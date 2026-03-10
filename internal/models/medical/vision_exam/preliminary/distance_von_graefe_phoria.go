package preliminary
type DistanceVonGraefePhoria struct {
	IDDistanceVonGraefePhoria int64   `gorm:"column:id_distance_von_graefe_phoria;primaryKey;autoIncrement" json:"id_distance_von_graefe_phoria"`
	HDistVgp                  *string `gorm:"column:h_dist_vgp;type:varchar(255)"                           json:"h_dist_vgp,omitempty"`
	VDistVgp                  *string `gorm:"column:v_dist_vgp;type:varchar(255)"                           json:"v_dist_vgp,omitempty"`
}
func (DistanceVonGraefePhoria) TableName() string { return "distance_von_graefe_phoria" }
func (d *DistanceVonGraefePhoria) ToMap() map[string]interface{} {
	return map[string]interface{}{"id_distance_von_graefe_phoria": d.IDDistanceVonGraefePhoria, "h_dist_vgp": d.HDistVgp, "v_dist_vgp": d.VDistVgp}
}
