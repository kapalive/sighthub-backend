package preliminary
type UnaidedPHDistance struct {
	IDUnaidedPHDistance int64   `gorm:"column:id_unaided_ph_distance;primaryKey;autoIncrement" json:"id_unaided_ph_distance"`
	Od20                *string `gorm:"column:od_20;type:varchar(255)"                         json:"od_20,omitempty"`
	Os20                *string `gorm:"column:os_20;type:varchar(255)"                         json:"os_20,omitempty"`
}
func (UnaidedPHDistance) TableName() string { return "unaided_ph_distance" }
func (u *UnaidedPHDistance) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_unaided_ph_distance": u.IDUnaidedPHDistance, "od_20": u.Od20, "os_20": u.Os20}
}
