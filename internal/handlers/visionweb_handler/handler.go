package visionweb_handler

import (
	"encoding/json"
	"net/http"

	"sighthub-backend/internal/services/visionweb_service"
)

type Handler struct {
	svc *visionweb_service.Service
}

func NewHandler(svc *visionweb_service.Service) *Handler {
	return &Handler{svc: svc}
}

// POST /api/price_book/import-vw
// No body — imports all labs configured with source='vision_web' in lab table
func (h *Handler) ImportFromVisionWeb(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.ImportAllVisionWeb()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
