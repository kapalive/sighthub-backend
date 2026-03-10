package preliminary
type DistVergenceTest struct {
	IDDistVergenceTest int     `gorm:"column:id_dist_vergence_testing;primaryKey;autoIncrement" json:"id_dist_vergence_testing"`
	Bi1                *string `gorm:"column:bi1;type:varchar(50)"                              json:"bi1,omitempty"`
	Bo1                *string `gorm:"column:bo1;type:varchar(50)"                              json:"bo1,omitempty"`
	Bi2                *string `gorm:"column:bi2;type:varchar(50)"                              json:"bi2,omitempty"`
	Bo2                *string `gorm:"column:bo2;type:varchar(50)"                              json:"bo2,omitempty"`
	Bi3                *string `gorm:"column:bi3;type:varchar(50)"                              json:"bi3,omitempty"`
	Bo3                *string `gorm:"column:bo3;type:varchar(50)"                              json:"bo3,omitempty"`
}
func (DistVergenceTest) TableName() string { return "dist_vergence_testing" }
func (d *DistVergenceTest) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_dist_vergence_testing": d.IDDistVergenceTest,
		"bi1": d.Bi1, "bo1": d.Bo1, "bi2": d.Bi2, "bo2": d.Bo2, "bi3": d.Bi3, "bo3": d.Bo3,
	}
}
