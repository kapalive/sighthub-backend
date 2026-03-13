package pos_terminal_handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	posSvc "sighthub-backend/internal/services/pos_terminal_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct {
	svc *posSvc.Service
}

func New(svc *posSvc.Service) *Handler { return &Handler{svc: svc} }

// GET /pos/terminals
func (h *Handler) ListTerminals(w http.ResponseWriter, r *http.Request) {
	el, err := h.getEL(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	result, err := h.svc.ListTerminals(el)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /pos/terminals
func (h *Handler) CreateTerminal(w http.ResponseWriter, r *http.Request) {
	el, err := h.getEL(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	var req posSvc.CreateTerminalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.svc.CreateTerminal(el, req)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusCreated)
}

// PUT /pos/terminals/{terminal_id}
func (h *Handler) UpdateTerminal(w http.ResponseWriter, r *http.Request) {
	el, err := h.getEL(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	tid, err := strconv.Atoi(mux.Vars(r)["terminal_id"])
	if err != nil {
		jsonError(w, "invalid terminal_id", http.StatusBadRequest)
		return
	}
	var req posSvc.UpdateTerminalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.svc.UpdateTerminal(el, tid, req)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /pos/terminals/{terminal_id}/default
func (h *Handler) SetDefaultTerminal(w http.ResponseWriter, r *http.Request) {
	el, err := h.getEL(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	tid, err := strconv.Atoi(mux.Vars(r)["terminal_id"])
	if err != nil {
		jsonError(w, "invalid terminal_id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.SetDefaultTerminal(el, tid)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// DELETE /pos/terminals/{terminal_id}
func (h *Handler) DeleteTerminal(w http.ResponseWriter, r *http.Request) {
	el, err := h.getEL(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	tid, err := strconv.Atoi(mux.Vars(r)["terminal_id"])
	if err != nil {
		jsonError(w, "invalid terminal_id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.DeleteTerminal(el, tid)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /invoice/{invoice_id}/pos/start
func (h *Handler) PosStart(w http.ResponseWriter, r *http.Request) {
	el, err := h.getEL(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	invoiceID, err := strconv.ParseInt(mux.Vars(r)["invoice_id"], 10, 64)
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	var req posSvc.PosStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.svc.PosStart(el, invoiceID, req)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /invoice/{invoice_id}/pos/commit
func (h *Handler) PosCommit(w http.ResponseWriter, r *http.Request) {
	el, err := h.getEL(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	invoiceID, err := strconv.ParseInt(mux.Vars(r)["invoice_id"], 10, 64)
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}
	var req posSvc.PosCommitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.svc.PosCommit(el, invoiceID, req)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /pos/spin-config/provision
func (h *Handler) ProvisionSpinConfig(w http.ResponseWriter, r *http.Request) {
	el, err := h.getEL(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	var req posSvc.ProvisionSpinConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := h.svc.ProvisionSpinConfig(el, req)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /pos/spin-config
func (h *Handler) GetSpinConfig(w http.ResponseWriter, r *http.Request) {
	el, err := h.getEL(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	result, err := h.svc.GetSpinConfig(el)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /pos/tx/{tx_id}
func (h *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	el, err := h.getEL(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	txID, err := strconv.ParseInt(mux.Vars(r)["tx_id"], 10, 64)
	if err != nil {
		jsonError(w, "invalid tx_id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GetTransaction(el, txID)
	if err != nil {
		jsonError(w, err.Error(), httpStatus(err))
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (h *Handler) getEL(r *http.Request) (*posSvc.EmpLocation, error) {
	username := pkgAuth.UsernameFromContext(r.Context())
	return h.svc.GetEmpLocation(username)
}

func httpStatus(err error) int {
	switch {
	case errors.Is(err, posSvc.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, posSvc.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, posSvc.ErrBadRequest):
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
