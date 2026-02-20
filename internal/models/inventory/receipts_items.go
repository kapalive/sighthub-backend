// internal/models/inventory/receipts_items.go
package inventory

import (
	"fmt"
	"sighthub-backend/internal/models/interfaces" // Импортируем интерфейсы
	"time"
)

// Модель для товаров в получении инвентаря (ReceiptsItems)
type ReceiptsItems struct {
	IDReceiptsItems int64     `gorm:"column:id_receipts_items;primaryKey;autoIncrement" json:"id_receipts_items"`
	InvoiceID       int64     `gorm:"column:invoice_id;not null" json:"invoice_id"`
	InventoryID     int64     `gorm:"column:inventory_id;not null" json:"inventory_id"`
	DateTime        time.Time `gorm:"column:datetime;default:CURRENT_TIMESTAMP" json:"datetime"`
}

// TableName задаёт имя таблицы в БД
func (ReceiptsItems) TableName() string {
	return "receipts_items"
}

// ToMap превращает объект в карту для удобства работы с данными
func (r *ReceiptsItems) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_receipts_items": r.IDReceiptsItems,
		"invoice_id":        r.InvoiceID,
		"inventory_id":      r.InventoryID,
		"datetime":          r.DateTime,
	}
}

// String метод для печати объекта
func (r *ReceiptsItems) String() string {
	return fmt.Sprintf("<ReceiptsItems invoice_id=%d, inventory_id=%d, datetime=%s>", r.InvoiceID, r.InventoryID, r.DateTime)
}

// Получить инвойс через интерфейс Invoice
func (r *ReceiptsItems) GetInvoice(invoiceVendor interfaces.InvoiceInterface) (map[string]interface{}, error) {
	return invoiceVendor.GetInvoiceByID(r.InvoiceID)
}

// Получить инвентарь через интерфейс Inventory
func (r *ReceiptsItems) GetInventory(inventoryVendor interfaces.InventoryInterface) (map[string]interface{}, error) {
	return inventoryVendor.GetInventoryByID(r.InventoryID)
}
