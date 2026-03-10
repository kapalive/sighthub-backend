// internal/models/insurance/insurance_payment.go
package insurance

import "time"

// InsurancePayment ⇄ insurance_payment
type InsurancePayment struct {
	IDInsurancePayment   int64      `gorm:"column:id_insurance_payment;primaryKey;autoIncrement"  json:"id_insurance_payment"`
	InvoiceID            int64      `gorm:"column:invoice_id;not null;index"                      json:"invoice_id"`
	InsurancePolicyID    int64      `gorm:"column:insurance_policy_id;not null"                   json:"insurance_policy_id"`
	PaymentTypeID        int        `gorm:"column:payment_type_id;not null"                       json:"payment_type_id"`
	Amount               string     `gorm:"column:amount;type:numeric(10,2);not null"             json:"amount"`
	ReferenceNumber      *string    `gorm:"column:reference_number;type:varchar(100)"             json:"reference_number,omitempty"`
	Note                 *string    `gorm:"column:note;type:text"                                 json:"note,omitempty"`
	EmployeeID           *int64     `gorm:"column:employee_id"                                    json:"employee_id,omitempty"`
	CreatedAt            *time.Time `gorm:"column:created_at;type:timestamptz;default:now()"      json:"created_at,omitempty"`
}

func (InsurancePayment) TableName() string { return "insurance_payment" }

func (i *InsurancePayment) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_insurance_payment": i.IDInsurancePayment,
		"invoice_id":           i.InvoiceID,
		"insurance_policy_id":  i.InsurancePolicyID,
		"payment_type_id":      i.PaymentTypeID,
		"amount":               i.Amount,
		"reference_number":     i.ReferenceNumber,
		"note":                 i.Note,
		"employee_id":          i.EmployeeID,
	}
	if i.CreatedAt != nil {
		m["created_at"] = i.CreatedAt.Format(time.RFC3339)
	} else {
		m["created_at"] = nil
	}
	return m
}
