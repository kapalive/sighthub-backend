package request_appointment_service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	apptModel "sighthub-backend/internal/models/appointment"
	empModel "sighthub-backend/internal/models/employees"
	insModel "sighthub-backend/internal/models/insurance"
	locModel "sighthub-backend/internal/models/location"
	mktModel "sighthub-backend/internal/models/marketing"
	patModel "sighthub-backend/internal/models/patients"
	schedModel "sighthub-backend/internal/models/schedule"
	pkgEmail "sighthub-backend/pkg/email"
)

// ── Service ──────────────────────────────────────────────────────────────────

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ── DTOs ─────────────────────────────────────────────────────────────────────

type ShowcaseLocationItem struct {
	IDLocation int    `json:"id_location"`
	FullName   string `json:"full_name"`
	Hash       string `json:"hash"`
}

type DoctorItem struct {
	DoctorID int    `json:"doctor_id"`
	Doctor   string `json:"doctor"`
}

type SlotItem struct {
	Hour         string `json:"hour"`
	Availability bool   `json:"availability"`
}

type DaySlotResult struct {
	Date         string     `json:"date"`
	IsWorkingDay bool       `json:"is_working_day"`
	Hours        []SlotItem `json:"hours"`
}

type CreateRequestAppointmentInput struct {
	Hash                    string  `json:"hash"`
	DoctorID                int64   `json:"doctor_id"`
	PatientID               *int64  `json:"patient_id"`
	FirstName               string  `json:"first_name"`
	LastName                string  `json:"last_name"`
	Dob                     string  `json:"dob"` // YYYY-MM-DD
	Phone                   string  `json:"phone"`
	Email                   string  `json:"email"`
	RequestingDate          string  `json:"requesting_date"` // YYYY-MM-DD
	RequestingTime          string  `json:"requesting_time"` // HH:MM:SS
	ProfessionalServiceTypeID *int64 `json:"professional_service_type_id"`
	InsurancePolicyID       *int64  `json:"insurance_policy_id"`
	InsuranceCompanyID      *int64  `json:"insurance_company_id"`
	GroupNumber             *string `json:"group_number"`
	MemberNumber            *string `json:"member_number"`
	HolderType              *string `json:"holder_type"`
	Note                    *string `json:"note"`
}

type RequestAppointmentResult struct {
	IDRequestAppointment    int64   `json:"id_request_appointment"`
	PatientID               *int64  `json:"patient_id"`
	FirstName               string  `json:"first_name"`
	LastName                string  `json:"last_name"`
	Dob                     *string `json:"dob"`
	Phone                   string  `json:"phone"`
	Email                   *string `json:"email"`
	RequestingDate          string  `json:"requesting_date"`
	RequestingTime          string  `json:"requesting_time"`
	ProfessionalServiceTypeID *int64 `json:"professional_service_type_id"`
	InsurancePolicyID       *int64  `json:"insurance_policy_id"`
	InsuranceCompanyID      *int64  `json:"insurance_company_id"`
	GroupNumber             *string `json:"group_number"`
	MemberNumber            *string `json:"member_number"`
	HolderType              *string `json:"holder_type"`
	Note                    *string `json:"note"`
}

// IntakeFormInput covers both create and update payloads.
type IntakeFormInput struct {
	Hash          string `json:"hash"` // only for create
	AppointmentID int64  `json:"appointment_id"` // only for create

	LastName         *string `json:"lastName"`
	FirstName        *string `json:"firstName"`
	MiddleInitial    *string `json:"middleInitial"`
	DateOfBirth      *string `json:"dateOfBirth"`
	Age              *int16  `json:"age"`
	Gender           *string `json:"gender"`
	CellPhone        *string `json:"cellPhone"`
	HomePhone        *string `json:"homePhone"`
	Address          *string `json:"address"`
	City             *string `json:"city"`
	Zip              *string `json:"zip"`
	ReferredBy       *string `json:"referredBy"`
	EmailAddress     *string `json:"emailAddress"`
	Insurance        *string `json:"insurance"`
	PolicyHolderName *string `json:"policyHolderName"`
	PolicyHolderDOB  *string `json:"policyHolderDOB"`
	LastEyeExam      *string `json:"lastEyeExam"`
	BeenHereBefore   *string `json:"beenHereBefore"`
	WhenBefore       *string `json:"whenBefore"`
	EmergencyContact *string `json:"emergencyContact"`
	WearsGlasses     *string `json:"wearsGlasses"`
	PrimaryCareProvider *string `json:"primaryCareProvider"`
	PrimaryCarePhone    *string `json:"primaryCarePhone"`
	Surgeries        *string `json:"surgeries"`
	DiagnosisType    *string `json:"diagnosisType"`
	DiagnosisDate    *string `json:"diagnosisDate"`
	OtherConditions  *string `json:"otherConditions"`
	SignaturePath    *string `json:"signaturePath"`
	SmokingStatus    *string `json:"smokingStatus"`
	DrinkingStatus   *string `json:"drinkingStatus"`

	ChiefComplaint  map[string]interface{} `json:"chiefComplaint"`
	EyeMedicalHistory map[string]interface{} `json:"eyeMedicalHistory"`
	SignatureData   map[string]interface{} `json:"signatureData"`

	MedicalHistory []map[string]interface{} `json:"medicalHistory"`
	Medications    []string                 `json:"medications"`
	Allergies      []string                 `json:"allergies"`
}

type MedicalHistoryItem struct {
	Name string `json:"name"`
	Yes  bool   `json:"yes"`
	No   bool   `json:"no"`
}

type IntakeFormResult struct {
	LastName         *string `json:"lastName"`
	FirstName        *string `json:"firstName"`
	MiddleInitial    *string `json:"middleInitial"`
	DateOfBirth      string  `json:"dateOfBirth"`
	Age              string  `json:"age"`
	Gender           *string `json:"gender"`
	CellPhone        *string `json:"cellPhone"`
	HomePhone        *string `json:"homePhone"`
	Address          *string `json:"address"`
	City             *string `json:"city"`
	Zip              *string `json:"zip"`
	ReferredBy       *string `json:"referredBy"`
	EmailAddress     *string `json:"emailAddress"`
	Insurance        *string `json:"insurance"`
	PolicyHolderName *string `json:"policyHolderName"`
	PolicyHolderDOB  string  `json:"policyHolderDOB"`
	LastEyeExam      *string `json:"lastEyeExam"`
	BeenHereBefore   *string `json:"beenHereBefore"`
	WhenBefore       *string `json:"whenBefore"`
	EmergencyContact *string `json:"emergencyContact"`
	WearsGlasses     *string `json:"wearsGlasses"`
	PrimaryCareProvider *string `json:"primaryCareProvider"`
	PrimaryCarePhone    *string `json:"primaryCarePhone"`
	Surgeries        *string `json:"surgeries"`
	DiagnosisType    *string `json:"diagnosisType"`
	DiagnosisDate    string  `json:"diagnosisDate"`
	OtherConditions  *string `json:"otherConditions"`
	SmokingStatus    *string `json:"smokingStatus"`
	DrinkingStatus   *string `json:"drinkingStatus"`
	SignaturePath    *string `json:"signaturePath"`

	ReasonForVisit    map[string]interface{} `json:"reasonForVisit"`
	ChiefComplaint    map[string]interface{} `json:"chiefComplaint"`
	EyeMedicalHistory map[string]interface{} `json:"eyeMedicalHistory"`
	SignatureData     map[string]interface{} `json:"signatureData"`

	MedicalHistory []MedicalHistoryItem `json:"medicalHistory"`
	Medications    []string             `json:"medications"`
	Allergies      []string             `json:"allergies"`
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (s *Service) getLocationByHash(hash string) (*locModel.Location, error) {
	if hash == "" {
		return nil, errors.New("hash is required")
	}
	var store locModel.Store
	if err := s.db.Where("hash = ?", hash).First(&store).Error; err != nil {
		return nil, errors.New("location not found")
	}
	var loc locModel.Location
	if err := s.db.Where("store_id = ? AND warehouse_id IS NULL", store.IDStore).First(&loc).Error; err != nil {
		return nil, errors.New("location not found")
	}
	loc.Store = &store
	return &loc, nil
}

func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}

func parseTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("15:04:05", s)
	if err != nil {
		t2, err2 := time.Parse("15:04", s)
		if err2 != nil {
			return nil
		}
		return &t2
	}
	return &t
}

func formatDate(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func timeMinutes(t time.Time) int {
	return t.Hour()*60 + t.Minute()
}

func combineDatetime(d time.Time, t time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.UTC)
}

func boolFromMap(m map[string]interface{}, key string) *bool {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	if b, ok := v.(bool); ok {
		return &b
	}
	return nil
}

func strFromMap(m map[string]interface{}, key string) *string {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	if s, ok := v.(string); ok {
		return &s
	}
	return nil
}

// ── Public methods ────────────────────────────────────────────────────────────

// GetLocations returns all showcase locations with store hash.
func (s *Service) GetLocations() ([]ShowcaseLocationItem, error) {
	var locs []locModel.Location
	if err := s.db.Where("showcase = ?", true).Preload("Store").Find(&locs).Error; err != nil {
		return nil, err
	}
	result := make([]ShowcaseLocationItem, 0, len(locs))
	for _, loc := range locs {
		hash := ""
		if loc.Store != nil {
			hash = loc.Store.Hash
		}
		result = append(result, ShowcaseLocationItem{
			IDLocation: loc.IDLocation,
			FullName:   loc.FullName,
			Hash:       hash,
		})
	}
	return result, nil
}

// GetDoctors returns doctors (job_title_id=1) for the given store hash.
func (s *Service) GetDoctors(hash string) ([]DoctorItem, error) {
	loc, err := s.getLocationByHash(hash)
	if err != nil {
		return nil, err
	}

	var doctors []empModel.Employee
	if err := s.db.Where("job_title_id = 1 AND store_payroll_id = ?", loc.IDLocation).
		Find(&doctors).Error; err != nil {
		return nil, err
	}
	if len(doctors) == 0 {
		return nil, errors.New("no doctors found for this location")
	}

	result := make([]DoctorItem, 0, len(doctors))
	for _, d := range doctors {
		result = append(result, DoctorItem{
			DoctorID: d.IDEmployee,
			Doctor:   fmt.Sprintf("Dr. %s %s", d.FirstName, d.LastName),
		})
	}
	return result, nil
}

// GetDoctorSlotAvailability returns 15-minute availability slots for a doctor over 7 days.
func (s *Service) GetDoctorSlotAvailability(hash string, doctorID int64, startDateStr string) ([]DaySlotResult, error) {
	loc, err := s.getLocationByHash(hash)
	if err != nil {
		return nil, err
	}

	var doctor empModel.Employee
	if err := s.db.Where("id_employee = ?", doctorID).First(&doctor).Error; err != nil {
		return nil, errors.New("doctor not found")
	}

	startDate := time.Now().UTC().Truncate(24 * time.Hour)
	if startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = t
		} else {
			return nil, errors.New("invalid start_date format, use YYYY-MM-DD")
		}
	}
	endDate := startDate.AddDate(0, 0, 6)

	result := make([]DaySlotResult, 0, 7)
	current := startDate
	for !current.After(endDate) {
		dayName := strings.ToLower(current.Weekday().String())

		var schedEntry schedModel.Schedule
		hasSchedule := s.db.Where("employee_id = ? AND day_of_week = ?", doctorID, dayName).
			First(&schedEntry).Error == nil

		var calEntry schedModel.Calendar
		hasCalendar := s.db.Where("employee_id = ? AND date = ?", doctorID, current.Format("2006-01-02")).
			First(&calEntry).Error == nil

		var startTime, endTime *time.Time
		var lunchStart, lunchEnd *time.Time

		if hasCalendar && !calEntry.IsWorkingDay {
			result = append(result, DaySlotResult{
				Date:         current.Format("2006-01-02"),
				IsWorkingDay: false,
				Hours:        []SlotItem{},
			})
			current = current.AddDate(0, 0, 1)
			continue
		}

		if hasSchedule {
			st := schedEntry.StartTime
			et := schedEntry.EndTime
			startTime = &st
			endTime = &et
			lunchStart = schedEntry.LunchStart
			lunchEnd = schedEntry.LunchEnd
		}
		if hasCalendar {
			if calEntry.TimeStart != nil {
				startTime = calEntry.TimeStart
			}
			if calEntry.TimeEnd != nil {
				endTime = calEntry.TimeEnd
			}
		}

		if startTime == nil || endTime == nil {
			result = append(result, DaySlotResult{
				Date:         current.Format("2006-01-02"),
				IsWorkingDay: false,
				Hours:        []SlotItem{},
			})
			current = current.AddDate(0, 0, 1)
			continue
		}

		// Load appointments for this doctor+location+date via schedule join
		var appts []apptModel.Appointment
		s.db.Joins("JOIN schedule ON schedule.id_schedule = appointment.schedule_id").
			Where("schedule.employee_id = ? AND appointment.location_id = ? AND appointment.appointment_date = ?",
				doctorID, loc.IDLocation, current.Format("2006-01-02")).
			Find(&appts)

		busySlots := map[string]bool{}
		for _, a := range appts {
			st := combineDatetime(current, a.StartTime)
			et := combineDatetime(current, a.EndTime)
			for st.Before(et) {
				busySlots[st.Format("15:04")] = true
				st = st.Add(15 * time.Minute)
			}
		}

		hours := []SlotItem{}
		cur := combineDatetime(current, *startTime)
		end := combineDatetime(current, *endTime)
		var lunchStartDt, lunchEndDt time.Time
		hasLunch := lunchStart != nil && lunchEnd != nil
		if hasLunch {
			lunchStartDt = combineDatetime(current, *lunchStart)
			lunchEndDt = combineDatetime(current, *lunchEnd)
		}

		for cur.Before(end) {
			slot := cur.Format("15:04")
			available := !busySlots[slot]
			if hasLunch && !cur.Before(lunchStartDt) && cur.Before(lunchEndDt) {
				available = false
			}
			hours = append(hours, SlotItem{Hour: slot, Availability: available})
			cur = cur.Add(15 * time.Minute)
		}

		result = append(result, DaySlotResult{
			Date:         current.Format("2006-01-02"),
			IsWorkingDay: true,
			Hours:        hours,
		})
		current = current.AddDate(0, 0, 1)
	}
	return result, nil
}

// CreateRequestAppointment creates a request appointment for a patient.
func (s *Service) CreateRequestAppointment(input CreateRequestAppointmentInput) (*RequestAppointmentResult, error) {
	loc, err := s.getLocationByHash(input.Hash)
	if err != nil {
		return nil, err
	}

	var setting locModel.LocationAppointmentSettings
	if err := s.db.Where("location_id = ?", loc.IDLocation).First(&setting).Error; err != nil {
		return nil, errors.New("request appointment disabled for this location")
	}
	if !setting.RequestAppointmentEnabled {
		return nil, errors.New("request appointment disabled for this location")
	}

	if input.DoctorID == 0 {
		return nil, errors.New("doctor_id is required")
	}
	var doctor empModel.Employee
	if err := s.db.Where("id_employee = ?", input.DoctorID).First(&doctor).Error; err != nil {
		return nil, errors.New("doctor not found for this location")
	}
	if doctor.StorePayrollID == nil || *doctor.StorePayrollID != int64(loc.IDLocation) {
		return nil, errors.New("doctor not found for this location")
	}

	if input.RequestingDate == "" || input.RequestingTime == "" {
		return nil, errors.New("requesting_date and requesting_time are required")
	}
	reqDate := parseDate(input.RequestingDate)
	if reqDate == nil {
		return nil, errors.New("invalid requesting_date format")
	}
	reqTime := parseTime(input.RequestingTime)
	if reqTime == nil {
		return nil, errors.New("invalid requesting_time format")
	}

	// Patient info
	firstName, lastName, phone, email := input.FirstName, input.LastName, input.Phone, input.Email
	var dob *time.Time
	if input.Dob != "" {
		dob = parseDate(input.Dob)
	}
	var patientID *int64

	if input.PatientID != nil {
		var pat patModel.Patient
		if err := s.db.Where("id_patient = ?", *input.PatientID).First(&pat).Error; err != nil {
			return nil, errors.New("patient not found")
		}
		firstName = pat.FirstName
		lastName = pat.LastName
		dob = pat.DOB
		if pat.Phone != nil {
			phone = *pat.Phone
		}
		if pat.Email != nil {
			email = *pat.Email
		}
		patientID = input.PatientID
	} else {
		if firstName == "" || lastName == "" || phone == "" {
			return nil, errors.New("first_name, last_name, and phone are required")
		}
	}

	// Insurance info
	var insCompanyID *int64
	var groupNumber, memberNumber, holderType *string
	var insPolicyID *int64

	if input.InsurancePolicyID != nil {
		var policy insModel.InsurancePolicy
		if err := s.db.Where("id_insurance_policy = ?", *input.InsurancePolicyID).First(&policy).Error; err != nil {
			return nil, errors.New("insurance policy not found")
		}
		cid := int64(policy.InsuranceCompanyID)
		insCompanyID = &cid
		groupNumber = policy.GroupNumber
		insPolicyID = input.InsurancePolicyID

		if patientID != nil {
			var holder patModel.InsuranceHolderPatients
			if err := s.db.Where("patient_id = ? AND insurance_policy_id = ?", *patientID, *input.InsurancePolicyID).
				First(&holder).Error; err == nil {
				holderType = &holder.HolderType
				memberNumber = holder.MemberNumber
			}
		}
	} else {
		insCompanyID = input.InsuranceCompanyID
		groupNumber = input.GroupNumber
		memberNumber = input.MemberNumber
		holderType = input.HolderType
		insPolicyID = nil
	}

	ra := apptModel.RequestAppointment{
		DoctorID:                  input.DoctorID,
		PatientID:                 patientID,
		FirstName:                 firstName,
		LastName:                  lastName,
		Dob:                       dob,
		Phone:                     phone,
		RequestingDate:            *reqDate,
		RequestingTime:            *reqTime,
		ProfessionalServiceTypeID: input.ProfessionalServiceTypeID,
		InsurancePolicyID:         insPolicyID,
		InsuranceCompanyID:        insCompanyID,
		GroupNumber:               groupNumber,
		MemberNumber:              memberNumber,
		HolderType:                holderType,
		Note:                      input.Note,
		Processed:                 boolPtr(false),
	}
	if email != "" {
		ra.Email = &email
	}
	if err := s.db.Create(&ra).Error; err != nil {
		return nil, err
	}

	// Send confirmation email (best-effort)
	if email != "" {
		locID := int64(loc.IDLocation)
		template := pkgEmail.GetTemplateForCategory(s.db, "appointment")
		orgName := ""
		if loc.Store != nil {
			if loc.Store.BusinessName != nil {
				orgName = *loc.Store.BusinessName
			} else if loc.Store.FullName != nil {
				orgName = *loc.Store.FullName
			}
		}
		_ = pkgEmail.SendViaDB(s.db, email, "Your Appointment Request", template, map[string]interface{}{
			"patient_name":     strings.TrimSpace(firstName + " " + lastName),
			"appointment_date": input.RequestingDate,
			"appointment_time": input.RequestingTime,
			"provider_name":    strings.TrimSpace(doctor.FirstName + " " + doctor.LastName),
			"location":         loc.FullName,
			"organization_name": orgName,
		}, &locID)
	}

	// Build result
	var dobStr *string
	if ra.Dob != nil && !ra.Dob.IsZero() {
		s := ra.Dob.Format("2006-01-02")
		dobStr = &s
	}
	reqDateStr := ra.RequestingDate.Format("2006-01-02")
	reqTimeStr := ra.RequestingTime.Format("15:04:05")

	return &RequestAppointmentResult{
		IDRequestAppointment:    ra.IDRequestAppointment,
		PatientID:               ra.PatientID,
		FirstName:               ra.FirstName,
		LastName:                ra.LastName,
		Dob:                     dobStr,
		Phone:                   ra.Phone,
		Email:                   ra.Email,
		RequestingDate:          reqDateStr,
		RequestingTime:          reqTimeStr,
		ProfessionalServiceTypeID: ra.ProfessionalServiceTypeID,
		InsurancePolicyID:       ra.InsurancePolicyID,
		InsuranceCompanyID:      ra.InsuranceCompanyID,
		GroupNumber:             ra.GroupNumber,
		MemberNumber:            ra.MemberNumber,
		HolderType:              ra.HolderType,
		Note:                    ra.Note,
	}, nil
}

// CreateIntakeForm creates an intake form for a given appointment.
func (s *Service) CreateIntakeForm(input IntakeFormInput) (int64, error) {
	loc, err := s.getLocationByHash(input.Hash)
	if err != nil {
		return 0, err
	}

	var setting locModel.LocationAppointmentSettings
	if err := s.db.Where("location_id = ?", loc.IDLocation).First(&setting).Error; err != nil || !setting.IntakeFormEnabled {
		return 0, errors.New("intake form disabled for this location")
	}

	var appt apptModel.Appointment
	if err := s.db.Where("id_appointment = ?", input.AppointmentID).First(&appt).Error; err != nil {
		return 0, errors.New("appointment not found")
	}

	form := mktModel.IntakeFormData{
		AppointmentID:    input.AppointmentID,
		LastName:         input.LastName,
		FirstName:        input.FirstName,
		MiddleInitial:    input.MiddleInitial,
		DateOfBirth:      parseDate(strVal(input.DateOfBirth)),
		Age:              input.Age,
		Gender:           input.Gender,
		CellPhone:        input.CellPhone,
		HomePhone:        input.HomePhone,
		Address:          input.Address,
		City:             input.City,
		Zip:              input.Zip,
		ReferredBy:       input.ReferredBy,
		EmailAddress:     input.EmailAddress,
		Insurance:        input.Insurance,
		PolicyHolderName: input.PolicyHolderName,
		PolicyHolderDob:  parseDate(strVal(input.PolicyHolderDOB)),
		LastEyeExam:      input.LastEyeExam,
		BeenHereBefore:   input.BeenHereBefore,
		WhenBefore:       input.WhenBefore,
		EmergencyContact: input.EmergencyContact,
		WearsGlasses:     input.WearsGlasses,
		PrimaryCareProvider: input.PrimaryCareProvider,
		PrimaryCarePhone:    input.PrimaryCarePhone,
		Surgeries:        input.Surgeries,
		DiagnosisType:    input.DiagnosisType,
		DiagnosisDate:    parseDate(strVal(input.DiagnosisDate)),
		OtherConditions:  input.OtherConditions,
		SignaturePath:    input.SignaturePath,
		SmokingStatus:    input.SmokingStatus,
		DrinkingStatus:   input.DrinkingStatus,
	}
	applyComplaint(&form, input.ChiefComplaint)
	applyEyeHistory(&form, input.EyeMedicalHistory)
	if input.SignatureData != nil {
		form.Signature = strFromMap(input.SignatureData, "signature")
		form.SignatureDate = parseDate(strVal(strFromMap(input.SignatureData, "signatureDate")))
		form.PatientOldRx = strFromMap(input.SignatureData, "patientOldRx")
	}

	var formID int64
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&form).Error; err != nil {
			return err
		}
		formID = form.IDIntakeFormData

		for _, item := range input.MedicalHistory {
			name, ok := item["name"].(string)
			if !ok || name == "" {
				continue
			}
			val, _ := item["yes"].(bool)
			if err := tx.Create(&mktModel.IntakeFormMedicalHistory{
				RequestID: formID,
				Name:      name,
				Value:     val,
			}).Error; err != nil {
				return err
			}
		}
		for _, med := range input.Medications {
			if err := tx.Create(&mktModel.IntakeFormMedications{
				RequestID: formID,
				Name:      med,
			}).Error; err != nil {
				return err
			}
		}
		for _, allergy := range input.Allergies {
			if err := tx.Create(&mktModel.IntakeFormAllergies{
				RequestID: formID,
				Name:      allergy,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return formID, nil
}

// GetIntakeForm returns intake form data including related rows.
func (s *Service) GetIntakeForm(id int64) (*IntakeFormResult, error) {
	var form mktModel.IntakeFormData
	err := s.db.Preload("MedicalHistory").Preload("Medications").Preload("Allergies").
		Where("id_intake_form_data = ?", id).First(&form).Error
	if err != nil {
		return nil, errors.New("intake form not found")
	}

	medHistory := make([]MedicalHistoryItem, 0, len(form.MedicalHistory))
	for _, h := range form.MedicalHistory {
		medHistory = append(medHistory, MedicalHistoryItem{Name: h.Name, Yes: h.Value, No: !h.Value})
	}
	meds := make([]string, 0, len(form.Medications))
	for _, m := range form.Medications {
		meds = append(meds, m.Name)
	}
	allergies := make([]string, 0, len(form.Allergies))
	for _, a := range form.Allergies {
		allergies = append(allergies, a.Name)
	}

	ageStr := ""
	if form.Age != nil {
		ageStr = fmt.Sprintf("%d", *form.Age)
	}

	return &IntakeFormResult{
		LastName:         form.LastName,
		FirstName:        form.FirstName,
		MiddleInitial:    form.MiddleInitial,
		DateOfBirth:      formatDate(form.DateOfBirth),
		Age:              ageStr,
		Gender:           form.Gender,
		CellPhone:        form.CellPhone,
		HomePhone:        form.HomePhone,
		Address:          form.Address,
		City:             form.City,
		Zip:              form.Zip,
		ReferredBy:       form.ReferredBy,
		EmailAddress:     form.EmailAddress,
		Insurance:        form.Insurance,
		PolicyHolderName: form.PolicyHolderName,
		PolicyHolderDOB:  formatDate(form.PolicyHolderDob),
		LastEyeExam:      form.LastEyeExam,
		BeenHereBefore:   form.BeenHereBefore,
		WhenBefore:       form.WhenBefore,
		EmergencyContact: form.EmergencyContact,
		WearsGlasses:     form.WearsGlasses,
		PrimaryCareProvider: form.PrimaryCareProvider,
		PrimaryCarePhone:    form.PrimaryCarePhone,
		Surgeries:        form.Surgeries,
		DiagnosisType:    form.DiagnosisType,
		DiagnosisDate:    formatDate(form.DiagnosisDate),
		OtherConditions:  form.OtherConditions,
		SmokingStatus:    form.SmokingStatus,
		DrinkingStatus:   form.DrinkingStatus,
		SignaturePath:    form.SignaturePath,
		ReasonForVisit: map[string]interface{}{
			"routineExam": boolVal(form.BlurryVision),
			"medicalExam": boolVal(form.DoubleVision),
			"clExam":      boolVal(form.Itching),
		},
		ChiefComplaint: map[string]interface{}{
			"blurryVision":         form.BlurryVision,
			"blurryVisionDistance": form.BlurryVisionDistance,
			"blurryVisionNear":     form.BlurryVisionNear,
			"itching":              form.Itching,
			"doubleVision":         form.DoubleVision,
			"burning":              form.Burning,
			"eyeInjury":            form.EyeInjury,
			"eyeInfection":         form.EyeInfection,
			"tearing":              form.Tearing,
			"floatersSpots":        form.FloatersSpots,
			"flashesOfLight":       form.FlashesOfLight,
			"pain":                 form.Pain,
			"lightSensitivity":     form.LightSensitivity,
		},
		EyeMedicalHistory: map[string]interface{}{
			"cataracts":         form.Cataracts,
			"glaucoma":          form.Glaucoma,
			"retinalProblems":   form.RetinalProblems,
			"cornealProblems":   form.CornealProblems,
			"cataractSurgery":   form.CataractSurgery,
			"cataractSurgeryEye": form.CataractSurgeryEye,
			"dateAndDoctorName": form.DateAndDoctorName,
		},
		SignatureData: map[string]interface{}{
			"signature":     form.Signature,
			"signatureDate": formatDate(form.SignatureDate),
			"patientOldRx":  form.PatientOldRx,
		},
		MedicalHistory: medHistory,
		Medications:    meds,
		Allergies:      allergies,
	}, nil
}

// UpdateIntakeForm replaces an intake form's data including related rows.
func (s *Service) UpdateIntakeForm(id int64, input IntakeFormInput) error {
	var form mktModel.IntakeFormData
	if err := s.db.Where("id_intake_form_data = ?", id).First(&form).Error; err != nil {
		return errors.New("intake form not found")
	}

	form.LastName = input.LastName
	form.FirstName = input.FirstName
	form.MiddleInitial = input.MiddleInitial
	form.DateOfBirth = parseDate(strVal(input.DateOfBirth))
	form.Age = input.Age
	form.Gender = input.Gender
	form.CellPhone = input.CellPhone
	form.HomePhone = input.HomePhone
	form.Address = input.Address
	form.City = input.City
	form.Zip = input.Zip
	form.ReferredBy = input.ReferredBy
	form.EmailAddress = input.EmailAddress
	form.Insurance = input.Insurance
	form.PolicyHolderName = input.PolicyHolderName
	form.PolicyHolderDob = parseDate(strVal(input.PolicyHolderDOB))
	form.LastEyeExam = input.LastEyeExam
	form.BeenHereBefore = input.BeenHereBefore
	form.WhenBefore = input.WhenBefore
	form.EmergencyContact = input.EmergencyContact
	form.WearsGlasses = input.WearsGlasses
	form.PrimaryCareProvider = input.PrimaryCareProvider
	form.PrimaryCarePhone = input.PrimaryCarePhone
	form.Surgeries = input.Surgeries
	form.DiagnosisType = input.DiagnosisType
	form.DiagnosisDate = parseDate(strVal(input.DiagnosisDate))
	form.OtherConditions = input.OtherConditions
	form.SignaturePath = input.SignaturePath
	form.SmokingStatus = input.SmokingStatus
	form.DrinkingStatus = input.DrinkingStatus
	applyComplaint(&form, input.ChiefComplaint)
	applyEyeHistory(&form, input.EyeMedicalHistory)
	if input.SignatureData != nil {
		form.Signature = strFromMap(input.SignatureData, "signature")
		form.SignatureDate = parseDate(strVal(strFromMap(input.SignatureData, "signatureDate")))
		form.PatientOldRx = strFromMap(input.SignatureData, "patientOldRx")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&form).Error; err != nil {
			return err
		}
		tx.Where("request_id = ?", form.IDIntakeFormData).Delete(&mktModel.IntakeFormMedicalHistory{})
		tx.Where("request_id = ?", form.IDIntakeFormData).Delete(&mktModel.IntakeFormMedications{})
		tx.Where("request_id = ?", form.IDIntakeFormData).Delete(&mktModel.IntakeFormAllergies{})

		for _, item := range input.MedicalHistory {
			name, ok := item["name"].(string)
			if !ok || name == "" {
				continue
			}
			val, _ := item["yes"].(bool)
			if err := tx.Create(&mktModel.IntakeFormMedicalHistory{
				RequestID: form.IDIntakeFormData,
				Name:      name,
				Value:     val,
			}).Error; err != nil {
				return err
			}
		}
		for _, med := range input.Medications {
			if err := tx.Create(&mktModel.IntakeFormMedications{
				RequestID: form.IDIntakeFormData,
				Name:      med,
			}).Error; err != nil {
				return err
			}
		}
		for _, allergy := range input.Allergies {
			if err := tx.Create(&mktModel.IntakeFormAllergies{
				RequestID: form.IDIntakeFormData,
				Name:      allergy,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// CheckAppointment verifies that an appointment is upcoming (status 3 or 5, future datetime).
func (s *Service) CheckAppointment(appointmentID int64) error {
	var appt apptModel.Appointment
	if err := s.db.Where("id_appointment = ?", appointmentID).First(&appt).Error; err != nil {
		return errors.New("appointment not found")
	}
	if appt.StatusAppointmentID != 3 && appt.StatusAppointmentID != 5 {
		return errors.New("appointment has invalid status")
	}
	apptDatetime := combineDatetime(appt.AppointmentDate, appt.EndTime)
	if apptDatetime.Before(time.Now()) {
		return errors.New("appointment time has already passed")
	}
	return nil
}

// ── Private helpers ───────────────────────────────────────────────────────────

func applyComplaint(form *mktModel.IntakeFormData, m map[string]interface{}) {
	form.BlurryVision = boolFromMap(m, "blurryVision")
	form.BlurryVisionDistance = strFromMap(m, "blurryVisionDistance")
	form.BlurryVisionNear = strFromMap(m, "blurryVisionNear")
	form.Itching = boolFromMap(m, "itching")
	form.DoubleVision = boolFromMap(m, "doubleVision")
	form.Burning = boolFromMap(m, "burning")
	form.EyeInjury = boolFromMap(m, "eyeInjury")
	form.EyeInfection = boolFromMap(m, "eyeInfection")
	form.Tearing = boolFromMap(m, "tearing")
	form.FloatersSpots = boolFromMap(m, "floatersSpots")
	form.FlashesOfLight = boolFromMap(m, "flashesOfLight")
	form.Pain = boolFromMap(m, "pain")
	form.LightSensitivity = boolFromMap(m, "lightSensitivity")
}

func applyEyeHistory(form *mktModel.IntakeFormData, m map[string]interface{}) {
	form.Cataracts = boolFromMap(m, "cataracts")
	form.Glaucoma = boolFromMap(m, "glaucoma")
	form.RetinalProblems = boolFromMap(m, "retinalProblems")
	form.CornealProblems = boolFromMap(m, "cornealProblems")
	form.CataractSurgery = boolFromMap(m, "cataractSurgery")
	form.CataractSurgeryEye = strFromMap(m, "cataractSurgeryEye")
	form.DateAndDoctorName = strFromMap(m, "dateAndDoctorName")
}

func boolVal(b *bool) bool {
	return b != nil && *b
}

func boolPtr(b bool) *bool {
	return &b
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
