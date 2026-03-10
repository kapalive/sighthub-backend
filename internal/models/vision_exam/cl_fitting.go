// internal/models/vision_exam/cl_fitting.go
package vision_exam

import "time"

// Fitting1 ↔ table: fitting_1
type Fitting1 struct {
	IDFitting1         int64   `gorm:"column:id_fitting_1;primaryKey;autoIncrement" json:"id_fitting_1"`
	OdBrand            *string `gorm:"column:od_brand;type:text"                    json:"od_brand,omitempty"`
	OsBrand            *string `gorm:"column:os_brand;type:text"                    json:"os_brand,omitempty"`
	OdBCur             *string `gorm:"column:od_b_cur;type:varchar(255)"             json:"od_b_cur,omitempty"`
	OsBCur             *string `gorm:"column:os_b_cur;type:varchar(255)"             json:"os_b_cur,omitempty"`
	OdDia              *string `gorm:"column:od_dia;type:varchar(255)"               json:"od_dia,omitempty"`
	OsDia              *string `gorm:"column:os_dia;type:varchar(255)"               json:"os_dia,omitempty"`
	OdPwr              *string `gorm:"column:od_pwr;type:varchar(255)"               json:"od_pwr,omitempty"`
	OsPwr              *string `gorm:"column:os_pwr;type:varchar(255)"               json:"os_pwr,omitempty"`
	OdCyl              *string `gorm:"column:od_cyl;type:varchar(255)"               json:"od_cyl,omitempty"`
	OsCyl              *string `gorm:"column:os_cyl;type:varchar(255)"               json:"os_cyl,omitempty"`
	OdAxis             *string `gorm:"column:od_axis;type:varchar(255)"              json:"od_axis,omitempty"`
	OsAxis             *string `gorm:"column:os_axis;type:varchar(255)"              json:"os_axis,omitempty"`
	OdAdd              *string `gorm:"column:od_add;type:varchar(255)"               json:"od_add,omitempty"`
	OsAdd              *string `gorm:"column:os_add;type:varchar(255)"               json:"os_add,omitempty"`
	OdDva20            *string `gorm:"column:od_dva_20;type:varchar(255)"            json:"od_dva_20,omitempty"`
	OsDva20            *string `gorm:"column:os_dva_20;type:varchar(255)"            json:"os_dva_20,omitempty"`
	OdNva20            *string `gorm:"column:od_nva_20;type:varchar(255)"            json:"od_nva_20,omitempty"`
	OsNva20            *string `gorm:"column:os_nva_20;type:varchar(255)"            json:"os_nva_20,omitempty"`
	OdOverRefraction   *string `gorm:"column:od_over_refraction;type:text"           json:"od_over_refraction,omitempty"`
	OsOverRefraction   *string `gorm:"column:os_over_refraction;type:text"           json:"os_over_refraction,omitempty"`
	OdFinal            *bool   `gorm:"column:od_final"                              json:"od_final,omitempty"`
	OsFinal            *bool   `gorm:"column:os_final"                              json:"os_final,omitempty"`
	Evaluation         *string `gorm:"column:evaluation;type:text"                  json:"evaluation,omitempty"`
	DominantEye        string  `gorm:"column:dominant_eye;not null;default:'n/a'"   json:"dominant_eye"`
}
func (Fitting1) TableName() string { return "fitting_1" }
func (f *Fitting1) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_fitting_1": f.IDFitting1, "od_brand": f.OdBrand, "os_brand": f.OsBrand,
		"od_b_cur": f.OdBCur, "os_b_cur": f.OsBCur, "od_dia": f.OdDia, "os_dia": f.OsDia,
		"od_pwr": f.OdPwr, "os_pwr": f.OsPwr, "od_cyl": f.OdCyl, "os_cyl": f.OsCyl,
		"od_axis": f.OdAxis, "os_axis": f.OsAxis, "od_add": f.OdAdd, "os_add": f.OsAdd,
		"od_dva_20": f.OdDva20, "os_dva_20": f.OsDva20, "od_nva_20": f.OdNva20, "os_nva_20": f.OsNva20,
		"od_over_refraction": f.OdOverRefraction, "os_over_refraction": f.OsOverRefraction,
		"od_final": f.OdFinal, "os_final": f.OsFinal, "evaluation": f.Evaluation, "dominant_eye": f.DominantEye,
	}
}

// Fitting2 ↔ table: fitting_2 (same fields as Fitting1)
type Fitting2 struct {
	IDFitting2         int64   `gorm:"column:id_fitting_2;primaryKey;autoIncrement" json:"id_fitting_2"`
	OdBrand            *string `gorm:"column:od_brand;type:text"                    json:"od_brand,omitempty"`
	OsBrand            *string `gorm:"column:os_brand;type:text"                    json:"os_brand,omitempty"`
	OdBCur             *string `gorm:"column:od_b_cur;type:varchar(255)"             json:"od_b_cur,omitempty"`
	OsBCur             *string `gorm:"column:os_b_cur;type:varchar(255)"             json:"os_b_cur,omitempty"`
	OdDia              *string `gorm:"column:od_dia;type:varchar(255)"               json:"od_dia,omitempty"`
	OsDia              *string `gorm:"column:os_dia;type:varchar(255)"               json:"os_dia,omitempty"`
	OdPwr              *string `gorm:"column:od_pwr;type:varchar(255)"               json:"od_pwr,omitempty"`
	OsPwr              *string `gorm:"column:os_pwr;type:varchar(255)"               json:"os_pwr,omitempty"`
	OdCyl              *string `gorm:"column:od_cyl;type:varchar(255)"               json:"od_cyl,omitempty"`
	OsCyl              *string `gorm:"column:os_cyl;type:varchar(255)"               json:"os_cyl,omitempty"`
	OdAxis             *string `gorm:"column:od_axis;type:varchar(255)"              json:"od_axis,omitempty"`
	OsAxis             *string `gorm:"column:os_axis;type:varchar(255)"              json:"os_axis,omitempty"`
	OdAdd              *string `gorm:"column:od_add;type:varchar(255)"               json:"od_add,omitempty"`
	OsAdd              *string `gorm:"column:os_add;type:varchar(255)"               json:"os_add,omitempty"`
	OdDva20            *string `gorm:"column:od_dva_20;type:varchar(255)"            json:"od_dva_20,omitempty"`
	OsDva20            *string `gorm:"column:os_dva_20;type:varchar(255)"            json:"os_dva_20,omitempty"`
	OdNva20            *string `gorm:"column:od_nva_20;type:varchar(255)"            json:"od_nva_20,omitempty"`
	OsNva20            *string `gorm:"column:os_nva_20;type:varchar(255)"            json:"os_nva_20,omitempty"`
	OdOverRefraction   *string `gorm:"column:od_over_refraction;type:text"           json:"od_over_refraction,omitempty"`
	OsOverRefraction   *string `gorm:"column:os_over_refraction;type:text"           json:"os_over_refraction,omitempty"`
	OdFinal            *bool   `gorm:"column:od_final"                              json:"od_final,omitempty"`
	OsFinal            *bool   `gorm:"column:os_final"                              json:"os_final,omitempty"`
	Evaluation         *string `gorm:"column:evaluation;type:text"                  json:"evaluation,omitempty"`
	DominantEye        string  `gorm:"column:dominant_eye;not null;default:'n/a'"   json:"dominant_eye"`
}
func (Fitting2) TableName() string { return "fitting_2" }
func (f *Fitting2) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_fitting_2": f.IDFitting2, "od_brand": f.OdBrand, "os_brand": f.OsBrand,
		"dominant_eye": f.DominantEye,
	}
}

// Fitting3 ↔ table: fitting_3
type Fitting3 struct {
	IDFitting3         int64   `gorm:"column:id_fitting_3;primaryKey;autoIncrement" json:"id_fitting_3"`
	OdBrand            *string `gorm:"column:od_brand;type:text"                    json:"od_brand,omitempty"`
	OsBrand            *string `gorm:"column:os_brand;type:text"                    json:"os_brand,omitempty"`
	OdBCur             *string `gorm:"column:od_b_cur;type:varchar(255)"             json:"od_b_cur,omitempty"`
	OsBCur             *string `gorm:"column:os_b_cur;type:varchar(255)"             json:"os_b_cur,omitempty"`
	OdDia              *string `gorm:"column:od_dia;type:varchar(255)"               json:"od_dia,omitempty"`
	OsDia              *string `gorm:"column:os_dia;type:varchar(255)"               json:"os_dia,omitempty"`
	OdPwr              *string `gorm:"column:od_pwr;type:varchar(255)"               json:"od_pwr,omitempty"`
	OsPwr              *string `gorm:"column:os_pwr;type:varchar(255)"               json:"os_pwr,omitempty"`
	OdCyl              *string `gorm:"column:od_cyl;type:varchar(255)"               json:"od_cyl,omitempty"`
	OsCyl              *string `gorm:"column:os_cyl;type:varchar(255)"               json:"os_cyl,omitempty"`
	OdAxis             *string `gorm:"column:od_axis;type:varchar(255)"              json:"od_axis,omitempty"`
	OsAxis             *string `gorm:"column:os_axis;type:varchar(255)"              json:"os_axis,omitempty"`
	OdAdd              *string `gorm:"column:od_add;type:varchar(255)"               json:"od_add,omitempty"`
	OsAdd              *string `gorm:"column:os_add;type:varchar(255)"               json:"os_add,omitempty"`
	OdDva20            *string `gorm:"column:od_dva_20;type:varchar(255)"            json:"od_dva_20,omitempty"`
	OsDva20            *string `gorm:"column:os_dva_20;type:varchar(255)"            json:"os_dva_20,omitempty"`
	OdNva20            *string `gorm:"column:od_nva_20;type:varchar(255)"            json:"od_nva_20,omitempty"`
	OsNva20            *string `gorm:"column:os_nva_20;type:varchar(255)"            json:"os_nva_20,omitempty"`
	OdOverRefraction   *string `gorm:"column:od_over_refraction;type:text"           json:"od_over_refraction,omitempty"`
	OsOverRefraction   *string `gorm:"column:os_over_refraction;type:text"           json:"os_over_refraction,omitempty"`
	OdFinal            *bool   `gorm:"column:od_final"                              json:"od_final,omitempty"`
	OsFinal            *bool   `gorm:"column:os_final"                              json:"os_final,omitempty"`
	Evaluation         *string `gorm:"column:evaluation;type:text"                  json:"evaluation,omitempty"`
	DominantEye        string  `gorm:"column:dominant_eye;not null;default:'n/a'"   json:"dominant_eye"`
}
func (Fitting3) TableName() string { return "fitting_3" }
func (f *Fitting3) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_fitting_3": f.IDFitting3, "dominant_eye": f.DominantEye}
}

// FirstTrial ↔ table: first_trial
type FirstTrial struct {
	IDFirstTrial         int64      `gorm:"column:id_first_trial;primaryKey;autoIncrement" json:"id_first_trial"`
	OdBrand              *string    `gorm:"column:od_brand;type:varchar(255)"  json:"od_brand,omitempty"`
	OsBrand              *string    `gorm:"column:os_brand;type:varchar(255)"  json:"os_brand,omitempty"`
	OdBCur               *string    `gorm:"column:od_b_cur;type:varchar(255)"  json:"od_b_cur,omitempty"`
	OsBCur               *string    `gorm:"column:os_b_cur;type:varchar(255)"  json:"os_b_cur,omitempty"`
	OdDia                *string    `gorm:"column:od_dia;type:varchar(255)"    json:"od_dia,omitempty"`
	OsDia                *string    `gorm:"column:os_dia;type:varchar(255)"    json:"os_dia,omitempty"`
	OdPwr                *string    `gorm:"column:od_pwr;type:varchar(255)"    json:"od_pwr,omitempty"`
	OsPwr                *string    `gorm:"column:os_pwr;type:varchar(255)"    json:"os_pwr,omitempty"`
	OdCyl                *string    `gorm:"column:od_cyl;type:varchar(255)"    json:"od_cyl,omitempty"`
	OsCyl                *string    `gorm:"column:os_cyl;type:varchar(255)"    json:"os_cyl,omitempty"`
	OdAxis               *string    `gorm:"column:od_axis;type:varchar(255)"   json:"od_axis,omitempty"`
	OsAxis               *string    `gorm:"column:os_axis;type:varchar(255)"   json:"os_axis,omitempty"`
	OdAdd                *string    `gorm:"column:od_add;type:varchar(255)"    json:"od_add,omitempty"`
	OsAdd                *string    `gorm:"column:os_add;type:varchar(255)"    json:"os_add,omitempty"`
	OdDva20              *string    `gorm:"column:od_dva_20;type:varchar(255)" json:"od_dva_20,omitempty"`
	OsDva20              *string    `gorm:"column:os_dva_20;type:varchar(255)" json:"os_dva_20,omitempty"`
	OdNva20              *string    `gorm:"column:od_nva_20;type:varchar(255)" json:"od_nva_20,omitempty"`
	OsNva20              *string    `gorm:"column:os_nva_20;type:varchar(255)" json:"os_nva_20,omitempty"`
	Trial                bool       `gorm:"column:trial;not null;default:false"              json:"trial"`
	Final                bool       `gorm:"column:final;not null;default:false"              json:"final"`
	NeedToOrder          bool       `gorm:"column:need_to_order;not null;default:false"      json:"need_to_order"`
	DispenseFromStock    bool       `gorm:"column:dispense_from_stock;not null;default:false" json:"dispense_from_stock"`
	FrontDeskNote        *string    `gorm:"column:front_desk_note;type:text"                 json:"front_desk_note,omitempty"`
	ExpireDate           *time.Time `gorm:"column:expire_date;type:date"                     json:"expire_date,omitempty"`
}
func (FirstTrial) TableName() string { return "first_trial" }
func (f *FirstTrial) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_first_trial": f.IDFirstTrial,
		"trial": f.Trial, "final": f.Final,
		"need_to_order": f.NeedToOrder, "dispense_from_stock": f.DispenseFromStock,
	}
	if f.ExpireDate != nil { m["expire_date"] = f.ExpireDate.Format("2006-01-02") } else { m["expire_date"] = nil }
	return m
}

// SecondTrial ↔ table: second_trial
type SecondTrial struct {
	IDSecondTrial     int64   `gorm:"column:id_second_trial;primaryKey;autoIncrement" json:"id_second_trial"`
	OdBrand           *string `gorm:"column:od_brand;type:varchar(255)" json:"od_brand,omitempty"`
	OsBrand           *string `gorm:"column:os_brand;type:varchar(255)" json:"os_brand,omitempty"`
	OdBCur            *string `gorm:"column:od_b_cur;type:varchar(255)" json:"od_b_cur,omitempty"`
	OsBCur            *string `gorm:"column:os_b_cur;type:varchar(255)" json:"os_b_cur,omitempty"`
	OdDia             *string `gorm:"column:od_dia;type:varchar(255)"   json:"od_dia,omitempty"`
	OsDia             *string `gorm:"column:os_dia;type:varchar(255)"   json:"os_dia,omitempty"`
	OdPwr             *string `gorm:"column:od_pwr;type:varchar(255)"   json:"od_pwr,omitempty"`
	OsPwr             *string `gorm:"column:os_pwr;type:varchar(255)"   json:"os_pwr,omitempty"`
	OdCyl             *string `gorm:"column:od_cyl;type:varchar(255)"   json:"od_cyl,omitempty"`
	OsCyl             *string `gorm:"column:os_cyl;type:varchar(255)"   json:"os_cyl,omitempty"`
	OdAxis            *string `gorm:"column:od_axis;type:varchar(255)"  json:"od_axis,omitempty"`
	OsAxis            *string `gorm:"column:os_axis;type:varchar(255)"  json:"os_axis,omitempty"`
	OdAdd             *string `gorm:"column:od_add;type:varchar(255)"   json:"od_add,omitempty"`
	OsAdd             *string `gorm:"column:os_add;type:varchar(255)"   json:"os_add,omitempty"`
	Trial             bool    `gorm:"column:trial;not null;default:false"              json:"trial"`
	Final             bool    `gorm:"column:final;not null;default:false"              json:"final"`
	NeedToOrder       bool    `gorm:"column:need_to_order;not null;default:false"      json:"need_to_order"`
	DispenseFromStock bool    `gorm:"column:dispense_from_stock;not null;default:false" json:"dispense_from_stock"`
	FrontDeskNote     *string `gorm:"column:front_desk_note;type:text"                 json:"front_desk_note,omitempty"`
	TypeAdd           *string `gorm:"column:type_add;type:varchar(255)"                json:"type_add,omitempty"`
}
func (SecondTrial) TableName() string { return "second_trial" }
func (s *SecondTrial) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_second_trial": s.IDSecondTrial, "trial": s.Trial, "final": s.Final}
}

// ThirdTrial ↔ table: third_trial
type ThirdTrial struct {
	IDThirdTrial      int64   `gorm:"column:id_third_trial;primaryKey;autoIncrement" json:"id_third_trial"`
	OdBrand           *string `gorm:"column:od_brand;type:varchar(255)" json:"od_brand,omitempty"`
	OsBrand           *string `gorm:"column:os_brand;type:varchar(255)" json:"os_brand,omitempty"`
	OdBCur            *string `gorm:"column:od_b_cur;type:varchar(255)" json:"od_b_cur,omitempty"`
	OsBCur            *string `gorm:"column:os_b_cur;type:varchar(255)" json:"os_b_cur,omitempty"`
	OdDia             *string `gorm:"column:od_dia;type:varchar(255)"   json:"od_dia,omitempty"`
	OsDia             *string `gorm:"column:os_dia;type:varchar(255)"   json:"os_dia,omitempty"`
	OdPwr             *string `gorm:"column:od_pwr;type:varchar(255)"   json:"od_pwr,omitempty"`
	OsPwr             *string `gorm:"column:os_pwr;type:varchar(255)"   json:"os_pwr,omitempty"`
	OdCyl             *string `gorm:"column:od_cyl;type:varchar(255)"   json:"od_cyl,omitempty"`
	OsCyl             *string `gorm:"column:os_cyl;type:varchar(255)"   json:"os_cyl,omitempty"`
	OdAxis            *string `gorm:"column:od_axis;type:varchar(255)"  json:"od_axis,omitempty"`
	OsAxis            *string `gorm:"column:os_axis;type:varchar(255)"  json:"os_axis,omitempty"`
	Trial             bool    `gorm:"column:trial;not null;default:false"              json:"trial"`
	Final             bool    `gorm:"column:final;not null;default:false"              json:"final"`
	NeedToOrder       bool    `gorm:"column:need_to_order;not null;default:false"      json:"need_to_order"`
	DispenseFromStock bool    `gorm:"column:dispense_from_stock;not null;default:false" json:"dispense_from_stock"`
}
func (ThirdTrial) TableName() string { return "third_trial" }
func (t *ThirdTrial) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_third_trial": t.IDThirdTrial, "trial": t.Trial, "final": t.Final}
}

// LabDesign ↔ table: lab_design
type LabDesign struct {
	IDLabDesign    int64   `gorm:"column:id_lab_design;primaryKey;autoIncrement" json:"id_lab_design"`
	ColorOd        *string `gorm:"column:color_od;type:varchar(255)" json:"color_od,omitempty"`
	ColorOs        *string `gorm:"column:color_os;type:varchar(255)" json:"color_os,omitempty"`
	K1Od           *string `gorm:"column:k1_od;type:varchar(255)"    json:"k1_od,omitempty"`
	K1Os           *string `gorm:"column:k1_os;type:varchar(255)"    json:"k1_os,omitempty"`
	K2Od           *string `gorm:"column:k2_od;type:varchar(255)"    json:"k2_od,omitempty"`
	K2Os           *string `gorm:"column:k2_os;type:varchar(255)"    json:"k2_os,omitempty"`
	SphOd          *string `gorm:"column:sph_od;type:varchar(255)"   json:"sph_od,omitempty"`
	SphOs          *string `gorm:"column:sph_os;type:varchar(255)"   json:"sph_os,omitempty"`
	CylOd          *string `gorm:"column:cyl_od;type:varchar(255)"   json:"cyl_od,omitempty"`
	CylOs          *string `gorm:"column:cyl_os;type:varchar(255)"   json:"cyl_os,omitempty"`
	AxisOd         *string `gorm:"column:axis_od;type:varchar(255)"  json:"axis_od,omitempty"`
	AxisOs         *string `gorm:"column:axis_os;type:varchar(255)"  json:"axis_os,omitempty"`
	AddOd          *string `gorm:"column:add_od;type:varchar(255)"   json:"add_od,omitempty"`
	AddOs          *string `gorm:"column:add_os;type:varchar(255)"   json:"add_os,omitempty"`
	OverallDiaOd   *string `gorm:"column:overall_dia_od;type:varchar(255)" json:"overall_dia_od,omitempty"`
	OverallDiaOs   *string `gorm:"column:overall_dia_os;type:varchar(255)" json:"overall_dia_os,omitempty"`
	DvaOd          *string `gorm:"column:dva_od;type:varchar(255)"   json:"dva_od,omitempty"`
	DvaOs          *string `gorm:"column:dva_os;type:varchar(255)"   json:"dva_os,omitempty"`
	NvaOd          *string `gorm:"column:nva_od;type:varchar(255)"   json:"nva_od,omitempty"`
	NvaOs          *string `gorm:"column:nva_os;type:varchar(255)"   json:"nva_os,omitempty"`
	FrontDeskNote  *string `gorm:"column:front_desk_note;type:text"  json:"front_desk_note,omitempty"`
}
func (LabDesign) TableName() string { return "lab_design" }
func (l *LabDesign) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_lab_design": l.IDLabDesign}
}

// DrDesign ↔ table: dr_design
type DrDesign struct {
	IDDrDesign     int64   `gorm:"column:id_dr_design;primaryKey;autoIncrement" json:"id_dr_design"`
	MaterialOd     *string `gorm:"column:material_od;type:varchar(255)" json:"material_od,omitempty"`
	MaterialOs     *string `gorm:"column:material_os;type:varchar(255)" json:"material_os,omitempty"`
	ColorOd        *string `gorm:"column:color_od;type:varchar(255)"    json:"color_od,omitempty"`
	ColorOs        *string `gorm:"column:color_os;type:varchar(255)"    json:"color_os,omitempty"`
	BaseCurveOd    *string `gorm:"column:base_curve_od;type:varchar(255)" json:"base_curve_od,omitempty"`
	BaseCurveOs    *string `gorm:"column:base_curve_os;type:varchar(255)" json:"base_curve_os,omitempty"`
	DiaOd          *string `gorm:"column:dia_od;type:varchar(255)"      json:"dia_od,omitempty"`
	DiaOs          *string `gorm:"column:dia_os;type:varchar(255)"      json:"dia_os,omitempty"`
	PowerOd        *string `gorm:"column:power_od;type:varchar(255)"    json:"power_od,omitempty"`
	PowerOs        *string `gorm:"column:power_os;type:varchar(255)"    json:"power_os,omitempty"`
	CylOd          *string `gorm:"column:cyl_od;type:varchar(255)"      json:"cyl_od,omitempty"`
	CylOs          *string `gorm:"column:cyl_os;type:varchar(255)"      json:"cyl_os,omitempty"`
	AxisOd         *string `gorm:"column:axis_od;type:varchar(255)"     json:"axis_od,omitempty"`
	AxisOs         *string `gorm:"column:axis_os;type:varchar(255)"     json:"axis_os,omitempty"`
	AddOd          *string `gorm:"column:add_od;type:varchar(255)"      json:"add_od,omitempty"`
	AddOs          *string `gorm:"column:add_os;type:varchar(255)"      json:"add_os,omitempty"`
	CtrThkOd       *string `gorm:"column:ctr_thk_od;type:varchar(255)"  json:"ctr_thk_od,omitempty"`
	CtrThkOs       *string `gorm:"column:ctr_thk_os;type:varchar(255)"  json:"ctr_thk_os,omitempty"`
	PerfCurveOd    *string `gorm:"column:perf_curve_od;type:varchar(255)" json:"perf_curve_od,omitempty"`
	PerfCurveOs    *string `gorm:"column:perf_curve_os;type:varchar(255)" json:"perf_curve_os,omitempty"`
	LenticOd       bool    `gorm:"column:lentic_od;not null;default:false" json:"lentic_od"`
	LenticOs       bool    `gorm:"column:lentic_os;not null;default:false" json:"lentic_os"`
	DotOd          bool    `gorm:"column:dot_od;not null;default:false"    json:"dot_od"`
	DotOs          bool    `gorm:"column:dot_os;not null;default:false"    json:"dot_os"`
	DvaOd          *string `gorm:"column:dva_od;type:varchar(255)"      json:"dva_od,omitempty"`
	DvaOs          *string `gorm:"column:dva_os;type:varchar(255)"      json:"dva_os,omitempty"`
	NvaOd          *string `gorm:"column:nva_od;type:varchar(255)"      json:"nva_od,omitempty"`
	NvaOs          *string `gorm:"column:nva_os;type:varchar(255)"      json:"nva_os,omitempty"`
	FrontDeskNote  *string `gorm:"column:front_desk_note;type:text"     json:"front_desk_note,omitempty"`
}
func (DrDesign) TableName() string { return "dr_design" }
func (d *DrDesign) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_dr_design": d.IDDrDesign}
}

// GasPermeable ↔ table: gas_permeable
type GasPermeable struct {
	IDGasPermeable int64  `gorm:"column:id_gas_permeable;primaryKey;autoIncrement" json:"id_gas_permeable"`
	LabDesignID    *int64 `gorm:"column:lab_design_id"                             json:"lab_design_id,omitempty"`
	DrDesignID     *int64 `gorm:"column:dr_design_id"                              json:"dr_design_id,omitempty"`

	LabDesign *LabDesign `gorm:"foreignKey:LabDesignID;references:IDLabDesign" json:"-"`
	DrDesign  *DrDesign  `gorm:"foreignKey:DrDesignID;references:IDDrDesign"   json:"-"`
}
func (GasPermeable) TableName() string { return "gas_permeable" }
func (g *GasPermeable) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_gas_permeable": g.IDGasPermeable, "lab_design_id": g.LabDesignID, "dr_design_id": g.DrDesignID}
}

// ClFitting ↔ table: cl_fitting
type ClFitting struct {
	IDClFitting    int64  `gorm:"column:id_cl_fitting;primaryKey;autoIncrement" json:"id_cl_fitting"`
	Fitting1ID     int64  `gorm:"column:fitting_1_id;not null;uniqueIndex"      json:"fitting_1_id"`
	Fitting2ID     *int64 `gorm:"column:fitting_2_id;uniqueIndex"               json:"fitting_2_id,omitempty"`
	Fitting3ID     *int64 `gorm:"column:fitting_3_id;uniqueIndex"               json:"fitting_3_id,omitempty"`
	FirstTrialID   int64  `gorm:"column:first_trial_id;not null;uniqueIndex"    json:"first_trial_id"`
	SecondTrialID  *int64 `gorm:"column:second_trial_id;uniqueIndex"            json:"second_trial_id,omitempty"`
	ThirdTrialID   *int64 `gorm:"column:third_trial_id;uniqueIndex"             json:"third_trial_id,omitempty"`
	GasPermeableID *int64 `gorm:"column:gas_permeable_id;uniqueIndex"           json:"gas_permeable_id,omitempty"`
	DrNote         *string `gorm:"column:dr_note;type:text"                     json:"dr_note,omitempty"`
	EyeExamID      int64  `gorm:"column:eye_exam_id;not null"                   json:"eye_exam_id"`

	Fitting1    *Fitting1    `gorm:"foreignKey:Fitting1ID;references:IDFitting1"       json:"-"`
	Fitting2    *Fitting2    `gorm:"foreignKey:Fitting2ID;references:IDFitting2"       json:"-"`
	Fitting3    *Fitting3    `gorm:"foreignKey:Fitting3ID;references:IDFitting3"       json:"-"`
	FirstTrial  *FirstTrial  `gorm:"foreignKey:FirstTrialID;references:IDFirstTrial"   json:"-"`
	SecondTrial *SecondTrial `gorm:"foreignKey:SecondTrialID;references:IDSecondTrial" json:"-"`
	ThirdTrial  *ThirdTrial  `gorm:"foreignKey:ThirdTrialID;references:IDThirdTrial"   json:"-"`
	GasPermeable *GasPermeable `gorm:"foreignKey:GasPermeableID;references:IDGasPermeable" json:"-"`
}
func (ClFitting) TableName() string { return "cl_fitting" }
func (c *ClFitting) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_cl_fitting": c.IDClFitting, "fitting_1_id": c.Fitting1ID,
		"fitting_2_id": c.Fitting2ID, "fitting_3_id": c.Fitting3ID,
		"first_trial_id": c.FirstTrialID, "second_trial_id": c.SecondTrialID,
		"third_trial_id": c.ThirdTrialID, "gas_permeable_id": c.GasPermeableID,
		"dr_note": c.DrNote, "eye_exam_id": c.EyeExamID,
	}
	return m
}
