package inventory_handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	invSvc "sighthub-backend/internal/services/inventory_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case float64:
		return int64(val)
	case string:
		n, _ := strconv.ParseInt(val, 10, 64)
		return n
	case nil:
		return 0
	default:
		n, _ := strconv.ParseInt(fmt.Sprintf("%v", val), 10, 64)
		return n
	}
}

// GET /count_sheets
func (h *Handler) GetCountSheets(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var brandID, vendorID *int64
	if v := r.URL.Query().Get("brand_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			brandID = &id
		}
	}
	if v := r.URL.Query().Get("vendor_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			vendorID = &id
		}
	}
	var dateFrom, dateTo *string
	if v := r.URL.Query().Get("date_from"); v != "" {
		dateFrom = &v
	}
	if v := r.URL.Query().Get("date_to"); v != "" {
		dateTo = &v
	}

	result, err := h.svc.GetCountSheets(username, brandID, vendorID, dateFrom, dateTo)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// POST /count_sheets
func (h *Handler) CreateCountSheet(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var input invSvc.CreateCountSheetInput
	if err := decodeJSON(r, &input); err != nil {
		jsonResponse(w, 400, map[string]string{"error": "invalid request body"})
		return
	}
	result, err := h.svc.CreateCountSheet(username, input)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 201, result)
}

// DELETE /count_sheets
func (h *Handler) DeleteCountSheet(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	idCountSheet, err := strconv.ParseInt(r.URL.Query().Get("id_count_sheet"), 10, 64)
	if err != nil || idCountSheet == 0 {
		jsonResponse(w, 400, map[string]string{"error": "id_count_sheet is required"})
		return
	}
	if err := h.svc.DeleteCountSheet(username, idCountSheet); err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, map[string]interface{}{
		"status":  200,
		"message": "Count sheet deleted successfully",
	})
}

// GET /count_sheets/{id_count_sheet}
func (h *Handler) GetCountSheetInfo(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	idCountSheet, err := strconv.ParseInt(mux.Vars(r)["id_count_sheet"], 10, 64)
	if err != nil {
		jsonResponse(w, 400, map[string]string{"error": "invalid id_count_sheet"})
		return
	}
	result, err := h.svc.GetCountSheetInfo(username, idCountSheet)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// PUT /count_sheets/{id_count_sheet}
func (h *Handler) UpdateCountSheetNotes(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	idCountSheet, err := strconv.ParseInt(mux.Vars(r)["id_count_sheet"], 10, 64)
	if err != nil {
		jsonResponse(w, 400, map[string]string{"error": "invalid id_count_sheet"})
		return
	}
	var body struct {
		Notes string `json:"notes"`
	}
	if err := decodeJSON(r, &body); err != nil {
		jsonResponse(w, 400, map[string]string{"error": "field 'notes' is required"})
		return
	}
	if err := h.svc.UpdateCountSheetNotes(username, idCountSheet, body.Notes); err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, map[string]string{"message": "Notes updated successfully"})
}

// GET /count_sheets/items
func (h *Handler) GetCountSheetItems(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	idCountSheet, err := strconv.ParseInt(r.URL.Query().Get("id_count_sheet"), 10, 64)
	if err != nil || idCountSheet == 0 {
		jsonResponse(w, 400, map[string]string{"error": "id_count_sheet is required"})
		return
	}
	result, err := h.svc.GetCountSheetItems(username, idCountSheet)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// POST /count_sheets/items
func (h *Handler) AddItemToCountSheet(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var raw map[string]interface{}
	if err := decodeJSON(r, &raw); err != nil {
		jsonResponse(w, 400, map[string]string{"error": "invalid request body"})
		return
	}

	idCountSheet := toInt64(raw["id_count_sheet"])
	skuInt := toInt64(raw["sku"])

	if idCountSheet == 0 || skuInt == 0 {
		jsonResponse(w, 400, map[string]string{"error": "id_count_sheet and sku are required"})
		return
	}
	sku := strconv.FormatInt(skuInt, 10)
	result, err := h.svc.AddItemToCountSheet(username, idCountSheet, sku)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// DELETE /count_sheets/items
func (h *Handler) DeleteItemFromCountSheet(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	idCountSheet, _ := strconv.ParseInt(r.URL.Query().Get("id_count_sheet"), 10, 64)
	itemID, _ := strconv.ParseInt(r.URL.Query().Get("item_id"), 10, 64)
	if idCountSheet == 0 || itemID == 0 {
		jsonResponse(w, 400, map[string]string{"error": "id_count_sheet and item_id are required"})
		return
	}
	result, err := h.svc.DeleteItemFromCountSheet(username, idCountSheet, itemID)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// POST /count_sheets/close/{id_count_sheet}
func (h *Handler) CloseCountSheet(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	idCountSheet, err := strconv.ParseInt(mux.Vars(r)["id_count_sheet"], 10, 64)
	if err != nil {
		jsonResponse(w, 400, map[string]string{"error": "invalid id_count_sheet"})
		return
	}
	result, err := h.svc.CloseCountSheet(username, idCountSheet)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}
