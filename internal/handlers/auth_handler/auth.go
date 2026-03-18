package auth_handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"sighthub-backend/config"
	"sighthub-backend/internal/services/auth_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

// Handler — auth хэндлер
type Handler struct {
	svc *auth_service.AuthService
	cfg *config.Config
}

func New(svc *auth_service.AuthService, cfg *config.Config) *Handler {
	return &Handler{svc: svc, cfg: cfg}
}

// Login — POST /auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Username == "" || body.Password == "" {
		jsonError(w, "username and password are required", http.StatusBadRequest)
		return
	}

	result, err := h.svc.Login(r.Context(), body.Username, body.Password, r.RemoteAddr, r.Header.Get("User-Agent"))
	if err != nil {
		var blErr *auth_service.BlacklistError
		if errors.As(err, &blErr) {
			jsonResponse(w, map[string]interface{}{
				"error":          "Account is temporarily locked due to multiple failed login attempts. Try again later.",
				"time_remaining": blErr.TimeRemaining,
			}, http.StatusForbidden)
			return
		}
		switch err {
		case auth_service.ErrInactive:
			jsonError(w, "Account is inactive", http.StatusForbidden)
		case auth_service.ErrInvalidCreds:
			jsonError(w, "Invalid credentials", http.StatusUnauthorized)
		default:
			jsonError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	h.setRefreshCookie(w, result.RefreshToken)
	jsonResponse(w, map[string]string{"access_token": result.AccessToken}, http.StatusOK)
}

// LoginWithPin — POST /auth/login_with_pin
func (h *Handler) LoginWithPin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ExpressLogin string `json:"express_login"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ExpressLogin == "" {
		jsonError(w, "PIN code is required", http.StatusBadRequest)
		return
	}

	pin := body.ExpressLogin
	if len(pin) != 5 {
		jsonError(w, "PIN code must be exactly 5 digits", http.StatusBadRequest)
		return
	}
	for _, c := range pin {
		if c < '0' || c > '9' {
			jsonError(w, "PIN code must be exactly 5 digits", http.StatusBadRequest)
			return
		}
	}

	result, err := h.svc.LoginWithPin(r.Context(), pin, r.RemoteAddr, r.Header.Get("User-Agent"))
	if err != nil {
		switch err {
		case auth_service.ErrInvalidPin:
			jsonError(w, "Invalid PIN code", http.StatusUnauthorized)
		case auth_service.ErrInactive:
			jsonError(w, "Account is inactive", http.StatusForbidden)
		default:
			jsonError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	h.setRefreshCookie(w, result.RefreshToken)
	jsonResponse(w, map[string]string{"access_token": result.AccessToken}, http.StatusOK)
}

// Refresh — POST /auth/token/refresh (refresh_token из httponly cookie)
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		jsonError(w, "Missing refresh token", http.StatusBadRequest)
		return
	}

	result, err := h.svc.Refresh(r.Context(), cookie.Value)
	if err != nil {
		switch err {
		case auth_service.ErrNoRefreshToken:
			jsonError(w, "Missing refresh token", http.StatusBadRequest)
		case auth_service.ErrExpiredRefresh:
			jsonError(w, "Expired refresh token", http.StatusUnauthorized)
		case auth_service.ErrRevoked:
			jsonError(w, "Refresh token has been revoked. Please login again.", http.StatusForbidden)
		default:
			jsonError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	h.setRefreshCookie(w, result.RefreshToken)
	jsonResponse(w, map[string]string{"access_token": result.AccessToken}, http.StatusOK)
}

// Logout — POST /auth/logout (защищён JWTMiddleware)
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	if username == "" {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.svc.Logout(r.Context(), username); err != nil {
		if err == auth_service.ErrNotFound {
			jsonError(w, "User not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		MaxAge:   -1,
		Path:     "/",
	})
	jsonResponse(w, map[string]string{"message": "Logged out successfully"}, http.StatusOK)
}

// TokenCheck — POST /auth/token-check
func (h *Handler) TokenCheck(w http.ResponseWriter, r *http.Request) {
	var body struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.AccessToken == "" {
		jsonError(w, "access_token is required", http.StatusBadRequest)
		return
	}

	tok := strings.TrimPrefix(body.AccessToken, "Bearer ")
	claims, err := pkgAuth.ParseToken(tok, h.cfg.JWTSecretKey)
	if err != nil {
		jsonResponse(w, map[string]interface{}{"valid": false, "error": err.Error()}, http.StatusOK)
		return
	}

	blacklisted, _ := h.svc.IsTokenBlacklisted(r.Context(), claims.ID)
	if blacklisted {
		jsonResponse(w, map[string]bool{"valid": false}, http.StatusOK)
		return
	}

	jsonResponse(w, map[string]bool{"valid": true}, http.StatusOK)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func (h *Handler) setRefreshCookie(w http.ResponseWriter, token string) {
	sameSite := http.SameSiteLaxMode
	switch h.cfg.CookieSameSite {
	case "None":
		sameSite = http.SameSiteNoneMode
	case "Strict":
		sameSite = http.SameSiteStrictMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: sameSite,
		Path:     "/",
	})
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
