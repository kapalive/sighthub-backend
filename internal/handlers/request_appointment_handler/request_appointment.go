package request_appointment_handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	svc "sighthub-backend/internal/services/request_appointment_service"
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

// GET /api/request-appointment/locations
func (h *Handler) GetLocations(w http.ResponseWriter, r *http.Request) {
	locs, err := h.svc.GetLocations()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, locs)
}

// GET /api/request-appointment/doctors?hash=...
func (h *Handler) GetDoctors(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		writeError(w, http.StatusBadRequest, "hash is required")
		return
	}
	doctors, err := h.svc.GetDoctors(hash)
	if err != nil {
		if err.Error() == "no doctors found for this location" {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusNotFound, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, doctors)
}

// GET /api/request-appointment/doctor-slot-availability?doctor_id=...&hash=...&start_date=...
func (h *Handler) GetDoctorSlotAvailability(w http.ResponseWriter, r *http.Request) {
	doctorIDStr := r.URL.Query().Get("doctor_id")
	hash := r.URL.Query().Get("hash")
	if doctorIDStr == "" || hash == "" {
		writeError(w, http.StatusBadRequest, "doctor_id and hash are required")
		return
	}
	doctorID, err := strconv.ParseInt(doctorIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid doctor_id")
		return
	}
	startDate := r.URL.Query().Get("start_date")
	slots, err := h.svc.GetDoctorSlotAvailability(hash, doctorID, startDate)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "location not found" || err.Error() == "doctor not found" {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, slots)
}

// POST /api/request-appointment/request-appointment
func (h *Handler) CreateRequestAppointment(w http.ResponseWriter, r *http.Request) {
	var input svc.CreateRequestAppointmentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := h.svc.CreateRequestAppointment(input)
	if err != nil {
		switch err.Error() {
		case "location not found", "doctor not found for this location", "patient not found", "insurance policy not found":
			writeError(w, http.StatusNotFound, err.Error())
		case "request appointment disabled for this location":
			writeError(w, http.StatusForbidden, err.Error())
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

// POST /api/request-appointment/intake-form
func (h *Handler) CreateIntakeForm(w http.ResponseWriter, r *http.Request) {
	var input svc.IntakeFormInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id, err := h.svc.CreateIntakeForm(input)
	if err != nil {
		switch err.Error() {
		case "location not found", "appointment not found":
			writeError(w, http.StatusNotFound, err.Error())
		case "intake form disabled for this location":
			writeError(w, http.StatusForbidden, err.Error())
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"message": "Intake form submitted", "id": id})
}

// GET /api/request-appointment/intake-form/{id}
func (h *Handler) GetIntakeForm(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	result, err := h.svc.GetIntakeForm(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// PUT /api/request-appointment/intake-form/{id}
func (h *Handler) UpdateIntakeForm(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var input svc.IntakeFormInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateIntakeForm(id, input); err != nil {
		if err.Error() == "intake form not found" {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"message": "Intake form updated", "id": id})
}

// GET /api/request-appointment/check-appointment/{id}
func (h *Handler) CheckAppointment(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.CheckAppointment(id); err != nil {
		switch err.Error() {
		case "appointment not found":
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Appointment is upcoming"})
}
