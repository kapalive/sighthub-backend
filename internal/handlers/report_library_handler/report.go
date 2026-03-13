package report_library_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"

	"sighthub-backend/internal/middleware"
	svc "sighthub-backend/internal/services/report_library_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct {
	svc *svc.Service
	db  *gorm.DB
}

func New(s *svc.Service, db *gorm.DB) *Handler { return &Handler{svc: s, db: db} }

// ─── helpers ────────────────────────────────────────────────────────────────

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func permittedIDs(r *http.Request, db *gorm.DB) []int {
	username := pkgAuth.UsernameFromContext(r.Context())
	if username == "" {
		return nil
	}
	return middleware.GetPermittedLocationIDs(db, username, 11)
}

func parseDate(s, def string) string {
	if s == "" {
		return def
	}
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return def
	}
	return s
}

func today() string {
	return time.Now().Format("2006-01-02")
}

// ─── GET /all_reports ───────────────────────────────────────────────────────

func (h *Handler) AllReports(w http.ResponseWriter, r *http.Request) {
	reportList := map[string][]map[string]string{
		"Sales": {
			{"label": "Invoice Summary", "path": "/invoice-summary"},
			{"label": "Invoice Classification", "path": "/invoice-classification"},
			{"label": "Vendor/Brand Margin Report", "path": "/vendor-brand-margin-report"},
			{"label": "Sales by Location", "path": "/sales-by-location"},
			{"label": "Sales by Frame", "path": "/sales-by-frame"},
			{"label": "Sales Average", "path": "/sales-average"},
			{"label": "Sales Breakdown by Product", "path": "/sales-breakdown-by-product"},
			{"label": "Sales by Employee - Detailed (Fully)", "path": "/sales-by-emp-detailed"},
			{"label": "Gift Card Balance", "path": "/gift-card-balance"},
			{"label": "Gift Card Activities", "path": "/gift-card-activities"},
			{"label": "SMS Purchases", "path": "/sms-purchases"},
		},
		"Inventory": {
			{"label": "List of Orders Placed", "path": "/list-of-orders-placed"},
			{"label": "List of Receipts", "path": "/list-of-receipts"},
			{"label": "Receipt by Brand", "path": "/receipt-by-brand"},
			{"label": "Missing Inventory", "path": "/missing-inventory"},
			{"label": "Inventory Work Flow", "path": "/inventory-work-flow"},
			{"label": "Inventory Analysis", "path": "/inventory-analysis"},
			{"label": "WOS Lens Order", "path": "/wos-lens-order"},
		},
		"Audit Logs": {
			{"label": "Audit Logins", "path": "/audit-logins"},
			{"label": "Audit Invoices", "path": "/audit-invoices"},
			{"label": "Audit Invoice Lines", "path": "/audit-invoice-lines"},
			{"label": "Audit Doctor Exams", "path": "/audit-dr-exams"},
			{"label": "Audit Reports", "path": "/audit-reports"},
			{"label": "Audit Files", "path": "/audit-files"},
			{"label": "Email Log", "path": "/email-log"},
			{"label": "SMS Log", "path": "/sms-log"},
			{"label": "Fax Log", "path": "/fax-log"},
			{"label": "Phone Log", "path": "/phone-log"},
			{"label": "Patient Communication Log", "path": "/patient-communication-log"},
		},
		"Performance": {
			{"label": "Lab Margin", "path": "/lab-margin"},
			{"label": "Discounts", "path": "/discounts"},
			{"label": "Inventory Turns (by Brand)", "path": "/inventory-turns"},
			{"label": "Score Card", "path": "/score-card"},
			{"label": "Elite Report", "path": "/elite-report"},
			{"label": "Appointment Search", "path": "/appointment-search"},
			{"label": "Appointment by Type", "path": "/appointment-by-type"},
			{"label": "Appointment by Insurance", "path": "/appointment-by-insurance"},
			{"label": "Invoice Time Analysis", "path": "/invoice-time-analysis"},
		},
		"Doctor Reports": {
			{"label": "Revenue by Doctor", "path": "/revenue-by-doctor"},
			{"label": "Professional Fees by Doctor/Payments Received", "path": "/prof-fees-by-dr"},
			{"label": "Appointment Stats", "path": "/appointment-stats"},
			{"label": "Sales (Doctor Location) - Exams", "path": "/sales-doctor-location-exams"},
			{"label": "Appointment Sales", "path": "/appointment-sales"},
			{"label": "Referral Source", "path": "/referral-source"},
		},
		"Insurance": {
			{"label": "Insurance Statistics", "path": "/insurance-statistics"},
		},
		"Marketing": {
			{"label": "Live Survey Results", "path": "/live-survey-results"},
			{"label": "Mailing List", "path": "/mailing-list"},
			{"label": "List of Birthdays", "path": "/list-of-birthdays"},
			{"label": "List of All Patient/Customers", "path": "/list-of-patients-customers"},
			{"label": "Nonprofit Report", "path": "/nonprofit-report"},
		},
		"Accounting": {
			{"label": "Monthly Summary", "path": "/monthly-summary"},
			{"label": "Accounts Receivable Aging Report", "path": "/accounts-receivable-aging-report"},
			{"label": "Accounts Receivable as of a Specific Date", "path": "/accounts-receivable-specific-date"},
			{"label": "Accounts Receivable Analysis", "path": "/accounts-receivable-analysis"},
			{"label": "List of Transfer Credit", "path": "/list-of-transfer-credit"},
			{"label": "Credit Card Payments", "path": "/credit-card-payments"},
			{"label": "Payment Summary", "path": "/payment-summary"},
			{"label": "Payment Details", "path": "/payment-details"},
			{"label": "Payment Categories", "path": "/payment-categories"},
			{"label": "Payment Overview", "path": "/payment-overview"},
			{"label": "Payment Report", "path": "/payment-report"},
			{"label": "Accounts Payable Due By Month", "path": "/accounts-payable-due-by-month"},
		},
	}
	jsonOK(w, reportList)
}

// ─── GET /sales/invoice_summary ─────────────────────────────────────────────

func (h *Handler) InvoiceSummary(w http.ResponseWriter, r *http.Request) {
	locIDs := permittedIDs(r, h.db)
	if len(locIDs) == 0 {
		jsonError(w, "no permitted locations", 403)
		return
	}

	t := today()
	startDate := parseDate(r.URL.Query().Get("start_date"), t)
	endDate := parseDate(r.URL.Query().Get("end_date"), t)

	data, totals, err := h.svc.InvoiceSummary(locIDs, startDate, endDate)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	jsonOK(w, map[string]interface{}{"data": data, "totals": totals})
}

// ─── GET /sales/invoice_classification ──────────────────────────────────────

func (h *Handler) InvoiceClassification(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	locID, _ := strconv.Atoi(q.Get("location_id"))
	if locID == 0 {
		jsonError(w, "location_id is required", 400)
		return
	}

	t := today()
	startDate := parseDate(q.Get("start_date"), t)
	endDate := parseDate(q.Get("end_date"), t)
	classType := q.Get("classification_type")

	data, err := h.svc.InvoiceClassification(locID, startDate, endDate, classType)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	var totalCount, totalSales float64
	for _, d := range data {
		totalCount += d.Count
		totalSales += d.TotalSales
	}
	avgSale := 0.0
	if totalCount > 0 {
		avgSale = totalSales / totalCount
	}

	jsonOK(w, map[string]interface{}{
		"data": data,
		"totals": map[string]interface{}{
			"total_count": totalCount,
			"total_sales": totalSales,
			"avg_sale":    avgSale,
		},
	})
}

// ─── GET /sales/vendor_brand_margin_report ──────────────────────────────────

func (h *Handler) VendorBrandMarginReport(w http.ResponseWriter, r *http.Request) {
	locIDs := permittedIDs(r, h.db)
	if len(locIDs) == 0 {
		jsonError(w, "no permitted locations", 403)
		return
	}

	t := today()
	startDate := parseDate(r.URL.Query().Get("start_date"), t)
	endDate := parseDate(r.URL.Query().Get("end_date"), t)

	data, err := h.svc.VendorBrandMarginReport(locIDs, startDate, endDate)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]interface{}{"data": data})
}

// ─── GET /sales/sales_by_location ───────────────────────────────────────────

func (h *Handler) SalesByLocation(w http.ResponseWriter, r *http.Request) {
	locIDs := permittedIDs(r, h.db)
	if len(locIDs) == 0 {
		jsonError(w, "no permitted locations", 403)
		return
	}

	t := today()
	startDate := parseDate(r.URL.Query().Get("start_date"), t)
	endDate := parseDate(r.URL.Query().Get("end_date"), t)

	data, err := h.svc.SalesByLocation(locIDs, startDate, endDate)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	var totalIns, totalPt float64
	for _, d := range data {
		totalIns += d.InsuranceSales
		totalPt += d.PatientSales
	}

	jsonOK(w, map[string]interface{}{
		"data": data,
		"totals": map[string]interface{}{
			"total_insurance_sales": totalIns,
			"total_patient_sales":   totalPt,
		},
	})
}

// ─── GET /sales/sales_by_frame ──────────────────────────────────────────────

func (h *Handler) SalesByFrame(w http.ResponseWriter, r *http.Request) {
	locIDs := permittedIDs(r, h.db)
	if len(locIDs) == 0 {
		jsonError(w, "no permitted locations", 403)
		return
	}

	t := today()
	startDate := parseDate(r.URL.Query().Get("start_date"), t)
	endDate := parseDate(r.URL.Query().Get("end_date"), t)

	var locFilter *int
	if v := r.URL.Query().Get("location_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			locFilter = &id
		}
	}

	data, err := h.svc.SalesByFrame(locIDs, startDate, endDate, locFilter)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]interface{}{"data": data})
}

// ─── GET /sales/sales_average ───────────────────────────────────────────────

func (h *Handler) SalesAverage(w http.ResponseWriter, r *http.Request) {
	locIDs := permittedIDs(r, h.db)
	if len(locIDs) == 0 {
		jsonError(w, "no permitted locations", 403)
		return
	}

	t := today()
	startDate := parseDate(r.URL.Query().Get("start_date"), t)
	endDate := parseDate(r.URL.Query().Get("end_date"), t)

	data, err := h.svc.SalesAverage(locIDs, startDate, endDate)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]interface{}{"data": data})
}

// ─── GET /sales/sales_breakdown_by_product_type ─────────────────────────────

func (h *Handler) SalesBreakdownByProductType(w http.ResponseWriter, r *http.Request) {
	locIDs := permittedIDs(r, h.db)
	if len(locIDs) == 0 {
		jsonError(w, "no permitted locations", 403)
		return
	}

	t := today()
	startDate := parseDate(r.URL.Query().Get("start_date"), t)
	endDate := parseDate(r.URL.Query().Get("end_date"), t)

	data, err := h.svc.SalesBreakdownByProductType(locIDs, startDate, endDate)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	var totalCost, totalInvoice float64
	for _, d := range data {
		totalCost += d.Cost
		totalInvoice += d.InvoiceTotal
	}

	jsonOK(w, map[string]interface{}{
		"data": data,
		"totals": map[string]interface{}{
			"total_cost":           totalCost,
			"total_invoice_amount": totalInvoice,
		},
	})
}

// ─── GET /sales/sales_by_employee ───────────────────────────────────────────

func (h *Handler) SalesByEmployee(w http.ResponseWriter, r *http.Request) {
	locIDs := permittedIDs(r, h.db)
	if len(locIDs) == 0 {
		jsonError(w, "no permitted locations", 403)
		return
	}

	t := today()
	startDate := parseDate(r.URL.Query().Get("start_date"), t)
	endDate := parseDate(r.URL.Query().Get("end_date"), t)

	var empID *int
	if v := r.URL.Query().Get("employee_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			empID = &id
		}
	}

	data, err := h.svc.SalesByEmployee(locIDs, startDate, endDate, empID)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	var grossTotal, netTotal, ptTotal, insTotal, discTotal float64
	for _, d := range data {
		grossTotal += d.GrossSales
		netTotal += d.NetSales
		ptTotal += d.PtPay
		insTotal += d.InsPay
		discTotal += d.Discount
	}

	jsonOK(w, map[string]interface{}{
		"data": data,
		"total": map[string]interface{}{
			"gross_sales": grossTotal,
			"net_sales":   netTotal,
			"pt_pay":      ptTotal,
			"ins_pay":     insTotal,
			"discount":    discTotal,
		},
	})
}

// ─── GET /sales/gift_card_balance ───────────────────────────────────────────

func (h *Handler) GiftCardBalance(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GiftCardBalance()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

// ─── GET /sales/gift_card_details ───────────────────────────────────────────

func (h *Handler) GiftCardDetails(w http.ResponseWriter, r *http.Request) {
	cardCode := r.URL.Query().Get("card_code")
	if cardCode == "" {
		jsonError(w, "Gift card code is required", 400)
		return
	}

	data, err := h.svc.GiftCardDetailsInfo(cardCode)
	if err != nil {
		jsonError(w, err.Error(), 404)
		return
	}
	jsonOK(w, data)
}

// ─── GET /sales/gift_card_activity ──────────────────────────────────────────

func (h *Handler) GiftCardActivity(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	now := time.Now().UTC()
	endDate := now
	startDate := now.Add(-30 * 24 * time.Hour)

	if v := q.Get("end_date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			endDate = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	}
	if v := q.Get("start_date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			startDate = t
		}
	}

	var locID *int
	if v := q.Get("location_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			locID = &id
		}
	}

	data, err := h.svc.GiftCardActivity(locID, startDate, endDate)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

// ─── GET /first_questionnaire/referral ──────────────────────────────────────

func (h *Handler) QuestionnaireReferral(w http.ResponseWriter, r *http.Request) {
	locID, startDT, endDT, err := h.parseDateRangeAndLocation(r)
	if err != nil {
		jsonError(w, err.Error(), 404)
		return
	}

	data, err := h.svc.QuestionnaireReferral(locID, startDT, endDT)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

// ─── GET /first_questionnaire/reasons ───────────────────────────────────────

func (h *Handler) QuestionnaireReasons(w http.ResponseWriter, r *http.Request) {
	locID, startDT, endDT, err := h.parseDateRangeAndLocation(r)
	if err != nil {
		jsonError(w, err.Error(), 404)
		return
	}

	data, err := h.svc.QuestionnaireReasons(locID, startDT, endDT)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

// ─── helper: parse date range + location for questionnaire endpoints ────────

func (h *Handler) parseDateRangeAndLocation(r *http.Request) (int, time.Time, time.Time, error) {
	username := pkgAuth.UsernameFromContext(r.Context())

	// location_id: from query or from employee record
	q := r.URL.Query()
	locID, _ := strconv.Atoi(q.Get("location_id"))
	if locID == 0 {
		empLocID, err := h.svc.GetEmployeeLocationID(username)
		if err != nil {
			return 0, time.Time{}, time.Time{}, err
		}
		locID = empLocID
	}

	now := time.Now().UTC()
	endDT := now
	startDT := now.Add(-30 * 24 * time.Hour)

	if v := q.Get("end"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			endDT = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		} else if t, err := time.Parse(time.RFC3339, v); err == nil {
			endDT = t
		}
	}
	if v := q.Get("start"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			startDT = t
		} else if t, err := time.Parse(time.RFC3339, v); err == nil {
			startDT = t
		}
	}

	if startDT.After(endDT) {
		startDT, endDT = endDT, startDT
	}

	return locID, startDT, endDT, nil
}
