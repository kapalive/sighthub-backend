package recall

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/general"
	locModel "sighthub-backend/internal/models/location"
	"sighthub-backend/internal/models/patients"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct{ DB *gorm.DB }

func New(db *gorm.DB) *Handler { return &Handler{DB: db} }

func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data) //nolint:errcheck
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}

func (h *Handler) getEmployeeAndLocation(username string) (*empModel.Employee, *locModel.Location, error) {
	var login authModel.EmployeeLogin
	if err := h.DB.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, nil, err
	}
	var emp empModel.Employee
	if err := h.DB.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, nil, err
	}
	if emp.LocationID == nil {
		return &emp, nil, nil
	}
	var loc locModel.Location
	if err := h.DB.First(&loc, *emp.LocationID).Error; err != nil {
		return &emp, nil, err
	}
	return &emp, &loc, nil
}

// GET /api/patient/{patient_id}/recall
func (h *Handler) GetRecalls(w http.ResponseWriter, r *http.Request) {
	idPatient, err := strconv.ParseInt(mux.Vars(r)["patient_id"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid patient_id", http.StatusBadRequest)
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	_, loc, err := h.getEmployeeAndLocation(username)
	if err != nil || loc == nil {
		jsonError(w, "Employee or location not found", http.StatusNotFound)
		return
	}

	// Check patient belongs to location
	var count int64
	h.DB.Model(&patients.Patient{}).
		Where("id_patient = ? AND location_id = ?", idPatient, loc.IDLocation).Count(&count)
	if count == 0 {
		jsonError(w, "Patient not found", http.StatusNotFound)
		return
	}

	var recalls []general.PlaningCommunication
	h.DB.Where("patient_id = ? AND location_id = ? AND date::date >= CURRENT_DATE",
		idPatient, loc.IDLocation).
		Order("date ASC").Find(&recalls)

	result := make([]map[string]interface{}, 0, len(recalls))
	for _, rc := range recalls {
		var dateStr interface{}
		if !rc.Date.IsZero() {
			dateStr = rc.Date.Format(time.RFC3339)
		}
		result = append(result, map[string]interface{}{
			"recall_id":   rc.IDPlaningCommunication,
			"date":        dateStr,
			"reason":      rc.Reason,
			"note":        rc.Note,
			"location_id": rc.LocationID,
		})
	}

	jsonOK(w, map[string]interface{}{"recalls": result})
}

// POST /api/patient/{patient_id}/recall
func (h *Handler) CreateRecall(w http.ResponseWriter, r *http.Request) {
	idPatient, err := strconv.ParseInt(mux.Vars(r)["patient_id"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid patient_id", http.StatusBadRequest)
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	_, loc, err := h.getEmployeeAndLocation(username)
	if err != nil || loc == nil {
		jsonError(w, "Employee or location not found", http.StatusNotFound)
		return
	}

	var count int64
	h.DB.Model(&patients.Patient{}).
		Where("id_patient = ? AND location_id = ?", idPatient, loc.IDLocation).Count(&count)
	if count == 0 {
		jsonError(w, "Patient not found", http.StatusNotFound)
		return
	}

	var input struct {
		Reason *string `json:"reason"`
		Note   *string `json:"note"`
		Date   string  `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if input.Date == "" {
		jsonError(w, "date is required", http.StatusBadRequest)
		return
	}

	recallDt, err := time.Parse(time.RFC3339, input.Date)
	if err != nil {
		recallDt, err = time.Parse("2006-01-02T15:04:05", input.Date)
		if err != nil {
			recallDt, err = time.Parse("2006-01-02", input.Date)
			if err != nil {
				jsonError(w, "invalid date format, expected ISO8601", http.StatusBadRequest)
				return
			}
		}
	}

	// Get Call communication type
	var callType general.CommunicationType
	if err := h.DB.Where("communication_type = ?", "Call").First(&callType).Error; err != nil {
		jsonError(w, "communication type 'Call' not found", http.StatusInternalServerError)
		return
	}

	locID := int64(loc.IDLocation)
	recall := general.PlaningCommunication{
		PatientID:           idPatient,
		CommunicationTypeID: callType.CommunicationTypeID,
		Reason:              input.Reason,
		Note:                input.Note,
		Date:                recallDt,
		LocationID:          &locID,
	}

	if err := h.DB.Create(&recall).Error; err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, recall.ToMap())
}

// DELETE /api/patient/{patient_id}/recall/{recall_id}
func (h *Handler) DeleteRecall(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idPatient, err := strconv.ParseInt(vars["patient_id"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid patient_id", http.StatusBadRequest)
		return
	}
	recallID, err := strconv.ParseInt(vars["recall_id"], 10, 64)
	if err != nil {
		jsonError(w, "Invalid recall_id", http.StatusBadRequest)
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	_, loc, err := h.getEmployeeAndLocation(username)
	if err != nil || loc == nil {
		jsonError(w, "Employee or location not found", http.StatusNotFound)
		return
	}

	var count int64
	h.DB.Model(&patients.Patient{}).
		Where("id_patient = ? AND location_id = ?", idPatient, loc.IDLocation).Count(&count)
	if count == 0 {
		jsonError(w, "Patient not found", http.StatusNotFound)
		return
	}

	var recall general.PlaningCommunication
	if err := h.DB.Where("id_planing_communication = ? AND patient_id = ? AND location_id = ?",
		recallID, idPatient, loc.IDLocation).First(&recall).Error; err != nil {
		jsonError(w, "Recall not found", http.StatusNotFound)
		return
	}

	if err := h.DB.Delete(&recall).Error; err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]string{"message": "Recall deleted"})
}
