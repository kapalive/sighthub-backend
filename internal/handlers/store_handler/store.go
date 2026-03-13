package store_handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	storeSvc "sighthub-backend/internal/services/store_service"
)

type Handler struct {
	svc *storeSvc.Service
}

func New(svc *storeSvc.Service) *Handler { return &Handler{svc: svc} }

// GET /
func (h *Handler) GetAllStores(w http.ResponseWriter, r *http.Request) {
	stores, err := h.svc.GetAllStores()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, stores, http.StatusOK)
}

// POST /
func (h *Handler) CreateStore(w http.ResponseWriter, r *http.Request) {
	var input storeSvc.StoreInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	resp, err := h.svc.CreateStore(input)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, resp, http.StatusCreated)
}

// GET /{store_id}
func (h *Handler) GetStore(w http.ResponseWriter, r *http.Request) {
	storeID, err := parseID(r, "store_id")
	if err != nil {
		jsonError(w, "invalid store_id", http.StatusBadRequest)
		return
	}
	resp, err := h.svc.GetStore(storeID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, resp, http.StatusOK)
}

// GET /{store_id}/request-appointment-link
func (h *Handler) GetRequestAppointmentLink(w http.ResponseWriter, r *http.Request) {
	storeID, err := parseID(r, "store_id")
	if err != nil {
		jsonError(w, "invalid store_id", http.StatusBadRequest)
		return
	}
	link, err := h.svc.GetRequestAppointmentLink(storeID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusForbidden)
		return
	}
	jsonResponse(w, map[string]string{"link": link}, http.StatusOK)
}

// PUT /{store_id}
func (h *Handler) UpdateStore(w http.ResponseWriter, r *http.Request) {
	storeID, err := parseID(r, "store_id")
	if err != nil {
		jsonError(w, "invalid store_id", http.StatusBadRequest)
		return
	}
	var input storeSvc.StoreInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateStore(storeID, input); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, map[string]string{"message": "Store updated successfully"}, http.StatusOK)
}

// PUT /{store_id}/activate
func (h *Handler) ActivateStore(w http.ResponseWriter, r *http.Request) {
	storeID, err := parseID(r, "store_id")
	if err != nil {
		jsonError(w, "invalid store_id", http.StatusBadRequest)
		return
	}
	var body map[string]interface{}
	json.NewDecoder(r.Body).Decode(&body) //nolint:errcheck

	var active *bool
	if v, ok := body["active"].(bool); ok {
		active = &v
	}

	respData, statusCode, err := h.svc.ActivateStore(storeID, active)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, respData, statusCode)
}

// GET /warehouses
func (h *Handler) GetWarehouses(w http.ResponseWriter, r *http.Request) {
	warehouses, err := h.svc.GetWarehouses()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, warehouses, http.StatusOK)
}

// GET /warehouses/{warehouse_id}
func (h *Handler) GetWarehouse(w http.ResponseWriter, r *http.Request) {
	whID, err := parseID(r, "warehouse_id")
	if err != nil {
		jsonError(w, "invalid warehouse_id", http.StatusBadRequest)
		return
	}
	resp, err := h.svc.GetWarehouse(whID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, resp, http.StatusOK)
}

// PUT /warehouses/{warehouse_id}
func (h *Handler) UpdateWarehouse(w http.ResponseWriter, r *http.Request) {
	whID, err := parseID(r, "warehouse_id")
	if err != nil {
		jsonError(w, "invalid warehouse_id", http.StatusBadRequest)
		return
	}
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateWarehouse(whID, data); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, map[string]string{"message": "Warehouse updated successfully"}, http.StatusOK)
}

// POST /warehouses
func (h *Handler) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	whID, err := h.svc.CreateWarehouse(data)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, map[string]int{"id_warehouse": whID}, http.StatusCreated)
}

// GET /sales_tax_list
func (h *Handler) GetSalesTaxList(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.GetSalesTaxList()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, list, http.StatusOK)
}

// GET /sales_taxes
func (h *Handler) GetSalesTaxes(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.GetSalesTaxes()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, list, http.StatusOK)
}

// --- utils ---

func parseID(r *http.Request, key string) (int, error) {
	return strconv.Atoi(mux.Vars(r)[key])
}

func jsonResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data) //nolint:errcheck
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}
