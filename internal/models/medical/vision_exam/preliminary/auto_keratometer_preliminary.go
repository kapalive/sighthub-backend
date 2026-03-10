package preliminary
type AutoKeratometerPreliminary struct {
	IDAutoKeratometerPreliminary int64   `gorm:"column:id_auto_keratometer_preliminary;primaryKey;autoIncrement" json:"id_auto_keratometer_preliminary"`
	OdPw1                        *string `gorm:"column:od_pw1;type:varchar(255)"                                 json:"od_pw1,omitempty"`
	OsPw1                        *string `gorm:"column:os_pw1;type:varchar(255)"                                 json:"os_pw1,omitempty"`
	OdPw2                        *string `gorm:"column:od_pw2;type:varchar(255)"                                 json:"od_pw2,omitempty"`
	OsPw2                        *string `gorm:"column:os_pw2;type:varchar(255)"                                 json:"os_pw2,omitempty"`
	OdAxis                       *string `gorm:"column:od_axis;type:varchar(255)"                                json:"od_axis,omitempty"`
	OsAxis                       *string `gorm:"column:os_axis;type:varchar(255)"                                json:"os_axis,omitempty"`
}
func (AutoKeratometerPreliminary) TableName() string { return "auto_keratometer_preliminary" }
func (a *AutoKeratometerPreliminary) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_auto_keratometer_preliminary": a.IDAutoKeratometerPreliminary,
		"od_pw1": a.OdPw1, "os_pw1": a.OsPw1, "od_pw2": a.OdPw2, "os_pw2": a.OsPw2,
		"od_axis": a.OdAxis, "os_axis": a.OsAxis,
	}
}
