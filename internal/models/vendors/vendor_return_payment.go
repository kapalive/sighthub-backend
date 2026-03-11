package vendors

import (
	"fmt"
	"time"

	"sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/general"
)

// VendorReturnPayment ⇄ table: vendor_return_payment
type VendorReturnPayment struct {
	IDVendorReturnPayment   int64     `gorm:"column:id_vendor_return_payment;primaryKey;autoIncrement" json:"id_vendor_return_payment"`
	ReturnToVendorInvoiceID int64     `gorm:"column:return_to_vendor_invoice_id;not null;index"        json:"return_to_vendor_invoice_id"`
	Amount                  float64   `gorm:"column:amount;type:numeric(10,2);not null"                json:"amount"`
	PaymentMethodID         int       `gorm:"column:payment_method_id;not null"                        json:"payment_method_id"`
	EmployeeID              *int64    `gorm:"column:employee_id"                                       json:"employee_id,omitempty"`
	PaymentTimestamp        time.Time `gorm:"column:payment_timestamp;default:CURRENT_TIMESTAMP"       json:"payment_timestamp"`
	Notes                   *string   `gorm:"column:notes;type:varchar(255)"                           json:"notes,omitempty"`

	PaymentMethod *general.PaymentMethod `gorm:"foreignKey:PaymentMethodID;references:IDPaymentMethod" json:"payment_method,omitempty"`
	Employee      *employees.Employee    `gorm:"foreignKey:EmployeeID;references:IDEmployee"           json:"employee,omitempty"`
}

func (VendorReturnPayment) TableName() string { return "vendor_return_payment" }

func (v *VendorReturnPayment) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_vendor_return_payment":    v.IDVendorReturnPayment,
		"return_to_vendor_invoice_id": v.ReturnToVendorInvoiceID,
		"amount":                      v.Amount,
		"payment_method_id":           v.PaymentMethodID,
		"employee_id":                 v.EmployeeID,
		"payment_timestamp":           v.PaymentTimestamp.Format(time.RFC3339),
		"notes":                       v.Notes,
	}
	if v.PaymentMethod != nil {
		m["payment_method"] = v.PaymentMethod.MethodName
	}
	return m
}

func (v *VendorReturnPayment) String() string {
	return fmt.Sprintf("<VendorReturnPayment id=%d amount=%.2f>", v.IDVendorReturnPayment, v.Amount)
}
