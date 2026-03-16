package referral_service

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	authModel      "sighthub-backend/internal/models/auth"
	empModel       "sighthub-backend/internal/models/employees"
	assessModel    "sighthub-backend/internal/models/medical/vision_exam/assessment"
	extSleModel    "sighthub-backend/internal/models/medical/vision_exam/external_sle"
	posteriorModel "sighthub-backend/internal/models/medical/vision_exam/posterior"
	prelimModel    "sighthub-backend/internal/models/medical/vision_exam/preliminary"
	refrModel      "sighthub-backend/internal/models/medical/vision_exam/refraction"
	referralModel  "sighthub-backend/internal/models/medical/vision_exam/referral"
	patModel       "sighthub-backend/internal/models/patients"
	visionModel    "sighthub-backend/internal/models/vision_exam"
	"sighthub-backend/pkg/activitylog"
	"sighthub-backend/pkg/email"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ─── helpers ──────────────────────────────────────────────────────────────────

func (s *Service) getEmployee(username string) (*empModel.Employee, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, errors.New("employee not found")
	}
	var emp empModel.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, errors.New("employee not found")
	}
	return &emp, nil
}

func (s *Service) validateExamOwnership(emp *empModel.Employee, examID int64) (*visionModel.EyeExam, error) {
	var exam visionModel.EyeExam
	if err := s.db.Where("id_eye_exam = ?", examID).First(&exam).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return nil, errors.New("not authorized")
	}
	return &exam, nil
}

func strVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func doctorFullInfo(d *referralModel.ReferralDoctor) map[string]interface{} {
	if d == nil {
		return nil
	}
	return map[string]interface{}{
		"id_referral_doctor": d.IDReferralDoctor,
		"first_name":         strVal(d.FirstName),
		"last_name":          strVal(d.LastName),
		"fax":                strVal(d.Fax),
	}
}

// ─── input / result types ─────────────────────────────────────────────────────

type SaveReferralInput struct {
	ToReferralDoctorID *int64  `json:"to_referral_doctor_id"`
	CcReferralDoctorID *int64  `json:"cc_referral_doctor_id"`
	TitleLetter        *string `json:"title_letter"`
	IntroLetter        *string `json:"intro_letter"`
	TestsLetter        *string `json:"tests_letter"`
	IssueLetter        *string `json:"issue_letter"`
}

type CreateDoctorInput struct {
	Salutation *string `json:"salutation"`
	Npi        *string `json:"npi"`
	LastName   *string `json:"last_name"`
	FirstName  *string `json:"first_name"`
	Address    *string `json:"address"`
	Address2   *string `json:"address2"`
	City       *string `json:"city"`
	State      *string `json:"state"`
	Zip        *string `json:"zip"`
	Phone      *string `json:"phone"`
	Fax        *string `json:"fax"`
	Email      *string `json:"email"`
}

type ReferralInfo struct {
	Type             string  `json:"type"`
	IDReferralDoctor int64   `json:"id_referral_doctor"`
	FirstName        *string `json:"first_name"`
	LastName         *string `json:"last_name"`
}

type LetterListItem struct {
	IDReferralLetter int64         `json:"id_referral_letter"`
	TitleLetter      *string       `json:"title_letter"`
	ReferralInfo     *ReferralInfo `json:"referral_info"`
}

type GetReferralResult struct {
	ExamID         int64            `json:"exam_id"`
	Exists         bool             `json:"exists"`
	ReferralLetters []LetterListItem `json:"referral_letters"`
}

type LetterDetailResult struct {
	IDReferralLetter   int64                  `json:"id_referral_letter"`
	TitleLetter        *string                `json:"title_letter"`
	IntroLetter        *string                `json:"intro_letter"`
	TestsLetter        *string                `json:"tests_letter"`
	IssueLetter        *string                `json:"issue_letter"`
	EyeExamID          int64                  `json:"eye_exam_id"`
	ToReferralDoctorID *int64                 `json:"to_referral_doctor_id"`
	CcReferralDoctorID *int64                 `json:"cc_referral_doctor_id"`
	ToReferral         map[string]interface{} `json:"to_referral"`
	CcReferral         map[string]interface{} `json:"cc_referral"`
}

type DoctorListResult struct {
	IDReferralDoctor int64   `json:"id_referral_doctor"`
	LastName         *string `json:"last_name"`
	FirstName        *string `json:"first_name"`
	Address          string  `json:"address"`
	Phone            *string `json:"phone"`
	Fax              *string `json:"fax"`
}

// ─── letter helpers ───────────────────────────────────────────────────────────

func (s *Service) letterToDetail(letter *referralModel.ReferralLetter) LetterDetailResult {
	result := LetterDetailResult{
		IDReferralLetter:   letter.IDReferralLetter,
		TitleLetter:        letter.TitleLetter,
		IntroLetter:        letter.IntroLetter,
		TestsLetter:        letter.TestsLetter,
		IssueLetter:        letter.IssueLetter,
		EyeExamID:          letter.EyeExamID,
		ToReferralDoctorID: letter.ToReferralDoctorID,
		CcReferralDoctorID: letter.CcReferralDoctorID,
	}
	if letter.ToReferralDoctorID != nil {
		var doc referralModel.ReferralDoctor
		if s.db.First(&doc, *letter.ToReferralDoctorID).Error == nil {
			result.ToReferral = doctorFullInfo(&doc)
		}
	}
	if letter.CcReferralDoctorID != nil {
		var doc referralModel.ReferralDoctor
		if s.db.First(&doc, *letter.CcReferralDoctorID).Error == nil {
			result.CcReferral = doctorFullInfo(&doc)
		}
	}
	return result
}

// ─── referral letter CRUD ─────────────────────────────────────────────────────

func (s *Service) SaveReferral(username string, examID int64, input SaveReferralInput) (map[string]interface{}, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}
	exam, err := s.validateExamOwnership(emp, examID)
	if err != nil {
		return nil, err
	}
	if exam.Passed {
		return nil, errors.New("cannot create referral letter for a completed exam")
	}

	if input.ToReferralDoctorID != nil {
		var doc referralModel.ReferralDoctor
		if err := s.db.First(&doc, *input.ToReferralDoctorID).Error; err != nil {
			return nil, fmt.Errorf("to referral doctor with ID %d not found", *input.ToReferralDoctorID)
		}
	}
	if input.CcReferralDoctorID != nil {
		var doc referralModel.ReferralDoctor
		if err := s.db.First(&doc, *input.CcReferralDoctorID).Error; err != nil {
			return nil, fmt.Errorf("cc referral doctor with ID %d not found", *input.CcReferralDoctorID)
		}
	}

	letter := referralModel.ReferralLetter{
		EyeExamID:          examID,
		ToReferralDoctorID: input.ToReferralDoctorID,
		CcReferralDoctorID: input.CcReferralDoctorID,
		TitleLetter:        input.TitleLetter,
		IntroLetter:        input.IntroLetter,
		TestsLetter:        input.TestsLetter,
		IssueLetter:        input.IssueLetter,
	}
	if err := s.db.Create(&letter).Error; err != nil {
		return nil, err
	}
	activitylog.Log(s.db, "referral", "create", activitylog.WithEntity(examID))
	return map[string]interface{}{
		"id_referral_letter":   letter.IDReferralLetter,
		"title_letter":         letter.TitleLetter,
		"intro_letter":         letter.IntroLetter,
		"tests_letter":         letter.TestsLetter,
		"issue_letter":         letter.IssueLetter,
		"eye_exam_id":          letter.EyeExamID,
		"to_referral_doctor_id": letter.ToReferralDoctorID,
		"cc_referral_doctor_id": letter.CcReferralDoctorID,
	}, nil
}

func (s *Service) GetReferral(examID int64) (GetReferralResult, error) {
	var exam visionModel.EyeExam
	if err := s.db.Where("id_eye_exam = ?", examID).First(&exam).Error; err != nil {
		return GetReferralResult{}, errors.New("exam not found")
	}
	var letters []referralModel.ReferralLetter
	s.db.Where("eye_exam_id = ?", examID).Find(&letters)

	if len(letters) == 0 {
		return GetReferralResult{ExamID: examID, Exists: false, ReferralLetters: []LetterListItem{}}, nil
	}

	items := make([]LetterListItem, 0, len(letters))
	for _, l := range letters {
		item := LetterListItem{
			IDReferralLetter: l.IDReferralLetter,
			TitleLetter:      l.TitleLetter,
		}
		if l.ToReferralDoctorID != nil {
			var doc referralModel.ReferralDoctor
			if s.db.First(&doc, *l.ToReferralDoctorID).Error == nil {
				item.ReferralInfo = &ReferralInfo{
					Type:             "to",
					IDReferralDoctor: doc.IDReferralDoctor,
					FirstName:        doc.FirstName,
					LastName:         doc.LastName,
				}
			}
		} else if l.CcReferralDoctorID != nil {
			var doc referralModel.ReferralDoctor
			if s.db.First(&doc, *l.CcReferralDoctorID).Error == nil {
				item.ReferralInfo = &ReferralInfo{
					Type:             "cc",
					IDReferralDoctor: doc.IDReferralDoctor,
					FirstName:        doc.FirstName,
					LastName:         doc.LastName,
				}
			}
		}
		items = append(items, item)
	}
	return GetReferralResult{ExamID: examID, Exists: true, ReferralLetters: items}, nil
}

func (s *Service) GetReferralLetterByID(examID, letterID int64) (LetterDetailResult, error) {
	var exam visionModel.EyeExam
	if err := s.db.Where("id_eye_exam = ?", examID).First(&exam).Error; err != nil {
		return LetterDetailResult{}, errors.New("exam not found")
	}
	var letter referralModel.ReferralLetter
	if err := s.db.Where("eye_exam_id = ? AND id_referral_letter = ?", examID, letterID).First(&letter).Error; err != nil {
		return LetterDetailResult{}, errors.New("referral letter not found")
	}
	return s.letterToDetail(&letter), nil
}

func (s *Service) UpdateReferralLetter(username string, examID, letterID int64, rawData map[string]interface{}) (LetterDetailResult, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return LetterDetailResult{}, err
	}
	exam, err := s.validateExamOwnership(emp, examID)
	if err != nil {
		return LetterDetailResult{}, err
	}
	if exam.Passed {
		return LetterDetailResult{}, errors.New("cannot update: exam has already been completed")
	}

	var letter referralModel.ReferralLetter
	if err := s.db.Where("eye_exam_id = ? AND id_referral_letter = ?", examID, letterID).First(&letter).Error; err != nil {
		return LetterDetailResult{}, errors.New("referral letter not found")
	}

	updates := map[string]interface{}{}
	if v, ok := rawData["title_letter"]; ok {
		if s, ok := v.(string); ok {
			updates["title_letter"] = s
		} else {
			updates["title_letter"] = nil
		}
	}
	if v, ok := rawData["intro_letter"]; ok {
		if s, ok := v.(string); ok {
			updates["intro_letter"] = s
		} else {
			updates["intro_letter"] = nil
		}
	}
	if v, ok := rawData["tests_letter"]; ok {
		if s, ok := v.(string); ok {
			updates["tests_letter"] = s
		} else {
			updates["tests_letter"] = nil
		}
	}
	if v, ok := rawData["issue_letter"]; ok {
		if s, ok := v.(string); ok {
			updates["issue_letter"] = s
		} else {
			updates["issue_letter"] = nil
		}
	}

	// Handle doctor ID fields (allow null to clear)
	if v, ok := rawData["to_referral_doctor_id"]; ok {
		if v == nil {
			updates["to_referral_doctor_id"] = nil
		} else {
			id, err := toInt64(v)
			if err != nil {
				return LetterDetailResult{}, errors.New("invalid to_referral_doctor_id")
			}
			var doc referralModel.ReferralDoctor
			if err := s.db.First(&doc, id).Error; err != nil {
				return LetterDetailResult{}, fmt.Errorf("referral doctor with ID %d not found", id)
			}
			updates["to_referral_doctor_id"] = id
		}
	}
	if v, ok := rawData["cc_referral_doctor_id"]; ok {
		if v == nil {
			updates["cc_referral_doctor_id"] = nil
		} else {
			id, err := toInt64(v)
			if err != nil {
				return LetterDetailResult{}, errors.New("invalid cc_referral_doctor_id")
			}
			var doc referralModel.ReferralDoctor
			if err := s.db.First(&doc, id).Error; err != nil {
				return LetterDetailResult{}, fmt.Errorf("referral doctor with ID %d not found", id)
			}
			updates["cc_referral_doctor_id"] = id
		}
	}

	if len(updates) > 0 {
		if err := s.db.Model(&letter).Updates(updates).Error; err != nil {
			return LetterDetailResult{}, err
		}
	}
	activitylog.Log(s.db, "referral", "update", activitylog.WithEntity(examID))

	// reload letter
	s.db.Where("eye_exam_id = ? AND id_referral_letter = ?", examID, letterID).First(&letter)
	return s.letterToDetail(&letter), nil
}

func toInt64(v interface{}) (int64, error) {
	switch val := v.(type) {
	case float64:
		return int64(val), nil
	case int64:
		return val, nil
	case int:
		return int64(val), nil
	default:
		return 0, errors.New("not a number")
	}
}

func (s *Service) DeleteReferralLetterOrDoctor(username string, examID, letterID int64, doctorType string) (string, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return "", err
	}
	if _, err := s.validateExamOwnership(emp, examID); err != nil {
		return "", err
	}

	var letter referralModel.ReferralLetter
	if err := s.db.Where("eye_exam_id = ? AND id_referral_letter = ?", examID, letterID).First(&letter).Error; err != nil {
		return "", errors.New("referral letter not found")
	}

	if doctorType != "" {
		if doctorType != "to" && doctorType != "cc" {
			return "", errors.New("invalid doctor_type: use 'to' or 'cc'")
		}
		updates := map[string]interface{}{}
		if doctorType == "to" {
			updates["to_referral_doctor_id"] = nil
		} else {
			updates["cc_referral_doctor_id"] = nil
		}
		if err := s.db.Model(&letter).Updates(updates).Error; err != nil {
			return "", err
		}
		activitylog.Log(s.db, "referral", "delete", activitylog.WithEntity(letter.IDReferralLetter))
		return fmt.Sprintf("Referral doctor (%s) removed successfully", doctorType), nil
	}

	if err := s.db.Delete(&letter).Error; err != nil {
		return "", err
	}
	activitylog.Log(s.db, "referral", "delete", activitylog.WithEntity(letterID))
	return "Referral letter deleted successfully", nil
}

// ─── doctors CRUD ─────────────────────────────────────────────────────────────

func (s *Service) GetAllReferralDoctors() ([]DoctorListResult, error) {
	var docs []referralModel.ReferralDoctor
	if err := s.db.Find(&docs).Error; err != nil {
		return nil, err
	}
	result := make([]DoctorListResult, 0, len(docs))
	for _, d := range docs {
		parts := []string{strVal(d.Address), strVal(d.Address2), strVal(d.City), strVal(d.State), strVal(d.Zip)}
		filtered := []string{}
		for _, p := range parts {
			if p != "" {
				filtered = append(filtered, p)
			}
		}
		result = append(result, DoctorListResult{
			IDReferralDoctor: d.IDReferralDoctor,
			LastName:         d.LastName,
			FirstName:        d.FirstName,
			Address:          strings.Join(filtered, ", "),
			Phone:            d.Phone,
			Fax:              d.Fax,
		})
	}
	return result, nil
}

func (s *Service) CreateReferralDoctor(input CreateDoctorInput) (*referralModel.ReferralDoctor, error) {
	doc := referralModel.ReferralDoctor{
		Salutation: input.Salutation,
		Npi:        input.Npi,
		LastName:   input.LastName,
		FirstName:  input.FirstName,
		Address:    input.Address,
		Address2:   input.Address2,
		City:       input.City,
		State:      input.State,
		Zip:        input.Zip,
		Phone:      input.Phone,
		Fax:        input.Fax,
		Email:      input.Email,
	}
	if err := s.db.Create(&doc).Error; err != nil {
		return nil, err
	}
	activitylog.Log(s.db, "referral", "doctor_create", activitylog.WithEntity(doc.IDReferralDoctor))
	return &doc, nil
}

func (s *Service) UpdateReferralDoctor(doctorID int64, input CreateDoctorInput) (*referralModel.ReferralDoctor, error) {
	var doc referralModel.ReferralDoctor
	if err := s.db.First(&doc, doctorID).Error; err != nil {
		return nil, errors.New("referral doctor not found")
	}
	if input.Salutation != nil {
		doc.Salutation = input.Salutation
	}
	if input.Npi != nil {
		doc.Npi = input.Npi
	}
	if input.LastName != nil {
		doc.LastName = input.LastName
	}
	if input.FirstName != nil {
		doc.FirstName = input.FirstName
	}
	if input.Address != nil {
		doc.Address = input.Address
	}
	if input.Address2 != nil {
		doc.Address2 = input.Address2
	}
	if input.City != nil {
		doc.City = input.City
	}
	if input.State != nil {
		doc.State = input.State
	}
	if input.Zip != nil {
		doc.Zip = input.Zip
	}
	if input.Phone != nil {
		doc.Phone = input.Phone
	}
	if input.Fax != nil {
		doc.Fax = input.Fax
	}
	if input.Email != nil {
		doc.Email = input.Email
	}
	if err := s.db.Save(&doc).Error; err != nil {
		return nil, err
	}
	activitylog.Log(s.db, "referral", "doctor_update", activitylog.WithEntity(doctorID))
	return &doc, nil
}

func (s *Service) DeleteReferralDoctor(doctorID int64) error {
	var doc referralModel.ReferralDoctor
	if err := s.db.First(&doc, doctorID).Error; err != nil {
		return errors.New("referral doctor not found")
	}
	if err := s.db.Delete(&doc).Error; err != nil {
		return err
	}
	activitylog.Log(s.db, "referral", "doctor_delete", activitylog.WithEntity(doctorID))
	return nil
}

// ─── HTML / print ─────────────────────────────────────────────────────────────

func (s *Service) GetReferralLetterHTML(letterID int64) (string, error) {
	var letter referralModel.ReferralLetter
	if err := s.db.First(&letter, letterID).Error; err != nil {
		return "", errors.New("referral letter not found")
	}
	var exam visionModel.EyeExam
	if err := s.db.Where("id_eye_exam = ?", letter.EyeExamID).First(&exam).Error; err != nil {
		return "", errors.New("eye exam not found")
	}

	var toDoc, ccDoc map[string]interface{}
	if letter.ToReferralDoctorID != nil {
		var d referralModel.ReferralDoctor
		if s.db.First(&d, *letter.ToReferralDoctorID).Error == nil {
			toDoc = doctorFullInfo(&d)
		}
	}
	if letter.CcReferralDoctorID != nil {
		var d referralModel.ReferralDoctor
		if s.db.First(&d, *letter.CcReferralDoctorID).Error == nil {
			ccDoc = doctorFullInfo(&d)
		}
	}

	// Patient
	patient := map[string]interface{}{}
	var pat patModel.Patient
	if s.db.First(&pat, exam.PatientID).Error == nil {
		addrParts := []string{}
		for _, p := range []*string{pat.StreetAddress, pat.AddressLine2, pat.City, pat.State, pat.ZipCode} {
			if p != nil && *p != "" {
				addrParts = append(addrParts, *p)
			}
		}
		dobStr := "N/A"
		if pat.DOB != nil {
			dobStr = pat.DOB.Format("01/02/2006")
		}
		phone := "N/A"
		if pat.Phone != nil {
			phone = *pat.Phone
		}
		patient = map[string]interface{}{
			"first_name": pat.FirstName,
			"last_name":  pat.LastName,
			"dob":        dobStr,
			"address":    strings.Join(addrParts, ", "),
			"phone":      phone,
		}
	}

	// Doctor / employee
	doctor := map[string]interface{}{
		"first_name":    "N/A",
		"last_name":     "N/A",
		"printing_name": "N/A",
		"job_title":     "N/A",
		"dr_npi_number": "N/A",
	}
	var emp empModel.Employee
	if s.db.First(&emp, exam.EmployeeID).Error == nil {
		jobTitle := "N/A"
		if emp.JobTitleID != nil {
			var jt empModel.JobTitle
			if s.db.First(&jt, *emp.JobTitleID).Error == nil {
				jobTitle = jt.Title
			}
		}
		printingName := fmt.Sprintf("%s %s", emp.FirstName, emp.LastName)
		npiNum := "N/A"
		var npiRec empModel.DoctorNpiNumber
		if s.db.Where("employee_id = ?", emp.IDEmployee).First(&npiRec).Error == nil {
			if npiRec.PrintingName != nil && *npiRec.PrintingName != "" {
				printingName = *npiRec.PrintingName
			}
			npiNum = npiRec.DRNPINumber
		}
		doctor = map[string]interface{}{
			"first_name":    emp.FirstName,
			"last_name":     emp.LastName,
			"printing_name": printingName,
			"job_title":     jobTitle,
			"dr_npi_number": npiNum,
		}
	}

	// Provider / location
	provider := map[string]interface{}{}
	var loc visionModel.EyeExam // reuse to avoid circular; actually use location model directly
	_ = loc
	// query location directly
	type Location struct {
		IDLocation    int     `gorm:"column:id_location"`
		FullName      string  `gorm:"column:full_name"`
		StreetAddress *string `gorm:"column:street_address"`
		AddressLine2  *string `gorm:"column:address_line_2"`
		City          *string `gorm:"column:city"`
		State         *string `gorm:"column:state"`
		PostalCode    *string `gorm:"column:postal_code"`
		Phone         *string `gorm:"column:phone"`
	}
	var location Location
	if s.db.Table("location").Where("id_location = ?", exam.LocationID).First(&location).Error == nil {
		locParts := []string{}
		for _, p := range []*string{location.StreetAddress, location.AddressLine2, location.City, location.State, location.PostalCode} {
			if p != nil && *p != "" {
				locParts = append(locParts, *p)
			}
		}
		phone := "N/A"
		if location.Phone != nil {
			phone = *location.Phone
		}
		provider = map[string]interface{}{
			"name":    location.FullName,
			"address": strings.Join(locParts, ", "),
			"phone":   phone,
		}
	}
	doctor["provider"] = provider

	// Drawing from ExternalSleEye
	var drawingPath interface{} = nil
	var extSle extSleModel.ExternalSleEye
	if s.db.Where("eye_exam_id = ?", exam.IDEyeExam).First(&extSle).Error == nil && extSle.AddDrawing != nil {
		drawingPath = *extSle.AddDrawing
	}

	title := "Referral Letter"
	if letter.TitleLetter != nil {
		title = *letter.TitleLetter
	}
	intro := "No introduction provided."
	if letter.IntroLetter != nil {
		intro = *letter.IntroLetter
	}
	tests := "No test details available."
	if letter.TestsLetter != nil {
		tests = *letter.TestsLetter
	}
	issue := "No issue details provided."
	if letter.IssueLetter != nil {
		issue = *letter.IssueLetter
	}

	// Convert \n to <br> for HTML rendering
	nl2br := func(s string) template.HTML {
		escaped := template.HTMLEscapeString(s)
		return template.HTML(strings.ReplaceAll(escaped, "\n", "<br>\n"))
	}

	ctx := map[string]interface{}{
		"title":       title,
		"intro":       nl2br(intro),
		"tests":       nl2br(tests),
		"drawing":     drawingPath,
		"issue":       nl2br(issue),
		"to_doctor":   toDoc,
		"cc_doctor":   ccDoc,
		"from_doctor": doctor,
		"patient":     patient,
	}

	return s.renderReferralTemplate(ctx)
}

func (s *Service) renderReferralTemplate(ctx map[string]interface{}) (string, error) {
	// Try to load template from disk
	tmplDir := os.Getenv("PDF_TEMPLATES_DIR")
	if tmplDir == "" {
		// default: relative to binary
		ex, _ := os.Executable()
		tmplDir = filepath.Join(filepath.Dir(ex), "internal", "templates", "pdf")
	}
	tmplPath := filepath.Join(tmplDir, "referral_template.html")

	var tmplContent []byte
	var err error
	tmplContent, err = os.ReadFile(tmplPath)
	if err != nil {
		// fallback: use simple inline template
		return "", fmt.Errorf("template not found at %s: %w", tmplPath, err)
	}

	tmpl, err := template.New("referral").Parse(string(tmplContent))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ─── build tests ──────────────────────────────────────────────────────────────

func (s *Service) BuildTests(examID int64, params map[string]string) (string, string, error) {
	emp, err := s.buildGetEmployee(examID)
	if err != nil {
		return "", "", err
	}

	var exam visionModel.EyeExam
	if err := s.db.Where("id_eye_exam = ?", examID).First(&exam).Error; err != nil {
		return "", "", errors.New("exam not found")
	}

	var responseText strings.Builder
	drawingPath := ""

	if params["uva_dist"] == "true" {
		var prelim prelimModel.PreliminaryEyeExam
		if s.db.Where("eye_exam_id = ?", examID).First(&prelim).Error == nil && prelim.UnaidedVADistanceID != nil {
			var uva prelimModel.UnaidedVADistance
			if s.db.First(&uva, *prelim.UnaidedVADistanceID).Error == nil {
				responseText.WriteString(fmt.Sprintf("\n\nDistance Unaided VA:\nOD (20/): %s\nOS (20/): %s\nOU (20/): %s\n",
					strVal(uva.Od20), strVal(uva.Os20), strVal(uva.Ou20)))
			} else {
				responseText.WriteString("\n\nUVA Dist:\nNo data available in UnaidedVADistance.")
			}
		} else {
			responseText.WriteString("\n\nUVA Dist:\nNo related PreliminaryEyeExam found.")
		}
	}

	if params["uva_near"] == "true" {
		var prelim prelimModel.PreliminaryEyeExam
		if s.db.Where("eye_exam_id = ?", examID).First(&prelim).Error == nil && prelim.UnaidedVANearID != nil {
			var uva prelimModel.UnaidedVANear
			if s.db.First(&uva, *prelim.UnaidedVANearID).Error == nil {
				responseText.WriteString(fmt.Sprintf("\n\nNear Unaid VA:\nOD (20/): %s\nOS (20/): %s\nOU (20/): %s\n",
					strVal(uva.Od20), strVal(uva.Os20), strVal(uva.Ou20)))
			} else {
				responseText.WriteString("\n\nUVA Near:\nNo data available in UnaidedVANear.")
			}
		} else {
			responseText.WriteString("\n\nUVA Near:\nNo related PreliminaryEyeExam found.")
		}
	}

	if params["bcva"] == "true" {
		var refr refrModel.RefractionEye
		if s.db.Where("eye_exam_id = ?", examID).First(&refr).Error == nil {
			var final refrModel.RefractionFinal
			if s.db.First(&final, refr.FinalID).Error == nil {
				responseText.WriteString(fmt.Sprintf("\n\nDistance Corrected VA:\nOD (DVA 20/): %s\nOS (DVA 20/): %s\nOU (DVA 20/): %s\n",
					strVal(final.OdDva20), strVal(final.OsDva20), strVal(final.OuDva20)))
			} else {
				responseText.WriteString("\n\nBCVA:\nNo data available in Final.")
			}
		} else {
			responseText.WriteString("\n\nBCVA:\nNo related RefractionEye found.")
		}
	}

	if params["cover_test"] == "true" {
		var prelim prelimModel.PreliminaryEyeExam
		if s.db.Where("eye_exam_id = ?", examID).First(&prelim).Error == nil {
			responseText.WriteString(fmt.Sprintf("\n\nCover Test:\nDistance Cover Test: %s\n",
				strVal(prelim.DistanceCoverTest)))
		}
	}

	if params["normal_findings_test"] == "true" {
		var prelim prelimModel.PreliminaryEyeExam
		if s.db.Where("eye_exam_id = ?", examID).First(&prelim).Error == nil {
			responseText.WriteString(fmt.Sprintf("\n\nNormal:\nNear Cover Test: %s\n",
				strVal(prelim.NearCoverTest)))
		}
	}

	if params["external_sle"] == "true" {
		var extSles []extSleModel.ExternalSleEye
		s.db.Where("eye_exam_id = ?", examID).Find(&extSles)
		if len(extSles) > 0 {
			for _, sle := range extSles {
				var findings extSleModel.FindingsExternalSle
				if s.db.First(&findings, sle.FindingsExternalSleID).Error == nil {
					responseText.WriteString(fmt.Sprintf(
						"\n\nFindings External SLE:\nExternals:\n  %s\n\nLids Lashes:\n  OD: %s\n  OS: %s\n\nConjunctiva Sclera:\n  OD: %s\n  OS: %s\n\nCornea:\n  OD: %s\n  OS: %s\n\nTear Film:\n  OD: %s\n  OS: %s\n\nAnterior Chamber:\n  OD: %s\n  OS: %s\n\nIris:\n  OD: %s\n  OS: %s\n\nLens:\n  OD: %s\n  OS: %s\n",
						strOrNo(findings.Externals),
						strOrNo(findings.OdLidsLashes), strOrNo(findings.OsLidsLashes),
						strOrNo(findings.OdConjunctivaSclera), strOrNo(findings.OsConjunctivaSclera),
						strOrNo(findings.OdCornea), strOrNo(findings.OsCornea),
						strOrNo(findings.OdTearFilm), strOrNo(findings.OsTearFilm),
						strOrNo(findings.OdAnteriorChamber), strOrNo(findings.OsAnteriorChamber),
						strOrNo(findings.OdIris), strOrNo(findings.OsIris),
						strOrNo(findings.OdLens), strOrNo(findings.OsLens),
					))
				}
			}
		} else {
			responseText.WriteString("\n\nFindings External SLE:\nNo related ExternalSleEye found.\n")
		}
	}

	if params["normal_findings_extern"] == "true" {
		responseText.WriteString("\n\nNormal Findings:\nData for Normal Findings (External / SLE): not specified")
	}

	if params["externalsle_draw"] == "true" {
		var extSle extSleModel.ExternalSleEye
		if s.db.Where("eye_exam_id = ?", examID).First(&extSle).Error == nil && extSle.AddDrawing != nil {
			drawingPath = *extSle.AddDrawing
			responseText.WriteString(fmt.Sprintf("\n\nExternal / SLE Drawing:\n%s", *extSle.AddDrawing))
		} else {
			responseText.WriteString("\n\nExternal / SLE Drawing:\nNo drawing available.")
		}
	}

	if params["externalsle_notes"] == "true" {
		var extSle extSleModel.ExternalSleEye
		if s.db.Where("eye_exam_id = ?", examID).First(&extSle).Error == nil {
			responseText.WriteString(fmt.Sprintf("\n\nNotes (External / SLE):\n%s\n", strOrNo(extSle.Note)))
		} else {
			responseText.WriteString("\n\nNotes (External / SLE):\nNo related ExternalSleEye found.\n")
		}
	}

	if params["posterior_findings"] == "true" {
		var posteriorEye posteriorModel.PosteriorEye
		if s.db.Where("eye_exam_id = ?", examID).First(&posteriorEye).Error == nil {
			var findings posteriorModel.FindingsPosterior
			if s.db.First(&findings, posteriorEye.FindingsPosteriorID).Error == nil {
				responseText.WriteString(fmt.Sprintf(
					"\n\nPosterior Findings:\nView:\n  OD: %s\n  OS: %s\n\nVitreous:\n  OD: %s\n  OS: %s\n\nMacula:\n  OD: %s\n  OS: %s\n\nBackground:\n  OD: %s\n  OS: %s\n\nVessels:\n  OD: %s\n  OS: %s\n\nPeripheral Fundus:\n  OD: %s\n  OS: %s\n\nOptic Nerve:\n  OD: %s\n  OS: %s\n",
					strOrNo(findings.OdView), strOrNo(findings.OsView),
					strOrNo(findings.OdVitreous), strOrNo(findings.OsVitreous),
					strOrNo(findings.OdMacula), strOrNo(findings.OsMacula),
					strOrNo(findings.OdBackground), strOrNo(findings.OsBackground),
					strOrNo(findings.OdVessels), strOrNo(findings.OsVessels),
					strOrNo(findings.OdPeripheralFundus), strOrNo(findings.OsPeripheralFundus),
					strOrNo(findings.OdOpticsNerve), strOrNo(findings.OsOpticsNerve),
				))
			} else {
				responseText.WriteString("\n\nPosterior Findings:\nNo related FindingsPosterior found.\n")
			}
		} else {
			responseText.WriteString("\n\nPosterior Findings:\nNo related PosteriorEye found.\n")
		}
	}

	if params["posterior_draw"] == "true" {
		responseText.WriteString("\n\nPosterior Drawing:\nData for Posterior Drawing: not specified")
	}

	if params["normal_findings_poster"] == "true" {
		responseText.WriteString("\n\nNormal Findings:\nData for Normal Findings (Posterior Findings): not specified")
	}

	if params["poster_notes"] == "true" {
		var posteriorEye posteriorModel.PosteriorEye
		if s.db.Where("eye_exam_id = ?", examID).First(&posteriorEye).Error == nil {
			responseText.WriteString(fmt.Sprintf("\n\nNotes (Posterior Findings):\n%s\n", strOrNo(posteriorEye.Note)))
		} else {
			responseText.WriteString("\n\nNotes (Posterior Findings):\nNo related PosteriorEye found.\n")
		}
	}

	if params["special_test"] == "true" {
		responseText.WriteString("\n\nSpecial Testing:\nData for Special Testing: not specified")
	}

	if params["special_notes"] == "true" {
		responseText.WriteString("\n\nNotes:\nData for Notes (Special Testing): not specified")
	}

	if params["tonometry"] == "true" {
		endDate := params["data_end"]
		startDate := params["data_start"]
		if endDate == "" {
			endDate = time.Now().Format("2006-01-02")
		}
		if startDate == "" {
			startDate = time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
		}

		var extSle extSleModel.ExternalSleEye
		if s.db.Where("eye_exam_id = ?", examID).First(&extSle).Error != nil {
			responseText.WriteString("\n\nTonometry:\nNo related ExternalSleEye found.\n")
		} else {
			var tonos []extSleModel.TonometryEye
			s.db.Where("external_sle_eye_id = ? AND date_tonometry_eye >= ? AND date_tonometry_eye <= ?",
				extSle.IDExternalSleEye, startDate, endDate).Find(&tonos)
			if len(tonos) > 0 {
				responseText.WriteString(fmt.Sprintf("\n\nTonometry (Start: %s, End: %s):\n", startDate, endDate))
				for _, t := range tonos {
					dateStr := "No data"
					if t.DateTonometryEye != nil {
						dateStr = t.DateTonometryEye.Format("2006-01-02")
					}
					timeStr := "No data"
					if t.TimeTonometryEye != nil {
						timeStr = *t.TimeTonometryEye
					}
					responseText.WriteString(fmt.Sprintf("Method: %s\nDate: %s\nTime: %s\nOD: %s\nOS: %s\n\n",
						strOrNo(t.MethodTonometry), dateStr, timeStr,
						strOrNo(t.OdTonometryEye), strOrNo(t.OsTonometryEye)))
				}
			} else {
				responseText.WriteString(fmt.Sprintf("\n\nTonometry (Start: %s, End: %s):\nNo data found for the specified period.\n", startDate, endDate))
			}
		}
	}

	if params["refraction_final"] == "true" {
		var refr refrModel.RefractionEye
		if s.db.Where("eye_exam_id = ?", examID).First(&refr).Error == nil {
			var final refrModel.RefractionFinal
			if s.db.First(&final, refr.FinalID).Error == nil {
				expireStr := "No data"
				if final.ExpireDate != nil {
					expireStr = final.ExpireDate.Format("2006-01-02")
				}
				responseText.WriteString(fmt.Sprintf(
					"\n\nRefraction (Final):\nOD SPH: %s\nOS SPH: %s\nOD CYL: %s\nOS CYL: %s\nOD Axis: %s\nOS Axis: %s\nOD Add: %s\nOS Add: %s\nOD Horizontal Prism: %s\nOS Horizontal Prism: %s\nOD Vertical Prism: %s\nOS Vertical Prism: %s\nOD DVA (20/): %s\nOS DVA (20/): %s\nOU DVA (20/): %s\nOD NVA (20/): %s\nOS NVA (20/): %s\nOU NVA (20/): %s\nOD PD: %s\nOS PD: %s\nOU PD: %s\nOD NPD: %s\nOS NPD: %s\nOU NPD: %s\nExpire Date: %s\n",
					strOrNo(final.OdSph), strOrNo(final.OsSph),
					strOrNo(final.OdCyl), strOrNo(final.OsCyl),
					strOrNo(final.OdAxis), strOrNo(final.OsAxis),
					strOrNo(final.OdAdd), strOrNo(final.OsAdd),
					strOrNo(final.OdHPrism), strOrNo(final.OsHPrism),
					strOrNo(final.OdVPrism), strOrNo(final.OsVPrism),
					strOrNo(final.OdDva20), strOrNo(final.OsDva20), strOrNo(final.OuDva20),
					strOrNo(final.OdNva20), strOrNo(final.OsNva20), strOrNo(final.OuNva20),
					strOrNo(final.OdPd), strOrNo(final.OsPd), strOrNo(final.OuPd),
					strOrNo(final.OdNpd), strOrNo(final.OsNpd), strOrNo(final.OuNpd),
					expireStr,
				))
			} else {
				responseText.WriteString("\n\nRefraction (Final):\nNo related Final record found.\n")
			}
		} else {
			responseText.WriteString("\n\nRefraction (Final):\nNo related RefractionEye found.\n")
		}
	}

	if params["diagnosis"] == "true" {
		var assessments []assessModel.AssessmentEye
		s.db.Where("eye_exam_id = ?", examID).Find(&assessments)
		if len(assessments) == 0 {
			responseText.WriteString("\n\nDiagnosis:\nNo assessments found.\n")
		} else {
			var diagList []string
			for _, a := range assessments {
				var diagnoses []assessModel.AssessmentDiagnosis
				s.db.Where("assessment_eye_id = ?", a.IDAssessmentEye).Find(&diagnoses)
				for _, d := range diagnoses {
					code := "No code"
					if d.Code != nil {
						code = *d.Code
					}
					title := "No title"
					if d.Title != nil {
						title = *d.Title
					}
					diagList = append(diagList, fmt.Sprintf("%s: %s", code, title))
				}
			}
			if len(diagList) > 0 {
				responseText.WriteString("\n\nDiagnosis:\n" + strings.Join(diagList, "\n") + "\n")
			} else {
				responseText.WriteString("\n\nDiagnosis:\nNo diagnoses found.\n")
			}
		}
	}

	text := strings.TrimSpace(responseText.String())
	if text == "" {
		return "", "", errors.New("no selected parameters")
	}

	// NPI
	npi := "Not available"
	var npiRec empModel.DoctorNpiNumber
	if s.db.Where("employee_id = ?", emp.IDEmployee).First(&npiRec).Error == nil {
		npi = npiRec.DRNPINumber
	}
	text += fmt.Sprintf("\n\nNPI: %s", npi)

	return text, drawingPath, nil
}

// buildGetEmployee loads employee for the exam owner (used in BuildTests)
func (s *Service) buildGetEmployee(examID int64) (*empModel.Employee, error) {
	var exam visionModel.EyeExam
	if err := s.db.Where("id_eye_exam = ?", examID).First(&exam).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	var emp empModel.Employee
	if err := s.db.First(&emp, exam.EmployeeID).Error; err != nil {
		return nil, errors.New("employee not found")
	}
	return &emp, nil
}

func strOrNo(p *string) string {
	if p == nil || *p == "" {
		return "No data"
	}
	return *p
}

// ─── SendReferralEmail ───────────────────────────────────────────────────────

func (s *Service) SendReferralEmail(letterID int64, toEmail, subject string) error {
	// Get letter to find exam → location
	var letter referralModel.ReferralLetter
	if err := s.db.First(&letter, letterID).Error; err != nil {
		return errors.New("letter not found")
	}

	var exam visionModel.EyeExam
	if err := s.db.First(&exam, letter.EyeExamID).Error; err != nil {
		return errors.New("exam not found")
	}

	htmlBody, err := s.GetReferralLetterHTML(letterID)
	if err != nil {
		return fmt.Errorf("failed to render letter: %w", err)
	}

	// Use location-specific SMTP config
	locID := int64(exam.LocationID)
	smtpCfg, err := email.GetSMTPForLocation(s.db, &locID)
	if err != nil {
		return fmt.Errorf("smtp not configured: %w", err)
	}

	if subject == "" {
		subject = "Referral Letter"
	}

	return email.Send(smtpCfg, toEmail, subject, htmlBody)
}
