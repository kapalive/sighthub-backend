// internal/models/invoices/invoice_services_item.go
package invoices

import (
	"fmt"
)

// InvoiceServicesItem ↔ table: invoice_services_item
type InvoiceServicesItem struct {
	IDInvoiceServicesItem int64    `gorm:"column:id_invoice_services_item;primaryKey;autoIncrement"       json:"id_invoice_services_item"`
	InvoiceID             int64    `gorm:"column:invoice_id;not null"                                     json:"invoice_id"`
	ProfessionalServiceID *int64   `gorm:"column:professional_service_id"                                 json:"professional_service_id,omitempty"`
	LensOrderID           *int64   `gorm:"column:lens_order_id"                                           json:"lens_order_id,omitempty"`
	ContactLensItemID     *int64   `gorm:"column:contact_lens_item_id"                                    json:"contact_lens_item_id,omitempty"`
	AdditionalServiceID   *int64   `gorm:"column:additional_service_id"                                   json:"additional_service_id,omitempty"`
	Quantity              int      `gorm:"column:quantity;not null;default:1"                             json:"quantity"`
	Price                 float64  `gorm:"column:price;type:numeric(10,2);not null;default:0"             json:"price"`
	Total                 *float64 `gorm:"column:total;type:numeric(10,2)"                                json:"total,omitempty"`

	// В Python-модели были relation'ы; здесь оставляем только FK.
	// Связь с Invoice (тот же пакет), без циклов с другими пакетами.
	Invoice *Invoice `gorm:"foreignKey:InvoiceID;references:IDInvoice" json:"invoice,omitempty"`
}

func (InvoiceServicesItem) TableName() string { return "invoice_services_item" }

func (i *InvoiceServicesItem) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_invoice_services_item": i.IDInvoiceServicesItem,
		"invoice_id":               i.InvoiceID,
		"professional_service_id":  i.ProfessionalServiceID,
		"lens_order_id":            i.LensOrderID,
		"contact_lens_item_id":     i.ContactLensItemID,
		"additional_service_id":    i.AdditionalServiceID,
		"quantity":                 i.Quantity,
		"price":                    i.Price,
		"total":                    i.Total,
	}
}

func (i *InvoiceServicesItem) String() string {
	return fmt.Sprintf("<InvoiceServicesItem %d | invoice=%d | qty=%d | price=%.2f>",
		i.IDInvoiceServicesItem, i.InvoiceID, i.Quantity, i.Price)
}

// Опционально: конструктор с дефолтами
func NewInvoiceServicesItem(invoiceID int64) *InvoiceServicesItem {
	return &InvoiceServicesItem{
		InvoiceID: invoiceID,
		Quantity:  1,
		Price:     0,
	}
}
