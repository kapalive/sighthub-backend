// internal/models/lab_ticket/lab_ticket.go
package lab_ticket

import (
	"fmt"
	"time"

	"sighthub-backend/internal/models/vendors"
)

// LabTicket ↔ table: lab_ticket
type LabTicket struct {
	IDLabTicket                   int64      `gorm:"column:id_lab_ticket;primaryKey;autoIncrement"                   json:"id_lab_ticket"`
	NumberTicket                  string     `gorm:"column:number_ticket;type:varchar(7);not null"                   json:"number_ticket"`
	GOrC                          *string    `gorm:"column:g_or_c;type:varchar(16)"                                  json:"g_or_c,omitempty"` // "g" | "c"
	DateCreate                    *time.Time `gorm:"column:date_create;type:date"                                    json:"date_create,omitempty"`
	DatePromise                   *time.Time `gorm:"column:date_promise;type:date"                                   json:"date_promise,omitempty"`
	ShipTo                        *string    `gorm:"column:ship_to;type:varchar(255)"                                json:"ship_to,omitempty"`
	LabID                         *int       `gorm:"column:lab_id"                                                   json:"lab_id,omitempty"`
	LabTicketStatusID             int64      `gorm:"column:lab_ticket_status_id;not null"                            json:"lab_ticket_status_id"`
	PatientID                     int64      `gorm:"column:patient_id;not null"                                      json:"patient_id"`
	OrdersLensID                  *int64     `gorm:"column:orders_lens_id"                                           json:"orders_lens_id,omitempty"`
	InvoiceID                     int64      `gorm:"column:invoice_id;not null"                                      json:"invoice_id"`
	Tray                          *string    `gorm:"column:tray;type:text"                                           json:"tray,omitempty"`
	Notified                      *string    `gorm:"column:notified;type:text"                                       json:"notified,omitempty"`
	Amt                           *string    `gorm:"column:amt;type:varchar(20)"                                     json:"amt,omitempty"`
	LabTicketPowersID             *int64     `gorm:"column:lab_ticket_powers_id"                                     json:"lab_ticket_powers_id,omitempty"`
	LabTicketLensID               *int64     `gorm:"column:lab_ticket_lens_id"                                       json:"lab_ticket_lens_id,omitempty"`
	LabTicketFrameID              *int64     `gorm:"column:lab_ticket_frame_id"                                      json:"lab_ticket_frame_id,omitempty"`
	LabTicketPowersContactID      *int64     `gorm:"column:lab_ticket_powers_contact_id"                             json:"lab_ticket_powers_contact_id,omitempty"`
	LabTicketContactID            *int64     `gorm:"column:lab_ticket_contact_id"                                    json:"lab_ticket_contact_id,omitempty"`
	OurNote                       *string    `gorm:"column:our_note;type:text"                                       json:"our_note,omitempty"`
	LabInstructions               *string    `gorm:"column:lab_instructions;type:text"                               json:"lab_instructions,omitempty"`
	EmployeeID                    int64      `gorm:"column:employee_id;not null"                                     json:"employee_id"`
	VwOrderID                     *string    `gorm:"column:vw_order_id;type:varchar(100)"                            json:"vw_order_id,omitempty"`

	// preload relations (lab_id now references vendor where lab=true)
	Lab                  *vendors.Vendor       `gorm:"foreignKey:LabID;references:IDVendor"                                        json:"-"`
	LabTicketStatus      *LabTicketStatus      `gorm:"foreignKey:LabTicketStatusID;references:IDLabTicketStatus"                   json:"-"`
	Powers               *LabTicketPowers      `gorm:"foreignKey:LabTicketPowersID;references:IDLabTicketPowers"                   json:"-"`
	Lens                 *LabTicketLens        `gorm:"foreignKey:LabTicketLensID;references:IDLabTicketLens"                       json:"-"`
	Frame                *LabTicketFrame       `gorm:"foreignKey:LabTicketFrameID;references:IDLabTicketFrame"                     json:"-"`
	PowersContact        *LabTicketPowersContact `gorm:"foreignKey:LabTicketPowersContactID;references:IDLabTicketPowersContact"   json:"-"`
	Contact              *LabTicketContact     `gorm:"foreignKey:LabTicketContactID;references:IDLabTicketContact"                 json:"-"`
}

func (LabTicket) TableName() string { return "lab_ticket" }

func (l *LabTicket) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_lab_ticket":                l.IDLabTicket,
		"number_ticket":                l.NumberTicket,
		"g_or_c":                       l.GOrC,
		"lab_id":                       l.LabID,
		"lab_ticket_status_id":         l.LabTicketStatusID,
		"patient_id":                   l.PatientID,
		"orders_lens_id":               l.OrdersLensID,
		"invoice_id":                   l.InvoiceID,
		"tray":                         l.Tray,
		"notified":                     l.Notified,
		"amt":                          l.Amt,
		"lab_ticket_powers_id":         l.LabTicketPowersID,
		"lab_ticket_lens_id":           l.LabTicketLensID,
		"lab_ticket_frame_id":          l.LabTicketFrameID,
		"lab_ticket_powers_contact_id": l.LabTicketPowersContactID,
		"lab_ticket_contact_id":        l.LabTicketContactID,
		"our_note":                     l.OurNote,
		"lab_instructions":             l.LabInstructions,
		"employee_id":                  l.EmployeeID,
		"vw_order_id":                  l.VwOrderID,
	}
	if l.DateCreate != nil {
		d := l.DateCreate.Format("2006-01-02")
		m["date_create"] = d
	} else {
		m["date_create"] = nil
	}
	if l.DatePromise != nil {
		d := l.DatePromise.Format("2006-01-02")
		m["date_promise"] = d
	} else {
		m["date_promise"] = nil
	}
	if l.LabTicketStatus != nil {
		m["lab_ticket_status"] = l.LabTicketStatus.ToMap()
	}
	if l.Lab != nil {
		m["lab"] = l.Lab.ToMap()
	}
	if l.Powers != nil {
		m["powers"] = l.Powers.ToMap()
	}
	if l.Lens != nil {
		m["lens"] = l.Lens.ToMap()
	}
	if l.Frame != nil {
		m["frame"] = l.Frame.ToMap()
	}
	if l.PowersContact != nil {
		m["powers_contact"] = l.PowersContact.ToMap()
	}
	if l.Contact != nil {
		m["contact"] = l.Contact.ToMap()
	}
	return m
}

func (l *LabTicket) String() string {
	return fmt.Sprintf("<LabTicket %s>", l.NumberTicket)
}
