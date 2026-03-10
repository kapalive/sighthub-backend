package preliminary
type AutorefractorPreliminary struct {
	IDAutorefractorPreliminary int64   `gorm:"column:id_autorefractor_preliminary;primaryKey;autoIncrement" json:"id_autorefractor_preliminary"`
	OdSph                      *string `gorm:"column:od_sph;type:varchar(255)"                              json:"od_sph,omitempty"`
	OsSph                      *string `gorm:"column:os_sph;type:varchar(255)"                              json:"os_sph,omitempty"`
	OdCyl                      *string `gorm:"column:od_cyl;type:varchar(255)"                              json:"od_cyl,omitempty"`
	OsCyl                      *string `gorm:"column:os_cyl;type:varchar(255)"                              json:"os_cyl,omitempty"`
	OdAxis                     *string `gorm:"column:od_axis;type:varchar(255)"                             json:"od_axis,omitempty"`
	OsAxis                     *string `gorm:"column:os_axis;type:varchar(255)"                             json:"os_axis,omitempty"`
	Pd                         *string `gorm:"column:pd;type:varchar(255)"                                  json:"pd,omitempty"`
}
func (AutorefractorPreliminary) TableName() string { return "autorefractor_preliminary" }
func (a *AutorefractorPreliminary) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_autorefractor_preliminary": a.IDAutorefractorPreliminary,
		"od_sph": a.OdSph, "os_sph": a.OsSph, "od_cyl": a.OdCyl, "os_cyl": a.OsCyl,
		"od_axis": a.OdAxis, "os_axis": a.OsAxis, "pd": a.Pd,
	}
}
