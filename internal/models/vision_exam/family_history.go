// internal/models/vision_exam/family_history.go
package vision_exam

// FamilyHistory ↔ table: family_history
type FamilyHistory struct {
	IDFamilyHistory    int64   `gorm:"column:id_family_history;primaryKey;autoIncrement" json:"id_family_history"`
	Cataract           bool    `gorm:"column:cataract;not null;default:false"            json:"cataract"`
	Glaucoma           bool    `gorm:"column:glaucoma;not null;default:false"            json:"glaucoma"`
	MacularDegeneration bool   `gorm:"column:macular_degeneration;not null;default:false" json:"macular_degeneration"`
	Hypertension       bool    `gorm:"column:hypertension;not null;default:false"        json:"hypertension"`
	Diabetes           bool    `gorm:"column:diabetes;not null;default:false"            json:"diabetes"`
	Cancer             bool    `gorm:"column:cancer;not null;default:false"              json:"cancer"`
	HeartDisease       bool    `gorm:"column:heart_disease;not null;default:false"       json:"heart_disease"`
	Note               *string `gorm:"column:note;type:text"                             json:"note,omitempty"`
}

func (FamilyHistory) TableName() string { return "family_history" }

func (f *FamilyHistory) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_family_history":    f.IDFamilyHistory,
		"cataract":             f.Cataract,
		"glaucoma":             f.Glaucoma,
		"macular_degeneration": f.MacularDegeneration,
		"hypertension":         f.Hypertension,
		"diabetes":             f.Diabetes,
		"cancer":               f.Cancer,
		"heart_disease":        f.HeartDisease,
		"note":                 f.Note,
	}
}
