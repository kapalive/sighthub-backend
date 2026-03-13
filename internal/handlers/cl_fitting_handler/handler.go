package cl_fitting_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	clSvc "sighthub-backend/internal/services/cl_fitting_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct{ svc *clSvc.Service }

func New(svc *clSvc.Service) *Handler { return &Handler{svc: svc} }

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
	case strings.Contains(msg, "cl_fitting not found"):
		return http.StatusNotFound
	case strings.Contains(msg, "cannot update"):
		return http.StatusForbidden
	case strings.Contains(msg, "not authorized"):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func (h *Handler) SaveClFitting(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	var input clSvc.SaveClFittingInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.SaveClFitting(username, examID, input)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusCreated, result)
}

func (h *Handler) GetClFitting(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	result, err := h.svc.GetClFitting(examID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, result)
}

func (h *Handler) UpdateClFitting(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	var input clSvc.UpdateClFittingInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.UpdateClFitting(username, examID, input)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}

	jsonResponse(w, http.StatusOK, result)
}

func (h *Handler) GetContactLensBrands(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetContactLensBrands()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}
