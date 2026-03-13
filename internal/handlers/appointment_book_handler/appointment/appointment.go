package appointment

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"sighthub-backend/internal/middleware"
	svc "sighthub-backend/internal/services/appointment_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct {
	svc *svc.Service
	db  *gorm.DB
}

func New(s *svc.Service, db *gorm.DB) *Handler { return &Handler{svc: s, db: db} }

func jsonOK(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func errStatus(msg string) int {
	switch msg {
	case "not found", "patient not found", "location not found", "doctor not found",
		"status not found", "insurance policy not found":
		return http.StatusNotFound
	case "permission denied for this location", "this location is not available for selection",
		"permission denied":
		return http.StatusForbidden
	case "request appointment already processed":
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}

// GET /api/appointment-book/location
func (h *Handler) GetLocations(w http.ResponseWriter, r *http.Request) {
	permIDs := middleware.PermittedLocationIDsFromCtx(r.Context())
	list, err := h.svc.GetLocations(permIDs)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, list)
}

// PUT /api/appointment-book/set_location
func (h *Handler) SetLocation(w http.ResponseWriter, r *http.Request) {
	emp := middleware.EmployeeFromCtx(r.Context())
	if emp == nil {
		jsonError(w, "Employee not found", http.StatusNotFound)
		return
	}
	permIDs := middleware.PermittedLocationIDsFromCtx(r.Context())

	var body struct {
		LocationID int `json:"location_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.LocationID == 0 {
		jsonError(w, "Missing or invalid location_id", http.StatusBadRequest)
		return
	}

	result, err := h.svc.SetLocation(emp.IDEmployee, body.LocationID, permIDs)
	if err != nil {
		jsonError(w, err.Error(), errStatus(err.Error()))
		return
	}
	jsonOK(w, result)
}

// GET /api/appointment-book/doctors
func (h *Handler) GetDoctors(w http.ResponseWriter, r *http.Request) {
	locStr := r.URL.Query().Get("location_id")
	if locStr == "" {
		jsonError(w, "location_id is required", http.StatusBadRequest)
		return
	}
	locationID, err := strconv.Atoi(locStr)
	if err != nil {
		jsonError(w, "invalid location_id", http.StatusBadRequest)
		return
	}
	list, err := h.svc.GetDoctors(locationID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, list)
}

// GET /api/appointment-book/status-appointments
func (h *Handler) GetStatusAppointments(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.GetStatusAppointments()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, list)
}

// GET /api/appointment-book/appointments
func (h *Handler) GetAppointments(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	locStr := q.Get("location_id")
	if locStr == "" {
		jsonError(w, "location_id is required", http.StatusBadRequest)
		return
	}
	locationID, err := strconv.Atoi(locStr)
	if err != nil {
		jsonError(w, "invalid location_id", http.StatusBadRequest)
		return
	}

	startStr := q.Get("start_date")
	if startStr == "" {
		jsonError(w, "Start date is required", http.StatusBadRequest)
		return
	}
	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		jsonError(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	endDate := startDate
	if endStr := q.Get("end_date"); endStr != "" {
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			jsonError(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}
	if endDate.Before(startDate) {
		jsonError(w, "End date must be the same or later than start date", http.StatusBadRequest)
		return
	}

	input := svc.GetAppointmentsInput{
		LocationID: locationID,
		StartDate:  startDate,
		EndDate:    endDate,
	}
	if empStr := q.Get("employee_id"); empStr != "" {
		if empID, err := strconv.ParseInt(empStr, 10, 64); err == nil {
			input.EmployeeID = &empID
		}
	}

	result, err := h.svc.GetAppointments(input)
	if err != nil {
		jsonError(w, err.Error(), errStatus(err.Error()))
		return
	}
	jsonOK(w, result)
}

// GET /api/appointment-book/request-appointments
func (h *Handler) GetRequestAppointments(w http.ResponseWriter, r *http.Request) {
	locStr := r.URL.Query().Get("location_id")
	locationID := 0

	if locStr != "" {
		var err error
		locationID, err = strconv.Atoi(locStr)
		if err != nil {
			jsonError(w, "invalid location_id", http.StatusBadRequest)
			return
		}
	} else {
		username := pkgAuth.UsernameFromContext(r.Context())
		emp, _, err := resolveEmployee(h.db, username)
		if err != nil || emp == nil {
			jsonError(w, "Employee or location not found", http.StatusNotFound)
			return
		}
		if emp.LocationID == nil {
			jsonError(w, "Employee or location not found", http.StatusNotFound)
			return
		}
		locationID = int(*emp.LocationID)
	}

	list, err := h.svc.GetRequestAppointments(locationID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, list)
}

// POST /api/appointment-book/request-appointments/cancel
func (h *Handler) CancelRequestAppointment(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	emp, locID, err := resolveEmployee(h.db, username)
	if err != nil || emp == nil {
		jsonError(w, "Employee or location not found", http.StatusNotFound)
		return
	}

	var body struct {
		IDRequestAppointment *int64 `json:"id_request_appointment"`
		RequestAppointmentID *int64 `json:"request_appointment_id"`
		CancelNote           string `json:"cancel_note"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	raw := body.IDRequestAppointment
	if raw == nil {
		raw = body.RequestAppointmentID
	}
	if raw == nil {
		jsonError(w, "id_request_appointment is required and must be integer", http.StatusBadRequest)
		return
	}

	result, err := h.svc.CancelRequestAppointment(*raw, locID, body.CancelNote)
	if err != nil {
		if err.Error() == "not found" {
			jsonError(w, "Request appointment not found", http.StatusNotFound)
		} else {
			jsonError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	jsonOK(w, result)
}

// POST /api/appointment-book/new-appointment
func (h *Handler) CreateAppointment(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	emp, locID, err := resolveEmployee(h.db, username)
	if err != nil || emp == nil {
		jsonError(w, "Employee not found", http.StatusNotFound)
		return
	}

	var input svc.CreateAppointmentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	input.EmployeeID = emp.IDEmployee
	input.LocID = locID

	result, err := h.svc.CreateAppointment(input)
	if err != nil {
		jsonError(w, err.Error(), errStatus(err.Error()))
		return
	}
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, result)
}

// PUT /api/appointment-book/appointments/{id}
func (h *Handler) UpdateAppointment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid appointment ID", http.StatusBadRequest)
		return
	}
	var input svc.UpdateAppointmentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "No input data provided", http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateAppointment(id, input); err != nil {
		jsonError(w, err.Error(), errStatus(err.Error()))
		return
	}
	jsonOK(w, map[string]string{"message": "Appointment updated successfully"})
}

// PUT /api/appointment-book/appointments/{id}/status
func (h *Handler) UpdateAppointmentStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid appointment ID", http.StatusBadRequest)
		return
	}
	var body struct {
		StatusID *int `json:"status_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.StatusID == nil {
		jsonError(w, "status_id is required", http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateAppointmentStatus(id, *body.StatusID); err != nil {
		jsonError(w, err.Error(), errStatus(err.Error()))
		return
	}
	jsonOK(w, map[string]string{"message": "Appointment status updated successfully"})
}

// PUT /api/appointment-book/appointments/{id}/insurance
func (h *Handler) UpdateAppointmentInsurance(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid appointment ID", http.StatusBadRequest)
		return
	}
	var body struct {
		InsuranceID interface{} `json:"insurance_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "No input data provided", http.StatusBadRequest)
		return
	}

	var insuranceID *int64
	if body.InsuranceID != nil {
		switch v := body.InsuranceID.(type) {
		case float64:
			if v != 0 {
				id64 := int64(v)
				insuranceID = &id64
			}
		}
	}

	if err := h.svc.UpdateAppointmentInsurance(id, insuranceID); err != nil {
		jsonError(w, err.Error(), errStatus(err.Error()))
		return
	}
	jsonOK(w, map[string]string{"message": "Appointment insurance updated successfully"})
}

// DELETE /api/appointment-book/appointments/{id}
func (h *Handler) DeleteAppointment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid appointment ID", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteAppointment(id); err != nil {
		if err.Error() == "not found" {
			jsonError(w, "Appointment not found", http.StatusNotFound)
		} else {
			jsonError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	jsonOK(w, map[string]string{"message": "Appointment deleted successfully."})
}

// GET /api/appointment-book/location/{id}/work_hours
func (h *Handler) GetLocationWorkHours(w http.ResponseWriter, r *http.Request) {
	locationID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		jsonError(w, "Invalid location ID", http.StatusBadRequest)
		return
	}
	q := r.URL.Query()
	startStr := q.Get("start_date")
	if startStr == "" {
		jsonError(w, "Start date is required", http.StatusBadRequest)
		return
	}
	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		jsonError(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	endDate := startDate
	if endStr := q.Get("end_date"); endStr != "" {
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			jsonError(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}
	if endDate.Before(startDate) {
		jsonError(w, "End date must be the same or later than start date", http.StatusBadRequest)
		return
	}

	result, err := h.svc.GetLocationWorkHours(locationID, startDate, endDate)
	if err != nil {
		jsonError(w, err.Error(), errStatus(err.Error()))
		return
	}
	jsonOK(w, result)
}

// POST /api/appointment-book/doctor/{id}/set_lunch
func (h *Handler) SetDoctorLunch(w http.ResponseWriter, r *http.Request) {
	doctorID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid doctor ID", http.StatusBadRequest)
		return
	}

	var body struct {
		LunchStart          string `json:"lunch_start"`
		LunchDurationMinutes int   `json:"lunch_duration_minutes"`
		Date                string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if body.LunchStart == "" {
		jsonError(w, "Lunch start time is required", http.StatusBadRequest)
		return
	}
	lunchStart, err := time.Parse("15:04", body.LunchStart)
	if err != nil {
		jsonError(w, "Invalid lunch start time format. Use HH:MM", http.StatusBadRequest)
		return
	}

	targetDate := time.Now()
	if body.Date != "" {
		targetDate, err = time.Parse("2006-01-02", body.Date)
		if err != nil {
			jsonError(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}
	durationMin := body.LunchDurationMinutes
	if durationMin == 0 {
		durationMin = 30
	}

	result, err := h.svc.SetDoctorLunch(doctorID, targetDate, lunchStart, durationMin)
	if err != nil {
		jsonError(w, err.Error(), errStatus(err.Error()))
		return
	}
	jsonOK(w, result)
}

// POST /api/appointment-book/send_sms/intake_form_link/{id}
func (h *Handler) SendIntakeFormLink(w http.ResponseWriter, r *http.Request) {
	appointmentID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid appointment ID", http.StatusBadRequest)
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	emp, locID, err := resolveEmployee(h.db, username)
	if err != nil || emp == nil {
		jsonError(w, "Employee or location not found", http.StatusNotFound)
		return
	}

	result, err := h.svc.SendIntakeFormSMSForAppointment(appointmentID, emp.IDEmployee, locID)
	if err != nil {
		if err.Error() == "not found" {
			jsonError(w, "Appointment not found", http.StatusNotFound)
		} else if err.Error() == "permission denied" {
			jsonError(w, "Permission denied", http.StatusForbidden)
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	code := http.StatusOK
	if result.Status == "error" {
		code = http.StatusBadRequest
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"appointment_id": appointmentID,
		"status":         result.Status,
		"link":           result.Link,
		"warnings":       result.Warnings,
	})
}

// ─── helper ───────────────────────────────────────────────────────────────────

// resolveEmployee looks up Employee + location from JWT username via DB.
func resolveEmployee(db *gorm.DB, username string) (*empDB, int, error) {
	var login authDB
	if err := db.Table("employee_login").Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, 0, nil
	}
	var emp empDB
	if err := db.Table("employee").Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, 0, nil
	}
	locID := 0
	if emp.LocationID != nil {
		locID = int(*emp.LocationID)
	}
	return &emp, locID, nil
}

type authDB struct {
	IDEmployeeLogin int `gorm:"column:id_employee_login"`
}

type empDB struct {
	IDEmployee int    `gorm:"column:id_employee"`
	FirstName  string `gorm:"column:first_name"`
	LastName   string `gorm:"column:last_name"`
	LocationID *int64 `gorm:"column:location_id"`
}
