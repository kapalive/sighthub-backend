package preliminary
type Bruckner struct {
	IDBruckner int64   `gorm:"column:id_bruckner;primaryKey;autoIncrement" json:"id_bruckner"`
	Od         *string `gorm:"column:od;type:varchar(255)"                 json:"od,omitempty"`
	Os         *string `gorm:"column:os;type:varchar(255)"                 json:"os,omitempty"`
	GoodReflex *bool   `gorm:"column:good_reflex"                          json:"good_reflex,omitempty"`
}
func (Bruckner) TableName() string { return "bruckner" }
func (b *Bruckner) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_bruckner": b.IDBruckner, "od": b.Od, "os": b.Os, "good_reflex": b.GoodReflex}
}
