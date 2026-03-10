package preliminary
type NearVonGraefePhoria struct {
	IDNearVonGraefePhoria int64   `gorm:"column:id_near_von_graefe_phoria;primaryKey;autoIncrement" json:"id_near_von_graefe_phoria"`
	HNearVgp              *string `gorm:"column:h_near_vgp;type:varchar(255)"                       json:"h_near_vgp,omitempty"`
	VNearVgp              *string `gorm:"column:v_near_vgp;type:varchar(255)"                       json:"v_near_vgp,omitempty"`
}
func (NearVonGraefePhoria) TableName() string { return "near_von_graefe_phoria" }
func (n *NearVonGraefePhoria) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_near_von_graefe_phoria": n.IDNearVonGraefePhoria, "h_near_vgp": n.HNearVgp, "v_near_vgp": n.VNearVgp}
}
