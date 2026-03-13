package info

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	infoSvc "sighthub-backend/internal/services/patient_service/info"
	pkgAuth "sighthub-backend/pkg/auth"
)

// ─── Handler ──────────────────────────────────────────────────────────────────

type Handler struct {
	svc *infoSvc.Service
}

func New(svc *infoSvc.Service) *Handler { return &Handler{svc: svc} }

// ─── Utils ────────────────────────────────────────────────────────────────────

func jsonResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data) //nolint:errcheck
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}

func parsePatientID(r *http.Request) (int64, error) {
	v := mux.Vars(r)["patient_id"]
	id, err := strconv.ParseInt(v, 10, 64)
	return id, err
}

func cleanParam(v string) string {
	return v
}

func optStr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

// ─── GET /languages ───────────────────────────────────────────────────────────

func (h *Handler) GetLanguages(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.GetLanguages()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, list, http.StatusOK)
}

// ─── GET /country_codes ───────────────────────────────────────────────────────

func (h *Handler) GetCountryCodes(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.GetCountryCodes()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, list, http.StatusOK)
}

// ─── POST /add ────────────────────────────────────────────────────────────────

func (h *Handler) CreatePatient(w http.ResponseWriter, r *http.Request) {
	var input infoSvc.CreatePatientInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.CreatePatient(username, input)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "employee or location not found" {
			status = http.StatusNotFound
		}
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, result, http.StatusCreated)
}

// ─── GET /{patient_id} ────────────────────────────────────────────────────────

func (h *Handler) GetPatient(w http.ResponseWriter, r *http.Request) {
	patientID, err := parsePatientID(r)
	if err != nil {
		jsonError(w, "invalid patient_id", http.StatusBadRequest)
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	patient, err := h.svc.GetPatient(username, patientID)
	if err != nil {
		switch err.Error() {
		case "employee or location not found":
			jsonError(w, err.Error(), http.StatusNotFound)
		case "patient not found":
			jsonError(w, err.Error(), http.StatusNotFound)
		case "forbidden":
			jsonError(w, "Forbidden", http.StatusForbidden)
		default:
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonResponse(w, patient, http.StatusOK)
}

// ─── PUT /{patient_id} ────────────────────────────────────────────────────────

func (h *Handler) UpdatePatient(w http.ResponseWriter, r *http.Request) {
	patientID, err := parsePatientID(r)
	if err != nil {
		jsonError(w, "invalid patient_id", http.StatusBadRequest)
		return
	}

	var input infoSvc.UpdatePatientInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdatePatient(patientID, input); err != nil {
		status := http.StatusBadRequest
		if err.Error() == "patient not found" {
			status = http.StatusNotFound
		}
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]string{
		"status":  "success",
		"message": "Patient profile updated successfully.",
	}, http.StatusOK)
}

// ─── DELETE /remove/{patient_id} ──────────────────────────────────────────────

func (h *Handler) DeletePatient(w http.ResponseWriter, r *http.Request) {
	patientID, err := parsePatientID(r)
	if err != nil {
		jsonError(w, "invalid patient_id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeletePatient(patientID); err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── GET /{patient_id}/generate-filename-doc ──────────────────────────────────

func (h *Handler) GenerateFilenameDoc(w http.ResponseWriter, r *http.Request) {
	patientID, err := parsePatientID(r)
	if err != nil {
		jsonError(w, "invalid patient_id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GenerateFilenameDoc(patientID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── GET /{patient_id}/generate-filename-prescription ────────────────────────

func (h *Handler) GenerateFilenamePrescription(w http.ResponseWriter, r *http.Request) {
	patientID, err := parsePatientID(r)
	if err != nil {
		jsonError(w, "invalid patient_id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GenerateFilenamePrescription(patientID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── GET /{patient_id}/generate-filename-insurance-policy ────────────────────

func (h *Handler) GenerateFilenameInsurancePolicy(w http.ResponseWriter, r *http.Request) {
	patientID, err := parsePatientID(r)
	if err != nil {
		jsonError(w, "invalid patient_id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GenerateFilenameInsurancePolicy(patientID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── GET /patient/search ──────────────────────────────────────────────────────

func (h *Handler) SearchPatients(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	params := infoSvc.SearchParams{
		FirstName: optStr(cleanParam(q.Get("first_name"))),
		LastName:  optStr(cleanParam(q.Get("last_name"))),
		DOB:       optStr(cleanParam(q.Get("dob"))),
		City:      optStr(cleanParam(q.Get("city"))),
		State:     optStr(cleanParam(q.Get("state"))),
		Phone:     optStr(cleanParam(q.Get("phone"))),
		Email:     optStr(cleanParam(q.Get("email"))),
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	result, err := h.svc.SearchPatients(username, params)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "employee or location not found" {
			status = http.StatusNotFound
		} else if err.Error() == "dob must be YYYY-MM-DD" {
			status = http.StatusBadRequest
		}
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── GET /{patient_id}/prescriptions ─────────────────────────────────────────

func (h *Handler) GetPatientPrescriptions(w http.ResponseWriter, r *http.Request) {
	patientID, err := parsePatientID(r)
	if err != nil {
		jsonError(w, "invalid patient_id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GetPatientPrescriptions(patientID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── GET /appointments ────────────────────────────────────────────────────────

func (h *Handler) GetPatientAppointments(w http.ResponseWriter, r *http.Request) {
	patientIDStr := r.URL.Query().Get("id_patient")
	if patientIDStr == "" {
		jsonError(w, "id_patient is required", http.StatusBadRequest)
		return
	}
	patientID, err := strconv.ParseInt(patientIDStr, 10, 64)
	if err != nil {
		jsonError(w, "invalid id_patient", http.StatusBadRequest)
		return
	}

	result, err := h.svc.GetPatientAppointments(patientID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}
