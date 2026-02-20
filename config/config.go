// internal/config/config.go
package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config — все переменные окружения, включая то, что у тебя в .env.
// Ничего не убирал, только добавил новые поля.
type Config struct {
	// Secrets
	SecretKey        string
	JWTSecretKey     string
	RefreshSecretKey string

	// DB
	DBUsername string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string

	// HTTP
	Port string

	// URLs (старые поля)
	UploadAPIURL     string
	DownloadAPIURL   string
	StorageServerURI string

	// Recaptcha / Google (старые поля)
	RecaptchaSiteKey       string
	RecaptchaSecretKey     string
	GoogleProjectID        string
	GoogleApplicationCreds string // GOOGLE_APPLICATION_CREDENTIALS (путь к JSON)
	FlaskEnv               string

	// Cookies
	CookieSameSite string
	CookieSecure   bool

	// New: PayPal
	PayPalClientID     string
	PayPalClientSecret string
	PayPalMerchantID   string

	// New: domains (разобраны в слайс)
	DomainsRaw string   // исходная строка из .env
	Domains    []string // разобранный список доменов

	// New: Sezzle / Afterpay
	SezzlePrivateKey   string
	SezzlePublicKey    string
	AfterpayPrivateKey string
	AfterpayMerchantID string

	// New: SMTP / email
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	EmailSender  string

	AliasEmailSender  string
	AliasEmailReciper string

	// New: Telnyx
	TelnyxAPIKey     string
	TelnyxPublicKey  string
	TelnyxFromNumber string

	// New: Front/Back URLs
	BackendURL  string
	FrontendURL string

	// New: Firebase Storage
	FirebaseBucket        string // FIREBASE_BUCKET (например arts-of-optics-b919b.appspot.com)
	FirebaseDefaultPrefix string // опционально: FIREBASE_DEFAULT_PREFIX
}

// LoadConfig — грузит .env (если есть рядом/в WD) и собирает конфиг
func LoadConfig() (*Config, error) {
	// Загружаем .env (если нет — просто используем переменные окружения)
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using environment variables")
	}

	cfg := &Config{
		// Secrets
		SecretKey:        os.Getenv("SECRET_KEY"),
		JWTSecretKey:     os.Getenv("JWT_SECRET_KEY"),
		RefreshSecretKey: os.Getenv("REFRESH_SECRET_KEY"),

		// DB
		DBUsername: os.Getenv("DB_USERNAME"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBName:     os.Getenv("DB_NAME"),

		// HTTP
		Port: os.Getenv("PORT"),

		// URLs (старые поля)
		UploadAPIURL:     os.Getenv("UPLOAD_API_URL"),
		DownloadAPIURL:   os.Getenv("DOWNLOAD_API_URL"),
		StorageServerURI: os.Getenv("STORAGE_SERVER_URI"),

		// Recaptcha / Google (старые поля)
		RecaptchaSiteKey:       os.Getenv("RECAPTCHA_SITE_KEY"),
		RecaptchaSecretKey:     os.Getenv("RECAPTCHA_SECRET_KEY"),
		GoogleProjectID:        os.Getenv("GOOGLE_PROJECT_ID"),
		GoogleApplicationCreds: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		FlaskEnv:               os.Getenv("FLASK_ENV"),

		// Cookies
		CookieSameSite: os.Getenv("COOKIE_SAMESITE"),
		CookieSecure:   os.Getenv("COOKIE_SECURE") == "true",

		// PayPal
		PayPalClientID:     os.Getenv("PAYPAL_CLIENT_ID"),
		PayPalClientSecret: os.Getenv("PAYPAL_CLIENT_SECRET"),
		PayPalMerchantID:   os.Getenv("PAYPAL_MERCHANT_ID"),

		// Domains
		DomainsRaw: os.Getenv("DOMAINS"),

		// Sezzle / Afterpay
		SezzlePrivateKey:   os.Getenv("SEZZLE_PRIVATE_KEY"),
		SezzlePublicKey:    os.Getenv("SEZZLE_PUBLIC_KEY"),
		AfterpayPrivateKey: os.Getenv("AFTERPAY_PRIVATE_KEY"),
		AfterpayMerchantID: os.Getenv("AFTERPAY_MERCHANT_ID"),

		// SMTP / email
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		EmailSender:  os.Getenv("EMAIL_SENDER"),

		AliasEmailSender:  os.Getenv("ALIAS_EMAIL_SENDER"),
		AliasEmailReciper: os.Getenv("ALIAS_EMAIL_RECIPER"),

		// Telnyx
		TelnyxAPIKey:     os.Getenv("TELNYX_API_KEY"),
		TelnyxPublicKey:  os.Getenv("TELNYX_PUBLIC_KEY"),
		TelnyxFromNumber: os.Getenv("TELNYX_FROM_NUMBER"),

		// Front/Back
		BackendURL:  os.Getenv("BACKEND_URL"),
		FrontendURL: os.Getenv("FRONTEND_URL"),

		// Firebase
		FirebaseBucket:        os.Getenv("FIREBASE_BUCKET"),
		FirebaseDefaultPrefix: os.Getenv("FIREBASE_DEFAULT_PREFIX"),
	}

	// Разобрать домены в массив
	if cfg.DomainsRaw != "" {
		parts := strings.Split(cfg.DomainsRaw, ",")
		cfg.Domains = make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				cfg.Domains = append(cfg.Domains, p)
			}
		}
	}

	return cfg, nil
}
