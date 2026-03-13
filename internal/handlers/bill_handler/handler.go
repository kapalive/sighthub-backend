package bill_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	billSvc "sighthub-backend/internal/services/bill_service"
)

type Handler struct{ svc *billSvc.Service }

func New(svc *billSvc.Service) *Handler { return &Handler{svc: svc} }

func jsonResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func errorStatus(err error) int {
	if strings.Contains(err.Error(), "not found") {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

// GET /forms
func (h *Handler) GetFormData(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetFormData()
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}

// GET /forms/{id}/fields
func (h *Handler) GetFormFields(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id_form"]
	formID, err := strconv.Atoi(idStr)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid id_form"})
		return
	}
	fields, err := h.svc.GetFormFields(formID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{"fields": fields})
}

// POST /forms/{id}/submit
func (h *Handler) SubmitForm(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id_form"]
	formID, err := strconv.Atoi(idStr)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid id_form"})
		return
	}
	if err := h.svc.ValidateForm(formID); err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Data saved successfully"})
}
