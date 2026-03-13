package helpdesk_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	pkgAuth "sighthub-backend/pkg/auth"
	svc "sighthub-backend/internal/services/helpdesk_service"
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

// GET /api/helpdesk/?page=&per_page=&status=&q=&order=
func (h *Handler) ListTickets(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	if username == "" {
		writeError(w, http.StatusBadRequest, "cannot extract employee_login from token identity")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page == 0 {
		page = 1
	}
	if perPage == 0 {
		perPage = 20
	}
	status := r.URL.Query().Get("status")
	q := r.URL.Query().Get("q")
	order := r.URL.Query().Get("order")
	if order == "" {
		order = "desc"
	}

	result, err := h.svc.ListTickets(username, page, perPage, status, q, order)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// POST /api/helpdesk/
func (h *Handler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	if username == "" {
		writeError(w, http.StatusBadRequest, "cannot extract employee_login from token identity")
		return
	}

	var body struct {
		Subject     string   `json:"subject"`
		Description string   `json:"description"`
		Screenshots []string `json:"screenshots"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	subject := strings.TrimSpace(body.Subject)
	description := strings.TrimSpace(body.Description)

	if subject == "" {
		writeError(w, http.StatusBadRequest, "subject is required")
		return
	}
	if description == "" {
		writeError(w, http.StatusBadRequest, "description is required")
		return
	}
	if len(subject) > 200 {
		writeError(w, http.StatusBadRequest, "subject max length is 200")
		return
	}
	if len(description) > 5000 {
		writeError(w, http.StatusBadRequest, "description max length is 5000")
		return
	}

	screenshots := body.Screenshots
	if screenshots == nil {
		screenshots = []string{}
	}
	if len(screenshots) > 5 {
		writeError(w, http.StatusBadRequest, "screenshots max count is 5")
		return
	}
	clean := make([]string, 0, len(screenshots))
	for _, u := range screenshots {
		clean = append(clean, strings.TrimSpace(u))
	}

	// Extract source domain from request headers
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	host = strings.Split(host, ",")[0]
	host = strings.TrimSpace(strings.Split(host, ":")[0])
	var sourceDomain *string
	if host != "" {
		sourceDomain = &host
	}

	ticket, err := h.svc.CreateTicket(username, subject, description, clean, sourceDomain)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"message": "created", "ticket": ticket})
}

// GET /api/helpdesk/{ticket_id}
func (h *Handler) GetTicket(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	if username == "" {
		writeError(w, http.StatusBadRequest, "cannot extract employee_login from token identity")
		return
	}

	idStr := mux.Vars(r)["ticket_id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket_id")
		return
	}

	ticket, err := h.svc.GetTicket(id, username)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ticket": ticket})
}

// POST /api/helpdesk/{ticket_id} — external webhook, HMAC-protected, no JWT
func (h *Handler) ReceiveReply(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["ticket_id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ticket_id")
		return
	}

	gotHMAC := strings.TrimSpace(r.Header.Get("X-Trusted-HMAC"))

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		body = map[string]interface{}{}
	}

	ticket, err := h.svc.ReceiveReply(id, gotHMAC, body)
	if err != nil {
		switch err.Error() {
		case "unauthorized":
			writeError(w, http.StatusUnauthorized, err.Error())
		case "ticket not found":
			writeError(w, http.StatusNotFound, err.Error())
		case "HELPDESK_REPLY_HMAC not configured":
			writeError(w, http.StatusInternalServerError, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"message": "ok", "ticket": ticket})
}
