// internal/models/inventory/inventory_transaction.go
package inventory

import (
	"fmt"
	"sighthub-backend/internal/models/interfaces" // Импортируем интерфейс
	"time"
)

// Модель для транзакции инвентаря (InventoryTransaction)
type InventoryTransaction struct {
	IDTransaction    int64     `gorm:"column:id_transaction;primaryKey;autoIncrement" json:"id_transaction"`
	InventoryID      int64     `gorm:"column:inventory_id;not null" json:"inventory_id"`
	FromLocationID   int64     `gorm:"column:from_location_id" json:"from_location_id"`
	ToLocationID     int64     `gorm:"column:to_location_id" json:"to_location_id"`
	TransferredBy    int64     `gorm:"column:transferred_by;not null" json:"transferred_by"`
	InvoiceID        int64     `gorm:"column:invoice_id;not null" json:"invoice_id"`
	OldInvoiceID     *int64    `gorm:"column:old_invoice_id" json:"old_invoice_id,omitempty"`
	StatusItems      string    `gorm:"column:status_items;not null" json:"status_items"`
	TransactionType  string    `gorm:"column:transaction_type;default:'Transfer';not null" json:"transaction_type"`
	DateTransaction  time.Time `gorm:"column:date_transaction;default:CURRENT_TIMESTAMP" json:"date_transaction"`
	InventoryCountID *int64    `gorm:"column:inventory_count_id" json:"inventory_count_id,omitempty"`
	Notes            *string   `gorm:"column:notes" json:"notes,omitempty"`
}

// TableName задаёт имя таблицы в БД
func (InventoryTransaction) TableName() string {
	return "inventory_transaction"
}

// ToMap превращает объект в карту для удобства работы с данными
func (i *InventoryTransaction) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_transaction":     i.IDTransaction,
		"inventory_id":       i.InventoryID,
		"from_location_id":   i.FromLocationID,
		"to_location_id":     i.ToLocationID,
		"transferred_by":     i.TransferredBy,
		"invoice_id":         i.InvoiceID,
		"old_invoice_id":     i.OldInvoiceID,
		"status_items":       i.StatusItems,
		"transaction_type":   i.TransactionType,
		"date_transaction":   i.DateTransaction,
		"inventory_count_id": i.InventoryCountID,
		"notes":              i.Notes,
	}
}

// String метод для печати объекта
func (i *InventoryTransaction) String() string {
	return fmt.Sprintf("<InventoryTransaction %d | InventoryID: %d | FromLocationID: %d | ToLocationID: %d | InvoiceID: %d | Status: %s | InventoryCountID: %d>",
		i.IDTransaction, i.InventoryID, i.FromLocationID, i.ToLocationID, i.InvoiceID, i.StatusItems, i.InventoryCountID)
}

// Получить данные о счете через интерфейс Invoice
func (i *InventoryTransaction) GetInvoice(invoiceVendor interfaces.InvoiceInterface) (map[string]interface{}, error) {
	// Используем интерфейс для получения инвойса
	return invoiceVendor.ToMap(), nil
}

// Получить данные о старом счете через интерфейс Invoice
func (i *InventoryTransaction) GetOldInvoice(invoiceVendor interfaces.InvoiceInterface) (map[string]interface{}, error) {
	if i.OldInvoiceID == nil {
		return nil, nil
	}
	// Используем интерфейс для получения старого инвойса
	return invoiceVendor.ToMap(), nil
}
