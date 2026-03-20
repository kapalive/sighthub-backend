package integration_handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	intSvc "sighthub-backend/internal/services/integration_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct {
	svc *intSvc.Service
}

func NewHandler(svc *intSvc.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// GET /api/settings/integration/{code}
func (h *Handler) GetIntegration(w http.ResponseWriter, r *http.Request) {
	code := mux.Vars(r)["code"]
	username := pkgAuth.UsernameFromContext(r.Context())
	locationID := h.getLocationID(r, username)
	if locationID == 0 {
		h.jsonError(w, "location not found", 404)
		return
	}

	result, err := h.svc.GetIntegration(code, locationID)
	if err != nil {
		h.jsonError(w, err.Error(), 400)
		return
	}
	h.jsonOK(w, result)
}

// POST /api/settings/integration/{code}
func (h *Handler) SetIntegration(w http.ResponseWriter, r *http.Request) {
	code := mux.Vars(r)["code"]
	username := pkgAuth.UsernameFromContext(r.Context())
	locationID := h.getLocationID(r, username)
	if locationID == 0 {
		h.jsonError(w, "location not found", 404)
		return
	}

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.jsonError(w, "invalid JSON", 400)
		return
	}
	if body.Username == "" || body.Password == "" {
		h.jsonError(w, "username and password are required", 400)
		return
	}

	result, err := h.svc.SetIntegration(code, locationID, body.Username, body.Password)
	if err != nil {
		h.jsonError(w, err.Error(), 500)
		return
	}
	h.jsonOK(w, result)
}

// GET /api/settings/integration
func (h *Handler) ListIntegrations(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.ListIntegrations()
	if err != nil {
		h.jsonError(w, err.Error(), 500)
		return
	}
	h.jsonOK(w, result)
}

func (h *Handler) getLocationID(r *http.Request, username string) int64 {
	if v := r.URL.Query().Get("location_id"); v != "" {
		id, _ := strconv.ParseInt(v, 10, 64)
		return id
	}
	// Get from employee
	var locID int64
	h.svc.DB().Raw(`
		SELECT e.location_id FROM employee e
		JOIN employee_login el ON el.id_employee_login = e.employee_login_id
		WHERE el.employee_login = ?`, username).Scan(&locID)
	return locID
}
