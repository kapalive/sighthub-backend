package payment

import (
	"encoding/json"
	"net/http"
	"strconv"

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

func pathID(r *http.Request) int {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	return id
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListPaymentMethods()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MethodName string  `json:"method_name"`
		Category   string  `json:"category"`
		ShortName  *string `json:"short_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	if body.MethodName == "" || body.Category == "" {
		jsonError(w, "method_name and category are required", 400)
		return
	}
	data, err := h.svc.CreatePaymentMethod(body.MethodName, body.Category, body.ShortName)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.UpdatePaymentMethod(pathID(r), body)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeletePaymentMethod(pathID(r)); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted"})
}
