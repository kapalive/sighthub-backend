package cl_fitting_service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	empLoginModel "sighthub-backend/internal/models/auth"
	clModel "sighthub-backend/internal/models/medical/vision_exam/cl_fitting"
	empModel "sighthub-backend/internal/models/employees"
	vendorModel "sighthub-backend/internal/models/vendors"
	visionModel "sighthub-backend/internal/models/vision_exam"
	"sighthub-backend/pkg/activitylog"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ---------- input types ----------

type FittingInput struct {
	OdBrand          *string `json:"od_brand"`
	OsBrand          *string `json:"os_brand"`
	OdBCur           *string `json:"od_b_cur"`
	OsBCur           *string `json:"os_b_cur"`
	OdDia            *string `json:"od_dia"`
	OsDia            *string `json:"os_dia"`
	OdPwr            *string `json:"od_pwr"`
	OsPwr            *string `json:"os_pwr"`
	OdCyl            *string `json:"od_cyl"`
	OsCyl            *string `json:"os_cyl"`
	OdAxis           *string `json:"od_axis"`
	OsAxis           *string `json:"os_axis"`
	OdAdd            *string `json:"od_add"`
	OsAdd            *string `json:"os_add"`
	OdDva20          *string `json:"od_dva_20"`
	OsDva20          *string `json:"os_dva_20"`
	OdNva20          *string `json:"od_nva_20"`
	OsNva20          *string `json:"os_nva_20"`
	OdOverRefraction *string `json:"od_over_refraction"`
	OsOverRefraction *string `json:"os_over_refraction"`
	OdFinal          *bool   `json:"od_final"`
	OsFinal          *bool   `json:"os_final"`
	Evaluation       *string `json:"evaluation"`
	DominantEye      *string `json:"dominant_eye"`
}

type TrialInput struct {
	OdBrand           *string `json:"od_brand"`
	OsBrand           *string `json:"os_brand"`
	OdBCur            *string `json:"od_b_cur"`
	OsBCur            *string `json:"os_b_cur"`
	OdDia             *string `json:"od_dia"`
	OsDia             *string `json:"os_dia"`
	OdPwr             *string `json:"od_pwr"`
	OsPwr             *string `json:"os_pwr"`
	OdCyl             *string `json:"od_cyl"`
	OsCyl             *string `json:"os_cyl"`
	OdAxis            *string `json:"od_axis"`
	OsAxis            *string `json:"os_axis"`
	OdAdd             *string `json:"od_add"`
	OsAdd             *string `json:"os_add"`
	OdDva20           *string `json:"od_dva_20"`
	OsDva20           *string `json:"os_dva_20"`
	OdNva20           *string `json:"od_nva_20"`
	OsNva20           *string `json:"os_nva_20"`
	Trial             *bool   `json:"trial"`
	Final             *bool   `json:"final"`
	NeedToOrder       *bool   `json:"need_to_order"`
	DispenseFromStock *bool   `json:"dispense_from_stock"`
	FrontDeskNote     *string `json:"front_desk_note"`
	ExpireDate        *string `json:"expire_date"`
	TypeAdd           *string `json:"type_add"`
}

type LabDesignInput struct {
	OdColor       *string `json:"od_color"`
	OsColor       *string `json:"os_color"`
	OdK1          *string `json:"od_k1"`
	OsK1          *string `json:"os_k1"`
	OdK2          *string `json:"od_k2"`
	OsK2          *string `json:"os_k2"`
	OdSph         *string `json:"od_sph"`
	OsSph         *string `json:"os_sph"`
	OdCyl         *string `json:"od_cyl"`
	OsCyl         *string `json:"os_cyl"`
	OdAxis        *string `json:"od_axis"`
	OsAxis        *string `json:"os_axis"`
	OdAdd         *string `json:"od_add"`
	OsAdd         *string `json:"os_add"`
	OdOverallDia  *string `json:"od_overall_dia"`
	OsOverallDia  *string `json:"os_overall_dia"`
	OdDva         *string `json:"od_dva"`
	OsDva         *string `json:"os_dva"`
	OdNva         *string `json:"od_nva"`
	OsNva         *string `json:"os_nva"`
	FrontDeskNote *string `json:"front_desk_note"`
}

type DrDesignInput struct {
	OdMaterial    *string `json:"od_material"`
	OsMaterial    *string `json:"os_material"`
	OdColor       *string `json:"od_color"`
	OsColor       *string `json:"os_color"`
	OdBaseCurve   *string `json:"od_base_curve"`
	OsBaseCurve   *string `json:"os_base_curve"`
	OdDia         *string `json:"od_dia"`
	OsDia         *string `json:"os_dia"`
	OdPower       *string `json:"od_power"`
	OsPower       *string `json:"os_power"`
	OdCyl         *string `json:"od_cyl"`
	OsCyl         *string `json:"os_cyl"`
	OdAxis        *string `json:"od_axis"`
	OsAxis        *string `json:"os_axis"`
	OdAdd         *string `json:"od_add"`
	OsAdd         *string `json:"os_add"`
	OdCtrThk      *string `json:"od_ctr_thk"`
	OsCtrThk      *string `json:"os_ctr_thk"`
	OdPerfCurve   *string `json:"od_perf_curve"`
	OsPerfCurve   *string `json:"os_perf_curve"`
	OdLentic      *bool   `json:"od_lentic"`
	OsLentic      *bool   `json:"os_lentic"`
	OdDot         *bool   `json:"od_dot"`
	OsDot         *bool   `json:"os_dot"`
	OdDva         *string `json:"od_dva"`
	OsDva         *string `json:"os_dva"`
	OdNva         *string `json:"od_nva"`
	OsNva         *string `json:"os_nva"`
	FrontDeskNote *string `json:"front_desk_note"`
}

type GasPermeableInput struct {
	LabDesign *LabDesignInput `json:"lab_design"`
	DrDesign  *DrDesignInput  `json:"dr_design"`
}

type SaveClFittingInput struct {
	Fitting1     FittingInput      `json:"fitting_1"`
	Fitting2     FittingInput      `json:"fitting_2"`
	Fitting3     FittingInput      `json:"fitting_3"`
	FirstTrial   TrialInput        `json:"first_trial"`
	SecondTrial  TrialInput        `json:"second_trial"`
	ThirdTrial   TrialInput        `json:"third_trial"`
	GasPermeable GasPermeableInput `json:"gas_permeable"`
	DrNote       *string           `json:"dr_note"`
}

type UpdateClFittingInput struct {
	Fitting1     *FittingInput      `json:"fitting_1"`
	Fitting2     *FittingInput      `json:"fitting_2"`
	Fitting3     *FittingInput      `json:"fitting_3"`
	FirstTrial   *TrialInput        `json:"first_trial"`
	SecondTrial  *TrialInput        `json:"second_trial"`
	ThirdTrial   *TrialInput        `json:"third_trial"`
	GasPermeable *GasPermeableInput `json:"gas_permeable"`
	DrNote       *string            `json:"dr_note"`
}

// ---------- helpers ----------

func boolVal(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
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

func computeExpireDate(odBrand, osBrand *string, expireDateStr *string) *time.Time {
	hasBrand := (odBrand != nil && *odBrand != "") || (osBrand != nil && *osBrand != "")
	if !hasBrand {
		return nil
	}
	if expireDateStr == nil || *expireDateStr == "" {
		t := time.Now().UTC().AddDate(1, 0, 0)
		return &t
	}
	t, err := time.Parse("2006-01-02", *expireDateStr)
	if err != nil {
		t = time.Now().UTC().AddDate(1, 0, 0)
	}
	return &t
}

func fittingFromInput(in FittingInput) clModel.Fitting1 {
	de := "n/a"
	if in.DominantEye != nil {
		de = *in.DominantEye
	}
	return clModel.Fitting1{
		OdBrand: in.OdBrand, OsBrand: in.OsBrand,
		OdBCur: in.OdBCur, OsBCur: in.OsBCur,
		OdDia: in.OdDia, OsDia: in.OsDia,
		OdPwr: in.OdPwr, OsPwr: in.OsPwr,
		OdCyl: in.OdCyl, OsCyl: in.OsCyl,
		OdAxis: in.OdAxis, OsAxis: in.OsAxis,
		OdAdd: in.OdAdd, OsAdd: in.OsAdd,
		OdDva20: in.OdDva20, OsDva20: in.OsDva20,
		OdNva20: in.OdNva20, OsNva20: in.OsNva20,
		OdOverRefraction: in.OdOverRefraction, OsOverRefraction: in.OsOverRefraction,
		OdFinal: in.OdFinal, OsFinal: in.OsFinal,
		Evaluation:  in.Evaluation,
		DominantEye: de,
	}
}

// ---------- SaveClFitting ----------

func (s *Service) SaveClFitting(username string, examID int64, input SaveClFittingInput) (map[string]interface{}, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}

	var exam visionModel.EyeExam
	if err := s.db.Where("id_eye_exam = ?", examID).First(&exam).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return nil, errors.New("not authorized to update this exam")
	}
	if exam.Passed {
		return nil, errors.New("cannot update a completed exam")
	}

	// Fitting1
	f1 := fittingFromInput(input.Fitting1)
	if err := s.db.Create(&f1).Error; err != nil {
		return nil, err
	}

	// Fitting2
	f2tmp := fittingFromInput(input.Fitting2)
	f2 := clModel.Fitting2{
		OdBrand: f2tmp.OdBrand, OsBrand: f2tmp.OsBrand,
		OdBCur: f2tmp.OdBCur, OsBCur: f2tmp.OsBCur,
		OdDia: f2tmp.OdDia, OsDia: f2tmp.OsDia,
		OdPwr: f2tmp.OdPwr, OsPwr: f2tmp.OsPwr,
		OdCyl: f2tmp.OdCyl, OsCyl: f2tmp.OsCyl,
		OdAxis: f2tmp.OdAxis, OsAxis: f2tmp.OsAxis,
		OdAdd: f2tmp.OdAdd, OsAdd: f2tmp.OsAdd,
		OdDva20: f2tmp.OdDva20, OsDva20: f2tmp.OsDva20,
		OdNva20: f2tmp.OdNva20, OsNva20: f2tmp.OsNva20,
		OdOverRefraction: f2tmp.OdOverRefraction, OsOverRefraction: f2tmp.OsOverRefraction,
		OdFinal: f2tmp.OdFinal, OsFinal: f2tmp.OsFinal,
		Evaluation:  f2tmp.Evaluation,
		DominantEye: f2tmp.DominantEye,
	}
	if err := s.db.Create(&f2).Error; err != nil {
		return nil, err
	}

	// Fitting3
	f3tmp := fittingFromInput(input.Fitting3)
	f3 := clModel.Fitting3{
		OdBrand: f3tmp.OdBrand, OsBrand: f3tmp.OsBrand,
		OdBCur: f3tmp.OdBCur, OsBCur: f3tmp.OsBCur,
		OdDia: f3tmp.OdDia, OsDia: f3tmp.OsDia,
		OdPwr: f3tmp.OdPwr, OsPwr: f3tmp.OsPwr,
		OdCyl: f3tmp.OdCyl, OsCyl: f3tmp.OsCyl,
		OdAxis: f3tmp.OdAxis, OsAxis: f3tmp.OsAxis,
		OdAdd: f3tmp.OdAdd, OsAdd: f3tmp.OsAdd,
		OdDva20: f3tmp.OdDva20, OsDva20: f3tmp.OsDva20,
		OdNva20: f3tmp.OdNva20, OsNva20: f3tmp.OsNva20,
		OdOverRefraction: f3tmp.OdOverRefraction, OsOverRefraction: f3tmp.OsOverRefraction,
		OdFinal: f3tmp.OdFinal, OsFinal: f3tmp.OsFinal,
		Evaluation:  f3tmp.Evaluation,
		DominantEye: f3tmp.DominantEye,
	}
	if err := s.db.Create(&f3).Error; err != nil {
		return nil, err
	}

	// FirstTrial — expire_date logic
	expireDate := computeExpireDate(input.FirstTrial.OdBrand, input.FirstTrial.OsBrand, input.FirstTrial.ExpireDate)
	ft := clModel.FirstTrial{
		OdBrand: input.FirstTrial.OdBrand, OsBrand: input.FirstTrial.OsBrand,
		OdBCur: input.FirstTrial.OdBCur, OsBCur: input.FirstTrial.OsBCur,
		OdDia: input.FirstTrial.OdDia, OsDia: input.FirstTrial.OsDia,
		OdPwr: input.FirstTrial.OdPwr, OsPwr: input.FirstTrial.OsPwr,
		OdCyl: input.FirstTrial.OdCyl, OsCyl: input.FirstTrial.OsCyl,
		OdAxis: input.FirstTrial.OdAxis, OsAxis: input.FirstTrial.OsAxis,
		OdAdd: input.FirstTrial.OdAdd, OsAdd: input.FirstTrial.OsAdd,
		OdDva20: input.FirstTrial.OdDva20, OsDva20: input.FirstTrial.OsDva20,
		OdNva20: input.FirstTrial.OdNva20, OsNva20: input.FirstTrial.OsNva20,
		Trial:             boolVal(input.FirstTrial.Trial),
		Final:             boolVal(input.FirstTrial.Final),
		NeedToOrder:       boolVal(input.FirstTrial.NeedToOrder),
		DispenseFromStock: boolVal(input.FirstTrial.DispenseFromStock),
		FrontDeskNote:     input.FirstTrial.FrontDeskNote,
		ExpireDate:        expireDate,
	}
	if err := s.db.Create(&ft).Error; err != nil {
		return nil, err
	}

	// SecondTrial
	st := clModel.SecondTrial{
		OdBrand: input.SecondTrial.OdBrand, OsBrand: input.SecondTrial.OsBrand,
		OdBCur: input.SecondTrial.OdBCur, OsBCur: input.SecondTrial.OsBCur,
		OdDia: input.SecondTrial.OdDia, OsDia: input.SecondTrial.OsDia,
		OdPwr: input.SecondTrial.OdPwr, OsPwr: input.SecondTrial.OsPwr,
		OdCyl: input.SecondTrial.OdCyl, OsCyl: input.SecondTrial.OsCyl,
		OdAxis: input.SecondTrial.OdAxis, OsAxis: input.SecondTrial.OsAxis,
		OdAdd: input.SecondTrial.OdAdd, OsAdd: input.SecondTrial.OsAdd,
		OdDva20: input.SecondTrial.OdDva20, OsDva20: input.SecondTrial.OsDva20,
		OdNva20: input.SecondTrial.OdNva20, OsNva20: input.SecondTrial.OsNva20,
		Trial:             boolVal(input.SecondTrial.Trial),
		Final:             boolVal(input.SecondTrial.Final),
		NeedToOrder:       boolVal(input.SecondTrial.NeedToOrder),
		DispenseFromStock: boolVal(input.SecondTrial.DispenseFromStock),
		FrontDeskNote:     input.SecondTrial.FrontDeskNote,
		TypeAdd:           input.SecondTrial.TypeAdd,
	}
	if err := s.db.Create(&st).Error; err != nil {
		return nil, err
	}

	// ThirdTrial
	tt := clModel.ThirdTrial{
		OdBrand: input.ThirdTrial.OdBrand, OsBrand: input.ThirdTrial.OsBrand,
		OdBCur: input.ThirdTrial.OdBCur, OsBCur: input.ThirdTrial.OsBCur,
		OdDia: input.ThirdTrial.OdDia, OsDia: input.ThirdTrial.OsDia,
		OdPwr: input.ThirdTrial.OdPwr, OsPwr: input.ThirdTrial.OsPwr,
		OdCyl: input.ThirdTrial.OdCyl, OsCyl: input.ThirdTrial.OsCyl,
		OdAxis: input.ThirdTrial.OdAxis, OsAxis: input.ThirdTrial.OsAxis,
		OdAdd: input.ThirdTrial.OdAdd, OsAdd: input.ThirdTrial.OsAdd,
		OdDva20: input.ThirdTrial.OdDva20, OsDva20: input.ThirdTrial.OsDva20,
		OdNva20: input.ThirdTrial.OdNva20, OsNva20: input.ThirdTrial.OsNva20,
		Trial:             boolVal(input.ThirdTrial.Trial),
		Final:             boolVal(input.ThirdTrial.Final),
		NeedToOrder:       boolVal(input.ThirdTrial.NeedToOrder),
		DispenseFromStock: boolVal(input.ThirdTrial.DispenseFromStock),
		FrontDeskNote:     input.ThirdTrial.FrontDeskNote,
		TypeAdd:           input.ThirdTrial.TypeAdd,
	}
	if err := s.db.Create(&tt).Error; err != nil {
		return nil, err
	}

	// LabDesign
	var ld clModel.LabDesign
	if input.GasPermeable.LabDesign != nil {
		li := input.GasPermeable.LabDesign
		ld = clModel.LabDesign{
			ColorOd: li.OdColor, ColorOs: li.OsColor,
			K1Od: li.OdK1, K1Os: li.OsK1,
			K2Od: li.OdK2, K2Os: li.OsK2,
			SphOd: li.OdSph, SphOs: li.OsSph,
			CylOd: li.OdCyl, CylOs: li.OsCyl,
			AxisOd: li.OdAxis, AxisOs: li.OsAxis,
			AddOd: li.OdAdd, AddOs: li.OsAdd,
			OverallDiaOd: li.OdOverallDia, OverallDiaOs: li.OsOverallDia,
			DvaOd: li.OdDva, DvaOs: li.OsDva,
			NvaOd: li.OdNva, NvaOs: li.OsNva,
			FrontDeskNote: li.FrontDeskNote,
		}
	}
	if err := s.db.Create(&ld).Error; err != nil {
		return nil, err
	}

	// DrDesign
	var dd clModel.DrDesign
	if input.GasPermeable.DrDesign != nil {
		di := input.GasPermeable.DrDesign
		dd = clModel.DrDesign{
			MaterialOd: di.OdMaterial, MaterialOs: di.OsMaterial,
			ColorOd: di.OdColor, ColorOs: di.OsColor,
			BaseCurveOd: di.OdBaseCurve, BaseCurveOs: di.OsBaseCurve,
			DiaOd: di.OdDia, DiaOs: di.OsDia,
			PowerOd: di.OdPower, PowerOs: di.OsPower,
			CylOd: di.OdCyl, CylOs: di.OsCyl,
			AxisOd: di.OdAxis, AxisOs: di.OsAxis,
			AddOd: di.OdAdd, AddOs: di.OsAdd,
			CtrThkOd: di.OdCtrThk, CtrThkOs: di.OsCtrThk,
			PerfCurveOd: di.OdPerfCurve, PerfCurveOs: di.OsPerfCurve,
			LenticOd: boolVal(di.OdLentic), LenticOs: boolVal(di.OsLentic),
			DotOd: boolVal(di.OdDot), DotOs: boolVal(di.OsDot),
			DvaOd: di.OdDva, DvaOs: di.OsDva,
			NvaOd: di.OdNva, NvaOs: di.OsNva,
			FrontDeskNote: di.FrontDeskNote,
		}
	}
	if err := s.db.Create(&dd).Error; err != nil {
		return nil, err
	}

	// GasPermeable
	ldID := ld.IDLabDesign
	ddID := dd.IDDrDesign
	gp := clModel.GasPermeable{LabDesignID: &ldID, DrDesignID: &ddID}
	if err := s.db.Create(&gp).Error; err != nil {
		return nil, err
	}

	// ClFitting
	f2id := f2.IDFitting2
	f3id := f3.IDFitting3
	stid := st.IDSecondTrial
	ttid := tt.IDThirdTrial
	gpid := gp.IDGasPermeable
	cf := clModel.ClFitting{
		Fitting1ID:     f1.IDFitting1,
		Fitting2ID:     &f2id,
		Fitting3ID:     &f3id,
		FirstTrialID:   ft.IDFirstTrial,
		SecondTrialID:  &stid,
		ThirdTrialID:   &ttid,
		GasPermeableID: &gpid,
		EyeExamID:      examID,
		DrNote:         input.DrNote,
	}
	if err := s.db.Create(&cf).Error; err != nil {
		return nil, err
	}

	activitylog.Log(s.db, "exam_cl_fitting", "save", activitylog.WithEntity(examID))

	return cf.ToMap(), nil
}

// ---------- GetClFitting ----------

func (s *Service) GetClFitting(examID int64) (map[string]interface{}, error) {
	var cf clModel.ClFitting
	if err := s.db.Where("eye_exam_id = ?", examID).First(&cf).Error; err != nil {
		return map[string]interface{}{
			"exam_id":    examID,
			"exists":     false,
			"cl_fitting": nil,
			"fitting_1":  nil,
			"first_trial": nil,
		}, nil
	}

	result := map[string]interface{}{
		"exam_id":    examID,
		"exists":     true,
		"cl_fitting": cf.ToMap(),
		"dr_note":    cf.DrNote,
	}

	// Fitting1
	if cf.Fitting1ID != 0 {
		var f1 clModel.Fitting1
		if s.db.First(&f1, cf.Fitting1ID).Error == nil {
			result["fitting_1"] = f1.ToMap()
		}
	}
	// Fitting2
	if cf.Fitting2ID != nil {
		var f2 clModel.Fitting2
		if s.db.First(&f2, *cf.Fitting2ID).Error == nil {
			result["fitting_2"] = f2.ToMap()
		}
	}
	// Fitting3
	if cf.Fitting3ID != nil {
		var f3 clModel.Fitting3
		if s.db.First(&f3, *cf.Fitting3ID).Error == nil {
			result["fitting_3"] = f3.ToMap()
		}
	}
	// FirstTrial
	if cf.FirstTrialID != 0 {
		var ft clModel.FirstTrial
		if s.db.First(&ft, cf.FirstTrialID).Error == nil {
			result["first_trial"] = ft.ToMap()
		}
	}
	// SecondTrial
	if cf.SecondTrialID != nil {
		var st clModel.SecondTrial
		if s.db.First(&st, *cf.SecondTrialID).Error == nil {
			result["second_trial"] = st.ToMap()
		}
	}
	// ThirdTrial
	if cf.ThirdTrialID != nil {
		var tt clModel.ThirdTrial
		if s.db.First(&tt, *cf.ThirdTrialID).Error == nil {
			result["third_trial"] = tt.ToMap()
		}
	}
	// GasPermeable
	if cf.GasPermeableID != nil {
		var gp clModel.GasPermeable
		if s.db.First(&gp, *cf.GasPermeableID).Error == nil {
			gpMap := map[string]interface{}{
				"id_gas_permeable": gp.IDGasPermeable,
				"lab_design":       nil,
				"dr_design":        nil,
			}
			if gp.LabDesignID != nil {
				var ld clModel.LabDesign
				if s.db.First(&ld, *gp.LabDesignID).Error == nil {
					gpMap["lab_design"] = ld.ToMap()
				}
			}
			if gp.DrDesignID != nil {
				var dd clModel.DrDesign
				if s.db.First(&dd, *gp.DrDesignID).Error == nil {
					gpMap["dr_design"] = dd.ToMap()
				}
			}
			result["gas_permeable"] = gpMap
		}
	}

	return result, nil
}

// ---------- UpdateClFitting ----------

func (s *Service) UpdateClFitting(username string, examID int64, input UpdateClFittingInput) (map[string]interface{}, error) {
	emp, err := s.getEmployee(username)
	if err != nil {
		return nil, err
	}

	var exam visionModel.EyeExam
	if err := s.db.Where("id_eye_exam = ?", examID).First(&exam).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	if exam.Passed {
		return nil, errors.New("cannot update a completed exam")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return nil, errors.New("not authorized to update this exam")
	}

	var cf clModel.ClFitting
	if err := s.db.Where("eye_exam_id = ?", examID).First(&cf).Error; err != nil {
		return nil, errors.New("cl_fitting not found")
	}

	// Fitting1
	if input.Fitting1 != nil {
		var f1 clModel.Fitting1
		if s.db.First(&f1, cf.Fitting1ID).Error == nil {
			in := input.Fitting1
			if in.OdBrand != nil { f1.OdBrand = in.OdBrand }
			if in.OsBrand != nil { f1.OsBrand = in.OsBrand }
			if in.OdBCur != nil { f1.OdBCur = in.OdBCur }
			if in.OsBCur != nil { f1.OsBCur = in.OsBCur }
			if in.OdDia != nil { f1.OdDia = in.OdDia }
			if in.OsDia != nil { f1.OsDia = in.OsDia }
			if in.OdPwr != nil { f1.OdPwr = in.OdPwr }
			if in.OsPwr != nil { f1.OsPwr = in.OsPwr }
			if in.OdCyl != nil { f1.OdCyl = in.OdCyl }
			if in.OsCyl != nil { f1.OsCyl = in.OsCyl }
			if in.OdAxis != nil { f1.OdAxis = in.OdAxis }
			if in.OsAxis != nil { f1.OsAxis = in.OsAxis }
			if in.OdAdd != nil { f1.OdAdd = in.OdAdd }
			if in.OsAdd != nil { f1.OsAdd = in.OsAdd }
			if in.OdDva20 != nil { f1.OdDva20 = in.OdDva20 }
			if in.OsDva20 != nil { f1.OsDva20 = in.OsDva20 }
			if in.OdNva20 != nil { f1.OdNva20 = in.OdNva20 }
			if in.OsNva20 != nil { f1.OsNva20 = in.OsNva20 }
			if in.OdOverRefraction != nil { f1.OdOverRefraction = in.OdOverRefraction }
			if in.OsOverRefraction != nil { f1.OsOverRefraction = in.OsOverRefraction }
			if in.OdFinal != nil { f1.OdFinal = in.OdFinal }
			if in.OsFinal != nil { f1.OsFinal = in.OsFinal }
			if in.Evaluation != nil { f1.Evaluation = in.Evaluation }
			if in.DominantEye != nil { f1.DominantEye = *in.DominantEye }
			s.db.Save(&f1)
		}
	}

	// Fitting2
	if input.Fitting2 != nil && cf.Fitting2ID != nil {
		var f2 clModel.Fitting2
		if s.db.First(&f2, *cf.Fitting2ID).Error == nil {
			in := input.Fitting2
			if in.OdBrand != nil { f2.OdBrand = in.OdBrand }
			if in.OsBrand != nil { f2.OsBrand = in.OsBrand }
			if in.OdBCur != nil { f2.OdBCur = in.OdBCur }
			if in.OsBCur != nil { f2.OsBCur = in.OsBCur }
			if in.OdDia != nil { f2.OdDia = in.OdDia }
			if in.OsDia != nil { f2.OsDia = in.OsDia }
			if in.OdPwr != nil { f2.OdPwr = in.OdPwr }
			if in.OsPwr != nil { f2.OsPwr = in.OsPwr }
			if in.OdCyl != nil { f2.OdCyl = in.OdCyl }
			if in.OsCyl != nil { f2.OsCyl = in.OsCyl }
			if in.OdAxis != nil { f2.OdAxis = in.OdAxis }
			if in.OsAxis != nil { f2.OsAxis = in.OsAxis }
			if in.OdAdd != nil { f2.OdAdd = in.OdAdd }
			if in.OsAdd != nil { f2.OsAdd = in.OsAdd }
			if in.OdDva20 != nil { f2.OdDva20 = in.OdDva20 }
			if in.OsDva20 != nil { f2.OsDva20 = in.OsDva20 }
			if in.OdNva20 != nil { f2.OdNva20 = in.OdNva20 }
			if in.OsNva20 != nil { f2.OsNva20 = in.OsNva20 }
			if in.OdOverRefraction != nil { f2.OdOverRefraction = in.OdOverRefraction }
			if in.OsOverRefraction != nil { f2.OsOverRefraction = in.OsOverRefraction }
			if in.OdFinal != nil { f2.OdFinal = in.OdFinal }
			if in.OsFinal != nil { f2.OsFinal = in.OsFinal }
			if in.Evaluation != nil { f2.Evaluation = in.Evaluation }
			if in.DominantEye != nil { f2.DominantEye = *in.DominantEye }
			s.db.Save(&f2)
		}
	}

	// Fitting3
	if input.Fitting3 != nil && cf.Fitting3ID != nil {
		var f3 clModel.Fitting3
		if s.db.First(&f3, *cf.Fitting3ID).Error == nil {
			in := input.Fitting3
			if in.OdBrand != nil { f3.OdBrand = in.OdBrand }
			if in.OsBrand != nil { f3.OsBrand = in.OsBrand }
			if in.OdBCur != nil { f3.OdBCur = in.OdBCur }
			if in.OsBCur != nil { f3.OsBCur = in.OsBCur }
			if in.OdDia != nil { f3.OdDia = in.OdDia }
			if in.OsDia != nil { f3.OsDia = in.OsDia }
			if in.OdPwr != nil { f3.OdPwr = in.OdPwr }
			if in.OsPwr != nil { f3.OsPwr = in.OsPwr }
			if in.OdCyl != nil { f3.OdCyl = in.OdCyl }
			if in.OsCyl != nil { f3.OsCyl = in.OsCyl }
			if in.OdAxis != nil { f3.OdAxis = in.OdAxis }
			if in.OsAxis != nil { f3.OsAxis = in.OsAxis }
			if in.OdAdd != nil { f3.OdAdd = in.OdAdd }
			if in.OsAdd != nil { f3.OsAdd = in.OsAdd }
			if in.OdDva20 != nil { f3.OdDva20 = in.OdDva20 }
			if in.OsDva20 != nil { f3.OsDva20 = in.OsDva20 }
			if in.OdNva20 != nil { f3.OdNva20 = in.OdNva20 }
			if in.OsNva20 != nil { f3.OsNva20 = in.OsNva20 }
			if in.OdOverRefraction != nil { f3.OdOverRefraction = in.OdOverRefraction }
			if in.OsOverRefraction != nil { f3.OsOverRefraction = in.OsOverRefraction }
			if in.OdFinal != nil { f3.OdFinal = in.OdFinal }
			if in.OsFinal != nil { f3.OsFinal = in.OsFinal }
			if in.Evaluation != nil { f3.Evaluation = in.Evaluation }
			if in.DominantEye != nil { f3.DominantEye = *in.DominantEye }
			s.db.Save(&f3)
		}
	}

	// FirstTrial
	if input.FirstTrial != nil {
		var ft clModel.FirstTrial
		if s.db.First(&ft, cf.FirstTrialID).Error == nil {
			in := input.FirstTrial
			odBrand := ft.OdBrand
			osBrand := ft.OsBrand
			if in.OdBrand != nil { odBrand = in.OdBrand }
			if in.OsBrand != nil { osBrand = in.OsBrand }
			expireDate := computeExpireDate(odBrand, osBrand, in.ExpireDate)

			ft.OdBrand = odBrand
			ft.OsBrand = osBrand
			if in.OdBCur != nil { ft.OdBCur = in.OdBCur }
			if in.OsBCur != nil { ft.OsBCur = in.OsBCur }
			if in.OdDia != nil { ft.OdDia = in.OdDia }
			if in.OsDia != nil { ft.OsDia = in.OsDia }
			if in.OdPwr != nil { ft.OdPwr = in.OdPwr }
			if in.OsPwr != nil { ft.OsPwr = in.OsPwr }
			if in.OdCyl != nil { ft.OdCyl = in.OdCyl }
			if in.OsCyl != nil { ft.OsCyl = in.OsCyl }
			if in.OdAxis != nil { ft.OdAxis = in.OdAxis }
			if in.OsAxis != nil { ft.OsAxis = in.OsAxis }
			if in.OdAdd != nil { ft.OdAdd = in.OdAdd }
			if in.OsAdd != nil { ft.OsAdd = in.OsAdd }
			if in.OdDva20 != nil { ft.OdDva20 = in.OdDva20 }
			if in.OsDva20 != nil { ft.OsDva20 = in.OsDva20 }
			if in.OdNva20 != nil { ft.OdNva20 = in.OdNva20 }
			if in.OsNva20 != nil { ft.OsNva20 = in.OsNva20 }
			if in.Trial != nil { ft.Trial = *in.Trial }
			if in.Final != nil { ft.Final = *in.Final }
			if in.NeedToOrder != nil { ft.NeedToOrder = *in.NeedToOrder }
			if in.DispenseFromStock != nil { ft.DispenseFromStock = *in.DispenseFromStock }
			if in.FrontDeskNote != nil { ft.FrontDeskNote = in.FrontDeskNote }
			ft.ExpireDate = expireDate
			s.db.Save(&ft)
		}
	}

	// SecondTrial
	if input.SecondTrial != nil && cf.SecondTrialID != nil {
		var st clModel.SecondTrial
		if s.db.First(&st, *cf.SecondTrialID).Error == nil {
			in := input.SecondTrial
			if in.OdBrand != nil { st.OdBrand = in.OdBrand }
			if in.OsBrand != nil { st.OsBrand = in.OsBrand }
			if in.OdBCur != nil { st.OdBCur = in.OdBCur }
			if in.OsBCur != nil { st.OsBCur = in.OsBCur }
			if in.OdDia != nil { st.OdDia = in.OdDia }
			if in.OsDia != nil { st.OsDia = in.OsDia }
			if in.OdPwr != nil { st.OdPwr = in.OdPwr }
			if in.OsPwr != nil { st.OsPwr = in.OsPwr }
			if in.OdCyl != nil { st.OdCyl = in.OdCyl }
			if in.OsCyl != nil { st.OsCyl = in.OsCyl }
			if in.OdAxis != nil { st.OdAxis = in.OdAxis }
			if in.OsAxis != nil { st.OsAxis = in.OsAxis }
			if in.OdAdd != nil { st.OdAdd = in.OdAdd }
			if in.OsAdd != nil { st.OsAdd = in.OsAdd }
			if in.OdDva20 != nil { st.OdDva20 = in.OdDva20 }
			if in.OsDva20 != nil { st.OsDva20 = in.OsDva20 }
			if in.OdNva20 != nil { st.OdNva20 = in.OdNva20 }
			if in.OsNva20 != nil { st.OsNva20 = in.OsNva20 }
			if in.Trial != nil { st.Trial = *in.Trial }
			if in.Final != nil { st.Final = *in.Final }
			if in.NeedToOrder != nil { st.NeedToOrder = *in.NeedToOrder }
			if in.DispenseFromStock != nil { st.DispenseFromStock = *in.DispenseFromStock }
			if in.FrontDeskNote != nil { st.FrontDeskNote = in.FrontDeskNote }
			if in.TypeAdd != nil { st.TypeAdd = in.TypeAdd }
			s.db.Save(&st)
		}
	}

	// ThirdTrial
	if input.ThirdTrial != nil && cf.ThirdTrialID != nil {
		var tt clModel.ThirdTrial
		if s.db.First(&tt, *cf.ThirdTrialID).Error == nil {
			in := input.ThirdTrial
			if in.OdBrand != nil { tt.OdBrand = in.OdBrand }
			if in.OsBrand != nil { tt.OsBrand = in.OsBrand }
			if in.OdBCur != nil { tt.OdBCur = in.OdBCur }
			if in.OsBCur != nil { tt.OsBCur = in.OsBCur }
			if in.OdDia != nil { tt.OdDia = in.OdDia }
			if in.OsDia != nil { tt.OsDia = in.OsDia }
			if in.OdPwr != nil { tt.OdPwr = in.OdPwr }
			if in.OsPwr != nil { tt.OsPwr = in.OsPwr }
			if in.OdCyl != nil { tt.OdCyl = in.OdCyl }
			if in.OsCyl != nil { tt.OsCyl = in.OsCyl }
			if in.OdAxis != nil { tt.OdAxis = in.OdAxis }
			if in.OsAxis != nil { tt.OsAxis = in.OsAxis }
			if in.OdAdd != nil { tt.OdAdd = in.OdAdd }
			if in.OsAdd != nil { tt.OsAdd = in.OsAdd }
			if in.OdDva20 != nil { tt.OdDva20 = in.OdDva20 }
			if in.OsDva20 != nil { tt.OsDva20 = in.OsDva20 }
			if in.OdNva20 != nil { tt.OdNva20 = in.OdNva20 }
			if in.OsNva20 != nil { tt.OsNva20 = in.OsNva20 }
			if in.Trial != nil { tt.Trial = *in.Trial }
			if in.Final != nil { tt.Final = *in.Final }
			if in.NeedToOrder != nil { tt.NeedToOrder = *in.NeedToOrder }
			if in.DispenseFromStock != nil { tt.DispenseFromStock = *in.DispenseFromStock }
			if in.FrontDeskNote != nil { tt.FrontDeskNote = in.FrontDeskNote }
			if in.TypeAdd != nil { tt.TypeAdd = in.TypeAdd }
			s.db.Save(&tt)
		}
	}

	// GasPermeable
	if input.GasPermeable != nil && cf.GasPermeableID != nil {
		var gp clModel.GasPermeable
		if s.db.First(&gp, *cf.GasPermeableID).Error == nil {
			if input.GasPermeable.LabDesign != nil && gp.LabDesignID != nil {
				var ld clModel.LabDesign
				if s.db.First(&ld, *gp.LabDesignID).Error == nil {
					li := input.GasPermeable.LabDesign
					if li.OdColor != nil { ld.ColorOd = li.OdColor }
					if li.OsColor != nil { ld.ColorOs = li.OsColor }
					if li.OdK1 != nil { ld.K1Od = li.OdK1 }
					if li.OsK1 != nil { ld.K1Os = li.OsK1 }
					if li.OdK2 != nil { ld.K2Od = li.OdK2 }
					if li.OsK2 != nil { ld.K2Os = li.OsK2 }
					if li.OdSph != nil { ld.SphOd = li.OdSph }
					if li.OsSph != nil { ld.SphOs = li.OsSph }
					if li.OdCyl != nil { ld.CylOd = li.OdCyl }
					if li.OsCyl != nil { ld.CylOs = li.OsCyl }
					if li.OdAxis != nil { ld.AxisOd = li.OdAxis }
					if li.OsAxis != nil { ld.AxisOs = li.OsAxis }
					if li.OdAdd != nil { ld.AddOd = li.OdAdd }
					if li.OsAdd != nil { ld.AddOs = li.OsAdd }
					if li.OdOverallDia != nil { ld.OverallDiaOd = li.OdOverallDia }
					if li.OsOverallDia != nil { ld.OverallDiaOs = li.OsOverallDia }
					if li.OdDva != nil { ld.DvaOd = li.OdDva }
					if li.OsDva != nil { ld.DvaOs = li.OsDva }
					if li.OdNva != nil { ld.NvaOd = li.OdNva }
					if li.OsNva != nil { ld.NvaOs = li.OsNva }
					if li.FrontDeskNote != nil { ld.FrontDeskNote = li.FrontDeskNote }
					s.db.Save(&ld)
				}
			}
			if input.GasPermeable.DrDesign != nil && gp.DrDesignID != nil {
				var dd clModel.DrDesign
				if s.db.First(&dd, *gp.DrDesignID).Error == nil {
					di := input.GasPermeable.DrDesign
					if di.OdMaterial != nil { dd.MaterialOd = di.OdMaterial }
					if di.OsMaterial != nil { dd.MaterialOs = di.OsMaterial }
					if di.OdColor != nil { dd.ColorOd = di.OdColor }
					if di.OsColor != nil { dd.ColorOs = di.OsColor }
					if di.OdBaseCurve != nil { dd.BaseCurveOd = di.OdBaseCurve }
					if di.OsBaseCurve != nil { dd.BaseCurveOs = di.OsBaseCurve }
					if di.OdDia != nil { dd.DiaOd = di.OdDia }
					if di.OsDia != nil { dd.DiaOs = di.OsDia }
					if di.OdPower != nil { dd.PowerOd = di.OdPower }
					if di.OsPower != nil { dd.PowerOs = di.OsPower }
					if di.OdCyl != nil { dd.CylOd = di.OdCyl }
					if di.OsCyl != nil { dd.CylOs = di.OsCyl }
					if di.OdAxis != nil { dd.AxisOd = di.OdAxis }
					if di.OsAxis != nil { dd.AxisOs = di.OsAxis }
					if di.OdAdd != nil { dd.AddOd = di.OdAdd }
					if di.OsAdd != nil { dd.AddOs = di.OsAdd }
					if di.OdCtrThk != nil { dd.CtrThkOd = di.OdCtrThk }
					if di.OsCtrThk != nil { dd.CtrThkOs = di.OsCtrThk }
					if di.OdPerfCurve != nil { dd.PerfCurveOd = di.OdPerfCurve }
					if di.OsPerfCurve != nil { dd.PerfCurveOs = di.OsPerfCurve }
					if di.OdLentic != nil { dd.LenticOd = *di.OdLentic }
					if di.OsLentic != nil { dd.LenticOs = *di.OsLentic }
					if di.OdDot != nil { dd.DotOd = *di.OdDot }
					if di.OsDot != nil { dd.DotOs = *di.OsDot }
					if di.OdDva != nil { dd.DvaOd = di.OdDva }
					if di.OsDva != nil { dd.DvaOs = di.OsDva }
					if di.OdNva != nil { dd.NvaOd = di.OdNva }
					if di.OsNva != nil { dd.NvaOs = di.OsNva }
					if di.FrontDeskNote != nil { dd.FrontDeskNote = di.FrontDeskNote }
					s.db.Save(&dd)
				}
			}
		}
	}

	// ClFitting dr_note
	if input.DrNote != nil {
		cf.DrNote = input.DrNote
		s.db.Save(&cf)
	}

	activitylog.Log(s.db, "exam_cl_fitting", "update", activitylog.WithEntity(examID))

	return cf.ToMap(), nil
}

// ---------- GetContactLensBrands ----------

func (s *Service) GetContactLensBrands() ([]map[string]interface{}, error) {
	var brands []vendorModel.BrandContactLens
	if err := s.db.Select("id_brand_contact_lens, brand_name").Find(&brands).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, len(brands))
	for i, b := range brands {
		result[i] = map[string]interface{}{
			"id_brand_contact_lens": b.IDBrandContactLens,
			"brand_name":            b.BrandName,
		}
	}
	return result, nil
}
