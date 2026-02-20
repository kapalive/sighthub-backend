// internal/models/inventory/inventory.go
package inventory

import (
	"fmt"
	"time"

	"sighthub-backend/internal/models/types"
)

type Inventory struct {
	IDInventory          int64                      `gorm:"column:id_inventory;primaryKey"               json:"id_inventory"`
	SKU                  string                     `gorm:"column:sku;type:varchar(12);not null"         json:"sku"`
	CreatedDate          time.Time                  `gorm:"column:created_date;default:CURRENT_TIMESTAMP" json:"created_date"`
	StatusItemsInventory types.StatusItemsInventory `gorm:"column:status_items_inventory;not null"       json:"status_items_inventory"`
	LocationID           int64                      `gorm:"column:location_id;not null"                  json:"location_id"`
	ModelID              *int64                     `gorm:"column:model_id"                              json:"model_id,omitempty"`
	InvoiceID            int64                      `gorm:"column:invoice_id;not null"                   json:"invoice_id"`
	OrdersLensID         *int64                     `gorm:"column:orders_lens_id"                        json:"orders_lens_id,omitempty"`
	EmployeeID           *int64                     `gorm:"column:employee_id"                           json:"employee_id,omitempty"`
	VariantCRSLProductID *int                       `gorm:"column:variant_cr_sl_product_id"              json:"variant_cr_sl_product_id,omitempty"`

	// ВАЖНО: никаких кросс-пакетных relation-полей здесь.
	// Любые preload/join — через репозиторий (см. ниже).
}

func (Inventory) TableName() string { return "inventory" }

// Реализуем интерфейс InventoryInterface
func (i *Inventory) ID() int64 {
	return i.IDInventory
}

func (i *Inventory) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_inventory":             i.IDInventory,
		"sku":                      i.SKU,
		"created_date":             i.CreatedDate,
		"status_items_inventory":   i.StatusItemsInventory,
		"location_id":              i.LocationID,
		"model_id":                 i.ModelID,
		"invoice_id":               i.InvoiceID,
		"orders_lens_id":           i.OrdersLensID,
		"employee_id":              i.EmployeeID,
		"variant_cr_sl_product_id": i.VariantCRSLProductID,
	}
}

func (i *Inventory) String() string {
	return fmt.Sprintf("<Inventory %s>", i.SKU)
}

func (i *Inventory) GenerateSKU() {
	var mid int64
	if i.ModelID != nil {
		mid = *i.ModelID
	}
	i.SKU = fmt.Sprintf("%03d/%03d", mid%1000, time.Now().Unix()%1000)
}
