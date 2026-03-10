package preliminary
type BloodPressure struct {
	IDBloodPressure int64   `gorm:"column:id_blood_pressure;primaryKey;autoIncrement" json:"id_blood_pressure"`
	Sbp             *string `gorm:"column:sbp;type:varchar(255)"                      json:"sbp,omitempty"`
	Dbp             *string `gorm:"column:dbp;type:varchar(255)"                      json:"dbp,omitempty"`
}
func (BloodPressure) TableName() string { return "blood_pressure" }
func (b *BloodPressure) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_blood_pressure": b.IDBloodPressure, "sbp": b.Sbp, "dbp": b.Dbp}
}
