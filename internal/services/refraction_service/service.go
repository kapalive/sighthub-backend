package refraction_service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	empModel      "sighthub-backend/internal/models/employees"
	empLoginModel "sighthub-backend/internal/models/auth"
	refrModel     "sighthub-backend/internal/models/medical/vision_exam/refraction"
	visionModel   "sighthub-backend/internal/models/vision_exam"
	"sighthub-backend/pkg/activitylog"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

func boolVal(b *bool) bool {
	if b != nil {
		return *b
	}
	return false
}

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
	if err := s.db.Where("id_eye_exam = ?", examID).First(&exam).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	return &exam, nil
}

// ─── Input types ─────────────────────────────────────────────────────────────

type RetinoscopyInput struct {
	OdSph        *string `json:"od_sph"`
	OsSph        *string `json:"os_sph"`
	OdCyl        *string `json:"od_cyl"`
	OsCyl        *string `json:"os_cyl"`
	OdAxis       *string `json:"od_axis"`
	OsAxis       *string `json:"os_axis"`
	OdAdd        *string `json:"od_add"`
	OsAdd        *string `json:"os_add"`
	OdHPrism     *string `json:"od_h_prism"`
	OdHPrismList *string `json:"od_h_prism_list"`
	OsHPrism     *string `json:"os_h_prism"`
	OsHPrismList *string `json:"os_h_prism_list"`
	OdVPrism     *string `json:"od_v_prism"`
	OdVPrismList *string `json:"od_v_prism_list"`
	OsVPrism     *string `json:"os_v_prism"`
	OsVPrismList *string `json:"os_v_prism_list"`
	OdDva20      *string `json:"od_dva_20"`
	OsDva20      *string `json:"os_dva_20"`
	OdNva20      *string `json:"od_nva_20"`
	OsNva20      *string `json:"os_nva_20"`
	OuDva20      *string `json:"ou_dva_20"`
	OuNva20      *string `json:"ou_nva_20"`
	OdFinal      *bool   `json:"od_final"`
	OsFinal      *bool   `json:"os_final"`
}

type CycloInput struct {
	OdSph        *string `json:"od_sph"`
	OsSph        *string `json:"os_sph"`
	OdCyl        *string `json:"od_cyl"`
	OsCyl        *string `json:"os_cyl"`
	OdAxis       *string `json:"od_axis"`
	OsAxis       *string `json:"os_axis"`
	OdAdd        *string `json:"od_add"`
	OsAdd        *string `json:"os_add"`
	OdHPrism     *string `json:"od_h_prism"`
	OdHPrismList *string `json:"od_h_prism_list"`
	OsHPrism     *string `json:"os_h_prism"`
	OsHPrismList *string `json:"os_h_prism_list"`
	OdVPrism     *string `json:"od_v_prism"`
	OdVPrismList *string `json:"od_v_prism_list"`
	OsVPrism     *string `json:"os_v_prism"`
	OsVPrismList *string `json:"os_v_prism_list"`
	OdDva20      *string `json:"od_dva_20"`
	OsDva20      *string `json:"os_dva_20"`
	OdNva20      *string `json:"od_nva_20"`
	OsNva20      *string `json:"os_nva_20"`
	OuDva20      *string `json:"ou_dva_20"`
	OuNva20      *string `json:"ou_nva_20"`
	OdFinal      *bool   `json:"od_final"`
	OsFinal      *bool   `json:"os_final"`
	OdPd         *string `json:"od_pd"`
	OsPd         *string `json:"os_pd"`
	OuPd         *string `json:"ou_pd"`
	OdNpd        *string `json:"od_npd"`
	OsNpd        *string `json:"os_npd"`
	OuNpd        *string `json:"ou_npd"`
}

type ManifestInput struct {
	OdSph        *string `json:"od_sph"`
	OsSph        *string `json:"os_sph"`
	OdCyl        *string `json:"od_cyl"`
	OsCyl        *string `json:"os_cyl"`
	OdAxis       *string `json:"od_axis"`
	OsAxis       *string `json:"os_axis"`
	OdAdd        *string `json:"od_add"`
	OsAdd        *string `json:"os_add"`
	OdHPrism     *string `json:"od_h_prism"`
	OdHPrismList *string `json:"od_h_prism_list"`
	OsHPrism     *string `json:"os_h_prism"`
	OsHPrismList *string `json:"os_h_prism_list"`
	OdVPrism     *string `json:"od_v_prism"`
	OdVPrismList *string `json:"od_v_prism_list"`
	OsVPrism     *string `json:"os_v_prism"`
	OsVPrismList *string `json:"os_v_prism_list"`
	OdDva20      *string `json:"od_dva_20"`
	OsDva20      *string `json:"os_dva_20"`
	OdNva20      *string `json:"od_nva_20"`
	OsNva20      *string `json:"os_nva_20"`
	OuDva20      *string `json:"ou_dva_20"`
	OuNva20      *string `json:"ou_nva_20"`
	OdFinal      *bool   `json:"od_final"`
	OsFinal      *bool   `json:"os_final"`
	OdPd         *string `json:"od_pd"`
	OsPd         *string `json:"os_pd"`
	OuPd         *string `json:"ou_pd"`
	OdNpd        *string `json:"od_npd"`
	OsNpd        *string `json:"os_npd"`
	OuNpd        *string `json:"ou_npd"`
}

type FinalInput struct {
	OdSph        *string `json:"od_sph"`
	OsSph        *string `json:"os_sph"`
	OdCyl        *string `json:"od_cyl"`
	OsCyl        *string `json:"os_cyl"`
	OdAxis       *string `json:"od_axis"`
	OsAxis       *string `json:"os_axis"`
	OdAdd        *string `json:"od_add"`
	OsAdd        *string `json:"os_add"`
	OdHPrism     *string `json:"od_h_prism"`
	OdHPrismList *string `json:"od_h_prism_list"`
	OsHPrism     *string `json:"os_h_prism"`
	OsHPrismList *string `json:"os_h_prism_list"`
	OdVPrism     *string `json:"od_v_prism"`
	OdVPrismList *string `json:"od_v_prism_list"`
	OsVPrism     *string `json:"os_v_prism"`
	OsVPrismList *string `json:"os_v_prism_list"`
	OdDva20      *string `json:"od_dva_20"`
	OsDva20      *string `json:"os_dva_20"`
	OdNva20      *string `json:"od_nva_20"`
	OsNva20      *string `json:"os_nva_20"`
	OuDva20      *string `json:"ou_dva_20"`
	OuNva20      *string `json:"ou_nva_20"`
	OdPd         *string `json:"od_pd"`
	OsPd         *string `json:"os_pd"`
	OuPd         *string `json:"ou_pd"`
	OdNpd        *string `json:"od_npd"`
	OsNpd        *string `json:"os_npd"`
	OuNpd        *string `json:"ou_npd"`
	ExpireDate   *string `json:"expire_date"`
	Note         *string `json:"note"`
}

type Final2Input struct {
	OdSph        *string `json:"od_sph"`
	OsSph        *string `json:"os_sph"`
	OdCyl        *string `json:"od_cyl"`
	OsCyl        *string `json:"os_cyl"`
	OdAxis       *string `json:"od_axis"`
	OsAxis       *string `json:"os_axis"`
	OdAdd        *string `json:"od_add"`
	OsAdd        *string `json:"os_add"`
	OdHPrism     *string `json:"od_h_prism"`
	OdHPrismList *string `json:"od_h_prism_list"`
	OsHPrism     *string `json:"os_h_prism"`
	OsHPrismList *string `json:"os_h_prism_list"`
	OdVPrism     *string `json:"od_v_prism"`
	OdVPrismList *string `json:"od_v_prism_list"`
	OsVPrism     *string `json:"os_v_prism"`
	OsVPrismList *string `json:"os_v_prism_list"`
	OdDva20      *string `json:"od_dva_20"`
	OsDva20      *string `json:"os_dva_20"`
	OuDva20      *string `json:"ou_dva_20"`
	OdPd         *string `json:"od_pd"`
	OsPd         *string `json:"os_pd"`
	OuPd         *string `json:"ou_pd"`
	OdNpd        *string `json:"od_npd"`
	OsNpd        *string `json:"os_npd"`
	OuNpd        *string `json:"ou_npd"`
	Desc         *string `json:"desc"`
	Note         *string `json:"note"`
}

type Final3Input struct {
	OdSph  *string `json:"od_sph"`
	OsSph  *string `json:"os_sph"`
	OdCyl  *string `json:"od_cyl"`
	OsCyl  *string `json:"os_cyl"`
	OdAxis *string `json:"od_axis"`
	OsAxis *string `json:"os_axis"`
	OdAdd  *string `json:"od_add"`
	OsAdd  *string `json:"os_add"`
	OdDva20 *string `json:"od_dva_20"`
	OsDva20 *string `json:"os_dva_20"`
	OuDva20 *string `json:"ou_dva_20"`
	OdPd   *string `json:"od_pd"`
	OsPd   *string `json:"os_pd"`
	OuPd   *string `json:"ou_pd"`
	OdNpd  *string `json:"od_npd"`
	OsNpd  *string `json:"os_npd"`
	OuNpd  *string `json:"ou_npd"`
}

type SaveRefractionInput struct {
	Retinoscopy RetinoscopyInput `json:"retinoscopy"`
	Cyclo       CycloInput       `json:"cyclo"`
	Manifest    ManifestInput    `json:"manifest"`
	Final       FinalInput       `json:"final"`
	Final2      Final2Input      `json:"final2"`
	Final3      Final3Input      `json:"final3"`
	DrNote      *string          `json:"dr_note"`
}

type UpdateRefractionInput struct {
	Retinoscopy *RetinoscopyInput `json:"retinoscopy"`
	Cyclo       *CycloInput       `json:"cyclo"`
	Manifest    *ManifestInput    `json:"manifest"`
	Final       *FinalInput       `json:"final"`
	Final2      *Final2Input      `json:"final2"`
	Final3      *Final3Input      `json:"final3"`
	DrNote      *string           `json:"dr_note"`
}

// ─── SaveRefraction ───────────────────────────────────────────────────────────

func (s *Service) SaveRefraction(username string, examID int64, input SaveRefractionInput) error {
	emp, err := s.getEmployee(username)
	if err != nil {
		return err
	}

	exam, err := s.getExam(examID)
	if err != nil {
		return err
	}

	if exam.Passed {
		return errors.New("cannot update refraction for a completed exam")
	}

	if exam.EmployeeID != int64(emp.IDEmployee) {
		return errors.New("not authorized")
	}

	// Create Retinoscopy
	r := refrModel.Retinoscopy{
		OdSph:        input.Retinoscopy.OdSph,
		OsSph:        input.Retinoscopy.OsSph,
		OdCyl:        input.Retinoscopy.OdCyl,
		OsCyl:        input.Retinoscopy.OsCyl,
		OdAxis:       input.Retinoscopy.OdAxis,
		OsAxis:       input.Retinoscopy.OsAxis,
		OdAdd:        input.Retinoscopy.OdAdd,
		OsAdd:        input.Retinoscopy.OsAdd,
		OdHPrism:     input.Retinoscopy.OdHPrism,
		OdHPrismList: input.Retinoscopy.OdHPrismList,
		OsHPrism:     input.Retinoscopy.OsHPrism,
		OsHPrismList: input.Retinoscopy.OsHPrismList,
		OdVPrism:     input.Retinoscopy.OdVPrism,
		OdVPrismList: input.Retinoscopy.OdVPrismList,
		OsVPrism:     input.Retinoscopy.OsVPrism,
		OsVPrismList: input.Retinoscopy.OsVPrismList,
		OdDva20:      input.Retinoscopy.OdDva20,
		OsDva20:      input.Retinoscopy.OsDva20,
		OdNva20:      input.Retinoscopy.OdNva20,
		OsNva20:      input.Retinoscopy.OsNva20,
		OuDva20:      input.Retinoscopy.OuDva20,
		OuNva20:      input.Retinoscopy.OuNva20,
		OdFinal:      boolVal(input.Retinoscopy.OdFinal),
		OsFinal:      boolVal(input.Retinoscopy.OsFinal),
	}
	if err := s.db.Create(&r).Error; err != nil {
		return err
	}

	// Create Cyclo
	c := refrModel.Cyclo{
		OdSph:        input.Cyclo.OdSph,
		OsSph:        input.Cyclo.OsSph,
		OdCyl:        input.Cyclo.OdCyl,
		OsCyl:        input.Cyclo.OsCyl,
		OdAxis:       input.Cyclo.OdAxis,
		OsAxis:       input.Cyclo.OsAxis,
		OdAdd:        input.Cyclo.OdAdd,
		OsAdd:        input.Cyclo.OsAdd,
		OdHPrism:     input.Cyclo.OdHPrism,
		OdHPrismList: input.Cyclo.OdHPrismList,
		OsHPrism:     input.Cyclo.OsHPrism,
		OsHPrismList: input.Cyclo.OsHPrismList,
		OdVPrism:     input.Cyclo.OdVPrism,
		OdVPrismList: input.Cyclo.OdVPrismList,
		OsVPrism:     input.Cyclo.OsVPrism,
		OsVPrismList: input.Cyclo.OsVPrismList,
		OdDva20:      input.Cyclo.OdDva20,
		OsDva20:      input.Cyclo.OsDva20,
		OdNva20:      input.Cyclo.OdNva20,
		OsNva20:      input.Cyclo.OsNva20,
		OuDva20:      input.Cyclo.OuDva20,
		OuNva20:      input.Cyclo.OuNva20,
		OdFinal:      boolVal(input.Cyclo.OdFinal),
		OsFinal:      boolVal(input.Cyclo.OsFinal),
		OdPd:         input.Cyclo.OdPd,
		OsPd:         input.Cyclo.OsPd,
		OuPd:         input.Cyclo.OuPd,
		OdNpd:        input.Cyclo.OdNpd,
		OsNpd:        input.Cyclo.OsNpd,
		OuNpd:        input.Cyclo.OuNpd,
	}
	if err := s.db.Create(&c).Error; err != nil {
		return err
	}

	// Create Manifest
	m := refrModel.Manifest{
		OdSph:        input.Manifest.OdSph,
		OsSph:        input.Manifest.OsSph,
		OdCyl:        input.Manifest.OdCyl,
		OsCyl:        input.Manifest.OsCyl,
		OdAxis:       input.Manifest.OdAxis,
		OsAxis:       input.Manifest.OsAxis,
		OdAdd:        input.Manifest.OdAdd,
		OsAdd:        input.Manifest.OsAdd,
		OdHPrism:     input.Manifest.OdHPrism,
		OdHPrismList: input.Manifest.OdHPrismList,
		OsHPrism:     input.Manifest.OsHPrism,
		OsHPrismList: input.Manifest.OsHPrismList,
		OdVPrism:     input.Manifest.OdVPrism,
		OdVPrismList: input.Manifest.OdVPrismList,
		OsVPrism:     input.Manifest.OsVPrism,
		OsVPrismList: input.Manifest.OsVPrismList,
		OdDva20:      input.Manifest.OdDva20,
		OsDva20:      input.Manifest.OsDva20,
		OdNva20:      input.Manifest.OdNva20,
		OsNva20:      input.Manifest.OsNva20,
		OuDva20:      input.Manifest.OuDva20,
		OuNva20:      input.Manifest.OuNva20,
		OdFinal:      boolVal(input.Manifest.OdFinal),
		OsFinal:      boolVal(input.Manifest.OsFinal),
		OdPd:         input.Manifest.OdPd,
		OsPd:         input.Manifest.OsPd,
		OuPd:         input.Manifest.OuPd,
		OdNpd:        input.Manifest.OdNpd,
		OsNpd:        input.Manifest.OsNpd,
		OuNpd:        input.Manifest.OuNpd,
	}
	if err := s.db.Create(&m).Error; err != nil {
		return err
	}

	// Create RefractionFinal
	var expireDate *time.Time
	if input.Final.ExpireDate != nil {
		t, err := time.Parse("2006-01-02", *input.Final.ExpireDate)
		if err == nil {
			expireDate = &t
		}
	}
	if expireDate == nil {
		now := time.Now().UTC()
		expireDate = &now
	}
	f := refrModel.RefractionFinal{
		OdSph:        input.Final.OdSph,
		OsSph:        input.Final.OsSph,
		OdCyl:        input.Final.OdCyl,
		OsCyl:        input.Final.OsCyl,
		OdAxis:       input.Final.OdAxis,
		OsAxis:       input.Final.OsAxis,
		OdAdd:        input.Final.OdAdd,
		OsAdd:        input.Final.OsAdd,
		OdHPrism:     input.Final.OdHPrism,
		OdHPrismList: input.Final.OdHPrismList,
		OsHPrism:     input.Final.OsHPrism,
		OsHPrismList: input.Final.OsHPrismList,
		OdVPrism:     input.Final.OdVPrism,
		OdVPrismList: input.Final.OdVPrismList,
		OsVPrism:     input.Final.OsVPrism,
		OsVPrismList: input.Final.OsVPrismList,
		OdDva20:      input.Final.OdDva20,
		OsDva20:      input.Final.OsDva20,
		OdNva20:      input.Final.OdNva20,
		OsNva20:      input.Final.OsNva20,
		OuDva20:      input.Final.OuDva20,
		OuNva20:      input.Final.OuNva20,
		OdPd:         input.Final.OdPd,
		OsPd:         input.Final.OsPd,
		OuPd:         input.Final.OuPd,
		OdNpd:        input.Final.OdNpd,
		OsNpd:        input.Final.OsNpd,
		OuNpd:        input.Final.OuNpd,
		ExpireDate:   expireDate,
		Note:         input.Final.Note,
	}
	if err := s.db.Create(&f).Error; err != nil {
		return err
	}

	// Create Final2
	f2 := refrModel.Final2{
		OdSph:        input.Final2.OdSph,
		OsSph:        input.Final2.OsSph,
		OdCyl:        input.Final2.OdCyl,
		OsCyl:        input.Final2.OsCyl,
		OdAxis:       input.Final2.OdAxis,
		OsAxis:       input.Final2.OsAxis,
		OdAdd:        input.Final2.OdAdd,
		OsAdd:        input.Final2.OsAdd,
		OdHPrism:     input.Final2.OdHPrism,
		OdHPrismList: input.Final2.OdHPrismList,
		OsHPrism:     input.Final2.OsHPrism,
		OsHPrismList: input.Final2.OsHPrismList,
		OdVPrism:     input.Final2.OdVPrism,
		OdVPrismList: input.Final2.OdVPrismList,
		OsVPrism:     input.Final2.OsVPrism,
		OsVPrismList: input.Final2.OsVPrismList,
		OdDva20:      input.Final2.OdDva20,
		OsDva20:      input.Final2.OsDva20,
		OuDva20:      input.Final2.OuDva20,
		OdPd:         input.Final2.OdPd,
		OsPd:         input.Final2.OsPd,
		OuPd:         input.Final2.OuPd,
		OdNpd:        input.Final2.OdNpd,
		OsNpd:        input.Final2.OsNpd,
		OuNpd:        input.Final2.OuNpd,
		Desc:         input.Final2.Desc,
		Note:         input.Final2.Note,
	}
	if err := s.db.Create(&f2).Error; err != nil {
		return err
	}

	// Create Final3
	f3 := refrModel.Final3{
		OdSph:   input.Final3.OdSph,
		OsSph:   input.Final3.OsSph,
		OdCyl:   input.Final3.OdCyl,
		OsCyl:   input.Final3.OsCyl,
		OdAxis:  input.Final3.OdAxis,
		OsAxis:  input.Final3.OsAxis,
		OdAdd:   input.Final3.OdAdd,
		OsAdd:   input.Final3.OsAdd,
		OdDva20: input.Final3.OdDva20,
		OsDva20: input.Final3.OsDva20,
		OuDva20: input.Final3.OuDva20,
		OdPd:    input.Final3.OdPd,
		OsPd:    input.Final3.OsPd,
		OuPd:    input.Final3.OuPd,
		OdNpd:   input.Final3.OdNpd,
		OsNpd:   input.Final3.OsNpd,
		OuNpd:   input.Final3.OuNpd,
	}
	if err := s.db.Create(&f3).Error; err != nil {
		return err
	}

	// Upsert RefractionEye
	var re refrModel.RefractionEye
	s.db.Where("eye_exam_id = ?", examID).First(&re)
	re.EyeExamID = examID
	re.RetinoscopyID = r.IDRetinoscopy
	re.CycloID = c.IDCyclo
	re.ManifestID = m.IDManifest
	re.FinalID = f.IDFinal
	id2 := f2.IDFinal2
	re.Final2ID = &id2
	id3 := f3.IDFinal3
	re.Final3ID = &id3
	re.DrNote = input.DrNote
	if err := s.db.Save(&re).Error; err != nil {
		return err
	}

	activitylog.Log(s.db, "exam_refraction", "save", activitylog.WithEntity(examID))
	return nil
}

// ─── GetRefraction ────────────────────────────────────────────────────────────

func (s *Service) GetRefraction(examID int64) (map[string]interface{}, error) {
	if _, err := s.getExam(examID); err != nil {
		return nil, err
	}

	var re refrModel.RefractionEye
	if err := s.db.Where("eye_exam_id = ?", examID).First(&re).Error; err != nil {
		return map[string]interface{}{
			"exam_id":        examID,
			"exists":         false,
			"refraction_eye": nil,
			"retinoscopy":    nil,
			"cyclo":          nil,
			"manifest":       nil,
			"final":          nil,
			"final2":         nil,
			"final3":         nil,
			"dr_note":        nil,
		}, nil
	}

	var retin refrModel.Retinoscopy
	s.db.First(&retin, re.RetinoscopyID)

	var cyclo refrModel.Cyclo
	s.db.First(&cyclo, re.CycloID)

	var mani refrModel.Manifest
	s.db.First(&mani, re.ManifestID)

	var final refrModel.RefractionFinal
	s.db.First(&final, re.FinalID)

	result := map[string]interface{}{
		"exam_id":        examID,
		"exists":         true,
		"refraction_eye": re.ToMap(),
		"retinoscopy":    retin.ToMap(),
		"cyclo":          cyclo.ToMap(),
		"manifest":       mani.ToMap(),
		"final":          final.ToMap(),
		"final2":         nil,
		"final3":         nil,
		"dr_note":        re.DrNote,
	}

	if re.Final2ID != nil {
		var f2 refrModel.Final2
		if err := s.db.First(&f2, *re.Final2ID).Error; err == nil {
			result["final2"] = f2.ToMap()
		}
	}

	if re.Final3ID != nil {
		var f3 refrModel.Final3
		if err := s.db.First(&f3, *re.Final3ID).Error; err == nil {
			result["final3"] = f3.ToMap()
		}
	}

	return result, nil
}

// ─── UpdateRefraction ─────────────────────────────────────────────────────────

func (s *Service) UpdateRefraction(username string, examID int64, input UpdateRefractionInput) (map[string]interface{}, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}

	exam, err := s.getExam(examID)
	if err != nil {
		return nil, err
	}

	if exam.Passed {
		return nil, errors.New("cannot update refraction for a completed exam")
	}

	if exam.EmployeeID != int64(emp.IDEmployee) {
		return nil, errors.New("not authorized")
	}

	var re refrModel.RefractionEye
	if err := s.db.Where("eye_exam_id = ?", examID).First(&re).Error; err != nil {
		return nil, errors.New("refraction not found")
	}

	if input.Retinoscopy != nil {
		var retin refrModel.Retinoscopy
		if err := s.db.First(&retin, re.RetinoscopyID).Error; err == nil {
			if input.Retinoscopy.OdSph != nil        { retin.OdSph = input.Retinoscopy.OdSph }
			if input.Retinoscopy.OsSph != nil        { retin.OsSph = input.Retinoscopy.OsSph }
			if input.Retinoscopy.OdCyl != nil        { retin.OdCyl = input.Retinoscopy.OdCyl }
			if input.Retinoscopy.OsCyl != nil        { retin.OsCyl = input.Retinoscopy.OsCyl }
			if input.Retinoscopy.OdAxis != nil       { retin.OdAxis = input.Retinoscopy.OdAxis }
			if input.Retinoscopy.OsAxis != nil       { retin.OsAxis = input.Retinoscopy.OsAxis }
			if input.Retinoscopy.OdAdd != nil        { retin.OdAdd = input.Retinoscopy.OdAdd }
			if input.Retinoscopy.OsAdd != nil        { retin.OsAdd = input.Retinoscopy.OsAdd }
			if input.Retinoscopy.OdHPrism != nil     { retin.OdHPrism = input.Retinoscopy.OdHPrism }
			if input.Retinoscopy.OdHPrismList != nil { retin.OdHPrismList = input.Retinoscopy.OdHPrismList }
			if input.Retinoscopy.OsHPrism != nil     { retin.OsHPrism = input.Retinoscopy.OsHPrism }
			if input.Retinoscopy.OsHPrismList != nil { retin.OsHPrismList = input.Retinoscopy.OsHPrismList }
			if input.Retinoscopy.OdVPrism != nil     { retin.OdVPrism = input.Retinoscopy.OdVPrism }
			if input.Retinoscopy.OdVPrismList != nil { retin.OdVPrismList = input.Retinoscopy.OdVPrismList }
			if input.Retinoscopy.OsVPrism != nil     { retin.OsVPrism = input.Retinoscopy.OsVPrism }
			if input.Retinoscopy.OsVPrismList != nil { retin.OsVPrismList = input.Retinoscopy.OsVPrismList }
			if input.Retinoscopy.OdDva20 != nil      { retin.OdDva20 = input.Retinoscopy.OdDva20 }
			if input.Retinoscopy.OsDva20 != nil      { retin.OsDva20 = input.Retinoscopy.OsDva20 }
			if input.Retinoscopy.OdNva20 != nil      { retin.OdNva20 = input.Retinoscopy.OdNva20 }
			if input.Retinoscopy.OsNva20 != nil      { retin.OsNva20 = input.Retinoscopy.OsNva20 }
			if input.Retinoscopy.OuDva20 != nil      { retin.OuDva20 = input.Retinoscopy.OuDva20 }
			if input.Retinoscopy.OuNva20 != nil      { retin.OuNva20 = input.Retinoscopy.OuNva20 }
			if input.Retinoscopy.OdFinal != nil      { retin.OdFinal = *input.Retinoscopy.OdFinal }
			if input.Retinoscopy.OsFinal != nil      { retin.OsFinal = *input.Retinoscopy.OsFinal }
			s.db.Save(&retin)
		}
	}

	if input.Cyclo != nil {
		var cyclo refrModel.Cyclo
		if err := s.db.First(&cyclo, re.CycloID).Error; err == nil {
			if input.Cyclo.OdSph != nil        { cyclo.OdSph = input.Cyclo.OdSph }
			if input.Cyclo.OsSph != nil        { cyclo.OsSph = input.Cyclo.OsSph }
			if input.Cyclo.OdCyl != nil        { cyclo.OdCyl = input.Cyclo.OdCyl }
			if input.Cyclo.OsCyl != nil        { cyclo.OsCyl = input.Cyclo.OsCyl }
			if input.Cyclo.OdAxis != nil       { cyclo.OdAxis = input.Cyclo.OdAxis }
			if input.Cyclo.OsAxis != nil       { cyclo.OsAxis = input.Cyclo.OsAxis }
			if input.Cyclo.OdAdd != nil        { cyclo.OdAdd = input.Cyclo.OdAdd }
			if input.Cyclo.OsAdd != nil        { cyclo.OsAdd = input.Cyclo.OsAdd }
			if input.Cyclo.OdHPrism != nil     { cyclo.OdHPrism = input.Cyclo.OdHPrism }
			if input.Cyclo.OdHPrismList != nil { cyclo.OdHPrismList = input.Cyclo.OdHPrismList }
			if input.Cyclo.OsHPrism != nil     { cyclo.OsHPrism = input.Cyclo.OsHPrism }
			if input.Cyclo.OsHPrismList != nil { cyclo.OsHPrismList = input.Cyclo.OsHPrismList }
			if input.Cyclo.OdVPrism != nil     { cyclo.OdVPrism = input.Cyclo.OdVPrism }
			if input.Cyclo.OdVPrismList != nil { cyclo.OdVPrismList = input.Cyclo.OdVPrismList }
			if input.Cyclo.OsVPrism != nil     { cyclo.OsVPrism = input.Cyclo.OsVPrism }
			if input.Cyclo.OsVPrismList != nil { cyclo.OsVPrismList = input.Cyclo.OsVPrismList }
			if input.Cyclo.OdDva20 != nil      { cyclo.OdDva20 = input.Cyclo.OdDva20 }
			if input.Cyclo.OsDva20 != nil      { cyclo.OsDva20 = input.Cyclo.OsDva20 }
			if input.Cyclo.OdNva20 != nil      { cyclo.OdNva20 = input.Cyclo.OdNva20 }
			if input.Cyclo.OsNva20 != nil      { cyclo.OsNva20 = input.Cyclo.OsNva20 }
			if input.Cyclo.OuDva20 != nil      { cyclo.OuDva20 = input.Cyclo.OuDva20 }
			if input.Cyclo.OuNva20 != nil      { cyclo.OuNva20 = input.Cyclo.OuNva20 }
			if input.Cyclo.OdFinal != nil      { cyclo.OdFinal = *input.Cyclo.OdFinal }
			if input.Cyclo.OsFinal != nil      { cyclo.OsFinal = *input.Cyclo.OsFinal }
			if input.Cyclo.OdPd != nil         { cyclo.OdPd = input.Cyclo.OdPd }
			if input.Cyclo.OsPd != nil         { cyclo.OsPd = input.Cyclo.OsPd }
			if input.Cyclo.OuPd != nil         { cyclo.OuPd = input.Cyclo.OuPd }
			if input.Cyclo.OdNpd != nil        { cyclo.OdNpd = input.Cyclo.OdNpd }
			if input.Cyclo.OsNpd != nil        { cyclo.OsNpd = input.Cyclo.OsNpd }
			if input.Cyclo.OuNpd != nil        { cyclo.OuNpd = input.Cyclo.OuNpd }
			s.db.Save(&cyclo)
		}
	}

	if input.Manifest != nil {
		var mani refrModel.Manifest
		if err := s.db.First(&mani, re.ManifestID).Error; err == nil {
			if input.Manifest.OdSph != nil        { mani.OdSph = input.Manifest.OdSph }
			if input.Manifest.OsSph != nil        { mani.OsSph = input.Manifest.OsSph }
			if input.Manifest.OdCyl != nil        { mani.OdCyl = input.Manifest.OdCyl }
			if input.Manifest.OsCyl != nil        { mani.OsCyl = input.Manifest.OsCyl }
			if input.Manifest.OdAxis != nil       { mani.OdAxis = input.Manifest.OdAxis }
			if input.Manifest.OsAxis != nil       { mani.OsAxis = input.Manifest.OsAxis }
			if input.Manifest.OdAdd != nil        { mani.OdAdd = input.Manifest.OdAdd }
			if input.Manifest.OsAdd != nil        { mani.OsAdd = input.Manifest.OsAdd }
			if input.Manifest.OdHPrism != nil     { mani.OdHPrism = input.Manifest.OdHPrism }
			if input.Manifest.OdHPrismList != nil { mani.OdHPrismList = input.Manifest.OdHPrismList }
			if input.Manifest.OsHPrism != nil     { mani.OsHPrism = input.Manifest.OsHPrism }
			if input.Manifest.OsHPrismList != nil { mani.OsHPrismList = input.Manifest.OsHPrismList }
			if input.Manifest.OdVPrism != nil     { mani.OdVPrism = input.Manifest.OdVPrism }
			if input.Manifest.OdVPrismList != nil { mani.OdVPrismList = input.Manifest.OdVPrismList }
			if input.Manifest.OsVPrism != nil     { mani.OsVPrism = input.Manifest.OsVPrism }
			if input.Manifest.OsVPrismList != nil { mani.OsVPrismList = input.Manifest.OsVPrismList }
			if input.Manifest.OdDva20 != nil      { mani.OdDva20 = input.Manifest.OdDva20 }
			if input.Manifest.OsDva20 != nil      { mani.OsDva20 = input.Manifest.OsDva20 }
			if input.Manifest.OdNva20 != nil      { mani.OdNva20 = input.Manifest.OdNva20 }
			if input.Manifest.OsNva20 != nil      { mani.OsNva20 = input.Manifest.OsNva20 }
			if input.Manifest.OuDva20 != nil      { mani.OuDva20 = input.Manifest.OuDva20 }
			if input.Manifest.OuNva20 != nil      { mani.OuNva20 = input.Manifest.OuNva20 }
			if input.Manifest.OdFinal != nil      { mani.OdFinal = *input.Manifest.OdFinal }
			if input.Manifest.OsFinal != nil      { mani.OsFinal = *input.Manifest.OsFinal }
			if input.Manifest.OdPd != nil         { mani.OdPd = input.Manifest.OdPd }
			if input.Manifest.OsPd != nil         { mani.OsPd = input.Manifest.OsPd }
			if input.Manifest.OuPd != nil         { mani.OuPd = input.Manifest.OuPd }
			if input.Manifest.OdNpd != nil        { mani.OdNpd = input.Manifest.OdNpd }
			if input.Manifest.OsNpd != nil        { mani.OsNpd = input.Manifest.OsNpd }
			if input.Manifest.OuNpd != nil        { mani.OuNpd = input.Manifest.OuNpd }
			s.db.Save(&mani)
		}
	}

	if input.Final != nil {
		var final refrModel.RefractionFinal
		if err := s.db.First(&final, re.FinalID).Error; err == nil {
			if input.Final.OdSph != nil        { final.OdSph = input.Final.OdSph }
			if input.Final.OsSph != nil        { final.OsSph = input.Final.OsSph }
			if input.Final.OdCyl != nil        { final.OdCyl = input.Final.OdCyl }
			if input.Final.OsCyl != nil        { final.OsCyl = input.Final.OsCyl }
			if input.Final.OdAxis != nil       { final.OdAxis = input.Final.OdAxis }
			if input.Final.OsAxis != nil       { final.OsAxis = input.Final.OsAxis }
			if input.Final.OdAdd != nil        { final.OdAdd = input.Final.OdAdd }
			if input.Final.OsAdd != nil        { final.OsAdd = input.Final.OsAdd }
			if input.Final.OdHPrism != nil     { final.OdHPrism = input.Final.OdHPrism }
			if input.Final.OdHPrismList != nil { final.OdHPrismList = input.Final.OdHPrismList }
			if input.Final.OsHPrism != nil     { final.OsHPrism = input.Final.OsHPrism }
			if input.Final.OsHPrismList != nil { final.OsHPrismList = input.Final.OsHPrismList }
			if input.Final.OdVPrism != nil     { final.OdVPrism = input.Final.OdVPrism }
			if input.Final.OdVPrismList != nil { final.OdVPrismList = input.Final.OdVPrismList }
			if input.Final.OsVPrism != nil     { final.OsVPrism = input.Final.OsVPrism }
			if input.Final.OsVPrismList != nil { final.OsVPrismList = input.Final.OsVPrismList }
			if input.Final.OdDva20 != nil      { final.OdDva20 = input.Final.OdDva20 }
			if input.Final.OsDva20 != nil      { final.OsDva20 = input.Final.OsDva20 }
			if input.Final.OdNva20 != nil      { final.OdNva20 = input.Final.OdNva20 }
			if input.Final.OsNva20 != nil      { final.OsNva20 = input.Final.OsNva20 }
			if input.Final.OuDva20 != nil      { final.OuDva20 = input.Final.OuDva20 }
			if input.Final.OuNva20 != nil      { final.OuNva20 = input.Final.OuNva20 }
			if input.Final.OdPd != nil         { final.OdPd = input.Final.OdPd }
			if input.Final.OsPd != nil         { final.OsPd = input.Final.OsPd }
			if input.Final.OuPd != nil         { final.OuPd = input.Final.OuPd }
			if input.Final.OdNpd != nil        { final.OdNpd = input.Final.OdNpd }
			if input.Final.OsNpd != nil        { final.OsNpd = input.Final.OsNpd }
			if input.Final.OuNpd != nil        { final.OuNpd = input.Final.OuNpd }
			if input.Final.Note != nil         { final.Note = input.Final.Note }
			if input.Final.ExpireDate != nil {
				t, err := time.Parse("2006-01-02", *input.Final.ExpireDate)
				if err == nil {
					final.ExpireDate = &t
				}
			}
			s.db.Save(&final)
		}
	}

	if input.Final2 != nil && re.Final2ID != nil {
		var f2 refrModel.Final2
		if err := s.db.First(&f2, *re.Final2ID).Error; err == nil {
			if input.Final2.OdSph != nil        { f2.OdSph = input.Final2.OdSph }
			if input.Final2.OsSph != nil        { f2.OsSph = input.Final2.OsSph }
			if input.Final2.OdCyl != nil        { f2.OdCyl = input.Final2.OdCyl }
			if input.Final2.OsCyl != nil        { f2.OsCyl = input.Final2.OsCyl }
			if input.Final2.OdAxis != nil       { f2.OdAxis = input.Final2.OdAxis }
			if input.Final2.OsAxis != nil       { f2.OsAxis = input.Final2.OsAxis }
			if input.Final2.OdAdd != nil        { f2.OdAdd = input.Final2.OdAdd }
			if input.Final2.OsAdd != nil        { f2.OsAdd = input.Final2.OsAdd }
			if input.Final2.OdHPrism != nil     { f2.OdHPrism = input.Final2.OdHPrism }
			if input.Final2.OdHPrismList != nil { f2.OdHPrismList = input.Final2.OdHPrismList }
			if input.Final2.OsHPrism != nil     { f2.OsHPrism = input.Final2.OsHPrism }
			if input.Final2.OsHPrismList != nil { f2.OsHPrismList = input.Final2.OsHPrismList }
			if input.Final2.OdVPrism != nil     { f2.OdVPrism = input.Final2.OdVPrism }
			if input.Final2.OdVPrismList != nil { f2.OdVPrismList = input.Final2.OdVPrismList }
			if input.Final2.OsVPrism != nil     { f2.OsVPrism = input.Final2.OsVPrism }
			if input.Final2.OsVPrismList != nil { f2.OsVPrismList = input.Final2.OsVPrismList }
			if input.Final2.OdDva20 != nil      { f2.OdDva20 = input.Final2.OdDva20 }
			if input.Final2.OsDva20 != nil      { f2.OsDva20 = input.Final2.OsDva20 }
			if input.Final2.OuDva20 != nil      { f2.OuDva20 = input.Final2.OuDva20 }
			if input.Final2.OdPd != nil         { f2.OdPd = input.Final2.OdPd }
			if input.Final2.OsPd != nil         { f2.OsPd = input.Final2.OsPd }
			if input.Final2.OuPd != nil         { f2.OuPd = input.Final2.OuPd }
			if input.Final2.OdNpd != nil        { f2.OdNpd = input.Final2.OdNpd }
			if input.Final2.OsNpd != nil        { f2.OsNpd = input.Final2.OsNpd }
			if input.Final2.OuNpd != nil        { f2.OuNpd = input.Final2.OuNpd }
			if input.Final2.Desc != nil         { f2.Desc = input.Final2.Desc }
			if input.Final2.Note != nil         { f2.Note = input.Final2.Note }
			s.db.Save(&f2)
		}
	}

	if input.Final3 != nil && re.Final3ID != nil {
		var f3 refrModel.Final3
		if err := s.db.First(&f3, *re.Final3ID).Error; err == nil {
			if input.Final3.OdSph != nil   { f3.OdSph = input.Final3.OdSph }
			if input.Final3.OsSph != nil   { f3.OsSph = input.Final3.OsSph }
			if input.Final3.OdCyl != nil   { f3.OdCyl = input.Final3.OdCyl }
			if input.Final3.OsCyl != nil   { f3.OsCyl = input.Final3.OsCyl }
			if input.Final3.OdAxis != nil  { f3.OdAxis = input.Final3.OdAxis }
			if input.Final3.OsAxis != nil  { f3.OsAxis = input.Final3.OsAxis }
			if input.Final3.OdAdd != nil   { f3.OdAdd = input.Final3.OdAdd }
			if input.Final3.OsAdd != nil   { f3.OsAdd = input.Final3.OsAdd }
			if input.Final3.OdDva20 != nil { f3.OdDva20 = input.Final3.OdDva20 }
			if input.Final3.OsDva20 != nil { f3.OsDva20 = input.Final3.OsDva20 }
			if input.Final3.OuDva20 != nil { f3.OuDva20 = input.Final3.OuDva20 }
			if input.Final3.OdPd != nil    { f3.OdPd = input.Final3.OdPd }
			if input.Final3.OsPd != nil    { f3.OsPd = input.Final3.OsPd }
			if input.Final3.OuPd != nil    { f3.OuPd = input.Final3.OuPd }
			if input.Final3.OdNpd != nil   { f3.OdNpd = input.Final3.OdNpd }
			if input.Final3.OsNpd != nil   { f3.OsNpd = input.Final3.OsNpd }
			if input.Final3.OuNpd != nil   { f3.OuNpd = input.Final3.OuNpd }
			s.db.Save(&f3)
		}
	}

	if input.DrNote != nil {
		re.DrNote = input.DrNote
		s.db.Save(&re)
	}

	activitylog.Log(s.db, "exam_refraction", "update", activitylog.WithEntity(examID))
	return s.GetRefraction(examID)
}
