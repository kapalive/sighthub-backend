package cl_fitting

// ThirdTrial ↔ table: third_trial
type ThirdTrial struct {
	IDThirdTrial      int64   `gorm:"column:id_third_trial;primaryKey;autoIncrement"   json:"id_third_trial"`
	OdBrand           *string `gorm:"column:od_brand;type:varchar(255)"                json:"od_brand,omitempty"`
	OsBrand           *string `gorm:"column:os_brand;type:varchar(255)"                json:"os_brand,omitempty"`
	OdBCur            *string `gorm:"column:od_b_cur;type:varchar(255)"                json:"od_b_cur,omitempty"`
	OsBCur            *string `gorm:"column:os_b_cur;type:varchar(255)"                json:"os_b_cur,omitempty"`
	OdDia             *string `gorm:"column:od_dia;type:varchar(255)"                  json:"od_dia,omitempty"`
	OsDia             *string `gorm:"column:os_dia;type:varchar(255)"                  json:"os_dia,omitempty"`
	OdPwr             *string `gorm:"column:od_pwr;type:varchar(255)"                  json:"od_pwr,omitempty"`
	OsPwr             *string `gorm:"column:os_pwr;type:varchar(255)"                  json:"os_pwr,omitempty"`
	OdCyl             *string `gorm:"column:od_cyl;type:varchar(255)"                  json:"od_cyl,omitempty"`
	OsCyl             *string `gorm:"column:os_cyl;type:varchar(255)"                  json:"os_cyl,omitempty"`
	OdAxis            *string `gorm:"column:od_axis;type:varchar(255)"                 json:"od_axis,omitempty"`
	OsAxis            *string `gorm:"column:os_axis;type:varchar(255)"                 json:"os_axis,omitempty"`
	OdAdd             *string `gorm:"column:od_add;type:varchar(255)"                  json:"od_add,omitempty"`
	OsAdd             *string `gorm:"column:os_add;type:varchar(255)"                  json:"os_add,omitempty"`
	OdDva20           *string `gorm:"column:od_dva_20;type:varchar(255)"               json:"od_dva_20,omitempty"`
	OsDva20           *string `gorm:"column:os_dva_20;type:varchar(255)"               json:"os_dva_20,omitempty"`
	OdNva20           *string `gorm:"column:od_nva_20;type:varchar(255)"               json:"od_nva_20,omitempty"`
	OsNva20           *string `gorm:"column:os_nva_20;type:varchar(255)"               json:"os_nva_20,omitempty"`
	Trial             bool    `gorm:"column:trial;not null;default:false"              json:"trial"`
	Final             bool    `gorm:"column:final;not null;default:false"              json:"final"`
	NeedToOrder       bool    `gorm:"column:need_to_order;not null;default:false"      json:"need_to_order"`
	DispenseFromStock bool    `gorm:"column:dispense_from_stock;not null;default:false" json:"dispense_from_stock"`
	FrontDeskNote     *string `gorm:"column:front_desk_note;type:text"                 json:"front_desk_note,omitempty"`
	TypeAdd           *string `gorm:"column:type_add;type:varchar(255)"                json:"type_add,omitempty"`
}
func (ThirdTrial) TableName() string { return "third_trial" }
func (t *ThirdTrial) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_third_trial": t.IDThirdTrial,
		"od_brand": t.OdBrand, "os_brand": t.OsBrand,
		"od_add": t.OdAdd, "os_add": t.OsAdd,
		"od_dva_20": t.OdDva20, "os_dva_20": t.OsDva20,
		"od_nva_20": t.OdNva20, "os_nva_20": t.OsNva20,
		"trial": t.Trial, "final": t.Final,
		"need_to_order": t.NeedToOrder, "dispense_from_stock": t.DispenseFromStock,
		"front_desk_note": t.FrontDeskNote, "type_add": t.TypeAdd,
	}
}
