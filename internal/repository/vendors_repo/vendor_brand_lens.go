package vendors_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vendors"
)

type VendorBrandLensRepo struct{ DB *gorm.DB }

func NewVendorBrandLensRepo(db *gorm.DB) *VendorBrandLensRepo {
	return &VendorBrandLensRepo{DB: db}
}

func (r *VendorBrandLensRepo) GetByID(id int) (*vendors.VendorBrandLens, error) {
	var item vendors.VendorBrandLens
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *VendorBrandLensRepo) GetByVendorID(vendorID int) ([]vendors.VendorBrandLens, error) {
	var items []vendors.VendorBrandLens
	return items, r.DB.Where("id_vendor = ?", vendorID).Find(&items).Error
}

func (r *VendorBrandLensRepo) Add(vendorID, brandLensID int) error {
	return r.DB.Create(&vendors.VendorBrandLens{
		IDVendor: vendorID, IDBrandLens: brandLensID,
	}).Error
}

func (r *VendorBrandLensRepo) Remove(id int) error {
	return r.DB.Delete(&vendors.VendorBrandLens{}, id).Error
}

func (r *VendorBrandLensRepo) RemoveByVendorAndBrand(vendorID, brandLensID int) error {
	return r.DB.Where("id_vendor = ? AND id_brand_lens = ?", vendorID, brandLensID).
		Delete(&vendors.VendorBrandLens{}).Error
}
