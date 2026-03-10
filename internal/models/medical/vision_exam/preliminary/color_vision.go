package preliminary
type ColorVision struct {
	IDColorVision int64   `gorm:"column:id_color_vision;primaryKey;autoIncrement" json:"id_color_vision"`
	Od1           *string `gorm:"column:od1;type:varchar(255)"                    json:"od1,omitempty"`
	Od2           *string `gorm:"column:od2;type:varchar(255)"                    json:"od2,omitempty"`
	Os1           *string `gorm:"column:os1;type:varchar(255)"                    json:"os1,omitempty"`
	Os2           *string `gorm:"column:os2;type:varchar(255)"                    json:"os2,omitempty"`
}
func (ColorVision) TableName() string { return "color_vision" }
func (c *ColorVision) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_color_vision": c.IDColorVision, "od1": c.Od1, "od2": c.Od2, "os1": c.Os1, "os2": c.Os2}
}
