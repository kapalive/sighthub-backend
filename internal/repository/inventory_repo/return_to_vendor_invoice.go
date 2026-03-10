// internal/repository/inventory_repo/return_to_vendor_invoice.go
// Этот файл — thin-прокси к vendors_repo.ReturnToVendorInvoiceRepo.
// Здесь хранится только inventory-специфичная операция: добавление позиции возврата.
package inventory_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vendors"
)

// ReturnToVendorItemRepo управляет позициями возврата поставщику на уровне инвентаря.
type ReturnToVendorItemRepo struct{ DB *gorm.DB }

func NewReturnToVendorItemRepo(db *gorm.DB) *ReturnToVendorItemRepo {
	return &ReturnToVendorItemRepo{DB: db}
}

// GetByInvoice возвращает все позиции возврата для данного return_to_vendor_invoice_id.
func (r *ReturnToVendorItemRepo) GetByInvoice(returnInvoiceID int64) ([]vendors.ReturnToVendorItem, error) {
	var rows []vendors.ReturnToVendorItem
	return rows, r.DB.Where("return_to_vendor_invoice_id = ?", returnInvoiceID).Find(&rows).Error
}

// AddItem добавляет позицию возврата.
func (r *ReturnToVendorItemRepo) AddItem(item *vendors.ReturnToVendorItem) error {
	return r.DB.Create(item).Error
}

// DeleteItem удаляет позицию возврата по ID.
func (r *ReturnToVendorItemRepo) DeleteItem(id int64) error {
	return r.DB.Delete(&vendors.ReturnToVendorItem{}, id).Error
}

// GetByInventoryID возвращает историю возвратов для единицы инвентаря.
func (r *ReturnToVendorItemRepo) GetByInventoryID(inventoryID int64) ([]vendors.ReturnToVendorItem, error) {
	var rows []vendors.ReturnToVendorItem
	return rows, r.DB.Where("inventory_id = ?", inventoryID).Find(&rows).Error
}

func (r *ReturnToVendorItemRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
