package preliminary

import "time"

type EntranceGlasses struct {
	IDEntranceGlasses int64      `gorm:"column:id_entrance_glasses;primaryKey;autoIncrement" json:"id_entrance_glasses"`
	Data              *time.Time `gorm:"column:data;type:date"                               json:"data,omitempty"`
	OdSph             *string    `gorm:"column:od_sph;type:varchar(10)"                      json:"od_sph,omitempty"`
	OsSph             *string    `gorm:"column:os_sph;type:varchar(10)"                      json:"os_sph,omitempty"`
	OdCyl             *string    `gorm:"column:od_cyl;type:varchar(10)"                      json:"od_cyl,omitempty"`
	OsCyl             *string    `gorm:"column:os_cyl;type:varchar(10)"                      json:"os_cyl,omitempty"`
	OdAxis            *string    `gorm:"column:od_axis;type:varchar(10)"                     json:"od_axis,omitempty"`
	OsAxis            *string    `gorm:"column:os_axis;type:varchar(10)"                     json:"os_axis,omitempty"`
	OdAdd             *string    `gorm:"column:od_add;type:varchar(10)"                      json:"od_add,omitempty"`
	OsAdd             *string    `gorm:"column:os_add;type:varchar(10)"                      json:"os_add,omitempty"`
	OdHPrism          *string    `gorm:"column:od_h_prism;type:varchar(10)"                  json:"od_h_prism,omitempty"`
	OsHPrism          *string    `gorm:"column:os_h_prism;type:varchar(10)"                  json:"os_h_prism,omitempty"`
	OdVPrism          *string    `gorm:"column:od_v_prism;type:varchar(10)"                  json:"od_v_prism,omitempty"`
	OsVPrism          *string    `gorm:"column:os_v_prism;type:varchar(10)"                  json:"os_v_prism,omitempty"`
	Note              *string    `gorm:"column:note;type:text"                               json:"note,omitempty"`
}
func (EntranceGlasses) TableName() string { return "entrance_glasses" }
func (e *EntranceGlasses) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_entrance_glasses": e.IDEntranceGlasses,
		"od_sph": e.OdSph, "os_sph": e.OsSph, "od_cyl": e.OdCyl, "os_cyl": e.OsCyl,
		"od_axis": e.OdAxis, "os_axis": e.OsAxis, "od_add": e.OdAdd, "os_add": e.OsAdd,
		"od_h_prism": e.OdHPrism, "os_h_prism": e.OsHPrism,
		"od_v_prism": e.OdVPrism, "os_v_prism": e.OsVPrism, "note": e.Note,
	}
	if e.Data != nil { m["data"] = e.Data.Format("2006-01-02") } else { m["data"] = nil }
	return m
}
