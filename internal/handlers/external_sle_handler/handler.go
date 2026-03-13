package external_sle_handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	sleSvc "sighthub-backend/internal/services/external_sle_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct{ svc *sleSvc.Service }

func New(svc *sleSvc.Service) *Handler { return &Handler{svc: svc} }

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
	case strings.Contains(msg, "external_sle not found"):
		return http.StatusNotFound
	case strings.Contains(msg, "cannot update"):
		return http.StatusForbidden
	case strings.Contains(msg, "not authorized"):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func (h *Handler) SaveExternalSle(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	var input sleSvc.SaveExternalSleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.SaveExternalSle(username, examID, input)
	if err != nil {
		var ve *sleSvc.ValidationError
		if errors.As(err, &ve) {
			jsonResponse(w, http.StatusBadRequest, map[string]interface{}{"errors": ve.Errors})
			return
		}
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusCreated, result)
}

func (h *Handler) GetExternalSle(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	result, err := h.svc.GetExternalSle(examID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

func (h *Handler) UpdateExternalSle(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	var input sleSvc.UpdateExternalSleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.UpdateExternalSle(username, examID, input)
	if err != nil {
		var ve *sleSvc.ValidationError
		if errors.As(err, &ve) {
			jsonResponse(w, http.StatusBadRequest, map[string]interface{}{"errors": ve.Errors})
			return
		}
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}
