package communication

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/general"
	locModel "sighthub-backend/internal/models/location"
	"sighthub-backend/internal/models/patients"
	pkgAuth "sighthub-backend/pkg/auth"
	"sighthub-backend/pkg/communication"
	pkgEmail "sighthub-backend/pkg/email"
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

func parsePatientID(r *http.Request) (int64, error) {
	v := mux.Vars(r)["patient_id"]
	return strconv.ParseInt(v, 10, 64)
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

func (h *Handler) getCommunicationTypeID(typeName string) (int, error) {
	var ct general.CommunicationType
	if err := h.DB.Where("communication_type = ?", typeName).First(&ct).Error; err != nil {
		return 0, err
	}
	return ct.CommunicationTypeID, nil
}

func (h *Handler) employeeName(employeeID int64) string {
	var emp empModel.Employee
	if err := h.DB.First(&emp, employeeID).Error; err != nil {
		return "Unknown Employee"
	}
	return fmt.Sprintf("%s %s", emp.LastName, emp.FirstName)
}

func (h *Handler) locationName(locationID int) string {
	var loc locModel.Location
	if err := h.DB.First(&loc, locationID).Error; err != nil {
		return "Unknown location"
	}
	if loc.ShortName != nil {
		return *loc.ShortName
	}
	return loc.FullName
}

// GET /api/patient/{patient_id}/sms-history
func (h *Handler) GetSMSHistory(w http.ResponseWriter, r *http.Request) {
	idPatient, err := parsePatientID(r)
	if err != nil {
		jsonError(w, "Invalid patient_id", http.StatusBadRequest)
		return
	}

	// Check patient exists
	var count int64
	h.DB.Model(&patients.Patient{}).Where("id_patient = ?", idPatient).Count(&count)
	if count == 0 {
		jsonError(w, "Patient not found", http.StatusNotFound)
		return
	}

	smsTypeID, err := h.getCommunicationTypeID("SMS")
	if err != nil {
		jsonOK(w, []interface{}{})
		return
	}

	var logs []patients.PatientCommunicationHistory
	h.DB.Where("patient_id = ? AND communication_type_id = ?", idPatient, smsTypeID).
		Order("communication_datetime DESC").Find(&logs)

	result := make([]map[string]interface{}, 0, len(logs))
	for _, sms := range logs {
		result = append(result, map[string]interface{}{
			"date":     sms.CommunicationDatetime.Format("01/02/2006"),
			"content":  sms.Content,
			"location": h.locationName(sms.LocationID),
			"by":       h.employeeName(sms.EmployeeID),
		})
	}

	jsonOK(w, result)
}

// GET /api/patient/{patient_id}/email-history
func (h *Handler) GetEmailHistory(w http.ResponseWriter, r *http.Request) {
	idPatient, err := parsePatientID(r)
	if err != nil {
		jsonError(w, "Invalid patient_id", http.StatusBadRequest)
		return
	}

	var count int64
	h.DB.Model(&patients.Patient{}).Where("id_patient = ?", idPatient).Count(&count)
	if count == 0 {
		jsonError(w, "Patient not found", http.StatusNotFound)
		return
	}

	emailTypeID, err := h.getCommunicationTypeID("Email")
	if err != nil {
		jsonOK(w, []interface{}{})
		return
	}

	var logs []patients.PatientCommunicationHistory
	h.DB.Where("patient_id = ? AND communication_type_id = ?", idPatient, emailTypeID).
		Order("communication_datetime DESC").Find(&logs)

	result := make([]map[string]interface{}, 0, len(logs))
	for _, email := range logs {
		result = append(result, map[string]interface{}{
			"date":     email.CommunicationDatetime.Format("01/02/2006"),
			"subject":  email.Content,
			"location": h.locationName(email.LocationID),
			"by":       h.employeeName(email.EmployeeID),
		})
	}

	jsonOK(w, result)
}

// GET /api/patient/{patient_id}/call-history
func (h *Handler) GetCallHistory(w http.ResponseWriter, r *http.Request) {
	idPatient, err := parsePatientID(r)
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

	// Call history is location-scoped: check patient belongs to this location
	var patient patients.Patient
	if err := h.DB.Where("id_patient = ? AND location_id = ?", idPatient, loc.IDLocation).
		First(&patient).Error; err != nil {
		jsonOK(w, []interface{}{})
		return
	}

	callTypeID, err := h.getCommunicationTypeID("Call")
	if err != nil {
		jsonOK(w, []interface{}{})
		return
	}

	var logs []patients.PatientCommunicationHistory
	h.DB.Where("patient_id = ? AND communication_type_id = ? AND location_id = ?",
		idPatient, callTypeID, loc.IDLocation).
		Order("communication_datetime DESC").Find(&logs)

	locName := loc.FullName
	if loc.ShortName != nil {
		locName = *loc.ShortName
	}
	patientName := fmt.Sprintf("%s %s", patient.FirstName, patient.LastName)

	result := make([]map[string]interface{}, 0, len(logs))
	for _, comm := range logs {
		result = append(result, map[string]interface{}{
			"date_time":     comm.CommunicationDatetime.Format("2006-01-02T15:04:05"),
			"call":          comm.Content,
			"patient_name":  patientName,
			"location":      locName,
			"employee_name": h.employeeName(comm.EmployeeID),
		})
	}

	jsonOK(w, result)
}

// POST /api/patient/{patient_id}/send-sms
func (h *Handler) SendSMS(w http.ResponseWriter, r *http.Request) {
	idPatient, err := parsePatientID(r)
	if err != nil {
		jsonError(w, "Invalid patient_id", http.StatusBadRequest)
		return
	}

	var patient patients.Patient
	if err := h.DB.First(&patient, idPatient).Error; err != nil {
		jsonError(w, "Patient not found", http.StatusNotFound)
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	emp, loc, err := h.getEmployeeAndLocation(username)
	if err != nil || emp == nil || loc == nil {
		jsonError(w, "Employee or location not found", http.StatusNotFound)
		return
	}

	var input struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if input.Content == "" {
		jsonError(w, "SMS content is required", http.StatusBadRequest)
		return
	}

	smsTypeID, err := h.getCommunicationTypeID("SMS")
	if err != nil {
		jsonError(w, "SMS communication type not found", http.StatusNotFound)
		return
	}

	if patient.Phone == nil || *patient.Phone == "" {
		jsonError(w, "Patient has no phone number", http.StatusBadRequest)
		return
	}
	smsResult := communication.SendSMS(*patient.Phone, input.Content)
	if smsResult.Status != "accepted" {
		errMsg := smsResult.Error
		if errMsg == "" {
			errMsg = "Unknown SMS error"
		}
		jsonError(w, errMsg, http.StatusInternalServerError)
		return
	}

	entry := patients.PatientCommunicationHistory{
		PatientID:             idPatient,
		CommunicationTypeID:   smsTypeID,
		Content:               input.Content,
		Description:           strPtr("SMS notification sent"),
		CommunicationDatetime: time.Now(),
		EmployeeID:            int64(emp.IDEmployee),
		LocationID:            loc.IDLocation,
	}
	if err := h.DB.Create(&entry).Error; err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, map[string]interface{}{
		"message":         "SMS sent successfully",
		"twilio_response": smsResult,
	})
}

// POST /api/patient/{patient_id}/log-call
func (h *Handler) LogCall(w http.ResponseWriter, r *http.Request) {
	idPatient, err := parsePatientID(r)
	if err != nil {
		jsonError(w, "Invalid patient_id", http.StatusBadRequest)
		return
	}

	var count int64
	h.DB.Model(&patients.Patient{}).Where("id_patient = ?", idPatient).Count(&count)
	if count == 0 {
		jsonError(w, "Patient not found", http.StatusNotFound)
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	emp, loc, err := h.getEmployeeAndLocation(username)
	if err != nil || emp == nil || loc == nil {
		jsonError(w, "Employee or location not found", http.StatusNotFound)
		return
	}

	var input struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if input.Content == "" {
		jsonError(w, "Call note content is required", http.StatusBadRequest)
		return
	}

	callTypeID, err := h.getCommunicationTypeID("Call")
	if err != nil {
		jsonError(w, "Call communication type not found", http.StatusNotFound)
		return
	}

	entry := patients.PatientCommunicationHistory{
		PatientID:             idPatient,
		CommunicationTypeID:   callTypeID,
		Content:               input.Content,
		Description:           strPtr("Phone call logged"),
		CommunicationDatetime: time.Now(),
		EmployeeID:            int64(emp.IDEmployee),
		LocationID:            loc.IDLocation,
	}
	if err := h.DB.Create(&entry).Error; err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, map[string]string{"message": "Call logged successfully"})
}

// POST /api/patient/{patient_id}/send-email
func (h *Handler) SendEmail(w http.ResponseWriter, r *http.Request) {
	idPatient, err := parsePatientID(r)
	if err != nil {
		jsonError(w, "Invalid patient_id", http.StatusBadRequest)
		return
	}

	var patient patients.Patient
	if err := h.DB.First(&patient, idPatient).Error; err != nil {
		jsonError(w, "Patient not found", http.StatusNotFound)
		return
	}

	username := pkgAuth.UsernameFromContext(r.Context())
	emp, loc, err := h.getEmployeeAndLocation(username)
	if err != nil || emp == nil || loc == nil {
		jsonError(w, "Employee or location not found", http.StatusNotFound)
		return
	}

	var input struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if input.Subject == "" || input.Body == "" {
		jsonError(w, "Email subject and body are required", http.StatusBadRequest)
		return
	}

	emailTypeID, err := h.getCommunicationTypeID("Email")
	if err != nil {
		jsonError(w, "Email communication type not found", http.StatusNotFound)
		return
	}

	entry := patients.PatientCommunicationHistory{
		PatientID:             idPatient,
		CommunicationTypeID:   emailTypeID,
		Content:               input.Subject,
		Description:           &input.Body,
		CommunicationDatetime: time.Now(),
		EmployeeID:            int64(emp.IDEmployee),
		LocationID:            loc.IDLocation,
	}
	if err := h.DB.Create(&entry).Error; err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if patient.Email != nil && *patient.Email != "" {
		orgName := ""
		if loc.StoreID != 0 {
			var store locModel.Store
			if err := h.DB.First(&store, loc.StoreID).Error; err == nil {
				if store.BusinessName != nil {
					orgName = *store.BusinessName
				} else if store.FullName != nil {
					orgName = *store.FullName
				}
			}
		}
		patientName := strings.TrimSpace(fmt.Sprintf("%s %s", patient.FirstName, patient.LastName))
		locID := int64(loc.IDLocation)
		_ = pkgEmail.SendViaDB(
			h.DB,
			*patient.Email,
			input.Subject,
			"default.html",
			map[string]interface{}{
				"patient_name":      patientName,
				"subject":           input.Subject,
				"body":              input.Body,
				"organization_name": orgName,
			},
			&locID,
		)
	}

	w.WriteHeader(http.StatusCreated)
	jsonOK(w, map[string]string{"message": "Email logged successfully"})
}

func strPtr(s string) *string { return &s }
