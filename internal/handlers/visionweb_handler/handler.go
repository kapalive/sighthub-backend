package visionweb_handler

import (
	"encoding/json"
	"net/http"
	"sync"

	"sighthub-backend/internal/services/visionweb_service"
)

type Handler struct {
	svc *visionweb_service.Service
	mu  sync.Mutex

	// last import result (async)
	importing bool
	lastResult *visionweb_service.ImportAllResult
	lastError  string
}

func NewHandler(svc *visionweb_service.Service) *Handler {
	return &Handler{svc: svc}
}

// POST /api/price_book/import-vw
// Starts async import, returns 202 immediately
func (h *Handler) ImportFromVisionWeb(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	if h.importing {
		h.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "import already in progress"})
		return
	}
	h.importing = true
	h.lastResult = nil
	h.lastError = ""
	h.mu.Unlock()

	go func() {
		result, err := h.svc.ImportAllVisionWeb()
		h.mu.Lock()
		h.importing = false
		if err != nil {
			h.lastError = err.Error()
		} else {
			h.lastResult = result
		}
		h.mu.Unlock()
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"message": "import started"})
}

// GET /api/price_book/import-vw/status
func (h *Handler) ImportStatus(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")

	if h.importing {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "importing",
		})
		return
	}

	if h.lastError != "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  h.lastError,
		})
		return
	}

	if h.lastResult != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "done",
			"result": h.lastResult,
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "idle",
	})
}
