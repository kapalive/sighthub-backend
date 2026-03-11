package vendors_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vendors"
)

type VendorRepo struct{ DB *gorm.DB }

func NewVendorRepo(db *gorm.DB) *VendorRepo { return &VendorRepo{DB: db} }

func (r *VendorRepo) GetAll() ([]vendors.Vendor, error) {
	var items []vendors.Vendor
	return items, r.DB.Find(&items).Error
}

func (r *VendorRepo) GetByID(id int) (*vendors.Vendor, error) {
	var item vendors.Vendor
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *VendorRepo) Search(query string) ([]vendors.Vendor, error) {
	var items []vendors.Vendor
	q := "%" + query + "%"
	return items, r.DB.Where("vendor_name ILIKE ? OR short_name ILIKE ?", q, q).Find(&items).Error
}

func (r *VendorRepo) Create(item *vendors.Vendor) error {
	return r.DB.Create(item).Error
}

func (r *VendorRepo) Save(item *vendors.Vendor) error {
	return r.DB.Save(item).Error
}

func (r *VendorRepo) Delete(id int) error {
	return r.DB.Delete(&vendors.Vendor{}, id).Error
}
