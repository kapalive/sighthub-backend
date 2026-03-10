// internal/models/insurance/insurance_payment_type.go
package insurance

import "time"

// InsurancePaymentType ⇄ insurance_payment_type
type InsurancePaymentType struct {
	IDInsurancePaymentType int        `gorm:"column:id_insurance_payment_type;primaryKey;autoIncrement" json:"id_insurance_payment_type"`
	Name                   string     `gorm:"column:name;type:varchar(100);not null;uniqueIndex"         json:"name"`
	Description            *string    `gorm:"column:description;type:varchar(255)"                       json:"description,omitempty"`
	Active                 bool       `gorm:"column:active;not null;default:true"                        json:"active"`
	CreatedAt              *time.Time `gorm:"column:created_at;type:timestamptz;default:now()"           json:"created_at,omitempty"`
}

func (InsurancePaymentType) TableName() string { return "insurance_payment_type" }

func (i *InsurancePaymentType) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_insurance_payment_type": i.IDInsurancePaymentType,
		"name":                      i.Name,
		"description":               i.Description,
		"active":                    i.Active,
	}
}
