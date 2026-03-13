package posterior_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	posteriorSvc "sighthub-backend/internal/services/posterior_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct{ svc *posteriorSvc.Service }

func New(svc *posteriorSvc.Service) *Handler { return &Handler{svc: svc} }

func jsonResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func parsePathInt64(r *http.Request, key string) (int64, error) {
	v, ok := mux.Vars(r)[key]
	if !ok {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseInt(v, 10, 64)
}

func errorStatus(err error) int {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "employee not found"):
		return http.StatusNotFound
	case strings.Contains(msg, "exam not found"):
		return http.StatusNotFound
	case strings.Contains(msg, "posterior_eye not found"),
		strings.Contains(msg, "findings_posterior not found"),
		strings.Contains(msg, "cup_disc_ratio_posterior not found"):
		return http.StatusNotFound
	case strings.Contains(msg, "cannot update"), strings.Contains(msg, "not authorized"):
		return http.StatusForbidden
	case strings.Contains(msg, "already exists"):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (h *Handler) SavePosterior(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	var input posteriorSvc.SavePosteriorInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.SavePosterior(username, examID, input)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusCreated, result)
}

func (h *Handler) GetPosterior(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	result, err := h.svc.GetPosterior(examID)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

func (h *Handler) UpdatePosterior(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	var input posteriorSvc.UpdatePosteriorInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.UpdatePosterior(username, examID, input)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}
