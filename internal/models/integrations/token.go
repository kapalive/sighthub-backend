package integrations

import "time"

// IntegrationToken stores OAuth tokens obtained for a given integration credential.
type IntegrationToken struct {
	IDIntegrationToken      int64     `gorm:"column:id_integration_token;primaryKey"       json:"id_integration_token"`
	IntegrationCredentialID int64     `gorm:"column:integration_credential_id;uniqueIndex" json:"integration_credential_id"`
	AccessToken             string    `gorm:"column:access_token;type:text"                json:"access_token"`
	RefreshToken            *string   `gorm:"column:refresh_token;type:text"               json:"refresh_token,omitempty"`
	ExpiresIn               *int64    `gorm:"column:expires_in"                            json:"expires_in,omitempty"`
	ObtainedAt              time.Time `gorm:"column:obtained_at"                           json:"obtained_at"`
	CreatedAt               time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP"  json:"created_at"`
	UpdatedAt               time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"  json:"updated_at"`

	Credential *IntegrationCredential `gorm:"foreignKey:IntegrationCredentialID;references:IDIntegrationCredential" json:"-"`
}

func (IntegrationToken) TableName() string { return "integration_token" }
