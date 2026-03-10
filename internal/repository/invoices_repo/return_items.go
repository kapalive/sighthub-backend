// internal/repository/invoices_repo/return_items.go
package invoices_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/invoices"
)

type ReturnItemRepo struct{ DB *gorm.DB }

func NewReturnItemRepo(db *gorm.DB) *ReturnItemRepo { return &ReturnItemRepo{DB: db} }

func (r *ReturnItemRepo) GetByReturnID(returnID int64) ([]invoices.ReturnItem, error) {
	var rows []invoices.ReturnItem
	err := r.DB.Preload("InvoiceItemSale").
		Where("return_id = ?", returnID).
		Find(&rows).Error
	return rows, err
}

func (r *ReturnItemRepo) Create(item *invoices.ReturnItem) error {
	return r.DB.Create(item).Error
}

func (r *ReturnItemRepo) Delete(id int64) error {
	return r.DB.Delete(&invoices.ReturnItem{}, id).Error
}

func (r *ReturnItemRepo) DeleteByReturnID(returnID int64) error {
	return r.DB.Where("return_id = ?", returnID).Delete(&invoices.ReturnItem{}).Error
}

func (r *ReturnItemRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
