package inventory_handler

import (
	"encoding/json"
	"net/http"
	"strings"

	invSvc "sighthub-backend/internal/services/inventory_service"
)

type Handler struct{ svc *invSvc.Service }

func New(svc *invSvc.Service) *Handler { return &Handler{svc: svc} }

func jsonResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func httpStatus(err error) int {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "not found"):
		return http.StatusNotFound
	case strings.Contains(msg, "required"), strings.Contains(msg, "invalid"),
		strings.Contains(msg, "already exists"), strings.Contains(msg, "already"),
		strings.Contains(msg, "provide either"):
		return http.StatusBadRequest
	case strings.Contains(msg, "permission"), strings.Contains(msg, "denied"):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
