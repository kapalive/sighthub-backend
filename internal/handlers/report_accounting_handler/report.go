package report_accounting_handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sighthub-backend/internal/middleware"
	svc "sighthub-backend/internal/services/report_accounting_service"
)

type Handler struct {
	svc *svc.Service
}

func New(s *svc.Service) *Handler { return &Handler{svc: s} }

// ─── helpers ─────────────────────────────────────────────────────────────────

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func jsonResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// validateLocation checks location_id param against permitted locations from StorePermission middleware.
func validateLocation(r *http.Request) (int, bool) {
	locStr := r.URL.Query().Get("location_id")
	if locStr == "" {
		return 0, false
	}
	locID, err := strconv.Atoi(locStr)
	if err != nil {
		return 0, false
	}
	permitted := middleware.PermittedLocationIDsFromCtx(r.Context())
	for _, id := range permitted {
		if id == locID {
			return locID, true
		}
	}
	return locID, false // not permitted
}

// resolveLocationIDs handles: empty → employee default, "all" → all permitted, specific id.
func resolveLocationIDs(r *http.Request) ([]int, int, string) {
	locStr := strings.TrimSpace(r.URL.Query().Get("location_id"))
	permitted := middleware.PermittedLocationIDsFromCtx(r.Context())
	permSet := map[int]struct{}{}
	for _, id := range permitted {
		permSet[id] = struct{}{}
	}

	if locStr == "" {
		emp := middleware.EmployeeFromCtx(r.Context())
		if emp == nil || emp.LocationID == nil {
			return nil, 0, ""
		}
		defLoc := int(*emp.LocationID)
		if _, ok := permSet[defLoc]; !ok {
			return nil, 0, "Permission denied for default location"
		}
		return []int{defLoc}, 0, ""
	}
	if strings.ToLower(locStr) == "all" {
		return permitted, 0, ""
	}
	lid, err := strconv.Atoi(locStr)
	if err != nil {
		return nil, http.StatusBadRequest, "Invalid location_id"
	}
	if _, ok := permSet[lid]; !ok {
		return nil, http.StatusForbidden, "Permission denied for this location"
	}
	return []int{lid}, 0, ""
}

func parseDate(s string) (time.Time, bool) {
	t, err := time.Parse("2006-01-02", s)
	return t, err == nil
}

// ─── 1. GET /monthly_summary ─────────────────────────────────────────────────

func (h *Handler) MonthlySummary(w http.ResponseWriter, r *http.Request) {
	locID, ok := validateLocation(r)
	if !ok {
		if locID == 0 {
			jsonError(w, "location_id is required", http.StatusBadRequest)
		} else {
			jsonError(w, "Permission denied for this location", http.StatusForbidden)
		}
		return
	}

	now := time.Now().UTC()
	month := int(now.Month())
	year := now.Year()
	if v := r.URL.Query().Get("month"); v != "" {
		if m, err := strconv.Atoi(v); err == nil {
			month = m
		}
	}
	if v := r.URL.Query().Get("year"); v != "" {
		if y, err := strconv.Atoi(v); err == nil {
			year = y
		}
	}

	result, err := h.svc.MonthlySummary(locID, month, year)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.URL.Query().Get("output") == "csv" {
		csv := svc.MonthlySummaryCSV(result)
		csv.ServeHTTP(w, fmt.Sprintf("monthly_summary_%d_%02d.csv", year, month))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── 2. GET /daily_detail ────────────────────────────────────────────────────

func (h *Handler) DailyDetail(w http.ResponseWriter, r *http.Request) {
	locID, ok := validateLocation(r)
	if !ok {
		if locID == 0 {
			jsonError(w, "location_id is required", http.StatusBadRequest)
		} else {
			jsonError(w, "Permission denied for this location", http.StatusForbidden)
		}
		return
	}

	dateStr := r.URL.Query().Get("date")
	targetDate, ok := parseDate(dateStr)
	if !ok {
		jsonError(w, "date is required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	result, err := h.svc.DailyDetail(locID, targetDate)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.URL.Query().Get("output") == "csv" {
		csv := svc.DailyDetailCSV(result)
		csv.ServeHTTP(w, fmt.Sprintf("daily_detail_%s.csv", dateStr))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── 3. GET /payment_summary ─────────────────────────────────────────────────

func (h *Handler) PaymentSummary(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	targetDate, ok := parseDate(dateStr)
	if !ok {
		jsonError(w, "date is required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	locIDs, errCode, errMsg := resolveLocationIDs(r)
	if errMsg != "" {
		if errCode == 0 {
			errCode = http.StatusForbidden
		}
		jsonError(w, errMsg, errCode)
		return
	}
	if len(locIDs) == 0 {
		jsonResponse(w, map[string]interface{}{"data": []interface{}{}, "grand_total": 0}, http.StatusOK)
		return
	}

	result, err := h.svc.PaymentSummary(locIDs, targetDate)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── 4. GET /payment_details ─────────────────────────────────────────────────

func (h *Handler) PaymentDetails(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	startDate, ok1 := parseDate(startStr)
	endDate, ok2 := parseDate(endStr)
	if !ok1 || !ok2 {
		jsonError(w, "start and end dates required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	locIDs, errCode, errMsg := resolveLocationIDs(r)
	if errMsg != "" {
		if errCode == 0 {
			errCode = http.StatusForbidden
		}
		jsonError(w, errMsg, errCode)
		return
	}
	if len(locIDs) == 0 {
		jsonResponse(w, map[string]interface{}{"data": []interface{}{}, "total_pt_bal": 0, "total_ins_bal": 0}, http.StatusOK)
		return
	}

	result, err := h.svc.PaymentDetails(locIDs, startDate, endDate)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── 5. GET /payment_categories ──────────────────────────────────────────────

func (h *Handler) PaymentCategories(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	startDate, ok1 := parseDate(startStr)
	endDate, ok2 := parseDate(endStr)
	if !ok1 || !ok2 {
		jsonError(w, "start and end dates required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	locIDs, errCode, errMsg := resolveLocationIDs(r)
	if errMsg != "" {
		if errCode == 0 {
			errCode = http.StatusForbidden
		}
		jsonError(w, errMsg, errCode)
		return
	}
	if len(locIDs) == 0 {
		jsonResponse(w, map[string]interface{}{
			"data": []interface{}{}, "total_pmt": 0,
			"total_insurance_payment": 0, "total_pt_pmt": 0,
		}, http.StatusOK)
		return
	}

	var paymentTypeID *int
	if v := r.URL.Query().Get("payment_type"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			paymentTypeID = &id
		}
	}

	insuranceFilter := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("insurance")))
	var insuranceCompanyID *int
	if insuranceFilter != "" && insuranceFilter != "all_payments" && insuranceFilter != "all_insurance" {
		if id, err := strconv.Atoi(insuranceFilter); err == nil {
			insuranceCompanyID = &id
		} else {
			jsonError(w, "Invalid insurance filter", http.StatusBadRequest)
			return
		}
	}

	result, err := h.svc.PaymentCategories(locIDs, startDate, endDate, paymentTypeID, insuranceFilter, insuranceCompanyID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── 6. GET /insurance_companies ─────────────────────────────────────────────

func (h *Handler) InsuranceCompanies(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.InsuranceCompanies()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── 7. GET /payment_types ───────────────────────────────────────────────────

func (h *Handler) PaymentTypes(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.PaymentTypes()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── 8. GET /ar_insurance_aging ──────────────────────────────────────────────

func (h *Handler) ARInsuranceAging(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	startDate, ok1 := parseDate(startStr)
	endDate, ok2 := parseDate(endStr)
	if !ok1 || !ok2 {
		jsonError(w, "start and end dates required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	locIDs, errCode, errMsg := resolveLocationIDs(r)
	if errMsg != "" {
		if errCode == 0 {
			errCode = http.StatusForbidden
		}
		jsonError(w, errMsg, errCode)
		return
	}
	if len(locIDs) == 0 {
		jsonResponse(w, map[string]interface{}{"data": []interface{}{}, "totals": map[string]interface{}{}}, http.StatusOK)
		return
	}

	searchBy := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("search_by")))

	insuranceFilter := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("insurance")))
	var insuranceCompanyID *int
	if insuranceFilter != "" && insuranceFilter != "all" && insuranceFilter != "all_insurances" {
		if id, err := strconv.Atoi(insuranceFilter); err == nil {
			insuranceCompanyID = &id
		} else {
			jsonError(w, "Invalid insurance filter", http.StatusBadRequest)
			return
		}
	}

	result, err := h.svc.ARInsuranceAging(locIDs, startDate, endDate, searchBy, insuranceCompanyID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}
