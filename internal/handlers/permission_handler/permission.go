package permission_handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	svc "sighthub-backend/internal/services/permission_service"
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

func locationIDFromQuery(r *http.Request) (int, bool) {
	s := r.URL.Query().Get("location_id")
	if s == "" {
		return 0, false
	}
	id, err := strconv.Atoi(s)
	return id, err == nil
}

// GET /api/employee/permissions/{block}/:employee_id
func (h *Handler) GetBlockPermissions(w http.ResponseWriter, r *http.Request) {
	employeeID, err := employeeIDFromVars(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	blockID, err := strconv.Atoi(mux.Vars(r)["block_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid block_id")
		return
	}
	locationID, ok := locationIDFromQuery(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "location_id is required")
		return
	}
	result, err := h.svc.GetBlockPermissions(employeeID, blockID, locationID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// POST /api/employee/permissions/{block}/:employee_id
func (h *Handler) SetBlockPermission(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	employeeID, err := employeeIDFromVars(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	blockID, err := strconv.Atoi(mux.Vars(r)["block_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid block_id")
		return
	}
	locationID, ok := locationIDFromQuery(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "location_id is required")
		return
	}
	var body struct {
		CombinationID int  `json:"combination_id"`
		Granted       bool `json:"granted"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.SetBlockPermission(username, employeeID, blockID, locationID, body.CombinationID, body.Granted); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// GET /api/employee/permissions/warehouses/:employee_id
func (h *Handler) GetWarehouseAccess(w http.ResponseWriter, r *http.Request) {
	employeeID, err := employeeIDFromVars(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	locationID, ok := locationIDFromQuery(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "location_id is required")
		return
	}
	result, err := h.svc.GetWarehouseAccess(employeeID, locationID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// POST /api/employee/permissions/warehouses/:employee_id
func (h *Handler) SetWarehouseAccess(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	employeeID, err := employeeIDFromVars(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	locationID, ok := locationIDFromQuery(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "location_id is required")
		return
	}
	var body struct {
		WarehouseID int  `json:"warehouse_id"`
		Granted     bool `json:"granted"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.SetWarehouseAccess(username, employeeID, locationID, body.WarehouseID, body.Granted); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Access updated"})
}
