package integrations

import "time"

// IntegrationOAuthState stores temporary PKCE state for an in-progress OAuth flow.
type IntegrationOAuthState struct {
	IDIntegrationOAuthState int64     `gorm:"column:id_integration_oauth_state;primaryKey" json:"id_integration_oauth_state"`
	IntegrationCredentialID int64     `gorm:"column:integration_credential_id;not null"    json:"integration_credential_id"`
	State                   string    `gorm:"column:state;type:varchar(255);uniqueIndex"   json:"state"`
	CodeVerifier            string    `gorm:"column:code_verifier;type:varchar(255)"        json:"code_verifier"`
	CreatedAt               time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP"  json:"created_at"`
	ExpiresAt               time.Time `gorm:"column:expires_at"                             json:"expires_at"`

	Credential *IntegrationCredential `gorm:"foreignKey:IntegrationCredentialID;references:IDIntegrationCredential" json:"-"`
}

func (IntegrationOAuthState) TableName() string { return "integration_oauth_state" }
