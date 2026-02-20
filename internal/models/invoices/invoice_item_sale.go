// internal/models/invoices/invoice_item_sale.go
package invoices

import (
	"fmt"
)

// InvoiceItemSale соответствует таблице invoice_item_sale
type InvoiceItemSale struct {
	IDInvoiceSale int64    `gorm:"column:id_invoice_sale;primaryKey;autoIncrement"            json:"id_invoice_sale"`
	InvoiceID     int64    `gorm:"column:invoice_id;not null"                                  json:"invoice_id"`
	ItemType      string   `gorm:"column:item_type;type:varchar(50);not null"                  json:"item_type"`
	ItemID        *int64   `gorm:"column:item_id"                                              json:"item_id"` // nullable
	Description   string   `gorm:"column:description;type:varchar(255);not null"               json:"description"`
	Quantity      int      `gorm:"column:quantity;not null;default:1"                          json:"quantity"`
	Price         float64  `gorm:"column:price;type:numeric(10,2);not null;default:0"          json:"price"`
	Cost          float64  `gorm:"column:cost;type:numeric(10,2);not null;default:0"           json:"cost"`
	Discount      float64  `gorm:"column:discount;type:numeric(10,2);not null;default:0"       json:"discount"`
	Total         float64  `gorm:"column:total;type:numeric(10,2);not null;default:0"          json:"total"`
	Taxable       *bool    `gorm:"column:taxable"                                              json:"taxable"` // nullable
	TotalTax      float64  `gorm:"column:total_tax;type:numeric(10,4);not null;default:0"      json:"total_tax"`
	InsBalance    *float64 `gorm:"column:ins_balance;type:numeric(10,2)"                       json:"ins_balance"` // nullable

	// Связь с Invoice (тот же пакет)
	Invoice *Invoice `gorm:"foreignKey:InvoiceID;references:IDInvoice" json:"invoice,omitempty"`
}

func (InvoiceItemSale) TableName() string { return "invoice_item_sale" }

func (i *InvoiceItemSale) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_invoice_sale": i.IDInvoiceSale,
		"invoice_id":      i.InvoiceID,
		"item_type":       i.ItemType,
		"item_id":         i.ItemID,
		"description":     i.Description,
		"quantity":        i.Quantity,
		"price":           i.Price,
		"cost":            i.Cost,
		"discount":        i.Discount,
		"total":           i.Total,
		"taxable":         i.Taxable,
		"total_tax":       i.TotalTax,
		"ins_balance":     i.InsBalance,
	}
}

func (i *InvoiceItemSale) String() string {
	return fmt.Sprintf("<InvoiceItemSale %d - %s>", i.IDInvoiceSale, i.Description)
}
