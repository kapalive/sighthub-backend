// internal/repository/inventory_repo/inventory_transfer.go
package inventory_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/inventory"
	"sighthub-backend/internal/models/types"
)

type InventoryTransferRepo struct{ DB *gorm.DB }

func NewInventoryTransferRepo(db *gorm.DB) *InventoryTransferRepo {
	return &InventoryTransferRepo{DB: db}
}

// GetByInvoiceID возвращает трансферы по инвойсу.
func (r *InventoryTransferRepo) GetByInvoiceID(invoiceID int64) ([]inventory.InventoryTransfer, error) {
	var rows []inventory.InventoryTransfer
	return rows, r.DB.Where("invoice_id = ?", invoiceID).Find(&rows).Error
}

// GetByInventoryID возвращает историю трансферов единицы инвентаря.
func (r *InventoryTransferRepo) GetByInventoryID(inventoryID int64) ([]inventory.InventoryTransfer, error) {
	var rows []inventory.InventoryTransfer
	return rows, r.DB.
		Where("inventory_id = ?", inventoryID).
		Order("date_transfer DESC").
		Find(&rows).Error
}

// GetByLocation возвращает трансферы куда/откуда локации.
func (r *InventoryTransferRepo) GetByLocation(locationID int64) ([]inventory.InventoryTransfer, error) {
	var rows []inventory.InventoryTransfer
	return rows, r.DB.
		Where("from_location_id = ? OR to_location_id = ?", locationID, locationID).
		Order("date_transfer DESC").
		Find(&rows).Error
}

// CreateTransferInput — данные нового трансфера.
type CreateTransferInput struct {
	InventoryID    int64
	FromLocationID int64
	ToLocationID   int64
	TransferredBy  int64
	ReceivedBy     int64
	StatusItems    types.StatusItemsInventory
	InvoiceID      int64
	OldInvoiceID   *int64
	InvoiceFrom    *int64
	InvoiceTo      *int64
	SystemNote     *string
}

// Create создаёт запись трансфера.
func (r *InventoryTransferRepo) Create(inp CreateTransferInput) (*inventory.InventoryTransfer, error) {
	tr := &inventory.InventoryTransfer{
		InventoryID:    inp.InventoryID,
		FromLocationID: inp.FromLocationID,
		ToLocationID:   inp.ToLocationID,
		DateTransfer:   time.Now(),
		TransferredBy:  inp.TransferredBy,
		ReceivedBy:     inp.ReceivedBy,
		StatusItems:    inp.StatusItems,
		InvoiceID:      inp.InvoiceID,
		OldInvoiceID:   inp.OldInvoiceID,
		InvoiceFrom:    inp.InvoiceFrom,
		InvoiceTo:      inp.InvoiceTo,
		SystemNote:     inp.SystemNote,
	}
	return tr, r.DB.Create(tr).Error
}

// CreateWithTx создаёт трансфер внутри существующей DB-транзакции.
func (r *InventoryTransferRepo) CreateWithTx(tx *gorm.DB, inp CreateTransferInput) (*inventory.InventoryTransfer, error) {
	tr := &inventory.InventoryTransfer{
		InventoryID:    inp.InventoryID,
		FromLocationID: inp.FromLocationID,
		ToLocationID:   inp.ToLocationID,
		DateTransfer:   time.Now(),
		TransferredBy:  inp.TransferredBy,
		ReceivedBy:     inp.ReceivedBy,
		StatusItems:    inp.StatusItems,
		InvoiceID:      inp.InvoiceID,
		OldInvoiceID:   inp.OldInvoiceID,
		InvoiceFrom:    inp.InvoiceFrom,
		InvoiceTo:      inp.InvoiceTo,
		SystemNote:     inp.SystemNote,
	}
	return tr, tx.Create(tr).Error
}

// UpdateStatus обновляет статус трансфера.
func (r *InventoryTransferRepo) UpdateStatus(id int64, status types.StatusItemsInventory) error {
	return r.DB.Model(&inventory.InventoryTransfer{}).
		Where("id_transfer = ?", id).
		Update("status_items", status).Error
}

func (r *InventoryTransferRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
