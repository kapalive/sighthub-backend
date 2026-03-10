// internal/models/lab_ticket/lab_ticket_contact.go
package lab_ticket

import (
	"fmt"
	"sighthub-backend/internal/models/vendors"
)

// LabTicketContact ↔ table: lab_ticket_contact
type LabTicketContact struct {
	IDLabTicketContact                   int64   `gorm:"column:id_lab_ticket_contact;primaryKey;autoIncrement"                         json:"id_lab_ticket_contact"`
	LabTicketContactLensServicesID       int     `gorm:"column:lab_ticket_contact_lens_services_id;not null"                            json:"lab_ticket_contact_lens_services_id"`
	ODAnnualSupply                       bool    `gorm:"column:od_annual_supply;not null;default:true"                                  json:"od_annual_supply"`
	OSAnnualSupply                       bool    `gorm:"column:os_annual_supply;not null;default:true"                                  json:"os_annual_supply"`
	ODTotalQty                           *int    `gorm:"column:od_total_qty"                                                            json:"od_total_qty,omitempty"`
	OSTotalQty                           *int    `gorm:"column:os_total_qty"                                                            json:"os_total_qty,omitempty"`
	Reasons                              *string `gorm:"column:reasons;type:varchar(255)"                                               json:"reasons,omitempty"`
	Modality                             *string `gorm:"column:modality;type:varchar(255)"                                              json:"modality,omitempty"`
	BrandContactLensID                   *int64  `gorm:"column:brand_contact_lens_id"                                                   json:"brand_contact_lens_id,omitempty"`
	ManufacturerID                       *int64  `gorm:"column:manufacturer_id"                                                         json:"manufacturer_id,omitempty"`

	ContactLensService *LabTicketContactLensService  `gorm:"foreignKey:LabTicketContactLensServicesID;references:IDLabTicketContactLensServices" json:"-"`
	BrandContactLens   *vendors.BrandContactLens     `gorm:"foreignKey:BrandContactLensID;references:IDBrandContactLens"                        json:"-"`
	Manufacturer       *vendors.Manufacturer         `gorm:"foreignKey:ManufacturerID;references:IDManufacturer"                                json:"-"`
}

func (LabTicketContact) TableName() string { return "lab_ticket_contact" }

func (l *LabTicketContact) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_lab_ticket_contact":                  l.IDLabTicketContact,
		"lab_ticket_contact_lens_services_id":    l.LabTicketContactLensServicesID,
		"od_annual_supply":                       l.ODAnnualSupply,
		"os_annual_supply":                       l.OSAnnualSupply,
		"od_total_qty":                           l.ODTotalQty,
		"os_total_qty":                           l.OSTotalQty,
		"reasons":                                l.Reasons,
		"modality":                               l.Modality,
		"brand_contact_lens_id":                  l.BrandContactLensID,
		"manufacturer_id":                        l.ManufacturerID,
	}
	if l.ContactLensService != nil {
		m["contact_lens_service"] = l.ContactLensService.ToMap()
	}
	return m
}

func (l *LabTicketContact) String() string {
	return fmt.Sprintf("<LabTicketContact %d>", l.IDLabTicketContact)
}
