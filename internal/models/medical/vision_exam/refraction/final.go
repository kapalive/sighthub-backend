package refraction

import "time"

// RefractionFinal ↔ table: final
type RefractionFinal struct {
	IDFinal      int64      `gorm:"column:id_final;primaryKey;autoIncrement" json:"id_final"`
	OdSph        *string    `gorm:"column:od_sph;type:varchar(255)"  json:"od_sph,omitempty"`
	OsSph        *string    `gorm:"column:os_sph;type:varchar(255)"  json:"os_sph,omitempty"`
	OdCyl        *string    `gorm:"column:od_cyl;type:varchar(255)"  json:"od_cyl,omitempty"`
	OsCyl        *string    `gorm:"column:os_cyl;type:varchar(255)"  json:"os_cyl,omitempty"`
	OdAxis       *string    `gorm:"column:od_axis;type:varchar(255)" json:"od_axis,omitempty"`
	OsAxis       *string    `gorm:"column:os_axis;type:varchar(255)" json:"os_axis,omitempty"`
	OdAdd        *string    `gorm:"column:od_add;type:varchar(255)"  json:"od_add,omitempty"`
	OsAdd        *string    `gorm:"column:os_add;type:varchar(255)"  json:"os_add,omitempty"`
	OdHPrism     *string    `gorm:"column:od_h_prism;type:varchar(255)"      json:"od_h_prism,omitempty"`
	OdHPrismList *string    `gorm:"column:od_h_prism_list;type:varchar(255)" json:"od_h_prism_list,omitempty"`
	OsHPrism     *string    `gorm:"column:os_h_prism;type:varchar(255)"      json:"os_h_prism,omitempty"`
	OsHPrismList *string    `gorm:"column:os_h_prism_list;type:varchar(255)" json:"os_h_prism_list,omitempty"`
	OdVPrism     *string    `gorm:"column:od_v_prism;type:varchar(255)"      json:"od_v_prism,omitempty"`
	OdVPrismList *string    `gorm:"column:od_v_prism_list;type:varchar(255)" json:"od_v_prism_list,omitempty"`
	OsVPrism     *string    `gorm:"column:os_v_prism;type:varchar(255)"      json:"os_v_prism,omitempty"`
	OsVPrismList *string    `gorm:"column:os_v_prism_list;type:varchar(255)" json:"os_v_prism_list,omitempty"`
	OdDva20      *string    `gorm:"column:od_dva_20;type:varchar(255)" json:"od_dva_20,omitempty"`
	OsDva20      *string    `gorm:"column:os_dva_20;type:varchar(255)" json:"os_dva_20,omitempty"`
	OdNva20      *string    `gorm:"column:od_nva_20;type:varchar(255)" json:"od_nva_20,omitempty"`
	OsNva20      *string    `gorm:"column:os_nva_20;type:varchar(255)" json:"os_nva_20,omitempty"`
	OuDva20      *string    `gorm:"column:ou_dva_20;type:varchar(255)" json:"ou_dva_20,omitempty"`
	OuNva20      *string    `gorm:"column:ou_nva_20;type:varchar(255)" json:"ou_nva_20,omitempty"`
	OdPd         *string    `gorm:"column:od_pd;type:varchar(255)"  json:"od_pd,omitempty"`
	OsPd         *string    `gorm:"column:os_pd;type:varchar(255)"  json:"os_pd,omitempty"`
	OuPd         *string    `gorm:"column:ou_pd;type:varchar(255)"  json:"ou_pd,omitempty"`
	OdNpd        *string    `gorm:"column:od_npd;type:varchar(255)" json:"od_npd,omitempty"`
	OsNpd        *string    `gorm:"column:os_npd;type:varchar(255)" json:"os_npd,omitempty"`
	OuNpd        *string    `gorm:"column:ou_npd;type:varchar(255)" json:"ou_npd,omitempty"`
	ExpireDate   *time.Time `gorm:"column:expire_date;type:date"    json:"expire_date,omitempty"`
	Note         *string    `gorm:"column:note;type:varchar(255)"   json:"note,omitempty"`
}
func (RefractionFinal) TableName() string { return "final" }
func (f *RefractionFinal) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_final": f.IDFinal,
		"od_sph": f.OdSph, "os_sph": f.OsSph, "od_cyl": f.OdCyl, "os_cyl": f.OsCyl,
		"od_axis": f.OdAxis, "os_axis": f.OsAxis, "od_add": f.OdAdd, "os_add": f.OsAdd,
		"od_h_prism": f.OdHPrism, "od_h_prism_list": f.OdHPrismList,
		"os_h_prism": f.OsHPrism, "os_h_prism_list": f.OsHPrismList,
		"od_v_prism": f.OdVPrism, "od_v_prism_list": f.OdVPrismList,
		"os_v_prism": f.OsVPrism, "os_v_prism_list": f.OsVPrismList,
		"od_dva_20": f.OdDva20, "os_dva_20": f.OsDva20,
		"od_nva_20": f.OdNva20, "os_nva_20": f.OsNva20,
		"ou_dva_20": f.OuDva20, "ou_nva_20": f.OuNva20,
		"od_pd": f.OdPd, "os_pd": f.OsPd, "ou_pd": f.OuPd,
		"od_npd": f.OdNpd, "os_npd": f.OsNpd, "ou_npd": f.OuNpd,
		"note": f.Note,
	}
	if f.ExpireDate != nil { m["expire_date"] = f.ExpireDate.Format("2006-01-02") } else { m["expire_date"] = nil }
	return m
}
