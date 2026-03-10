// internal/models/general/helpdesk_ticket.go
package general

import "time"

// HelpdeskTicket ⇄ helpdesk_ticket
type HelpdeskTicket struct {
	IDHelpdeskTicket int64      `gorm:"column:id_helpdesk_ticket;primaryKey;autoIncrement"    json:"id_helpdesk_ticket"`
	EmployeeLogin    string     `gorm:"column:employee_login;type:text;not null"              json:"employee_login"`
	Subject          string     `gorm:"column:subject;type:text;not null"                     json:"subject"`
	Description      string     `gorm:"column:description;type:text;not null"                 json:"description"`
	SourceDomain     *string    `gorm:"column:source_domain;type:text"                        json:"source_domain,omitempty"`
	Status           string     `gorm:"column:status;type:text;not null;default:'new'"        json:"status"`
	CreatedAt        time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:now()" json:"created_at"`
	ForwardedAt      *time.Time `gorm:"column:forwarded_at;type:timestamptz"                  json:"forwarded_at,omitempty"`
	ForwardOk        bool       `gorm:"column:forward_ok;not null;default:false"              json:"forward_ok"`
	ForwardError     *string    `gorm:"column:forward_error;type:text"                        json:"forward_error,omitempty"`
	ForwardResponse  *string    `gorm:"column:forward_response;type:jsonb"                    json:"forward_response,omitempty"`
}

func (HelpdeskTicket) TableName() string { return "helpdesk_ticket" }

func (h *HelpdeskTicket) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_helpdesk_ticket": h.IDHelpdeskTicket,
		"employee_login":     h.EmployeeLogin,
		"subject":            h.Subject,
		"description":        h.Description,
		"source_domain":      h.SourceDomain,
		"status":             h.Status,
		"created_at":         h.CreatedAt.Format(time.RFC3339),
		"forward_ok":         h.ForwardOk,
		"forward_error":      h.ForwardError,
	}
	if h.ForwardedAt != nil {
		m["forwarded_at"] = h.ForwardedAt.Format(time.RFC3339)
	} else {
		m["forwarded_at"] = nil
	}
	return m
}
