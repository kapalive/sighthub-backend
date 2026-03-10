package preliminary
type Motility struct {
	IDMotility int64   `gorm:"column:id_motility;primaryKey;autoIncrement" json:"id_motility"`
	Od         *string `gorm:"column:od;type:varchar(255)"                 json:"od,omitempty"`
	Os         *string `gorm:"column:os;type:varchar(255)"                 json:"os,omitempty"`
}
func (Motility) TableName() string { return "motility" }
func (m *Motility) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_motility": m.IDMotility, "od": m.Od, "os": m.Os}
}
