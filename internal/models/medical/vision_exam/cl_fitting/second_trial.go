package cl_fitting

// SecondTrial ↔ table: second_trial
type SecondTrial struct {
	IDSecondTrial     int64   `gorm:"column:id_second_trial;primaryKey;autoIncrement"  json:"id_second_trial"`
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
	Trial             bool    `gorm:"column:trial;not null;default:false"              json:"trial"`
	Final             bool    `gorm:"column:final;not null;default:false"              json:"final"`
	NeedToOrder       bool    `gorm:"column:need_to_order;not null;default:false"      json:"need_to_order"`
	DispenseFromStock bool    `gorm:"column:dispense_from_stock;not null;default:false" json:"dispense_from_stock"`
	FrontDeskNote     *string `gorm:"column:front_desk_note;type:text"                 json:"front_desk_note,omitempty"`
	TypeAdd           *string `gorm:"column:type_add;type:varchar(255)"                json:"type_add,omitempty"`
}
func (SecondTrial) TableName() string { return "second_trial" }
func (s *SecondTrial) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_second_trial": s.IDSecondTrial,
		"od_brand": s.OdBrand, "os_brand": s.OsBrand,
		"trial": s.Trial, "final": s.Final,
		"need_to_order": s.NeedToOrder, "dispense_from_stock": s.DispenseFromStock,
		"front_desk_note": s.FrontDeskNote, "type_add": s.TypeAdd,
	}
}
