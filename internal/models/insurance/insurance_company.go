package insurance

import "fmt"

// InsuranceCompany ⇄ insurance_company
type InsuranceCompany struct {
	IDInsuranceCompany int     `gorm:"column:id_insurance_company;primaryKey"         json:"id_insurance_company"`
	CompanyName        string  `gorm:"column:company_name;type:varchar(100);not null" json:"company_name"`
	ContactNumber      *string `gorm:"column:contact_number;type:varchar(15)"         json:"contact_number,omitempty"`
	ContactEmail       *string `gorm:"column:contact_email;type:varchar(100)"         json:"contact_email,omitempty"`
	Address            *string `gorm:"column:address;type:varchar(100)"               json:"address,omitempty"`
	AddressLine2       *string `gorm:"column:address_line_2;type:varchar(100)"        json:"address_line_2,omitempty"`
	City               *string `gorm:"column:city;type:varchar(50)"                   json:"city,omitempty"`
	State              *string `gorm:"column:state;type:varchar(2)"                   json:"state,omitempty"`
	ZipCode            *string `gorm:"column:zip_code;type:varchar(10)"               json:"zip_code,omitempty"`
}

func (InsuranceCompany) TableName() string { return "insurance_company" }

// ToMap — аналог Python to_dict()
func (c *InsuranceCompany) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_insurance_company": c.IDInsuranceCompany,
		"company_name":         c.CompanyName,
		"contact_number":       c.ContactNumber,
		"contact_email":        c.ContactEmail,
		"address":              c.Address,
		"address_line_2":       c.AddressLine2,
		"city":                 c.City,
		"state":                c.State,
		"zip_code":             c.ZipCode,
	}
}

func (c *InsuranceCompany) String() string {
	return fmt.Sprintf("<InsuranceCompany %s>", c.CompanyName)
}
