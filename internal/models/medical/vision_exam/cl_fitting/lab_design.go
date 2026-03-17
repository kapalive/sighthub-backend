package cl_fitting

// LabDesign ↔ table: lab_design
type LabDesign struct {
	IDLabDesign   int64   `gorm:"column:id_lab_design;primaryKey;autoIncrement"  json:"id_lab_design"`
	ColorOd       *string `gorm:"column:color_od;type:varchar(255)"              json:"color_od,omitempty"`
	ColorOs       *string `gorm:"column:color_os;type:varchar(255)"              json:"color_os,omitempty"`
	K1Od          *string `gorm:"column:k1_od;type:varchar(255)"                 json:"k1_od,omitempty"`
	K1Os          *string `gorm:"column:k1_os;type:varchar(255)"                 json:"k1_os,omitempty"`
	K2Od          *string `gorm:"column:k2_od;type:varchar(255)"                 json:"k2_od,omitempty"`
	K2Os          *string `gorm:"column:k2_os;type:varchar(255)"                 json:"k2_os,omitempty"`
	SphOd         *string `gorm:"column:sph_od;type:varchar(255)"                json:"sph_od,omitempty"`
	SphOs         *string `gorm:"column:sph_os;type:varchar(255)"                json:"sph_os,omitempty"`
	CylOd         *string `gorm:"column:cyl_od;type:varchar(255)"                json:"cyl_od,omitempty"`
	CylOs         *string `gorm:"column:cyl_os;type:varchar(255)"                json:"cyl_os,omitempty"`
	AxisOd        *string `gorm:"column:axis_od;type:varchar(255)"               json:"axis_od,omitempty"`
	AxisOs        *string `gorm:"column:axis_os;type:varchar(255)"               json:"axis_os,omitempty"`
	AddOd         *string `gorm:"column:add_od;type:varchar(255)"                json:"add_od,omitempty"`
	AddOs         *string `gorm:"column:add_os;type:varchar(255)"                json:"add_os,omitempty"`
	OverallDiaOd  *string `gorm:"column:overall_dia_od;type:varchar(255)"        json:"overall_dia_od,omitempty"`
	OverallDiaOs  *string `gorm:"column:overall_dia_os;type:varchar(255)"        json:"overall_dia_os,omitempty"`
	DvaOd         *string `gorm:"column:dva_od;type:varchar(255)"                json:"dva_od,omitempty"`
	DvaOs         *string `gorm:"column:dva_os;type:varchar(255)"                json:"dva_os,omitempty"`
	NvaOd         *string `gorm:"column:nva_od;type:varchar(255)"                json:"nva_od,omitempty"`
	NvaOs         *string `gorm:"column:nva_os;type:varchar(255)"                json:"nva_os,omitempty"`
	FrontDeskNote *string `gorm:"column:front_desk_note;type:text"               json:"front_desk_note,omitempty"`
}
func (LabDesign) TableName() string { return "lab_design" }
func (l *LabDesign) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_lab_design":  l.IDLabDesign,
		"od_color":       l.ColorOd,
		"os_color":       l.ColorOs,
		"od_k1":          l.K1Od,
		"os_k1":          l.K1Os,
		"od_k2":          l.K2Od,
		"os_k2":          l.K2Os,
		"od_sph":         l.SphOd,
		"os_sph":         l.SphOs,
		"od_cyl":         l.CylOd,
		"os_cyl":         l.CylOs,
		"od_axis":        l.AxisOd,
		"os_axis":        l.AxisOs,
		"od_add":         l.AddOd,
		"os_add":         l.AddOs,
		"od_overall_dia": l.OverallDiaOd,
		"os_overall_dia": l.OverallDiaOs,
		"od_dva":         l.DvaOd,
		"os_dva":         l.DvaOs,
		"od_nva":         l.NvaOd,
		"os_nva":         l.NvaOs,
		"front_desk_note": l.FrontDeskNote,
	}
}
