package cl_fitting

// Fitting3 ↔ table: fitting_3
type Fitting3 struct {
	IDFitting3    int64   `gorm:"column:id_fitting_3;primaryKey;autoIncrement" json:"id_fitting_3"`
	OdBrand          *string `gorm:"column:od_brand;type:text"           json:"od_brand,omitempty"`
	OsBrand          *string `gorm:"column:os_brand;type:text"           json:"os_brand,omitempty"`
	OdBCur           *string `gorm:"column:od_b_cur;type:varchar(255)"   json:"od_b_cur,omitempty"`
	OsBCur           *string `gorm:"column:os_b_cur;type:varchar(255)"   json:"os_b_cur,omitempty"`
	OdDia            *string `gorm:"column:od_dia;type:varchar(255)"     json:"od_dia,omitempty"`
	OsDia            *string `gorm:"column:os_dia;type:varchar(255)"     json:"os_dia,omitempty"`
	OdPwr            *string `gorm:"column:od_pwr;type:varchar(255)"     json:"od_pwr,omitempty"`
	OsPwr            *string `gorm:"column:os_pwr;type:varchar(255)"     json:"os_pwr,omitempty"`
	OdCyl            *string `gorm:"column:od_cyl;type:varchar(255)"     json:"od_cyl,omitempty"`
	OsCyl            *string `gorm:"column:os_cyl;type:varchar(255)"     json:"os_cyl,omitempty"`
	OdAxis           *string `gorm:"column:od_axis;type:varchar(255)"    json:"od_axis,omitempty"`
	OsAxis           *string `gorm:"column:os_axis;type:varchar(255)"    json:"os_axis,omitempty"`
	OdAdd            *string `gorm:"column:od_add;type:varchar(255)"     json:"od_add,omitempty"`
	OsAdd            *string `gorm:"column:os_add;type:varchar(255)"     json:"os_add,omitempty"`
	OdDva20          *string `gorm:"column:od_dva_20;type:varchar(255)"  json:"od_dva_20,omitempty"`
	OsDva20          *string `gorm:"column:os_dva_20;type:varchar(255)"  json:"os_dva_20,omitempty"`
	OdNva20          *string `gorm:"column:od_nva_20;type:varchar(255)"  json:"od_nva_20,omitempty"`
	OsNva20          *string `gorm:"column:os_nva_20;type:varchar(255)"  json:"os_nva_20,omitempty"`
	OdOverRefraction *string `gorm:"column:od_over_refraction;type:text" json:"od_over_refraction,omitempty"`
	OsOverRefraction *string `gorm:"column:os_over_refraction;type:text" json:"os_over_refraction,omitempty"`
	OdFinal          *bool   `gorm:"column:od_final"                     json:"od_final,omitempty"`
	OsFinal          *bool   `gorm:"column:os_final"                     json:"os_final,omitempty"`
	Evaluation       *string `gorm:"column:evaluation;type:text"         json:"evaluation,omitempty"`
	DominantEye      string  `gorm:"column:dominant_eye;not null;default:'n/a'" json:"dominant_eye"`
}
func (Fitting3) TableName() string { return "fitting_3" }
func (f *Fitting3) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_fitting_3": f.IDFitting3,
		"od_brand": f.OdBrand, "os_brand": f.OsBrand,
		"od_b_cur": f.OdBCur, "os_b_cur": f.OsBCur,
		"od_dia": f.OdDia, "os_dia": f.OsDia,
		"od_pwr": f.OdPwr, "os_pwr": f.OsPwr,
		"od_cyl": f.OdCyl, "os_cyl": f.OsCyl,
		"od_axis": f.OdAxis, "os_axis": f.OsAxis,
		"od_add": f.OdAdd, "os_add": f.OsAdd,
		"od_dva_20": f.OdDva20, "os_dva_20": f.OsDva20,
		"od_nva_20": f.OdNva20, "os_nva_20": f.OsNva20,
		"od_over_refraction": f.OdOverRefraction, "os_over_refraction": f.OsOverRefraction,
		"od_final": f.OdFinal, "os_final": f.OsFinal,
		"evaluation": f.Evaluation, "dominant_eye": f.DominantEye,
	}
}
