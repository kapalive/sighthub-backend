package preliminary
type AidedVADistance struct {
	IDAidedVADistance int64   `gorm:"column:id_aided_va_distance;primaryKey;autoIncrement" json:"id_aided_va_distance"`
	Od20              *string `gorm:"column:od_20;type:varchar(255)"                       json:"od_20,omitempty"`
	Os20              *string `gorm:"column:os_20;type:varchar(255)"                       json:"os_20,omitempty"`
	Ou20              *string `gorm:"column:ou_20;type:varchar(255)"                       json:"ou_20,omitempty"`
}
func (AidedVADistance) TableName() string { return "aided_va_distance" }
func (a *AidedVADistance) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_aided_va_distance": a.IDAidedVADistance, "od_20": a.Od20, "os_20": a.Os20, "ou_20": a.Ou20}
}
