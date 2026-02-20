// internal/models/invoices/return_invoices.go
package invoices

import (
	"fmt"
	"time"
)

// ReturnInvoice соответствует таблице return_invoices
type ReturnInvoice struct {
	ReturnID     int64     `gorm:"column:return_id;primaryKey;autoIncrement"                 json:"return_id"`
	InvoiceID    int64     `gorm:"column:invoice_id;not null"                                 json:"invoice_id"`
	ReturnReason *string   `gorm:"column:return_reason;type:text"                             json:"return_reason,omitempty"`
	ReturnedQty  int       `gorm:"column:returned_quantity;not null;default:1"                json:"returned_quantity"`
	ReturnAmount float64   `gorm:"column:return_amount;type:numeric(10,2);not null;default:0" json:"return_amount"`
	ReturnDate   time.Time `gorm:"column:return_date;type:timestamptz;default:CURRENT_TIMESTAMP" json:"return_date"`
	Status       string    `gorm:"column:status;type:varchar(50);default:Initialized"         json:"status"`

	// Связь с Invoice (в том же пакете)
	// Предполагается, что у Invoice поле PK называется IDInvoice и column:id_invoice
	Invoice *Invoice `gorm:"foreignKey:InvoiceID;references:IDInvoice;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"invoice,omitempty"`
}

func (ReturnInvoice) TableName() string { return "return_invoices" }

func (r *ReturnInvoice) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"return_id":         r.ReturnID,
		"invoice_id":        r.InvoiceID,
		"return_reason":     r.ReturnReason,
		"returned_quantity": r.ReturnedQty,
		"return_amount":     r.ReturnAmount,
		"return_date":       r.ReturnDate, // time.Time — отдадите как ISO-8601 через json.Marshal
		"status":            r.Status,
	}
}

func (r *ReturnInvoice) String() string {
	return fmt.Sprintf("<ReturnInvoice %d - Invoice %d>", r.ReturnID, r.InvoiceID)
}

// Опционально: удобный конструктор с дефолтами
func NewReturnInvoice(invoiceID int64, amount float64, qty int, reason *string) *ReturnInvoice {
	if qty <= 0 {
		qty = 1
	}
	return &ReturnInvoice{
		InvoiceID:    invoiceID,
		ReturnReason: reason,
		ReturnedQty:  qty,
		ReturnAmount: amount,
		Status:       "Initialized",
	}
}
