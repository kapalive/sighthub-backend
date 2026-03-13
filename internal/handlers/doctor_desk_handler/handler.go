package doctor_desk_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	ddSvc "sighthub-backend/internal/services/doctor_desk_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct{ svc *ddSvc.Service }

func New(svc *ddSvc.Service) *Handler { return &Handler{svc: svc} }

// ─── helpers ──────────────────────────────────────────────────────────────────

func jsonResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func parsePathInt64(r *http.Request, key string) (int64, error) {
	return strconv.ParseInt(mux.Vars(r)[key], 10, 64)
}

func parsePathInt(r *http.Request, key string) (int, error) {
	return strconv.Atoi(mux.Vars(r)[key])
}

// ─── GetAppointments  GET /api/doctor-desk/appointments ──────────────────────

func (h *Handler) GetAppointments(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	// Resolve location_id
	locationID, err := h.svc.GetEmployeeLocation(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if v := r.URL.Query().Get("location_id"); v != "" {
		if id, e := strconv.Atoi(v); e == nil {
			locationID = id
		}
	}

	// Resolve start_date
	date := time.Now().Truncate(24 * time.Hour)
	if v := r.URL.Query().Get("start_date"); v != "" {
		if t, e := time.Parse("2006-01-02", v); e == nil {
			date = t
		} else {
			http.Error(w, "invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	// Optional employee_id
	var employeeID *int64
	if v := r.URL.Query().Get("employee_id"); v != "" {
		if id, e := strconv.ParseInt(v, 10, 64); e == nil {
			employeeID = &id
		}
	}

	result, err := h.svc.GetAppointments(locationID, date, employeeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, []interface{}{result}, http.StatusOK)
}

// ─── SearchPatients  GET /api/doctor-desk/patient-search ─────────────────────

func (h *Handler) SearchPatients(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	firstName := strings.TrimSpace(r.URL.Query().Get("first_name"))
	lastName := strings.TrimSpace(r.URL.Query().Get("last_name"))
	chart := strings.TrimSpace(r.URL.Query().Get("chart"))

	patients, err := h.svc.SearchPatients(username, firstName, lastName, chart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if len(patients) == 0 {
		jsonResponse(w, map[string]interface{}{
			"patients": []interface{}{},
			"message":  "No patient found with the provided parameters",
		}, http.StatusOK)
		return
	}
	jsonResponse(w, map[string]interface{}{"patients": patients}, http.StatusOK)
}

// ─── UpdateAppointmentStatus  PUT /api/doctor-desk/appointment/status/{id} ───

func (h *Handler) UpdateAppointmentStatus(w http.ResponseWriter, r *http.Request) {
	appointmentID, err := parsePathInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid appointment id", http.StatusBadRequest)
		return
	}

	var body struct {
		StatusID *int `json:"status_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.StatusID == nil {
		http.Error(w, "status_id is required", http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdateAppointmentStatus(appointmentID, *body.StatusID); err != nil {
		if err.Error() == "appointment not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"message": "Appointment status updated successfully"}, http.StatusOK)
}

// ─── GetDoctors  GET /api/doctor-desk/doctors ─────────────────────────────────

func (h *Handler) GetDoctors(w http.ResponseWriter, r *http.Request) {
	locationIDStr := r.URL.Query().Get("location_id")
	if locationIDStr == "" {
		http.Error(w, "location_id is required", http.StatusBadRequest)
		return
	}
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		http.Error(w, "invalid location_id", http.StatusBadRequest)
		return
	}

	doctors, err := h.svc.GetDoctors(locationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, doctors, http.StatusOK)
}

// ─── GetShowcaseLocations  GET /api/doctor-desk/location ──────────────────────

func (h *Handler) GetShowcaseLocations(w http.ResponseWriter, r *http.Request) {
	locs, err := h.svc.GetShowcaseLocations()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, locs, http.StatusOK)
}

// ─── GetPatientExams  GET /api/doctor-desk/patient/{id}/exams ─────────────────

func (h *Handler) GetPatientExams(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	patientID, err := parsePathInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid patient id", http.StatusBadRequest)
		return
	}

	exams, err := h.svc.GetPatientExams(username, patientID)
	if err != nil {
		if err.Error() == "employee or location not found" || err.Error() == "patient not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]interface{}{"exams": exams}, http.StatusOK)
}

// ─── GetUnsignedExams  GET /api/doctor-desk/exams/unsigned ───────────────────

func (h *Handler) GetUnsignedExams(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var locationID *int
	if v := r.URL.Query().Get("location_id"); v != "" {
		if id, e := strconv.Atoi(v); e == nil {
			locationID = &id
		}
	}

	exams, err := h.svc.GetUnsignedExams(username, locationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, map[string]interface{}{"exams": exams}, http.StatusOK)
}

// ─── File handlers ────────────────────────────────────────────────────────────

func (h *Handler) GetPatientFiles(w http.ResponseWriter, r *http.Request) {
	patientID, err := parsePathInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid patient id", http.StatusBadRequest)
		return
	}
	files, err := h.svc.GetPatientFiles(patientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, map[string]interface{}{"files": files}, http.StatusOK)
}

func (h *Handler) UploadExamFile(w http.ResponseWriter, r *http.Request) {
	patientID, err := parsePathInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid patient id", http.StatusBadRequest)
		return
	}

	var body struct {
		FilePath string `json:"file_path"`
		FileName string `json:"file_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.FilePath == "" {
		http.Error(w, "file_path is required", http.StatusBadRequest)
		return
	}

	fileID, err := h.svc.UploadExamFile(patientID, body.FilePath, body.FileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]interface{}{
		"message": "File uploaded and metadata saved successfully",
		"file_id": fileID,
	}, http.StatusCreated)
}

func (h *Handler) UpdateExamFile(w http.ResponseWriter, r *http.Request) {
	fileID, err := parsePathInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid file id", http.StatusBadRequest)
		return
	}

	var body struct {
		FileName string `json:"file_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.FileName == "" {
		http.Error(w, "file_name is required", http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdateExamFile(fileID, body.FileName); err != nil {
		if err.Error() == "file not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"message": "File information updated successfully"}, http.StatusOK)
}

func (h *Handler) GetExamFile(w http.ResponseWriter, r *http.Request) {
	fileID, err := parsePathInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid file id", http.StatusBadRequest)
		return
	}
	file, err := h.svc.GetExamFile(fileID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, file, http.StatusOK)
}

func (h *Handler) DeleteExamFile(w http.ResponseWriter, r *http.Request) {
	fileID, err := parsePathInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid file id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteExamFile(fileID); err != nil {
		if err.Error() == "file not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]string{"message": "File deleted successfully"}, http.StatusOK)
}

// ─── Patient Notes handlers ───────────────────────────────────────────────────

func (h *Handler) GetPatientNotes(w http.ResponseWriter, r *http.Request) {
	patientID, err := parsePathInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid patient id", http.StatusBadRequest)
		return
	}
	notes, err := h.svc.GetPatientNotes(patientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, map[string]interface{}{"notes": notes}, http.StatusOK)
}

func (h *Handler) CreatePatientNote(w http.ResponseWriter, r *http.Request) {
	patientID, err := parsePathInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid patient id", http.StatusBadRequest)
		return
	}

	var body struct {
		Note      string  `json:"note"`
		Top       bool    `json:"top"`
		AlertDate *string `json:"alert_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.Note == "" {
		http.Error(w, "note text is required", http.StatusBadRequest)
		return
	}

	var alertDate *time.Time
	if body.AlertDate != nil && *body.AlertDate != "" {
		t, err := time.Parse("2006-01-02", *body.AlertDate)
		if err != nil {
			http.Error(w, "invalid alert_date format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		alertDate = &t
	}

	noteID, err := h.svc.CreatePatientNote(patientID, body.Note, body.Top, alertDate)
	if err != nil {
		if err.Error() == "patient not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]interface{}{
		"message": "Patient note created successfully",
		"note_id": noteID,
	}, http.StatusCreated)
}

func (h *Handler) UpdatePatientNote(w http.ResponseWriter, r *http.Request) {
	patientID, err := parsePathInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid patient id", http.StatusBadRequest)
		return
	}
	noteID, err := parsePathInt64(r, "note_id")
	if err != nil {
		http.Error(w, "invalid note id", http.StatusBadRequest)
		return
	}

	var body struct {
		Note      *string `json:"note"`
		Top       *bool   `json:"top"`
		AlertDate *string `json:"alert_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// alert_date: present and non-null → parse; absent → clear (Python sets to None if not provided)
	var alertDate *time.Time
	clearAlertDate := body.AlertDate == nil
	if body.AlertDate != nil && *body.AlertDate != "" {
		t, err := time.Parse("2006-01-02", *body.AlertDate)
		if err != nil {
			http.Error(w, "invalid alert_date format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		alertDate = &t
		clearAlertDate = false
	}

	if err := h.svc.UpdatePatientNote(patientID, noteID, body.Note, body.Top, alertDate, clearAlertDate); err != nil {
		switch err.Error() {
		case "patient not found", "note not found for this patient":
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonResponse(w, map[string]string{"message": "Patient note updated successfully"}, http.StatusOK)
}

func (h *Handler) GetPatientNote(w http.ResponseWriter, r *http.Request) {
	patientID, err := parsePathInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid patient id", http.StatusBadRequest)
		return
	}
	noteID, err := parsePathInt64(r, "note_id")
	if err != nil {
		http.Error(w, "invalid note id", http.StatusBadRequest)
		return
	}
	note, err := h.svc.GetPatientNote(patientID, noteID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, note, http.StatusOK)
}

func (h *Handler) DeletePatientNote(w http.ResponseWriter, r *http.Request) {
	patientID, err := parsePathInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid patient id", http.StatusBadRequest)
		return
	}
	noteID, err := parsePathInt64(r, "note_id")
	if err != nil {
		http.Error(w, "invalid note id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeletePatientNote(patientID, noteID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, map[string]string{"message": "Patient note deleted successfully"}, http.StatusOK)
}

// ─── Patient Info handlers ────────────────────────────────────────────────────

func (h *Handler) GetPatientInfo(w http.ResponseWriter, r *http.Request) {
	patientID, err := parsePathInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid patient id", http.StatusBadRequest)
		return
	}
	info, err := h.svc.GetPatientInfo(patientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, info, http.StatusOK)
}

func (h *Handler) UpdatePatientInfo(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	patientID, err := parsePathInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid patient id", http.StatusBadRequest)
		return
	}

	var body struct {
		FrontDeskNote *string `json:"front_desk_note"`
		ExpireDate    *string `json:"expire_date"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	var expireDate *time.Time
	if body.ExpireDate != nil && *body.ExpireDate != "" {
		t, err := time.Parse("2006-01-02", *body.ExpireDate)
		if err != nil {
			http.Error(w, "invalid expire_date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		expireDate = &t
	}

	if err := h.svc.UpdatePatientInfo(username, patientID, body.FrontDeskNote, expireDate); err != nil {
		switch err.Error() {
		case "employee or location not found", "patient not found",
			"patient has no eye exams in this location",
			"recall information not found":
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonResponse(w, map[string]string{"message": "Patient recall updated successfully"}, http.StatusOK)
}

// ─── LogCall  POST /api/doctor-desk/{id}/log-call ─────────────────────────────

func (h *Handler) LogCall(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	patientID, err := parsePathInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid patient id", http.StatusBadRequest)
		return
	}

	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Content == "" {
		http.Error(w, "call note content is required", http.StatusBadRequest)
		return
	}

	if err := h.svc.LogCall(username, patientID, body.Content); err != nil {
		switch err.Error() {
		case "employee or location not found", "patient not found",
			"call communication type not found":
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonResponse(w, map[string]string{"message": "Call logged successfully"}, http.StatusCreated)
}

// ─── Medical history handlers ─────────────────────────────────────────────────

func (h *Handler) GetMedications(w http.ResponseWriter, r *http.Request) {
	patientID, err := parsePathInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid patient id", http.StatusBadRequest)
		return
	}
	meds, err := h.svc.GetMedications(patientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, meds, http.StatusOK)
}

func (h *Handler) GetAllergies(w http.ResponseWriter, r *http.Request) {
	patientID, err := parsePathInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid patient id", http.StatusBadRequest)
		return
	}
	allergies, err := h.svc.GetAllergies(patientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, allergies, http.StatusOK)
}

func (h *Handler) GetDiagnoses(w http.ResponseWriter, r *http.Request) {
	patientID, err := parsePathInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid patient id", http.StatusBadRequest)
		return
	}
	diagnoses, err := h.svc.GetDiagnoses(patientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, diagnoses, http.StatusOK)
}
