// internal/models/general/country.go
package general

import "fmt"

type Country struct {
	IDCountry int    `gorm:"column:id_country;primaryKey;autoIncrement"            json:"id_country"`
	Code      string `gorm:"column:code;type:varchar(3);not null;uniqueIndex:uniq_country_code"     json:"code"`
	Country   string `gorm:"column:country;type:varchar(100);not null;uniqueIndex:uniq_country_name" json:"country"`
}

func (Country) TableName() string { return "country" }

func (c *Country) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_country": c.IDCountry,
		"code":       c.Code,
		"country":    c.Country,
	}
}

func (c *Country) String() string {
	return fmt.Sprintf("<Country %s | %s>", c.Code, c.Country)
}
