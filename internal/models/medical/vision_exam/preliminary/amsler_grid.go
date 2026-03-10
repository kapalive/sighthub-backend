package preliminary
type AmslerGrid struct {
	IDAmslerGrid int64   `gorm:"column:id_amsler_grid;primaryKey;autoIncrement" json:"id_amsler_grid"`
	Od           *string `gorm:"column:od;type:varchar(255)"                    json:"od,omitempty"`
	Os           *string `gorm:"column:os;type:varchar(255)"                    json:"os,omitempty"`
}
func (AmslerGrid) TableName() string { return "amsler_grid" }
func (a *AmslerGrid) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_amsler_grid": a.IDAmslerGrid, "od": a.Od, "os": a.Os}
}
