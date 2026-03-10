package config

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

type Config struct {
	// App
	AppEnv string `json:"app_env"`
	Debug  bool   `json:"debug"`
	Port   string `json:"port"`

	// Secrets
	SecretKey        string `json:"secret_key"`
	JWTSecretKey     string `json:"jwt_secret_key"`
	RefreshSecretKey string `json:"refresh_secret_key"`

	// DB
	DBUsername string `json:"db_username"`
	DBPassword string `json:"db_password"`
	DBHost     string `json:"db_host"`
	DBPort     string `json:"db_port"`
	DBName     string `json:"db_name"`

	// Storage / URLs
	UploadAPIURL     string `json:"upload_api_url"`
	DownloadAPIURL   string `json:"download_api_url"`
	StorageServerURI string `json:"storage_server_uri"`

	// reCAPTCHA / Google
	RecaptchaSiteKey       string `json:"recaptcha_site_key"`
	RecaptchaSecretKey     string `json:"recaptcha_secret_key"`
	GoogleProjectID        string `json:"google_project_id"`
	GoogleApplicationCreds string `json:"google_application_credentials"`

	// Cookies
	CookieSameSite string `json:"cookie_samesite"`
	CookieSecure   bool   `json:"cookie_secure"`

	// JWT / Signing
	PrivateKeyPath string `json:"private_key_path"`
	Issuer         string `json:"issuer"`

	// SMTP / Email
	SMTPHost          string `json:"smtp_host"`
	SMTPPort          string `json:"smtp_port"`
	SMTPUsername      string `json:"smtp_username"`
	SMTPPassword      string `json:"smtp_password"`
	EmailSender       string `json:"email_sender"`
	AliasEmailSender  string `json:"alias_email_sender"`
	AliasEmailReciper string `json:"alias_email_reciper"`

	// SMS
	SMSApiURL     string `json:"sms_api_url"`
	SMSApiToken   string `json:"sms_api_token"`
	SMSHmacSecret string `json:"sms_hmac_secret"`

	// SRFax
	SRFaxURL       string `json:"srfax_url"`
	SRFaxAccountID string `json:"srfax_account_id"`
	SRFaxPassword  string `json:"srfax_password"`
	SRFaxCallerID  string `json:"srfax_caller_id"`

	// Twilio
	TwilioAccountSID          string `json:"twilio_account_sid"`
	TwilioAuthToken           string `json:"twilio_auth_token"`
	TwilioPhoneNumber         string `json:"twilio_phone_number"`
	TwilioCampaignSID         string `json:"twilio_campaign_sid"`
	TwilioMessagingServiceSID string `json:"twilio_messaging_service_sid"`
	TwilioBrandSID            string `json:"twilio_brand_sid"`

	// Helpdesk
	HelpdeskForwardURL  string `json:"helpdesk_forward_url"`
	HelpdeskForwardHMAC string `json:"helpdesk_forward_hmac"`
	HelpdeskReplyHMAC   string `json:"helpdesk_reply_hmac"`

	// PayPal
	PayPalClientID     string `json:"paypal_client_id"`
	PayPalClientSecret string `json:"paypal_client_secret"`
	PayPalMerchantID   string `json:"paypal_merchant_id"`

	// Sezzle / Afterpay
	SezzlePrivateKey   string `json:"sezzle_private_key"`
	SezzlePublicKey    string `json:"sezzle_public_key"`
	AfterpayPrivateKey string `json:"afterpay_private_key"`
	AfterpayMerchantID string `json:"afterpay_merchant_id"`

	// Telnyx
	TelnyxAPIKey     string `json:"telnyx_api_key"`
	TelnyxPublicKey  string `json:"telnyx_public_key"`
	TelnyxFromNumber string `json:"telnyx_from_number"`

	// Front/Back URLs
	BackendURL  string `json:"backend_url"`
	FrontendURL string `json:"frontend_url"`

	// Redis
	RedisAddr string `json:"redis_addr"`

	// Firebase Storage
	FirebaseBucket        string `json:"firebase_bucket"`
	FirebaseDefaultPrefix string `json:"firebase_default_prefix"`

	// Domains (raw строка из JSON, разбирается в Domains)
	DomainsRaw string   `json:"domains"`
	Domains    []string `json:"-"`

	// Zeiss OAuth
	ZeissAppBaseURL        string `json:"zeiss_app_base_url"`
	ZeissAuthorizeURL      string `json:"zeiss_authorize_url"`
	ZeissTokenURL          string `json:"zeiss_token_url"`
	ZeissPolicy            string `json:"zeiss_policy"`
	ZeissFrontRedirectPath string `json:"zeiss_front_redirect_path"`
	ZeissFrontCallbackPath string `json:"zeiss_front_callback_path"`
	ZeissClientID          string `json:"zeiss_client_id"`
	ZeissOAuthState        string `json:"zeiss_oauth_state"`
}

// LoadConfig читает config/config.json (production) или config/config.development.json
// в зависимости от переменной окружения APP_ENV.
// Можно явно указать файл через CONFIG_FILE.
func LoadConfig() (*Config, error) {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		appEnv := os.Getenv("APP_ENV")
		if appEnv == "development" {
			configFile = "config/config.development.json"
		} else {
			configFile = "config/config.json"
		}
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Printf("config file %q not found: %v", configFile, err)
		data = []byte("{}")
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if cfg.RedisAddr == "" {
		cfg.RedisAddr = "localhost:6379"
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
