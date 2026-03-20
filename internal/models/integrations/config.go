package integrations

import "time"

type IntegrationConfig struct {
	IDIntegrationConfig  int64      `gorm:"column:id_integration_config;primaryKey" json:"id_integration_config"`
	IntegrationCompanyID int64      `gorm:"column:integration_company_id;not null"  json:"integration_company_id"`
	LocationID           int64      `gorm:"column:location_id;not null"             json:"location_id"`
	Username             string     `gorm:"column:username"                         json:"username"`
	Password             string     `gorm:"column:password"                         json:"password"`
	Connected            bool       `gorm:"column:connected;default:false"          json:"connected"`
	LastCheck            *time.Time `gorm:"column:last_check"                       json:"last_check"`
	CreatedAt            time.Time  `gorm:"column:created_at"                       json:"created_at"`
	UpdatedAt            time.Time  `gorm:"column:updated_at"                       json:"updated_at"`
}

func (IntegrationConfig) TableName() string { return "integration_config" }
