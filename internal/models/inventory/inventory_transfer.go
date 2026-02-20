// internal/models/inventory/inventory_transfer.go
package inventory

import (
	"fmt"
	"sighthub-backend/internal/models/interfaces" // Импортируем интерфейсы
	"time"
)

// Модель для трансфера инвентаря (InventoryTransfer)
type InventoryTransfer struct {
	IDTransfer     int64     `gorm:"column:id_transfer;primaryKey;autoIncrement" json:"id_transfer"`
	InventoryID    int64     `gorm:"column:inventory_id;not null" json:"inventory_id"`
	FromLocationID int64     `gorm:"column:from_location_id;not null" json:"from_location_id"`
	ToLocationID   int64     `gorm:"column:to_location_id;not null" json:"to_location_id"`
	DateTransfer   time.Time `gorm:"column:date_transfer;default:CURRENT_TIMESTAMP" json:"date_transfer"`
	TransferredBy  int64     `gorm:"column:transferred_by" json:"transferred_by"`
	ReceivedBy     int64     `gorm:"column:received_by" json:"received_by"`
	StatusItems    string    `gorm:"column:status_items;not null" json:"status_items"`
	InvoiceID      int64     `gorm:"column:invoice_id;not null" json:"invoice_id"`
	OldInvoiceID   *int64    `gorm:"column:old_invoice_id" json:"old_invoice_id,omitempty"`
	InvoiceFrom    *int64    `gorm:"column:invoice_from" json:"invoice_from,omitempty"`
	InvoiceTo      *int64    `gorm:"column:invoice_to" json:"invoice_to,omitempty"`
	SystemNote     *string   `gorm:"column:system_note" json:"system_note,omitempty"`
}

// TableName задаёт имя таблицы в БД
func (InventoryTransfer) TableName() string {
	return "inventory_transfer"
}

// ToMap превращает объект в карту для удобства работы с данными
func (i *InventoryTransfer) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_transfer":      i.IDTransfer,
		"inventory_id":     i.InventoryID,
		"from_location_id": i.FromLocationID,
		"to_location_id":   i.ToLocationID,
		"date_transfer":    i.DateTransfer,
		"transferred_by":   i.TransferredBy,
		"received_by":      i.ReceivedBy,
		"status_items":     i.StatusItems,
		"invoice_id":       i.InvoiceID,
		"old_invoice_id":   i.OldInvoiceID,
		"invoice_from":     i.InvoiceFrom,
		"invoice_to":       i.InvoiceTo,
		"system_note":      i.SystemNote,
	}
}

// String метод для печати объекта
func (i *InventoryTransfer) String() string {
	return fmt.Sprintf("<InventoryTransfer %d | InventoryID: %d | FromLocationID: %d | ToLocationID: %d | InvoiceID: %d | Status: %s>",
		i.IDTransfer, i.InventoryID, i.FromLocationID, i.ToLocationID, i.InvoiceID, i.StatusItems)
}

// Получить данные о местоположении "From" через интерфейс
func (i *InventoryTransfer) GetFromLocation(locationVendor interfaces.LocationInterface) (map[string]interface{}, error) {
	return locationVendor.GetLocationByID(i.FromLocationID)
}

// Получить данные о местоположении "To" через интерфейс
func (i *InventoryTransfer) GetToLocation(locationVendor interfaces.LocationInterface) (map[string]interface{}, error) {
	return locationVendor.GetLocationByID(i.ToLocationID)
}

// Получить данные о сотруднике, который выполнил трансфер
func (i *InventoryTransfer) GetTransferredBy(employeeVendor interfaces.EmployeeInterface) (map[string]interface{}, error) {
	return employeeVendor.GetEmployeeByID(i.TransferredBy)
}

// Получить данные о сотруднике, который принял трансфер
func (i *InventoryTransfer) GetReceivedBy(employeeVendor interfaces.EmployeeInterface) (map[string]interface{}, error) {
	return employeeVendor.GetEmployeeByID(i.ReceivedBy)
}

// Получить инвойс по ID через интерфейс Invoice
func (i *InventoryTransfer) GetInvoice(invoiceVendor interfaces.InvoiceInterface) (map[string]interface{}, error) {
	return invoiceVendor.GetInvoiceByID(i.InvoiceID)
}

// Получить старый инвойс через интерфейс Invoice
func (i *InventoryTransfer) GetOldInvoice(invoiceVendor interfaces.InvoiceInterface) (map[string]interface{}, error) {
	if i.OldInvoiceID == nil {
		return nil, nil
	}
	return invoiceVendor.GetInvoiceByID(*i.OldInvoiceID)
}

// Получить инвойс "From" через интерфейс Invoice
func (i *InventoryTransfer) GetInvoiceFrom(invoiceVendor interfaces.InvoiceInterface) (map[string]interface{}, error) {
	if i.InvoiceFrom == nil {
		return nil, nil
	}
	return invoiceVendor.GetInvoiceByID(*i.InvoiceFrom)
}

// Получить инвойс "To" через интерфейс Invoice
func (i *InventoryTransfer) GetInvoiceTo(invoiceVendor interfaces.InvoiceInterface) (map[string]interface{}, error) {
	if i.InvoiceTo == nil {
		return nil, nil
	}
	return invoiceVendor.GetInvoiceByID(*i.InvoiceTo)
}
