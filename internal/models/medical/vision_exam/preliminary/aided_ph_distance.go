package preliminary
type AidedPHDistance struct {
	IDAidedPHDistance int64   `gorm:"column:id_aided_ph_distance;primaryKey;autoIncrement" json:"id_aided_ph_distance"`
	Od20              *string `gorm:"column:od_20;type:varchar(255)"                       json:"od_20,omitempty"`
	Os20              *string `gorm:"column:os_20;type:varchar(255)"                       json:"os_20,omitempty"`
}
func (AidedPHDistance) TableName() string { return "aided_ph_distance" }
func (a *AidedPHDistance) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_aided_ph_distance": a.IDAidedPHDistance, "od_20": a.Od20, "os_20": a.Os20}
}
