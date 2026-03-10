// internal/models/medical/medical_record.go
package medical

// MedicalRecord ↔ table: medical_record
type MedicalRecord struct {
	IDMedicalRecord       int64   `gorm:"column:id_medical_record;primaryKey;autoIncrement" json:"id_medical_record"`
	Occupation            *string `gorm:"column:occupation;type:varchar(255)"               json:"occupation,omitempty"`
	PersistingPatientNote *string `gorm:"column:persisting_patient_note;type:varchar(255)"  json:"persisting_patient_note,omitempty"`
}

func (MedicalRecord) TableName() string { return "medical_record" }

func (m *MedicalRecord) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_medical_record":       m.IDMedicalRecord,
		"occupation":              m.Occupation,
		"persisting_patient_note": m.PersistingPatientNote,
	}
}
