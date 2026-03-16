package inventory_handler

import (
	"net/http"
	"strconv"
	"strings"

	invSvc "sighthub-backend/internal/services/inventory_service"
)

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

// GET /all-filters
func (h *Handler) GetInventoryByFilter(w http.ResponseWriter, r *http.Request) {
	params := invSvc.FilterParams{
		LocationIDs: parseMultiInt(r, "location_id"),
		VendorIDs:   parseMultiInt(r, "vendor_id"),
		BrandIDs:    parseMultiInt(r, "brand_id"),
		ProductIDs:  parseMultiInt(r, "product_id"),
		ModelIDs:    parseMultiInt(r, "variant_id"),
		Statuses:    parseMultiStr(r, "status"),
		VendorNames: parseMultiStr(r, "vendor_name"),
		GroupBy:     r.URL.Query().Get("group_by"),
		Output:      r.URL.Query().Get("output"),
		Page:        1,
		PerPage:     25,
	}

	// brand names: support both ?brand= and ?brand_name=
	brandNames := parseMultiStr(r, "brand")
	if len(brandNames) == 0 {
		brandNames = parseMultiStr(r, "brand_name")
	}
	params.BrandNames = brandNames

	if v := r.URL.Query().Get("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			params.Page = p
		}
	}
	if v := r.URL.Query().Get("per_page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			params.PerPage = p
		}
	}

	// Resolve effective location IDs via permissions → fallback
	effectiveLocationIDs, err := resolveLocationIDs(h.db, r, params.LocationIDs)
	if err != nil || len(effectiveLocationIDs) == 0 {
		jsonResponse(w, 200, &invSvc.FilterResult{Items: []map[string]interface{}{}})
		return
	}

	if params.Output == "csv" || params.Output == "csv_detail" {
		rows, err := h.svc.GetInventoryCSV(params, effectiveLocationIDs)
		if err != nil {
			jsonResponse(w, 500, map[string]string{"error": err.Error()})
			return
		}
		csvData := invSvc.FormatCSV(rows)
		w.Header().Set("Content-Disposition", "attachment; filename=inventory.csv")
		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(200)
		w.Write([]byte(csvData))
		return
	}

	result, err := h.svc.GetInventoryByFilter(params, effectiveLocationIDs)
	if err != nil {
		jsonResponse(w, 500, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}
