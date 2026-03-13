package ordered_inventory_handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	orderedSvc "sighthub-backend/internal/services/ordered_inventory_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct {
	svc *orderedSvc.Service
}

func New(svc *orderedSvc.Service) *Handler { return &Handler{svc: svc} }

// POST /api/ordered-inventory/add
func (h *Handler) AddOrderedItem(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	var req orderedSvc.AddOrderedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.svc.AddOrderedItem(el, req)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusCreated)
}

// GET /api/ordered-inventory/pending?vendor_id=N
func (h *Handler) GetPendingItems(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	vendorIDStr := r.URL.Query().Get("vendor_id")
	vendorID, err := strconv.ParseInt(vendorIDStr, 10, 64)
	if err != nil || vendorID == 0 {
		jsonError(w, "vendor_id is required", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GetPendingItems(el, vendorID)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /api/ordered-inventory/receive
func (h *Handler) ReceiveOrderedItem(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	el, err := h.svc.GetEmpLocation(username)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	var req orderedSvc.ReceiveOrderedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.svc.ReceiveOrderedItem(el, req)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func httpStatus(err error) int {
	switch {
	case errors.Is(err, orderedSvc.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, orderedSvc.ErrBadRequest):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

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
