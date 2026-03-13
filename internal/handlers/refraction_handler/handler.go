package refraction_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	refrSvc "sighthub-backend/internal/services/refraction_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct{ svc *refrSvc.Service }

func New(svc *refrSvc.Service) *Handler { return &Handler{svc: svc} }

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
	case strings.Contains(msg, "refraction not found"):
		return http.StatusNotFound
	case strings.Contains(msg, "cannot update"):
		return http.StatusForbidden
	case strings.Contains(msg, "not authorized"):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func (h *Handler) SaveRefraction(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	var input refrSvc.SaveRefractionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	if err := h.svc.SaveRefraction(username, examID, input); err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]string{"message": "Refraction saved successfully"})
}

func (h *Handler) GetRefraction(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	result, err := h.svc.GetRefraction(examID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, result)
}

func (h *Handler) UpdateRefraction(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	var input refrSvc.UpdateRefractionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.UpdateRefraction(username, examID, input)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, result)
}
