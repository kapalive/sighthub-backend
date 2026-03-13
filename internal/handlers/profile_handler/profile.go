package profile_handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"sighthub-backend/internal/services/profile_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct {
	svc *profile_service.Service
}

func New(svc *profile_service.Service) *Handler {
	return &Handler{svc: svc}
}

// ─── ExpressPass  POST /api/profile/express_pass ──────────────────────────────

func (h *Handler) ExpressPass(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Password == "" {
		jsonError(w, "Password is required", http.StatusBadRequest)
		return
	}

	pin, err := h.svc.ExpressPass(r.Context(), username, body.Password)
	if err != nil {
		switch err {
		case profile_service.ErrUserNotFound:
			jsonError(w, err.Error(), http.StatusNotFound)
		case profile_service.ErrLockedOut:
			jsonError(w, err.Error(), http.StatusForbidden)
		case profile_service.ErrInvalidPassword:
			jsonError(w, "Invalid password", http.StatusUnauthorized)
		default:
			jsonError(w, err.Error(), http.StatusNotFound)
		}
		return
	}

	jsonResponse(w, map[string]string{"express_login": pin}, http.StatusOK)
}

// ─── GetInfo  GET /api/profile/info ──────────────────────────────────────────

func (h *Handler) GetInfo(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	info, err := h.svc.GetInfo(r.Context(), username)
	if err != nil {
		switch err {
		case profile_service.ErrUserNotFound:
			jsonError(w, "User not found", http.StatusNotFound)
		case profile_service.ErrEmployeeNotFound:
			jsonError(w, "Employee not found", http.StatusNotFound)
		default:
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	jsonResponse(w, info, http.StatusOK)
}

// ─── ChangePassword  POST /api/profile/change_password ───────────────────────

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil ||
		body.CurrentPassword == "" || body.NewPassword == "" {
		jsonError(w, "Both current and new passwords are required", http.StatusBadRequest)
		return
	}

	if err := h.svc.ChangePassword(r.Context(), username, body.CurrentPassword, body.NewPassword); err != nil {
		switch err {
		case profile_service.ErrUserNotFound:
			jsonError(w, "User not found", http.StatusNotFound)
		case profile_service.ErrLockedOut:
			jsonError(w, err.Error(), http.StatusForbidden)
		default:
			jsonError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	jsonResponse(w, map[string]string{"message": "Password updated successfully"}, http.StatusOK)
}

// ─── GetInvoiceByID  GET /api/profile/invoice/{invoice_id} ───────────────────

func (h *Handler) GetInvoiceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID, err := strconv.ParseInt(vars["invoice_id"], 10, 64)
	if err != nil {
		jsonError(w, "invalid invoice_id", http.StatusBadRequest)
		return
	}

	result, err := h.svc.GetInvoiceByID(r.Context(), invoiceID)
	if err != nil {
		if err == profile_service.ErrInvoiceNotFound {
			jsonError(w, "Invoice not found", http.StatusNotFound)
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	jsonResponse(w, result, http.StatusOK)
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
