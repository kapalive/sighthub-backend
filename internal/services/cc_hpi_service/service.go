package cc_hpi_service

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/pkg/activitylog"

	empLoginModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	visionModel "sighthub-backend/internal/models/vision_exam"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ─── helpers ─────────────────────────────────────────────────────────────────

func (s *Service) getEmployee(username string) (*empModel.Employee, error) {
	var login empLoginModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, errors.New("employee not found")
	}
	var emp empModel.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, errors.New("employee not found")
	}
	return &emp, nil
}

func (s *Service) getExam(examID int64) (*visionModel.EyeExam, error) {
	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	return &exam, nil
}

// ─── Input types ─────────────────────────────────────────────────────────────

type ChiefComplaintInput struct {
	NoteChiefComplaint *string `json:"note_chief_complaint"`
	Location           *string `json:"location"`
	Quality            *string `json:"quality"`
	Severity           *string `json:"severity"`
	Duration           *string `json:"duration"`
	Timing             *string `json:"timing"`
	Context            *string `json:"context"`
	Factors            *string `json:"factors"`
	Symptoms           *string `json:"symptoms"`
}

type SecondaryComplaintInput struct {
	NoteSecondaryComplaint *string `json:"note_secondary_complaint"`
	Location               *string `json:"location"`
	Quality                *string `json:"quality"`
	Severity               *string `json:"severity"`
	Duration               *string `json:"duration"`
	Timing                 *string `json:"timing"`
	Context                *string `json:"context"`
	Factors                *string `json:"factors"`
	Symptoms               *string `json:"symptoms"`
}

type SaveCcHpiInput struct {
	ChiefComplaintHpiEye     ChiefComplaintInput     `json:"chief_complaint_hpi_eye"`
	SecondaryComplaintHpiEye SecondaryComplaintInput `json:"secondary_complaint_hpi_eye"`
	ChiefComplaintNote       *string                 `json:"chief_complaint_note"`
}

type UpdateCcHpiInput struct {
	ChiefComplaintHpiEye     *ChiefComplaintInput     `json:"chief_complaint_hpi_eye"`
	SecondaryComplaintHpiEye *SecondaryComplaintInput `json:"secondary_complaint_hpi_eye"`
	ChiefComplaintNote       *string                  `json:"chief_complaint_note"`
}

// ─── SaveCcHpi ────────────────────────────────────────────────────────────────

func (s *Service) SaveCcHpi(username string, examID int64, input SaveCcHpiInput) (map[string]interface{}, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}

	exam, err := s.getExam(examID)
	if err != nil {
		return nil, err
	}

	if exam.EmployeeID != int64(emp.IDEmployee) {
		return nil, errors.New("you are not authorized to update this exam")
	}
	if exam.Passed {
		return nil, errors.New("cannot create cc_hpi for a completed exam")
	}

	// Check if CcHpiEye already exists
	var existing visionModel.CcHpiEye
	if err := s.db.Where("eye_exam_id = ?", examID).First(&existing).Error; err == nil {
		return nil, errors.New("cc_hpi already exists for this exam")
	}

	// Create ChiefComplaintHPIEye
	d := input.ChiefComplaintHpiEye
	chief := visionModel.ChiefComplaintHPIEye{
		NoteChiefComplaint: d.NoteChiefComplaint,
		Location:           d.Location,
		Quality:            d.Quality,
		Severity:           d.Severity,
		Duration:           d.Duration,
		Timing:             d.Timing,
		Context:            d.Context,
		Factors:            d.Factors,
		Symptoms:           d.Symptoms,
	}
	if err := s.db.Create(&chief).Error; err != nil {
		return nil, err
	}

	// Create SecondaryComplaintHPIEye
	sd := input.SecondaryComplaintHpiEye
	secondary := visionModel.SecondaryComplaintHPIEye{
		NoteSecondaryComplaint: sd.NoteSecondaryComplaint,
		Location:               sd.Location,
		Quality:                sd.Quality,
		Severity:               sd.Severity,
		Duration:               sd.Duration,
		Timing:                 sd.Timing,
		Context:                sd.Context,
		Factors:                sd.Factors,
		Symptoms:               sd.Symptoms,
	}
	if err := s.db.Create(&secondary).Error; err != nil {
		return nil, err
	}

	// Create CcHpiEye
	ccHpi := visionModel.CcHpiEye{
		EyeExamID:                  examID,
		ChiefComplaintHPIEyeID:     &chief.IDChiefComplaintHPIEye,
		SecondaryComplaintHPIEyeID: &secondary.IDSecondaryComplaintHPIEye,
		ChiefComplaintNote:         input.ChiefComplaintNote,
	}
	if err := s.db.Create(&ccHpi).Error; err != nil {
		return nil, err
	}

	activitylog.Log(s.db, "exam_cc_hpi", "save",
		activitylog.WithEntity(examID),
	)

	return map[string]interface{}{
		"id_cc_hpi_eye":                  ccHpi.IDCcHpiEye,
		"eye_exam_id":                    ccHpi.EyeExamID,
		"chief_complaint_hpi_eye_id":     ccHpi.ChiefComplaintHPIEyeID,
		"secondary_complaint_hpi_eye_id": ccHpi.SecondaryComplaintHPIEyeID,
		"chief_complaint_note":           ccHpi.ChiefComplaintNote,
		"chief_complaint_hpi_eye":        chief.ToMap(),
		"secondary_complaint_hpi_eye":    secondary.ToMap(),
	}, nil
}

// ─── GetCcHpi ─────────────────────────────────────────────────────────────────

func (s *Service) GetCcHpi(examID int64) (map[string]interface{}, error) {
	if _, err := s.getExam(examID); err != nil {
		return nil, err
	}

	var ccHpi visionModel.CcHpiEye
	if err := s.db.Where("eye_exam_id = ?", examID).First(&ccHpi).Error; err != nil {
		return map[string]interface{}{
			"exam_id":                      examID,
			"exists":                       false,
			"cc_hpi_eye":                   nil,
			"chief_complaint_hpi_eye":      nil,
			"secondary_complaint_hpi_eye":  nil,
			"chief_complaint_note":         nil,
		}, nil
	}

	var chief *visionModel.ChiefComplaintHPIEye
	if ccHpi.ChiefComplaintHPIEyeID != nil {
		var c visionModel.ChiefComplaintHPIEye
		if err := s.db.First(&c, *ccHpi.ChiefComplaintHPIEyeID).Error; err == nil {
			chief = &c
		}
	}

	var secondary *visionModel.SecondaryComplaintHPIEye
	if ccHpi.SecondaryComplaintHPIEyeID != nil {
		var sc visionModel.SecondaryComplaintHPIEye
		if err := s.db.First(&sc, *ccHpi.SecondaryComplaintHPIEyeID).Error; err == nil {
			secondary = &sc
		}
	}

	var chiefMap, secondaryMap interface{}
	if chief != nil {
		chiefMap = chief.ToMap()
	}
	if secondary != nil {
		secondaryMap = secondary.ToMap()
	}

	return map[string]interface{}{
		"exam_id":                     examID,
		"exists":                      true,
		"cc_hpi_eye":                  ccHpi.ToMap(),
		"chief_complaint_hpi_eye":     chiefMap,
		"secondary_complaint_hpi_eye": secondaryMap,
		"chief_complaint_note":        ccHpi.ChiefComplaintNote,
	}, nil
}

// ─── UpdateCcHpi ──────────────────────────────────────────────────────────────

func (s *Service) UpdateCcHpi(username string, examID int64, input UpdateCcHpiInput) (map[string]interface{}, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}

	exam, err := s.getExam(examID)
	if err != nil {
		return nil, err
	}

	if exam.Passed {
		return nil, errors.New("cannot update cc_hpi for a completed exam")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return nil, errors.New("you are not authorized to update this exam")
	}

	var ccHpi visionModel.CcHpiEye
	if err := s.db.Where("eye_exam_id = ?", examID).First(&ccHpi).Error; err != nil {
		return nil, errors.New("cc_hpi record not found for this exam")
	}

	// Update chief_complaint_note on CcHpiEye
	if input.ChiefComplaintNote != nil {
		ccHpi.ChiefComplaintNote = input.ChiefComplaintNote
		s.db.Save(&ccHpi)
	}

	var chief *visionModel.ChiefComplaintHPIEye
	if input.ChiefComplaintHpiEye != nil && ccHpi.ChiefComplaintHPIEyeID != nil {
		var c visionModel.ChiefComplaintHPIEye
		if err := s.db.First(&c, *ccHpi.ChiefComplaintHPIEyeID).Error; err == nil {
			d := input.ChiefComplaintHpiEye
			if d.NoteChiefComplaint != nil {
				c.NoteChiefComplaint = d.NoteChiefComplaint
			}
			if d.Location != nil {
				c.Location = d.Location
			}
			if d.Quality != nil {
				c.Quality = d.Quality
			}
			if d.Severity != nil {
				c.Severity = d.Severity
			}
			if d.Duration != nil {
				c.Duration = d.Duration
			}
			if d.Timing != nil {
				c.Timing = d.Timing
			}
			if d.Context != nil {
				c.Context = d.Context
			}
			if d.Factors != nil {
				c.Factors = d.Factors
			}
			if d.Symptoms != nil {
				c.Symptoms = d.Symptoms
			}
			s.db.Save(&c)
			chief = &c
		}
	} else if ccHpi.ChiefComplaintHPIEyeID != nil {
		var c visionModel.ChiefComplaintHPIEye
		if err := s.db.First(&c, *ccHpi.ChiefComplaintHPIEyeID).Error; err == nil {
			chief = &c
		}
	}

	var secondary *visionModel.SecondaryComplaintHPIEye
	if input.SecondaryComplaintHpiEye != nil && ccHpi.SecondaryComplaintHPIEyeID != nil {
		var sc visionModel.SecondaryComplaintHPIEye
		if err := s.db.First(&sc, *ccHpi.SecondaryComplaintHPIEyeID).Error; err == nil {
			sd := input.SecondaryComplaintHpiEye
			if sd.NoteSecondaryComplaint != nil {
				sc.NoteSecondaryComplaint = sd.NoteSecondaryComplaint
			}
			if sd.Location != nil {
				sc.Location = sd.Location
			}
			if sd.Quality != nil {
				sc.Quality = sd.Quality
			}
			if sd.Severity != nil {
				sc.Severity = sd.Severity
			}
			if sd.Duration != nil {
				sc.Duration = sd.Duration
			}
			if sd.Timing != nil {
				sc.Timing = sd.Timing
			}
			if sd.Context != nil {
				sc.Context = sd.Context
			}
			if sd.Factors != nil {
				sc.Factors = sd.Factors
			}
			if sd.Symptoms != nil {
				sc.Symptoms = sd.Symptoms
			}
			s.db.Save(&sc)
			secondary = &sc
		}
	} else if ccHpi.SecondaryComplaintHPIEyeID != nil {
		var sc visionModel.SecondaryComplaintHPIEye
		if err := s.db.First(&sc, *ccHpi.SecondaryComplaintHPIEyeID).Error; err == nil {
			secondary = &sc
		}
	}

	activitylog.Log(s.db, "exam_cc_hpi", "update",
		activitylog.WithEntity(examID),
	)

	var chiefMap, secondaryMap interface{}
	if chief != nil {
		chiefMap = chief.ToMap()
	}
	if secondary != nil {
		secondaryMap = secondary.ToMap()
	}

	return map[string]interface{}{
		"cc_hpi_eye":                  ccHpi.ToMap(),
		"chief_complaint_hpi_eye":     chiefMap,
		"secondary_complaint_hpi_eye": secondaryMap,
	}, nil
}
