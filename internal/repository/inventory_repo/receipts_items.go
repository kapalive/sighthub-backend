// internal/repository/inventory_repo/receipts_items.go
package inventory_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/inventory"
	"sighthub-backend/internal/models/types"
)

type ReceiptsItemsRepo struct{ DB *gorm.DB }

func NewReceiptsItemsRepo(db *gorm.DB) *ReceiptsItemsRepo {
	return &ReceiptsItemsRepo{DB: db}
}

// GetByInvoiceID возвращает позиции получения для инвойса.
func (r *ReceiptsItemsRepo) GetByInvoiceID(invoiceID int64) ([]inventory.ReceiptsItems, error) {
	var rows []inventory.ReceiptsItems
	return rows, r.DB.Where("invoice_id = ?", invoiceID).Find(&rows).Error
}

// ReceiptItemDetail — расширенная строка с данными об инвентаре.
type ReceiptItemDetail struct {
	IDReceiptsItems int64     `json:"id_receipts_items"`
	InvoiceID       int64     `json:"invoice_id"`
	InventoryID     int64     `json:"inventory_id"`
	DateTime        time.Time `json:"datetime"`
	SKU             string    `json:"sku"`
	StatusItems     string    `json:"status_items_inventory"`
	LocationID      int64     `json:"location_id"`
}

// GetDetailByInvoice возвращает позиции получения с данными инвентаря.
func (r *ReceiptsItemsRepo) GetDetailByInvoice(invoiceID int64) ([]ReceiptItemDetail, error) {
	var rows []ReceiptItemDetail
	err := r.DB.
		Table("receipts_items ri").
		Select("ri.id_receipts_items, ri.invoice_id, ri.inventory_id, ri.datetime, i.sku, i.status_items_inventory, i.location_id").
		Joins("JOIN inventory i ON i.id_inventory = ri.inventory_id").
		Where("ri.invoice_id = ?", invoiceID).
		Scan(&rows).Error
	return rows, err
}

// Confirm добавляет единицу инвентаря в список полученных и
// обновляет её статус на "available" в одной транзакции.
func (r *ReceiptsItemsRepo) Confirm(invoiceID, inventoryID int64) (*inventory.ReceiptsItems, error) {
	ri := &inventory.ReceiptsItems{
		InvoiceID:   invoiceID,
		InventoryID: inventoryID,
		DateTime:    time.Now(),
	}
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ri).Error; err != nil {
			return err
		}
		return tx.Model(&inventory.Inventory{}).
			Where("id_inventory = ?", inventoryID).
			Update("status_items_inventory", types.StatusInventoryReadyForSale).Error
	})
	return ri, err
}

// Delete удаляет запись о получении.
func (r *ReceiptsItemsRepo) Delete(id int64) error {
	return r.DB.Delete(&inventory.ReceiptsItems{}, id).Error
}

func (r *ReceiptsItemsRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
