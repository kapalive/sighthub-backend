// internal/models/invoices/return_items.go
package invoices

import (
	"fmt"
)

// ReturnItem ↔ table: return_items
type ReturnItem struct {
	IDReturnItem int64    `gorm:"column:id_return_item;primaryKey;autoIncrement"   json:"id_return_item"`
	ReturnID     int64    `gorm:"column:return_id;not null"                        json:"return_id"`
	ItemSaleID   int64    `gorm:"column:item_sale_id;not null"                     json:"item_sale_id"`
	ReturnAmount *float64 `gorm:"column:return_amount;type:numeric(10,2)"          json:"return_amount,omitempty"`

	// Relations (оба в этом же пакете)
	ReturnInvoice   *ReturnInvoice   `gorm:"foreignKey:ReturnID;references:ReturnID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"return_invoice,omitempty"`
	InvoiceItemSale *InvoiceItemSale `gorm:"foreignKey:ItemSaleID;references:IDInvoiceSale"                                      json:"invoice_item_sale,omitempty"`

	// Доп. поле для имени товара (резолвится на уровне сервиса, чтобы не плодить циклы)
	ItemName *string `gorm:"-" json:"-"`
}

func (ReturnItem) TableName() string { return "return_items" }

func (r *ReturnItem) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_return_item": r.IDReturnItem,
		"return_id":      r.ReturnID,
		"item_sale_id":   r.ItemSaleID,
		"return_amount":  r.ReturnAmount,
	}

	// вложенные данные о позиции счета (как в Python to_dict)
	if r.InvoiceItemSale != nil {
		inv := map[string]interface{}{
			"id_invoice_sale": r.InvoiceItemSale.IDInvoiceSale,
			"item_id":         r.InvoiceItemSale.ItemID,
			"item_type":       r.InvoiceItemSale.ItemType,
			"price":           r.InvoiceItemSale.Price,
		}
		// опционально добавим item_name, если сервис его проставил
		if r.ItemName != nil {
			inv["item_name"] = *r.ItemName
		}
		m["invoice_item"] = inv
	}

	return m
}

func (r *ReturnItem) String() string {
	return fmt.Sprintf("<ReturnItem %d - Return %d - Item %d>", r.IDReturnItem, r.ReturnID, r.ItemSaleID)
}

// Опционально: удобный конструктор
func NewReturnItem(returnID, itemSaleID int64) *ReturnItem {
	return &ReturnItem{
		ReturnID:   returnID,
		ItemSaleID: itemSaleID,
	}
}
