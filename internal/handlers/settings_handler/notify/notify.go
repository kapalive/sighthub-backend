package notify

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	svc "sighthub-backend/internal/services/settings_service"
)

type Handler struct{ svc *svc.Service }

func New(s *svc.Service) *Handler { return &Handler{svc: s} }

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	action := mux.Vars(r)["action"]
	data, err := h.svc.GetNotifySetting(action)
	if err != nil {
		jsonError(w, err.Error(), 404)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) Upsert(w http.ResponseWriter, r *http.Request) {
	action := mux.Vars(r)["action"]
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.UpsertNotifySetting(action, body)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, data)
}
