package marketing

type IntakeFormMedicalHistory struct {
	IDMedicalHistory int64  `gorm:"column:id_medical_history;primaryKey" json:"id_medical_history"`
	RequestID        int64  `gorm:"column:request_id;not null"           json:"request_id"`
	Name             string `gorm:"column:name;type:text;not null"       json:"name"`
	Value            bool   `gorm:"column:value;not null"                json:"value"`
}

func (IntakeFormMedicalHistory) TableName() string { return "intake_form_medical_history" }
