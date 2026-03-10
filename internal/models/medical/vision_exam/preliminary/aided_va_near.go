package preliminary
type AidedVANear struct {
	IDAidedVANear int64   `gorm:"column:id_aided_va_near;primaryKey;autoIncrement" json:"id_aided_va_near"`
	Od20          *string `gorm:"column:od_20;type:varchar(255)"                   json:"od_20,omitempty"`
	Os20          *string `gorm:"column:os_20;type:varchar(255)"                   json:"os_20,omitempty"`
	Ou20          *string `gorm:"column:ou_20;type:varchar(255)"                   json:"ou_20,omitempty"`
}
func (AidedVANear) TableName() string { return "aided_va_near" }
func (a *AidedVANear) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_aided_va_near": a.IDAidedVANear, "od_20": a.Od20, "os_20": a.Os20, "ou_20": a.Ou20}
}
