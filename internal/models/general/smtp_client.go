// internal/models/general/smtp_client.go
package general

import (
	"fmt"
	"time"
)

type SmtpClient struct {
	IDSmtpClient int       `gorm:"column:id_smtp_client;primaryKey;autoIncrement" json:"id_smtp_client"`
	Label        string    `gorm:"column:label;type:varchar(255);not null;default:'Default'" json:"label"`
	NameKey      string    `gorm:"column:name_key;type:varchar(100);not null;default:'default'" json:"name_key"`
	LocationID   *int      `gorm:"column:location_id"                             json:"location_id"`
	SMTPHost     string    `gorm:"column:smtp_host;type:varchar(255);not null"    json:"smtp_host"`
	SMTPPort     int       `gorm:"column:smtp_port;not null;default:587"          json:"smtp_port"`
	SMTPUsername string    `gorm:"column:smtp_username;type:varchar(255);not null" json:"smtp_username"`
	SMTPPassword string    `gorm:"column:smtp_password;type:varchar(255);not null" json:"-"`
	FromDisplay  *string   `gorm:"column:from_display;type:varchar(255)"          json:"from_display"`
	SenderEmail  *string   `gorm:"column:sender_email;type:varchar(255)"          json:"sender_email"`
	UseTLS       bool      `gorm:"column:use_tls;not null;default:true"           json:"use_tls"`
	UseSSL       bool      `gorm:"column:use_ssl;not null;default:false"          json:"use_ssl"`
	Active       bool      `gorm:"column:active;not null;default:true"            json:"active"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"               json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"               json:"updated_at"`
}

func (SmtpClient) TableName() string { return "smtp_client" }

func (s *SmtpClient) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_smtp_client": s.IDSmtpClient,
		"label":          s.Label,
		"name_key":       s.NameKey,
		"location_id":    s.LocationID,
		"smtp_host":      s.SMTPHost,
		"smtp_port":      s.SMTPPort,
		"smtp_username":  s.SMTPUsername,
		"from_display":   s.FromDisplay,
		"sender_email":   s.SenderEmail,
		"use_tls":        s.UseTLS,
		"use_ssl":        s.UseSSL,
		"active":         s.Active,
		"created_at":     s.CreatedAt.Format(time.RFC3339),
		"updated_at":     s.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *SmtpClient) String() string {
	mode := "PLAIN"
	if s.UseTLS {
		mode = "TLS"
	} else if s.UseSSL {
		mode = "SSL"
	}
	return fmt.Sprintf("<SmtpClient %s:%d (%s)>", s.SMTPHost, s.SMTPPort, mode)
}
