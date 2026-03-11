package service

type ProfessionalServiceType struct {
	IDMedicalServiceType int    `gorm:"column:id_medical_service_type;primaryKey" json:"id_medical_service_type"`
	Title                string `gorm:"column:title;size:50;not null"             json:"title"`
}

func (ProfessionalServiceType) TableName() string { return "professional_service_type" }
