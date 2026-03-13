package exam_eye_service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	authModel    "sighthub-backend/internal/models/auth"
	empModel     "sighthub-backend/internal/models/employees"
	locModel     "sighthub-backend/internal/models/location"
	noteModel    "sighthub-backend/internal/models/medical/vision_exam"
	patModel     "sighthub-backend/internal/models/patients"
	visionModel  "sighthub-backend/internal/models/vision_exam"
	"sighthub-backend/pkg/activitylog"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ─── helpers ──────────────────────────────────────────────────────────────────

func (s *Service) getEmployeeAndLocation(username string) (*empModel.Employee, *locModel.Location, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("username = ?", username).First(&login).Error; err != nil {
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

func getSubRoutes(examTypeID int64) []string {
	mapping := map[int64][]string{
		1:  {"/history", "/cc_hpi", "/preliminary", "/refraction", "/cl_fitting", "/external_sle", "/posterior", "/special", "/assessment", "/referral", "/super"},
		2:  {"/history", "/cc_hpi", "/preliminary", "/refraction", "/cl_fitting", "/external_sle", "/posterior", "/assessment", "/referral", "/super"},
		3:  {"/preliminary", "/cl_fitting", "/external_sle", "/assessment", "/super"},
		4:  {"/history", "/preliminary", "/cl_fitting", "/external_sle", "/assessment", "/super"},
		5:  {"/external_sle", "/posterior", "/assessment", "/super"},
		6:  {"/history", "/cc_hpi", "/external_sle", "/posterior", "/assessment", "/referral", "/super"},
		7:  {"/cc_hpi", "/refraction", "/assessment", "/super"},
		8:  {"/cc_hpi", "/refraction", "/assessment", "/super"},
		9:  {"/assessment", "/super"},
		10: {"/cc_hpi", "/assessment", "/super"},
	}
	routes, ok := mapping[examTypeID]
	if !ok {
		return []string{}
	}
	return routes
}

func (s *Service) doctorName(employeeID int64, emp *empModel.Employee) string {
	var npi empModel.DoctorNpiNumber
	if err := s.db.Where("employee_id = ?", employeeID).First(&npi).Error; err == nil {
		if npi.PrintingName != nil && *npi.PrintingName != "" {
			return *npi.PrintingName
		}
	}
	return emp.FirstName + " " + emp.LastName
}

// ─── GetExamTypes ─────────────────────────────────────────────────────────────

func (s *Service) GetExamTypes() ([]map[string]interface{}, error) {
	var types []visionModel.EyeExamType
	if err := s.db.Find(&types).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(types))
	for i, t := range types {
		out[i] = t.ToMap()
	}
	return out, nil
}

// ─── StartNewExam ─────────────────────────────────────────────────────────────

type NewExamResult struct {
	ExamID     int64    `json:"exam_id"`
	ExamName   string   `json:"exam_name"`
	PatientAge int      `json:"patient_age"`
	SubRoutes  []string `json:"sub_routes"`
}

func (s *Service) StartNewExam(username string, patientID, examTypeID int64) (*NewExamResult, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var patient patModel.Patient
	if err := s.db.First(&patient, patientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}

	var examType visionModel.EyeExamType
	if err := s.db.Where("id_eye_exam_type = ?", examTypeID).First(&examType).Error; err != nil {
		return nil, errors.New("exam type not found")
	}

	exam := visionModel.EyeExam{
		EyeExamDate:   time.Now().UTC(),
		EmployeeID:    int64(emp.IDEmployee),
		EyeExamTypeID: examTypeID,
		LocationID:    loc.IDLocation,
		PatientID:     patientID,
		Passed:        false,
	}
	if err := s.db.Create(&exam).Error; err != nil {
		return nil, err
	}

	activitylog.Log(s.db, "exam", "create",
		activitylog.WithEntity(exam.IDEyeExam),
		activitylog.WithDetails(map[string]interface{}{
			"patient_id":        patientID,
			"eye_exam_type_id":  examTypeID,
			"exam_type_name":    examType.ExamTypeName,
		}),
	)

	return &NewExamResult{
		ExamID:     exam.IDEyeExam,
		ExamName:   examType.ExamTypeName,
		PatientAge: calculateAge(patient.DOB),
		SubRoutes:  getSubRoutes(examTypeID),
	}, nil
}

// ─── SubmitExam ───────────────────────────────────────────────────────────────

func (s *Service) SubmitExam(username string, examID int64) error {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}

	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return errors.New("exam not found")
	}
	if exam.Passed {
		return errors.New("cannot update a completed exam")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return errors.New("you are not authorized to update this exam")
	}

	if err := s.db.Model(&exam).Update("passed", true).Error; err != nil {
		return err
	}
	activitylog.Log(s.db, "exam", "submit",
		activitylog.WithEntity(examID),
		activitylog.WithDetails(map[string]interface{}{"patient_id": exam.PatientID}),
	)
	return nil
}

// ─── GetExamDetails ───────────────────────────────────────────────────────────

type ExamDetailResult struct {
	IDExam     int64    `json:"id_exam"`
	LocationID int      `json:"location_id"`
	Location   string   `json:"location"`
	EmployeeID int64    `json:"employee_id"`
	Doctor     string   `json:"doctor"`
	PatientID  int64    `json:"patient_id"`
	Patient    string   `json:"patient"`
	Passed     bool     `json:"passed"`
	Date       string   `json:"date"`
	SubRoutes  []string `json:"sub_routes"`
}

func (s *Service) GetExamDetails(examID int64) (*ExamDetailResult, error) {
	var exam visionModel.EyeExam
	if err := s.db.Preload("Location").Preload("Employee").First(&exam, examID).Error; err != nil {
		return nil, errors.New("eye exam not found")
	}
	if exam.Location == nil {
		return nil, errors.New("location not found")
	}
	if exam.Employee == nil {
		return nil, errors.New("employee not found")
	}

	var patient patModel.Patient
	if err := s.db.First(&patient, exam.PatientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}

	locName := "Unknown Location"
	if exam.Location.FullName != "" {
		locName = exam.Location.FullName
	}

	return &ExamDetailResult{
		IDExam:     exam.IDEyeExam,
		LocationID: exam.LocationID,
		Location:   locName,
		EmployeeID: exam.EmployeeID,
		Doctor:     s.doctorName(exam.EmployeeID, exam.Employee),
		PatientID:  exam.PatientID,
		Patient:    patient.FirstName + " " + patient.LastName,
		Passed:     exam.Passed,
		Date:       exam.EyeExamDate.Format(time.RFC3339),
		SubRoutes:  getSubRoutes(exam.EyeExamTypeID),
	}, nil
}

// ─── CancelExam ───────────────────────────────────────────────────────────────

func (s *Service) CancelExam(username string, examID int64) error {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}

	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return errors.New("exam not found")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return errors.New("you are not authorized to cancel this exam")
	}
	if exam.Passed {
		return errors.New("cannot cancel a completed exam")
	}

	activitylog.Log(s.db, "exam", "cancel",
		activitylog.WithEntity(examID),
		activitylog.WithDetails(map[string]interface{}{"patient_id": exam.PatientID}),
	)
	return s.db.Delete(&exam).Error
}

// ─── UnlockExam ───────────────────────────────────────────────────────────────

func (s *Service) UnlockExam(username string, examID int64) error {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}

	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return errors.New("exam not found")
	}
	if !exam.Passed {
		return errors.New("exam is already unlocked")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return errors.New("you are not authorized to unlock this exam")
	}

	if err := s.db.Model(&exam).Update("passed", false).Error; err != nil {
		return err
	}
	activitylog.Log(s.db, "exam", "unlock",
		activitylog.WithEntity(examID),
		activitylog.WithDetails(map[string]interface{}{"patient_id": exam.PatientID}),
	)
	return nil
}

// ─── Notes ───────────────────────────────────────────────────────────────────

func noteToMap(n noteModel.ExamEyeNotes) map[string]interface{} {
	return map[string]interface{}{
		"id_exam_eye_notes":  n.IDExamEyeNotes,
		"exam_eye_note_doc_id": n.ExamEyeNoteDocID,
		"note":               n.Note,
		"priority":           n.Priority,
	}
}

func (s *Service) GetNotes(username, objectName, fieldName string) ([]map[string]interface{}, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var table noteModel.ExamEyeNoteTable
	if err := s.db.Where("table_name = ?", objectName).First(&table).Error; err != nil {
		return nil, errors.New("table '" + objectName + "' not found")
	}

	var col noteModel.ExamEyeNoteCol
	if err := s.db.Where("column_name = ? AND exam_eye_note_table_id = ?", fieldName, table.IDExamEyeNoteTable).
		First(&col).Error; err != nil {
		return nil, errors.New("column '" + fieldName + "' not found in table '" + objectName + "'")
	}

	var docs []noteModel.ExamEyeNoteDoc
	if err := s.db.Where("exam_eye_note_col_id = ? AND employee_id = ?", col.IDExamEyeNoteCol, int64(emp.IDEmployee)).
		Find(&docs).Error; err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return []map[string]interface{}{}, nil
	}

	docIDs := make([]int64, len(docs))
	for i, d := range docs {
		docIDs[i] = d.IDExamEyeNoteDoc
	}

	var notes []noteModel.ExamEyeNotes
	if err := s.db.Where("exam_eye_note_doc_id IN ?", docIDs).
		Order("priority DESC").Find(&notes).Error; err != nil {
		return nil, err
	}

	out := make([]map[string]interface{}, len(notes))
	for i, n := range notes {
		out[i] = noteToMap(n)
	}
	return out, nil
}

func (s *Service) AddNote(username, objectName, fieldName, noteContent string, priority int) (map[string]interface{}, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var table noteModel.ExamEyeNoteTable
	if err := s.db.Where("table_name = ?", objectName).First(&table).Error; err != nil {
		return nil, errors.New("table '" + objectName + "' not found")
	}

	var col noteModel.ExamEyeNoteCol
	if err := s.db.Where("column_name = ? AND exam_eye_note_table_id = ?", fieldName, table.IDExamEyeNoteTable).
		First(&col).Error; err != nil {
		return nil, errors.New("column '" + fieldName + "' not found in table '" + objectName + "'")
	}

	var doc noteModel.ExamEyeNoteDoc
	if err := s.db.Where("exam_eye_note_col_id = ? AND employee_id = ?", col.IDExamEyeNoteCol, int64(emp.IDEmployee)).
		First(&doc).Error; err != nil {
		doc = noteModel.ExamEyeNoteDoc{
			ExamEyeNoteColID: col.IDExamEyeNoteCol,
			EmployeeID:       int64(emp.IDEmployee),
		}
		if err := s.db.Create(&doc).Error; err != nil {
			return nil, err
		}
	}

	newNote := noteModel.ExamEyeNotes{
		ExamEyeNoteDocID: doc.IDExamEyeNoteDoc,
		Note:             noteContent,
		Priority:         priority,
	}
	if err := s.db.Create(&newNote).Error; err != nil {
		return nil, err
	}
	return noteToMap(newNote), nil
}

func (s *Service) UpdateNote(noteID int64, priority *int) (map[string]interface{}, error) {
	var note noteModel.ExamEyeNotes
	if err := s.db.First(&note, noteID).Error; err != nil {
		return nil, errors.New("note not found")
	}

	if priority != nil {
		var conflict noteModel.ExamEyeNotes
		if err := s.db.Where("priority = ? AND exam_eye_note_doc_id = ?", *priority, note.ExamEyeNoteDocID).
			First(&conflict).Error; err == nil && conflict.IDExamEyeNotes != note.IDExamEyeNotes {
			// swap priorities
			oldPriority := note.Priority
			if err := s.db.Model(&conflict).Update("priority", oldPriority).Error; err != nil {
				return nil, err
			}
			if err := s.db.Model(&note).Update("priority", *priority).Error; err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"message":          "Priority swapped successfully",
				"updated_note":     noteToMap(note),
				"conflicting_note": noteToMap(conflict),
			}, nil
		}
		if err := s.db.Model(&note).Update("priority", *priority).Error; err != nil {
			return nil, err
		}
	}
	return noteToMap(note), nil
}

func (s *Service) DeleteNote(username string, noteID int64) error {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}

	var note noteModel.ExamEyeNotes
	if err := s.db.First(&note, noteID).Error; err != nil {
		return errors.New("note not found")
	}

	var doc noteModel.ExamEyeNoteDoc
	if err := s.db.First(&doc, note.ExamEyeNoteDocID).Error; err != nil {
		return errors.New("associated note document not found")
	}
	if doc.EmployeeID != int64(emp.IDEmployee) {
		return errors.New("you do not have permission to delete this note")
	}

	return s.db.Delete(&note).Error
}

// ─── ChangeExamType ───────────────────────────────────────────────────────────

type ChangeExamTypeResult struct {
	Message        string   `json:"message"`
	ExamID         int64    `json:"exam_id"`
	OldExamTypeID  int64    `json:"old_exam_type_id,omitempty"`
	NewExamTypeID  int64    `json:"new_exam_type_id"`
	NewExamTypeName string  `json:"new_exam_type_name,omitempty"`
	SubRoutes      []string `json:"sub_routes"`
}

func (s *Service) ChangeExamType(username string, examID, newTypeID int64) (*ChangeExamTypeResult, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return nil, errors.New("you are not authorized to change this exam")
	}
	if exam.Passed {
		return nil, errors.New("cannot change type of a completed exam")
	}

	var newType visionModel.EyeExamType
	if err := s.db.Where("id_eye_exam_type = ?", newTypeID).First(&newType).Error; err != nil {
		return nil, errors.New("exam type not found")
	}

	oldTypeID := exam.EyeExamTypeID
	if oldTypeID == newTypeID {
		return &ChangeExamTypeResult{
			Message:       "Exam type unchanged",
			ExamID:        exam.IDEyeExam,
			NewExamTypeID: exam.EyeExamTypeID,
			SubRoutes:     getSubRoutes(exam.EyeExamTypeID),
		}, nil
	}

	if err := s.db.Model(&exam).Update("eye_exam_type_id", newTypeID).Error; err != nil {
		return nil, err
	}
	activitylog.Log(s.db, "exam", "change_type",
		activitylog.WithEntity(examID),
		activitylog.WithDetails(map[string]interface{}{
			"old_type_id":   oldTypeID,
			"new_type_id":   newTypeID,
			"new_type_name": newType.ExamTypeName,
		}),
	)

	return &ChangeExamTypeResult{
		Message:         "Exam type updated successfully",
		ExamID:          exam.IDEyeExam,
		OldExamTypeID:   oldTypeID,
		NewExamTypeID:   newTypeID,
		NewExamTypeName: newType.ExamTypeName,
		SubRoutes:       getSubRoutes(newTypeID),
	}, nil
}
