package refraction

type Retinoscopy struct {
	IDRetinoscopy int64   `gorm:"column:id_retinoscopy;primaryKey;autoIncrement" json:"id_retinoscopy"`
	OdSph         *string `gorm:"column:od_sph;type:varchar(255)"  json:"od_sph,omitempty"`
	OsSph         *string `gorm:"column:os_sph;type:varchar(255)"  json:"os_sph,omitempty"`
	OdCyl         *string `gorm:"column:od_cyl;type:varchar(255)"  json:"od_cyl,omitempty"`
	OsCyl         *string `gorm:"column:os_cyl;type:varchar(255)"  json:"os_cyl,omitempty"`
	OdAxis        *string `gorm:"column:od_axis;type:varchar(255)" json:"od_axis,omitempty"`
	OsAxis        *string `gorm:"column:os_axis;type:varchar(255)" json:"os_axis,omitempty"`
	OdAdd         *string `gorm:"column:od_add;type:varchar(255)"  json:"od_add,omitempty"`
	OsAdd         *string `gorm:"column:os_add;type:varchar(255)"  json:"os_add,omitempty"`
	OdHPrism      *string `gorm:"column:od_h_prism;type:varchar(255)"      json:"od_h_prism,omitempty"`
	OdHPrismList  *string `gorm:"column:od_h_prism_list;type:varchar(255)" json:"od_h_prism_list,omitempty"`
	OsHPrism      *string `gorm:"column:os_h_prism;type:varchar(255)"      json:"os_h_prism,omitempty"`
	OsHPrismList  *string `gorm:"column:os_h_prism_list;type:varchar(255)" json:"os_h_prism_list,omitempty"`
	OdVPrism      *string `gorm:"column:od_v_prism;type:varchar(255)"      json:"od_v_prism,omitempty"`
	OdVPrismList  *string `gorm:"column:od_v_prism_list;type:varchar(255)" json:"od_v_prism_list,omitempty"`
	OsVPrism      *string `gorm:"column:os_v_prism;type:varchar(255)"      json:"os_v_prism,omitempty"`
	OsVPrismList  *string `gorm:"column:os_v_prism_list;type:varchar(255)" json:"os_v_prism_list,omitempty"`
	OdDva20       *string `gorm:"column:od_dva_20;type:varchar(255)" json:"od_dva_20,omitempty"`
	OsDva20       *string `gorm:"column:os_dva_20;type:varchar(255)" json:"os_dva_20,omitempty"`
	OdNva20       *string `gorm:"column:od_nva_20;type:varchar(255)" json:"od_nva_20,omitempty"`
	OsNva20       *string `gorm:"column:os_nva_20;type:varchar(255)" json:"os_nva_20,omitempty"`
	OuDva20       *string `gorm:"column:ou_dva_20;type:varchar(255)" json:"ou_dva_20,omitempty"`
	OuNva20       *string `gorm:"column:ou_nva_20;type:varchar(255)" json:"ou_nva_20,omitempty"`
	OdFinal       bool    `gorm:"column:od_final;not null;default:false" json:"od_final"`
	OsFinal       bool    `gorm:"column:os_final;not null;default:false" json:"os_final"`
}
func (Retinoscopy) TableName() string { return "retinoscopy" }
func (r *Retinoscopy) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_retinoscopy": r.IDRetinoscopy,
		"od_sph": r.OdSph, "os_sph": r.OsSph, "od_cyl": r.OdCyl, "os_cyl": r.OsCyl,
		"od_axis": r.OdAxis, "os_axis": r.OsAxis, "od_add": r.OdAdd, "os_add": r.OsAdd,
		"od_h_prism": r.OdHPrism, "od_h_prism_list": r.OdHPrismList,
		"os_h_prism": r.OsHPrism, "os_h_prism_list": r.OsHPrismList,
		"od_v_prism": r.OdVPrism, "od_v_prism_list": r.OdVPrismList,
		"os_v_prism": r.OsVPrism, "os_v_prism_list": r.OsVPrismList,
		"od_dva_20": r.OdDva20, "os_dva_20": r.OsDva20,
		"od_nva_20": r.OdNva20, "os_nva_20": r.OsNva20,
		"ou_dva_20": r.OuDva20, "ou_nva_20": r.OuNva20,
		"od_final": r.OdFinal, "os_final": r.OsFinal,
	}
}
