package ticket_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	ticketSvc "sighthub-backend/internal/services/ticket_service"
)

type Handler struct{ svc *ticketSvc.Service }

func New(svc *ticketSvc.Service) *Handler { return &Handler{svc: svc} }

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

// POST /{ticket_id}/notify-patient
func (h *Handler) NotifyPatient(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["ticket_id"]
	ticketID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid ticket_id"})
		return
	}
	result, err := h.svc.NotifyPatient(ticketID)
	if err != nil {
		jsonResponse(w, errorStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, result)
}
