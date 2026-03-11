package vendors_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vendors"
)

type RepContactVendorRepo struct{ DB *gorm.DB }

func NewRepContactVendorRepo(db *gorm.DB) *RepContactVendorRepo {
	return &RepContactVendorRepo{DB: db}
}

func (r *RepContactVendorRepo) GetByID(id int) (*vendors.RepContactVendor, error) {
	var item vendors.RepContactVendor
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *RepContactVendorRepo) GetByVendorID(vendorID int) ([]vendors.RepContactVendor, error) {
	var items []vendors.RepContactVendor
	return items, r.DB.Where("vendor_id = ?", vendorID).Find(&items).Error
}

func (r *RepContactVendorRepo) Create(item *vendors.RepContactVendor) error {
	return r.DB.Create(item).Error
}

func (r *RepContactVendorRepo) Save(item *vendors.RepContactVendor) error {
	return r.DB.Save(item).Error
}

func (r *RepContactVendorRepo) Delete(id int) error {
	return r.DB.Delete(&vendors.RepContactVendor{}, id).Error
}
