package service

type ProfessionalServiceScope struct {
	IDProfessionalServiceScope int    `gorm:"column:id_professional_service_scope;primaryKey" json:"id_professional_service_scope"`
	Title                      string `gorm:"column:title;size:50;not null"                   json:"title"`
}

func (ProfessionalServiceScope) TableName() string { return "professional_service_scope" }
