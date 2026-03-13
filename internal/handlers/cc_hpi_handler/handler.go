package cc_hpi_handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	ccHpiSvc "sighthub-backend/internal/services/cc_hpi_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct{ svc *ccHpiSvc.Service }

func New(svc *ccHpiSvc.Service) *Handler { return &Handler{svc: svc} }

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func parsePathInt64(r *http.Request, key string) (int64, bool) {
	v, ok := mux.Vars(r)[key]
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseInt(v, 10, 64)
	return n, err == nil
}

// POST /{exam_id}
func (h *Handler) SaveCcHpi(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	var input ccHpiSvc.SaveCcHpiInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.svc.SaveCcHpi(username, examID, input)
	if err != nil {
		switch err.Error() {
		case "employee not found", "exam not found":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "you are not authorized to update this exam", "cannot create cc_hpi for a completed exam",
			"cc_hpi already exists for this exam":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"message":       "CcHpiEye saved successfully",
		"id_cc_hpi_eye": result["id_cc_hpi_eye"],
		"data":          result,
	})
}

// GET /{exam_id}
func (h *Handler) GetCcHpi(w http.ResponseWriter, r *http.Request) {
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	result, err := h.svc.GetCcHpi(examID)
	if err != nil {
		jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// PUT /{exam_id}
func (h *Handler) UpdateCcHpi(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	examID, ok := parsePathInt64(r, "exam_id")
	if !ok {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid exam_id"})
		return
	}

	var input ccHpiSvc.UpdateCcHpiInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := h.svc.UpdateCcHpi(username, examID, input)
	if err != nil {
		switch err.Error() {
		case "employee not found", "exam not found", "cc_hpi record not found for this exam":
			jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case "cannot update cc_hpi for a completed exam", "you are not authorized to update this exam":
			jsonResponse(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{"message": "CcHpiEye updated successfully", "data": result})
}
