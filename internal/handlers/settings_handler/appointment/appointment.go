package appointment

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

// ── Appointment Reasons ────────────────────────────────────────────────────

func (h *Handler) ListReasons(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.ListAppointmentReasons()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) CreateReason(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Reason      string  `json:"reason"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	if body.Reason == "" {
		jsonError(w, "reason is required", 400)
		return
	}
	data, err := h.svc.CreateAppointmentReason(body.Reason, body.Description)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) UpdateReason(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.UpdateAppointmentReason(id, body)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) DeleteReason(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if err := h.svc.DeleteAppointmentReason(id); err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, map[string]string{"message": "Deleted"})
}

// ── Locations (showcase) ───────────────────────────────────────────────────

func (h *Handler) GetShowcaseLocations(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetShowcaseLocations()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

// ── Location Appointment Settings ──────────────────────────────────────────

func (h *Handler) SetRequestAppointment(w http.ResponseWriter, r *http.Request) {
	var body struct {
		LocationID int  `json:"location_id"`
		Enabled    bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.SetRequestAppointment(body.LocationID, body.Enabled)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) GetRequestAppointmentSettings(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetRequestAppointmentSettings()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) SetIntakeForm(w http.ResponseWriter, r *http.Request) {
	var body struct {
		LocationID int  `json:"location_id"`
		Enabled    bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	data, err := h.svc.SetIntakeForm(body.LocationID, body.Enabled)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) GetIntakeFormSettings(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetIntakeFormSettings()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonOK(w, data)
}

func (h *Handler) SetAppointmentDuration(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Duration int `json:"appointment_duration"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid JSON", 400)
		return
	}
	dur, err := h.svc.SetAppointmentDuration(body.Duration)
	if err != nil {
		jsonError(w, err.Error(), 400)
		return
	}
	jsonOK(w, map[string]interface{}{"appointment_duration": dur})
}

func (h *Handler) GetAppointmentDuration(w http.ResponseWriter, r *http.Request) {
	dur, err := h.svc.GetAppointmentDuration()
	if err != nil {
		jsonError(w, err.Error(), 404)
		return
	}
	jsonOK(w, map[string]interface{}{"appointment_duration": dur})
}
