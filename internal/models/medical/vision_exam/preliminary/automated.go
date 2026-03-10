package preliminary
type Automated struct {
	IDAutomated int64   `gorm:"column:id_automated;primaryKey;autoIncrement" json:"id_automated"`
	Od          *string `gorm:"column:od;type:varchar(255)"                  json:"od,omitempty"`
	Os          *string `gorm:"column:os;type:varchar(255)"                  json:"os,omitempty"`
}
func (Automated) TableName() string { return "automated" }
func (a *Automated) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_automated": a.IDAutomated, "od": a.Od, "os": a.Os}
}
