package refraction

// Final2 ↔ table: final2 (column id_final)
type Final2 struct {
	IDFinal2     int64   `gorm:"column:id_final;primaryKey;autoIncrement" json:"id_final"`
	OdSph        *string `gorm:"column:od_sph;type:varchar(255)"  json:"od_sph,omitempty"`
	OsSph        *string `gorm:"column:os_sph;type:varchar(255)"  json:"os_sph,omitempty"`
	OdCyl        *string `gorm:"column:od_cyl;type:varchar(255)"  json:"od_cyl,omitempty"`
	OsCyl        *string `gorm:"column:os_cyl;type:varchar(255)"  json:"os_cyl,omitempty"`
	OdAxis       *string `gorm:"column:od_axis;type:varchar(255)" json:"od_axis,omitempty"`
	OsAxis       *string `gorm:"column:os_axis;type:varchar(255)" json:"os_axis,omitempty"`
	OdAdd        *string `gorm:"column:od_add;type:varchar(255)"  json:"od_add,omitempty"`
	OsAdd        *string `gorm:"column:os_add;type:varchar(255)"  json:"os_add,omitempty"`
	OdHPrism     *string `gorm:"column:od_h_prism;type:varchar(255)"      json:"od_h_prism,omitempty"`
	OdHPrismList *string `gorm:"column:od_h_prism_list;type:varchar(255)" json:"od_h_prism_list,omitempty"`
	OsHPrism     *string `gorm:"column:os_h_prism;type:varchar(255)"      json:"os_h_prism,omitempty"`
	OsHPrismList *string `gorm:"column:os_h_prism_list;type:varchar(255)" json:"os_h_prism_list,omitempty"`
	OdVPrism     *string `gorm:"column:od_v_prism;type:varchar(255)"      json:"od_v_prism,omitempty"`
	OdVPrismList *string `gorm:"column:od_v_prism_list;type:varchar(255)" json:"od_v_prism_list,omitempty"`
	OsVPrism     *string `gorm:"column:os_v_prism;type:varchar(255)"      json:"os_v_prism,omitempty"`
	OsVPrismList *string `gorm:"column:os_v_prism_list;type:varchar(255)" json:"os_v_prism_list,omitempty"`
	OdDva20      *string `gorm:"column:od_dva_20;type:varchar(255)" json:"od_dva_20,omitempty"`
	OsDva20      *string `gorm:"column:os_dva_20;type:varchar(255)" json:"os_dva_20,omitempty"`
	OuDva20      *string `gorm:"column:ou_dva_20;type:varchar(255)" json:"ou_dva_20,omitempty"`
	OdPd         *string `gorm:"column:od_pd;type:varchar(255)"  json:"od_pd,omitempty"`
	OsPd         *string `gorm:"column:os_pd;type:varchar(255)"  json:"os_pd,omitempty"`
	OuPd         *string `gorm:"column:ou_pd;type:varchar(255)"  json:"ou_pd,omitempty"`
	OdNpd        *string `gorm:"column:od_npd;type:varchar(255)" json:"od_npd,omitempty"`
	OsNpd        *string `gorm:"column:os_npd;type:varchar(255)" json:"os_npd,omitempty"`
	OuNpd        *string `gorm:"column:ou_npd;type:varchar(255)" json:"ou_npd,omitempty"`
	Desc         *string `gorm:"column:desc;type:varchar(255)"   json:"desc,omitempty"`
	Note         *string `gorm:"column:note;type:varchar(255)"   json:"note,omitempty"`
}
func (Final2) TableName() string { return "final2" }
func (f *Final2) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_final": f.IDFinal2,
		"od_sph": f.OdSph, "os_sph": f.OsSph, "od_cyl": f.OdCyl, "os_cyl": f.OsCyl,
		"od_axis": f.OdAxis, "os_axis": f.OsAxis, "od_add": f.OdAdd, "os_add": f.OsAdd,
		"od_dva_20": f.OdDva20, "os_dva_20": f.OsDva20, "ou_dva_20": f.OuDva20,
		"od_pd": f.OdPd, "os_pd": f.OsPd, "ou_pd": f.OuPd,
		"od_npd": f.OdNpd, "os_npd": f.OsNpd, "ou_npd": f.OuNpd,
		"desc": f.Desc, "note": f.Note,
	}
}
