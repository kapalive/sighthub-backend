package special_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	specialSvc "sighthub-backend/internal/services/special_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct{ svc *specialSvc.Service }

func New(svc *specialSvc.Service) *Handler { return &Handler{svc: svc} }

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
	case strings.Contains(msg, "special_eye_exam not found"),
		strings.Contains(msg, "special_eye_file not found"):
		return http.StatusNotFound
	case strings.Contains(msg, "cannot update"), strings.Contains(msg, "not authorized"):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

func (h *Handler) SaveSpecial(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	var input specialSvc.SaveSpecialInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.SaveSpecial(username, examID, input)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusCreated, result)
}

func (h *Handler) GetSpecial(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	result, err := h.svc.GetSpecial(examID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

func (h *Handler) UpdateSpecial(w http.ResponseWriter, r *http.Request) {
	examID, err := parsePathInt64(r, "exam_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}
	var input specialSvc.UpdateSpecialInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.UpdateSpecial(username, examID, input)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

func (h *Handler) DeleteSpecialEyeFile(w http.ResponseWriter, r *http.Request) {
	fileID, err := parsePathInt64(r, "file_id")
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid file_id"})
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	id, err := h.svc.DeleteSpecialEyeFile(username, fileID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"message":          "File deleted successfully",
		"id_special_eye_file": id,
	})
}
