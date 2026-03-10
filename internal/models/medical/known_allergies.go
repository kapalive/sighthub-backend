// internal/models/medical/known_allergies.go
package medical

// KnownAllergies ↔ table: known_allergies
type KnownAllergies struct {
	IDKnownAllergies int64  `gorm:"column:id_known_allergies;primaryKey;autoIncrement" json:"id_known_allergies"`
	Title            string `gorm:"column:title;type:varchar(255);not null"            json:"title"`
	EyeExamID        int64  `gorm:"column:eye_exam_id;not null"                        json:"eye_exam_id"`
}

func (KnownAllergies) TableName() string { return "known_allergies" }

func (m *KnownAllergies) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_known_allergies": m.IDKnownAllergies,
		"title":              m.Title,
		"eye_exam_id":        m.EyeExamID,
	}
}
