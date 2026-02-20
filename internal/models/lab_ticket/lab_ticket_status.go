// internal/models/lab_ticket/lab_ticket_status.go
package lab_ticket

import "fmt"

// LabTicketStatus ↔ table: lab_ticket_status
type LabTicketStatus struct {
	IDLabTicketStatus int64  `gorm:"column:id_lab_ticket_status;primaryKey;autoIncrement" json:"id_lab_ticket_status"`
	TicketStatus      string `gorm:"column:ticket_status;type:text;not null"            json:"ticket_status"`
}

func (LabTicketStatus) TableName() string { return "lab_ticket_status" }

func (l *LabTicketStatus) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_lab_ticket_status": l.IDLabTicketStatus,
		"ticket_status":        l.TicketStatus,
	}
}

func (l *LabTicketStatus) String() string {
	return fmt.Sprintf("<LabTicketStatus %s>", l.TicketStatus)
}
