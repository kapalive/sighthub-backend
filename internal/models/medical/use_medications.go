// internal/models/medical/use_medications.go
package medical

// UseMedications ↔ table: use_medications
type UseMedications struct {
	IDUseMedications int64   `gorm:"column:id_use_medications;primaryKey;autoIncrement" json:"id_use_medications"`
	Title            string  `gorm:"column:title;type:varchar(255);not null"            json:"title"`
	FormulationType  *string `gorm:"column:formulation_type;type:varchar(150)"          json:"formulation_type,omitempty"`
	Strength         *string `gorm:"column:strength;type:varchar(100)"                  json:"strength,omitempty"`
	EyeExamID        int64   `gorm:"column:eye_exam_id;not null"                        json:"eye_exam_id"`
}

func (UseMedications) TableName() string { return "use_medications" }

func (m *UseMedications) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_use_medications": m.IDUseMedications,
		"title":              m.Title,
		"formulation_type":   m.FormulationType,
		"strength":           m.Strength,
		"eye_exam_id":        m.EyeExamID,
	}
}
