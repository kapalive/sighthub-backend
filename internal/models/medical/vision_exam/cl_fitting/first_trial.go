package cl_fitting

import "time"

// FirstTrial ↔ table: first_trial
type FirstTrial struct {
	IDFirstTrial      int64      `gorm:"column:id_first_trial;primaryKey;autoIncrement"   json:"id_first_trial"`
	OdBrand           *string    `gorm:"column:od_brand;type:varchar(255)"                json:"od_brand,omitempty"`
	OsBrand           *string    `gorm:"column:os_brand;type:varchar(255)"                json:"os_brand,omitempty"`
	OdBCur            *string    `gorm:"column:od_b_cur;type:varchar(255)"                json:"od_b_cur,omitempty"`
	OsBCur            *string    `gorm:"column:os_b_cur;type:varchar(255)"                json:"os_b_cur,omitempty"`
	OdDia             *string    `gorm:"column:od_dia;type:varchar(255)"                  json:"od_dia,omitempty"`
	OsDia             *string    `gorm:"column:os_dia;type:varchar(255)"                  json:"os_dia,omitempty"`
	OdPwr             *string    `gorm:"column:od_pwr;type:varchar(255)"                  json:"od_pwr,omitempty"`
	OsPwr             *string    `gorm:"column:os_pwr;type:varchar(255)"                  json:"os_pwr,omitempty"`
	OdCyl             *string    `gorm:"column:od_cyl;type:varchar(255)"                  json:"od_cyl,omitempty"`
	OsCyl             *string    `gorm:"column:os_cyl;type:varchar(255)"                  json:"os_cyl,omitempty"`
	OdAxis            *string    `gorm:"column:od_axis;type:varchar(255)"                 json:"od_axis,omitempty"`
	OsAxis            *string    `gorm:"column:os_axis;type:varchar(255)"                 json:"os_axis,omitempty"`
	OdAdd             *string    `gorm:"column:od_add;type:varchar(255)"                  json:"od_add,omitempty"`
	OsAdd             *string    `gorm:"column:os_add;type:varchar(255)"                  json:"os_add,omitempty"`
	OdDva20           *string    `gorm:"column:od_dva_20;type:varchar(255)"               json:"od_dva_20,omitempty"`
	OsDva20           *string    `gorm:"column:os_dva_20;type:varchar(255)"               json:"os_dva_20,omitempty"`
	OdNva20           *string    `gorm:"column:od_nva_20;type:varchar(255)"               json:"od_nva_20,omitempty"`
	OsNva20           *string    `gorm:"column:os_nva_20;type:varchar(255)"               json:"os_nva_20,omitempty"`
	Trial             bool       `gorm:"column:trial;not null;default:false"              json:"trial"`
	Final             bool       `gorm:"column:final;not null;default:false"              json:"final"`
	NeedToOrder       bool       `gorm:"column:need_to_order;not null;default:false"      json:"need_to_order"`
	DispenseFromStock bool       `gorm:"column:dispense_from_stock;not null;default:false" json:"dispense_from_stock"`
	FrontDeskNote     *string    `gorm:"column:front_desk_note;type:text"                 json:"front_desk_note,omitempty"`
	ExpireDate        *time.Time `gorm:"column:expire_date;type:date"                     json:"expire_date,omitempty"`
}
func (FirstTrial) TableName() string { return "first_trial" }
func (f *FirstTrial) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_first_trial": f.IDFirstTrial,
		"od_brand": f.OdBrand, "os_brand": f.OsBrand,
		"od_b_cur": f.OdBCur, "os_b_cur": f.OsBCur,
		"od_dia": f.OdDia, "os_dia": f.OsDia,
		"od_pwr": f.OdPwr, "os_pwr": f.OsPwr,
		"od_cyl": f.OdCyl, "os_cyl": f.OsCyl,
		"od_axis": f.OdAxis, "os_axis": f.OsAxis,
		"od_add": f.OdAdd, "os_add": f.OsAdd,
		"od_dva_20": f.OdDva20, "os_dva_20": f.OsDva20,
		"od_nva_20": f.OdNva20, "os_nva_20": f.OsNva20,
		"trial": f.Trial, "final": f.Final,
		"need_to_order": f.NeedToOrder, "dispense_from_stock": f.DispenseFromStock,
		"front_desk_note": f.FrontDeskNote,
	}
	if f.ExpireDate != nil { m["expire_date"] = f.ExpireDate.Format("2006-01-02") } else { m["expire_date"] = nil }
	return m
}
