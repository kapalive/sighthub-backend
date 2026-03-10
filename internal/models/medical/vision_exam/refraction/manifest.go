package refraction

type Manifest struct {
	IDManifest      int64   `gorm:"column:id_manifest;primaryKey;autoIncrement" json:"id_manifest"`
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
	OdPd          *string `gorm:"column:od_pd;type:varchar(255)"  json:"od_pd,omitempty"`
	OsPd          *string `gorm:"column:os_pd;type:varchar(255)"  json:"os_pd,omitempty"`
	OuPd          *string `gorm:"column:ou_pd;type:varchar(255)"  json:"ou_pd,omitempty"`
	OdNpd         *string `gorm:"column:od_npd;type:varchar(255)" json:"od_npd,omitempty"`
	OsNpd         *string `gorm:"column:os_npd;type:varchar(255)" json:"os_npd,omitempty"`
	OuNpd         *string `gorm:"column:ou_npd;type:varchar(255)" json:"ou_npd,omitempty"`
}
func (Manifest) TableName() string { return "manifest" }
func (v *Manifest) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_manifest": v.IDManifest,
		"od_sph": v.OdSph, "os_sph": v.OsSph, "od_cyl": v.OdCyl, "os_cyl": v.OsCyl,
		"od_axis": v.OdAxis, "os_axis": v.OsAxis, "od_add": v.OdAdd, "os_add": v.OsAdd,
		"od_h_prism": v.OdHPrism, "od_h_prism_list": v.OdHPrismList,
		"os_h_prism": v.OsHPrism, "os_h_prism_list": v.OsHPrismList,
		"od_v_prism": v.OdVPrism, "od_v_prism_list": v.OdVPrismList,
		"os_v_prism": v.OsVPrism, "os_v_prism_list": v.OsVPrismList,
		"od_dva_20": v.OdDva20, "os_dva_20": v.OsDva20,
		"od_nva_20": v.OdNva20, "os_nva_20": v.OsNva20,
		"ou_dva_20": v.OuDva20, "ou_nva_20": v.OuNva20,
		"od_final": v.OdFinal, "os_final": v.OsFinal,
		"od_pd": v.OdPd, "os_pd": v.OsPd, "ou_pd": v.OuPd,
		"od_npd": v.OdNpd, "os_npd": v.OsNpd, "ou_npd": v.OuNpd,
	}
}
