package report_sales_handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"

	"sighthub-backend/internal/middleware"
	svc "sighthub-backend/internal/services/report_sales_service"
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

// GET /breakdown
func (h *Handler) Breakdown(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// permitted locations
	username := pkgAuth.UsernameFromContext(r.Context())
	if username == "" {
		jsonError(w, "unauthorized", 401)
		return
	}
	allowedIDs := middleware.GetPermittedLocationIDs(h.db, username, 11)
	if len(allowedIDs) == 0 {
		jsonError(w, "No permitted locations", 403)
		return
	}

	// date range
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := today.Add(-24 * time.Hour)

	dateStart := yesterday
	dateEnd := today.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	if v := q.Get("date_start"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			dateStart = t
		} else {
			jsonError(w, "Invalid date format. Use YYYY-MM-DD", 400)
			return
		}
	}
	if v := q.Get("date_end"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			dateEnd = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.UTC)
		} else {
			jsonError(w, "Invalid date format. Use YYYY-MM-DD", 400)
			return
		}
	}

	// location
	var effectiveLocIDs []int
	if v := q.Get("location_id"); v != "" {
		locID, err := strconv.Atoi(v)
		if err != nil {
			jsonError(w, "invalid location_id", 400)
			return
		}
		if !contains(allowedIDs, locID) {
			jsonError(w, "Permission denied for this location", 403)
			return
		}
		effectiveLocIDs = []int{locID}
	} else {
		empLocID, err := h.svc.GetEmployeeLocationID(username)
		if err == nil && contains(allowedIDs, empLocID) {
			effectiveLocIDs = []int{empLocID}
		} else {
			effectiveLocIDs = allowedIDs
		}
	}

	// optional filters
	var employeeID *int
	if v := q.Get("employee_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			employeeID = &id
		}
	}

	outputFormat := q.Get("output")

	result, err := h.svc.Breakdown(effectiveLocIDs, dateStart, dateEnd, employeeID)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	// CSV output
	if outputFormat == "csv" {
		csv := csvutil.New()
		csv.Row(
			"Date", "Invoice", "Name", "Inv. Total", "Cost",
			"Oph Fr Only", "Oph Fr Rx", "Sun Plano", "Sun Rx",
			"L Only", "Prof Fees", "Contacts", "Other",
			"Discount", "Tax", "Employee", "Cost Info Missing",
		)

		for _, d := range result.Data {
			csv.Row(
				d.Date, d.Invoice, d.Name,
				csvutil.F(d.InvTotal), csvutil.F(d.Cost),
				csvutil.F(d.OphFrOnly), csvutil.F(d.OphFrRx),
				csvutil.F(d.SunPlano), csvutil.F(d.SunRx),
				csvutil.F(d.LOnly), csvutil.F(d.ProfFees),
				csvutil.F(d.Contacts), csvutil.F(d.Other),
				csvutil.F(d.Discount), csvutil.F(d.Tax),
				d.Employee, d.CostInfoMissing,
			)
		}

		t := result.Totals
		csv.EmptyRow()
		csv.Row(
			"", "", "Total",
			csvutil.F(t.InvTotal), csvutil.F(t.Cost),
			csvutil.F(t.OphFrOnly), csvutil.F(t.OphFrRx),
			csvutil.F(t.SunPlano), csvutil.F(t.SunRx),
			csvutil.F(t.LOnly), csvutil.F(t.ProfFees),
			csvutil.F(t.Contacts), csvutil.F(t.Other),
			csvutil.F(t.Discount), csvutil.F(t.Tax),
			"", "",
		)

		p := result.PercentNetOfTax
		csv.Row(
			"", "", "Percent Net of Tax",
			"", "",
			fmt.Sprintf("%.1f", p.OphFrOnly), fmt.Sprintf("%.1f", p.OphFrRx),
			fmt.Sprintf("%.1f", p.SunPlano), fmt.Sprintf("%.1f", p.SunRx),
			fmt.Sprintf("%.1f", p.LOnly), fmt.Sprintf("%.1f", p.ProfFees),
			fmt.Sprintf("%.1f", p.Contacts), fmt.Sprintf("%.1f", p.Other),
			"", "", "", "",
		)

		csv.Row(fmt.Sprintf("Record Count: %d", result.RecordCount))

		csv.ServeHTTP(w, "sales_breakdown_by_product_type.csv")
		return
	}

	jsonOK(w, result)
}

func contains(ids []int, id int) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}
