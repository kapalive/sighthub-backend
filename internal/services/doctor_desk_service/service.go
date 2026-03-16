package doctor_desk_service

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	appointmentModel "sighthub-backend/internal/models/appointment"
	authModel        "sighthub-backend/internal/models/auth"
	empModel         "sighthub-backend/internal/models/employees"
	genModel         "sighthub-backend/internal/models/general"
	locModel         "sighthub-backend/internal/models/location"
	medModel         "sighthub-backend/internal/models/medical"
	assessModel      "sighthub-backend/internal/models/medical/vision_exam/assessment"
	clFitModel       "sighthub-backend/internal/models/medical/vision_exam/cl_fitting"
	medVisionModel   "sighthub-backend/internal/models/medical/vision_exam"
	patModel         "sighthub-backend/internal/models/patients"
	schedModel       "sighthub-backend/internal/models/schedule"
	visionModel      "sighthub-backend/internal/models/vision_exam"
	"sighthub-backend/pkg/activitylog"
	"sighthub-backend/pkg/recall"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ─── helpers ──────────────────────────────────────────────────────────────────

func (s *Service) getEmployeeAndLocation(username string) (*empModel.Employee, *locModel.Location, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, nil, errors.New("employee or location not found")
	}
	var emp empModel.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, nil, errors.New("employee or location not found")
	}
	if emp.LocationID == nil {
		return nil, nil, errors.New("employee or location not found")
	}
	var loc locModel.Location
	if err := s.db.First(&loc, *emp.LocationID).Error; err != nil {
		return nil, nil, errors.New("employee or location not found")
	}
	return &emp, &loc, nil
}

func calculateAge(dob *time.Time) int {
	if dob == nil {
		return 0
	}
	now := time.Now()
	age := now.Year() - dob.Year()
	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		age--
	}
	return age
}

func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// getScheduleForDay returns working day info from Calendar or WorkShift fallback.
func (s *Service) getScheduleForDay(workShiftID int64, date time.Time) (isWorking bool, timeStart, timeEnd *string) {
	var cal schedModel.Calendar
	if err := s.db.Where("work_shift_id = ? AND date::date = ?", workShiftID, date.Format("2006-01-02")).
		First(&cal).Error; err == nil {
		if !cal.IsWorkingDay {
			return false, nil, nil
		}
		var ts, te *string
		if cal.TimeStart != nil {
			v := cal.TimeStart.Format("15:04:05")
			ts = &v
		}
		if cal.TimeEnd != nil {
			v := cal.TimeEnd.Format("15:04:05")
			te = &v
		}
		return true, ts, te
	}

	// Fallback: WorkShift standard schedule
	var ws empModel.WorkShift
	if err := s.db.First(&ws, workShiftID).Error; err != nil {
		return false, nil, nil
	}

	weekday := int(date.Weekday()) // 0=Sunday, 1=Monday, …, 6=Saturday

	type dayDef struct {
		working bool
		start   time.Time
		end     time.Time
		pStart  *time.Time
		pEnd    *time.Time
	}

	var dd dayDef
	switch weekday {
	case 1:
		dd = dayDef{working: ws.Monday, start: ws.MondayTimeStart, end: ws.MondayTimeEnd}
	case 2:
		dd = dayDef{working: ws.Tuesday, start: ws.TuesdayTimeStart, end: ws.TuesdayTimeEnd}
	case 3:
		dd = dayDef{working: ws.Wednesday, start: ws.WednesdayTimeStart, end: ws.WednesdayTimeEnd}
	case 4:
		dd = dayDef{working: ws.Thursday, start: ws.ThursdayTimeStart, end: ws.ThursdayTimeEnd}
	case 5:
		dd = dayDef{working: ws.Friday, start: ws.FridayTimeStart, end: ws.FridayTimeEnd}
	case 6:
		dd = dayDef{working: ws.Saturday, pStart: ws.SaturdayTimeStart, pEnd: ws.SaturdayTimeEnd}
	case 0:
		dd = dayDef{working: ws.Sunday, pStart: ws.SundayTimeStart, pEnd: ws.SundayTimeEnd}
	}

	if !dd.working {
		return false, nil, nil
	}

	var tsStr, teStr string
	if dd.pStart != nil {
		tsStr = dd.pStart.Format("15:04:05")
	} else {
		tsStr = dd.start.Format("15:04:05")
	}
	if dd.pEnd != nil {
		teStr = dd.pEnd.Format("15:04:05")
	} else {
		teStr = dd.end.Format("15:04:05")
	}
	return true, &tsStr, &teStr
}

func (s *Service) doctorName(employeeID int64, emp *empModel.Employee) string {
	var npi empModel.DoctorNpiNumber
	if s.db.Where("employee_id = ?", employeeID).First(&npi).Error == nil && npi.PrintingName != nil {
		return *npi.PrintingName
	}
	if emp != nil {
		return emp.FirstName + " " + emp.LastName
	}
	return "Unknown"
}

// ─── Result types ──────────────────────────────────────────────────────────────

type WorkHoursResult struct {
	TimeStart           *string `json:"time_start"`
	TimeEnd             *string `json:"time_end"`
	LunchStart          interface{} `json:"lunch_start"`
	LunchEnd            interface{} `json:"lunch_end"`
	AppointmentDuration *int    `json:"appointment_duration"`
}

type AppointmentItem struct {
	AppointmentID int64   `json:"appointment_id"`
	Title         string  `json:"title"`
	Start         string  `json:"start"`
	End           string  `json:"end"`
	DoctorID      int64   `json:"doctor_id"`
	Doctor        string  `json:"doctor"`
	Notes         *string `json:"notes"`
	ID            int64   `json:"id"`
	InsuranceID   *int64  `json:"insurance_id"`
	StatusID      int     `json:"status_id"`
	Status        string  `json:"status"`
}

type AppointmentsResult struct {
	Date              string          `json:"date"`
	IsWorkingDay      bool            `json:"is_working_day"`
	LocationWorkHours WorkHoursResult `json:"location_work_hours"`
	Appointments      []AppointmentItem `json:"appointments"`
}

type PatientSearchResult struct {
	IDPatient int64  `json:"id_patient"`
	Name      string `json:"name"`
	Age       int    `json:"age"`
	Address   string `json:"address"`
}

type DoctorResult struct {
	DoctorID int    `json:"doctor_id"`
	Doctor   string `json:"doctor"`
}

type ShowcaseLocationResult struct {
	LocationID int    `json:"location_id"`
	FullName   string `json:"full_name"`
}

type ExamResult struct {
	ExamID   int64   `json:"exam_id"`
	ExamDate *string `json:"exam_date"`
	Doctor   string  `json:"doctor"`
	DoctorID int64   `json:"doctor_id"`
	ExamType string  `json:"exam_type"`
	Passed   bool    `json:"passed"`
}

type UnsignedExamResult struct {
	ExamID      int64   `json:"exam_id"`
	ExamDate    *string `json:"exam_date"`
	ExamType    string  `json:"exam_type"`
	PatientID   int64   `json:"patient_id"`
	PatientName string  `json:"patient_name"`
	LocationID  int     `json:"location_id"`
	Passed      bool    `json:"passed"`
}

type FileResult struct {
	FileID     int64  `json:"file_id"`
	UploadDate string `json:"upload_date"`
	FileName   string `json:"file_name"`
	FilePath   string `json:"file_path"`
}

type FileDetailResult struct {
	FileID     int64  `json:"file_id"`
	UploadDate string `json:"upload_date"`
	FileName   string `json:"file_name"`
	FilePath   string `json:"file_path"`
	PatientID  int64  `json:"patient_id"`
}

type NoteResult struct {
	IDPatientNotes int64   `json:"id_patient_notes"`
	Note           string  `json:"note"`
	Top            bool    `json:"top"`
	AlertDate      *string `json:"alert_date"`
}

type NoteDetailResult struct {
	IDPatientNotes int64   `json:"id_patient_notes"`
	Note           string  `json:"note"`
	Top            bool    `json:"top"`
	AlertDate      *string `json:"alert_date"`
	PatientID      int64   `json:"patient_id"`
}

type InsuranceInfo struct {
	CompanyName *string `json:"company_name"`
	GroupNumber *string `json:"group_number"`
	HolderType  *string `json:"holder_type"`
}

type RecallInfo struct {
	FrontDeskNote *string `json:"front_desk_note"`
	ExpireDate    *string `json:"expire_date"`
}

type PatientInfoResult struct {
	Patient   map[string]interface{} `json:"patient"`
	Insurance InsuranceInfo          `json:"insurance"`
	Recall    []RecallInfo           `json:"recall"`
}

type MedResult struct {
	Year  int    `json:"year"`
	Title string `json:"title"`
}

// ─── GetAppointments ──────────────────────────────────────────────────────────

func (s *Service) GetAppointments(locationID int, date time.Time, employeeID *int64) (*AppointmentsResult, error) {
	var loc locModel.Location
	if err := s.db.First(&loc, locationID).Error; err != nil {
		return nil, errors.New("location not found")
	}

	// Appointment setting
	var setting locModel.LocationAppointmentSettings
	s.db.Where("location_id = ?", locationID).First(&setting)

	// Working hours
	isWorking := false
	var timeStart, timeEnd *string
	if loc.WorkShiftID != nil {
		isWorking, timeStart, timeEnd = s.getScheduleForDay(int64(*loc.WorkShiftID), date)
	}

	dayOfWeek := strings.ToLower(date.Weekday().String())

	var appts []appointmentModel.Appointment
	q := s.db.
		Joins("JOIN schedule ON schedule.id_schedule = appointment.schedule_id").
		Joins("JOIN employee ON employee.id_employee = schedule.employee_id").
		Where("appointment.location_id = ? AND appointment.appointment_date = ? AND schedule.day_of_week = ?",
			locationID, date.Format("2006-01-02"), dayOfWeek).
		Preload("Schedule").
		Preload("Schedule.Employee").
		Preload("Patient").
		Preload("StatusAppointment")
	if employeeID != nil {
		q = q.Where("employee.id_employee = ?", *employeeID)
	}
	q.Find(&appts)

	items := make([]AppointmentItem, 0, len(appts))
	for _, a := range appts {
		title := ""
		if a.Patient != nil {
			title = a.Patient.LastName + " " + a.Patient.FirstName
		}
		status := "Unknown"
		if a.StatusAppointment != nil {
			status = a.StatusAppointment.StatusAppointment
		}
		var docID int64
		docName := "Unknown"
		if a.Schedule != nil && a.Schedule.Employee != nil {
			docID = int64(a.Schedule.Employee.IDEmployee)
			docName = s.doctorName(docID, a.Schedule.Employee)
		}
		items = append(items, AppointmentItem{
			AppointmentID: a.IDAppointment,
			Title:         title,
			Start:         a.StartTime.Format("15:04"),
			End:           a.EndTime.Format("15:04"),
			DoctorID:      docID,
			Doctor:        docName,
			Notes:         a.Notes,
			ID:            a.PatientID,
			InsuranceID:   a.InsurancePolicyID,
			StatusID:      a.StatusAppointmentID,
			Status:        status,
		})
	}

	result := &AppointmentsResult{
		Date:         date.Format("2006-01-02"),
		IsWorkingDay: isWorking,
		LocationWorkHours: WorkHoursResult{
			TimeStart:           timeStart,
			TimeEnd:             timeEnd,
			LunchStart:          nil,
			LunchEnd:            nil,
			AppointmentDuration: setting.AppointmentDuration,
		},
		Appointments: items,
	}
	return result, nil
}

// GetEmployeeLocation returns employee's default location ID, for handler use.
func (s *Service) GetEmployeeLocation(username string) (int, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return 0, err
	}
	return loc.IDLocation, nil
}

// ─── SearchPatients ───────────────────────────────────────────────────────────

func (s *Service) SearchPatients(username, firstName, lastName, chart string) ([]PatientSearchResult, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	q := s.db.Model(&patModel.Patient{}).Where("location_id = ?", loc.IDLocation)
	if firstName != "" {
		q = q.Where("first_name ILIKE ?", firstName+"%")
	}
	if lastName != "" {
		q = q.Where("last_name ILIKE ?", lastName+"%")
	}
	if chart != "" {
		q = q.Where("chart = ?", chart)
	}

	var patients []patModel.Patient
	q.Order("last_name ASC, first_name ASC").Limit(25).Find(&patients)

	results := make([]PatientSearchResult, 0, len(patients))
	for _, p := range patients {
		addr := strings.Trim(safeStr(p.StreetAddress)+", "+safeStr(p.AddressLine2), ", ")
		results = append(results, PatientSearchResult{
			IDPatient: p.IDPatient,
			Name:      p.LastName + ", " + p.FirstName,
			Age:       calculateAge(p.DOB),
			Address:   addr,
		})
	}
	return results, nil
}

// ─── UpdateAppointmentStatus ──────────────────────────────────────────────────

func (s *Service) UpdateAppointmentStatus(appointmentID int64, statusID int) error {
	var appt appointmentModel.Appointment
	if err := s.db.First(&appt, appointmentID).Error; err != nil {
		return errors.New("appointment not found")
	}
	if err := s.db.Model(&appt).Update("status_appointment_id", statusID).Error; err != nil {
		return err
	}
	activitylog.Log(s.db, "appointment", "status_update",
		activitylog.WithEntity(appointmentID),
		activitylog.WithDetails(map[string]interface{}{"new_status_id": statusID}))
	return nil
}

// ─── GetDoctors ───────────────────────────────────────────────────────────────

func (s *Service) GetDoctors(locationID int) ([]DoctorResult, error) {
	var doctors []empModel.Employee
	s.db.Where("job_title_id = 1 AND store_payroll_id = ?", locationID).Find(&doctors)

	results := make([]DoctorResult, 0, len(doctors))
	for _, d := range doctors {
		results = append(results, DoctorResult{
			DoctorID: d.IDEmployee,
			Doctor:   "Dr. " + d.FirstName + " " + d.LastName,
		})
	}
	return results, nil
}

// ─── GetShowcaseLocations ─────────────────────────────────────────────────────

func (s *Service) GetShowcaseLocations() ([]ShowcaseLocationResult, error) {
	var locs []locModel.Location
	s.db.Where("showcase = true AND store_active = true AND warehouse_id IS NULL").Find(&locs)

	results := make([]ShowcaseLocationResult, 0, len(locs))
	for _, l := range locs {
		results = append(results, ShowcaseLocationResult{
			LocationID: l.IDLocation,
			FullName:   l.FullName,
		})
	}
	return results, nil
}

// ─── GetPatientExams ──────────────────────────────────────────────────────────

func (s *Service) GetPatientExams(username string, patientID int64) ([]ExamResult, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var patient patModel.Patient
	if err := s.db.Where("id_patient = ? AND location_id = ?", patientID, loc.IDLocation).
		First(&patient).Error; err != nil {
		return nil, errors.New("patient not found")
	}

	var exams []visionModel.EyeExam
	s.db.Where("patient_id = ? AND location_id = ?", patientID, loc.IDLocation).
		Preload("Employee").Preload("EyeExamType").
		Order("eye_exam_date DESC").Find(&exams)

	results := make([]ExamResult, 0, len(exams))
	for _, e := range exams {
		examType := "Unknown"
		if e.EyeExamType != nil {
			examType = e.EyeExamType.ExamTypeName
		}
		docName := "Unknown"
		if e.Employee != nil {
			docName = s.doctorName(e.EmployeeID, e.Employee)
		}
		d := e.EyeExamDate.Format(time.RFC3339)
		results = append(results, ExamResult{
			ExamID:   e.IDEyeExam,
			ExamDate: &d,
			Doctor:   docName,
			DoctorID: e.EmployeeID,
			ExamType: examType,
			Passed:   e.Passed,
		})
	}
	return results, nil
}

// ─── GetUnsignedExams ─────────────────────────────────────────────────────────

func (s *Service) GetUnsignedExams(username string, locationID *int) ([]UnsignedExamResult, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	locID := loc.IDLocation
	if locationID != nil {
		locID = *locationID
	}

	var exams []visionModel.EyeExam
	s.db.Where("employee_id = ? AND passed = false AND location_id = ?", emp.IDEmployee, locID).
		Preload("EyeExamType").
		Order("eye_exam_date DESC").Find(&exams)

	results := make([]UnsignedExamResult, 0, len(exams))
	for _, e := range exams {
		examType := "Unknown"
		if e.EyeExamType != nil {
			examType = e.EyeExamType.ExamTypeName
		}
		patName := "Unknown"
		var pat patModel.Patient
		if s.db.First(&pat, e.PatientID).Error == nil {
			patName = pat.FirstName + " " + pat.LastName
		}
		d := e.EyeExamDate.Format(time.RFC3339)
		results = append(results, UnsignedExamResult{
			ExamID:      e.IDEyeExam,
			ExamDate:    &d,
			ExamType:    examType,
			PatientID:   e.PatientID,
			PatientName: patName,
			LocationID:  e.LocationID,
			Passed:      e.Passed,
		})
	}
	return results, nil
}

// ─── File CRUD ────────────────────────────────────────────────────────────────

func (s *Service) GetPatientFiles(patientID int64) ([]FileResult, error) {
	var pat patModel.Patient
	if err := s.db.First(&pat, patientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}
	var files []medVisionModel.ExamResultEyesFiles
	s.db.Where("patient_id = ?", patientID).Find(&files)

	results := make([]FileResult, 0, len(files))
	for _, f := range files {
		results = append(results, FileResult{
			FileID:     f.IDExamResultEyesFiles,
			UploadDate: f.DateUpload.Format(time.RFC3339),
			FileName:   f.NameFile,
			FilePath:   f.PathToFile,
		})
	}
	return results, nil
}

func (s *Service) UploadExamFile(patientID int64, filePath, fileName string) (int64, error) {
	// Strip /mnt/tank/data prefix if present
	cleaned := filePath
	if idx := strings.Index(filePath, "/mnt/tank/data/"); idx >= 0 {
		cleaned = filePath[idx+len("/mnt/tank/data/"):]
	}
	cleaned = strings.TrimSpace(cleaned)

	f := medVisionModel.ExamResultEyesFiles{
		PatientID:  patientID,
		PathToFile: cleaned,
		NameFile:   fileName,
	}
	if err := s.db.Create(&f).Error; err != nil {
		return 0, err
	}
	activitylog.Log(s.db, "exam_file", "upload",
		activitylog.WithEntity(patientID),
		activitylog.WithDetails(map[string]interface{}{"filename": fileName}))
	return f.IDExamResultEyesFiles, nil
}

func (s *Service) UpdateExamFile(fileID int64, fileName string) error {
	var f medVisionModel.ExamResultEyesFiles
	if err := s.db.First(&f, fileID).Error; err != nil {
		return errors.New("file not found")
	}
	f.NameFile = fileName
	if err := s.db.Save(&f).Error; err != nil {
		return err
	}
	activitylog.Log(s.db, "exam_file", "update", activitylog.WithEntity(fileID))
	return nil
}

func (s *Service) GetExamFile(fileID int64) (*FileDetailResult, error) {
	var f medVisionModel.ExamResultEyesFiles
	if err := s.db.First(&f, fileID).Error; err != nil {
		return nil, errors.New("file not found")
	}
	return &FileDetailResult{
		FileID:     f.IDExamResultEyesFiles,
		UploadDate: f.DateUpload.Format(time.RFC3339),
		FileName:   f.NameFile,
		FilePath:   f.PathToFile,
		PatientID:  f.PatientID,
	}, nil
}

func (s *Service) DeleteExamFile(fileID int64) error {
	var f medVisionModel.ExamResultEyesFiles
	if err := s.db.First(&f, fileID).Error; err != nil {
		return errors.New("file not found")
	}
	if err := s.db.Delete(&f).Error; err != nil {
		return err
	}
	activitylog.Log(s.db, "exam_file", "delete", activitylog.WithEntity(fileID))
	return nil
}

// ─── Patient Notes CRUD ───────────────────────────────────────────────────────

func (s *Service) GetPatientNotes(patientID int64) ([]NoteResult, error) {
	var pat patModel.Patient
	if err := s.db.First(&pat, patientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}
	var notes []patModel.PatientNotes
	s.db.Where("patient_id = ?", patientID).Find(&notes)

	results := make([]NoteResult, 0, len(notes))
	for _, n := range notes {
		var ad *string
		if n.AlertDate != nil {
			v := n.AlertDate.Format("2006-01-02")
			ad = &v
		}
		results = append(results, NoteResult{
			IDPatientNotes: n.IDPatientNotes,
			Note:           n.Note,
			Top:            n.Top,
			AlertDate:      ad,
		})
	}
	return results, nil
}

func (s *Service) CreatePatientNote(patientID int64, noteText string, top bool, alertDate *time.Time) (int64, error) {
	var pat patModel.Patient
	if err := s.db.First(&pat, patientID).Error; err != nil {
		return 0, errors.New("patient not found")
	}
	note := patModel.PatientNotes{
		Note:      noteText,
		Top:       top,
		AlertDate: alertDate,
		PatientID: patientID,
	}
	if err := s.db.Create(&note).Error; err != nil {
		return 0, err
	}
	var adStr *string
	if alertDate != nil {
		v := alertDate.Format("2006-01-02")
		adStr = &v
	}
	activitylog.Log(s.db, "patient_note", "create",
		activitylog.WithEntity(patientID),
		activitylog.WithDetails(map[string]interface{}{"alert_date": adStr}))
	return note.IDPatientNotes, nil
}

func (s *Service) UpdatePatientNote(patientID, noteID int64, noteText *string, top *bool, alertDate *time.Time, clearAlertDate bool) error {
	var pat patModel.Patient
	if err := s.db.First(&pat, patientID).Error; err != nil {
		return errors.New("patient not found")
	}
	var note patModel.PatientNotes
	if err := s.db.Where("id_patient_notes = ? AND patient_id = ?", noteID, patientID).First(&note).Error; err != nil {
		return errors.New("note not found for this patient")
	}
	if noteText != nil {
		note.Note = *noteText
	}
	if top != nil {
		note.Top = *top
	}
	if alertDate != nil {
		note.AlertDate = alertDate
	} else if clearAlertDate {
		note.AlertDate = nil
	}
	if err := s.db.Save(&note).Error; err != nil {
		return err
	}
	activitylog.Log(s.db, "patient_note", "update", activitylog.WithEntity(noteID))
	return nil
}

func (s *Service) GetPatientNote(patientID, noteID int64) (*NoteDetailResult, error) {
	var pat patModel.Patient
	if err := s.db.First(&pat, patientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}
	var note patModel.PatientNotes
	if err := s.db.Where("id_patient_notes = ? AND patient_id = ?", noteID, patientID).First(&note).Error; err != nil {
		return nil, errors.New("note not found for this patient")
	}
	var ad *string
	if note.AlertDate != nil {
		v := note.AlertDate.Format("2006-01-02")
		ad = &v
	}
	return &NoteDetailResult{
		IDPatientNotes: note.IDPatientNotes,
		Note:           note.Note,
		Top:            note.Top,
		AlertDate:      ad,
		PatientID:      note.PatientID,
	}, nil
}

func (s *Service) DeletePatientNote(patientID, noteID int64) error {
	var pat patModel.Patient
	if err := s.db.First(&pat, patientID).Error; err != nil {
		return errors.New("patient not found")
	}
	var note patModel.PatientNotes
	if err := s.db.Where("id_patient_notes = ? AND patient_id = ?", noteID, patientID).First(&note).Error; err != nil {
		return errors.New("note not found for this patient")
	}
	if err := s.db.Delete(&note).Error; err != nil {
		return err
	}
	activitylog.Log(s.db, "patient_note", "delete", activitylog.WithEntity(noteID))
	return nil
}

// ─── GetPatientInfo ───────────────────────────────────────────────────────────

func (s *Service) GetPatientInfo(patientID int64) (*PatientInfoResult, error) {
	var pat patModel.Patient
	if err := s.db.First(&pat, patientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}

	// Insurance via raw SQL
	var ins struct {
		CompanyName string  `gorm:"column:company_name"`
		GroupNumber *string `gorm:"column:group_number"`
		HolderType  *string `gorm:"column:holder_type"`
	}
	s.db.Raw(`
		SELECT ic.company_name, ip.group_number, ihp.holder_type
		FROM insurance_company ic
		JOIN insurance_policy ip ON ip.insurance_company_id = ic.id_insurance_company
		JOIN insurance_holder_patients ihp ON ip.id_insurance_policy = ihp.insurance_policy_id
		WHERE ihp.patient_id = ?
		LIMIT 1
	`, patientID).Scan(&ins)

	// Recall info via CL fittings
	var exams []visionModel.EyeExam
	s.db.Where("patient_id = ?", patientID).Find(&exams)

	recallInfo := make([]RecallInfo, 0)
	for _, exam := range exams {
		var fittings []clFitModel.ClFitting
		s.db.Where("eye_exam_id = ?", exam.IDEyeExam).Find(&fittings)
		for _, fit := range fittings {
			var ft clFitModel.FirstTrial
			if err := s.db.First(&ft, fit.FirstTrialID).Error; err == nil {
				ri := RecallInfo{FrontDeskNote: ft.FrontDeskNote}
				if ft.ExpireDate != nil {
					ed := ft.ExpireDate.Format("2006-01-02")
					ri.ExpireDate = &ed
				}
				recallInfo = append(recallInfo, ri)
			}
		}
	}

	var dobStr *string
	if pat.DOB != nil {
		v := pat.DOB.Format("2006-01-02")
		dobStr = &v
	}

	return &PatientInfoResult{
		Patient: map[string]interface{}{
			"last_name":      pat.LastName,
			"first_name":     pat.FirstName,
			"street_address": pat.StreetAddress,
			"address_line_2": pat.AddressLine2,
			"city":           pat.City,
			"state":          pat.State,
			"zip_code":       pat.ZipCode,
			"dob":            dobStr,
		},
		Insurance: InsuranceInfo{
			CompanyName: &ins.CompanyName,
			GroupNumber: ins.GroupNumber,
			HolderType:  ins.HolderType,
		},
		Recall: recallInfo,
	}, nil
}

// ─── UpdatePatientInfo ────────────────────────────────────────────────────────

func (s *Service) UpdatePatientInfo(username string, patientID int64, frontDeskNote *string, expireDate *time.Time) error {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}

	var patient patModel.Patient
	if err := s.db.Where("id_patient = ? AND location_id = ?", patientID, loc.IDLocation).
		First(&patient).Error; err != nil {
		return errors.New("patient not found")
	}

	var latestExam visionModel.EyeExam
	if err := s.db.Where("patient_id = ? AND location_id = ?", patientID, loc.IDLocation).
		Order("eye_exam_date DESC").First(&latestExam).Error; err != nil {
		return errors.New("patient has no eye exams in this location")
	}

	var clFitting clFitModel.ClFitting
	if err := s.db.Where("eye_exam_id = ?", latestExam.IDEyeExam).First(&clFitting).Error; err != nil {
		return errors.New("recall information not found")
	}

	var firstTrial clFitModel.FirstTrial
	if err := s.db.First(&firstTrial, clFitting.FirstTrialID).Error; err != nil {
		return errors.New("recall information not found")
	}

	if frontDeskNote != nil {
		firstTrial.FrontDeskNote = frontDeskNote
	}
	if expireDate != nil {
		firstTrial.ExpireDate = expireDate
	}

	if err := s.db.Save(&firstTrial).Error; err != nil {
		return err
	}

	recall.UpsertCLRecall(
		s.db,
		patientID,
		latestExam.IDEyeExam,
		int64(loc.IDLocation),
		firstTrial.ExpireDate,
		firstTrial.FrontDeskNote,
	)

	activitylog.Log(s.db, "patient", "info_update", activitylog.WithEntity(patientID))
	return nil
}

// ─── LogCall ──────────────────────────────────────────────────────────────────

func (s *Service) LogCall(username string, patientID int64, content string) error {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}

	var pat patModel.Patient
	if err := s.db.First(&pat, patientID).Error; err != nil {
		return errors.New("patient not found")
	}

	var ct genModel.CommunicationType
	if err := s.db.Where("communication_type = ?", "Call").First(&ct).Error; err != nil {
		return errors.New("call communication type not found")
	}

	desc := "Phone call logged"
	entry := patModel.PatientCommunicationHistory{
		PatientID:             patientID,
		CommunicationTypeID:   ct.CommunicationTypeID,
		CommunicationDatetime: time.Now(),
		Description:           &desc,
		Content:               content,
		LocationID:            loc.IDLocation,
		EmployeeID:            int64(emp.IDEmployee),
	}
	if err := s.db.Create(&entry).Error; err != nil {
		return err
	}
	activitylog.Log(s.db, "patient", "call", activitylog.WithEntity(patientID))
	return nil
}

// ─── GetMedications ───────────────────────────────────────────────────────────

func (s *Service) GetMedications(patientID int64) ([]MedResult, error) {
	var exams []visionModel.EyeExam
	s.db.Where("patient_id = ?", patientID).Find(&exams)
	if len(exams) == 0 {
		return []MedResult{}, nil
	}

	examIDs := make([]int64, len(exams))
	for i, e := range exams {
		examIDs[i] = e.IDEyeExam
	}

	var meds []medModel.UseMedications
	s.db.Where("eye_exam_id IN ?", examIDs).Find(&meds)
	if len(meds) == 0 {
		return []MedResult{}, nil
	}

	// Build exam ID → year map
	examYear := make(map[int64]int, len(exams))
	for _, e := range exams {
		examYear[e.IDEyeExam] = e.EyeExamDate.Year()
	}

	results := make([]MedResult, 0, len(meds))
	for _, m := range meds {
		year, ok := examYear[m.EyeExamID]
		if !ok {
			continue
		}
		parts := []string{}
		if m.Title != "" {
			parts = append(parts, m.Title)
		}
		if m.FormulationType != nil && *m.FormulationType != "" {
			parts = append(parts, *m.FormulationType)
		}
		if m.Strength != nil && *m.Strength != "" {
			parts = append(parts, *m.Strength)
		}
		results = append(results, MedResult{Year: year, Title: strings.TrimSpace(strings.Join(parts, " "))})
	}

	// Sort by year descending
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Year > results[i].Year {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	return results, nil
}

// ─── GetAllergies ─────────────────────────────────────────────────────────────

func (s *Service) GetAllergies(patientID int64) ([]MedResult, error) {
	var exams []visionModel.EyeExam
	s.db.Where("patient_id = ?", patientID).Find(&exams)
	if len(exams) == 0 {
		return []MedResult{}, nil
	}

	examIDs := make([]int64, len(exams))
	for i, e := range exams {
		examIDs[i] = e.IDEyeExam
	}

	var allergies []medModel.KnownAllergies
	s.db.Where("eye_exam_id IN ?", examIDs).Find(&allergies)
	if len(allergies) == 0 {
		return []MedResult{}, nil
	}

	examYear := make(map[int64]int, len(exams))
	for _, e := range exams {
		examYear[e.IDEyeExam] = e.EyeExamDate.Year()
	}

	seen := make(map[[2]interface{}]bool)
	results := make([]MedResult, 0)
	for _, a := range allergies {
		year, ok := examYear[a.EyeExamID]
		if !ok {
			continue
		}
		key := [2]interface{}{a.Title, year}
		if !seen[key] {
			seen[key] = true
			results = append(results, MedResult{Year: year, Title: a.Title})
		}
	}

	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Year > results[i].Year {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	return results, nil
}

// ─── GetDiagnoses ─────────────────────────────────────────────────────────────

func (s *Service) GetDiagnoses(patientID int64) ([]MedResult, error) {
	var exams []visionModel.EyeExam
	s.db.Where("patient_id = ?", patientID).Find(&exams)
	if len(exams) == 0 {
		return []MedResult{}, nil
	}

	examIDs := make([]int64, len(exams))
	for i, e := range exams {
		examIDs[i] = e.IDEyeExam
	}

	var assessments []assessModel.AssessmentEye
	s.db.Where("eye_exam_id IN ?", examIDs).Find(&assessments)
	if len(assessments) == 0 {
		return []MedResult{}, nil
	}

	assessIDs := make([]int64, len(assessments))
	for i, a := range assessments {
		assessIDs[i] = a.IDAssessmentEye
	}

	// Map: assessment_eye_id → exam_id
	assessExam := make(map[int64]int64, len(assessments))
	for _, a := range assessments {
		assessExam[a.IDAssessmentEye] = a.EyeExamID
	}

	// Map: exam_id → year
	examYear := make(map[int64]int, len(exams))
	for _, e := range exams {
		examYear[e.IDEyeExam] = e.EyeExamDate.Year()
	}

	var diagnoses []assessModel.AssessmentDiagnosis
	s.db.Where("assessment_eye_id IN ?", assessIDs).Find(&diagnoses)
	if len(diagnoses) == 0 {
		return []MedResult{}, nil
	}

	results := make([]MedResult, 0, len(diagnoses))
	for _, d := range diagnoses {
		examID := assessExam[d.AssessmentEyeID]
		year := examYear[examID]
		title := ""
		if d.Title != nil {
			title = *d.Title
		}
		results = append(results, MedResult{Year: year, Title: title})
	}

	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Year > results[i].Year {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	return results, nil
}
