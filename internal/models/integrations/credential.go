package integrations

import (
	"encoding/json"
	"time"
)

// IntegrationCredential stores OAuth / API credentials per provider per location.
// If LocationID is NULL the credential applies business-wide.
type IntegrationCredential struct {
	IDIntegrationCredential int64            `gorm:"column:id_integration_credential;primaryKey" json:"id_integration_credential"`
	LocationID              *int64           `gorm:"column:location_id"                          json:"location_id,omitempty"`
	Provider                string           `gorm:"column:provider;type:varchar(50);not null"   json:"provider"`
	ClientID                string           `gorm:"column:client_id;type:varchar(255);not null" json:"client_id"`
	ClientSecret            *string          `gorm:"column:client_secret;type:varchar(255)"      json:"client_secret,omitempty"`
	Config                  json.RawMessage  `gorm:"column:config;type:jsonb"                    json:"config,omitempty"`
	Active                  bool             `gorm:"column:active;default:true"                  json:"active"`
	CreatedAt               time.Time        `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt               time.Time        `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (IntegrationCredential) TableName() string { return "integration_credential" }

// ZeissConfig is the provider-specific config stored in the Config JSONB field.
type ZeissConfig struct {
	AuthorizeURL      string `json:"authorize_url"`
	TokenURL          string `json:"token_url"`
	Policy            string `json:"policy"`
	BaseURL           string `json:"base_url"`
	FrontRedirectPath string `json:"front_redirect_path"`
	FrontCallbackPath string `json:"front_callback_path"`
	OAuthState        string `json:"oauth_state"`
}

// ParseZeissConfig extracts Zeiss-specific config from the JSONB field.
func (c *IntegrationCredential) ParseZeissConfig() (*ZeissConfig, error) {
	var zc ZeissConfig
	if err := json.Unmarshal(c.Config, &zc); err != nil {
		return nil, err
	}
	// defaults
	if zc.AuthorizeURL == "" {
		zc.AuthorizeURL = "https://id-ip.zeiss.com/oauth2/v2.0/authorize"
	}
	if zc.TokenURL == "" {
		zc.TokenURL = "https://id-ip.zeiss.com/oauth2/v2.0/token"
	}
	if zc.Policy == "" {
		zc.Policy = "B2C_1A_ZeissIdNormalSignIn"
	}
	if zc.FrontRedirectPath == "" {
		zc.FrontRedirectPath = "/test"
	}
	if zc.FrontCallbackPath == "" {
		zc.FrontCallbackPath = "/zeiss/callback"
	}
	if zc.OAuthState == "" {
		zc.OAuthState = "zeiss_eyesync_oauth_state"
	}
	return &zc, nil
}
