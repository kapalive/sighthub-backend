package service

type ProfessionalService struct {
	IDProfessionalService        int64    `gorm:"column:id_professional_service;primaryKey"      json:"id_professional_service"`
	ItemNumber                   string   `gorm:"column:item_number;size:30;not null"             json:"item_number"`
	CptHcpcsCode                 *string  `gorm:"column:cpt_hcpcs_code;size:12"                  json:"cpt_hcpcs_code,omitempty"`
	ProfessionalServiceScopeID   *int     `gorm:"column:professional_service_scope_id"           json:"professional_service_scope_id,omitempty"`
	ProfessionalServiceTypeID    *int     `gorm:"column:professional_service_type_id"            json:"professional_service_type_id,omitempty"`
	InvoiceDesc                  *string  `gorm:"column:invoice_desc;size:100"                   json:"invoice_desc,omitempty"`
	Price                        float64  `gorm:"column:price;type:numeric(10,2);default:0.00"   json:"price"`
	Cost                         float64  `gorm:"column:cost;type:numeric(10,2);default:0.00"    json:"cost"`
	Sort1                        *float64 `gorm:"column:sort1"                                   json:"sort1,omitempty"`
	Sort2                        *float64 `gorm:"column:sort2"                                   json:"sort2,omitempty"`
	ReferringPhysician           bool     `gorm:"column:referring_physician;default:false"        json:"referring_physician"`
	Visible                      bool     `gorm:"column:visible;default:true"                    json:"visible"`
	MfrNumber                    *string  `gorm:"column:mfr_number;size:60"                      json:"mfr_number,omitempty"`

	Scope *ProfessionalServiceScope `gorm:"foreignKey:ProfessionalServiceScopeID;references:IDProfessionalServiceScope" json:"scope,omitempty"`
	Type  *ProfessionalServiceType  `gorm:"foreignKey:ProfessionalServiceTypeID;references:IDMedicalServiceType"        json:"type,omitempty"`
}

func (ProfessionalService) TableName() string { return "professional_service" }
