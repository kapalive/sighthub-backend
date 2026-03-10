// internal/models/invoices/invoice_insurance_policy.go
package invoices

// InvoiceInsurancePolicy ⇄ table: invoice_insurance_policy
type InvoiceInsurancePolicy struct {
	IDInvoiceInsurancePolicy int   `gorm:"column:id_invoice_insurance_policy;primaryKey;autoIncrement" json:"id_invoice_insurance_policy"`
	InvoiceID                int64 `gorm:"column:invoice_id;not null"                                   json:"invoice_id"`
	InsurancePolicyID        int64 `gorm:"column:insurance_policy_id;not null"                          json:"insurance_policy_id"`
}

func (InvoiceInsurancePolicy) TableName() string { return "invoice_insurance_policy" }

func (i *InvoiceInsurancePolicy) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_invoice_insurance_policy": i.IDInvoiceInsurancePolicy,
		"invoice_id":                  i.InvoiceID,
		"insurance_policy_id":         i.InsurancePolicyID,
	}
}
