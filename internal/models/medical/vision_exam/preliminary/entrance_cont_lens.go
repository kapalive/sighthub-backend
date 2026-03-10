package preliminary

import "time"

type EntranceContLens struct {
	IDEntranceContLens int64      `gorm:"column:id_entrance_cont_lens;primaryKey;autoIncrement" json:"id_entrance_cont_lens"`
	Data               *time.Time `gorm:"column:data;type:date"                                 json:"data,omitempty"`
	OdBrand            *string    `gorm:"column:od_brand;type:varchar(100)"                     json:"od_brand,omitempty"`
	OsBrand            *string    `gorm:"column:os_brand;type:varchar(100)"                     json:"os_brand,omitempty"`
	OdBaseC            *string    `gorm:"column:od_base_c;type:varchar(50)"                     json:"od_base_c,omitempty"`
	OsBaseC            *string    `gorm:"column:os_base_c;type:varchar(50)"                     json:"os_base_c,omitempty"`
	OdDia              *string    `gorm:"column:od_dia;type:varchar(50)"                        json:"od_dia,omitempty"`
	OsDia              *string    `gorm:"column:os_dia;type:varchar(50)"                        json:"os_dia,omitempty"`
	OdPwr              *string    `gorm:"column:od_pwr;type:varchar(50)"                        json:"od_pwr,omitempty"`
	OsPwr              *string    `gorm:"column:os_pwr;type:varchar(50)"                        json:"os_pwr,omitempty"`
	OdCyl              *string    `gorm:"column:od_cyl;type:varchar(50)"                        json:"od_cyl,omitempty"`
	OsCyl              *string    `gorm:"column:os_cyl;type:varchar(50)"                        json:"os_cyl,omitempty"`
	OdAxis             *string    `gorm:"column:od_axis;type:varchar(50)"                       json:"od_axis,omitempty"`
	OsAxis             *string    `gorm:"column:os_axis;type:varchar(50)"                       json:"os_axis,omitempty"`
	OdAdd              *string    `gorm:"column:od_add;type:varchar(50)"                        json:"od_add,omitempty"`
	OsAdd              *string    `gorm:"column:os_add;type:varchar(50)"                        json:"os_add,omitempty"`
}
func (EntranceContLens) TableName() string { return "entrance_cont_lens" }
func (e *EntranceContLens) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_entrance_cont_lens": e.IDEntranceContLens,
		"od_brand": e.OdBrand, "os_brand": e.OsBrand,
		"od_base_c": e.OdBaseC, "os_base_c": e.OsBaseC,
		"od_dia": e.OdDia, "os_dia": e.OsDia,
		"od_pwr": e.OdPwr, "os_pwr": e.OsPwr,
		"od_cyl": e.OdCyl, "os_cyl": e.OsCyl,
		"od_axis": e.OdAxis, "os_axis": e.OsAxis,
		"od_add": e.OdAdd, "os_add": e.OsAdd,
	}
	if e.Data != nil { m["data"] = e.Data.Format("2006-01-02") } else { m["data"] = nil }
	return m
}
