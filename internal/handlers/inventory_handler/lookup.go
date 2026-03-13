package inventory_handler

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	invSvc "sighthub-backend/internal/services/inventory_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

// GET /lookup
func (h *Handler) LookupBySKU(w http.ResponseWriter, r *http.Request) {
	rawSKU := r.URL.Query().Get("sku")
	if rawSKU == "" {
		jsonResponse(w, 400, map[string]string{"error": "SKU is required"})
		return
	}
	result, err := h.svc.LookupBySKU(rawSKU)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// GET /history
func (h *Handler) GetInventoryHistory(w http.ResponseWriter, r *http.Request) {
	inventoryID, _ := strconv.ParseInt(r.URL.Query().Get("inventory_id"), 10, 64)
	rawSKU := r.URL.Query().Get("sku")

	if inventoryID == 0 && rawSKU == "" {
		jsonResponse(w, 400, map[string]string{"error": "specify inventory_id or sku"})
		return
	}
	result, err := h.svc.GetInventoryHistory(inventoryID, rawSKU)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// PUT /update_price/{inventory_id}
func (h *Handler) UpdatePrice(w http.ResponseWriter, r *http.Request) {
	invID, err := strconv.ParseInt(mux.Vars(r)["inventory_id"], 10, 64)
	if err != nil {
		jsonResponse(w, 400, map[string]string{"error": "invalid inventory_id"})
		return
	}

	var input invSvc.UpdatePriceInput
	if err := decodeJSON(r, &input); err != nil {
		jsonResponse(w, 400, map[string]string{"error": "invalid request body"})
		return
	}
	result, err := h.svc.UpdatePrice(invID, input)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// POST /update_state
func (h *Handler) UpdateInventoryState(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var input invSvc.StateChangeInput
	if err := decodeJSON(r, &input); err != nil {
		jsonResponse(w, 400, map[string]string{"error": "invalid request body"})
		return
	}
	result, err := h.svc.UpdateInventoryState(username, input)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}
