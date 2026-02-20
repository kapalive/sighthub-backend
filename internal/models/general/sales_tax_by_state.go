// internal/models/general/sales_tax_by_state.go
package general

import (
	"fmt"
	"strconv"
	"time"
)

type SalesTaxByState struct {
	IDSalesTax      int       `gorm:"column:id_sales_tax;primaryKey"                                    json:"id_sales_tax"`
	StateCode       string    `gorm:"column:state_code;type:varchar(2);not null"                         json:"state_code"`
	StateName       string    `gorm:"column:state_name;type:varchar(30);not null"                        json:"state_name"`
	SalesTaxPercent float64   `gorm:"column:sales_tax_percent;type:numeric(5,4);not null;check:sales_tax_percent >= 0" json:"sales_tax_percent"`
	EffectiveDate   time.Time `gorm:"column:effective_date;type:date;not null"                           json:"effective_date"`
	TaxActive       bool      `gorm:"column:tax_active;not null;default:true"                            json:"tax_active"`
	CountryID       int       `gorm:"column:country_id;not null"                                         json:"country_id"`

	Country *Country `gorm:"foreignKey:CountryID;references:IDCountry" json:"-"`
}

func (SalesTaxByState) TableName() string { return "sales_tax_by_state" }

func (s *SalesTaxByState) ToMap() map[string]interface{} {
	var eff *string
	if !s.EffectiveDate.IsZero() {
		v := s.EffectiveDate.Format("2006-01-02")
		eff = &v
	}

	return map[string]interface{}{
		"id_sales_tax":      s.IDSalesTax,
		"state_code":        s.StateCode,
		"state_name":        s.StateName,
		"sales_tax_percent": strconv.FormatFloat(s.SalesTaxPercent, 'f', 4, 64),
		"effective_date":    eff,
		"tax_active":        s.TaxActive,
		"country_id":        s.CountryID,
	}
}

func (s *SalesTaxByState) String() string {
	return fmt.Sprintf("<SalesTaxByState %s (%s): %s effective from %s>",
		s.StateCode,
		s.StateName,
		strconv.FormatFloat(s.SalesTaxPercent, 'f', 4, 64),
		s.EffectiveDate.Format("2006-01-02"),
	)
}
