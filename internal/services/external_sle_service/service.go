package external_sle_service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	empLoginModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	sleModel "sighthub-backend/internal/models/medical/vision_exam/external_sle"
	visionModel "sighthub-backend/internal/models/vision_exam"
	"sighthub-backend/pkg/activitylog"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ValidationError carries field-level validation errors (HTTP 400).
type ValidationError struct{ Errors []string }

func (e *ValidationError) Error() string { return "validation error" }

// ---------- input types ----------

type FindingsInput struct {
	Externals           *string `json:"externals"`
	OdLidsLashes        *string `json:"od_lids_lashes"`
	OsLidsLashes        *string `json:"os_lids_lashes"`
	OdConjunctivaSclera *string `json:"od_conjunctiva_sclera"`
	OsConjunctivaSclera *string `json:"os_conjunctiva_sclera"`
	OdCornea            *string `json:"od_cornea"`
	OsCornea            *string `json:"os_cornea"`
	OdTearFilm          *string `json:"od_tear_film"`
	OsTearFilm          *string `json:"os_tear_film"`
	OdAnteriorChamber   *string `json:"od_anterior_chamber"`
	OsAnteriorChamber   *string `json:"os_anterior_chamber"`
	OdIris              *string `json:"od_iris"`
	OsIris              *string `json:"os_iris"`
	OdLens              *string `json:"od_lens"`
	OsLens              *string `json:"os_lens"`
}

type GonioscopyInput struct {
	OdSup   *string `json:"od_sup"`
	OsSup   *string `json:"os_sup"`
	OdInf   *string `json:"od_inf"`
	OsInf   *string `json:"os_inf"`
	OdNasal *string `json:"od_nasal"`
	OsNasal *string `json:"os_nasal"`
	OdTemp  *string `json:"od_temp"`
	OsTemp  *string `json:"os_temp"`
}

type PachInput struct {
	Od *string `json:"od"`
	Os *string `json:"os"`
}

type VisualFieldsInput struct {
	SuperoTemporoOd  *string `json:"supero_temporo_od"`
	SuperoTemporoOs  *string `json:"supero_temporo_os"`
	SuperoNasalOd    *string `json:"supero_nasal_od"`
	SuperoNasalOs    *string `json:"supero_nasal_os"`
	InferoTemporalOd *string `json:"infero_temporal_od"`
	InferoTemporalOs *string `json:"infero_temporal_os"`
	InferoNasalOd    *string `json:"infero_nasal_od"`
	InferoNasalOs    *string `json:"infero_nasal_os"`
	Instrument       *string `json:"instrument"`
	Test             *string `json:"test"`
	Reason           *string `json:"reason"`
	Result           *string `json:"result"`
	Recommendations  *string `json:"recommendations"`
	Comments         *string `json:"comments"`
}

type TonometryInput struct {
	MethodTonometry  string  `json:"method_tonometry"`
	DateTonometryEye *string `json:"date_tonometry_eye"`
	TimeTonometryEye *string `json:"time_tonometry_eye"`
	OdTonometryEye   *string `json:"od_tonometry_eye"`
	OsTonometryEye   *string `json:"os_tonometry_eye"`
}

type SaveExternalSleInput struct {
	FindingsExternalSle   FindingsInput     `json:"findings_external_sle"`
	GonioscopyExternalSle GonioscopyInput   `json:"gonioscopy_external_sle"`
	PachExternalSle       PachInput         `json:"pach_external_sle"`
	VisualFields          VisualFieldsInput `json:"visual_fields"`
	TonometryEyes         *TonometryInput   `json:"tonometry_eyes"`
	AddDrawing            string            `json:"add_drawing"`
	OdAngleEstimation     *string           `json:"od_angle_estimation"`
	OsAngleEstimation     *string           `json:"os_angle_estimation"`
	IopDropsFluress       *bool             `json:"iop_drops_fluress"`
	IopDropsProparacaine  *bool             `json:"iop_drops_proparacaine"`
	IopDropsFluoroStrip   *bool             `json:"iop_drops_fluoro_strip"`
	Note                  *string           `json:"note"`
}

type UpdateExternalSleInput struct {
	FindingsExternalSle   *FindingsInput     `json:"findings_external_sle"`
	GonioscopyExternalSle *GonioscopyInput   `json:"gonioscopy_external_sle"`
	PachExternalSle       *PachInput         `json:"pach_external_sle"`
	VisualFields          *VisualFieldsInput `json:"visual_fields"`
	TonometryEyes         *TonometryInput    `json:"tonometry_eyes"`
	AddDrawing            string             `json:"add_drawing"`
	OdAngleEstimation     *string            `json:"od_angle_estimation"`
	OsAngleEstimation     *string            `json:"os_angle_estimation"`
	IopDropsFluress       *bool              `json:"iop_drops_fluress"`
	IopDropsProparacaine  *bool              `json:"iop_drops_proparacaine"`
	IopDropsFluoroStrip   *bool              `json:"iop_drops_fluoro_strip"`
	Note                  *string            `json:"note"`
}

// ---------- helpers ----------

func boolPtr(b bool) *bool { return &b }

func cleanDrawingPath(path string) string {
	const prefix = "/mnt/tank/data/"
	if strings.Contains(path, prefix) {
		return strings.TrimSpace(strings.SplitN(path, prefix, 2)[1])
	}
	return path
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

func parseTonometry(sleID int64, in *TonometryInput) (*sleModel.TonometryEye, []string) {
	var errs []string
	if in.MethodTonometry == "" {
		errs = append(errs, "Missing 'method_tonometry'")
	} else if len(in.MethodTonometry) > 100 {
		errs = append(errs, "Invalid 'method_tonometry': must be a string with max length 100.")
	}
	var dateParsed *time.Time
	if in.DateTonometryEye != nil && *in.DateTonometryEye != "" {
		t, err := time.Parse("2006-01-02", *in.DateTonometryEye)
		if err != nil {
			errs = append(errs, "Invalid 'date_tonometry_eye': must be in 'YYYY-MM-DD' format.")
		} else {
			dateParsed = &t
		}
	}
	var timeStr *string
	if in.TimeTonometryEye != nil && *in.TimeTonometryEye != "" {
		if _, err := time.Parse("15:04", *in.TimeTonometryEye); err != nil {
			errs = append(errs, "Invalid 'time_tonometry_eye': must be in 'HH:MM' format.")
		} else {
			timeStr = in.TimeTonometryEye
		}
	}
	for _, f := range []struct {
		v    *string
		name string
	}{
		{in.OdTonometryEye, "od_tonometry_eye"},
		{in.OsTonometryEye, "os_tonometry_eye"},
	} {
		if f.v != nil && *f.v != "" {
			n, err := strconv.Atoi(*f.v)
			if err != nil || n < 0 || n > 999 {
				errs = append(errs, "Invalid '"+f.name+"': must be a number between 0 and 999.")
			}
		}
	}
	if len(errs) > 0 {
		return nil, errs
	}
	m := in.MethodTonometry
	return &sleModel.TonometryEye{
		ExternalSleEyeID: sleID,
		MethodTonometry:  &m,
		DateTonometryEye: dateParsed,
		TimeTonometryEye: timeStr,
		OdTonometryEye:   in.OdTonometryEye,
		OsTonometryEye:   in.OsTonometryEye,
	}, nil
}

// ---------- SaveExternalSle ----------

func (s *Service) SaveExternalSle(username string, examID int64, input SaveExternalSleInput) (map[string]interface{}, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}
	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	if exam.Passed {
		return nil, errors.New("cannot update a completed exam")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return nil, errors.New("not authorized")
	}

	fi := input.FindingsExternalSle
	findings := sleModel.FindingsExternalSle{
		Externals: fi.Externals, OdLidsLashes: fi.OdLidsLashes, OsLidsLashes: fi.OsLidsLashes,
		OdConjunctivaSclera: fi.OdConjunctivaSclera, OsConjunctivaSclera: fi.OsConjunctivaSclera,
		OdCornea: fi.OdCornea, OsCornea: fi.OsCornea,
		OdTearFilm: fi.OdTearFilm, OsTearFilm: fi.OsTearFilm,
		OdAnteriorChamber: fi.OdAnteriorChamber, OsAnteriorChamber: fi.OsAnteriorChamber,
		OdIris: fi.OdIris, OsIris: fi.OsIris,
		OdLens: fi.OdLens, OsLens: fi.OsLens,
	}
	if err := s.db.Create(&findings).Error; err != nil {
		return nil, err
	}

	gi := input.GonioscopyExternalSle
	gonioscopy := sleModel.GonioscopyExternalSle{
		OdSup: gi.OdSup, OsSup: gi.OsSup,
		OdInf: gi.OdInf, OsInf: gi.OsInf,
		OdNasal: gi.OdNasal, OsNasal: gi.OsNasal,
		OdTemp: gi.OdTemp, OsTemp: gi.OsTemp,
	}
	if err := s.db.Create(&gonioscopy).Error; err != nil {
		return nil, err
	}

	pi := input.PachExternalSle
	pach := sleModel.PachExternalSle{Od: pi.Od, Os: pi.Os}
	if err := s.db.Create(&pach).Error; err != nil {
		return nil, err
	}

	vfi := input.VisualFields
	vf := sleModel.VisualFields{
		SuperoTemporoOd: vfi.SuperoTemporoOd, SuperoTemporoOs: vfi.SuperoTemporoOs,
		SuperoNasalOd: vfi.SuperoNasalOd, SuperoNasalOs: vfi.SuperoNasalOs,
		InferoTemporalOd: vfi.InferoTemporalOd, InferoTemporalOs: vfi.InferoTemporalOs,
		InferoNasalOd: vfi.InferoNasalOd, InferoNasalOs: vfi.InferoNasalOs,
		Instrument: vfi.Instrument, Test: vfi.Test,
		Reason: vfi.Reason, Result: vfi.Result,
		Recommendations: vfi.Recommendations, Comments: vfi.Comments,
	}
	if err := s.db.Create(&vf).Error; err != nil {
		return nil, err
	}

	cleanPath := cleanDrawingPath(input.AddDrawing)
	var addDrawing *string
	if cleanPath != "" {
		addDrawing = &cleanPath
	}
	odAngle := "n/a"
	if input.OdAngleEstimation != nil {
		odAngle = *input.OdAngleEstimation
	}
	osAngle := "n/a"
	if input.OsAngleEstimation != nil {
		osAngle = *input.OsAngleEstimation
	}
	iopFluress := boolPtr(false)
	if input.IopDropsFluress != nil {
		iopFluress = input.IopDropsFluress
	}
	iopPropara := boolPtr(false)
	if input.IopDropsProparacaine != nil {
		iopPropara = input.IopDropsProparacaine
	}
	iopFluoro := boolPtr(false)
	if input.IopDropsFluoroStrip != nil {
		iopFluoro = input.IopDropsFluoroStrip
	}
	vfID := vf.IDVisualFields
	sle := sleModel.ExternalSleEye{
		EyeExamID:               examID,
		FindingsExternalSleID:   findings.IDFindingsExternalSle,
		GonioscopyExternalSleID: gonioscopy.IDGonioscopyExternalSle,
		PachExternalSleID:       pach.IDPachExternalSle,
		VisualFieldsID:          &vfID,
		AddDrawing:              addDrawing,
		OdAngleEstimation:       odAngle,
		OsAngleEstimation:       osAngle,
		IopDropsFluress:         iopFluress,
		IopDropsProparacaine:    iopPropara,
		IopDropsFluoroStrip:     iopFluoro,
		Note:                    input.Note,
	}
	if err := s.db.Create(&sle).Error; err != nil {
		return nil, err
	}

	if input.TonometryEyes != nil {
		t, valErrs := parseTonometry(sle.IDExternalSleEye, input.TonometryEyes)
		if len(valErrs) > 0 {
			return nil, &ValidationError{Errors: valErrs}
		}
		if err := s.db.Create(t).Error; err != nil {
			return nil, err
		}
	}

	activitylog.Log(s.db, "exam_sle", "save", activitylog.WithEntity(examID))
	return map[string]interface{}{
		"message":         "External SLE and related data saved successfully",
		"external_sle_id": sle.IDExternalSleEye,
	}, nil
}

// ---------- GetExternalSle ----------

func (s *Service) GetExternalSle(examID int64) (map[string]interface{}, error) {
	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return nil, errors.New("exam not found")
	}

	var patientExamIDs []int64
	s.db.Model(&visionModel.EyeExam{}).Where("patient_id = ?", exam.PatientID).Pluck("id_eye_exam", &patientExamIDs)

	var sleIDs []int64
	s.db.Model(&sleModel.ExternalSleEye{}).Where("eye_exam_id IN ?", patientExamIDs).Pluck("id_external_sle_eye", &sleIDs)

	var allTonometry []sleModel.TonometryEye
	s.db.Where("external_sle_eye_id IN ?", sleIDs).
		Order("date_tonometry_eye DESC, time_tonometry_eye DESC").
		Find(&allTonometry)

	var sle sleModel.ExternalSleEye
	if err := s.db.Where("eye_exam_id = ?", examID).First(&sle).Error; err != nil {
		return map[string]interface{}{
			"exam_id":   examID,
			"exists":    false,
			"tonometry": allTonometry,
		}, nil
	}

	var findings sleModel.FindingsExternalSle
	s.db.First(&findings, sle.FindingsExternalSleID)

	var gonioscopy sleModel.GonioscopyExternalSle
	s.db.First(&gonioscopy, sle.GonioscopyExternalSleID)

	var pach sleModel.PachExternalSle
	s.db.First(&pach, sle.PachExternalSleID)

	var vfPtr *sleModel.VisualFields
	if sle.VisualFieldsID != nil {
		var vf sleModel.VisualFields
		if s.db.First(&vf, *sle.VisualFieldsID).Error == nil {
			vfPtr = &vf
		}
	}

	return map[string]interface{}{
		"exam_id":                 examID,
		"exists":                  true,
		"od_angle_estimation":     sle.OdAngleEstimation,
		"os_angle_estimation":     sle.OsAngleEstimation,
		"add_drawing":             sle.AddDrawing,
		"iop_drops_fluress":       sle.IopDropsFluress,
		"iop_drops_proparacaine":  sle.IopDropsProparacaine,
		"iop_drops_fluoro_strip":  sle.IopDropsFluoroStrip,
		"note":                    sle.Note,
		"findings_external_sle":   findings,
		"gonioscopy_external_sle": gonioscopy,
		"pach_external_sle":       pach,
		"visual_fields":           vfPtr,
		"tonometry":               allTonometry,
	}, nil
}

// ---------- UpdateExternalSle ----------

func (s *Service) UpdateExternalSle(username string, examID int64, input UpdateExternalSleInput) (map[string]interface{}, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}
	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return nil, errors.New("not authorized")
	}
	if exam.Passed {
		return nil, errors.New("cannot update a completed exam")
	}

	var sle sleModel.ExternalSleEye
	if err := s.db.Where("eye_exam_id = ?", examID).First(&sle).Error; err != nil {
		return nil, errors.New("external_sle not found")
	}

	if input.FindingsExternalSle != nil {
		var f sleModel.FindingsExternalSle
		if s.db.First(&f, sle.FindingsExternalSleID).Error == nil {
			in := input.FindingsExternalSle
			if in.Externals != nil           { f.Externals = in.Externals }
			if in.OdLidsLashes != nil        { f.OdLidsLashes = in.OdLidsLashes }
			if in.OsLidsLashes != nil        { f.OsLidsLashes = in.OsLidsLashes }
			if in.OdConjunctivaSclera != nil { f.OdConjunctivaSclera = in.OdConjunctivaSclera }
			if in.OsConjunctivaSclera != nil { f.OsConjunctivaSclera = in.OsConjunctivaSclera }
			if in.OdCornea != nil            { f.OdCornea = in.OdCornea }
			if in.OsCornea != nil            { f.OsCornea = in.OsCornea }
			if in.OdTearFilm != nil          { f.OdTearFilm = in.OdTearFilm }
			if in.OsTearFilm != nil          { f.OsTearFilm = in.OsTearFilm }
			if in.OdAnteriorChamber != nil   { f.OdAnteriorChamber = in.OdAnteriorChamber }
			if in.OsAnteriorChamber != nil   { f.OsAnteriorChamber = in.OsAnteriorChamber }
			if in.OdIris != nil              { f.OdIris = in.OdIris }
			if in.OsIris != nil              { f.OsIris = in.OsIris }
			if in.OdLens != nil              { f.OdLens = in.OdLens }
			if in.OsLens != nil              { f.OsLens = in.OsLens }
			s.db.Save(&f)
		}
	}

	if input.GonioscopyExternalSle != nil {
		var g sleModel.GonioscopyExternalSle
		if s.db.First(&g, sle.GonioscopyExternalSleID).Error == nil {
			in := input.GonioscopyExternalSle
			if in.OdSup != nil   { g.OdSup = in.OdSup }
			if in.OsSup != nil   { g.OsSup = in.OsSup }
			if in.OdInf != nil   { g.OdInf = in.OdInf }
			if in.OsInf != nil   { g.OsInf = in.OsInf }
			if in.OdNasal != nil { g.OdNasal = in.OdNasal }
			if in.OsNasal != nil { g.OsNasal = in.OsNasal }
			if in.OdTemp != nil  { g.OdTemp = in.OdTemp }
			if in.OsTemp != nil  { g.OsTemp = in.OsTemp }
			s.db.Save(&g)
		}
	}

	if input.PachExternalSle != nil {
		var p sleModel.PachExternalSle
		if s.db.First(&p, sle.PachExternalSleID).Error == nil {
			in := input.PachExternalSle
			if in.Od != nil { p.Od = in.Od }
			if in.Os != nil { p.Os = in.Os }
			s.db.Save(&p)
		}
	}

	if input.VisualFields != nil && sle.VisualFieldsID != nil {
		var vf sleModel.VisualFields
		if s.db.First(&vf, *sle.VisualFieldsID).Error == nil {
			in := input.VisualFields
			if in.SuperoTemporoOd != nil  { vf.SuperoTemporoOd = in.SuperoTemporoOd }
			if in.SuperoTemporoOs != nil  { vf.SuperoTemporoOs = in.SuperoTemporoOs }
			if in.SuperoNasalOd != nil    { vf.SuperoNasalOd = in.SuperoNasalOd }
			if in.SuperoNasalOs != nil    { vf.SuperoNasalOs = in.SuperoNasalOs }
			if in.InferoTemporalOd != nil { vf.InferoTemporalOd = in.InferoTemporalOd }
			if in.InferoTemporalOs != nil { vf.InferoTemporalOs = in.InferoTemporalOs }
			if in.InferoNasalOd != nil    { vf.InferoNasalOd = in.InferoNasalOd }
			if in.InferoNasalOs != nil    { vf.InferoNasalOs = in.InferoNasalOs }
			if in.Instrument != nil       { vf.Instrument = in.Instrument }
			if in.Test != nil             { vf.Test = in.Test }
			if in.Reason != nil           { vf.Reason = in.Reason }
			if in.Result != nil           { vf.Result = in.Result }
			if in.Recommendations != nil  { vf.Recommendations = in.Recommendations }
			if in.Comments != nil         { vf.Comments = in.Comments }
			s.db.Save(&vf)
		}
	}

	if p := cleanDrawingPath(input.AddDrawing); p != "" {
		sle.AddDrawing = &p
	}
	if input.OdAngleEstimation != nil   { sle.OdAngleEstimation = *input.OdAngleEstimation }
	if input.OsAngleEstimation != nil   { sle.OsAngleEstimation = *input.OsAngleEstimation }
	if input.IopDropsFluress != nil     { sle.IopDropsFluress = input.IopDropsFluress }
	if input.IopDropsProparacaine != nil { sle.IopDropsProparacaine = input.IopDropsProparacaine }
	if input.IopDropsFluoroStrip != nil { sle.IopDropsFluoroStrip = input.IopDropsFluoroStrip }
	if input.Note != nil                { sle.Note = input.Note }
	s.db.Save(&sle)

	if input.TonometryEyes != nil {
		t, valErrs := parseTonometry(sle.IDExternalSleEye, input.TonometryEyes)
		if len(valErrs) > 0 {
			return nil, &ValidationError{Errors: valErrs}
		}
		s.db.Create(t)
	}

	activitylog.Log(s.db, "exam_sle", "update", activitylog.WithEntity(examID))
	return map[string]interface{}{
		"message": "ExternalSleEye updated successfully",
		"data":    sle,
	}, nil
}
