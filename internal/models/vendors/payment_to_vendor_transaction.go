// internal/models/vendors/payment_to_vendor_transaction.go
package vendors

import "time"

// PaymentToVendorTransaction ⇄ payment_to_vendor_transaction
type PaymentToVendorTransaction struct {
	IDPaymentVendorTransaction int64      `gorm:"column:id_payment_vendor_transaction;primaryKey;autoIncrement" json:"id_payment_vendor_transaction"`
	VendorID                   int        `gorm:"column:vendor_id;not null"                                     json:"vendor_id"`
	LocationID                 int64      `gorm:"column:location_id;not null"                                   json:"location_id"`
	EmployeeID                 int64      `gorm:"column:employee_id;not null"                                   json:"employee_id"`
	PaymentMethodID            int        `gorm:"column:payment_method_id;not null"                             json:"payment_method_id"`
	PaymentDate                time.Time  `gorm:"column:payment_date;type:date;not null"                        json:"-"`
	Amount                     string     `gorm:"column:amount;type:numeric(12,2);not null"                     json:"amount"`
	Note                       *string    `gorm:"column:note;type:varchar(255)"                                 json:"note,omitempty"`
	VendorLocationAccountID    *int64     `gorm:"column:vendor_location_account_id"                             json:"vendor_location_account_id,omitempty"`
	CreatedAt                  *time.Time `gorm:"column:created_at;type:timestamptz;default:now()"              json:"created_at,omitempty"`
}

func (PaymentToVendorTransaction) TableName() string { return "payment_to_vendor_transaction" }

func (p *PaymentToVendorTransaction) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_payment_vendor_transaction": p.IDPaymentVendorTransaction,
		"vendor_id":                     p.VendorID,
		"location_id":                   p.LocationID,
		"employee_id":                   p.EmployeeID,
		"payment_method_id":             p.PaymentMethodID,
		"payment_date":                  p.PaymentDate.Format("2006-01-02"),
		"amount":                        p.Amount,
		"note":                          p.Note,
		"vendor_location_account_id":    p.VendorLocationAccountID,
	}
}
