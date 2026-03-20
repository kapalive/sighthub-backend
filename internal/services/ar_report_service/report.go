package ar_report_service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/invoices"
	locModel "sighthub-backend/internal/models/location"
	patModel "sighthub-backend/internal/models/patients"
	reportModel "sighthub-backend/internal/models/reports"
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ─── helpers ──────────────────────────────────────────────────────────────────

func (s *Service) getEmployeeAndLocation(username string) (*empModel.Employee, *locModel.Location, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, nil, errors.New("employee or location not found")
	}
	var emp empModel.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, nil, errors.New("employee or location not found")
	}
	if emp.LocationID == nil {
		return nil, nil, errors.New("employee or location not found")
	}
	var loc locModel.Location
	if err := s.db.First(&loc, *emp.LocationID).Error; err != nil {
		return nil, nil, errors.New("employee or location not found")
	}
	return &emp, &loc, nil
}

// GetEmployeeLocation returns just the location ID for the given username — used by the handler.
func (s *Service) GetEmployeeLocation(username string) (*empModel.Employee, int, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, 0, err
	}
	return emp, loc.IDLocation, nil
}

// ─── BalanceDue ───────────────────────────────────────────────────────────────

type BalanceDueInvoice struct {
	InvoiceID     int64   `json:"invoice_id"`
	NumberInvoice string  `json:"number_invoice"`
	PatientID     int64   `json:"patient_id"`
	Date          *string `json:"date"`
	Age           int     `json:"age"`
	LastName      string  `json:"last_name"`
	FirstName     string  `json:"first_name"`
	CellNumber    string  `json:"cell_number"`
	HomeNumber    string  `json:"home_number"`
	Total         float64 `json:"total"`
	Paid          float64 `json:"paid"`
	Due           float64 `json:"due"`
	PtDue         float64 `json:"pt_due"`
	InsDue        float64 `json:"ins_due"`
	Insurance     float64 `json:"insurance"`
}

type BalanceDueGroup struct {
	EmployeeID   *int64              `json:"employee_id"`
	EmployeeName string              `json:"employee_name"`
	Invoices     []BalanceDueInvoice `json:"invoices"`
}

func (s *Service) GetBalanceDue(username string, locationID int, periodMonths int) ([]BalanceDueGroup, error) {
	type row struct {
		IDInvoice         int64      `gorm:"column:id_invoice"`
		NumberInvoice     string     `gorm:"column:number_invoice"`
		EmployeeID        *int64     `gorm:"column:employee_id"`
		EmployeeFirstName *string    `gorm:"column:employee_first_name"`
		EmployeeLastName  *string    `gorm:"column:employee_last_name"`
		DateCreate        *time.Time `gorm:"column:date_create"`
		PatientID         int64      `gorm:"column:patient_id"`
		AgeDays           *float64   `gorm:"column:age_days"`
		PatLastName       string     `gorm:"column:pat_last_name"`
		PatFirstName      string     `gorm:"column:pat_first_name"`
		CellNumber        *string    `gorm:"column:cell_number"`
		HomeNumber        *string    `gorm:"column:home_number"`
		TotalAmount       float64    `gorm:"column:total_amount"`
		PTBal             float64    `gorm:"column:pt_bal"`
		Due               float64    `gorm:"column:due"`
		InsBal            float64    `gorm:"column:ins_bal"`
	}

	var rows []row
	err := s.db.Raw(`
		SELECT
			i.id_invoice,
			i.number_invoice,
			i.employee_id,
			e.first_name  AS employee_first_name,
			e.last_name   AS employee_last_name,
			i.date_create,
			i.patient_id,
			EXTRACT(EPOCH FROM (NOW() - i.date_create)) / 86400.0 AS age_days,
			p.last_name   AS pat_last_name,
			p.first_name  AS pat_first_name,
			p.phone       AS cell_number,
			p.phone_home  AS home_number,
			i.total_amount,
			i.pt_bal,
			i.due,
			i.ins_bal
		FROM invoice i
		JOIN patient p ON p.id_patient = i.patient_id
		LEFT JOIN employee e ON e.id_employee = i.employee_id
		WHERE i.location_id = ?
		  AND i.number_invoice LIKE 'S%%'
		  AND (? = 0 OR i.date_create >= NOW() - MAKE_INTERVAL(months => ?))
		ORDER BY i.employee_id, i.date_create DESC
	`, locationID, periodMonths, periodMonths).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	grouped := map[interface{}]*BalanceDueGroup{}
	order := []interface{}{}

	for _, r := range rows {
		var key interface{}
		if r.EmployeeID != nil {
			key = *r.EmployeeID
		} else {
			key = nil
		}

		if _, ok := grouped[key]; !ok {
			name := "Unknown"
			if r.EmployeeFirstName != nil && r.EmployeeLastName != nil {
				name = *r.EmployeeFirstName + " " + *r.EmployeeLastName
			}
			grouped[key] = &BalanceDueGroup{
				EmployeeID:   r.EmployeeID,
				EmployeeName: name,
				Invoices:     []BalanceDueInvoice{},
			}
			order = append(order, key)
		}

		age := 0
		if r.AgeDays != nil {
			age = int(*r.AgeDays)
		}
		var dateStr *string
		if r.DateCreate != nil {
			s := r.DateCreate.Format(time.RFC3339)
			dateStr = &s
		}
		cell := "N/A"
		if r.CellNumber != nil && *r.CellNumber != "" {
			cell = *r.CellNumber
		}
		home := "N/A"
		if r.HomeNumber != nil && *r.HomeNumber != "" {
			home = *r.HomeNumber
		}

		grouped[key].Invoices = append(grouped[key].Invoices, BalanceDueInvoice{
			InvoiceID:     r.IDInvoice,
			NumberInvoice: r.NumberInvoice,
			PatientID:     r.PatientID,
			Date:          dateStr,
			Age:           age,
			LastName:      r.PatLastName,
			FirstName:     r.PatFirstName,
			CellNumber:    cell,
			HomeNumber:    home,
			Total:         r.TotalAmount,
			Paid:          r.PTBal,
			Due:           r.Due,
			PtDue:         r.PTBal,
			InsDue:        r.InsBal,
			Insurance:     r.InsBal,
		})
	}

	result := make([]BalanceDueGroup, 0, len(order))
	for _, k := range order {
		result = append(result, *grouped[k])
	}
	return result, nil
}

// ─── Credits ──────────────────────────────────────────────────────────────────

type CreditInvoice struct {
	InvoiceID     int64   `json:"invoice_id"`
	NumberInvoice string  `json:"number_invoice"`
	PatientID     int64   `json:"patient_id"`
	Date          *string `json:"date"`
	Age           int     `json:"age"`
	LastName      string  `json:"last_name"`
	FirstName     string  `json:"first_name"`
	CellNumber    string  `json:"cell_number"`
	HomeNumber    string  `json:"home_number"`
	Total         float64 `json:"total"`
	Paid          float64 `json:"paid"`
	Due           float64 `json:"due"`
	Insurance     float64 `json:"insurance"`
	Credit        float64 `json:"credit"`
}

type CreditGroup struct {
	EmployeeID   *int64          `json:"employee_id"`
	EmployeeName string          `json:"employee_name"`
	Invoices     []CreditInvoice `json:"invoices"`
}

func (s *Service) GetCredits(username string, locationID int) ([]CreditGroup, error) {
	type row struct {
		IDInvoice         int64      `gorm:"column:id_invoice"`
		NumberInvoice     string     `gorm:"column:number_invoice"`
		EmployeeID        *int64     `gorm:"column:employee_id"`
		EmployeeFirstName *string    `gorm:"column:employee_first_name"`
		EmployeeLastName  *string    `gorm:"column:employee_last_name"`
		PatientID         int64      `gorm:"column:patient_id"`
		DateCreate        *time.Time `gorm:"column:date_create"`
		AgeDays           *float64   `gorm:"column:age_days"`
		PatLastName       string     `gorm:"column:pat_last_name"`
		PatFirstName      string     `gorm:"column:pat_first_name"`
		CellNumber        *string    `gorm:"column:cell_number"`
		HomeNumber        *string    `gorm:"column:home_number"`
		TotalAmount       float64    `gorm:"column:total_amount"`
		FinalAmount       float64    `gorm:"column:final_amount"`
		PTBal             float64    `gorm:"column:pt_bal"`
		Due               float64    `gorm:"column:due"`
		InsBal            float64    `gorm:"column:ins_bal"`
		GiftCardBal       *float64   `gorm:"column:gift_card_bal"`
		PtPaid            float64    `gorm:"column:pt_paid"`
		InsPaid           float64    `gorm:"column:ins_paid"`
	}

	var rows []row
	err := s.db.Raw(`
		SELECT
			i.id_invoice,
			i.number_invoice,
			i.employee_id,
			e.first_name  AS employee_first_name,
			e.last_name   AS employee_last_name,
			i.patient_id,
			i.date_create,
			EXTRACT(EPOCH FROM (NOW() - i.date_create)) / 86400.0 AS age_days,
			p.last_name   AS pat_last_name,
			p.first_name  AS pat_first_name,
			p.cell_work   AS cell_number,
			p.phone_home  AS home_number,
			i.total_amount,
			i.final_amount,
			i.pt_bal,
			i.due,
			i.ins_bal,
			i.gift_card_bal,
			COALESCE(pt.pt_paid,  0) AS pt_paid,
			COALESCE(ins.ins_paid, 0) AS ins_paid
		FROM invoice i
		JOIN patient p ON p.id_patient = i.patient_id
		LEFT JOIN employee e ON e.id_employee = i.employee_id
		LEFT JOIN (
			SELECT invoice_id, SUM(amount) AS pt_paid
			FROM payment_history
			WHERE payment_method_id != 14
			GROUP BY invoice_id
		) pt  ON pt.invoice_id  = i.id_invoice
		LEFT JOIN (
			SELECT invoice_id, SUM(amount::numeric) AS ins_paid
			FROM insurance_payment
			GROUP BY invoice_id
		) ins ON ins.invoice_id = i.id_invoice
		WHERE i.location_id = ?
		  AND i.number_invoice LIKE 'S%'
		  AND (
			COALESCE(pt.pt_paid, 0)
			+ COALESCE(ins.ins_paid, 0)
			+ COALESCE(i.gift_card_bal, 0)
		  ) > COALESCE(i.final_amount, 0)
		ORDER BY i.employee_id, i.date_create DESC
	`, locationID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	grouped := map[interface{}]*CreditGroup{}
	order := []interface{}{}

	for _, r := range rows {
		var key interface{}
		if r.EmployeeID != nil {
			key = *r.EmployeeID
		} else {
			key = nil
		}
		if _, ok := grouped[key]; !ok {
			name := "Unknown"
			if r.EmployeeFirstName != nil && r.EmployeeLastName != nil {
				name = *r.EmployeeFirstName + " " + *r.EmployeeLastName
			}
			grouped[key] = &CreditGroup{
				EmployeeID:   r.EmployeeID,
				EmployeeName: name,
				Invoices:     []CreditInvoice{},
			}
			order = append(order, key)
		}

		age := 0
		if r.AgeDays != nil {
			age = int(*r.AgeDays)
		}
		var dateStr *string
		if r.DateCreate != nil {
			s := r.DateCreate.Format(time.RFC3339)
			dateStr = &s
		}
		cell := "N/A"
		if r.CellNumber != nil && *r.CellNumber != "" {
			cell = *r.CellNumber
		}
		home := "N/A"
		if r.HomeNumber != nil && *r.HomeNumber != "" {
			home = *r.HomeNumber
		}
		gcBal := 0.0
		if r.GiftCardBal != nil {
			gcBal = *r.GiftCardBal
		}
		credit := r.PtPaid + r.InsPaid + gcBal - r.FinalAmount
		if credit < 0 {
			credit = 0
		}

		grouped[key].Invoices = append(grouped[key].Invoices, CreditInvoice{
			InvoiceID:     r.IDInvoice,
			NumberInvoice: r.NumberInvoice,
			PatientID:     r.PatientID,
			Date:          dateStr,
			Age:           age,
			LastName:      r.PatLastName,
			FirstName:     r.PatFirstName,
			CellNumber:    cell,
			HomeNumber:    home,
			Total:         r.TotalAmount,
			Paid:          r.PTBal,
			Due:           r.Due,
			Insurance:     r.InsBal,
			Credit:        credit,
		})
	}

	result := make([]CreditGroup, 0, len(order))
	for _, k := range order {
		result = append(result, *grouped[k])
	}
	return result, nil
}

// ─── CountSheets — list ───────────────────────────────────────────────────────

type CountSheetSummary struct {
	IDCountSheet int     `json:"id_count_sheet"`
	CreatedDate  string  `json:"created_date"`
	Status       bool    `json:"status"`
	Quantity     string  `json:"quantity"`
	Notes        *string `json:"notes"`
}

func (s *Service) GetCountSheets(username string, dateFrom, dateTo *time.Time) ([]CountSheetSummary, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	query := s.db.Model(&reportModel.ARCount{}).Where("location_id = ?", loc.IDLocation)
	if dateFrom != nil {
		query = query.Where("created_date >= ?", *dateFrom)
	}
	if dateTo != nil {
		query = query.Where("created_date <= ?", *dateTo)
	}

	var sheets []reportModel.ARCount
	if err := query.Order("created_date DESC").Find(&sheets).Error; err != nil {
		return nil, err
	}

	result := make([]CountSheetSummary, 0, len(sheets))
	for _, cs := range sheets {
		var found, missing int64
		s.db.Model(&reportModel.TempCountAR{}).
			Where("ar_count_id = ? AND location_id = ?", cs.IDARCount, loc.IDLocation).
			Count(&found)
		s.db.Model(&reportModel.MissingAR{}).
			Where("ar_count_id = ? AND location_id = ?", cs.IDARCount, loc.IDLocation).
			Count(&missing)

		result = append(result, CountSheetSummary{
			IDCountSheet: cs.IDARCount,
			CreatedDate:  cs.CreatedDate.Format(time.RFC3339),
			Status:       cs.Status,
			Quantity:     quantityStr(int(found), int(found+missing)),
			Notes:        cs.Notes,
		})
	}
	return result, nil
}

// ─── CountSheets — create ─────────────────────────────────────────────────────

func (s *Service) CreateCountSheet(username string, notes string) (*CountSheetSummary, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var existing reportModel.ARCount
	if s.db.Where("location_id = ? AND status = true", loc.IDLocation).First(&existing).Error == nil {
		return nil, &ActiveCountSheetError{ID: existing.IDARCount}
	}

	var openInvoices []invoices.Invoice
	if err := s.db.Where(
		"location_id = ? AND number_invoice LIKE 'S%' AND due > 0",
		loc.IDLocation,
	).Find(&openInvoices).Error; err != nil {
		return nil, err
	}
	if len(openInvoices) == 0 {
		return nil, errors.New("no open AR invoices found for this location")
	}

	now := time.Now().UTC()
	cs := reportModel.ARCount{
		LocationID:        loc.IDLocation,
		Status:            true,
		PrepByDate:        now,
		PrepByEmployeeID:  emp.IDEmployee,
		CreatedDate:       now,
		CreatedEmployeeID: emp.IDEmployee,
		UpdatedDate:       now,
		UpdatedEmployeeID: emp.IDEmployee,
		Quantity:          len(openInvoices),
		Notes:             strPtr(notes),
	}
	if notes == "" {
		cs.Notes = nil
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&cs).Error; err != nil {
			return err
		}
		for _, inv := range openInvoices {
			m := reportModel.MissingAR{
				ARCountID:    cs.IDARCount,
				InvoiceID:    inv.IDInvoice,
				LocationID:   loc.IDLocation,
				ReportedDate: now,
			}
			emptyNotes := ""
			m.Notes = &emptyNotes
			if err := tx.Create(&m).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	total := len(openInvoices)
	return &CountSheetSummary{
		IDCountSheet: cs.IDARCount,
		CreatedDate:  cs.CreatedDate.Format(time.RFC3339),
		Status:       cs.Status,
		Quantity:     quantityStr(0, total),
		Notes:        cs.Notes,
	}, nil
}

// ActiveCountSheetError — returned when an active sheet already exists
type ActiveCountSheetError struct {
	ID int
}

func (e *ActiveCountSheetError) Error() string {
	return "an active AR count sheet already exists for this location"
}

// ─── CountSheets — delete ─────────────────────────────────────────────────────

func (s *Service) DeleteCountSheet(username string, idCountSheet int) error {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}

	var cs reportModel.ARCount
	if err := s.db.Where("id_ar_count = ? AND location_id = ?", idCountSheet, loc.IDLocation).
		First(&cs).Error; err != nil {
		return errors.New("count sheet not found or does not belong to your location")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		tx.Where("ar_count_id = ?", idCountSheet).Delete(&reportModel.TempCountAR{})
		tx.Where("ar_count_id = ?", idCountSheet).Delete(&reportModel.MissingAR{})
		return tx.Delete(&cs).Error
	})
}

// ─── CountSheets — get one ────────────────────────────────────────────────────

type CountSheetDetail struct {
	IDCountSheet int     `json:"id_count_sheet"`
	Location     string  `json:"location"`
	Status       bool    `json:"status"`
	Quantity     string  `json:"quantity"`
	Notes        *string `json:"notes"`
	CreatedBy    *string `json:"created_by"`
	CreatedDate  string  `json:"created_date"`
}

func (s *Service) GetCountSheetInfo(username string, idCountSheet int) (*CountSheetDetail, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var cs reportModel.ARCount
	if err := s.db.Where("id_ar_count = ? AND location_id = ?", idCountSheet, loc.IDLocation).
		First(&cs).Error; err != nil {
		return nil, errors.New("count sheet not found or does not belong to your location")
	}

	var found, missing int64
	s.db.Model(&reportModel.TempCountAR{}).
		Where("ar_count_id = ? AND location_id = ?", idCountSheet, loc.IDLocation).Count(&found)
	s.db.Model(&reportModel.MissingAR{}).
		Where("ar_count_id = ? AND location_id = ?", idCountSheet, loc.IDLocation).Count(&missing)

	var createdBy *string
	var creator empModel.Employee
	if s.db.First(&creator, cs.CreatedEmployeeID).Error == nil {
		name := creator.FirstName + " " + creator.LastName
		createdBy = &name
	}

	return &CountSheetDetail{
		IDCountSheet: cs.IDARCount,
		Location:     loc.FullName,
		Status:       cs.Status,
		Quantity:     quantityStr(int(found), int(found+missing)),
		Notes:        cs.Notes,
		CreatedBy:    createdBy,
		CreatedDate:  cs.CreatedDate.Format(time.RFC3339),
	}, nil
}

// ─── CountSheets — update notes ───────────────────────────────────────────────

func (s *Service) UpdateCountSheetNotes(username string, idCountSheet int, notes string) error {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}

	var cs reportModel.ARCount
	if err := s.db.First(&cs, idCountSheet).Error; err != nil {
		return errors.New("count sheet not found")
	}

	cs.Notes = &notes
	cs.UpdatedEmployeeID = emp.IDEmployee
	cs.UpdatedDate = time.Now().UTC()
	return s.db.Save(&cs).Error
}

// ─── CountSheets — close ──────────────────────────────────────────────────────

func (s *Service) CloseCountSheet(username string, idCountSheet int) error {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}

	var cs reportModel.ARCount
	if err := s.db.Where("id_ar_count = ? AND location_id = ?", idCountSheet, loc.IDLocation).
		First(&cs).Error; err != nil {
		return errors.New("count sheet not found or does not belong to your location")
	}
	if !cs.Status {
		return errors.New("count sheet is already closed")
	}

	cs.Status = false
	cs.UpdatedEmployeeID = emp.IDEmployee
	cs.UpdatedDate = time.Now().UTC()
	return s.db.Save(&cs).Error
}

// ─── CountSheets — get items ──────────────────────────────────────────────────

type InvoiceItem struct {
	InvoiceID     int64   `json:"invoice_id"`
	NumberInvoice string  `json:"number_invoice"`
	PatientName   *string `json:"patient_name"`
	Date          *string `json:"date"`
	Due           float64 `json:"due"`
}

type CountSheetItems struct {
	IDCountSheet int           `json:"id_count_sheet"`
	Location     string        `json:"location"`
	Status       bool          `json:"status"`
	Quantity     string        `json:"quantity"`
	CountedItems []InvoiceItem `json:"counted_items"`
	MissingItems []InvoiceItem `json:"missing_items"`
}

func (s *Service) GetCountSheetItems(username string, idCountSheet int) (*CountSheetItems, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var cs reportModel.ARCount
	if err := s.db.Where("id_ar_count = ? AND location_id = ?", idCountSheet, loc.IDLocation).
		First(&cs).Error; err != nil {
		return nil, errors.New("count sheet not found or does not belong to your location")
	}

	var counted []reportModel.TempCountAR
	s.db.Where("ar_count_id = ? AND location_id = ?", idCountSheet, loc.IDLocation).Find(&counted)

	var missing []reportModel.MissingAR
	s.db.Where("ar_count_id = ? AND location_id = ?", idCountSheet, loc.IDLocation).Find(&missing)

	countedList := s.invoiceItemsForIDs(counted)
	missingList := s.invoiceItemsForMissing(missing)

	total := len(countedList) + len(missingList)
	return &CountSheetItems{
		IDCountSheet: cs.IDARCount,
		Location:     loc.FullName,
		Status:       cs.Status,
		Quantity:     quantityStr(len(countedList), total),
		CountedItems: countedList,
		MissingItems: missingList,
	}, nil
}

// ─── CountSheets — add invoice ────────────────────────────────────────────────

type AddInvoiceResult struct {
	CountedItems []InvoiceItem `json:"counted_items"`
	MissingItems []InvoiceItem `json:"missing_items"`
}

func (s *Service) AddInvoiceToCountSheet(username string, idCountSheet int, invoiceInput string) (*AddInvoiceResult, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var cs reportModel.ARCount
	if err := s.db.Where("id_ar_count = ? AND location_id = ?", idCountSheet, loc.IDLocation).
		First(&cs).Error; err != nil {
		return nil, errors.New("count sheet not found or does not belong to your location")
	}
	if !cs.Status {
		return nil, errors.New("count sheet is closed")
	}

	inv := s.findInvoiceByInput(invoiceInput, loc.IDLocation)
	if inv == nil {
		return nil, errors.New("invoice not found in this location")
	}

	var alreadyCounted reportModel.TempCountAR
	if s.db.Where("ar_count_id = ? AND invoice_id = ? AND location_id = ?",
		idCountSheet, inv.IDInvoice, loc.IDLocation).First(&alreadyCounted).Error == nil {
		return nil, errors.New("this invoice has already been counted in this sheet")
	}

	var missingEntry reportModel.MissingAR
	if err := s.db.Where("ar_count_id = ? AND invoice_id = ? AND location_id = ?",
		idCountSheet, inv.IDInvoice, loc.IDLocation).First(&missingEntry).Error; err != nil {
		return nil, errors.New("invoice not found in the outstanding list for this count sheet")
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		temp := reportModel.TempCountAR{
			InvoiceID:  inv.IDInvoice,
			LocationID: loc.IDLocation,
			ARCountID:  idCountSheet,
			CountDate:  time.Now().UTC(),
		}
		if err := tx.Create(&temp).Error; err != nil {
			return err
		}
		return tx.Delete(&missingEntry).Error
	}); err != nil {
		return nil, err
	}

	var counted []reportModel.TempCountAR
	s.db.Where("ar_count_id = ? AND location_id = ?", idCountSheet, loc.IDLocation).Find(&counted)
	var missing []reportModel.MissingAR
	s.db.Where("ar_count_id = ? AND location_id = ?", idCountSheet, loc.IDLocation).Find(&missing)

	return &AddInvoiceResult{
		CountedItems: s.invoiceItemsForIDs(counted),
		MissingItems: s.invoiceItemsForMissing(missing),
	}, nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func (s *Service) findInvoiceByInput(input string, locationID int) *invoices.Invoice {
	input = strings.TrimSpace(input)

	var inv invoices.Invoice
	// Exact match
	if s.db.Where("number_invoice = ? AND location_id = ?", input, locationID).First(&inv).Error == nil {
		return &inv
	}
	// Digit-only: try suffix match and 'S' prefix
	if isDigits(input) {
		if s.db.Where("location_id = ? AND number_invoice ILIKE ?", locationID, "%"+input).
			First(&inv).Error == nil {
			return &inv
		}
		if s.db.Where("number_invoice = ? AND location_id = ?", "S"+input, locationID).
			First(&inv).Error == nil {
			return &inv
		}
	}
	// Case-insensitive contains fallback
	if s.db.Where("location_id = ? AND number_invoice ILIKE ?", locationID, "%"+input+"%").
		First(&inv).Error == nil {
		return &inv
	}
	return nil
}

func (s *Service) invoiceItemsForIDs(rows []reportModel.TempCountAR) []InvoiceItem {
	result := []InvoiceItem{}
	for _, r := range rows {
		var inv invoices.Invoice
		if s.db.First(&inv, r.InvoiceID).Error != nil {
			continue
		}
		result = append(result, s.invoiceToItem(&inv))
	}
	return result
}

func (s *Service) invoiceItemsForMissing(rows []reportModel.MissingAR) []InvoiceItem {
	result := []InvoiceItem{}
	for _, r := range rows {
		var inv invoices.Invoice
		if s.db.First(&inv, r.InvoiceID).Error != nil {
			continue
		}
		result = append(result, s.invoiceToItem(&inv))
	}
	return result
}

func (s *Service) invoiceToItem(inv *invoices.Invoice) InvoiceItem {
	item := InvoiceItem{
		InvoiceID:     inv.IDInvoice,
		NumberInvoice: inv.NumberInvoice,
		Due:           inv.Due,
	}
	d := inv.DateCreate.Format(time.RFC3339)
	item.Date = &d

	var pat patModel.Patient
	if s.db.First(&pat, inv.PatientID).Error == nil {
		name := pat.FirstName + " " + pat.LastName
		item.PatientName = &name
	}
	return item
}

func quantityStr(found, total int) string {
	return fmt.Sprintf("%d/%d", found, total)
}

func isDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func strPtr(s string) *string { return &s }

