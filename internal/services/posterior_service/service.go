package posterior_service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"

	empLoginModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	posteriorModel "sighthub-backend/internal/models/medical/vision_exam/posterior"
	visionModel "sighthub-backend/internal/models/vision_exam"
	"sighthub-backend/pkg/activitylog"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ---------- input types ----------

type FindingsPosteriorInput struct {
	OdView             *string `json:"od_view"`
	OsView             *string `json:"os_view"`
	OdVitreous         *string `json:"od_vitreous"`
	OsVitreous         *string `json:"os_vitreous"`
	OdMacula           *string `json:"od_macula"`
	OsMacula           *string `json:"os_macula"`
	OdBackground       *string `json:"od_background"`
	OsBackground       *string `json:"os_background"`
	OdVessels          *string `json:"od_vessels"`
	OsVessels          *string `json:"os_vessels"`
	OdPeripheralFundus *string `json:"od_peripheral_fundus"`
	OsPeripheralFundus *string `json:"os_peripheral_fundus"`
	OdOpticsNerve      *string `json:"od_optics_nerve"`
	OsOpticsNerve      *string `json:"os_optics_nerve"`
}

type CupDiscRatioInput struct {
	OdV *string `json:"od_v"`
	OsV *string `json:"os_v"`
	OdH *string `json:"od_h"`
	OsH *string `json:"os_h"`
}

type PosteriorEyeInput struct {
	InfoDirect                *bool   `json:"info_direct"`
	InfoBio                   *bool   `json:"info_bio"`
	Info90d                   *bool   `json:"info_90d"`
	InfoOptomap               *bool   `json:"info_optomap"`
	InfoRha                   *bool   `json:"info_rha"`
	InfoOther                 *string `json:"info_other"`
	MedicationPatientEducated *bool   `json:"medication_patient_educated"`
	MedicationIlationDeclined *bool   `json:"medication_ilation_declined"`
	MedicationParemyd         *bool   `json:"medication_paremyd"`
	MedicationAtropine        *bool   `json:"medication_atropine"`
	MedicationTropicamide     *bool   `json:"medication_tropicamide"`
	MedicationCyclopentolate  *bool   `json:"medication_cyclopentolate"`
	MedicationHomatropine     *bool   `json:"medication_homatropine"`
	MedicationPhenylephrine   *bool   `json:"medication_phenylephrine"`
	MedicationRha             *bool   `json:"medication_rha"`
	TimeDilated               *string `json:"time_dilated"`
	Other                     *string `json:"other"`
	Note                      *string `json:"note"`
	AddDrawing                *string `json:"add_drawing"`
}

type SavePosteriorInput struct {
	PosteriorEye      *PosteriorEyeInput      `json:"posterior_eye"`
	FindingsPosterior *FindingsPosteriorInput  `json:"findings_posterior"`
	CupDiscRatio      *CupDiscRatioInput       `json:"cup_disc_ratio_posterior"`
}

type UpdatePosteriorInput struct {
	PosteriorEye      *PosteriorEyeInput      `json:"posterior_eye"`
	FindingsPosterior *FindingsPosteriorInput  `json:"findings_posterior"`
	CupDiscRatio      *CupDiscRatioInput       `json:"cup_disc_ratio_posterior"`
}

type PosteriorResult struct {
	ExamID           int64                                  `json:"exam_id"`
	Exists           bool                                   `json:"exists"`
	PosteriorEye     *posteriorModel.PosteriorEye           `json:"posterior_eye"`
	FindingsPosterior *posteriorModel.FindingsPosterior     `json:"findings_posterior"`
	CupDiscRatio      *posteriorModel.CupDiscRatioPosterior `json:"cup_disc_ratio_posterior"`
}

// ---------- helpers ----------

var timeHHMM = regexp.MustCompile(`^\d{2}:\d{2}$`)

func boolDefault(p *bool, def bool) *bool {
	if p != nil {
		return p
	}
	b := def
	return &b
}

func parseTimeDilated(s *string) (*string, error) {
	if s == nil {
		return nil, nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil, nil
	}
	if !timeHHMM.MatchString(v) {
		return nil, fmt.Errorf("invalid time_dilated format %q, expected HH:MM", v)
	}
	return &v, nil
}

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

// cleanPosteriorEye zeros out GORM relation fields so they don't appear in JSON.
func cleanPosteriorEye(pe *posteriorModel.PosteriorEye) {
	pe.Findings = posteriorModel.FindingsPosterior{}
	pe.CupDiscRatio = posteriorModel.CupDiscRatioPosterior{}
}

// ---------- SavePosterior ----------

func (s *Service) SavePosterior(username string, examID int64, input SavePosteriorInput) (*PosteriorResult, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}

	var exam visionModel.EyeExam
	if err := s.db.Where("id_eye_exam = ?", examID).First(&exam).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return nil, errors.New("not authorized")
	}
	if exam.Passed {
		return nil, errors.New("cannot update: exam is completed")
	}

	var count int64
	s.db.Model(&posteriorModel.PosteriorEye{}).Where("eye_exam_id = ?", examID).Count(&count)
	if count > 0 {
		return nil, errors.New("posterior record already exists")
	}

	// Build sub-records from input (all fields default to nil if section absent)
	var fp posteriorModel.FindingsPosterior
	if fi := input.FindingsPosterior; fi != nil {
		fp = posteriorModel.FindingsPosterior{
			OdView: fi.OdView, OsView: fi.OsView,
			OdVitreous: fi.OdVitreous, OsVitreous: fi.OsVitreous,
			OdMacula: fi.OdMacula, OsMacula: fi.OsMacula,
			OdBackground: fi.OdBackground, OsBackground: fi.OsBackground,
			OdVessels: fi.OdVessels, OsVessels: fi.OsVessels,
			OdPeripheralFundus: fi.OdPeripheralFundus, OsPeripheralFundus: fi.OsPeripheralFundus,
			OdOpticsNerve: fi.OdOpticsNerve, OsOpticsNerve: fi.OsOpticsNerve,
		}
	}

	var cdr posteriorModel.CupDiscRatioPosterior
	if ci := input.CupDiscRatio; ci != nil {
		cdr = posteriorModel.CupDiscRatioPosterior{OdV: ci.OdV, OsV: ci.OsV, OdH: ci.OdH, OsH: ci.OsH}
	}

	var pei PosteriorEyeInput
	if input.PosteriorEye != nil {
		pei = *input.PosteriorEye
	}

	timeDilated, err := parseTimeDilated(pei.TimeDilated)
	if err != nil {
		return nil, err
	}

	var addDrawing *string
	if pei.AddDrawing != nil {
		v := strings.TrimSpace(*pei.AddDrawing)
		if v != "" {
			addDrawing = &v
		}
	}

	var pe posteriorModel.PosteriorEye
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&fp).Error; err != nil {
			return err
		}
		if err := tx.Create(&cdr).Error; err != nil {
			return err
		}

		pe = posteriorModel.PosteriorEye{
			InfoDirect:                boolDefault(pei.InfoDirect, false),
			InfoBio:                   boolDefault(pei.InfoBio, false),
			Info90d:                   boolDefault(pei.Info90d, false),
			InfoOptomap:               boolDefault(pei.InfoOptomap, false),
			InfoRha:                   boolDefault(pei.InfoRha, false),
			InfoOther:                 pei.InfoOther,
			MedicationPatientEducated: boolDefault(pei.MedicationPatientEducated, false),
			MedicationIlationDeclined: boolDefault(pei.MedicationIlationDeclined, false),
			MedicationParemyd:         boolDefault(pei.MedicationParemyd, false),
			MedicationAtropine:        boolDefault(pei.MedicationAtropine, false),
			MedicationTropicamide:     boolDefault(pei.MedicationTropicamide, false),
			MedicationCyclopentolate:  boolDefault(pei.MedicationCyclopentolate, false),
			MedicationHomatropine:     boolDefault(pei.MedicationHomatropine, false),
			MedicationPhenylephrine:   boolDefault(pei.MedicationPhenylephrine, false),
			MedicationRha:             boolDefault(pei.MedicationRha, false),
			TimeDilated:               timeDilated,
			Other:                     pei.Other,
			Note:                      pei.Note,
			AddDrawing:                addDrawing,
			FindingsPosteriorID:       fp.IDFindingsPosterior,
			CupDiscRatioPosteriorID:   cdr.IDCupDiscRatioPosterior,
			EyeExamID:                 examID,
		}
		if err := tx.Create(&pe).Error; err != nil {
			return err
		}
		activitylog.Log(tx, "exam_posterior", "save", activitylog.WithEntity(examID))
		return nil
	})
	if err != nil {
		return nil, err
	}

	cleanPosteriorEye(&pe)
	return &PosteriorResult{
		ExamID: examID, Exists: true,
		PosteriorEye:      &pe,
		FindingsPosterior: &fp,
		CupDiscRatio:      &cdr,
	}, nil
}

// ---------- GetPosterior ----------

func (s *Service) GetPosterior(examID int64) (*PosteriorResult, error) {
	var pe posteriorModel.PosteriorEye
	if err := s.db.Where("eye_exam_id = ?", examID).First(&pe).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &PosteriorResult{ExamID: examID, Exists: false}, nil
		}
		return nil, err
	}

	var fp posteriorModel.FindingsPosterior
	s.db.First(&fp, pe.FindingsPosteriorID)

	var cdr posteriorModel.CupDiscRatioPosterior
	s.db.First(&cdr, pe.CupDiscRatioPosteriorID)

	cleanPosteriorEye(&pe)
	return &PosteriorResult{
		ExamID: examID, Exists: true,
		PosteriorEye:      &pe,
		FindingsPosterior: &fp,
		CupDiscRatio:      &cdr,
	}, nil
}

// ---------- UpdatePosterior ----------

func (s *Service) UpdatePosterior(username string, examID int64, input UpdatePosteriorInput) (*PosteriorResult, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}

	var exam visionModel.EyeExam
	if err := s.db.Where("id_eye_exam = ?", examID).First(&exam).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return nil, errors.New("not authorized")
	}
	if exam.Passed {
		return nil, errors.New("cannot update: exam is completed")
	}

	var pe posteriorModel.PosteriorEye
	if err := s.db.Where("eye_exam_id = ?", examID).First(&pe).Error; err != nil {
		return nil, errors.New("posterior_eye not found")
	}

	var fp posteriorModel.FindingsPosterior
	if err := s.db.First(&fp, pe.FindingsPosteriorID).Error; err != nil {
		return nil, errors.New("findings_posterior not found")
	}

	var cdr posteriorModel.CupDiscRatioPosterior
	if err := s.db.First(&cdr, pe.CupDiscRatioPosteriorID).Error; err != nil {
		return nil, errors.New("cup_disc_ratio_posterior not found")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Update PosteriorEye fields
		if pei := input.PosteriorEye; pei != nil {
			if pei.InfoDirect != nil {
				pe.InfoDirect = pei.InfoDirect
			}
			if pei.InfoBio != nil {
				pe.InfoBio = pei.InfoBio
			}
			if pei.Info90d != nil {
				pe.Info90d = pei.Info90d
			}
			if pei.InfoOptomap != nil {
				pe.InfoOptomap = pei.InfoOptomap
			}
			if pei.InfoRha != nil {
				pe.InfoRha = pei.InfoRha
			}
			if pei.InfoOther != nil {
				pe.InfoOther = pei.InfoOther
			}
			if pei.MedicationPatientEducated != nil {
				pe.MedicationPatientEducated = pei.MedicationPatientEducated
			}
			if pei.MedicationIlationDeclined != nil {
				pe.MedicationIlationDeclined = pei.MedicationIlationDeclined
			}
			if pei.MedicationParemyd != nil {
				pe.MedicationParemyd = pei.MedicationParemyd
			}
			if pei.MedicationAtropine != nil {
				pe.MedicationAtropine = pei.MedicationAtropine
			}
			if pei.MedicationTropicamide != nil {
				pe.MedicationTropicamide = pei.MedicationTropicamide
			}
			if pei.MedicationCyclopentolate != nil {
				pe.MedicationCyclopentolate = pei.MedicationCyclopentolate
			}
			if pei.MedicationHomatropine != nil {
				pe.MedicationHomatropine = pei.MedicationHomatropine
			}
			if pei.MedicationPhenylephrine != nil {
				pe.MedicationPhenylephrine = pei.MedicationPhenylephrine
			}
			if pei.MedicationRha != nil {
				pe.MedicationRha = pei.MedicationRha
			}
			if pei.TimeDilated != nil {
				td, err := parseTimeDilated(pei.TimeDilated)
				if err != nil {
					return err
				}
				pe.TimeDilated = td
			}
			if pei.Other != nil {
				pe.Other = pei.Other
			}
			if pei.Note != nil {
				pe.Note = pei.Note
			}
			if pei.AddDrawing != nil {
				v := strings.TrimSpace(*pei.AddDrawing)
				if v == "" {
					pe.AddDrawing = nil
				} else {
					pe.AddDrawing = &v
				}
			}
			if err := tx.Save(&pe).Error; err != nil {
				return err
			}
		}

		// Update FindingsPosterior
		if fi := input.FindingsPosterior; fi != nil {
			if fi.OdView != nil {
				fp.OdView = fi.OdView
			}
			if fi.OsView != nil {
				fp.OsView = fi.OsView
			}
			if fi.OdVitreous != nil {
				fp.OdVitreous = fi.OdVitreous
			}
			if fi.OsVitreous != nil {
				fp.OsVitreous = fi.OsVitreous
			}
			if fi.OdMacula != nil {
				fp.OdMacula = fi.OdMacula
			}
			if fi.OsMacula != nil {
				fp.OsMacula = fi.OsMacula
			}
			if fi.OdBackground != nil {
				fp.OdBackground = fi.OdBackground
			}
			if fi.OsBackground != nil {
				fp.OsBackground = fi.OsBackground
			}
			if fi.OdVessels != nil {
				fp.OdVessels = fi.OdVessels
			}
			if fi.OsVessels != nil {
				fp.OsVessels = fi.OsVessels
			}
			if fi.OdPeripheralFundus != nil {
				fp.OdPeripheralFundus = fi.OdPeripheralFundus
			}
			if fi.OsPeripheralFundus != nil {
				fp.OsPeripheralFundus = fi.OsPeripheralFundus
			}
			if fi.OdOpticsNerve != nil {
				fp.OdOpticsNerve = fi.OdOpticsNerve
			}
			if fi.OsOpticsNerve != nil {
				fp.OsOpticsNerve = fi.OsOpticsNerve
			}
			if err := tx.Save(&fp).Error; err != nil {
				return err
			}
		}

		// Update CupDiscRatioPosterior
		if ci := input.CupDiscRatio; ci != nil {
			if ci.OdV != nil {
				cdr.OdV = ci.OdV
			}
			if ci.OsV != nil {
				cdr.OsV = ci.OsV
			}
			if ci.OdH != nil {
				cdr.OdH = ci.OdH
			}
			if ci.OsH != nil {
				cdr.OsH = ci.OsH
			}
			if err := tx.Save(&cdr).Error; err != nil {
				return err
			}
		}

		activitylog.Log(tx, "exam_posterior", "update", activitylog.WithEntity(examID))
		return nil
	})
	if err != nil {
		return nil, err
	}

	cleanPosteriorEye(&pe)
	return &PosteriorResult{
		ExamID: examID,
		PosteriorEye:      &pe,
		FindingsPosterior: &fp,
		CupDiscRatio:      &cdr,
	}, nil
}
