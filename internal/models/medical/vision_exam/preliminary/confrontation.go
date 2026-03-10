package preliminary
type Confrontation struct {
	IDConfrontation int64   `gorm:"column:id_confrontation;primaryKey;autoIncrement" json:"id_confrontation"`
	Od              *string `gorm:"column:od;type:varchar(255)"                      json:"od,omitempty"`
	Os              *string `gorm:"column:os;type:varchar(255)"                      json:"os,omitempty"`
}
func (Confrontation) TableName() string { return "confrontation" }
func (c *Confrontation) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_confrontation": c.IDConfrontation, "od": c.Od, "os": c.Os}
}
