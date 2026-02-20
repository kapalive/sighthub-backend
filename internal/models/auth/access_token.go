// internal/models/auth/access_token.go
package auth

import (
	"fmt"
	"time"
)

// AccessToken ⇄ access_token
type AccessToken struct {
	IDAccessToken          int        `gorm:"column:id_access_token;primaryKey"                                        json:"id_access_token"`
	Username               string     `gorm:"column:username;type:varchar(20);uniqueIndex;not null"                    json:"username"`
	AccessToken            *string    `gorm:"column:access_token;type:text"                                            json:"access_token,omitempty"`
	CurrentDatetimeAccess  *time.Time `gorm:"column:current_datetime_access;type:timestamptz;default:CURRENT_TIMESTAMP" json:"current_datetime_access,omitempty"`
	RefreshToken           *string    `gorm:"column:refresh_token;type:text"                                           json:"refresh_token,omitempty"`
	CurrentDatetimeRefresh *time.Time `gorm:"column:current_datetime_refresh;type:timestamptz;default:CURRENT_TIMESTAMP" json:"current_datetime_refresh,omitempty"`
}

func (AccessToken) TableName() string { return "access_token" }

// NewAccessToken — аналог Python __init__
func NewAccessToken(username string, accessToken, refreshToken *string) *AccessToken {
	now := time.Now().UTC()
	return &AccessToken{
		Username:               username,
		AccessToken:            accessToken,
		CurrentDatetimeAccess:  &now,
		RefreshToken:           refreshToken,
		CurrentDatetimeRefresh: &now,
	}
}

// UpdateAccessToken — аналог update_access_token(new_access_token, new_refresh_token=None)
func (t *AccessToken) UpdateAccessToken(newAccessToken string, newRefreshToken *string) {
	now := time.Now().UTC()
	t.AccessToken = &newAccessToken
	t.CurrentDatetimeAccess = &now
	if newRefreshToken != nil {
		t.RefreshToken = newRefreshToken
		t.CurrentDatetimeRefresh = &now
	}
}

// UpdateRefreshToken — аналог update_refresh_token(new_refresh_token)
func (t *AccessToken) UpdateRefreshToken(newRefreshToken string) {
	now := time.Now().UTC()
	t.RefreshToken = &newRefreshToken
	t.CurrentDatetimeRefresh = &now
}

// UpdateTokens — аналог update_tokens(new_access_token, new_refresh_token=None)
func (t *AccessToken) UpdateTokens(newAccessToken string, newRefreshToken *string) {
	now := time.Now().UTC()
	t.AccessToken = &newAccessToken
	t.CurrentDatetimeAccess = &now
	if newRefreshToken != nil {
		t.RefreshToken = newRefreshToken
		t.CurrentDatetimeRefresh = &now
	}
}

// ClearTokens — аналог clear_tokens()
func (t *AccessToken) ClearTokens() {
	t.AccessToken = nil
	t.CurrentDatetimeAccess = nil
	t.RefreshToken = nil
	t.CurrentDatetimeRefresh = nil
}

func (t *AccessToken) ToMap() map[string]interface{} {
	var accessAt, refreshAt interface{}
	if t.CurrentDatetimeAccess != nil && !t.CurrentDatetimeAccess.IsZero() {
		accessAt = t.CurrentDatetimeAccess.Format(time.RFC3339)
	} else {
		accessAt = nil
	}
	if t.CurrentDatetimeRefresh != nil && !t.CurrentDatetimeRefresh.IsZero() {
		refreshAt = t.CurrentDatetimeRefresh.Format(time.RFC3339)
	} else {
		refreshAt = nil
	}
	return map[string]interface{}{
		"id_access_token":          t.IDAccessToken,
		"username":                 t.Username,
		"access_token":             t.AccessToken,
		"current_datetime_access":  accessAt,
		"refresh_token":            t.RefreshToken,
		"current_datetime_refresh": refreshAt,
	}
}

func (t *AccessToken) String() string {
	return fmt.Sprintf("<AccessToken username=%s>", t.Username)
}
