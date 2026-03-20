package integrations

import "time"

type IntegrationCompany struct {
	IDIntegrationCompany int64     `gorm:"column:id_integration_company;primaryKey" json:"id_integration_company"`
	Name                 string    `gorm:"column:name"                              json:"name"`
	Code                 string    `gorm:"column:code"                              json:"code"`
	AuthURL              string    `gorm:"column:auth_url"                          json:"auth_url"`
	APIBaseURL           string    `gorm:"column:api_base_url"                      json:"api_base_url"`
	Active               bool      `gorm:"column:active;default:true"               json:"active"`
	CreatedAt            time.Time `gorm:"column:created_at"                        json:"created_at"`
}

func (IntegrationCompany) TableName() string { return "integration_company" }
