// internal/repository/inventory_repo/inventory_transaction.go
package inventory_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/inventory"
	"sighthub-backend/internal/models/types"
)

type InventoryTransactionRepo struct{ DB *gorm.DB }

func NewInventoryTransactionRepo(db *gorm.DB) *InventoryTransactionRepo {
	return &InventoryTransactionRepo{DB: db}
}

// GetByInventoryID возвращает историю транзакций для единицы инвентаря.
func (r *InventoryTransactionRepo) GetByInventoryID(inventoryID int64) ([]inventory.InventoryTransaction, error) {
	var rows []inventory.InventoryTransaction
	return rows, r.DB.
		Where("inventory_id = ?", inventoryID).
		Order("date_transaction DESC").
		Find(&rows).Error
}

// GetByInvoiceID возвращает транзакции по инвойсу.
func (r *InventoryTransactionRepo) GetByInvoiceID(invoiceID int64) ([]inventory.InventoryTransaction, error) {
	var rows []inventory.InventoryTransaction
	return rows, r.DB.Where("invoice_id = ?", invoiceID).Find(&rows).Error
}

// CreateInput — данные новой транзакции.
type CreateTransactionInput struct {
	InventoryID      int64
	FromLocationID   int64
	ToLocationID     int64
	TransferredBy    int64
	InvoiceID        int64
	OldInvoiceID     *int64
	StatusItems      types.StatusItemsInventory
	TransactionType  string
	InventoryCountID *int64
	Notes            *string
}

// Create записывает транзакцию инвентаря.
func (r *InventoryTransactionRepo) Create(inp CreateTransactionInput) (*inventory.InventoryTransaction, error) {
	txn := &inventory.InventoryTransaction{
		InventoryID:      inp.InventoryID,
		FromLocationID:   inp.FromLocationID,
		ToLocationID:     inp.ToLocationID,
		TransferredBy:    inp.TransferredBy,
		InvoiceID:        inp.InvoiceID,
		OldInvoiceID:     inp.OldInvoiceID,
		StatusItems:      inp.StatusItems,
		TransactionType:  inp.TransactionType,
		DateTransaction:  time.Now(),
		InventoryCountID: inp.InventoryCountID,
		Notes:            inp.Notes,
	}
	if txn.TransactionType == "" {
		txn.TransactionType = "Transfer"
	}
	return txn, r.DB.Create(txn).Error
}

// CreateWithTx создаёт транзакцию внутри существующей DB-транзакции.
func (r *InventoryTransactionRepo) CreateWithTx(tx *gorm.DB, inp CreateTransactionInput) (*inventory.InventoryTransaction, error) {
	txn := &inventory.InventoryTransaction{
		InventoryID:      inp.InventoryID,
		FromLocationID:   inp.FromLocationID,
		ToLocationID:     inp.ToLocationID,
		TransferredBy:    inp.TransferredBy,
		InvoiceID:        inp.InvoiceID,
		OldInvoiceID:     inp.OldInvoiceID,
		StatusItems:      inp.StatusItems,
		TransactionType:  inp.TransactionType,
		DateTransaction:  time.Now(),
		InventoryCountID: inp.InventoryCountID,
		Notes:            inp.Notes,
	}
	if txn.TransactionType == "" {
		txn.TransactionType = "Transfer"
	}
	return txn, tx.Create(txn).Error
}

func (r *InventoryTransactionRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
