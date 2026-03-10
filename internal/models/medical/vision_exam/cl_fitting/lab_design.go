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
		"id_lab_design": l.IDLabDesign,
		"color_od": l.ColorOd, "color_os": l.ColorOs,
		"k1_od": l.K1Od, "k1_os": l.K1Os, "k2_od": l.K2Od, "k2_os": l.K2Os,
		"sph_od": l.SphOd, "sph_os": l.SphOs, "cyl_od": l.CylOd, "cyl_os": l.CylOs,
		"axis_od": l.AxisOd, "axis_os": l.AxisOs, "add_od": l.AddOd, "add_os": l.AddOs,
		"overall_dia_od": l.OverallDiaOd, "overall_dia_os": l.OverallDiaOs,
		"dva_od": l.DvaOd, "dva_os": l.DvaOs, "nva_od": l.NvaOd, "nva_os": l.NvaOs,
		"front_desk_note": l.FrontDeskNote,
	}
}
