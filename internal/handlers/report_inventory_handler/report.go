package report_inventory_handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"sighthub-backend/internal/middleware"
	svc "sighthub-backend/internal/services/report_inventory_service"
	pkgAuth "sighthub-backend/pkg/auth"
	"sighthub-backend/pkg/csvutil"
)

type Handler struct {
	svc *svc.Service
	db  *gorm.DB
}

func New(s *svc.Service, db *gorm.DB) *Handler { return &Handler{svc: s, db: db} }

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

func permittedIDs(r *http.Request, db *gorm.DB) []int {
	username := pkgAuth.UsernameFromContext(r.Context())
	if username == "" {
		return nil
	}
	return middleware.GetPermittedLocationIDs(db, username, 11)
}

func employeeLocationID(r *http.Request) *int64 {
	emp := middleware.EmployeeFromCtx(r.Context())
	if emp != nil {
		return emp.LocationID
	}
	return nil
}

func parseMultiInt(r *http.Request, key string) []int64 {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil
	}
	var out []int64
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if v, err := strconv.ParseInt(s, 10, 64); err == nil {
			out = append(out, v)
		}
	}
	return out
}

func parseMultiStr(r *http.Request, key string) []string {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil
	}
	var out []string
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// resolveSimpleLocations resolves which location IDs to use based on
// explicit location_id param, employee default, and allowed IDs.
// Supports location_id=all, location_id=<int>, or empty (default to employee location).
func resolveSimpleLocations(r *http.Request, allowedIDs []int) ([]int, string, int) {
	emp := middleware.EmployeeFromCtx(r.Context())
	locStr := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("location_id")))

	if locStr == "" {
		if emp != nil && emp.LocationID != nil {
			lid := int(*emp.LocationID)
			for _, id := range allowedIDs {
				if id == lid {
					return []int{lid}, "", 0
				}
			}
		}
		return allowedIDs, "", 0
	}

	if locStr == "all" {
		return allowedIDs, "", 0
	}

	lid, err := strconv.Atoi(locStr)
	if err != nil {
		return nil, "Invalid location_id", http.StatusBadRequest
	}
	for _, id := range allowedIDs {
		if id == lid {
			return []int{lid}, "", 0
		}
	}
	return nil, "Permission denied for this location", http.StatusForbidden
}

// ─── 1. GET /frame-interaction ──────────────────────────────────────────────────

func (h *Handler) FrameInteraction(w http.ResponseWriter, r *http.Request) {
	allowedIDs := permittedIDs(r, h.db)
	if len(allowedIDs) == 0 {
		jsonError(w, "No permitted locations", http.StatusForbidden)
		return
	}

	effectiveIDs, errMsg, errCode := resolveSimpleLocations(r, allowedIDs)
	if errMsg != "" {
		jsonError(w, errMsg, errCode)
		return
	}

	today := time.Now().UTC()
	yesterday := today.AddDate(0, 0, -1)
	dateStart := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC)
	dateEnd := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 999999999, time.UTC)

	if v := r.URL.Query().Get("date_start"); v != "" {
		if d, ok := parseDate(v); ok {
			dateStart = d
		} else {
			jsonError(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}
	if v := r.URL.Query().Get("date_end"); v != "" {
		if d, ok := parseDate(v); ok {
			dateEnd = time.Date(d.Year(), d.Month(), d.Day(), 23, 59, 59, 999999999, time.UTC)
		} else {
			jsonError(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	vendorIDs := parseMultiInt(r, "vendor_id")
	brandIDs := parseMultiInt(r, "brand_id")
	statuses := parseMultiStr(r, "status")
	vendorNames := parseMultiStr(r, "vendor_name")
	brandNames := parseMultiStr(r, "brand")
	if len(brandNames) == 0 {
		brandNames = parseMultiStr(r, "brand_name")
	}

	items, totalCost, err := h.svc.FrameInteraction(effectiveIDs, dateStart, dateEnd, vendorIDs, brandIDs, statuses, vendorNames, brandNames)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	recordCount := len(items)

	if r.URL.Query().Get("output") == "csv" {
		cw := csvutil.New()
		cw.Row("Serial", "Item Nbr", "Model ID", "Description", "Status",
			"Work Flow", "Begin Date", "End Date", "Stock",
			"Employee Login", "Employee Name", "Cost")
		for _, it := range items {
			bd := ""
			if it.BeginDate != nil {
				bd = *it.BeginDate
			}
			ed := ""
			if it.EndDate != nil {
				ed = *it.EndDate
			}
			cw.Row(
				fmt.Sprintf("%d", it.Serial), it.ItemNbr, fmt.Sprintf("%d", it.ModelID),
				it.Description, it.Status, it.WorkFlow,
				bd, ed, it.Stock,
				it.EmployeeLogin, it.EmployeeName, csvutil.F(it.Cost),
			)
		}
		cw.EmptyRow()
		cw.Row(fmt.Sprintf("Record Count: %d", recordCount), "", "", "", "", "", "", "", "", "", "", csvutil.F(totalCost))
		cw.ServeHTTP(w, "inventory_work_flow.csv")
		return
	}

	jsonOK(w, map[string]interface{}{
		"data":         items,
		"record_count": recordCount,
		"total_cost":   totalCost,
	})
}

// ─── 2. GET /missing_inventory ──────────────────────────────────────────────────

func (h *Handler) MissingInventory(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	if startStr == "" || endStr == "" {
		jsonError(w, "start and end dates required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	startDate, ok1 := parseDate(startStr)
	endDate, ok2 := parseDate(endStr)
	if !ok1 || !ok2 {
		jsonError(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	tsStart := startDate
	tsEnd := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.UTC)

	allowedIDs := permittedIDs(r, h.db)
	if len(allowedIDs) == 0 {
		jsonError(w, "No permitted locations", http.StatusForbidden)
		return
	}

	locationIDs, errMsg, errCode := resolveSimpleLocations(r, allowedIDs)
	if errMsg != "" {
		jsonError(w, errMsg, errCode)
		return
	}

	items, totalCost, err := h.svc.MissingInventory(locationIDs, tsStart, tsEnd)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	recordCount := len(items)

	if strings.ToLower(r.URL.Query().Get("output")) == "csv" {
		cw := csvutil.New()
		cw.Row("Store", "Vendor", "Brand", "Model", "Cost",
			"F Number", "Serial Nbr", "In Stock Date")
		for _, it := range items {
			inStock := ""
			if it.InStockDate != nil {
				inStock = *it.InStockDate
			}
			cw.Row(it.Store, it.Vendor, it.Brand, it.Model,
				csvutil.F(it.Cost), it.FNumber,
				fmt.Sprintf("%d", it.Serial), inStock)
		}
		cw.EmptyRow()
		cw.Row(fmt.Sprintf("Record Count: %d", recordCount), "", "", "",
			csvutil.F(totalCost))
		cw.ServeHTTP(w, "missing_inventory.csv")
		return
	}

	jsonOK(w, map[string]interface{}{
		"data":         items,
		"record_count": recordCount,
		"total_cost":   totalCost,
		"start":        startDate.Format("2006-01-02"),
		"end":          endDate.Format("2006-01-02"),
	})
}

// ─── 3. GET /receipt_by_brand ───────────────────────────────────────────────────

func (h *Handler) ReceiptByBrand(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	if startStr == "" || endStr == "" {
		jsonError(w, "start and end dates required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	startDate, ok1 := parseDate(startStr)
	endDate, ok2 := parseDate(endStr)
	if !ok1 || !ok2 {
		jsonError(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	tsStart := startDate
	tsEnd := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.UTC)

	allowedIDs := permittedIDs(r, h.db)
	if len(allowedIDs) == 0 {
		jsonError(w, "No permitted locations", http.StatusForbidden)
		return
	}

	locationIDs, errMsg, errCode := resolveSimpleLocations(r, allowedIDs)
	if errMsg != "" {
		jsonError(w, errMsg, errCode)
		return
	}

	items, totalQty, totalCost, totalPrice, err := h.svc.ReceiptByBrand(locationIDs, tsStart, tsEnd)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	recordCount := len(items)

	if strings.ToLower(r.URL.Query().Get("output")) == "csv" {
		cw := csvutil.New()
		cw.Row("Vendor", "Brand", "Qty", "Cost", "Price")
		for _, it := range items {
			cw.Row(it.Vendor, it.Brand, fmt.Sprintf("%d", it.Qty),
				csvutil.F(it.Cost), csvutil.F(it.Price))
		}
		cw.EmptyRow()
		cw.Row(fmt.Sprintf("Record Count: %d", recordCount), "",
			fmt.Sprintf("%d", totalQty), csvutil.F(totalCost), csvutil.F(totalPrice))
		cw.ServeHTTP(w, "receipt_by_brand.csv")
		return
	}

	jsonOK(w, map[string]interface{}{
		"data":         items,
		"record_count": recordCount,
		"total_qty":    totalQty,
		"total_cost":   totalCost,
		"total_price":  totalPrice,
		"start":        startDate.Format("2006-01-02"),
		"end":          endDate.Format("2006-01-02"),
	})
}

// ─── 4. GET /list_of_receipts ───────────────────────────────────────────────────

func (h *Handler) ListOfReceipts(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	if startStr == "" || endStr == "" {
		jsonError(w, "start and end dates required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	startDate, ok1 := parseDate(startStr)
	endDate, ok2 := parseDate(endStr)
	if !ok1 || !ok2 {
		jsonError(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	tsStart := startDate
	tsEnd := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.UTC)

	allowedIDs := permittedIDs(r, h.db)
	if len(allowedIDs) == 0 {
		jsonError(w, "No permitted locations", http.StatusForbidden)
		return
	}

	locStr := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("location_id")))
	empLocID := employeeLocationID(r)

	locationIDs, errMsg := h.svc.ResolveReceiptLocations(locStr, allowedIDs, empLocID)
	if errMsg != "" {
		code := http.StatusBadRequest
		if strings.Contains(errMsg, "Permission denied") || strings.Contains(errMsg, "No permitted") {
			code = http.StatusForbidden
		}
		jsonError(w, errMsg, code)
		return
	}
	if len(locationIDs) == 0 {
		jsonOK(w, map[string]interface{}{"data": []interface{}{}, "record_count": 0, "totals": map[string]interface{}{}})
		return
	}

	items, totals, err := h.svc.ListOfReceipts(locationIDs, tsStart, tsEnd)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	recordCount := len(items)

	if strings.ToLower(r.URL.Query().Get("output")) == "csv" {
		cw := csvutil.New()
		cw.Row("Date", "Receipt #", "Vendor", "Location", "Invoice Dt",
			"Pack / Inv #", "Qty", "Sub Total", "S / H", "Tax", "Total", "Our Ref #")
		for _, it := range items {
			d, invD, packInv := "", "", ""
			if it.Date != nil {
				d = *it.Date
			}
			if it.InvoiceDate != nil {
				invD = *it.InvoiceDate
			}
			if it.PackInv != nil {
				packInv = *it.PackInv
			}
			cw.Row(d, it.ReceiptNo, it.Vendor, it.Location, invD,
				packInv, fmt.Sprintf("%d", it.Qty),
				csvutil.F(it.SubTotal), csvutil.F(it.ShippingHandling),
				csvutil.F(it.Tax), csvutil.F(it.Total), it.OrderRef)
		}
		cw.EmptyRow()
		cw.Row(
			fmt.Sprintf("Count: %d", recordCount), "", "", "", "", "",
			fmt.Sprintf("%d", totals.Qty), csvutil.F(totals.SubTotal),
			csvutil.F(totals.ShippingHandling), csvutil.F(totals.Tax),
			csvutil.F(totals.Total), "")
		cw.ServeHTTP(w, "list_of_receipts.csv")
		return
	}

	jsonOK(w, map[string]interface{}{
		"data":         items,
		"record_count": recordCount,
		"totals":       totals,
		"start":        startDate.Format("2006-01-02"),
		"end":          endDate.Format("2006-01-02"),
	})
}

// ─── 5. GET /internal_transfers ─────────────────────────────────────────────────

func (h *Handler) InternalTransfers(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	if startStr == "" || endStr == "" {
		jsonError(w, "start and end dates required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	startDate, ok1 := parseDate(startStr)
	endDate, ok2 := parseDate(endStr)
	if !ok1 || !ok2 {
		jsonError(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	tsStart := startDate
	tsEnd := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.UTC)

	allowedIDs := permittedIDs(r, h.db)
	if len(allowedIDs) == 0 {
		jsonError(w, "No permitted locations", http.StatusForbidden)
		return
	}

	locStr := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("location_id")))
	empLocID := employeeLocationID(r)

	locationIDs, errMsg := h.svc.ResolveAllLocations(locStr, allowedIDs, empLocID)
	if errMsg != "" {
		code := http.StatusBadRequest
		if strings.Contains(errMsg, "Permission denied") || strings.Contains(errMsg, "No permitted") {
			code = http.StatusForbidden
		}
		jsonError(w, errMsg, code)
		return
	}
	if len(locationIDs) == 0 {
		jsonOK(w, map[string]interface{}{"data": []interface{}{}, "record_count": 0, "totals": map[string]interface{}{}})
		return
	}

	items, totalCost, totalPrice, err := h.svc.InternalTransfers(locationIDs, tsStart, tsEnd)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	recordCount := len(items)

	if strings.ToLower(r.URL.Query().Get("output")) == "csv" {
		cw := csvutil.New()
		cw.Row("Date", "Serial", "From", "To", "Brand", "Model", "Type", "Cost", "Price")
		for _, it := range items {
			d := ""
			if it.Date != nil {
				d = *it.Date
			}
			cw.Row(d, fmt.Sprintf("%d", it.Serial),
				it.FromLocation, it.ToLocation,
				it.Brand, it.Model, it.TransactionType,
				csvutil.F(it.Cost), csvutil.F(it.Price))
		}
		cw.EmptyRow()
		cw.Row(fmt.Sprintf("Record Count: %d", recordCount), "", "", "", "", "", "",
			csvutil.F(totalCost), csvutil.F(totalPrice))
		cw.ServeHTTP(w, "internal_transfers.csv")
		return
	}

	jsonOK(w, map[string]interface{}{
		"data":         items,
		"record_count": recordCount,
		"total_cost":   totalCost,
		"total_price":  totalPrice,
		"start":        startDate.Format("2006-01-02"),
		"end":          endDate.Format("2006-01-02"),
	})
}

// ─── 6. GET /locations/can-receive ──────────────────────────────────────────────

func (h *Handler) CanReceiveLocations(w http.ResponseWriter, r *http.Request) {
	allowedIDs := permittedIDs(r, h.db)
	items, err := h.svc.CanReceiveLocations(allowedIDs)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, items)
}

// ─── 7. GET /locations/all ──────────────────────────────────────────────────────

func (h *Handler) AllLocations(w http.ResponseWriter, r *http.Request) {
	allowedIDs := permittedIDs(r, h.db)
	items, err := h.svc.AllLocations(allowedIDs)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, items)
}
