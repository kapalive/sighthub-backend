package sale_handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"sighthub-backend/internal/middleware"
	svc "sighthub-backend/internal/services/sale_service"
	pkgAuth "sighthub-backend/pkg/auth"
	"sighthub-backend/pkg/csvutil"
)

type Handler struct {
	svc *svc.Service
	db  *gorm.DB
}

func New(s *svc.Service, db *gorm.DB) *Handler { return &Handler{svc: s, db: db} }

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func username(r *http.Request) string {
	return pkgAuth.UsernameFromContext(r.Context())
}

func permittedIDs(r *http.Request, db *gorm.DB, permID int) []int {
	u := username(r)
	if u == "" {
		return nil
	}
	return middleware.GetPermittedLocationIDs(db, u, permID)
}

func isAllToken(s string) bool {
	v := strings.TrimSpace(strings.ToLower(s))
	return v == "all" || v == "*"
}

func normStr(s string) *string {
	v := strings.TrimSpace(s)
	if v == "" {
		return nil
	}
	return &v
}

// ── GET /stores ─────────────────────────────────────────────────────────────

func (h *Handler) GetLocations(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	data, err := h.svc.GetLocations(username)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

// ── GET /vendors ────────────────────────────────────────────────────────────

func (h *Handler) GetVendors(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetVendors()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

// ── GET /vendors/{vendor_id}/brands ─────────────────────────────────────────

func (h *Handler) GetVendorBrands(w http.ResponseWriter, r *http.Request) {
	vendorID, err := strconv.Atoi(mux.Vars(r)["vendor_id"])
	if err != nil {
		jsonError(w, "invalid vendor_id", 400)
		return
	}
	data, err := h.svc.GetVendorBrands(vendorID)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

// ── GET /employees ──────────────────────────────────────────────────────────

func (h *Handler) GetEmployees(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetEmployees()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

// ── GET /item ───────────────────────────────────────────────────────────────

func (h *Handler) GetItems(w http.ResponseWriter, r *http.Request) {
	u := username(r)
	empLocID, err := h.svc.GetEmployeeLocationID(u)
	if err != nil {
		jsonError(w, "Employee or location not found", 404)
		return
	}

	q := r.URL.Query()
	locationID := empLocID
	if v := q.Get("location_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			locationID = id
		}
	}

	now := time.Now()
	defaultStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	defaultEnd := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	dateStart := defaultStart
	dateEnd := defaultEnd

	dateStartStr := q.Get("date_start")
	if dateStartStr == "" {
		dateStartStr = q.Get("start_date")
	}
	dateEndStr := q.Get("date_end")
	if dateEndStr == "" {
		dateEndStr = q.Get("end_date")
	}

	if dateStartStr != "" {
		if t, err := time.Parse("2006-01-02", dateStartStr); err == nil {
			dateStart = t
		}
	}
	if dateEndStr != "" {
		if t, err := time.Parse("2006-01-02", dateEndStr); err == nil {
			dateEnd = t
		}
	}

	var employeeID *int
	if v := q.Get("employee_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			employeeID = &id
		}
	}

	// pb_key
	var pbKey *string
	pbKeyRaw := q.Get("pb_key")
	if pbKeyRaw != "" && !isAllToken(pbKeyRaw) {
		k := strings.TrimSpace(strings.ToLower(pbKeyRaw))
		if mapped, ok := svc.PB_KEY_MAP[k]; ok {
			pbKey = &mapped
		} else {
			trimmed := strings.TrimSpace(pbKeyRaw)
			pbKey = &trimmed
		}
	}

	invoiceContains := normStr(q.Get("invoice_contains"))

	var saleKeyFilter *string
	skf := q.Get("sale_key")
	if skf != "" && !isAllToken(skf) {
		saleKeyFilter = normStr(skf)
	}

	var saleKeyContains *string
	skc := q.Get("sale_key_contains")
	if skc != "" && !isAllToken(skc) {
		saleKeyContains = normStr(skc)
	}

	sunRaw := strings.ToLower(q.Get("sun_only"))
	sunOnly := sunRaw == "yes" || sunRaw == "true" || sunRaw == "1"
	totalBy := q.Get("total_by")

	var minBal, maxBal *float64
	if v := q.Get("minbal"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			minBal = &f
		}
	}
	if v := q.Get("maxbal"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			maxBal = &f
		}
	}

	var vendorID, brandID *int
	if v := q.Get("vendor_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			vendorID = &id
		}
	}
	if v := q.Get("brand_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			brandID = &id
		}
	}

	data, err := h.svc.GetSaleItems(
		locationID, dateStart, dateEnd,
		employeeID,
		pbKey, invoiceContains,
		saleKeyFilter, saleKeyContains,
		sunOnly, totalBy,
		minBal, maxBal,
		vendorID, brandID,
	)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	output := q.Get("output")
	if output == "csv" {
		csv := csvutil.New()
		csv.Row("invoice_number", "date", "sale_key", "description", "item_type",
			"quantity", "price", "total_amount", "pt_balance", "insurance_balance",
			"final_amount", "due_amount", "pb_cost", "pb_selling_price")
		for _, d := range data {
			csv.Row(
				d.InvoiceNumber, strPtrVal(d.Date), strPtrVal(d.SaleKey),
				d.Description, strPtrVal(d.ItemType),
				d.Quantity, d.Price, d.TotalAmount, d.PtBalance, d.InsuranceBalance,
				d.FinalAmount, d.DueAmount, strPtrVal(d.PbCost), strPtrVal(d.PbSellingPrice),
			)
		}
		csv.ServeHTTP(w, "sale_items.csv")
		return
	}

	jsonOK(w, data)
}

// ── GET /yearly_comparison_by_rep ───────────────────────────────────────────

func (h *Handler) YearlyComparisonByRep(w http.ResponseWriter, r *http.Request) {
	u := username(r)
	empLocID, err := h.svc.GetEmployeeLocationID(u)
	if err != nil {
		jsonError(w, "Employee or location not found", 404)
		return
	}

	q := r.URL.Query()
	locationID := empLocID
	if v := q.Get("location_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			locationID = id
		}
	}

	startMonth := svc.ParseMonth(q.Get("start_month"), 1)
	endMonth := svc.ParseMonth(q.Get("end_month"), 12)
	yearStart, yearEnd := svc.ParseYearRange(q.Get("year"))

	active, inactive, err := h.svc.YearlyComparisonByRep(locationID, yearStart, yearEnd, startMonth, endMonth)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	jsonOK(w, map[string]interface{}{
		"active":   active,
		"inactive": inactive,
	})
}

// ── GET /professional_codes ─────────────────────────────────────────────────

func (h *Handler) ProfessionalCodes(w http.ResponseWriter, r *http.Request) {
	u := username(r)
	empLocID, err := h.svc.GetEmployeeLocationID(u)
	if err != nil {
		jsonError(w, "Employee or location not found", 404)
		return
	}

	q := r.URL.Query()
	locationID := empLocID
	if v := q.Get("location_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			locationID = id
		}
	}

	now := time.Now()
	startDate := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if v := q.Get("start_date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			startDate = t
		} else {
			jsonError(w, "Invalid date format. Use YYYY-MM-DD.", 400)
			return
		}
	}
	if v := q.Get("end_date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			endDate = t
		} else {
			jsonError(w, "Invalid date format. Use YYYY-MM-DD.", 400)
			return
		}
	}

	items, summary, err := h.svc.ProfessionalCodes(locationID, startDate, endDate)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	jsonOK(w, map[string]interface{}{
		"items":   items,
		"summary": summary,
	})
}

// ── GET /yearly_comparison_by_brand ─────────────────────────────────────────

func (h *Handler) YearlyComparisonByBrand(w http.ResponseWriter, r *http.Request) {
	u := username(r)
	empLocID, err := h.svc.GetEmployeeLocationID(u)
	if err != nil {
		jsonError(w, "Employee or location not found", 404)
		return
	}

	q := r.URL.Query()

	var locationID *int
	locStr := q.Get("location_id")
	if locStr == "" {
		locationID = &empLocID
	} else if !isAllToken(locStr) {
		if id, err := strconv.Atoi(locStr); err == nil {
			locationID = &id
		} else {
			jsonError(w, "Invalid parameter value", 400)
			return
		}
	}

	var vendorID *int
	if v := q.Get("vendor"); v != "" && !isAllToken(v) {
		if id, err := strconv.Atoi(v); err == nil {
			vendorID = &id
		} else {
			jsonError(w, "Invalid parameter value", 400)
			return
		}
	}

	var brandID *int
	if v := q.Get("brand"); v != "" && !isAllToken(v) {
		if id, err := strconv.Atoi(v); err == nil {
			brandID = &id
		} else {
			jsonError(w, "Invalid parameter value", 400)
			return
		}
	}

	yearStart, yearEnd := svc.ParseYearRange(q.Get("year"))

	data, err := h.svc.YearlyComparisonByBrand(locationID, vendorID, brandID, yearStart, yearEnd)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	jsonOK(w, map[string]interface{}{"brands": data})
}

// ── GET /insurance ──────────────────────────────────────────────────────────

func (h *Handler) GetInsuranceCompanies(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetInsuranceCompanies()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

// ── GET /insurance_report ───────────────────────────────────────────────────

func (h *Handler) InsuranceReport(w http.ResponseWriter, r *http.Request) {
	u := username(r)
	empLocID, err := h.svc.GetEmployeeLocationID(u)
	if err != nil {
		jsonError(w, "Employee or location not found", 404)
		return
	}

	q := r.URL.Query()
	now := time.Now()

	var locationID *int
	locStr := q.Get("location_id")
	if locStr == "" {
		locationID = &empLocID
	} else if !isAllToken(locStr) {
		if id, err := strconv.Atoi(locStr); err == nil {
			locationID = &id
		} else {
			jsonError(w, "Invalid location_id", 400)
			return
		}
	}

	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endDate := startDate

	if v := q.Get("start"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			startDate = t
		} else {
			jsonError(w, "Invalid date format. Use YYYY-MM-DD.", 400)
			return
		}
	}
	if v := q.Get("end"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			endDate = t
		} else {
			jsonError(w, "Invalid date format. Use YYYY-MM-DD.", 400)
			return
		}
	}

	var insuranceID *int
	if v := q.Get("insurance_id"); v != "" && !isAllToken(v) {
		if id, err := strconv.Atoi(v); err == nil {
			insuranceID = &id
		} else {
			jsonError(w, "Invalid insurance_id. Must be an integer or 'ALL'.", 400)
			return
		}
	}

	data, err := h.svc.InsuranceReport(locationID, startDate, endDate, insuranceID)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	var totalFinal float64
	for _, d := range data {
		v, _ := strconv.ParseFloat(d.FinalAmount, 64)
		totalFinal += v
	}

	jsonOK(w, map[string]interface{}{
		"invoices": data,
		"summary": map[string]interface{}{
			"total_invoices":     len(data),
			"total_final_amount": fmt.Sprintf("%.2f", totalFinal),
		},
	})
}

// ── GET /commission ─────────────────────────────────────────────────────────

func (h *Handler) Commission(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	now := time.Now()
	// default: current pay period (Sun-Sat ish)
	defaultStart := now.AddDate(0, 0, -(int(now.Weekday()) + 1))
	defaultEnd := now.AddDate(0, 0, (12 - int(now.Weekday())))

	startDate := defaultStart
	endDate := defaultEnd

	if v := q.Get("start_date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			startDate = t
		} else {
			jsonError(w, "Invalid date format. Use YYYY-MM-DD.", 400)
			return
		}
	}
	if v := q.Get("end_date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			endDate = t
		} else {
			jsonError(w, "Invalid date format. Use YYYY-MM-DD.", 400)
			return
		}
	}

	var locationID *int
	if v := q.Get("location_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			locationID = &id
		}
	}

	data, err := h.svc.CommissionReport(locationID, startDate, endDate)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	jsonOK(w, map[string]interface{}{"employees": data})
}

// ── GET /sales_report ───────────────────────────────────────────────────────

func (h *Handler) SalesReport(w http.ResponseWriter, r *http.Request) {
	u := username(r)
	empLocID, err := h.svc.GetEmployeeLocationID(u)
	if err != nil {
		jsonError(w, "location or employee not found.", 404)
		return
	}

	data, err := h.svc.SalesReport(empLocID)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

// ── GET /referral_report ────────────────────────────────────────────────────

func (h *Handler) ReferralReport(w http.ResponseWriter, r *http.Request) {
	allowedIDs := permittedIDs(r, h.db, 81)
	if len(allowedIDs) == 0 {
		jsonError(w, "No permitted locations", 403)
		return
	}

	q := r.URL.Query()

	// location_id filter
	if v := q.Get("location_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			jsonError(w, "Invalid location_id", 400)
			return
		}
		if !contains(allowedIDs, id) {
			jsonError(w, "Access denied to location", 403)
			return
		}
		allowedIDs = []int{id}
	}

	var startDate, endDate *time.Time
	if v := q.Get("start_date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			startDate = &t
		} else {
			jsonError(w, "Invalid date format. Use YYYY-MM-DD", 400)
			return
		}
	}
	if v := q.Get("end_date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			endDate = &t
		} else {
			jsonError(w, "Invalid date format. Use YYYY-MM-DD", 400)
			return
		}
	}

	var visitReasonID, referralSourceID *int
	if v := q.Get("visit_reasons_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			visitReasonID = &id
		}
	}
	if v := q.Get("referral_sources_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			referralSourceID = &id
		}
	}

	data, err := h.svc.ReferralReport(allowedIDs, startDate, endDate, visitReasonID, referralSourceID)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	var startStr, endStr *string
	if startDate != nil {
		s := startDate.Format("2006-01-02")
		startStr = &s
	}
	if endDate != nil {
		s := endDate.Format("2006-01-02")
		endStr = &s
	}

	jsonOK(w, map[string]interface{}{
		"location_ids": allowedIDs,
		"total":        len(data),
		"start_date":   startStr,
		"end_date":     endStr,
		"referrals":    data,
	})
}

// ── helpers ─────────────────────────────────────────────────────────────────

func contains(ids []int, id int) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

func strPtrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
