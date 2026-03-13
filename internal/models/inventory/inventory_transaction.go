// internal/models/inventory/inventory_transaction.go
package inventory

import (
	"fmt"
	"sighthub-backend/internal/models/types"
	"time"
)

type InventoryTransaction struct {
	IDTransaction    int64                      `gorm:"column:id_transaction;primaryKey;autoIncrement" json:"id_transaction"`
	InventoryID      *int64                     `gorm:"column:inventory_id" json:"inventory_id,omitempty"`
	FromLocationID   *int64                     `gorm:"column:from_location_id" json:"from_location_id,omitempty"`
	ToLocationID     *int64                     `gorm:"column:to_location_id" json:"to_location_id,omitempty"`
	TransferredBy    int64                      `gorm:"column:transferred_by;not null" json:"transferred_by"`
	InvoiceID        *int64                     `gorm:"column:invoice_id" json:"invoice_id,omitempty"`
	OldInvoiceID     *int64                     `gorm:"column:old_invoice_id" json:"old_invoice_id,omitempty"`
	StatusItems      types.StatusItemsInventory `gorm:"column:status_items;not null" json:"status_items"`
	TransactionType  string                     `gorm:"column:transaction_type;default:'Transfer';not null" json:"transaction_type"`
	DateTransaction  time.Time                  `gorm:"column:date_transaction;default:CURRENT_TIMESTAMP" json:"date_transaction"`
	InventoryCountID *int64                     `gorm:"column:inventory_count_id" json:"inventory_count_id,omitempty"`
	Notes            *string                    `gorm:"column:notes" json:"notes,omitempty"`
}

func (InventoryTransaction) TableName() string {
	return "inventory_transaction"
}

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

func (i *InventoryTransaction) String() string {
	return fmt.Sprintf("<InventoryTransaction %d>", i.IDTransaction)
}
