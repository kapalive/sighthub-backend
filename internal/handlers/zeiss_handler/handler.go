package zeiss_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"sighthub-backend/internal/services/zeiss_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct {
	auth    *zeiss_service.AuthService
	catalog *zeiss_service.CatalogService
	db      *gorm.DB

	mu         sync.Mutex
	importing  bool
	lastResult *zeiss_service.ImportResult
	lastError  string
}

func New(auth *zeiss_service.AuthService, db *gorm.DB) *Handler {
	return &Handler{
		auth:    auth,
		catalog: zeiss_service.NewCatalogService(db, auth),
		db:      db,
	}
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) employeeID(r *http.Request) (int64, bool) {
	username := pkgAuth.UsernameFromContext(r.Context())
	if username == "" {
		return 0, false
	}
	// employee_login.employee_login -> id_employee_login -> employee.employee_login_id -> id_employee
	var result struct {
		IDEmployee int64
	}
	err := h.db.Raw(`
		SELECT e.id_employee
		FROM employee_login el
		JOIN employee e ON e.employee_login_id = el.id_employee_login
		WHERE el.employee_login = ?
	`, username).Scan(&result).Error
	if err != nil || result.IDEmployee == 0 {
		return 0, false
	}
	return result.IDEmployee, true
}

// GET /zeiss/auth/status
func (h *Handler) AuthStatus(w http.ResponseWriter, r *http.Request) {
	empID, ok := h.employeeID(r)
	if !ok {
		jsonError(w, "unauthorized", 401)
		return
	}
	jsonOK(w, h.auth.GetAuthStatus(empID))
}

// GET /zeiss/auth/url
func (h *Handler) AuthURL(w http.ResponseWriter, r *http.Request) {
	empID, ok := h.employeeID(r)
	if !ok {
		jsonError(w, "unauthorized", 401)
		return
	}
	// Determine origin: query param > Origin header > X-Forwarded-Host > Referer > Host
	origin := r.URL.Query().Get("origin")
	if origin == "" {
		origin = r.Header.Get("Origin")
	}
	if origin == "" {
		if fh := r.Header.Get("X-Forwarded-Host"); fh != "" {
			proto := r.Header.Get("X-Forwarded-Proto")
			if proto == "" {
				proto = "https"
			}
			origin = proto + "://" + fh
		}
	}
	if origin == "" {
		if ref := r.Header.Get("Referer"); ref != "" {
			// Extract scheme://host from Referer
			if idx := nthIndex(ref, "/", 3); idx > 0 {
				origin = ref[:idx]
			}
		}
	}
	if origin == "" {
		origin = "https://" + r.Host
	}
	result, err := h.auth.GenerateAuthURL(empID, origin)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, result)
}

// POST /zeiss/auth/callback
func (h *Handler) AuthCallback(w http.ResponseWriter, r *http.Request) {
	var body struct {
		State string `json:"state"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid JSON", 400)
		return
	}
	if body.State == "" || body.Code == "" {
		jsonError(w, "state and code are required", 400)
		return
	}
	if err := h.auth.ExchangeCode(body.State, body.Code); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Zeiss authentication successful"})
}

// POST /zeiss/auth/logout
func (h *Handler) AuthLogout(w http.ResponseWriter, r *http.Request) {
	empID, ok := h.employeeID(r)
	if !ok {
		jsonError(w, "unauthorized", 401)
		return
	}
	h.auth.Logout(empID)
	jsonOK(w, map[string]string{"message": "Zeiss session removed"})
}

// POST /zeiss/import-catalog — starts async import
func (h *Handler) ImportCatalog(w http.ResponseWriter, r *http.Request) {
	empID, ok := h.employeeID(r)
	if !ok {
		jsonError(w, "unauthorized", 401)
		return
	}
	var body struct {
		CustomerNumber string `json:"customer_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.CustomerNumber == "" {
		jsonError(w, "customer_number is required", 400)
		return
	}

	h.mu.Lock()
	if h.importing {
		h.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "import already in progress"})
		return
	}
	h.importing = true
	h.lastResult = nil
	h.lastError = ""
	h.mu.Unlock()

	go func() {
		result, err := h.catalog.ImportCatalog(empID, body.CustomerNumber)
		h.mu.Lock()
		h.importing = false
		if err != nil {
			h.lastError = err.Error()
		} else {
			h.lastResult = result
		}
		h.mu.Unlock()
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"message": "import started"})
}

// GET /zeiss/import-catalog/status
func (h *Handler) ImportCatalogStatus(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")

	if h.importing {
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "importing"})
		return
	}
	if h.lastError != "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "error", "error": h.lastError})
		return
	}
	if h.lastResult != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "done", "result": h.lastResult})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "idle"})
}

// GET /zeiss/allowed-treatments?lens_code=59752
func (h *Handler) GetAllowedTreatments(w http.ResponseWriter, r *http.Request) {
	empID, ok := h.employeeID(r)
	if !ok {
		jsonError(w, "unauthorized", 401)
		return
	}
	lensCode := r.URL.Query().Get("lens_code")
	if lensCode == "" {
		jsonError(w, "lens_code is required", 400)
		return
	}
	status := h.auth.GetAuthStatus(empID)
	if status.CustomerNumber == nil || *status.CustomerNumber == "" {
		jsonError(w, "zeiss auth required", 401)
		return
	}
	result, err := h.catalog.GetAllowedTreatments(empID, lensCode, *status.CustomerNumber)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, result)
}

// GET /zeiss/order-requirements/{ticket_id}
func (h *Handler) ZeissOrderRequirements(w http.ResponseWriter, r *http.Request) {
	empID, ok := h.employeeID(r)
	if !ok {
		jsonError(w, "unauthorized", 401)
		return
	}
	ticketIDStr := mux.Vars(r)["ticket_id"]
	ticketID, err := strconv.ParseInt(ticketIDStr, 10, 64)
	if err != nil {
		jsonError(w, "invalid ticket_id", 400)
		return
	}
	result := h.catalog.CheckZeissOrderRequirements(ticketID, empID)
	jsonOK(w, result)
}

func nthIndex(s string, sub string, n int) int {
	idx := 0
	for i := 0; i < n; i++ {
		pos := strings.Index(s[idx:], sub)
		if pos < 0 {
			return -1
		}
		idx += pos + len(sub)
	}
	return idx - len(sub)
}
