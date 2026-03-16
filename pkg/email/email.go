// pkg/email/email.go
// Аналог utils_email.py — SMTP-рассылка + рендеринг HTML шаблонов
package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

// SMTPConfig представляет строку из таблицы smtp_client.
type SMTPConfig struct {
	SmtpHost       string
	SmtpPort       int
	SmtpUsername   string
	SmtpPassword   string
	UseSSL         bool
	UseTLS         bool
	SenderEmail    string
	FromDisplay    string
}

// smtpClientRow — минимальная GORM-модель для выборки smtp_client.
type smtpClientRow struct {
	SmtpHost     string  `gorm:"column:smtp_host"`
	SmtpPort     int     `gorm:"column:smtp_port"`
	SmtpUsername string  `gorm:"column:smtp_username"`
	SmtpPassword string  `gorm:"column:smtp_password"`
	UseSSL       bool    `gorm:"column:use_ssl"`
	UseTLS       bool    `gorm:"column:use_tls"`
	SenderEmail  *string `gorm:"column:sender_email"`
	FromDisplay  *string `gorm:"column:from_display"`
	LocationID   *int64  `gorm:"column:location_id"`
}

func (smtpClientRow) TableName() string { return "smtp_client" }

// GetSMTPForLocation возвращает SMTP-конфиг для локации (с fallback на глобальный).
func GetSMTPForLocation(db *gorm.DB, locationID *int64) (*SMTPConfig, error) {
	var row smtpClientRow

	if locationID != nil {
		err := db.Where("active = true AND location_id = ?", *locationID).
			First(&row).Error
		if err == nil {
			return toConfig(row), nil
		}
	}

	// Глобальный fallback (location_id IS NULL)
	err := db.Where("active = true AND location_id IS NULL").First(&row).Error
	if err != nil {
		return nil, fmt.Errorf(
			"business mail is not configured. Please set up SMTP in Settings",
		)
	}
	return toConfig(row), nil
}

func toConfig(r smtpClientRow) *SMTPConfig {
	cfg := &SMTPConfig{
		SmtpHost:     r.SmtpHost,
		SmtpPort:     r.SmtpPort,
		SmtpUsername: r.SmtpUsername,
		SmtpPassword: r.SmtpPassword,
		UseSSL:       r.UseSSL,
		UseTLS:       r.UseTLS,
	}
	if r.SenderEmail != nil {
		cfg.SenderEmail = *r.SenderEmail
	} else {
		cfg.SenderEmail = r.SmtpUsername
	}
	if r.FromDisplay != nil {
		cfg.FromDisplay = *r.FromDisplay
	}
	return cfg
}

// GetTemplateForCategory возвращает имя файла шаблона (.html) для категории.
// Приоритет: org_email_template → email_template (is_default) → "default.html".
func GetTemplateForCategory(db *gorm.DB, category string) string {
	var name string

	// Ищем org-выбор
	err := db.Raw(`
		SELECT et.name
		FROM org_email_template ot
		JOIN email_template et ON et.id_email_template = ot.template_id
		WHERE ot.category = ?
		LIMIT 1`, category).Scan(&name).Error
	if err == nil && name != "" {
		return name + ".html"
	}

	// Дефолтный шаблон для категории
	err = db.Raw(`
		SELECT name FROM email_template
		WHERE category = ? AND is_default = true
		LIMIT 1`, category).Scan(&name).Error
	if err == nil && name != "" {
		return name + ".html"
	}

	return "default.html"
}

// RenderTemplate рендерит HTML шаблон из директории templates/email/.
// templatesDir — абсолютный путь к папке с шаблонами.
func RenderTemplate(templatesDir, templateFile string, ctx map[string]interface{}) (string, error) {
	path := filepath.Join(templatesDir, templateFile)
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return "", fmt.Errorf("parse template %q: %w", templateFile, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("render template %q: %w", templateFile, err)
	}
	return buf.String(), nil
}

// Send отправляет HTML-письмо через SMTP.
func Send(cfg *SMTPConfig, to, subject, htmlBody string) error {
	if to == "" {
		return fmt.Errorf("to email is empty")
	}

	from := cfg.SenderEmail
	if cfg.FromDisplay != "" {
		from = fmt.Sprintf("%s <%s>", cfg.FromDisplay, cfg.SenderEmail)
	}

	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=utf-8\r\n\r\n",
		from, to, subject,
	)
	msg := []byte(headers + htmlBody)

	addr := fmt.Sprintf("%s:%d", cfg.SmtpHost, cfg.SmtpPort)
	auth := smtp.PlainAuth("", cfg.SmtpUsername, cfg.SmtpPassword, cfg.SmtpHost)

	if cfg.UseSSL {
		tlsCfg := &tls.Config{ServerName: cfg.SmtpHost}
		conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 15 * time.Second}, "tcp", addr, tlsCfg)
		if err != nil {
			return fmt.Errorf("tls dial: %w", err)
		}
		client, err := smtp.NewClient(conn, cfg.SmtpHost)
		if err != nil {
			return fmt.Errorf("smtp new client: %w", err)
		}
		defer client.Close()
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
		if err := client.Mail(cfg.SenderEmail); err != nil {
			return err
		}
		if err := client.Rcpt(to); err != nil {
			return err
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write(msg)
		w.Close()
		return err
	}

	// Plain / STARTTLS
	return smtp.SendMail(addr, auth, cfg.SenderEmail, []string{to}, msg)
}

// SendViaDB — высокоуровневая обёртка: берёт SMTP из БД, рендерит шаблон и отправляет.
func SendViaDB(
	db *gorm.DB,
	to, subject, templateFile string,
	ctx map[string]interface{},
	locationID *int64,
) error {
	cfg, err := GetSMTPForLocation(db, locationID)
	if err != nil {
		return err
	}

	templatesDir := os.Getenv("EMAIL_TEMPLATES_DIR")
	if templatesDir == "" {
		// дефолтный путь относительно бинаря
		templatesDir = "internal/templates/email"
	}

	body, err := RenderTemplate(templatesDir, templateFile, ctx)
	if err != nil {
		return err
	}

	return Send(cfg, to, subject, body)
}
