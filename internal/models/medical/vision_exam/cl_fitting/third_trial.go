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
	Trial             bool    `gorm:"column:trial;not null;default:false"              json:"trial"`
	Final             bool    `gorm:"column:final;not null;default:false"              json:"final"`
	NeedToOrder       bool    `gorm:"column:need_to_order;not null;default:false"      json:"need_to_order"`
	DispenseFromStock bool    `gorm:"column:dispense_from_stock;not null;default:false" json:"dispense_from_stock"`
}
func (ThirdTrial) TableName() string { return "third_trial" }
func (t *ThirdTrial) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_third_trial": t.IDThirdTrial,
		"od_brand": t.OdBrand, "os_brand": t.OsBrand,
		"trial": t.Trial, "final": t.Final,
		"need_to_order": t.NeedToOrder, "dispense_from_stock": t.DispenseFromStock,
	}
}
