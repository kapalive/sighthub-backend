package vendors_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vendors"
)

type VendorBrandRepo struct{ DB *gorm.DB }

func NewVendorBrandRepo(db *gorm.DB) *VendorBrandRepo { return &VendorBrandRepo{DB: db} }

func (r *VendorBrandRepo) GetByID(id int) (*vendors.VendorBrand, error) {
	var item vendors.VendorBrand
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *VendorBrandRepo) GetByVendorID(vendorID int) ([]vendors.VendorBrand, error) {
	var items []vendors.VendorBrand
	return items, r.DB.Where("id_vendor = ?", vendorID).Find(&items).Error
}

func (r *VendorBrandRepo) GetBrandIDsByVendor(vendorID int) ([]int, error) {
	var ids []int
	return ids, r.DB.Model(&vendors.VendorBrand{}).
		Where("id_vendor = ?", vendorID).Pluck("id_brand", &ids).Error
}

func (r *VendorBrandRepo) Add(vendorID, brandID int) error {
	return r.DB.Create(&vendors.VendorBrand{IDVendor: vendorID, IDBrand: brandID}).Error
}

func (r *VendorBrandRepo) Remove(id int) error {
	return r.DB.Delete(&vendors.VendorBrand{}, id).Error
}

func (r *VendorBrandRepo) RemoveByVendorAndBrand(vendorID, brandID int) error {
	return r.DB.Where("id_vendor = ? AND id_brand = ?", vendorID, brandID).
		Delete(&vendors.VendorBrand{}).Error
}
