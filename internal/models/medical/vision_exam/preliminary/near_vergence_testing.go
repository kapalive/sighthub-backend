package preliminary
type NearVergenceTest struct {
	IDNearVergenceTest int     `gorm:"column:id_near_vergence_testing;primaryKey;autoIncrement" json:"id_near_vergence_testing"`
	Bi1                *string `gorm:"column:bi1;type:varchar(50)"                              json:"bi1,omitempty"`
	Bo1                *string `gorm:"column:bo1;type:varchar(50)"                              json:"bo1,omitempty"`
	Bi2                *string `gorm:"column:bi2;type:varchar(50)"                              json:"bi2,omitempty"`
	Bo2                *string `gorm:"column:bo2;type:varchar(50)"                              json:"bo2,omitempty"`
	Bi3                *string `gorm:"column:bi3;type:varchar(50)"                              json:"bi3,omitempty"`
	Bo3                *string `gorm:"column:bo3;type:varchar(50)"                              json:"bo3,omitempty"`
}
func (NearVergenceTest) TableName() string { return "near_vergence_testing" }
func (n *NearVergenceTest) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_near_vergence_testing": n.IDNearVergenceTest,
		"bi1": n.Bi1, "bo1": n.Bo1, "bi2": n.Bi2, "bo2": n.Bo2, "bi3": n.Bi3, "bo3": n.Bo3,
	}
}
