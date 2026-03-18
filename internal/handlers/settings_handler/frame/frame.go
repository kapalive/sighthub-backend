package frame

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

// ── Frame Shapes ───────────────────────────────────────────────────────────

func (h *Handler) ListShapes(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListFrameShapes()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) CreateShape(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title       string  `json:"frame_shape_name"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.CreateFrameShape(body.Title, body.Description)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) UpdateShape(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	if err := h.svc.UpdateFrameShape(pathID(r), body); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Updated"})
}

func (h *Handler) DeleteShape(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteFrameShape(pathID(r)); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted"})
}

// ── Frame Type Materials ───────────────────────────────────────────────────

func (h *Handler) ListTypeMaterials(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListFrameTypeMaterials()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) CreateTypeMaterial(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Material string `json:"material"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.CreateFrameTypeMaterial(body.Material)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) UpdateTypeMaterial(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Material string `json:"material"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	if err := h.svc.UpdateFrameTypeMaterial(pathID(r), body.Material); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Updated"})
}

func (h *Handler) DeleteTypeMaterial(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteFrameTypeMaterial(pathID(r)); err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted"})
}
