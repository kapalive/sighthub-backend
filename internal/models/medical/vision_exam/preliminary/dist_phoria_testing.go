package preliminary
type DistPhoriaTest struct {
	IDDistPhoriaTest int     `gorm:"column:id_dist_phoria_testing;primaryKey;autoIncrement" json:"id_dist_phoria_testing"`
	Horiz            *string `gorm:"column:horiz;type:varchar(50)"                          json:"horiz,omitempty"`
	Vert             *string `gorm:"column:vert;type:varchar(50)"                           json:"vert,omitempty"`
	HorizExo         bool    `gorm:"column:horiz_exo;not null;default:false"                json:"horiz_exo"`
	HorizEso         bool    `gorm:"column:horiz_eso;not null;default:false"                json:"horiz_eso"`
	HorizOrtho       bool    `gorm:"column:horiz_ortho;not null;default:false"              json:"horiz_ortho"`
	VertRh           bool    `gorm:"column:vert_rh;not null;default:false"                  json:"vert_rh"`
	VertLn           bool    `gorm:"column:vert_ln;not null;default:false"                  json:"vert_ln"`
	VertOrtho        bool    `gorm:"column:vert_ortho;not null;default:false"               json:"vert_ortho"`
}
func (DistPhoriaTest) TableName() string { return "dist_phoria_testing" }
func (d *DistPhoriaTest) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_dist_phoria_testing": d.IDDistPhoriaTest,
		"horiz": d.Horiz, "vert": d.Vert,
		"horiz_exo": d.HorizExo, "horiz_eso": d.HorizEso, "horiz_ortho": d.HorizOrtho,
		"vert_rh": d.VertRh, "vert_ln": d.VertLn, "vert_ortho": d.VertOrtho,
	}
}
