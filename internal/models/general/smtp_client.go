// internal/models/general/smtp_client.go
package general

import (
	"fmt"
	"time"
)

type SmtpClient struct {
	IDSmtpClient int       `gorm:"column:id_smtp_client;primaryKey;autoIncrement" json:"id_smtp_client"`
	SMTPHost     string    `gorm:"column:smtp_host;type:varchar(255);not null"    json:"smtp_host"`
	SMTPPort     int       `gorm:"column:smtp_port;not null;default:587"          json:"smtp_port"`
	SMTPUsername string    `gorm:"column:smtp_username;type:varchar(255);not null" json:"smtp_username"`
	SMTPPassword string    `gorm:"column:smtp_password;type:varchar(255);not null" json:"-"`
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
		"smtp_host":      s.SMTPHost,
		"smtp_port":      s.SMTPPort,
		"smtp_username":  s.SMTPUsername,
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
