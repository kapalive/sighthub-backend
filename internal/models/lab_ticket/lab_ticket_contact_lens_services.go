// internal/models/lab_ticket/lab_ticket_contact_lens_services.go
package lab_ticket

import "fmt"

// LabTicketContactLensService ↔ table: lab_ticket_contact_lens_services
type LabTicketContactLensService struct {
	IDLabTicketContactLensServices int64  `gorm:"column:id_lab_ticket_contact_lens_services;primaryKey;autoIncrement" json:"id_lab_ticket_contact_lens_services"`
	ServicesName                   string `gorm:"column:services_name;type:varchar(255);not null"                          json:"services_name"`
}

func (LabTicketContactLensService) TableName() string { return "lab_ticket_contact_lens_services" }

func (l *LabTicketContactLensService) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_lab_ticket_contact_lens_services": l.IDLabTicketContactLensServices,
		"services_name":                       l.ServicesName,
	}
}

func (l *LabTicketContactLensService) String() string {
	return fmt.Sprintf("<LabTicketContactLensService %d: %s>", l.IDLabTicketContactLensServices, l.ServicesName)
}
