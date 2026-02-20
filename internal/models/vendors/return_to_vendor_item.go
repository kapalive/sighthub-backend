// internal/models/vendors/return_to_vendor_item.go
package vendors

import (
	"fmt"
	"sighthub-backend/internal/models/interfaces"
)

// ReturnToVendorItem представляет модель для возврата товаров поставщикам
type ReturnToVendorItem struct {
	IDReturnToVendorItem    int64   `gorm:"column:id_return_to_vendor_item;primaryKey"                                   json:"id_return_to_vendor_item"`
	ReturnToVendorInvoiceID int64   `gorm:"column:return_to_vendor_invoice_id;not null;index"                            json:"return_to_vendor_invoice_id"`
	InventoryID             int64   `gorm:"column:inventory_id;not null;index"                                          json:"inventory_id"`
	ReasonReturn            string  `gorm:"column:reason_return;type:reason_return_vendor;not null"                      json:"reason_return"`
	PurchaseCost            float64 `gorm:"column:purchase_cost;type:numeric(10,2);not null;default:0.00"               json:"purchase_cost"`

	// Используем интерфейс вместо импорта модели
	ReturnToVendorInvoice ReturnToVendorInvoice         `gorm:"foreignKey:ReturnToVendorInvoiceID;references:IDReturnToVendorInvoice" json:"-"`
	Inventory             interfaces.InventoryInterface `gorm:"foreignKey:InventoryID;references:IDInventory"                         json:"-"`
}

func (ReturnToVendorItem) TableName() string { return "return_to_vendor_item" }

func (r *ReturnToVendorItem) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_return_to_vendor_item":    r.IDReturnToVendorItem,
		"return_to_vendor_invoice_id": r.ReturnToVendorInvoiceID,
		"inventory_id":                r.InventoryID,
		"purchase_cost":               r.PurchaseCost,
		"reason_return":               r.ReasonReturn,
	}
}

func (r *ReturnToVendorItem) String() string {
	return fmt.Sprintf("<ReturnToVendorItem id=%d reason=%s>", r.IDReturnToVendorItem, r.ReasonReturn)
}
