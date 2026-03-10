// internal/models/patients/payment_history.go
package patients

import "time"

// PaymentHistory ⇄ table: payment_history
type PaymentHistory struct {
	PaymentID            int64      `gorm:"column:payment_id;primaryKey;autoIncrement"                json:"payment_id"`
	PatientID            *int64     `gorm:"column:patient_id"                                         json:"patient_id,omitempty"`
	InvoiceID            int64      `gorm:"column:invoice_id;not null"                                json:"invoice_id"`
	Amount               float64    `gorm:"column:amount;type:numeric(10,2);not null"                  json:"amount"`
	PaymentTimestamp     time.Time  `gorm:"column:payment_timestamp;type:timestamptz;not null;default:now()" json:"payment_timestamp"`
	TransactionHash      *string    `gorm:"column:transaction_hash;type:text"                          json:"transaction_hash,omitempty"`
	PaymentMethodID      *int64     `gorm:"column:payment_method_id"                                  json:"payment_method_id,omitempty"`
	PaymentTransactionID *int64     `gorm:"column:payment_transaction_id"                             json:"payment_transaction_id,omitempty"`
	EmployeeID           *int64     `gorm:"column:employee_id"                                        json:"employee_id,omitempty"`
	Note                 *string    `gorm:"column:note;type:text"                                     json:"note,omitempty"`
}

func (PaymentHistory) TableName() string { return "payment_history" }

func (p *PaymentHistory) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"payment_id":             p.PaymentID,
		"patient_id":             p.PatientID,
		"invoice_id":             p.InvoiceID,
		"amount":                 p.Amount,
		"payment_timestamp":      p.PaymentTimestamp.Format(time.RFC3339),
		"transaction_hash":       p.TransactionHash,
		"payment_method_id":      p.PaymentMethodID,
		"payment_transaction_id": p.PaymentTransactionID,
		"employee_id":            p.EmployeeID,
		"note":                   p.Note,
	}
}
