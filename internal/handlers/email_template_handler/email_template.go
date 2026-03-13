package email_template_handler

import (
	"encoding/json"
	"net/http"

	svc "sighthub-backend/internal/services/email_template_service"
)

type Handler struct {
	svc *svc.Service
}

func New(s *svc.Service) *Handler {
	return &Handler{svc: s}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// GET /api/email-template/
func (h *Handler) GetAllTemplates(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetAllTemplates()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// POST /api/email-template/settings
func (h *Handler) SetOrgTemplate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TemplateID int `json:"template_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.TemplateID == 0 {
		writeError(w, http.StatusBadRequest, "template_id is required")
		return
	}
	result, err := h.svc.SetOrgTemplate(body.TemplateID)
	if err != nil {
		if err.Error() == "template not found" {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, result)
}
