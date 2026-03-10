// internal/models/diseases/diagnosis_level_6.go
package diseases

// DiagnosisLevel6 ↔ table: diagnosis_level_6 (ICD-10 level 6)
type DiagnosisLevel6 struct {
	IDLevel6    int64   `gorm:"column:id_level_6;primaryKey;autoIncrement"       json:"id_level_6"`
	DiagnosisID int64   `gorm:"column:diagnosis_id;not null"                     json:"diagnosis_id"`
	Code        string  `gorm:"column:code;type:varchar(10);not null;uniqueIndex" json:"code"`
	TitleLevel6 string  `gorm:"column:title_level_6;type:varchar(255);not null"  json:"title_level_6"`
	Description *string `gorm:"column:description;type:text"                     json:"description,omitempty"`

	Diagnosis *Diagnosis `gorm:"foreignKey:DiagnosisID;references:IDDiagnosis" json:"-"`
}

func (DiagnosisLevel6) TableName() string { return "diagnosis_level_6" }

func (d *DiagnosisLevel6) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_level_6":    d.IDLevel6,
		"diagnosis_id":  d.DiagnosisID,
		"code":          d.Code,
		"title_level_6": d.TitleLevel6,
		"description":   d.Description,
	}
}
