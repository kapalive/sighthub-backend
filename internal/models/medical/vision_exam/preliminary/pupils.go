package preliminary
type Pupils struct {
	IDPupils   int64   `gorm:"column:id_pupils;primaryKey;autoIncrement"  json:"id_pupils"`
	OdMmDim    *string `gorm:"column:od_mm_dim;type:varchar(255)"          json:"od_mm_dim,omitempty"`
	OdMmBright *string `gorm:"column:od_mm_bright;type:varchar(255)"       json:"od_mm_bright,omitempty"`
	OsMmDim    *string `gorm:"column:os_mm_dim;type:varchar(255)"          json:"os_mm_dim,omitempty"`
	OsMmBright *string `gorm:"column:os_mm_bright;type:varchar(255)"       json:"os_mm_bright,omitempty"`
	Perrla     bool    `gorm:"column:perrla;not null;default:false"        json:"perrla"`
	PerrlaText *string `gorm:"column:perrla_text;type:varchar(255)"        json:"perrla_text,omitempty"`
	Apd        bool    `gorm:"column:apd;not null;default:false"           json:"apd"`
	ApdText    *string `gorm:"column:apd_text;type:varchar(255)"           json:"apd_text,omitempty"`
}
func (Pupils) TableName() string { return "pupils" }
func (p *Pupils) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_pupils": p.IDPupils, "od_mm_dim": p.OdMmDim, "od_mm_bright": p.OdMmBright,
		"os_mm_dim": p.OsMmDim, "os_mm_bright": p.OsMmBright,
		"perrla": p.Perrla, "perrla_text": p.PerrlaText, "apd": p.Apd, "apd_text": p.ApdText,
	}
}
