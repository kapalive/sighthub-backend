package daily_close_service

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	generalModel "sighthub-backend/internal/models/general"
	invoiceModel "sighthub-backend/internal/models/invoices"
	locModel "sighthub-backend/internal/models/location"
	patientModel "sighthub-backend/internal/models/patients"
	reportsModel "sighthub-backend/internal/models/reports"
	"sighthub-backend/pkg/activitylog"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ── helpers ─────────────────────────────────────────────────────────────────

func (s *Service) getEmployeeAndLocation(username string) (*empModel.Employee, *locModel.Location, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, nil, errors.New("employee not found")
	}
	var emp empModel.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, nil, errors.New("employee not found")
	}
	if emp.LocationID == nil {
		return nil, nil, errors.New("employee not found")
	}
	var loc locModel.Location
	if err := s.db.First(&loc, *emp.LocationID).Error; err != nil {
		return nil, nil, errors.New("employee not found")
	}
	return &emp, &loc, nil
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func errorStatus(err error) int {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "not found"):
		return 404
	case strings.Contains(msg, "required"), strings.Contains(msg, "invalid"),
		strings.Contains(msg, "already exists"):
		return 400
	default:
		return 500
	}
}

// ── denomination helpers ────────────────────────────────────────────────────

type denomInfo struct {
	Field string
	Cents int
}

var denominationsMap = map[string]denomInfo{
	"0.01":   {Field: "cent_1", Cents: 1},
	"0.05":   {Field: "cent_5", Cents: 5},
	"0.10":   {Field: "cent_10", Cents: 10},
	"0.25":   {Field: "cent_25", Cents: 25},
	"0.50":   {Field: "cent_50", Cents: 50},
	"1.00":   {Field: "dollar_1", Cents: 100},
	"2.00":   {Field: "dollar_2", Cents: 200},
	"5.00":   {Field: "dollar_5", Cents: 500},
	"10.00":  {Field: "dollar_10", Cents: 1000},
	"20.00":  {Field: "dollar_20", Cents: 2000},
	"50.00":  {Field: "dollar_50", Cents: 5000},
	"100.00": {Field: "dollar_100", Cents: 10000},
}

func parseCashCounts(cashCounts map[string]interface{}) (totalCents int, sheet reportsModel.CashCountSheet) {
	getInt := func(key string) int {
		v, ok := cashCounts[key]
		if !ok {
			return 0
		}
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		default:
			return 0
		}
	}

	for denomKey, info := range denominationsMap {
		count := getInt(denomKey)
		totalCents += count * info.Cents
		switch info.Field {
		case "cent_1":
			sheet.Cent1 = count
		case "cent_5":
			sheet.Cent5 = count
		case "cent_10":
			sheet.Cent10 = count
		case "cent_25":
			sheet.Cent25 = count
		case "cent_50":
			sheet.Cent50 = count
		case "dollar_1":
			sheet.Dollar1 = count
		case "dollar_2":
			sheet.Dollar2 = count
		case "dollar_5":
			sheet.Dollar5 = count
		case "dollar_10":
			sheet.Dollar10 = count
		case "dollar_20":
			sheet.Dollar20 = count
		case "dollar_50":
			sheet.Dollar50 = count
		case "dollar_100":
			sheet.Dollar100 = count
		}
	}
	return totalCents, sheet
}

// ── CreateDailyClose ────────────────────────────────────────────────────────

type PaymentInput struct {
	PaymentMethodID int64    `json:"payment_method_id"`
	Amount          float64  `json:"amount"`
	Note            *string  `json:"note"`
}

type CreateDailyCloseInput struct {
	Date       string                 `json:"date"`
	Payments   []PaymentInput         `json:"payments"`
	CashCounts map[string]interface{} `json:"cash_counts"`
}

func (s *Service) CreateDailyClose(username string, input CreateDailyCloseInput) (map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	if input.Date == "" {
		return nil, errors.New("date is required. Use format YYYY-MM-DD")
	}
	closeDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return nil, errors.New("invalid date format. Use YYYY-MM-DD")
	}

	var count int64
	s.db.Model(&reportsModel.DailyClosePayment{}).
		Where("location_id = ? AND date = ?", loc.IDLocation, closeDate).
		Count(&count)
	if count > 0 {
		return nil, errors.New("report for this date and location already exists. Use PUT to update")
	}

	return s.saveDailyClose(loc.IDLocation, closeDate, input.Payments, input.CashCounts, "create")
}

// ── UpdateDailyClose ────────────────────────────────────────────────────────

func (s *Service) UpdateDailyClose(username string, input CreateDailyCloseInput) (map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	if input.Date == "" {
		return nil, errors.New("date is required. Use format YYYY-MM-DD")
	}
	closeDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return nil, errors.New("invalid date format. Use YYYY-MM-DD")
	}

	// Delete existing
	var existing []reportsModel.DailyClosePayment
	s.db.Where("location_id = ? AND date = ?", loc.IDLocation, closeDate).Find(&existing)
	for _, rep := range existing {
		s.db.Where("daily_close_payment_id = ?", rep.IDDailyClosePayment).Delete(&reportsModel.CashCountSheet{})
		s.db.Delete(&rep)
	}

	return s.saveDailyClose(loc.IDLocation, closeDate, input.Payments, input.CashCounts, "update")
}

func (s *Service) saveDailyClose(locationID int, closeDate time.Time, payments []PaymentInput, cashCounts map[string]interface{}, action string) (map[string]interface{}, error) {
	if cashCounts == nil {
		cashCounts = map[string]interface{}{}
	}

	totalCents, cashSheet := parseCashCounts(cashCounts)
	totalCashAmount := float64(totalCents) / 100.0

	// If cash counts provided but no cash payment (method 2), add one
	hasCash := false
	for _, p := range payments {
		if p.PaymentMethodID == 2 {
			hasCash = true
			break
		}
	}
	if totalCents > 0 && !hasCash {
		payments = append(payments, PaymentInput{
			PaymentMethodID: 2,
			Amount:          totalCashAmount,
		})
	}

	var createdPayments []map[string]interface{}
	var cashPaymentID int64

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, p := range payments {
			amount := p.Amount
			if p.PaymentMethodID == 2 {
				amount = totalCashAmount
			}

			dcp := reportsModel.DailyClosePayment{
				PaymentMethodID: p.PaymentMethodID,
				Date:            closeDate,
				Amount:          amount,
				LocationID:      int64(locationID),
				Note:            p.Note,
			}
			if err := tx.Create(&dcp).Error; err != nil {
				return err
			}

			if p.PaymentMethodID == 2 {
				cashPaymentID = dcp.IDDailyClosePayment
			}

			createdPayments = append(createdPayments, map[string]interface{}{
				"id_daily_close_payment": dcp.IDDailyClosePayment,
				"payment_method_id":     dcp.PaymentMethodID,
				"date":                  dcp.Date.Format("2006-01-02"),
				"amount":                dcp.Amount,
				"location_id":           dcp.LocationID,
				"note":                  dcp.Note,
			})
		}

		if cashPaymentID > 0 {
			cashSheet.DailyClosePaymentID = &cashPaymentID
			cashSheet.Date = closeDate
			if err := tx.Create(&cashSheet).Error; err != nil {
				return err
			}
		}

		activitylog.Log(tx, "daily_close", action,
			activitylog.WithDetails(map[string]interface{}{
				"date":        closeDate.Format("2006-01-02"),
				"location_id": locationID,
			}),
		)
		return nil
	})
	if err != nil {
		return nil, err
	}

	key := "created_payments"
	if action == "update" {
		key = "updated_payments"
	}

	return map[string]interface{}{
		"date": closeDate.Format("2006-01-02"),
		key:    createdPayments,
	}, nil
}

// ── GetDailyCloseSummary ────────────────────────────────────────────────────

func (s *Service) GetDailyCloseSummary(username, dateStr string) (map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	if dateStr == "" {
		return nil, errors.New("date query parameter is required. Use format YYYY-MM-DD")
	}
	closeDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, errors.New("invalid date format. Use YYYY-MM-DD")
	}

	var payments []reportsModel.DailyClosePayment
	s.db.Where("location_id = ? AND date = ?", loc.IDLocation, closeDate).Find(&payments)

	var result []map[string]interface{}
	for _, p := range payments {
		var pm generalModel.PaymentMethod
		methodName := ""
		if err := s.db.First(&pm, p.PaymentMethodID).Error; err == nil {
			methodName = pm.MethodName
		}

		entry := map[string]interface{}{
			"id_daily_close_payment": p.IDDailyClosePayment,
			"payment_method_id":     p.PaymentMethodID,
			"method_name":           methodName,
			"amount":                p.Amount,
		}

		if p.PaymentMethodID == 2 {
			var cs reportsModel.CashCountSheet
			if err := s.db.Where("daily_close_payment_id = ?", p.IDDailyClosePayment).First(&cs).Error; err == nil {
				entry["cash_counts"] = map[string]int{
					"0.01":   cs.Cent1,
					"0.05":   cs.Cent5,
					"0.10":   cs.Cent10,
					"0.25":   cs.Cent25,
					"0.50":   cs.Cent50,
					"1.00":   cs.Dollar1,
					"2.00":   cs.Dollar2,
					"5.00":   cs.Dollar5,
					"10.00":  cs.Dollar10,
					"20.00":  cs.Dollar20,
					"50.00":  cs.Dollar50,
					"100.00": cs.Dollar100,
				}
			}
		}

		result = append(result, entry)
	}

	return map[string]interface{}{
		"date":     closeDate.Format("2006-01-02"),
		"payments": result,
	}, nil
}

// ── GetPaymentMethods ───────────────────────────────────────────────────────

func (s *Service) GetPaymentMethods() ([]map[string]interface{}, error) {
	var methods []generalModel.PaymentMethod
	if err := s.db.Where("id_payment_method NOT IN ?", []int{1, 20}).Find(&methods).Error; err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	for _, m := range methods {
		result = append(result, m.ToMap())
	}
	return result, nil
}

// ── GetDailyCloseDetail ─────────────────────────────────────────────────────

func (s *Service) GetDailyCloseDetail(username, dateStr string) (map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	locationID := loc.IDLocation
	var currentDate time.Time
	if dateStr != "" {
		currentDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, errors.New("invalid date format. Use YYYY-MM-DD")
		}
	} else {
		currentDate = time.Now()
	}
	dateOnly := currentDate.Format("2006-01-02")

	var invoices []invoiceModel.Invoice
	s.db.Where("number_invoice LIKE ? AND patient_id IS NOT NULL AND location_id = ? AND DATE(created_at) = ?",
		"S%", locationID, dateOnly).Find(&invoices)

	if len(invoices) == 0 {
		return map[string]interface{}{
			"location_id":     locationID,
			"date":            dateOnly,
			"invoices_detail": []interface{}{},
			"summary": map[string]interface{}{
				"gross_revenue": 0.0,
				"adjust":        0.0,
				"intercompany":  0.0,
				"gift_card":     0.0,
				"sales_tax":     0.0,
				"net_revenue":   0.0,
				"taxable_sales": 0.0,
				"flash":         0.0,
			},
		}, nil
	}

	var invoiceIDs []int64
	for _, inv := range invoices {
		invoiceIDs = append(invoiceIDs, inv.IDInvoice)
	}

	// Taxable sums per invoice
	type taxRow struct {
		InvoiceID int64
		Total     float64
	}
	var taxRows []taxRow
	s.db.Model(&invoiceModel.InvoiceItemSale{}).
		Select("invoice_id, SUM(total) as total").
		Where("invoice_id IN ? AND taxable = true", invoiceIDs).
		Group("invoice_id").Scan(&taxRows)
	taxableSums := map[int64]float64{}
	for _, r := range taxRows {
		taxableSums[r.InvoiceID] = r.Total
	}

	// Patient payments (excluding gift card method 14)
	type paidRow struct {
		InvoiceID int64
		Total     float64
	}
	var ptRows []paidRow
	s.db.Model(&patientModel.PaymentHistory{}).
		Select("invoice_id, COALESCE(SUM(amount), 0) as total").
		Where("invoice_id IN ? AND (payment_method_id IS NULL OR payment_method_id != 14)", invoiceIDs).
		Group("invoice_id").Scan(&ptRows)
	ptPaidMap := map[int64]float64{}
	for _, r := range ptRows {
		ptPaidMap[r.InvoiceID] = r.Total
	}

	// Insurance payments
	var insRows []paidRow
	s.db.Raw(`SELECT invoice_id, COALESCE(SUM(amount::numeric), 0) as total FROM insurance_payment WHERE invoice_id IN ? GROUP BY invoice_id`, invoiceIDs).Scan(&insRows)
	insPaidMap := map[int64]float64{}
	for _, r := range insRows {
		insPaidMap[r.InvoiceID] = r.Total
	}

	var results []map[string]interface{}
	grossRevenue := 0.0
	adjust := 0.0
	giftCard := 0.0
	salesTax := 0.0
	taxableTotal := 0.0

	for _, inv := range invoices {
		ptSales := ptPaidMap[inv.IDInvoice]
		insSales := insPaidMap[inv.IDInvoice]
		total := ptSales + insSales
		taxable := taxableSums[inv.IDInvoice]
		tax := inv.TaxAmount
		balDue := inv.Due

		grossRevenue += inv.FinalAmount
		if inv.Discount != nil {
			adjust += *inv.Discount
		}
		if inv.GiftCardBal != nil {
			giftCard += *inv.GiftCardBal
		}
		salesTax += tax
		taxableTotal += taxable

		// Prep/Sell
		employeeName := ""
		doctorName := ""
		if inv.EmployeeID != nil {
			var emp empModel.Employee
			if s.db.First(&emp, *inv.EmployeeID).Error == nil {
				employeeName = emp.FirstName + " " + emp.LastName
			}
		}
		if inv.DoctorID != nil {
			var doc empModel.Employee
			if s.db.First(&doc, *inv.DoctorID).Error == nil {
				doctorName = doc.FirstName + " " + doc.LastName
			}
		}
		prepSell := strings.TrimSpace(employeeName + " / " + doctorName)

		// Customer
		customerStr := "No Patient"
		if inv.PatientID > 0 {
			var pat patientModel.Patient
			if s.db.First(&pat, inv.PatientID).Error == nil {
				custName := strings.TrimSpace(pat.FirstName + " " + pat.LastName)
				if pat.Email != nil && *pat.Email != "" {
					if custName != "" {
						customerStr = custName + " (email ✓)"
					} else {
						customerStr = *pat.Email + " (email ✓)"
					}
				} else if custName != "" {
					customerStr = custName
				} else {
					customerStr = "Unknown Customer"
				}
			}
		}

		results = append(results, map[string]interface{}{
			"invoice":   inv.NumberInvoice,
			"customer":  customerStr,
			"total":     total,
			"pt_sales":  ptSales,
			"ins_sales": insSales,
			"taxable":   taxable,
			"tax":       tax,
			"tax_1":     0.0,
			"tax_2":     0.0,
			"bal_due":   balDue,
			"prep_sell": prepSell,
		})
	}

	intercompany := 0.0
	netRevenue := grossRevenue - (adjust + intercompany + giftCard + salesTax)

	return map[string]interface{}{
		"location_id":     locationID,
		"date":            dateOnly,
		"invoices_detail": results,
		"summary": map[string]interface{}{
			"gross_revenue": grossRevenue,
			"adjust":        adjust,
			"intercompany":  intercompany,
			"gift_card":     giftCard,
			"sales_tax":     salesTax,
			"net_revenue":   netRevenue,
			"taxable_sales": taxableTotal,
			"flash":         grossRevenue,
		},
	}, nil
}

// ── GetInvoicesPayments ─────────────────────────────────────────────────────

func (s *Service) GetInvoicesPayments(username, dateStr string) (map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	locationID := loc.IDLocation
	var currentDate time.Time
	if dateStr != "" {
		currentDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, errors.New("invalid date format. Use YYYY-MM-DD")
		}
	} else {
		currentDate = time.Now()
	}
	dateOnly := currentDate.Format("2006-01-02")

	type paymentRow struct {
		InvoiceNumber string
		FirstName     string
		LastName      string
		Amount        float64
		MethodName    *string
		Due           float64
		Notified      *string
		DOB           *time.Time
		EmpFirstName  *string
		EmpLastName   *string
	}

	var rows []paymentRow
	s.db.Raw(`
		SELECT i.number_invoice as invoice_number,
		       p.first_name, p.last_name,
		       ph.amount,
		       pm.method_name,
		       i.due,
		       i.notified,
		       p.dob,
		       e.first_name as emp_first_name, e.last_name as emp_last_name
		FROM payment_history ph
		JOIN patient p ON ph.patient_id = p.id_patient
		JOIN invoice i ON ph.invoice_id = i.id_invoice
		LEFT JOIN payment_method pm ON ph.payment_method_id = pm.id_payment_method
		LEFT JOIN employee e ON ph.employee_id = e.id_employee
		WHERE i.location_id = ? AND DATE(ph.payment_timestamp) = ?
	`, locationID, dateOnly).Scan(&rows)

	var results []map[string]interface{}
	totalAmount := 0.0

	for _, r := range rows {
		name := strings.TrimSpace(r.FirstName + " " + r.LastName)
		paidBy := "Unknown"
		if r.MethodName != nil {
			paidBy = *r.MethodName
		}

		dlCC := "************"
		if r.Notified != nil && len(*r.Notified) >= 4 {
			dlCC = "************" + (*r.Notified)[len(*r.Notified)-4:]
		}

		dobExp := ""
		if r.DOB != nil {
			dobExp = r.DOB.Format("2006-01-02")
		}

		user := "Unknown"
		if r.EmpFirstName != nil && r.EmpLastName != nil {
			user = strings.TrimSpace(*r.EmpFirstName + " " + *r.EmpLastName)
		}

		totalAmount += r.Amount
		results = append(results, map[string]interface{}{
			"invoice":     r.InvoiceNumber,
			"name":        name,
			"amount":      r.Amount,
			"paid_by":     paidBy,
			"balance_due": r.Due,
			"dl_cc":       dlCC,
			"dob_exp":     dobExp,
			"user":        user,
		})
	}

	return map[string]interface{}{
		"location_id":     locationID,
		"date":            dateOnly,
		"invoices_detail": results,
		"summary": map[string]interface{}{
			"total_amount": totalAmount,
		},
	}, nil
}

// ── GetPaymentsSummary ──────────────────────────────────────────────────────

func (s *Service) GetPaymentsSummary(username, dateStr string) (map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	locationID := loc.IDLocation
	var currentDate time.Time
	if dateStr != "" {
		currentDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, errors.New("invalid date format. Use YYYY-MM-DD")
		}
	} else {
		currentDate = time.Now()
	}
	dateOnly := currentDate.Format("2006-01-02")

	type methodAmount struct {
		MethodName string
		Amount     float64
	}

	// 1. Computer payments (PaymentHistory)
	var computerRows []methodAmount
	s.db.Raw(`
		SELECT pm.method_name, SUM(ph.amount) as amount
		FROM payment_history ph
		JOIN payment_method pm ON ph.payment_method_id = pm.id_payment_method
		JOIN invoice i ON ph.invoice_id = i.id_invoice
		WHERE DATE(ph.payment_timestamp) = ? AND i.location_id = ?
		GROUP BY pm.method_name
	`, dateOnly, locationID).Scan(&computerRows)

	// 2. Count sheet payments (DailyClosePayment)
	var countSheetRows []methodAmount
	s.db.Raw(`
		SELECT pm.method_name, SUM(dcp.amount) as amount
		FROM daily_close_payment dcp
		JOIN payment_method pm ON dcp.payment_method_id = pm.id_payment_method
		WHERE dcp.date = ? AND dcp.location_id = ?
		GROUP BY pm.method_name
	`, dateOnly, locationID).Scan(&countSheetRows)

	// 3. Swiped (PaymentTransaction)
	var swipedRows []methodAmount
	s.db.Raw(`
		SELECT pt.payment_method as method_name, SUM(pt.amount::numeric) as amount
		FROM payment_transaction pt
		JOIN payment_history ph ON pt.id_payment_transaction = ph.payment_transaction_id
		JOIN invoice i ON ph.invoice_id = i.id_invoice
		WHERE DATE(pt.transaction_date) = ? AND i.location_id = ?
		GROUP BY pt.payment_method
	`, dateOnly, locationID).Scan(&swipedRows)

	summary := map[string]map[string]float64{}

	for _, r := range computerRows {
		summary[r.MethodName] = map[string]float64{
			"computer": r.Amount, "count_sheet": 0, "swiped": 0, "difference": 0,
		}
	}
	for _, r := range countSheetRows {
		if _, ok := summary[r.MethodName]; !ok {
			summary[r.MethodName] = map[string]float64{
				"computer": 0, "count_sheet": r.Amount, "swiped": 0, "difference": 0,
			}
		} else {
			summary[r.MethodName]["count_sheet"] = r.Amount
		}
	}
	for _, r := range swipedRows {
		if _, ok := summary[r.MethodName]; !ok {
			summary[r.MethodName] = map[string]float64{
				"computer": 0, "count_sheet": 0, "swiped": r.Amount, "difference": 0,
			}
		} else {
			summary[r.MethodName]["swiped"] = r.Amount
		}
	}

	// Calculate difference
	for _, data := range summary {
		data["difference"] = data["computer"] - (data["count_sheet"] + data["swiped"])
	}

	// Ensure all methods represented
	var allMethods []generalModel.PaymentMethod
	s.db.Find(&allMethods)
	for _, m := range allMethods {
		if _, ok := summary[m.MethodName]; !ok {
			summary[m.MethodName] = map[string]float64{
				"computer": 0, "count_sheet": 0, "swiped": 0, "difference": 0,
			}
		}
	}

	var result []map[string]interface{}
	totalComputer := 0.0
	totalCountSheet := 0.0
	totalSwiped := 0.0
	totalDiff := 0.0

	for method, data := range summary {
		result = append(result, map[string]interface{}{
			"Type":        method,
			"computer":    data["computer"],
			"count_sheet": data["count_sheet"],
			"swiped":      data["swiped"],
			"difference":  data["difference"],
		})
		totalComputer += data["computer"]
		totalCountSheet += data["count_sheet"]
		totalSwiped += data["swiped"]
		totalDiff += data["difference"]
	}

	result = append(result, map[string]interface{}{
		"Type":        "Total",
		"computer":    totalComputer,
		"count_sheet": totalCountSheet,
		"swiped":      totalSwiped,
		"difference":  totalDiff,
	})

	return map[string]interface{}{
		"location_id":      locationID,
		"date":             dateOnly,
		"payments_summary": result,
	}, nil
}

// ── GetTransferCreditSummary ────────────────────────────────────────────────

func (s *Service) GetTransferCreditSummary(username, dateStr string, invoiceID, patientID *int64) (map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	locationID := loc.IDLocation
	var currentDate time.Time
	if dateStr != "" {
		currentDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, errors.New("invalid date format. Use YYYY-MM-DD")
		}
	} else {
		currentDate = time.Now()
	}

	startDt := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, currentDate.Location())
	endDt := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 23, 59, 59, 999999999, currentDate.Location())

	// Positive = PaymentHistory with payment_method_id = 20
	posQuery := s.db.Model(&patientModel.PaymentHistory{}).
		Select("COALESCE(SUM(payment_history.amount), 0)").
		Joins("JOIN invoice ON payment_history.invoice_id = invoice.id_invoice").
		Where("payment_history.payment_method_id = 20 AND payment_history.payment_timestamp >= ? AND payment_history.payment_timestamp <= ? AND invoice.location_id = ?",
			startDt, endDt, locationID)
	if invoiceID != nil {
		posQuery = posQuery.Where("payment_history.invoice_id = ?", *invoiceID)
	}
	if patientID != nil {
		posQuery = posQuery.Where("payment_history.patient_id = ?", *patientID)
	}
	var positive float64
	posQuery.Scan(&positive)

	// Negative = TransferCredit.amount < 0
	negQuery := s.db.Model(&patientModel.TransferCredit{}).
		Select("COALESCE(SUM(transfer_credit.amount), 0)").
		Joins("JOIN invoice ON transfer_credit.invoice_id = invoice.id_invoice").
		Where("transfer_credit.amount < 0 AND transfer_credit.created_at >= ? AND transfer_credit.created_at <= ? AND invoice.location_id = ?",
			startDt, endDt, locationID)
	if invoiceID != nil {
		negQuery = negQuery.Where("transfer_credit.invoice_id = ?", *invoiceID)
	}
	if patientID != nil {
		negQuery = negQuery.Where("transfer_credit.patient_id = ?", *patientID)
	}
	var negative float64
	negQuery.Scan(&negative)

	return map[string]interface{}{
		"location_id": locationID,
		"date":        currentDate.Format("2006-01-02"),
		"summary": map[string]interface{}{
			"positive": positive,
			"negative": negative,
			"total":    positive + negative,
		},
	}, nil
}

// ── RenderDailyCloseReportHTML ──────────────────────────────────────────────

func (s *Service) RenderDailyCloseReportHTML(username, dateStr string) (map[string]interface{}, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	if dateStr == "" {
		return nil, errors.New("date query parameter is required")
	}
	currentDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, errors.New("invalid date format. Use YYYY-MM-DD")
	}

	// 1. Invoices detail
	invResp, err := s.GetDailyCloseDetail(username, dateStr)
	if err != nil {
		return nil, err
	}
	invoices, _ := invResp["invoices_detail"].([]map[string]interface{})
	summary, _ := invResp["summary"].(map[string]interface{})

	// 2. Payments
	pmResp, err := s.GetInvoicesPayments(username, dateStr)
	if err != nil {
		return nil, err
	}
	pmDetail, _ := pmResp["invoices_detail"].([]map[string]interface{})
	pmSummary, _ := pmResp["summary"].(map[string]interface{})

	var payments []map[string]interface{}
	for _, row := range pmDetail {
		payments = append(payments, map[string]interface{}{
			"invoice_number":  row["invoice"],
			"customer_name":   row["name"],
			"amount":          row["amount"],
			"payment_method":  row["paid_by"],
			"balance_due":     row["balance_due"],
			"dl_cc":           row["dl_cc"],
			"expiry":          row["dob_exp"],
			"user":            row["user"],
		})
	}

	// 3. Payments summary
	payResp, err := s.GetPaymentsSummary(username, dateStr)
	if err != nil {
		return nil, err
	}
	types, _ := payResp["payments_summary"].([]map[string]interface{})

	var paymentTypes []map[string]interface{}
	totalSummary := map[string]interface{}{}
	for _, row := range types {
		if row["Type"] == "Total" {
			totalSummary = row
			continue
		}
		diff, _ := row["difference"].(float64)
		paymentTypes = append(paymentTypes, map[string]interface{}{
			"type":            row["Type"],
			"computer":        row["computer"],
			"count_sheet":     row["count_sheet"],
			"computer_swiped": row["swiped"],
			"difference":      row["difference"],
			"highlight":       math.Abs(diff) > 0.009,
		})
	}

	// 4. Transfer credits
	creditResp, err := s.GetTransferCreditSummary(username, dateStr, nil, nil)
	if err != nil {
		return nil, err
	}
	creditSummary, _ := creditResp["summary"].(map[string]interface{})

	// Build invoices for template
	var templateInvoices []map[string]interface{}
	for _, inv := range invoices {
		templateInvoices = append(templateInvoices, map[string]interface{}{
			"number":        inv["invoice"],
			"customer_name": inv["customer"],
			"total":         inv["total"],
			"taxable":       inv["taxable"],
			"tax":           inv["tax"],
			"balance_due":   inv["bal_due"],
			"prep_sell":     inv["prep_sell"],
		})
	}

	grossRev := toFloat(summary["gross_revenue"])
	totalPayments := toFloat(pmSummary["total_amount"])

	data := map[string]interface{}{
		"store_name":       loc.FullName,
		"store_address":    ptrStr(loc.StreetAddress),
		"store_city_state": fmt.Sprintf("%s, %s %s", ptrStr(loc.City), ptrStr(loc.State), ptrStr(loc.PostalCode)),
		"store_phone":      ptrStr(loc.Phone),
		"store_code":       ptrStr(loc.ShortName),
		"report_date":      currentDate.Format("01/02/2006"),
		"page_number":      1,
		"total_pages":      1,
		"printed_by":       strings.ToUpper(emp.FirstName + " " + emp.LastName),
		"print_datetime":   strings.ToLower(time.Now().Format("01/02/2006 03:04:05 PM")),

		"invoices":      templateInvoices,
		"gross_revenue": grossRev,
		"inter_company": toFloat(summary["intercompany"]),
		"total_tax":     toFloat(summary["sales_tax"]),
		"net_revenue":   toFloat(summary["net_revenue"]),
		"taxable_sales": toFloat(summary["taxable_sales"]),

		"payments":       payments,
		"total_payments":  totalPayments,
		"change_in_ar":    grossRev - totalPayments,

		"payment_types":        paymentTypes,
		"total_computer":       toFloat(totalSummary["computer"]),
		"total_count_sheet":    toFloat(totalSummary["count_sheet"]),
		"total_computer_swiped": toFloat(totalSummary["swiped"]),
		"total_difference":     toFloat(totalSummary["difference"]),

		"transfer_credit_summary": creditSummary,
	}

	return data, nil
}

// ── RenderCountSheetHTML ────────────────────────────────────────────────────

func (s *Service) RenderCountSheetHTML(username, dateStr string) (map[string]interface{}, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	if dateStr == "" {
		return nil, errors.New("date query parameter is required")
	}
	currentDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, errors.New("invalid date format. Use YYYY-MM-DD")
	}

	summaryResp, err := s.GetDailyCloseSummary(username, dateStr)
	if err != nil {
		return nil, err
	}
	payments, _ := summaryResp["payments"].([]map[string]interface{})

	cashCounts := map[string]int{}
	totalCash := 0.0
	var electronicPayments []map[string]interface{}

	for _, p := range payments {
		pmID, _ := p["payment_method_id"].(int64)
		if pmID == 2 {
			if cc, ok := p["cash_counts"].(map[string]int); ok {
				cashCounts = cc
			}
			if amt, ok := p["amount"].(float64); ok {
				totalCash = amt
			}
		} else {
			electronicPayments = append(electronicPayments, map[string]interface{}{
				"name":      p["method_name"],
				"amount":    p["amount"],
				"highlight": false,
			})
		}
	}

	denomOrder := []string{"0.01", "0.05", "0.10", "0.25", "0.50", "1.00",
		"2.00", "5.00", "10.00", "20.00", "50.00", "100.00"}
	denomValues := map[string]float64{
		"0.01": 0.01, "0.05": 0.05, "0.10": 0.10, "0.25": 0.25,
		"0.50": 0.50, "1.00": 1.00, "2.00": 2.00, "5.00": 5.00,
		"10.00": 10.00, "20.00": 20.00, "50.00": 50.00, "100.00": 100.00,
	}
	var cashDenominations []map[string]interface{}
	for _, d := range denomOrder {
		cashDenominations = append(cashDenominations, map[string]interface{}{
			"value":     denomValues[d],
			"count":     cashCounts[d],
			"highlight": false,
		})
	}

	data := map[string]interface{}{
		"store_name":    loc.FullName,
		"store_address": ptrStr(loc.StreetAddress),
		"store_city":    ptrStr(loc.City),
		"store_state":   ptrStr(loc.State),
		"store_zip":     ptrStr(loc.PostalCode),
		"store_phone":   ptrStr(loc.Phone),
		"store_code":    ptrStr(loc.ShortName),
		"count_date":    currentDate.Format("01/02/2006"),
		"page_number":   1,
		"total_pages":   1,
		"printed_by":    strings.ToUpper(emp.FirstName + " " + emp.LastName),
		"printed_date":  strings.ToLower(time.Now().Format("01/02/2006 03:04:05 PM")),

		"cash_denominations": cashDenominations,
		"receipts_in_drawer": map[string]interface{}{"amount": 0.0, "highlight": false},
		"minus_opening_cash": map[string]interface{}{"amount": 0.0, "highlight": false},
		"total_cash":         totalCash,
		"electronic_payments": electronicPayments,
	}

	return data, nil
}

func toFloat(v interface{}) float64 {
	if v == nil {
		return 0.0
	}
	if f, ok := v.(float64); ok {
		return f
	}
	return 0.0
}
