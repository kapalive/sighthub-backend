package invoice

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	insModel "sighthub-backend/internal/models/insurance"
	"sighthub-backend/internal/models/invoices"
	patModel "sighthub-backend/internal/models/patients"
)

// ─── Input DTOs ───────────────────────────────────────────────────────────────

type PatientPaymentInput struct {
	Amount          string `json:"amount"`
	PaymentMethodID int64  `json:"payment_method_id"`
}

type CreditPaymentInput struct {
	CreditAmount string `json:"credit_amount"`
}

type DiscountInput struct {
	Discount float64 `json:"discount"`
}

type InsurancePaymentInput struct {
	Amount          string  `json:"amount"`
	PaymentTypeID   int     `json:"payment_type_id"`
	ReferenceNumber *string `json:"reference_number"`
	Note            *string `json:"note"`
}

type TransferCreditInput struct {
	PayerPatientID int64  `json:"payer_patient_id"`
	Amount         string `json:"amount"`
}

type UpdatePaymentInput struct {
	Amount          *string `json:"amount"`
	PaymentMethodID *int64  `json:"payment_method_id"`
	TransactionHash *string `json:"transaction_hash"`
}

// ─── AddPatientPayment ────────────────────────────────────────────────────────

type PatientPaymentResult struct {
	Message     string  `json:"message"`
	AmountPaid  float64 `json:"amount_paid"`
	OldDue      float64 `json:"old_due"`
	NewDue      float64 `json:"new_due"`
	CreditAdded float64 `json:"credit_added"`
}

func (s *Service) AddPatientPayment(username string, invoiceID int64, input PatientPaymentInput) (*PatientPaymentResult, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, errors.New("invoice not found")
	}
	if inv.LocationID != int64(loc.IDLocation) {
		return nil, errors.New("payments only at invoice location")
	}
	if inv.Finalized {
		return nil, errors.New("invoice is finalized")
	}

	amount, err := strconv.ParseFloat(input.Amount, 64)
	if err != nil || amount == 0 {
		return nil, errors.New("invalid amount")
	}

	isAdjustment := input.PaymentMethodID == 22
	if !isAdjustment && amount <= 0 {
		return nil, errors.New("amount must be > 0")
	}
	if isAdjustment && amount == 0 {
		return nil, errors.New("adjustment amount cannot be zero")
	}

	var result *PatientPaymentResult
	err = s.db.Transaction(func(tx *gorm.DB) error {
		oldDue := inv.Due

		empID := int64(emp.IDEmployee)
		ph := patModel.PaymentHistory{
			PatientID:        inv.PatientID,
			InvoiceID:        invoiceID,
			Amount:           amount,
			PaymentTimestamp: time.Now(),
			PaymentMethodID:  &input.PaymentMethodID,
			EmployeeID:       &empID,
		}
		if err := tx.Create(&ph).Error; err != nil {
			return err
		}

		if err := s.recalcInvoice(tx, &inv); err != nil {
			return err
		}

		creditAdded := 0.0
		if !isAdjustment {
			leftover := amount - oldDue
			if leftover > 0 {
				creditAdded = leftover
				var cb patModel.ClientBalance
				err := tx.Where("patient_id = ? AND location_id = ?", *inv.PatientID, inv.LocationID).First(&cb).Error
				if err != nil {
					cb = patModel.ClientBalance{
						PatientID:  *inv.PatientID,
						LocationID: int(inv.LocationID),
						Credit:     0,
					}
					if err2 := tx.Create(&cb).Error; err2 != nil {
						return err2
					}
				}
				cb.Credit += leftover
				if err := tx.Save(&cb).Error; err != nil {
					return err
				}
			}
		}

		result = &PatientPaymentResult{
			Message:     "Patient payment recorded",
			AmountPaid:  amount,
			OldDue:      oldDue,
			NewDue:      inv.Due,
			CreditAdded: creditAdded,
		}
		return nil
	})
	return result, err
}

// ─── PayWithCredit ────────────────────────────────────────────────────────────

type CreditPaymentResult struct {
	Message   string  `json:"message"`
	CreditUsed float64 `json:"credit_used"`
	NewDue    float64 `json:"new_due"`
	NewCredit float64 `json:"new_credit"`
}

func (s *Service) PayWithCredit(username string, invoiceID int64, input CreditPaymentInput) (*CreditPaymentResult, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, errors.New("invoice not found")
	}
	if inv.LocationID != int64(loc.IDLocation) {
		return nil, errors.New("credit payments only at invoice location")
	}
	if inv.Finalized {
		return nil, errors.New("invoice is finalized")
	}

	currentDue := inv.Due
	if currentDue <= 0 {
		return nil, errors.New("invoice has no due balance")
	}

	creditAmount, err := strconv.ParseFloat(input.CreditAmount, 64)
	if err != nil || creditAmount <= 0 {
		return nil, errors.New("amount must be > 0")
	}
	if creditAmount > currentDue {
		return nil, errors.New("cannot exceed invoice due")
	}

	var result *CreditPaymentResult
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var cb patModel.ClientBalance
		if err := tx.Where("patient_id = ? AND location_id = ?", inv.PatientID, inv.LocationID).First(&cb).Error; err != nil {
			return errors.New("not enough credit")
		}
		if cb.Credit < creditAmount {
			return errors.New("not enough credit")
		}
		cb.Credit -= creditAmount
		if err := tx.Save(&cb).Error; err != nil {
			return err
		}

		pmID := int64(20)
		empID := int64(emp.IDEmployee)
		ph := patModel.PaymentHistory{
			PatientID:        inv.PatientID,
			InvoiceID:        invoiceID,
			Amount:           creditAmount,
			PaymentTimestamp: time.Now(),
			PaymentMethodID:  &pmID,
			EmployeeID:       &empID,
		}
		if err := tx.Create(&ph).Error; err != nil {
			return err
		}

		if err := s.recalcInvoice(tx, &inv); err != nil {
			return err
		}

		result = &CreditPaymentResult{
			Message:    "Paid with patient credit",
			CreditUsed: creditAmount,
			NewDue:     inv.Due,
			NewCredit:  cb.Credit,
		}
		return nil
	})
	return result, err
}

// ─── AddDiscount ─────────────────────────────────────────────────────────────

func (s *Service) AddDiscount(username string, invoiceID int64, input DiscountInput) (float64, error) {
	_, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return 0, err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return 0, errors.New("invoice not found")
	}
	if inv.Finalized {
		return 0, errors.New("invoice is finalized")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		inv.Discount = &input.Discount
		if err := tx.Save(&inv).Error; err != nil {
			return err
		}
		return s.recalcInvoice(tx, &inv)
	})
	if err != nil {
		return 0, err
	}
	return inv.Due, nil
}

// ─── GetInsurancePaymentTypes ─────────────────────────────────────────────────

func (s *Service) GetInsurancePaymentTypes() ([]insModel.InsurancePaymentType, error) {
	var types []insModel.InsurancePaymentType
	err := s.db.Where("active = true").Find(&types).Error
	return types, err
}

// ─── AddInsurancePayment ──────────────────────────────────────────────────────

type InsurancePaymentResult struct {
	Message            string  `json:"message"`
	InsurancePaymentID int64   `json:"insurance_payment_id"`
	Amount             float64 `json:"amount"`
	PaymentTypeName    string  `json:"payment_type"`
	OldDue             float64 `json:"old_due"`
	NewDue             float64 `json:"new_due"`
}

func (s *Service) AddInsurancePayment(username string, invoiceID int64, input InsurancePaymentInput) (*InsurancePaymentResult, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, errors.New("invoice not found")
	}
	if inv.Finalized {
		return nil, errors.New("invoice is finalized")
	}

	var invIns invoices.InvoiceInsurancePolicy
	if err := s.db.Where("invoice_id = ?", invoiceID).First(&invIns).Error; err != nil {
		return nil, errors.New("no insurance policy attached")
	}

	var pt insModel.InsurancePaymentType
	if err := s.db.First(&pt, input.PaymentTypeID).Error; err != nil || !pt.Active {
		return nil, errors.New("invalid or inactive payment type")
	}

	amount, err := strconv.ParseFloat(input.Amount, 64)
	if err != nil || amount <= 0 {
		return nil, errors.New("amount must be > 0")
	}

	var result *InsurancePaymentResult
	err = s.db.Transaction(func(tx *gorm.DB) error {
		oldDue := inv.Due
		empID := int64(emp.IDEmployee)

		ip := insModel.InsurancePayment{
			InvoiceID:         invoiceID,
			InsurancePolicyID: invIns.InsurancePolicyID,
			PaymentTypeID:     input.PaymentTypeID,
			Amount:            fmt.Sprintf("%.2f", amount),
			ReferenceNumber:   input.ReferenceNumber,
			Note:              input.Note,
			EmployeeID:        &empID,
		}
		if err := tx.Create(&ip).Error; err != nil {
			return err
		}

		if err := s.recalcInvoice(tx, &inv); err != nil {
			return err
		}

		result = &InsurancePaymentResult{
			Message:            "Insurance payment recorded",
			InsurancePaymentID: ip.IDInsurancePayment,
			Amount:             amount,
			PaymentTypeName:    pt.Name,
			OldDue:             oldDue,
			NewDue:             inv.Due,
		}
		return nil
	})
	return result, err
}

// ─── GetInsurancePayments ─────────────────────────────────────────────────────

type InsurancePaymentDetail struct {
	IDInsurancePayment int64   `json:"id_insurance_payment"`
	Amount             string  `json:"amount"`
	PaymentTypeID      int     `json:"payment_type_id"`
	PaymentTypeName    *string `json:"payment_type_name"`
	ReferenceNumber    *string `json:"reference_number"`
	Note               *string `json:"note"`
	Employee           *string `json:"employee"`
	CreatedAt          *string `json:"created_at"`
}

func (s *Service) GetInsurancePayments(invoiceID int64) ([]InsurancePaymentDetail, error) {
	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, errors.New("invoice not found")
	}

	var payments []insModel.InsurancePayment
	if err := s.db.Where("invoice_id = ?", invoiceID).
		Order("created_at desc").Find(&payments).Error; err != nil {
		return nil, err
	}

	result := make([]InsurancePaymentDetail, 0, len(payments))
	for _, p := range payments {
		detail := InsurancePaymentDetail{
			IDInsurancePayment: p.IDInsurancePayment,
			Amount:             p.Amount,
			PaymentTypeID:      p.PaymentTypeID,
			ReferenceNumber:    p.ReferenceNumber,
			Note:               p.Note,
		}

		// Load payment type name
		var pt insModel.InsurancePaymentType
		if err := s.db.First(&pt, p.PaymentTypeID).Error; err == nil {
			detail.PaymentTypeName = &pt.Name
		}

		// Load employee name
		if p.EmployeeID != nil {
			type empName struct {
				FirstName string
				LastName  string
			}
			var en empName
			if err := s.db.Table("employee").Select("first_name, last_name").
				Where("id_employee = ?", *p.EmployeeID).Scan(&en).Error; err == nil {
				name := en.FirstName + " " + en.LastName
				detail.Employee = &name
			}
		}

		if p.CreatedAt != nil {
			s := p.CreatedAt.Format(time.RFC3339)
			detail.CreatedAt = &s
		}

		result = append(result, detail)
	}
	return result, nil
}

// ─── DeleteInsurancePayment ───────────────────────────────────────────────────

func (s *Service) DeleteInsurancePayment(username string, invoiceID, paymentID int64) (float64, error) {
	_, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return 0, err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return 0, errors.New("invoice not found")
	}
	if inv.Finalized {
		return 0, errors.New("invoice is finalized")
	}

	var ip insModel.InsurancePayment
	if err := s.db.Where("id_insurance_payment = ? AND invoice_id = ?", paymentID, invoiceID).First(&ip).Error; err != nil {
		return 0, errors.New("insurance payment not found")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&ip).Error; err != nil {
			return err
		}
		return s.recalcInvoice(tx, &inv)
	})
	if err != nil {
		return 0, err
	}
	return inv.Due, nil
}

// ─── GetPaymentHistory ────────────────────────────────────────────────────────

type PaymentHistoryItem struct {
	PaymentID       int64   `json:"payment_id"`
	PaymentTimestamp *string `json:"payment_timestamp"`
	Amount          float64 `json:"amount"`
	PaymentMethodID *int64  `json:"payment_method_id"`
	PaymentMethod   string  `json:"payment_method"`
	MaskedTx        *string `json:"masked_transaction"`
	Employee        string  `json:"employee"`
	Note            *string `json:"note"`
}

func (s *Service) GetPaymentHistory(invoiceID int64) ([]PaymentHistoryItem, error) {
	type row struct {
		PaymentID        int64
		PaymentTimestamp *time.Time
		Amount           float64
		PaymentMethodID  *int64
		MethodName       *string
		CardMaskedNumber *string
		FirstName        *string
		LastName         *string
		Note             *string
	}

	var rows []row
	err := s.db.Raw(`
		SELECT
			ph.payment_id,
			ph.payment_timestamp,
			ph.amount,
			ph.payment_method_id,
			pm.method_name,
			pt.card_masked_number,
			e.first_name,
			e.last_name,
			ph.note
		FROM payment_history ph
		LEFT JOIN payment_method pm ON ph.payment_method_id = pm.id_payment_method
		LEFT JOIN payment_transaction pt ON ph.payment_transaction_id = pt.id_payment_transaction
		LEFT JOIN employee e ON ph.employee_id = e.id_employee
		WHERE ph.invoice_id = ?
		ORDER BY ph.payment_timestamp DESC
	`, invoiceID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]PaymentHistoryItem, 0, len(rows))
	for _, r := range rows {
		item := PaymentHistoryItem{
			PaymentID:       r.PaymentID,
			Amount:          r.Amount,
			PaymentMethodID: r.PaymentMethodID,
			Note:            r.Note,
		}
		if r.PaymentTimestamp != nil {
			ts := r.PaymentTimestamp.Format("2006-01-02 15:04:05")
			item.PaymentTimestamp = &ts
		}
		if r.MethodName != nil {
			item.PaymentMethod = *r.MethodName
		} else {
			item.PaymentMethod = "Unknown"
		}
		item.MaskedTx = r.CardMaskedNumber
		if r.FirstName != nil && r.LastName != nil {
			item.Employee = *r.FirstName + " " + *r.LastName
		} else {
			item.Employee = "Unknown"
		}
		result = append(result, item)
	}
	return result, nil
}

// ─── GetCreditPayments ────────────────────────────────────────────────────────

type CreditPaymentItem struct {
	PaymentID        int64   `json:"payment_id"`
	PaymentTimestamp *string `json:"payment_timestamp"`
	Amount           float64 `json:"amount"`
	InvoiceID        int64   `json:"invoice_id"`
	InvoiceNumber    *string `json:"invoice_number"`
	Employee         string  `json:"employee"`
}

func (s *Service) GetCreditPayments(username string, patientID int64) ([]CreditPaymentItem, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	type row struct {
		PaymentID        int64
		PaymentTimestamp *time.Time
		Amount           float64
		InvoiceID        int64
		NumberInvoice    *string
		FirstName        *string
		LastName         *string
	}

	var rows []row
	err = s.db.Raw(`
		SELECT
			ph.payment_id,
			ph.payment_timestamp,
			ph.amount,
			ph.invoice_id,
			i.number_invoice,
			e.first_name,
			e.last_name
		FROM payment_history ph
		JOIN invoice i ON ph.invoice_id = i.id_invoice
		LEFT JOIN employee e ON ph.employee_id = e.id_employee
		WHERE ph.patient_id = ?
		  AND ph.payment_method_id = 20
		  AND i.location_id = ?
		ORDER BY ph.payment_timestamp DESC
	`, patientID, loc.IDLocation).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]CreditPaymentItem, 0, len(rows))
	for _, r := range rows {
		item := CreditPaymentItem{
			PaymentID:     r.PaymentID,
			Amount:        r.Amount,
			InvoiceID:     r.InvoiceID,
			InvoiceNumber: r.NumberInvoice,
		}
		if r.PaymentTimestamp != nil {
			ts := r.PaymentTimestamp.Format("2006-01-02 15:04:05")
			item.PaymentTimestamp = &ts
		}
		if r.FirstName != nil && r.LastName != nil {
			item.Employee = *r.FirstName + " " + *r.LastName
		} else {
			item.Employee = "Unknown"
		}
		result = append(result, item)
	}
	return result, nil
}

// ─── GetCreditBalance ─────────────────────────────────────────────────────────

type CreditBalanceResult struct {
	PatientID       int64   `json:"patient_id"`
	LocationID      int     `json:"location_id"`
	LocationShortName *string `json:"location_short_name"`
	CreditBalance   string  `json:"credit_balance"`
}

func (s *Service) GetCreditBalance(username string, patientID int64) (*CreditBalanceResult, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var cb patModel.ClientBalance
	credit := 0.0
	if err := s.db.Where("patient_id = ? AND location_id = ?", patientID, loc.IDLocation).
		First(&cb).Error; err == nil {
		credit = cb.Credit
	}

	return &CreditBalanceResult{
		PatientID:         patientID,
		LocationID:        loc.IDLocation,
		LocationShortName: loc.ShortName,
		CreditBalance:     fmt.Sprintf("%.2f", credit),
	}, nil
}

// ─── DeletePayment ────────────────────────────────────────────────────────────

func (s *Service) DeletePayment(username string, invoiceID, paymentID int64) (float64, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return 0, err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return 0, errors.New("invoice not found")
	}
	if inv.Finalized {
		return 0, errors.New("invoice is finalized")
	}
	if inv.LocationID != int64(loc.IDLocation) {
		return 0, errors.New("wrong location")
	}

	var ph patModel.PaymentHistory
	if err := s.db.Where("payment_id = ? AND invoice_id = ?", paymentID, invoiceID).First(&ph).Error; err != nil {
		return 0, errors.New("payment not found")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Return credit if was credit payment
		if ph.PaymentMethodID != nil && *ph.PaymentMethodID == 20 {
			var cb patModel.ClientBalance
			err := tx.Where("patient_id = ? AND location_id = ?", ph.PatientID, inv.LocationID).First(&cb).Error
			if err != nil {
				cb = patModel.ClientBalance{
					PatientID:  *ph.PatientID,
					LocationID: int(inv.LocationID),
					Credit:     ph.Amount,
				}
				return tx.Create(&cb).Error
			}
			cb.Credit += ph.Amount
			if err := tx.Save(&cb).Error; err != nil {
				return err
			}
		}

		if err := tx.Delete(&ph).Error; err != nil {
			return err
		}
		return s.recalcInvoice(tx, &inv)
	})
	if err != nil {
		return 0, err
	}
	return inv.Due, nil
}

// ─── UpdatePayment ────────────────────────────────────────────────────────────

type UpdatePaymentResult struct {
	Message   string  `json:"message"`
	PaymentID int64   `json:"payment_id"`
	Amount    float64 `json:"amount"`
	Due       float64 `json:"due"`
}

func (s *Service) UpdatePayment(username string, invoiceID, paymentID int64, input UpdatePaymentInput) (*UpdatePaymentResult, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, errors.New("invoice not found")
	}
	if inv.Finalized {
		return nil, errors.New("invoice is finalized")
	}

	var ph patModel.PaymentHistory
	if err := s.db.Where("payment_id = ? AND invoice_id = ?", paymentID, invoiceID).First(&ph).Error; err != nil {
		return nil, errors.New("payment not found")
	}

	newMethodID := ph.PaymentMethodID
	if input.PaymentMethodID != nil {
		newMethodID = input.PaymentMethodID
	}

	newAmount := ph.Amount
	if input.Amount != nil {
		parsed, err := strconv.ParseFloat(*input.Amount, 64)
		if err != nil || parsed <= 0 {
			return nil, errors.New("invalid amount")
		}
		newAmount = parsed
	}

	// Credit payments must be at invoice location
	oldIsCredit := ph.PaymentMethodID != nil && *ph.PaymentMethodID == 20
	newIsCredit := newMethodID != nil && *newMethodID == 20
	if (oldIsCredit || newIsCredit) && inv.LocationID != int64(loc.IDLocation) {
		return nil, errors.New("credit updates only at invoice location")
	}

	var result *UpdatePaymentResult
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Roll back old credit
		if oldIsCredit {
			var cb patModel.ClientBalance
			err := tx.Where("patient_id = ? AND location_id = ?", ph.PatientID, inv.LocationID).First(&cb).Error
			if err != nil {
				cb = patModel.ClientBalance{
					PatientID:  *ph.PatientID,
					LocationID: int(inv.LocationID),
					Credit:     0,
				}
				if err2 := tx.Create(&cb).Error; err2 != nil {
					return err2
				}
			}
			cb.Credit += ph.Amount
			if err := tx.Save(&cb).Error; err != nil {
				return err
			}
		}

		// Apply new credit
		if newIsCredit {
			var cb patModel.ClientBalance
			if err := tx.Where("patient_id = ? AND location_id = ?", ph.PatientID, inv.LocationID).First(&cb).Error; err != nil {
				return errors.New("not enough credit")
			}
			if cb.Credit < newAmount {
				return errors.New("not enough credit")
			}
			cb.Credit -= newAmount
			if err := tx.Save(&cb).Error; err != nil {
				return err
			}
		}

		ph.Amount = newAmount
		ph.PaymentMethodID = newMethodID
		ph.TransactionHash = input.TransactionHash
		empID := int64(emp.IDEmployee)
		ph.EmployeeID = &empID

		if err := tx.Save(&ph).Error; err != nil {
			return err
		}

		if err := s.recalcInvoice(tx, &inv); err != nil {
			return err
		}

		result = &UpdatePaymentResult{
			Message:   "Payment updated",
			PaymentID: ph.PaymentID,
			Amount:    newAmount,
			Due:       inv.Due,
		}
		return nil
	})
	return result, err
}

// ─── TransferCredit ───────────────────────────────────────────────────────────

type TransferCreditResult struct {
	Message           string  `json:"message"`
	AmountTransferred float64 `json:"amount_transferred"`
	NewDue            float64 `json:"new_due"`
	PayerNewCredit    float64 `json:"payer_new_credit"`
}

func (s *Service) TransferCredit(username string, invoiceID int64, input TransferCreditInput) (*TransferCreditResult, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, errors.New("invoice not found")
	}
	if inv.LocationID != int64(loc.IDLocation) {
		return nil, errors.New("invoice is not in your current location")
	}
	if inv.Finalized {
		return nil, errors.New("invoice is finalized")
	}

	currentDue := inv.Due
	if currentDue <= 0 {
		return nil, errors.New("invoice has no due balance")
	}

	amount, err := strconv.ParseFloat(input.Amount, 64)
	if err != nil || amount <= 0 {
		return nil, errors.New("amount must be > 0")
	}
	if amount > currentDue {
		return nil, errors.New("amount cannot exceed invoice due")
	}

	// Load payer patient name for note
	type patName struct {
		FirstName string
		LastName  string
	}
	var payer patName
	if err := s.db.Table("patient").Select("first_name, last_name").
		Where("id_patient = ?", input.PayerPatientID).Scan(&payer).Error; err != nil || payer.FirstName == "" {
		return nil, errors.New("payer patient not found")
	}

	var result *TransferCreditResult
	err = s.db.Transaction(func(tx *gorm.DB) error {
		var cb patModel.ClientBalance
		if err := tx.Where("patient_id = ? AND location_id = ?", input.PayerPatientID, loc.IDLocation).First(&cb).Error; err != nil {
			return errors.New("payer has insufficient credit balance")
		}
		if cb.Credit < amount {
			return errors.New("payer has insufficient credit balance")
		}
		cb.Credit -= amount
		if err := tx.Save(&cb).Error; err != nil {
			return err
		}

		pmID := int64(23)
		empID := int64(emp.IDEmployee)
		note := fmt.Sprintf("Transfer credit from %s %s", payer.FirstName, payer.LastName)
		ph := patModel.PaymentHistory{
			PatientID:        inv.PatientID,
			InvoiceID:        invoiceID,
			Amount:           amount,
			PaymentTimestamp: time.Now(),
			PaymentMethodID:  &pmID,
			EmployeeID:       &empID,
			Note:             &note,
		}
		if err := tx.Create(&ph).Error; err != nil {
			return err
		}

		if err := s.recalcInvoice(tx, &inv); err != nil {
			return err
		}

		result = &TransferCreditResult{
			Message:           "Invoice paid via credit transfer",
			AmountTransferred: amount,
			NewDue:            inv.Due,
			PayerNewCredit:    cb.Credit,
		}
		return nil
	})
	return result, err
}

