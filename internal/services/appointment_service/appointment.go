package appointment_service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	apptModel "sighthub-backend/internal/models/appointment"
	empModel "sighthub-backend/internal/models/employees"
	generalModel "sighthub-backend/internal/models/general"
	insuranceModel "sighthub-backend/internal/models/insurance"
	locModel "sighthub-backend/internal/models/location"
	mktModel "sighthub-backend/internal/models/marketing"
	patModel "sighthub-backend/internal/models/patients"
	schedModel "sighthub-backend/internal/models/schedule"
	svcModel "sighthub-backend/internal/models/service"
	"sighthub-backend/pkg/activitylog"
	"sighthub-backend/pkg/communication"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type LocationItem struct {
	LocationID int    `json:"location_id"`
	FullName   string `json:"full_name"`
	ShortName  string `json:"short_name"`
}

type SetLocationResult struct {
	LocationID int    `json:"location_id"`
	Location   string `json:"location"`
}

type DoctorItem struct {
	DoctorID int    `json:"doctor_id"`
	Doctor   string `json:"doctor"`
}

type WorkHoursDay struct {
	Date         string  `json:"date"`
	IsWorkingDay bool    `json:"is_working_day"`
	Message      *string `json:"message,omitempty"`
	TimeStart    *string `json:"time_start"`
	TimeEnd      *string `json:"time_end"`
	LunchStart   *string `json:"lunch_start,omitempty"`
	LunchEnd     *string `json:"lunch_end,omitempty"`
	ApptDuration *int    `json:"appointment_duration,omitempty"`
}

type AppointmentItem struct {
	AppointmentID int64   `json:"appointment_id"`
	Title         string  `json:"title"`
	Start         string  `json:"start"`
	End           string  `json:"end"`
	DoctorID      int64   `json:"doctor_id"`
	Doctor        string  `json:"doctor"`
	Notes         *string `json:"notes"`
	PatientID     int64   `json:"id"`
	InsuranceID   *int64  `json:"insurance_id"`
	StatusID      int     `json:"status_id"`
	Status        string  `json:"status"`
}

type DaySchedule struct {
	Date              string            `json:"date"`
	IsWorkingDay      bool              `json:"is_working_day"`
	LocationWorkHours WorkHoursDay      `json:"location_work_hours"`
	Appointments      []AppointmentItem `json:"appointments"`
}

type RequestAppointmentItem struct {
	IDRequestAppointment         int64   `json:"id_request_appointment"`
	DoctorID                     int64   `json:"doctor_id"`
	PatientID                    *int64  `json:"patient_id"`
	PatientName                  *string `json:"patient_name"`
	FirstName                    string  `json:"first_name"`
	LastName                     string  `json:"last_name"`
	DOB                          *string `json:"dob"`
	Accept                       *bool   `json:"accept"`
	Phone                        string  `json:"phone"`
	Email                        *string `json:"email"`
	RequestingDate               string  `json:"requesting_date"`
	RequestingTime               string  `json:"requesting_time"`
	ProfessionalServiceTypeID    *int64  `json:"professional_service_type_id"`
	ProfessionalServiceTypeTitle *string `json:"professional_service_type_title"`
	InsuranceCompanyID           *int64  `json:"insurance_company_id"`
	InsuranceCompanyName         *string `json:"insurance_company_name"`
	GroupNumber                  *string `json:"group_number"`
	MemberNumber                 *string `json:"member_number"`
	HolderType                   *string `json:"holder_type"`
	Note                         *string `json:"note"`
}

type CancelRequestResult struct {
	Message              string `json:"message"`
	IDRequestAppointment int64  `json:"id_request_appointment"`
	Accept               *bool  `json:"accept"`
	Processed            bool   `json:"processed"`
}

type CreateAppointmentInput struct {
	NewPatient           bool    `json:"new_patient"`
	IDRequestAppointment *int64  `json:"id_request_appointment"`
	PatientID            *int64  `json:"patient_id"`
	LocationID           *int    `json:"location_id"`
	DoctorID             *int64  `json:"doctor_id"`
	AppointmentDate      *string `json:"appointment_date"`
	StartTime            *string `json:"start_time"`
	EndTime              *string `json:"end_time"`
	InsuranceID          *int64  `json:"insurance_id"`
	StatusID             *int    `json:"status_id"`
	Notes                *string `json:"notes"`
	VisitReasonsID       *int    `json:"visit_reasons_id"`
	ReferralSourcesID    *int    `json:"referral_sources_id"`
	// set by handler from JWT
	EmployeeID int
	LocID      int
}

type AppointmentDetailResult struct {
	AppointmentID           int64                  `json:"appointment_id"`
	Location                map[string]interface{} `json:"location"`
	Doctor                  map[string]interface{} `json:"doctor"`
	Patient                 map[string]interface{} `json:"patient"`
	AppointmentDate         string                 `json:"appointment_date"`
	Time                    map[string]interface{} `json:"time"`
	Status                  *string                `json:"status"`
	Notes                   *string                `json:"notes"`
	Warnings                []string               `json:"warnings"`
	QuestionnaireReferralID *int64                 `json:"questionnaire_referral_id"`
	IntakeFormSMS           *IntakeFormSMSResult   `json:"intake_form_sms"`
}

type UpdateAppointmentInput struct {
	AppointmentDate *string `json:"appointment_date"`
	StartTime       *string `json:"start_time"`
	EndTime         *string `json:"end_time"`
	DoctorID        *int64  `json:"doctor_id"`
	Note            *string `json:"note"`
	Notes           *string `json:"notes"`
}

type IntakeFormSMSResult struct {
	Status   string   `json:"status"`
	Link     *string  `json:"link"`
	Warnings []string `json:"warnings"`
}

type LunchResult struct {
	Message    string `json:"message"`
	Date       string `json:"date"`
	LunchStart string `json:"lunch_start"`
	LunchEnd   string `json:"lunch_end"`
}

// ─── Locations ────────────────────────────────────────────────────────────────

func (s *Service) GetLocations(permittedLocIDs []int) ([]LocationItem, error) {
	permSet := make(map[int]struct{}, len(permittedLocIDs))
	for _, id := range permittedLocIDs {
		permSet[id] = struct{}{}
	}

	var stores []locModel.Store
	if err := s.db.Find(&stores).Error; err != nil {
		return nil, err
	}

	var out []LocationItem
	for _, store := range stores {
		var loc locModel.Location
		err := s.db.Where("store_id = ? AND warehouse_id IS NULL AND showcase = true AND store_active = true", store.IDStore).
			First(&loc).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			continue
		}
		if _, ok := permSet[loc.IDLocation]; !ok {
			continue
		}
		fullName := ""
		if store.FullName != nil {
			fullName = *store.FullName
		}
		shortName := ""
		if store.ShortName != nil {
			shortName = *store.ShortName
		}
		out = append(out, LocationItem{
			LocationID: loc.IDLocation,
			FullName:   fullName,
			ShortName:  shortName,
		})
	}
	return out, nil
}

func (s *Service) SetLocation(empID int, locationID int, permittedLocIDs []int) (*SetLocationResult, error) {
	permitted := false
	for _, id := range permittedLocIDs {
		if id == locationID {
			permitted = true
			break
		}
	}
	if !permitted {
		return nil, fmt.Errorf("permission denied for this location")
	}

	var loc locModel.Location
	err := s.db.Where("id_location = ? AND store_active = true AND showcase = true", locationID).First(&loc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("this location is not available for selection")
	}
	if err != nil {
		return nil, err
	}

	if err := s.db.Model(&empModel.Employee{}).Where("id_employee = ?", empID).Update("location_id", int64(locationID)).Error; err != nil {
		return nil, err
	}
	return &SetLocationResult{LocationID: loc.IDLocation, Location: loc.FullName}, nil
}

// ─── Doctors ──────────────────────────────────────────────────────────────────

func (s *Service) GetDoctors(locationID int) ([]DoctorItem, error) {
	var doctors []empModel.Employee
	if err := s.db.Where("job_title_id = 1 AND store_payroll_id = ?", locationID).Find(&doctors).Error; err != nil {
		return nil, err
	}
	out := make([]DoctorItem, len(doctors))
	for i, d := range doctors {
		out[i] = DoctorItem{
			DoctorID: d.IDEmployee,
			Doctor:   fmt.Sprintf("Dr. %s %s", d.FirstName, d.LastName),
		}
	}
	return out, nil
}

// ─── Statuses ─────────────────────────────────────────────────────────────────

func (s *Service) GetStatusAppointments() ([]apptModel.StatusAppointment, error) {
	var statuses []apptModel.StatusAppointment
	return statuses, s.db.Find(&statuses).Error
}

// ─── Appointment list ─────────────────────────────────────────────────────────

type GetAppointmentsInput struct {
	LocationID int
	StartDate  time.Time
	EndDate    time.Time
	EmployeeID *int64
}

func (s *Service) GetAppointments(input GetAppointmentsInput) ([]DaySchedule, error) {
	var loc locModel.Location
	if err := s.db.First(&loc, input.LocationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("location not found")
		}
		return nil, err
	}

	var apptDuration *int
	var settings locModel.LocationAppointmentSettings
	if err := s.db.Where("location_id = ?", input.LocationID).First(&settings).Error; err == nil {
		apptDuration = settings.AppointmentDuration
	}

	workShiftID := int64(0)
	if loc.WorkShiftID != nil {
		workShiftID = int64(*loc.WorkShiftID)
	}

	var result []DaySchedule
	current := input.StartDate
	for !current.After(input.EndDate) {
		dayOfWeek := strings.ToLower(current.Format("Monday"))
		wh := s.getWorkHoursForDate(workShiftID, current, apptDuration)

		apptQuery := s.db.
			Preload("Schedule.Employee").
			Preload("Patient").
			Preload("StatusAppointment").
			Joins("JOIN schedule ON schedule.id_schedule = appointment.schedule_id").
			Where("appointment.location_id = ? AND appointment.appointment_date = ? AND schedule.day_of_week = ?",
				input.LocationID, current.Format("2006-01-02"), dayOfWeek)

		if input.EmployeeID != nil {
			apptQuery = apptQuery.
				Joins("JOIN employee ON employee.id_employee = schedule.employee_id").
				Where("employee.id_employee = ?", *input.EmployeeID)
		}

		var appts []apptModel.Appointment
		apptQuery.Find(&appts)

		// Batch-load DoctorNpiNumber for all doctors in this day
		doctorIDs := map[int64]struct{}{}
		for _, a := range appts {
			if a.Schedule != nil && a.Schedule.Employee != nil {
				doctorIDs[int64(a.Schedule.Employee.IDEmployee)] = struct{}{}
			}
		}
		npiMap := s.loadNpiNumbers(doctorIDs)

		apptList := make([]AppointmentItem, 0, len(appts))
		for _, a := range appts {
			item := AppointmentItem{
				AppointmentID: a.IDAppointment,
				PatientID:     a.PatientID,
				InsuranceID:   a.InsurancePolicyID,
				StatusID:      a.StatusAppointmentID,
				Notes:         a.Notes,
			}
			if a.Patient != nil {
				item.Title = fmt.Sprintf("%s %s", a.Patient.LastName, a.Patient.FirstName)
			}
			if a.StartTime != "" {
				item.Start = truncTime(a.StartTime, 5) // "HH:MM"
			}
			if a.EndTime != "" {
				item.End = truncTime(a.EndTime, 5) // "HH:MM"
			}
			if a.Schedule != nil && a.Schedule.Employee != nil {
				emp := a.Schedule.Employee
				item.DoctorID = int64(emp.IDEmployee)
				npi := npiMap[int64(emp.IDEmployee)]
				item.Doctor = doctorDisplayName(emp, npi)
			}
			if a.StatusAppointment != nil {
				item.Status = a.StatusAppointment.StatusAppointment
			} else {
				item.Status = "Unknown"
			}
			apptList = append(apptList, item)
		}

		result = append(result, DaySchedule{
			Date:              current.Format("2006-01-02"),
			IsWorkingDay:      wh.IsWorkingDay,
			LocationWorkHours: wh,
			Appointments:      apptList,
		})
		current = current.AddDate(0, 0, 1)
	}
	return result, nil
}

// ─── Request appointments ─────────────────────────────────────────────────────

func (s *Service) GetRequestAppointments(locationID int) ([]RequestAppointmentItem, error) {
	var reqs []apptModel.RequestAppointment
	s.db.
		Joins("JOIN employee ON employee.id_employee = request_appointment.doctor_id").
		Where("employee.store_payroll_id = ? AND request_appointment.processed = false", locationID).
		Find(&reqs)

	// Batch-load related records
	insCompanyIDs := map[int64]struct{}{}
	svcTypeIDs := map[int64]struct{}{}
	patientIDs := map[int64]struct{}{}
	for _, r := range reqs {
		if r.InsuranceCompanyID != nil {
			insCompanyIDs[*r.InsuranceCompanyID] = struct{}{}
		}
		if r.ProfessionalServiceTypeID != nil {
			svcTypeIDs[*r.ProfessionalServiceTypeID] = struct{}{}
		}
		if r.PatientID != nil {
			patientIDs[*r.PatientID] = struct{}{}
		}
	}

	insMap := s.loadInsuranceCompanies(insCompanyIDs)
	svcMap := s.loadProfServiceTypes(svcTypeIDs)
	patMap := s.loadPatients(patientIDs)

	out := make([]RequestAppointmentItem, len(reqs))
	for i, r := range reqs {
		item := RequestAppointmentItem{
			IDRequestAppointment:      r.IDRequestAppointment,
			DoctorID:                  r.DoctorID,
			PatientID:                 r.PatientID,
			FirstName:                 r.FirstName,
			LastName:                  r.LastName,
			Accept:                    r.Accept,
			Phone:                     r.Phone,
			Email:                     r.Email,
			ProfessionalServiceTypeID: r.ProfessionalServiceTypeID,
			InsuranceCompanyID:        r.InsuranceCompanyID,
			GroupNumber:               r.GroupNumber,
			MemberNumber:              r.MemberNumber,
			HolderType:                r.HolderType,
			Note:                      r.Note,
		}
		if r.Dob != nil {
			d := r.Dob.Format("2006-01-02")
			item.DOB = &d
		}
		if !r.RequestingDate.IsZero() {
			item.RequestingDate = r.RequestingDate.Format("2006-01-02")
		}
		if r.RequestingTime != "" {
			item.RequestingTime = r.RequestingTime
		}
		if r.PatientID != nil {
			if p, ok := patMap[*r.PatientID]; ok {
				name := strings.TrimSpace(p.FirstName + " " + p.LastName)
				item.PatientName = &name
			}
		}
		if r.InsuranceCompanyID != nil {
			if ic, ok := insMap[*r.InsuranceCompanyID]; ok {
				item.InsuranceCompanyName = &ic.CompanyName
			}
		}
		if r.ProfessionalServiceTypeID != nil {
			if st, ok := svcMap[*r.ProfessionalServiceTypeID]; ok {
				item.ProfessionalServiceTypeTitle = &st.Title
			}
		}
		out[i] = item
	}
	return out, nil
}

func (s *Service) CancelRequestAppointment(reqID int64, locID int, cancelNote string) (*CancelRequestResult, error) {
	var req apptModel.RequestAppointment
	err := s.db.
		Joins("JOIN employee ON employee.id_employee = request_appointment.doctor_id").
		Where("request_appointment.id_request_appointment = ? AND employee.store_payroll_id = ?", reqID, locID).
		First(&req).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("not found")
	}
	if err != nil {
		return nil, err
	}

	// Already cancelled
	if req.Accept != nil && !*req.Accept {
		f := false
		return &CancelRequestResult{
			Message:              "Request appointment already cancelled",
			IDRequestAppointment: req.IDRequestAppointment,
			Accept:               &f,
			Processed:            req.Processed != nil && *req.Processed,
		}, nil
	}

	// Already processed
	if req.Processed != nil && *req.Processed {
		return nil, fmt.Errorf("request appointment already processed")
	}

	f, t := false, true
	req.Accept = &f
	req.Processed = &t

	if cancelNote != "" {
		existing := ""
		if req.Note != nil {
			existing = strings.TrimRight(*req.Note, " \n")
		}
		note := existing
		if existing != "" {
			note += "\n"
		}
		note += fmt.Sprintf("[CANCELLED] %s", cancelNote)
		req.Note = &note
	}

	if err := s.db.Save(&req).Error; err != nil {
		return nil, err
	}
	activitylog.Log(s.db, "appointment", "request_cancel", activitylog.WithEntity(req.IDRequestAppointment))

	fa := false
	return &CancelRequestResult{
		Message:              "Request appointment cancelled",
		IDRequestAppointment: req.IDRequestAppointment,
		Accept:               &fa,
		Processed:            true,
	}, nil
}

// ─── Create appointment ───────────────────────────────────────────────────────

func (s *Service) CreateAppointment(input CreateAppointmentInput) (*AppointmentDetailResult, error) {
	if input.PatientID == nil {
		return nil, fmt.Errorf("patient_id required")
	}
	if input.LocationID == nil {
		return nil, fmt.Errorf("location_id required")
	}
	if input.DoctorID == nil {
		return nil, fmt.Errorf("doctor_id required")
	}
	if input.AppointmentDate == nil || input.StartTime == nil || input.EndTime == nil {
		return nil, fmt.Errorf("date and time must be provided")
	}

	var patient patModel.Patient
	if err := s.db.First(&patient, input.PatientID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("patient not found")
		}
		return nil, err
	}

	var location locModel.Location
	if err := s.db.First(&location, input.LocationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("location not found")
		}
		return nil, err
	}

	// Validate doctor (must have job_title "Doctor")
	var doctor empModel.Employee
	err := s.db.
		Joins("JOIN job_title ON job_title.id_job_title = employee.job_title_id").
		Where("employee.id_employee = ? AND job_title.title = 'Doctor'", *input.DoctorID).
		First(&doctor).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("doctor not found")
	}
	if err != nil {
		return nil, err
	}

	// Load doctor NPI for display name
	npiMap := s.loadNpiNumbers(map[int64]struct{}{int64(doctor.IDEmployee): {}})
	doctorName := doctorDisplayName(&doctor, npiMap[int64(doctor.IDEmployee)])

	desiredDate, err := time.Parse("2006-01-02", *input.AppointmentDate)
	if err != nil {
		return nil, fmt.Errorf("invalid appointment_date format. Use YYYY-MM-DD")
	}
	startTime, err := time.Parse("15:04", *input.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time format. Use HH:MM")
	}
	endTime, err := time.Parse("15:04", *input.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time format. Use HH:MM")
	}

	// 1. Check past time — round current time up to end of minute
	now := time.Now()
	nowRounded := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location())
	requestedDT := time.Date(desiredDate.Year(), desiredDate.Month(), desiredDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, now.Location())
	if requestedDT.Before(nowRounded) {
		return nil, fmt.Errorf("cannot schedule appointment in the past")
	}

	// 2. Check doctor schedule
	dayOfWeek := strings.ToLower(desiredDate.Format("Monday"))
	schedEntry, err := s.findOrCreateSchedule(*input.DoctorID, dayOfWeek)
	if err != nil {
		return nil, err
	}

	schedStart, err := parseTimeStr(schedEntry.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid schedule start_time: %w", err)
	}
	schedEnd, err := parseTimeStr(schedEntry.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid schedule end_time: %w", err)
	}
	if !withinWorkHours(schedStart, startTime, schedEnd) {
		return nil, fmt.Errorf("requested time is outside the doctor's working hours")
	}

	apptDateTime := time.Date(desiredDate.Year(), desiredDate.Month(), desiredDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, time.Local)
	if apptDateTime.Before(time.Now()) {
		return nil, fmt.Errorf("cannot schedule appointment in the past")
	}

	if input.InsuranceID != nil {
		var ins insuranceModel.InsurancePolicy
		if err := s.db.First(&ins, input.InsuranceID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("insurance policy not found")
		}
	}

	var statusAppointment apptModel.StatusAppointment
	if input.StatusID != nil {
		if err := s.db.First(&statusAppointment, input.StatusID).Error; err != nil {
			return nil, fmt.Errorf("status not found")
		}
	} else {
		if err := s.db.Where("status_appointment = ?", "Scheduled").First(&statusAppointment).Error; err != nil {
			return nil, fmt.Errorf("default status 'Scheduled' not found")
		}
	}

	// Optional questionnaire referral
	var questionnaireReferralID *int64
	if input.VisitReasonsID != nil && input.ReferralSourcesID != nil {
		now := time.Now()
		empID64 := int64(input.EmployeeID)
		qr := mktModel.QuestionnaireReferral{
			PatientID:         input.PatientID,
			VisitReasonsID:    input.VisitReasonsID,
			ReferralSourcesID: input.ReferralSourcesID,
			LocationID:        input.LocationID,
			EmployeeID:        &empID64,
			DatetimeCreated:   &now,
		}
		s.db.Create(&qr)
		questionnaireReferralID = &qr.IDQuestionnaireReferral
	}

	schedID := int64(schedEntry.IDSchedule)
	appointment := apptModel.Appointment{
		ScheduleID:          &schedID,
		PatientID:           *input.PatientID,
		LocationID:          *input.LocationID,
		AppointmentDate:     desiredDate,
		StartTime:           startTime.Format("15:04:05"),
		EndTime:             endTime.Format("15:04:05"),
		StatusAppointmentID: statusAppointment.IDStatusAppointment,
		Notes:               input.Notes,
		InsurancePolicyID:   input.InsuranceID,
	}
	if err := s.db.Create(&appointment).Error; err != nil {
		return nil, err
	}

	// Mark request appointment processed
	if input.IDRequestAppointment != nil {
		t := true
		s.db.Model(&apptModel.RequestAppointment{}).
			Where("id_request_appointment = ?", *input.IDRequestAppointment).
			Update("processed", t)
	}

	// SMS
	var warnings []string
	formattedDate := desiredDate.Format("Monday, January 02, 2006")
	formattedStart := strings.TrimLeft(startTime.Format("3:04 PM"), "0")
	formattedEnd := strings.TrimLeft(endTime.Format("3:04 PM"), "0")

	smsMessage := fmt.Sprintf("Your appointment with Dr. %s is scheduled on %s from %s to %s at %s.",
		doctorName, formattedDate, formattedStart, formattedEnd, location.FullName)
	// Try DB template
	if tpl := s.getSMSTemplate("appointment", "confirmation"); tpl != "" {
		if rendered, err := communication.RenderSMSTemplate(tpl, map[string]string{
			"doctor":     doctorName,
			"date":       formattedDate,
			"start_time": formattedStart,
			"end_time":   formattedEnd,
			"location":   location.FullName,
		}); err == nil {
			smsMessage = rendered
		}
	}

	if patient.Phone == nil {
		warnings = append(warnings, "Patient has no phone number. SMS was not sent.")
	} else {
		res := communication.SendSMS(*patient.Phone, smsMessage)
		if res.Status == "accepted" {
			s.logSMSCommunication(patient.IDPatient, smsMessage, "SMS notification sent (new appointment)", input.EmployeeID, input.LocID)
		} else {
			warnings = append(warnings, fmt.Sprintf("SMS failed to send: %s", res.Error))
		}
	}

	activitylog.Log(s.db, "appointment", "create",
		activitylog.WithEntity(appointment.IDAppointment),
		activitylog.WithDetails(map[string]interface{}{
			"patient_id": appointment.PatientID,
			"date":       appointment.AppointmentDate.Format("2006-01-02"),
		}),
	)

	var intakeFormSMS *IntakeFormSMSResult
	if input.NewPatient {
		res := s.sendIntakeFormSMS(appointment.IDAppointment, input.EmployeeID, input.LocID)
		intakeFormSMS = res
		warnings = append(warnings, res.Warnings...)
	}

	statusStr := statusAppointment.StatusAppointment
	var patDOB *string
	if patient.DOB != nil {
		d := patient.DOB.Format("2006-01-02")
		patDOB = &d
	}

	return &AppointmentDetailResult{
		AppointmentID: appointment.IDAppointment,
		Location: map[string]interface{}{
			"location_id":   location.IDLocation,
			"location_name": location.FullName,
		},
		Doctor: map[string]interface{}{
			"id_employee": doctor.IDEmployee,
			"name":        doctorName,
		},
		Patient: map[string]interface{}{
			"patient_id": patient.IDPatient,
			"name":       strings.TrimSpace(patient.FirstName + " " + patient.LastName),
			"dob":        patDOB,
		},
		AppointmentDate: desiredDate.Format("2006-01-02"),
		Time: map[string]interface{}{
			"start": startTime.Format("15:04:05"),
			"end":   endTime.Format("15:04:05"),
		},
		Status:                  &statusStr,
		Notes:                   input.Notes,
		Warnings:                warnings,
		QuestionnaireReferralID: questionnaireReferralID,
		IntakeFormSMS:           intakeFormSMS,
	}, nil
}

// ─── Update appointment ───────────────────────────────────────────────────────

func (s *Service) UpdateAppointment(id int64, input UpdateAppointmentInput) error {
	var appointment apptModel.Appointment
	if err := s.db.Preload("Schedule").First(&appointment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("not found")
		}
		return err
	}
	if input.DoctorID == nil {
		return fmt.Errorf("doctor_id required")
	}

	desiredDateStr := appointment.AppointmentDate.Format("2006-01-02")
	if input.AppointmentDate != nil {
		desiredDateStr = *input.AppointmentDate
	}
	desiredDate, err := time.Parse("2006-01-02", desiredDateStr)
	if err != nil {
		return fmt.Errorf("invalid appointment_date format")
	}

	startTimeStr := truncTime(appointment.StartTime, 5) // "HH:MM"
	if input.StartTime != nil {
		startTimeStr = *input.StartTime
	}
	endTimeStr := truncTime(appointment.EndTime, 5) // "HH:MM"
	if input.EndTime != nil {
		endTimeStr = *input.EndTime
	}
	startTime, _ := time.Parse("15:04", startTimeStr)
	endTime, _ := time.Parse("15:04", endTimeStr)

	// 1. Check past time — round current time up to end of minute
	now := time.Now()
	nowRounded := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()+1, 0, 0, now.Location())
	requestedDT := time.Date(desiredDate.Year(), desiredDate.Month(), desiredDate.Day(),
		startTime.Hour(), startTime.Minute(), 0, 0, now.Location())
	if requestedDT.Before(nowRounded) {
		return fmt.Errorf("cannot schedule appointment in the past")
	}

	// 2. Validate doctor
	var doctor empModel.Employee
	doctorErr := s.db.
		Joins("JOIN job_title ON job_title.id_job_title = employee.job_title_id").
		Where("employee.id_employee = ? AND job_title.title = 'Doctor'", *input.DoctorID).
		First(&doctor).Error
	if errors.Is(doctorErr, gorm.ErrRecordNotFound) {
		return fmt.Errorf("doctor not found")
	}
	if doctorErr != nil {
		return doctorErr
	}

	// 3. Check doctor schedule
	dayOfWeek := strings.ToLower(desiredDate.Format("Monday"))
	schedEntry, err := s.findOrCreateSchedule(*input.DoctorID, dayOfWeek)
	if err != nil {
		return err
	}
	uSchedStart, err := parseTimeStr(schedEntry.StartTime)
	if err != nil {
		return fmt.Errorf("invalid schedule start_time: %w", err)
	}
	uSchedEnd, err := parseTimeStr(schedEntry.EndTime)
	if err != nil {
		return fmt.Errorf("invalid schedule end_time: %w", err)
	}
	if !withinWorkHours(uSchedStart, startTime, uSchedEnd) {
		return fmt.Errorf("requested time is outside the doctor's working hours")
	}

	schedID := int64(schedEntry.IDSchedule)
	appointment.ScheduleID = &schedID
	appointment.AppointmentDate = desiredDate
	appointment.StartTime = startTime.Format("15:04:05")
	appointment.EndTime = endTime.Format("15:04:05")

	note := input.Note
	if note == nil {
		note = input.Notes
	}
	if note != nil {
		appointment.Notes = note
	}

	if err := s.db.Save(&appointment).Error; err != nil {
		return err
	}
	activitylog.Log(s.db, "appointment", "update", activitylog.WithEntity(id))
	return nil
}

func (s *Service) UpdateAppointmentStatus(id int64, statusID int) error {
	var appointment apptModel.Appointment
	if err := s.db.First(&appointment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("not found")
		}
		return err
	}
	var status apptModel.StatusAppointment
	if err := s.db.First(&status, statusID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("status not found")
	}
	if err := s.db.Model(&appointment).Update("status_appointment_id", statusID).Error; err != nil {
		return err
	}
	activitylog.Log(s.db, "appointment", "status_update", activitylog.WithEntity(id))
	return nil
}

func (s *Service) UpdateAppointmentInsurance(id int64, insuranceID *int64) error {
	var appointment apptModel.Appointment
	if err := s.db.First(&appointment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("not found")
		}
		return err
	}

	if insuranceID != nil && *insuranceID != 0 {
		var ins insuranceModel.InsurancePolicy
		if err := s.db.First(&ins, *insuranceID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("insurance policy not found")
		}
		appointment.InsurancePolicyID = insuranceID
	} else {
		appointment.InsurancePolicyID = nil
	}

	if err := s.db.Save(&appointment).Error; err != nil {
		return err
	}
	activitylog.Log(s.db, "appointment", "insurance_update", activitylog.WithEntity(id))
	return nil
}

func (s *Service) DeleteAppointment(id int64) error {
	var appointment apptModel.Appointment
	if err := s.db.First(&appointment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("not found")
		}
		return err
	}
	today := time.Now().Truncate(24 * time.Hour)
	if appointment.AppointmentDate.Before(today) {
		return fmt.Errorf("cannot delete past appointments")
	}
	if err := s.db.Delete(&apptModel.Appointment{}, id).Error; err != nil {
		return err
	}
	activitylog.Log(s.db, "appointment", "delete", activitylog.WithEntity(id))
	return nil
}

// ─── Work hours ───────────────────────────────────────────────────────────────

func (s *Service) GetLocationWorkHours(locationID int, startDate, endDate time.Time) ([]WorkHoursDay, error) {
	var loc locModel.Location
	if err := s.db.First(&loc, locationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("location not found")
		}
		return nil, err
	}
	workShiftID := int64(0)
	if loc.WorkShiftID != nil {
		workShiftID = int64(*loc.WorkShiftID)
	}

	var result []WorkHoursDay
	current := startDate
	for !current.After(endDate) {
		result = append(result, s.getWorkHoursForDate(workShiftID, current, nil))
		current = current.AddDate(0, 0, 1)
	}
	return result, nil
}

// ─── Doctor lunch ─────────────────────────────────────────────────────────────

func (s *Service) SetDoctorLunch(doctorID int64, targetDate time.Time, lunchStart time.Time, durationMin int) (*LunchResult, error) {
	var doctor empModel.Employee
	if err := s.db.First(&doctor, doctorID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("doctor not found")
	}

	dayOfWeek := strings.ToLower(targetDate.Format("Monday"))
	var schedEntry schedModel.Schedule
	if err := s.db.Where("employee_id = ? AND day_of_week = ?", doctorID, dayOfWeek).First(&schedEntry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("doctor is not scheduled to work on this day")
		}
		return nil, err
	}
	if schedEntry.StartTime == "" || schedEntry.EndTime == "" {
		return nil, fmt.Errorf("doctor is not scheduled to work on this day")
	}

	lSchedStart, err := parseTimeStr(schedEntry.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid schedule start_time: %w", err)
	}
	lSchedEnd, err := parseTimeStr(schedEntry.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid schedule end_time: %w", err)
	}

	lunchEnd := lunchStart.Add(time.Duration(durationMin) * time.Minute)
	if !withinWorkHours(lSchedStart, lunchStart, lSchedEnd) ||
		timeMinutes(lunchEnd) > timeMinutes(lSchedEnd) {
		return nil, fmt.Errorf("lunch time is outside of working hours")
	}

	if err := s.db.Model(&schedEntry).Updates(map[string]interface{}{
		"lunch_start": lunchStart,
		"lunch_end":   lunchEnd,
	}).Error; err != nil {
		return nil, err
	}

	return &LunchResult{
		Message:    "Lunch time updated",
		Date:       targetDate.Format("2006-01-02"),
		LunchStart: lunchStart.Format("15:04"),
		LunchEnd:   lunchEnd.Format("15:04"),
	}, nil
}

// ─── Intake form SMS ──────────────────────────────────────────────────────────

func (s *Service) SendIntakeFormSMSForAppointment(appointmentID int64, empID, locID int) (*IntakeFormSMSResult, error) {
	var appointment apptModel.Appointment
	if err := s.db.First(&appointment, appointmentID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("not found")
	}

	if int(appointment.LocationID) != locID {
		return nil, fmt.Errorf("permission denied")
	}

	return s.sendIntakeFormSMS(appointmentID, empID, locID), nil
}

// ─── Private helpers ──────────────────────────────────────────────────────────

func (s *Service) getWorkHoursForDate(workShiftID int64, date time.Time, apptDuration *int) WorkHoursDay {
	day := WorkHoursDay{
		Date:         date.Format("2006-01-02"),
		ApptDuration: apptDuration,
	}

	// Check calendar override first
	if workShiftID != 0 {
		var cal schedModel.Calendar
		err := s.db.Where("work_shift_id = ? AND date = ?", workShiftID, date.Format("2006-01-02")).First(&cal).Error
		if err == nil {
			day.IsWorkingDay = cal.IsWorkingDay
			if cal.IsWorkingDay {
				if cal.TimeStart != nil {
					day.TimeStart = cal.TimeStart
				}
				if cal.TimeEnd != nil {
					day.TimeEnd = cal.TimeEnd
				}
			} else {
				msg := "location is closed on this day"
				day.Message = &msg
			}
			return day
		}
	}

	// Fall back to WorkShift standard schedule
	if workShiftID == 0 {
		return day
	}
	var ws empModel.WorkShift
	if err := s.db.First(&ws, workShiftID).Error; err != nil {
		return day
	}

	type dayData struct {
		working bool
		start   string
		end     string
	}
	wsMap := map[time.Weekday]dayData{
		time.Monday:    {ws.Monday, ws.MondayTimeStart, ws.MondayTimeEnd},
		time.Tuesday:   {ws.Tuesday, ws.TuesdayTimeStart, ws.TuesdayTimeEnd},
		time.Wednesday: {ws.Wednesday, ws.WednesdayTimeStart, ws.WednesdayTimeEnd},
		time.Thursday:  {ws.Thursday, ws.ThursdayTimeStart, ws.ThursdayTimeEnd},
		time.Friday:    {ws.Friday, ws.FridayTimeStart, ws.FridayTimeEnd},
	}
	if ws.SaturdayTimeStart != nil && ws.SaturdayTimeEnd != nil {
		wsMap[time.Saturday] = dayData{ws.Saturday, *ws.SaturdayTimeStart, *ws.SaturdayTimeEnd}
	} else {
		wsMap[time.Saturday] = dayData{ws.Saturday, "", ""}
	}
	if ws.SundayTimeStart != nil && ws.SundayTimeEnd != nil {
		wsMap[time.Sunday] = dayData{ws.Sunday, *ws.SundayTimeStart, *ws.SundayTimeEnd}
	} else {
		wsMap[time.Sunday] = dayData{ws.Sunday, "", ""}
	}

	di := wsMap[date.Weekday()]
	day.IsWorkingDay = di.working
	if di.working && di.start != "" {
		day.TimeStart = &di.start
		day.TimeEnd = &di.end
	} else if !di.working {
		msg := "location is closed on this day"
		day.Message = &msg
	}
	return day
}

func (s *Service) findOrCreateSchedule(doctorID int64, dayOfWeek string) (*schedModel.Schedule, error) {
	var schedEntry schedModel.Schedule
	err := s.db.Where("employee_id = ? AND day_of_week = ?", doctorID, dayOfWeek).First(&schedEntry).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		schedEntry = schedModel.Schedule{
			EmployeeID:          doctorID,
			DayOfWeek:           dayOfWeek,
			StartTime:           "09:00:00",
			EndTime:             "17:00:00",
			AppointmentDuration: 15,
		}
		if err := s.db.Create(&schedEntry).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return &schedEntry, nil
}

func (s *Service) getSMSTemplate(category, name string) string {
	var body string
	s.db.Table("sms_template").
		Where("category = ? AND name = ? AND active = true", category, name).
		Pluck("body", &body)
	return body
}

func (s *Service) logSMSCommunication(patientID int64, content, description string, empID, locID int) {
	var commType generalModel.CommunicationType
	if err := s.db.Where("communication_type = ?", "SMS").First(&commType).Error; err != nil {
		return
	}
	entry := patModel.PatientCommunicationHistory{
		PatientID:             patientID,
		CommunicationTypeID:   commType.CommunicationTypeID,
		CommunicationDatetime: time.Now(),
		Description:           &description,
		Content:               content,
		LocationID:            locID,
		EmployeeID:            int64(empID),
	}
	s.db.Create(&entry)
}

func (s *Service) sendIntakeFormSMS(appointmentID int64, empID, locID int) *IntakeFormSMSResult {
	result := &IntakeFormSMSResult{}

	var appointment apptModel.Appointment
	if err := s.db.Preload("Patient").Preload("Location.Store").First(&appointment, appointmentID).Error; err != nil {
		result.Warnings = append(result.Warnings, "Appointment not found")
		result.Status = "error"
		return result
	}
	if appointment.Patient == nil {
		result.Warnings = append(result.Warnings, "No patient linked to appointment")
		result.Status = "error"
		return result
	}
	if appointment.Location == nil || appointment.Location.Store == nil {
		result.Warnings = append(result.Warnings, "Store not linked to location")
		result.Status = "error"
		return result
	}

	store := appointment.Location.Store
	intakeURL := fmt.Sprintf("https://sighthub.cloud/intake/%s/app/%d", store.Hash, appointmentID)
	result.Link = &intakeURL

	smsMessage := "To save time at your visit, please fill out the intake form before arrival: " + intakeURL
	if tpl := s.getSMSTemplate("appointment", "intake_form"); tpl != "" {
		if rendered, err := communication.RenderSMSTemplate(tpl, map[string]string{
			"intake_url": intakeURL,
		}); err == nil {
			smsMessage = rendered
		}
	}

	if appointment.Patient.Phone == nil {
		result.Warnings = append(result.Warnings, "Patient has no phone number. SMS was not sent.")
		result.Status = "unknown"
		return result
	}

	res := communication.SendSMS(*appointment.Patient.Phone, smsMessage)
	result.Status = res.Status
	if res.Status == "accepted" {
		s.logSMSCommunication(appointment.Patient.IDPatient, smsMessage, "SMS notification sent (intake form)", empID, locID)
	} else {
		result.Warnings = append(result.Warnings, fmt.Sprintf("SMS failed: %s", res.Error))
	}
	return result
}

func (s *Service) loadNpiNumbers(ids map[int64]struct{}) map[int64]*empModel.DoctorNpiNumber {
	if len(ids) == 0 {
		return nil
	}
	idSlice := make([]int64, 0, len(ids))
	for id := range ids {
		idSlice = append(idSlice, id)
	}
	var rows []empModel.DoctorNpiNumber
	s.db.Where("employee_id IN ?", idSlice).Find(&rows)
	m := make(map[int64]*empModel.DoctorNpiNumber, len(rows))
	for i := range rows {
		if rows[i].EmployeeID != nil {
			m[*rows[i].EmployeeID] = &rows[i]
		}
	}
	return m
}

func (s *Service) loadInsuranceCompanies(ids map[int64]struct{}) map[int64]insuranceModel.InsuranceCompany {
	if len(ids) == 0 {
		return nil
	}
	idSlice := make([]int64, 0, len(ids))
	for id := range ids {
		idSlice = append(idSlice, id)
	}
	var rows []insuranceModel.InsuranceCompany
	s.db.Where("id_insurance_company IN ?", idSlice).Find(&rows)
	m := make(map[int64]insuranceModel.InsuranceCompany, len(rows))
	for _, r := range rows {
		m[int64(r.IDInsuranceCompany)] = r
	}
	return m
}

func (s *Service) loadProfServiceTypes(ids map[int64]struct{}) map[int64]svcModel.ProfessionalServiceType {
	if len(ids) == 0 {
		return nil
	}
	idSlice := make([]int64, 0, len(ids))
	for id := range ids {
		idSlice = append(idSlice, id)
	}
	var rows []svcModel.ProfessionalServiceType
	s.db.Where("id_medical_service_type IN ?", idSlice).Find(&rows)
	m := make(map[int64]svcModel.ProfessionalServiceType, len(rows))
	for _, r := range rows {
		m[int64(r.IDMedicalServiceType)] = r
	}
	return m
}

func (s *Service) loadPatients(ids map[int64]struct{}) map[int64]patModel.Patient {
	if len(ids) == 0 {
		return nil
	}
	idSlice := make([]int64, 0, len(ids))
	for id := range ids {
		idSlice = append(idSlice, id)
	}
	var rows []patModel.Patient
	s.db.Where("id_patient IN ?", idSlice).Find(&rows)
	m := make(map[int64]patModel.Patient, len(rows))
	for _, r := range rows {
		m[r.IDPatient] = r
	}
	return m
}

func doctorDisplayName(emp *empModel.Employee, npi *empModel.DoctorNpiNumber) string {
	if npi != nil && npi.PrintingName != nil {
		return *npi.PrintingName
	}
	return strings.TrimSpace(emp.FirstName + " " + emp.LastName)
}

func withinWorkHours(workStart, apptStart, workEnd time.Time) bool {
	return timeMinutes(workStart) <= timeMinutes(apptStart) && timeMinutes(apptStart) < timeMinutes(workEnd)
}

func timeMinutes(t time.Time) int {
	return t.Hour()*60 + t.Minute()
}

// truncTime safely truncates a time string (e.g. "10:00:00") to n characters.
// If the string is shorter than n, it is returned as-is.
func truncTime(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// parseTimeStr parses a time string like "10:00:00" or "10:00" into time.Time.
func parseTimeStr(s string) (time.Time, error) {
	if t, err := time.Parse("15:04:05", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("15:04", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("cannot parse time %q", s)
}
