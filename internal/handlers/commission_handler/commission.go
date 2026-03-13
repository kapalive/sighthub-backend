package commission_handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	svc "sighthub-backend/internal/services/commission_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct {
	svc *svc.Service
}

func New(s *svc.Service) *Handler {
	return &Handler{svc: s}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func employeeIDFromVars(r *http.Request) (int, error) {
	return strconv.Atoi(mux.Vars(r)["employee_id"])
}

// GET /api/employee/{employee_id}/commissions
func (h *Handler) GetCommissions(w http.ResponseWriter, r *http.Request) {
	_ = pkgAuth.UsernameFromContext(r.Context())
	employeeID, err := employeeIDFromVars(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	result, err := h.svc.GetCommissions(employeeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GET /api/employee/{employee_id}/commission/current
func (h *Handler) GetCurrentCommission(w http.ResponseWriter, r *http.Request) {
	_ = pkgAuth.UsernameFromContext(r.Context())
	employeeID, err := employeeIDFromVars(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	result, err := h.svc.GetCurrentCommission(employeeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GET /api/employee/{employee_id}/commissions/history
func (h *Handler) GetCommissionHistory(w http.ResponseWriter, r *http.Request) {
	_ = pkgAuth.UsernameFromContext(r.Context())
	employeeID, err := employeeIDFromVars(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	result, err := h.svc.GetCommissionHistory(employeeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// PUT /api/employee/{employee_id}/commissions/{commission_id}
func (h *Handler) UpdateCommission(w http.ResponseWriter, r *http.Request) {
	employeeID, err := employeeIDFromVars(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	commissionID, err := strconv.Atoi(mux.Vars(r)["commission_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid commission_id")
		return
	}
	var input svc.UpdateCommissionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateCommission(employeeID, commissionID, input); err != nil {
		if err.Error() == "commission not found" {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Commission updated successfully"})
}

// POST /api/employee/{employee_id}/commissions
func (h *Handler) CreateCommission(w http.ResponseWriter, r *http.Request) {
	employeeID, err := employeeIDFromVars(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	var input svc.CreateCommissionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id, err := h.svc.CreateCommission(employeeID, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":               "New commission created successfully",
		"id_employee_commissions": id,
	})
}
