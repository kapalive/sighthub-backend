package timecard_handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	svc "sighthub-backend/internal/services/timecard_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct {
	svc *svc.Service
}

func New(s *svc.Service) *Handler { return &Handler{svc: s} }

// GET /api/timecard/session_status?username=...
func (h *Handler) SessionStatus(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		jsonError(w, "Username is required", http.StatusBadRequest)
		return
	}

	result, err := h.svc.GetSessionStatus(username)
	if err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			jsonResponse(w, map[string]bool{"session": false}, http.StatusNotFound)
			return
		}
		if errors.Is(err, svc.ErrHistoryNotFound) {
			jsonResponse(w, map[string]bool{"session": false}, http.StatusOK)
			return
		}
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, result, http.StatusOK)
}

// POST /api/timecard/check
func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Username == "" || body.Password == "" {
		jsonError(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	result, err := h.svc.CheckInOut(body.Username, body.Password)
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidCreds):
			jsonError(w, "Invalid username or password", http.StatusUnauthorized)
		case errors.Is(err, svc.ErrInactive):
			jsonError(w, "Account is inactive", http.StatusForbidden)
		default:
			jsonError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	jsonResponse(w, result, http.StatusOK)
}

// POST /api/timecard/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Username == "" || body.Password == "" {
		jsonError(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	token, err := h.svc.Login(body.Username, body.Password)
	if err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidCreds):
			jsonError(w, "Invalid username or password", http.StatusUnauthorized)
		case errors.Is(err, svc.ErrInactive):
			jsonError(w, "Account is inactive", http.StatusForbidden)
		default:
			jsonError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	jsonResponse(w, map[string]string{"access_token": token}, http.StatusOK)
}

// GET /api/timecard/info  [JWT]
func (h *Handler) Info(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	today := time.Now()
	startDate := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	endDate := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 999999999, today.Location())

	if s := r.URL.Query().Get("start_date"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			jsonError(w, "Invalid start_date format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		startDate = t
	}
	if e := r.URL.Query().Get("end_date"); e != "" {
		t, err := time.Parse("2006-01-02", e)
		if err != nil {
			jsonError(w, "Invalid end_date format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		endDate = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
	}

	result, err := h.svc.GetInfo(username, startDate, endDate)
	if err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			jsonError(w, "Timecard account not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, result, http.StatusOK)
}

// POST /api/timecard/change_password  [JWT]
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var body struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.OldPassword == "" || body.NewPassword == "" {
		jsonError(w, "old_password and new_password are required", http.StatusBadRequest)
		return
	}

	if err := h.svc.ChangePassword(username, body.OldPassword, body.NewPassword); err != nil {
		switch {
		case errors.Is(err, svc.ErrInvalidCreds):
			jsonError(w, "Invalid credentials", http.StatusUnauthorized)
		case errors.Is(err, svc.ErrInactive):
			jsonError(w, "Account is inactive", http.StatusForbidden)
		default:
			jsonError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	jsonResponse(w, map[string]string{"message": "Password updated successfully"}, http.StatusCreated)
}

// PUT /api/timecard/history/{history_id}/note  [JWT]
func (h *Handler) UpdateHistoryNote(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	historyID, err := strconv.Atoi(mux.Vars(r)["history_id"])
	if err != nil {
		jsonError(w, "Invalid history id", http.StatusBadRequest)
		return
	}

	var body struct {
		Note *string `json:"note"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	if err := h.svc.UpdateHistoryNote(username, historyID, body.Note); err != nil {
		if errors.Is(err, svc.ErrHistoryNotFound) {
			jsonError(w, "History entry not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{"message": "Note updated successfully"}, http.StatusOK)
}

// POST /api/timecard/logout  [JWT]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, map[string]string{"message": "Logged out successfully"}, http.StatusOK)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

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
