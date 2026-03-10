package preliminary
type UnaidedVADistance struct {
	IDUnaidedVADistance int64   `gorm:"column:id_unaided_va_distance;primaryKey;autoIncrement" json:"id_unaided_va_distance"`
	Od20                *string `gorm:"column:od_20;type:varchar(255)"                         json:"od_20,omitempty"`
	Os20                *string `gorm:"column:os_20;type:varchar(255)"                         json:"os_20,omitempty"`
	Ou20                *string `gorm:"column:ou_20;type:varchar(255)"                         json:"ou_20,omitempty"`
}
func (UnaidedVADistance) TableName() string { return "unaided_va_distance" }
func (u *UnaidedVADistance) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_unaided_va_distance": u.IDUnaidedVADistance, "od_20": u.Od20, "os_20": u.Os20, "ou_20": u.Ou20}
}
