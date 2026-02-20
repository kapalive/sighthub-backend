// internal/models/lab_ticket/lab_ticket_powers_contact.go
package lab_ticket

import (
	"fmt"
	"time"
)

// Обертка под Postgres ENUM contact_lens_type_enum
type ContactLensType string

// LabTicketPowersContact ↔ table: lab_ticket_powers_contact
type LabTicketPowersContact struct {
	IDLabTicketPowersContact int64 `gorm:"column:id_lab_ticket_powers_contact;primaryKey;autoIncrement" json:"id_lab_ticket_powers_contact"`

	ODContLens *string `gorm:"column:od_cont_lens;type:varchar(200)" json:"od_cont_lens,omitempty"`
	OSContLens *string `gorm:"column:os_cont_lens;type:varchar(200)" json:"os_cont_lens,omitempty"`

	ODBC *string `gorm:"column:od_bc;type:varchar(20)" json:"od_bc,omitempty"`
	OSBC *string `gorm:"column:os_bc;type:varchar(20)" json:"os_bc,omitempty"`

	ODDia *float64 `gorm:"column:od_dia;type:numeric(5,2)" json:"od_dia,omitempty"`
	OSDia *float64 `gorm:"column:os_dia;type:numeric(5,2)" json:"os_dia,omitempty"`

	ODPwr *string `gorm:"column:od_pwr;type:varchar(20)" json:"od_pwr,omitempty"`
	OSPwr *string `gorm:"column:os_pwr;type:varchar(20)" json:"os_pwr,omitempty"`

	ODCyl *string `gorm:"column:od_cyl;type:varchar(20)" json:"od_cyl,omitempty"`
	OSCyl *string `gorm:"column:os_cyl;type:varchar(20)" json:"os_cyl,omitempty"`

	ODAxis *string `gorm:"column:od_axis;type:varchar(20)" json:"od_axis,omitempty"`
	OSAxis *string `gorm:"column:os_axis;type:varchar(20)" json:"os_axis,omitempty"`

	ODAdd *string `gorm:"column:od_add;type:varchar(20)" json:"od_add,omitempty"`
	OSAdd *string `gorm:"column:os_add;type:varchar(20)" json:"os_add,omitempty"`

	ODColor *string `gorm:"column:od_color;type:varchar(100)" json:"od_color,omitempty"`
	OSColor *string `gorm:"column:os_color;type:varchar(100)" json:"os_color,omitempty"`

	ODType *ContactLensType `gorm:"column:od_type;type:contact_lens_type_enum" json:"od_type,omitempty"`
	OSType *ContactLensType `gorm:"column:os_type;type:contact_lens_type_enum" json:"os_type,omitempty"`

	ExpirationDate *time.Time `gorm:"column:expiration_date;type:date" json:"expiration_date,omitempty"`

	ODHPrismDirection *HPrismDirection `gorm:"column:od_h_prism_direction;type:h_prism_direction_enum" json:"od_h_prism_direction,omitempty"`
	OSHPrismDirection *HPrismDirection `gorm:"column:os_h_prism_direction;type:h_prism_direction_enum" json:"os_h_prism_direction,omitempty"`
	ODVPrismDirection *VPrismDirection `gorm:"column:od_v_prism_direction;type:v_prism_direction_enum" json:"od_v_prism_direction,omitempty"`
	OSVPrismDirection *VPrismDirection `gorm:"column:os_v_prism_direction;type:v_prism_direction_enum" json:"os_v_prism_direction,omitempty"`
}

func (LabTicketPowersContact) TableName() string { return "lab_ticket_powers_contact" }

func (l *LabTicketPowersContact) ToMap() map[string]interface{} {
	var exp *string
	if l.ExpirationDate != nil {
		s := l.ExpirationDate.Format("2006-01-02")
		exp = &s
	}
	return map[string]interface{}{
		"id_lab_ticket_powers_contact": l.IDLabTicketPowersContact,
		"od_cont_lens":                 l.ODContLens,
		"os_cont_lens":                 l.OSContLens,
		"od_bc":                        l.ODBC,
		"os_bc":                        l.OSBC,
		"od_dia":                       l.ODDia,
		"os_dia":                       l.OSDia,
		"od_pwr":                       l.ODPwr,
		"os_pwr":                       l.OSPwr,
		"od_cyl":                       l.ODCyl,
		"os_cyl":                       l.OSCyl,
		"od_axis":                      l.ODAxis,
		"os_axis":                      l.OSAxis,
		"od_add":                       l.ODAdd,
		"os_add":                       l.OSAdd,
		"od_color":                     l.ODColor,
		"os_color":                     l.OSColor,
		"od_type":                      l.ODType,
		"os_type":                      l.OSType,
		"expiration_date":              exp,
		"od_h_prism_direction":         l.ODHPrismDirection,
		"os_h_prism_direction":         l.OSHPrismDirection,
		"od_v_prism_direction":         l.ODVPrismDirection,
		"os_v_prism_direction":         l.OSVPrismDirection,
	}
}

func (l *LabTicketPowersContact) String() string {
	return fmt.Sprintf("<LabTicketPowersContact %d>", l.IDLabTicketPowersContact)
}
