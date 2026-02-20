// internal/models/general/service_tax_by_state.go
package general

import (
	"fmt"
	"time"
)

type ServiceTaxByState struct {
	IDServiceTax  int       `gorm:"column:id_service_tax;primaryKey"       json:"id_service_tax"`
	StateCode     string    `gorm:"column:state_code;type:varchar(2);not null"   json:"state_code"`
	StateName     string    `gorm:"column:state_name;type:varchar(30);not null"  json:"state_name"`
	TaxPercent    float64   `gorm:"column:tax_percent;type:numeric(5,4);not null" json:"tax_percent"`
	EffectiveDate time.Time `gorm:"column:effective_date;type:date;not null"     json:"effective_date"`
	TaxActive     bool      `gorm:"column:tax_active;not null;default:true"      json:"tax_active"`
	CountryID     int       `gorm:"column:country_id;not null"                   json:"country_id"`
}

func (ServiceTaxByState) TableName() string { return "service_tax_by_state" }

func (s *ServiceTaxByState) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_service_tax": s.IDServiceTax,
		"state_code":     s.StateCode,
		"state_name":     s.StateName,
		"tax_percent":    fmt.Sprintf("%.4f", s.TaxPercent),
		"effective_date": s.EffectiveDate.Format("2006-01-02"),
		"tax_active":     s.TaxActive,
		"country_id":     s.CountryID,
	}
}

func (s *ServiceTaxByState) String() string {
	return fmt.Sprintf("<ServiceTax %s - %.4f%%>", s.StateCode, s.TaxPercent)
}
