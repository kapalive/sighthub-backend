package license_handler

import (
	"encoding/json"
	"net/http"

	svc "sighthub-backend/internal/services/license_service"
)

type Handler struct{ svc *svc.Service }

func New(s *svc.Service) *Handler { return &Handler{svc: s} }

func jsonOK(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// POST /api/license/kms/store
func (h *Handler) KMSStore(w http.ResponseWriter, r *http.Request) {
	var input svc.KMSStoreInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.svc.KMSStore(input)
	if err != nil {
		msg := err.Error()
		switch msg {
		case "hash store is required":
			jsonError(w, msg, http.StatusBadRequest)
		case "store not found", "location not found":
			jsonError(w, msg, http.StatusNotFound)
		default:
			jsonError(w, msg, http.StatusInternalServerError)
		}
		return
	}
	jsonOK(w, result)
}
