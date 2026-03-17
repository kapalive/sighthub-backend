package cl_fitting

// DrDesign ↔ table: dr_design
type DrDesign struct {
	IDDrDesign    int64   `gorm:"column:id_dr_design;primaryKey;autoIncrement"   json:"id_dr_design"`
	MaterialOd    *string `gorm:"column:material_od;type:varchar(255)"           json:"material_od,omitempty"`
	MaterialOs    *string `gorm:"column:material_os;type:varchar(255)"           json:"material_os,omitempty"`
	ColorOd       *string `gorm:"column:color_od;type:varchar(255)"              json:"color_od,omitempty"`
	ColorOs       *string `gorm:"column:color_os;type:varchar(255)"              json:"color_os,omitempty"`
	BaseCurveOd   *string `gorm:"column:base_curve_od;type:varchar(255)"         json:"base_curve_od,omitempty"`
	BaseCurveOs   *string `gorm:"column:base_curve_os;type:varchar(255)"         json:"base_curve_os,omitempty"`
	DiaOd         *string `gorm:"column:dia_od;type:varchar(255)"                json:"dia_od,omitempty"`
	DiaOs         *string `gorm:"column:dia_os;type:varchar(255)"                json:"dia_os,omitempty"`
	PowerOd       *string `gorm:"column:power_od;type:varchar(255)"              json:"power_od,omitempty"`
	PowerOs       *string `gorm:"column:power_os;type:varchar(255)"              json:"power_os,omitempty"`
	CylOd         *string `gorm:"column:cyl_od;type:varchar(255)"                json:"cyl_od,omitempty"`
	CylOs         *string `gorm:"column:cyl_os;type:varchar(255)"                json:"cyl_os,omitempty"`
	AxisOd        *string `gorm:"column:axis_od;type:varchar(255)"               json:"axis_od,omitempty"`
	AxisOs        *string `gorm:"column:axis_os;type:varchar(255)"               json:"axis_os,omitempty"`
	AddOd         *string `gorm:"column:add_od;type:varchar(255)"                json:"add_od,omitempty"`
	AddOs         *string `gorm:"column:add_os;type:varchar(255)"                json:"add_os,omitempty"`
	CtrThkOd      *string `gorm:"column:ctr_thk_od;type:varchar(255)"            json:"ctr_thk_od,omitempty"`
	CtrThkOs      *string `gorm:"column:ctr_thk_os;type:varchar(255)"            json:"ctr_thk_os,omitempty"`
	PerfCurveOd   *string `gorm:"column:perf_curve_od;type:varchar(255)"         json:"perf_curve_od,omitempty"`
	PerfCurveOs   *string `gorm:"column:perf_curve_os;type:varchar(255)"         json:"perf_curve_os,omitempty"`
	LenticOd      bool    `gorm:"column:lentic_od;not null;default:false"        json:"lentic_od"`
	LenticOs      bool    `gorm:"column:lentic_os;not null;default:false"        json:"lentic_os"`
	DotOd         bool    `gorm:"column:dot_od;not null;default:false"           json:"dot_od"`
	DotOs         bool    `gorm:"column:dot_os;not null;default:false"           json:"dot_os"`
	DvaOd         *string `gorm:"column:dva_od;type:varchar(255)"                json:"dva_od,omitempty"`
	DvaOs         *string `gorm:"column:dva_os;type:varchar(255)"                json:"dva_os,omitempty"`
	NvaOd         *string `gorm:"column:nva_od;type:varchar(255)"                json:"nva_od,omitempty"`
	NvaOs         *string `gorm:"column:nva_os;type:varchar(255)"                json:"nva_os,omitempty"`
	FrontDeskNote *string `gorm:"column:front_desk_note;type:text"               json:"front_desk_note,omitempty"`
}
func (DrDesign) TableName() string { return "dr_design" }
func (d *DrDesign) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_dr_design":  d.IDDrDesign,
		"od_material":   d.MaterialOd,
		"os_material":   d.MaterialOs,
		"od_color":      d.ColorOd,
		"os_color":      d.ColorOs,
		"od_base_curve": d.BaseCurveOd,
		"os_base_curve": d.BaseCurveOs,
		"od_dia":        d.DiaOd,
		"os_dia":        d.DiaOs,
		"od_power":      d.PowerOd,
		"os_power":      d.PowerOs,
		"od_cyl":        d.CylOd,
		"os_cyl":        d.CylOs,
		"od_axis":       d.AxisOd,
		"os_axis":       d.AxisOs,
		"od_add":        d.AddOd,
		"os_add":        d.AddOs,
		"od_ctr_thk":    d.CtrThkOd,
		"os_ctr_thk":    d.CtrThkOs,
		"od_perf_curve": d.PerfCurveOd,
		"os_perf_curve": d.PerfCurveOs,
		"od_lentic":     d.LenticOd,
		"os_lentic":     d.LenticOs,
		"od_dot":        d.DotOd,
		"os_dot":        d.DotOs,
		"od_dva":        d.DvaOd,
		"os_dva":        d.DvaOs,
		"od_nva":        d.NvaOd,
		"os_nva":        d.NvaOs,
		"front_desk_note": d.FrontDeskNote,
	}
}
