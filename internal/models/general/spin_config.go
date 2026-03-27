// internal/models/general/spin_config.go
package general

import "time"

// SpinConfig ⇄ spin_config
type SpinConfig struct {
	IDSpinConfig int64     `gorm:"column:id_spin_config;primaryKey;autoIncrement"                  json:"id_spin_config"`
	LocationID   int64     `gorm:"column:location_id;not null;uniqueIndex;index"                   json:"location_id"`
	SpinURL      *string   `gorm:"column:spin_url;type:text"                                       json:"spin_url,omitempty"`
	TimeoutSec   int       `gorm:"column:timeout_sec;not null;default:125"                         json:"timeout_sec"`
	AuthKey      string    `gorm:"column:auth_key;type:text;not null"                              json:"auth_key"`
	Active       bool      `gorm:"column:active;not null;default:true"                             json:"active"`
	IsSandbox    bool      `gorm:"column:is_sandbox;not null;default:true"                         json:"is_sandbox"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now()"       json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:now()"       json:"updated_at"`
	MerchantID   *string   `gorm:"column:merchant_id;type:varchar(64)"                             json:"merchant_id,omitempty"`
}

func (SpinConfig) TableName() string { return "spin_config" }

func (s *SpinConfig) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_spin_config": s.IDSpinConfig,
		"location_id":    s.LocationID,
		"spin_url":       s.SpinURL,
		"timeout_sec":    s.TimeoutSec,
		"auth_key":       s.AuthKey,
		"active":         s.Active,
		"created_at":     s.CreatedAt.Format(time.RFC3339),
		"updated_at":     s.UpdatedAt.Format(time.RFC3339),
		"merchant_id":    s.MerchantID,
	}
}
