package report_daily_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sighthub-backend/internal/middleware"
	svc "sighthub-backend/internal/services/report_daily_service"
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

func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func parseDate(s string) (time.Time, bool) {
	t, err := time.Parse("2006-01-02", s)
	return t, err == nil
}

func permittedIDs(r *http.Request) []int {
	return middleware.PermittedLocationIDsFromCtx(r.Context())
}

func validateLocationID(r *http.Request, param string) (int, int, string) {
	v := strings.TrimSpace(r.URL.Query().Get(param))
	if v == "" {
		return 0, http.StatusBadRequest, param + " is required"
	}
	id, err := strconv.Atoi(v)
	if err != nil {
		return 0, http.StatusBadRequest, "invalid " + param
	}
	for _, pid := range permittedIDs(r) {
		if pid == id {
			return id, 0, ""
		}
	}
	return 0, http.StatusForbidden, "Permission denied for this location"
}

func defaultLocationID(r *http.Request) int {
	emp := middleware.EmployeeFromCtx(r.Context())
	if emp != nil && emp.LocationID != nil {
		return int(*emp.LocationID)
	}
	return 0
}

// ─── 1. GET /sales ───────────────────────────────────────────────────────────

func (h *Handler) DailySalesSummary(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	targetDate := time.Now().UTC()
	if dateStr != "" {
		if d, ok := parseDate(dateStr); ok {
			targetDate = d
		}
	}

	result, err := h.svc.DailySalesSummary(permittedIDs(r), targetDate)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, result)
}

// ─── 2. GET /monthly_sales_summary ───────────────────────────────────────────

func (h *Handler) MonthlySalesSummary(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
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

	rows, err := h.svc.MonthlySalesSummary(permittedIDs(r), month, year)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]interface{}{"data": rows})
}

// ─── 3. GET /ytd_sales_summary ───────────────────────────────────────────────

func (h *Handler) YTDSalesSummary(w http.ResponseWriter, r *http.Request) {
	locStr := r.URL.Query().Get("location_id")
	if locStr == "" {
		jsonError(w, "location_id is required", http.StatusBadRequest)
		return
	}
	locID, errCode, errMsg := validateLocationID(r, "location_id")
	if errMsg != "" {
		jsonError(w, errMsg, errCode)
		return
	}

	today := time.Now().UTC()
	startOfYear := time.Date(today.Year(), 1, 1, 0, 0, 0, 0, time.UTC)

	startDate := startOfYear
	endDate := today
	if v := r.URL.Query().Get("start"); v != "" {
		if d, ok := parseDate(v); ok {
			startDate = d
		} else {
			jsonError(w, "Invalid date format. Use ISO format like 2024-01-01.", http.StatusBadRequest)
			return
		}
	} else if v := r.URL.Query().Get("start_date"); v != "" {
		if d, ok := parseDate(v); ok {
			startDate = d
		} else {
			jsonError(w, "Invalid date format. Use ISO format like 2024-01-01.", http.StatusBadRequest)
			return
		}
	}
	if v := r.URL.Query().Get("end"); v != "" {
		if d, ok := parseDate(v); ok {
			endDate = d
		} else {
			jsonError(w, "Invalid date format. Use ISO format like 2024-01-01.", http.StatusBadRequest)
			return
		}
	} else if v := r.URL.Query().Get("end_date"); v != "" {
		if d, ok := parseDate(v); ok {
			endDate = d
		} else {
			jsonError(w, "Invalid date format. Use ISO format like 2024-01-01.", http.StatusBadRequest)
			return
		}
	}

	var empID *int
	if v := r.URL.Query().Get("rep_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			empID = &id
		}
	}

	rows, err := h.svc.YTDSalesSummary(locID, startDate, endDate, empID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]interface{}{"data": rows})
}

// ─── 4. GET /sales_cash ─────────────────────────────────────────────────────

func (h *Handler) DailySalesCash(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	targetDate := time.Now().UTC()
	if dateStr != "" {
		if d, ok := parseDate(dateStr); ok {
			targetDate = d
		}
	}

	result, err := h.svc.DailySalesCash(permittedIDs(r), targetDate)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, result)
}

// ─── 5. GET /journal_report ──────────────────────────────────────────────────

func (h *Handler) JournalReport(w http.ResponseWriter, r *http.Request) {
	locID, errCode, errMsg := validateLocationID(r, "location_id")
	if errMsg != "" {
		jsonError(w, errMsg, errCode)
		return
	}

	now := time.Now().UTC()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := now

	if v := r.URL.Query().Get("start_date"); v != "" {
		if d, ok := parseDate(v); ok {
			startDate = d
		}
	}
	if v := r.URL.Query().Get("end_date"); v != "" {
		if d, ok := parseDate(v); ok {
			endDate = d
		}
	}

	summary := r.URL.Query().Get("summary") == "true"

	result, err := h.svc.JournalReport(locID, startDate, endDate, summary)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, result)
}

// ─── 6. GET /journal_transfer ────────────────────────────────────────────────

func (h *Handler) JournalTransfer(w http.ResponseWriter, r *http.Request) {
	locStr := r.URL.Query().Get("location_id")
	var locID int
	if locStr != "" {
		id, errCode, errMsg := validateLocationID(r, "location_id")
		if errMsg != "" {
			jsonError(w, errMsg, errCode)
			return
		}
		locID = id
	} else {
		locID = defaultLocationID(r)
		if locID == 0 {
			jsonError(w, "Employee or location not found", http.StatusNotFound)
			return
		}
	}

	now := time.Now().UTC()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := now

	if v := r.URL.Query().Get("start_date"); v != "" {
		if d, ok := parseDate(v); ok {
			startDate = d
		}
	}
	if v := r.URL.Query().Get("end_date"); v != "" {
		if d, ok := parseDate(v); ok {
			endDate = d
		}
	}

	result, err := h.svc.JournalTransfer(locID, startDate, endDate)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, result)
}

// ─── 7. GET /journal_receipts ────────────────────────────────────────────────

func (h *Handler) JournalReceipts(w http.ResponseWriter, r *http.Request) {
	locStr := r.URL.Query().Get("location_id")
	var locID int
	if locStr != "" {
		id, errCode, errMsg := validateLocationID(r, "location_id")
		if errMsg != "" {
			jsonError(w, errMsg, errCode)
			return
		}
		locID = id
	} else {
		locID = defaultLocationID(r)
		if locID == 0 {
			jsonError(w, "Employee or location not found", http.StatusNotFound)
			return
		}
	}

	now := time.Now().UTC()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := now

	if v := r.URL.Query().Get("start_date"); v != "" {
		if d, ok := parseDate(v); ok {
			startDate = d
		}
	}
	if v := r.URL.Query().Get("end_date"); v != "" {
		if d, ok := parseDate(v); ok {
			endDate = d
		}
	}

	result, err := h.svc.JournalReceipts(locID, startDate, endDate)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, result)
}

// ─── 8. GET /all_reports ─────────────────────────────────────────────────────

func (h *Handler) AllReports(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, svc.AllReports())
}

// ─── 9. GET /locations ───────────────────────────────────────────────────────

func (h *Handler) Locations(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ShowcaseLocations(permittedIDs(r))
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, items)
}

// ─── 10. GET /employees ──────────────────────────────────────────────────────

func (h *Handler) Employees(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.AllEmployees()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, items)
}
