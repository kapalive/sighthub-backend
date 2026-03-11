package vendors_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vendors"
)

type ReturnToVendorItemRepo struct{ DB *gorm.DB }

func NewReturnToVendorItemRepo(db *gorm.DB) *ReturnToVendorItemRepo {
	return &ReturnToVendorItemRepo{DB: db}
}

func (r *ReturnToVendorItemRepo) GetByID(id int64) (*vendors.ReturnToVendorItem, error) {
	var item vendors.ReturnToVendorItem
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *ReturnToVendorItemRepo) GetByInvoiceID(invoiceID int64) ([]vendors.ReturnToVendorItem, error) {
	var items []vendors.ReturnToVendorItem
	return items, r.DB.Where("return_to_vendor_invoice_id = ?", invoiceID).Find(&items).Error
}

func (r *ReturnToVendorItemRepo) Create(item *vendors.ReturnToVendorItem) error {
	return r.DB.Create(item).Error
}

func (r *ReturnToVendorItemRepo) Delete(id int64) error {
	return r.DB.Delete(&vendors.ReturnToVendorItem{}, id).Error
}

func (r *ReturnToVendorItemRepo) DeleteByInvoiceID(invoiceID int64) error {
	return r.DB.Where("return_to_vendor_invoice_id = ?", invoiceID).Delete(&vendors.ReturnToVendorItem{}).Error
}
