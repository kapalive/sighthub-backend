// internal/models/general/helpdesk_ticket_screenshot.go
package general

import "time"

// HelpdeskTicketScreenshot ⇄ helpdesk_ticket_screenshot
type HelpdeskTicketScreenshot struct {
	IDHelpdeskTicketScreenshot int64     `gorm:"column:id_helpdesk_ticket_screenshot;primaryKey;autoIncrement" json:"id_helpdesk_ticket_screenshot"`
	TicketID                   int64     `gorm:"column:ticket_id;not null"                                      json:"ticket_id"`
	URL                        string    `gorm:"column:url;type:text;not null"                                  json:"url"`
	CreatedAt                  time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now()"      json:"created_at"`
}

func (HelpdeskTicketScreenshot) TableName() string { return "helpdesk_ticket_screenshot" }
