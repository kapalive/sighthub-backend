package marketing

type IntakeFormAllergies struct {
	IDMedicalHistory int64  `gorm:"column:id_medical_history;primaryKey" json:"id_medical_history"`
	RequestID        int64  `gorm:"column:request_id;not null"           json:"request_id"`
	Name             string `gorm:"column:name;type:text;not null"       json:"name"`
}

func (IntakeFormAllergies) TableName() string { return "intake_form_allergies" }
