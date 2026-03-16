package claim_service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	empModel "sighthub-backend/internal/models/employees"
	insuranceModel "sighthub-backend/internal/models/insurance"
	invoiceModel "sighthub-backend/internal/models/invoices"
	superBill "sighthub-backend/internal/models/medical/vision_exam/super_bill"
	patientModel "sighthub-backend/internal/models/patients"
	svcModel "sighthub-backend/internal/models/service"
	modelTypes "sighthub-backend/internal/models/types"
	visionModel "sighthub-backend/internal/models/vision_exam"
	"sighthub-backend/pkg/activitylog"
)

// ── GET /invoices ─────────────────────────────────────────────────────────────

var validInsuranceStatuses = map[string]bool{
	"Billed": true, "Pending": true, "Paid": true,
	"Prep": true, "Canceled": true, "Accept": true,
}

type InsuranceInvoiceItem struct {
	InvoiceID        int64   `json:"invoice_id"`
	InvoiceNumber    string  `json:"invoice_number"`
	Date             *string `json:"date"`
	PatientFirstName string  `json:"patient_first_name"`
	PatientLastName  string  `json:"patient_last_name"`
	InsuranceStatus  string  `json:"insurance_status"`
}

type InsuranceInvoicesResult struct {
	Items   []InsuranceInvoiceItem `json:"items"`
	Total   int64                  `json:"total"`
	Page    int                    `json:"page"`
	PerPage int                    `json:"per_page"`
	Pages   int64                  `json:"pages"`
}

func (s *Service) GetInsuranceInvoices(username string, params map[string]string, page, perPage int) (*InsuranceInvoicesResult, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	db := s.db.Model(&invoiceModel.Invoice{}).
		Joins("LEFT JOIN invoice_insurance_policy iip ON iip.invoice_id = invoice.id_invoice").
		Joins("JOIN patient p ON p.id_patient = invoice.patient_id").
		Where("invoice.location_id = ?", loc.IDLocation).
		Where("(iip.invoice_id IS NOT NULL OR invoice.ins_bal > 0)")

	if status := params["status"]; status != "" {
		statuses := strings.Split(status, ",")
		valid := make([]string, 0, len(statuses))
		for _, st := range statuses {
			st = strings.TrimSpace(st)
			if !validInsuranceStatuses[st] {
				return nil, fmt.Errorf("invalid status value: %s", st)
			}
			valid = append(valid, st)
		}
		db = db.Where("invoice.paid_insurance_status IN ?", valid)
	}

	if ds := params["date_start"]; ds != "" {
		t, err := time.Parse("2006-01-02", ds)
		if err != nil {
			return nil, errors.New("invalid date_start format")
		}
		db = db.Where("DATE(invoice.date_create) >= ?", t.Format("2006-01-02"))
	}
	if de := params["date_end"]; de != "" {
		t, err := time.Parse("2006-01-02", de)
		if err != nil {
			return nil, errors.New("invalid date_end format")
		}
		db = db.Where("DATE(invoice.date_create) <= ?", t.Format("2006-01-02"))
	}

	if num := params["invoice_number"]; num != "" {
		db = db.Where("invoice.number_invoice ILIKE ?", "%"+num+"%")
	}

	if compIDStr := params["insurance_company_id"]; compIDStr != "" {
		db = db.Joins("JOIN insurance_policy ip2 ON ip2.id_insurance_policy = iip.insurance_policy_id").
			Where("ip2.insurance_company_id = ?", compIDStr)
	}

	if name := params["patient_name"]; name != "" {
		pat := "%" + name + "%"
		db = db.Where(
			"CONCAT(p.first_name, ' ', p.last_name) ILIKE ? OR CONCAT(p.last_name, ' ', p.first_name) ILIKE ?",
			pat, pat,
		)
	}

	db = db.Order("invoice.date_create DESC")

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	var invoices []invoiceModel.Invoice
	if err := db.Preload("Patient").
		Offset((page - 1) * perPage).Limit(perPage).
		Find(&invoices).Error; err != nil {
		return nil, err
	}

	items := make([]InsuranceInvoiceItem, 0, len(invoices))
	for _, inv := range invoices {
		var hasPayment bool
		s.db.Raw("SELECT EXISTS(SELECT 1 FROM insurance_payment WHERE invoice_id = ?)", inv.IDInvoice).Scan(&hasPayment)

		status := ""
		if inv.PaidInsuranceStatus != nil {
			status = string(*inv.PaidInsuranceStatus)
		}
		if status == "" {
			status = "Prep"
		}
		if hasPayment && status != "Paid" {
			status = "Paid"
		}

		d := inv.DateCreate.Format(time.RFC3339)
		firstName, lastName := "", ""
		if inv.Patient != nil {
			firstName = inv.Patient.FirstName
			lastName = inv.Patient.LastName
		}

		items = append(items, InsuranceInvoiceItem{
			InvoiceID:        inv.IDInvoice,
			InvoiceNumber:    inv.NumberInvoice,
			Date:             &d,
			PatientFirstName: firstName,
			PatientLastName:  lastName,
			InsuranceStatus:  status,
		})
	}

	pages := (total + int64(perPage) - 1) / int64(perPage)
	return &InsuranceInvoicesResult{
		Items:   items,
		Total:   total,
		Page:    page,
		PerPage: perPage,
		Pages:   pages,
	}, nil
}

// ── GET /insurance-companies ──────────────────────────────────────────────────

func (s *Service) GetInsuranceCompanies() ([]map[string]interface{}, error) {
	var companies []insuranceModel.InsuranceCompany
	if err := s.db.Order("company_name").Find(&companies).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, len(companies))
	for i, c := range companies {
		result[i] = map[string]interface{}{
			"id_insurance_company": c.IDInsuranceCompany,
			"company_name":         c.CompanyName,
		}
	}
	return result, nil
}

// ── GET /insurance-payment-types ──────────────────────────────────────────────

func (s *Service) GetInsurancePaymentTypes() ([]map[string]interface{}, error) {
	var pts []insuranceModel.InsurancePaymentType
	if err := s.db.Where("active = ?", true).Order("name").Find(&pts).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, len(pts))
	for i, pt := range pts {
		result[i] = map[string]interface{}{
			"payment_type_id": pt.IDInsurancePaymentType,
			"name":            pt.Name,
		}
	}
	return result, nil
}

// ── GET /invoices/:invoice_id/insurance-payment ───────────────────────────────

func (s *Service) GetInvoicePaymentSummary(invoiceID int64) (map[string]interface{}, error) {
	var inv invoiceModel.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}

	var count int64
	s.db.Model(&invoiceModel.InvoiceInsurancePolicy{}).Where("invoice_id = ?", invoiceID).Count(&count)
	if count == 0 {
		return nil, errors.New("no insurance policy attached")
	}

	var hasPayments bool
	s.db.Raw("SELECT EXISTS(SELECT 1 FROM insurance_payment WHERE invoice_id = ?)", invoiceID).Scan(&hasPayments)
	if !hasPayments {
		return nil, nil
	}

	patientPaid := s.sumPatientPayments(invoiceID)
	insurancePaid := s.sumInsurancePayments(invoiceID)
	giftCardPaid := 0.0
	if inv.GiftCardBal != nil {
		giftCardPaid = *inv.GiftCardBal
	}
	paidOther := patientPaid + giftCardPaid

	return map[string]interface{}{
		"invoice_number": inv.NumberInvoice,
		"invoice_total":  fmt.Sprintf("%.2f", inv.TotalAmount),
		"paid_other":     fmt.Sprintf("%.2f", paidOther),
		"paid_insurance": fmt.Sprintf("%.2f", insurancePaid),
		"pt_balance":     fmt.Sprintf("%.2f", inv.PTBal),
		"ins_balance":    fmt.Sprintf("%.2f", inv.InsBal),
		"total_balance":  fmt.Sprintf("%.2f", inv.Due),
	}, nil
}

// ── POST /invoices/:invoice_id/insurance-payment ──────────────────────────────

type AddInsurancePaymentInput struct {
	PaymentTypeID   int     `json:"payment_type_id"`
	Amount          float64 `json:"amount"`
	ReferenceNumber string  `json:"reference_number"`
	Note            string  `json:"note"`
	Adjust          float64 `json:"adjust"`
}

func (s *Service) AddInsurancePayment(username string, invoiceID int64, input AddInsurancePaymentInput) (map[string]interface{}, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var inv invoiceModel.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}
	if inv.Finalized {
		return nil, errors.New("invoice is finalized")
	}

	var invIns invoiceModel.InvoiceInsurancePolicy
	if err := s.db.Where("invoice_id = ?", invoiceID).First(&invIns).Error; err != nil {
		return nil, errors.New("no insurance policy attached")
	}

	if input.PaymentTypeID == 0 {
		return nil, errors.New("payment_type_id is required")
	}
	var pt insuranceModel.InsurancePaymentType
	if err := s.db.First(&pt, input.PaymentTypeID).Error; err != nil || !pt.Active {
		return nil, errors.New("invalid or inactive payment type")
	}

	if input.Amount <= 0 {
		return nil, errors.New("amount must be > 0")
	}

	oldDue := inv.Due
	var mainPaymentID, adjustPaymentID int64

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		empID := int64(emp.IDEmployee)
		ref := strings.TrimSpace(input.ReferenceNumber)
		note := strings.TrimSpace(input.Note)

		var refPtr, notePtr *string
		if ref != "" {
			refPtr = &ref
		}
		if note != "" {
			notePtr = &note
		}

		ip := insuranceModel.InsurancePayment{
			InvoiceID:         invoiceID,
			InsurancePolicyID: invIns.InsurancePolicyID,
			PaymentTypeID:     input.PaymentTypeID,
			Amount:            fmt.Sprintf("%.2f", input.Amount),
			ReferenceNumber:   refPtr,
			Note:              notePtr,
			EmployeeID:        &empID,
		}
		if err := tx.Create(&ip).Error; err != nil {
			return err
		}
		mainPaymentID = ip.IDInsurancePayment

		if input.Adjust != 0 {
			var adjustPT insuranceModel.InsurancePaymentType
			if err := tx.First(&adjustPT, 3).Error; err != nil || !adjustPT.Active {
				return errors.New("adjustment payment type not available")
			}
			adjNote := fmt.Sprintf("Adjustment for payment %d", mainPaymentID)
			if note != "" {
				adjNote = note
			}
			ipAdj := insuranceModel.InsurancePayment{
				InvoiceID:         invoiceID,
				InsurancePolicyID: invIns.InsurancePolicyID,
				PaymentTypeID:     3,
				Amount:            fmt.Sprintf("%.2f", input.Adjust),
				ReferenceNumber:   refPtr,
				Note:              &adjNote,
				EmployeeID:        &empID,
			}
			if err := tx.Create(&ipAdj).Error; err != nil {
				return err
			}
			adjustPaymentID = ipAdj.IDInsurancePayment
		}

		status := ""
		if inv.PaidInsuranceStatus != nil {
			status = string(*inv.PaidInsuranceStatus)
		}
		if status == "Prep" || status == "Accept" || status == "" {
			paid := modelTypes.PaidInsuranceStatusPaid
			inv.PaidInsuranceStatus = &paid
		}
		s.recalcInvoice(&inv)

		return tx.Save(&inv).Error
	})
	if txErr != nil {
		return nil, txErr
	}

	activitylog.Log(s.db, "claim", "payment_add",
		activitylog.WithEntity(invoiceID),
		activitylog.WithDetails(map[string]interface{}{
			"amount":          fmt.Sprintf("%.2f", input.Amount),
			"payment_type_id": input.PaymentTypeID,
		}))

	result := map[string]interface{}{
		"message":              "Insurance payment recorded",
		"insurance_payment_id": mainPaymentID,
		"amount":               fmt.Sprintf("%.2f", input.Amount),
		"payment_type":         pt.Name,
		"old_due":              fmt.Sprintf("%.2f", oldDue),
		"new_due":              fmt.Sprintf("%.2f", inv.Due),
	}
	if adjustPaymentID != 0 {
		result["adjustment_payment_id"] = adjustPaymentID
		result["adjustment_amount"] = fmt.Sprintf("%.2f", input.Adjust)
	}
	return result, nil
}

// ── GET /super-bill/:invoice_id ───────────────────────────────────────────────

func (s *Service) GetSuperBill(invoiceID int64) (map[string]interface{}, error) {
	var inv invoiceModel.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}

	var exam superBill.SuperEyeExam
	if err := s.db.Where("invoice_id = ?", invoiceID).First(&exam).Error; err != nil {
		return nil, nil // no super bill
	}

	var items []invoiceModel.InvoiceItemSale
	s.db.Where("invoice_id = ?", invoiceID).Find(&items)

	cptMap := make(map[string]interface{})
	for _, item := range items {
		if item.ItemID == nil {
			continue
		}
		itemID := *item.ItemID

		var ps svcModel.ProfessionalService
		s.db.First(&ps, itemID)

		var diagnoses []superBill.SuperBillDiagnosis
		s.db.Where("professional_service_id = ? AND super_eye_exam_id = ?", itemID, exam.IDSuperEyeExam).Find(&diagnoses)

		diseases := make([]map[string]interface{}, 0)
		for _, diag := range diagnoses {
			var diseaseRecs []superBill.DiseaseSuperBill
			s.db.Where("super_bill_diagnosis_id = ?", diag.IDSuperBillDiagnosis).Find(&diseaseRecs)
			for _, d := range diseaseRecs {
				diseases = append(diseases, map[string]interface{}{
					"id_disease_super_bill": d.IDDiseaseSuperBill,
					"level_id":              d.LevelID,
					"type":                  d.Type,
					"code":                  d.Code,
					"title":                 d.Title,
					"group_set":             d.GroupSet,
					"include":               d.Include,
				})
			}
		}

		ptBal := item.Total
		if item.PtBalance != nil {
			ptBal = *item.PtBalance
		}
		insBal := 0.0
		if item.InsBalance != nil {
			insBal = *item.InsBalance
		}

		cptMap[fmt.Sprintf("%d", itemID)] = map[string]interface{}{
			"item_id":        itemID,
			"cpt_hcpcs_code": ps.CptHcpcsCode,
			"description":    ps.InvoiceDesc,
			"quantity":       item.Quantity,
			"price":          fmt.Sprintf("%.2f", item.Price),
			"total":          fmt.Sprintf("%.2f", item.Total),
			"pt_balance":     fmt.Sprintf("%.2f", ptBal),
			"ins_balance":    fmt.Sprintf("%.2f", insBal),
			"diseases":       diseases,
		}
	}

	return map[string]interface{}{
		"total_amount":   fmt.Sprintf("%.2f", inv.TotalAmount),
		"cpt_hcpcs_code": cptMap,
	}, nil
}

// ── PUT /super-bill/:invoice_id ───────────────────────────────────────────────

func (s *Service) UpdateSuperBill(invoiceID int64, servicesData, diagnosesData map[string]interface{}) error {
	var inv invoiceModel.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invoice not found")
		}
		return err
	}

	var exam superBill.SuperEyeExam
	if err := s.db.Where("invoice_id = ?", invoiceID).First(&exam).Error; err != nil {
		return errors.New("super bill not found for this invoice")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for itemIDStr, svcPayload := range servicesData {
			payload, ok := svcPayload.(map[string]interface{})
			if !ok {
				continue
			}
			inclRaw, ok := payload["include"]
			if !ok {
				continue
			}
			svcInclude, _ := inclRaw.(bool)

			var diagnoses []superBill.SuperBillDiagnosis
			tx.Where("professional_service_id = ? AND super_eye_exam_id = ?", itemIDStr, exam.IDSuperEyeExam).Find(&diagnoses)
			for _, diag := range diagnoses {
				tx.Model(&superBill.DiseaseSuperBill{}).
					Where("super_bill_diagnosis_id = ?", diag.IDSuperBillDiagnosis).
					Update("include", svcInclude)
			}
		}

		for diseaseIDStr, diagPayload := range diagnosesData {
			payload, ok := diagPayload.(map[string]interface{})
			if !ok {
				continue
			}
			inclRaw, ok := payload["include"]
			if !ok {
				continue
			}
			diagInclude, _ := inclRaw.(bool)

			var disease superBill.DiseaseSuperBill
			if err := tx.First(&disease, diseaseIDStr).Error; err != nil {
				return fmt.Errorf("diagnosis %s not found", diseaseIDStr)
			}

			if disease.SuperBillDiagnosisID != nil {
				var diag superBill.SuperBillDiagnosis
				if tx.First(&diag, *disease.SuperBillDiagnosisID).Error == nil {
					if diag.SuperEyeExamID != exam.IDSuperEyeExam {
						return fmt.Errorf("diagnosis %s does not belong to this invoice's super bill", diseaseIDStr)
					}
				}
			}
			disease.Include = diagInclude
			tx.Save(&disease)
		}

		activitylog.Log(tx, "claim", "superbill_update", activitylog.WithEntity(invoiceID))
		return nil
	})
}

// ── GET /invoices/:invoice_id/patient ─────────────────────────────────────────

func (s *Service) GetInvoicePatient(invoiceID int64) (map[string]interface{}, error) {
	var inv invoiceModel.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}

	var patient patientModel.Patient
	if err := s.db.First(&patient, inv.PatientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}

	var dob *string
	if patient.DOB != nil {
		d := patient.DOB.Format("2006-01-02")
		dob = &d
	}

	return map[string]interface{}{
		"id_patient":     patient.IDPatient,
		"first_name":     patient.FirstName,
		"last_name":      patient.LastName,
		"dob":            dob,
		"gender":         patient.Gender,
		"phone":          patient.Phone,
		"phone_home":     patient.PhoneHome,
		"cell_work":      patient.CellWork,
		"email":          patient.Email,
		"street_address": patient.StreetAddress,
		"address_line_2": patient.AddressLine2,
		"city":           patient.City,
		"state":          patient.State,
		"zip_code":       patient.ZipCode,
	}, nil
}

// ── GET /invoices/:invoice_id/responsible-party ────────────────────────────────

func (s *Service) GetResponsibleParty(invoiceID int64) (map[string]interface{}, error) {
	var inv invoiceModel.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}

	var invIns invoiceModel.InvoiceInsurancePolicy
	if err := s.db.Where("invoice_id = ?", invoiceID).First(&invIns).Error; err != nil {
		return nil, errors.New("no insurance policy attached to this invoice")
	}

	var holder patientModel.InsuranceHolderPatients
	if err := s.db.Where("insurance_policy_id = ? AND holder_type = ?", invIns.InsurancePolicyID, "Self").
		First(&holder).Error; err != nil {
		return nil, errors.New("no primary holder found for this insurance policy")
	}

	var patient patientModel.Patient
	if err := s.db.First(&patient, holder.PatientID).Error; err != nil {
		return nil, errors.New("holder patient not found")
	}

	var dob *string
	if patient.DOB != nil {
		d := patient.DOB.Format("2006-01-02")
		dob = &d
	}

	return map[string]interface{}{
		"holder_type":    holder.HolderType,
		"member_number":  holder.MemberNumber,
		"id_patient":     patient.IDPatient,
		"first_name":     patient.FirstName,
		"last_name":      patient.LastName,
		"dob":            dob,
		"gender":         patient.Gender,
		"phone":          patient.Phone,
		"phone_home":     patient.PhoneHome,
		"cell_work":      patient.CellWork,
		"email":          patient.Email,
		"street_address": patient.StreetAddress,
		"address_line_2": patient.AddressLine2,
		"city":           patient.City,
		"state":          patient.State,
		"zip_code":       patient.ZipCode,
	}, nil
}

// ── PUT /invoices/:invoice_id/insurance-status ────────────────────────────────

var validInsuranceUpdateStatuses = map[string]bool{
	"Prep": true, "Accept": true, "Canceled": true, "Paid": true,
}

func (s *Service) UpdateInsuranceStatus(username string, invoiceID int64, status string) (map[string]interface{}, error) {
	_, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	if !validInsuranceUpdateStatuses[status] {
		return nil, errors.New("invalid status value")
	}

	var inv invoiceModel.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}

	newStatus := modelTypes.PaidInsuranceStatus(status)
	inv.PaidInsuranceStatus = &newStatus
	if err := s.db.Save(&inv).Error; err != nil {
		return nil, err
	}

	activitylog.Log(s.db, "claim", "insurance_status_update",
		activitylog.WithEntity(invoiceID),
		activitylog.WithDetails(map[string]interface{}{"status": status}))

	return map[string]interface{}{
		"message":    "Insurance status updated",
		"invoice_id": inv.IDInvoice,
		"status":     status,
	}, nil
}

// ── GET /invoices/:invoice_id/secondary-insurance ─────────────────────────────

func (s *Service) GetSecondaryInsurance(invoiceID int64) (map[string]interface{}, error) {
	var inv invoiceModel.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}

	var holder patientModel.InsuranceHolderPatients
	if err := s.db.Where("patient_id = ? AND position = ?", inv.PatientID, "Secondary").
		First(&holder).Error; err != nil {
		return nil, nil // no secondary insurance
	}

	var policy insuranceModel.InsurancePolicy
	if err := s.db.First(&policy, holder.InsurancePolicyID).Error; err != nil {
		return nil, nil
	}

	var companyName *string
	var company insuranceModel.InsuranceCompany
	if s.db.First(&company, policy.InsuranceCompanyID).Error == nil {
		cn := company.CompanyName
		companyName = &cn
	}

	return map[string]interface{}{
		"insurance_policy_id": policy.IDInsurancePolicy,
		"insurance_company":   companyName,
		"policy_holder":       holder.HolderType,
		"group_number":        policy.GroupNumber,
		"member_number":       holder.MemberNumber,
	}, nil
}

// ── GET /invoices/:invoice_id/insurance-info ──────────────────────────────────

func (s *Service) GetInvoiceInsuranceInfo(invoiceID int64) (map[string]interface{}, error) {
	var inv invoiceModel.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}

	var invIns invoiceModel.InvoiceInsurancePolicy
	if err := s.db.Where("invoice_id = ?", invoiceID).First(&invIns).Error; err != nil {
		return nil, nil
	}

	var policy insuranceModel.InsurancePolicy
	if err := s.db.First(&policy, invIns.InsurancePolicyID).Error; err != nil {
		return nil, nil
	}

	var companyName *string
	var company insuranceModel.InsuranceCompany
	if s.db.First(&company, policy.InsuranceCompanyID).Error == nil {
		cn := company.CompanyName
		companyName = &cn
	}

	result := map[string]interface{}{
		"insurance_policy_id": policy.IDInsurancePolicy,
		"insurance_company":   companyName,
		"policy_holder":       nil,
		"position":            nil,
		"group_number":        policy.GroupNumber,
		"member_number":       nil,
	}

	var holder patientModel.InsuranceHolderPatients
	if s.db.Where("insurance_policy_id = ? AND patient_id = ?", policy.IDInsurancePolicy, inv.PatientID).First(&holder).Error == nil {
		result["policy_holder"] = holder.HolderType
		result["position"] = holder.Position
		result["member_number"] = holder.MemberNumber
	}
	return result, nil
}

// ── GET /invoices/:invoice_id/claim-info ──────────────────────────────────────

func (s *Service) GetClaimInfo(invoiceID int64) (map[string]interface{}, error) {
	var inv invoiceModel.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}

	status := ""
	if inv.PaidInsuranceStatus != nil {
		status = string(*inv.PaidInsuranceStatus)
	}
	if status == "" {
		status = "Prep"
	}

	result := map[string]interface{}{
		"insurance_status":  status,
		"invoice_number":    inv.NumberInvoice,
		"invoice_date":      inv.DateCreate.Format(time.RFC3339),
		"doctor_first_name": nil,
		"doctor_last_name":  nil,
		"exam_date":         nil,
	}

	var exam superBill.SuperEyeExam
	if s.db.Where("invoice_id = ?", invoiceID).First(&exam).Error == nil {
		var eyeExam visionModel.EyeExam
		if s.db.First(&eyeExam, exam.EyeExamID).Error == nil {
			result["exam_date"] = eyeExam.EyeExamDate.Format(time.RFC3339)
			var doc empModel.Employee
			if s.db.First(&doc, eyeExam.EmployeeID).Error == nil {
				result["doctor_first_name"] = doc.FirstName
				result["doctor_last_name"] = doc.LastName
			}
		}
	}

	return result, nil
}

// ── GET /doctors ──────────────────────────────────────────────────────────────

func (s *Service) GetDoctors() ([]map[string]interface{}, error) {
	var doctors []empModel.Employee
	err := s.db.
		Joins("JOIN doctor_npi_number dnn ON dnn.employee_id = employee.id_employee").
		Where("employee.active = ?", true).
		Order("employee.last_name, employee.first_name").
		Find(&doctors).Error
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(doctors))
	for i, emp := range doctors {
		result[i] = map[string]interface{}{
			"doctor_id":   emp.IDEmployee,
			"doctor_name": fmt.Sprintf("Dr. %s %s", emp.FirstName, emp.LastName),
		}
	}
	return result, nil
}
