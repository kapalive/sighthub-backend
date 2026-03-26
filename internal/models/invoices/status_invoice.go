// internal/models/invoices/status_invoice.go
package invoices

import "fmt"

type StatusInvoice struct {
	IDStatusInvoice    int     `gorm:"column:id_status_invoice;primaryKey" json:"id_status_invoice"`
	StatusInvoiceValue string  `gorm:"column:status_invoice_value;type:varchar(255);not null" json:"name"`
	Icon               *string `gorm:"column:icon;type:varchar(255)" json:"icon,omitempty"`
	StatusType         string  `gorm:"column:status_type;type:varchar(20);not null;default:patient" json:"status_type"`
}

func (StatusInvoice) TableName() string { return "status_invoice" }

func (s *StatusInvoice) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_status_invoice": s.IDStatusInvoice,
		"name":              s.StatusInvoiceValue,
		"icon":              s.Icon,
		"status_type":       s.StatusType,
	}
}

func (s *StatusInvoice) String() string {
	return fmt.Sprintf("<StatusInvoice %s>", s.StatusInvoiceValue)
}
