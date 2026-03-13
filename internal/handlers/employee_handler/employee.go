package employee_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	svc "sighthub-backend/internal/services/employee_service"
	pkgAuth "sighthub-backend/pkg/auth"
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

// GET /api/employee/login/availability?username=
func (h *Handler) CheckLoginAvailability(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		writeError(w, http.StatusBadRequest, "username parameter is required")
		return
	}
	available, err := h.svc.CheckLoginAvailability(username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"available": available})
}

// GET /api/employee/location
func (h *Handler) GetLocations(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetLocations()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GET /api/employee/warehouses/{location_id}
func (h *Handler) GetWarehousesByLocation(w http.ResponseWriter, r *http.Request) {
	lid, err := strconv.Atoi(mux.Vars(r)["location_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid location_id")
		return
	}
	result, err := h.svc.GetWarehousesByLocation(lid)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GET /api/employee/roles
func (h *Handler) GetRoles(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetRoles()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GET /api/employee/
func (h *Handler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.ListEmployees()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GET /api/employee/employee/general
func (h *Handler) GetEmployeeGeneral(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	if username == "" {
		writeError(w, http.StatusBadRequest, "cannot extract username from token")
		return
	}
	result, err := h.svc.GetEmployeeGeneral(username)
	if err != nil {
		if err.Error() == "user not found" || err.Error() == "employee not found" {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GET /api/employee/{employee_id}
func (h *Handler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	eid, err := strconv.Atoi(mux.Vars(r)["employee_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	result, err := h.svc.GetEmployee(eid)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// POST /api/employee/add
func (h *Handler) AddEmployee(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	var input svc.AddEmployeeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id, err := h.svc.AddEmployee(username, input)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			writeError(w, http.StatusBadRequest, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":     "Employee added successfully",
		"id_employee": id,
	})
}

// PUT /api/employee/{employee_id}
func (h *Handler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	eid, err := strconv.Atoi(mux.Vars(r)["employee_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	var input svc.UpdateEmployeeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateEmployee(username, eid, input); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Employee updated successfully"})
}

// GET /api/employee/time_card
func (h *Handler) ListTimecards(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "true"
	result, err := h.svc.ListTimecards(activeOnly)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GET /api/employee/time_card/{timecard_login_id}
func (h *Handler) GetTimecardHistory(w http.ResponseWriter, r *http.Request) {
	tcID, err := strconv.Atoi(mux.Vars(r)["timecard_login_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid timecard_login_id")
		return
	}
	start, end := parseTimecardDateRange(r)
	result, err := h.svc.GetTimecardHistory(tcID, start, end, true)
	if err != nil {
		if err.Error() == "timecard account not found" {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// GET /api/employee/{employee_id}/time_card
func (h *Handler) GetEmployeeTimecardHistory(w http.ResponseWriter, r *http.Request) {
	eid, err := strconv.Atoi(mux.Vars(r)["employee_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	start, end := parseTimecardDateRange(r)
	result, err := h.svc.GetEmployeeTimecardHistory(eid, start, end)
	if err != nil {
		if err.Error() == "timecard account not found" || err.Error() == "employee not found" {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// POST /api/employee/time_card  OR  POST /api/employee/{employee_id}/time_card
func (h *Handler) CreateTimecard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var employeeID *int
	if eidStr, ok := vars["employee_id"]; ok && eidStr != "" {
		eid, err := strconv.Atoi(eidStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid employee_id")
			return
		}
		employeeID = &eid
	}

	var body struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Username == "" || body.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}
	if employeeID == nil && (body.FirstName == "" || body.LastName == "") {
		writeError(w, http.StatusBadRequest, "first_name and last_name are required")
		return
	}
	id, err := h.svc.CreateTimecard(employeeID, body.Username, body.Password, body.FirstName, body.LastName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Timecard login created successfully",
		"id":      id,
	})
}

// PUT /api/employee/{employee_id}/time_card
func (h *Handler) UpdateTimecard(w http.ResponseWriter, r *http.Request) {
	eid, err := strconv.Atoi(mux.Vars(r)["employee_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	var body struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateTimecard(eid, body.Username, body.Password, body.FirstName, body.LastName); err != nil {
		if err.Error() == "timecard account not found" || err.Error() == "employee not found" {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Timecard login updated successfully"})
}

// POST /api/employee/deactivate/{login_id}
func (h *Handler) DeactivateTimecard(w http.ResponseWriter, r *http.Request) {
	lid, err := strconv.Atoi(mux.Vars(r)["login_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid login_id")
		return
	}
	if err := h.svc.DeactivateTimecard(lid); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "User deactivated"})
}

// GET /api/employee/{employee_id}/schedule
func (h *Handler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	eid, err := strconv.Atoi(mux.Vars(r)["employee_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	var startDate, endDate *time.Time
	if s := r.URL.Query().Get("start_date"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid start_date format. Use YYYY-MM-DD")
			return
		}
		startDate = &t
		endDate = &t
	}
	if s := r.URL.Query().Get("end_date"); s != "" && startDate != nil {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid end_date format. Use YYYY-MM-DD")
			return
		}
		endDate = &t
	}
	result, err := h.svc.GetSchedule(eid, startDate, endDate)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// POST /api/employee/{employee_id}/schedule
func (h *Handler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	eid, err := strconv.Atoi(mux.Vars(r)["employee_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		data = map[string]interface{}{}
	}
	if err := h.svc.CreateSchedule(eid, data); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "Schedule saved"})
}

// PUT /api/employee/{employee_id}/schedule
func (h *Handler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	eid, err := strconv.Atoi(mux.Vars(r)["employee_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		data = map[string]interface{}{}
	}
	result, err := h.svc.UpdateSchedule(eid, data)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// POST /api/employee/{employee_id}/off_day
func (h *Handler) AddOffDay(w http.ResponseWriter, r *http.Request) {
	eid, err := strconv.Atoi(mux.Vars(r)["employee_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	var body struct {
		Date string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Date == "" {
		writeError(w, http.StatusBadRequest, "date is required")
		return
	}
	dateStr, err := h.svc.AddOffDay(eid, body.Date)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "Off day added", "date": dateStr})
}

// GET /api/employee/{employee_id}/off_day
func (h *Handler) ListOffDays(w http.ResponseWriter, r *http.Request) {
	eid, err := strconv.Atoi(mux.Vars(r)["employee_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	result, err := h.svc.ListOffDays(eid)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// DELETE /api/employee/{employee_id}/off_day/{date_str}
func (h *Handler) RemoveOffDay(w http.ResponseWriter, r *http.Request) {
	eid, err := strconv.Atoi(mux.Vars(r)["employee_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}
	dateStr := mux.Vars(r)["date_str"]
	removed, err := h.svc.RemoveOffDay(eid, dateStr)
	if err != nil {
		if err.Error() == "employee not found" || err.Error() == "off day not found" {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Off day removed", "date": removed})
}

// GET /job-titles
func (h *Handler) GetJobTitles(w http.ResponseWriter, r *http.Request) {
	titles, err := h.svc.GetJobTitles()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, titles)
}

func parseTimecardDateRange(r *http.Request) (time.Time, time.Time) {
	today := time.Now()
	start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 0, time.UTC)

	if s := r.URL.Query().Get("start_date"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			start = t
		}
	}
	if e := r.URL.Query().Get("end_date"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			end = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.UTC)
		}
	}
	return start, end
}
