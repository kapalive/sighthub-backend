package preliminary
type NearPhoriaTest struct {
	IDNearPhoriaTest int     `gorm:"column:id_near_phoria_testing;primaryKey;autoIncrement" json:"id_near_phoria_testing"`
	Horiz            *string `gorm:"column:horiz;type:varchar(50)"                          json:"horiz,omitempty"`
	Vert             *string `gorm:"column:vert;type:varchar(50)"                           json:"vert,omitempty"`
	GradientRatio1   *string `gorm:"column:gradient_ratio1;type:varchar(50)"                json:"gradient_ratio1,omitempty"`
	CalculatedRatio1 *string `gorm:"column:calculated_ratio1;type:varchar(50)"              json:"calculated_ratio1,omitempty"`
	GradientRatio2   *string `gorm:"column:gradient_ratio2;type:varchar(50)"                json:"gradient_ratio2,omitempty"`
	CalculatedRatio2 *string `gorm:"column:calculated_ratio2;type:varchar(50)"              json:"calculated_ratio2,omitempty"`
	HorizExo         bool    `gorm:"column:horiz_exo;not null;default:false"                json:"horiz_exo"`
	HorizEso         bool    `gorm:"column:horiz_eso;not null;default:false"                json:"horiz_eso"`
	HorizOrtho       bool    `gorm:"column:horiz_ortho;not null;default:false"              json:"horiz_ortho"`
	VertRh           bool    `gorm:"column:vert_rh;not null;default:false"                  json:"vert_rh"`
	VertLn           bool    `gorm:"column:vert_ln;not null;default:false"                  json:"vert_ln"`
	VertOrtho        bool    `gorm:"column:vert_ortho;not null;default:false"               json:"vert_ortho"`
}
func (NearPhoriaTest) TableName() string { return "near_phoria_testing" }
func (n *NearPhoriaTest) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_near_phoria_testing": n.IDNearPhoriaTest,
		"horiz": n.Horiz, "vert": n.Vert,
		"gradient_ratio1": n.GradientRatio1, "calculated_ratio1": n.CalculatedRatio1,
		"gradient_ratio2": n.GradientRatio2, "calculated_ratio2": n.CalculatedRatio2,
		"horiz_exo": n.HorizExo, "horiz_eso": n.HorizEso, "horiz_ortho": n.HorizOrtho,
		"vert_rh": n.VertRh, "vert_ln": n.VertLn, "vert_ortho": n.VertOrtho,
	}
}
