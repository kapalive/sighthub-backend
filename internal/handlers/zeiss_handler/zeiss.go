package zeiss_handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	svc "sighthub-backend/internal/services/zeiss_service"
)

type Handler struct {
	svc *svc.Service
}

func New(s *svc.Service) *Handler {
	return &Handler{svc: s}
}

// GET /api/integrations/zeiss/auth-url?location_id=X
func (h *Handler) AuthURL(w http.ResponseWriter, r *http.Request) {
	locID, err := parseLocationID(r)
	if err != nil {
		jsonError(w, "location_id is required", http.StatusBadRequest)
		return
	}

	resp, err := h.svc.GetAuthURL(r.Context(), locID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, svc.ErrCredentialNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, svc.ErrCredentialInactive) {
			status = http.StatusForbidden
		}
		jsonError(w, err.Error(), status)
		return
	}

	jsonResponse(w, resp, http.StatusOK)
}

// GET /api/integrations/zeiss/oauth2redirect?code=...&state=...
func (h *Handler) OAuth2Redirect(w http.ResponseWriter, r *http.Request) {
	errParam := r.URL.Query().Get("error")
	if errParam != "" {
		jsonError(w, errParam, http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	redirectURL, err := h.svc.HandleCallback(r.Context(), code, state)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, svc.ErrInvalidState) {
			status = http.StatusUnauthorized
		}
		jsonError(w, err.Error(), status)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// POST /api/integrations/zeiss/exchange
func (h *Handler) Exchange(w http.ResponseWriter, r *http.Request) {
	var req svc.ExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.Exchange(r.Context(), req); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, svc.ErrInvalidState) {
			status = http.StatusUnauthorized
		} else if errors.Is(err, svc.ErrTokenExchange) {
			status = http.StatusBadGateway
		}
		jsonError(w, err.Error(), status)
		return
	}

	jsonResponse(w, map[string]bool{"ok": true}, http.StatusOK)
}

// GET /api/integrations/zeiss/token?location_id=X
func (h *Handler) Token(w http.ResponseWriter, r *http.Request) {
	locID, err := parseLocationID(r)
	if err != nil {
		jsonError(w, "location_id is required", http.StatusBadRequest)
		return
	}

	resp, err := h.svc.GetToken(r.Context(), locID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, svc.ErrNotAuthenticated) || errors.Is(err, svc.ErrCredentialNotFound) {
			status = http.StatusUnauthorized
		}
		jsonError(w, err.Error(), status)
		return
	}

	jsonResponse(w, resp, http.StatusOK)
}

// POST /api/integrations/zeiss/refresh?location_id=X
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	locID, err := parseLocationID(r)
	if err != nil {
		jsonError(w, "location_id is required", http.StatusBadRequest)
		return
	}

	resp, err := h.svc.RefreshToken(r.Context(), locID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, svc.ErrNotAuthenticated) {
			status = http.StatusUnauthorized
		} else if errors.Is(err, svc.ErrNoRefreshToken) {
			status = http.StatusUnprocessableEntity
		}
		jsonError(w, err.Error(), status)
		return
	}

	jsonResponse(w, resp, http.StatusOK)
}

// POST /api/integrations/zeiss/logout?location_id=X
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	locID, err := parseLocationID(r)
	if err != nil {
		jsonError(w, "location_id is required", http.StatusBadRequest)
		return
	}

	if err := h.svc.Logout(r.Context(), locID); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]bool{"ok": true}, http.StatusOK)
}

// GET /api/integrations/zeiss/
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, map[string]string{"message": "Zeiss Routes are operational."}, http.StatusOK)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func parseLocationID(r *http.Request) (int64, error) {
	s := r.URL.Query().Get("location_id")
	if s == "" {
		return 0, errors.New("location_id is required")
	}
	return strconv.ParseInt(s, 10, 64)
}

func jsonResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	jsonResponse(w, map[string]string{"error": msg}, status)
}
