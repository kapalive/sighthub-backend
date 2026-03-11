package vendors_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vendors"
)

type VendorBrandContactLensRepo struct{ DB *gorm.DB }

func NewVendorBrandContactLensRepo(db *gorm.DB) *VendorBrandContactLensRepo {
	return &VendorBrandContactLensRepo{DB: db}
}

func (r *VendorBrandContactLensRepo) GetByID(id int) (*vendors.VendorBrandContactLens, error) {
	var item vendors.VendorBrandContactLens
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *VendorBrandContactLensRepo) GetByVendorID(vendorID int) ([]vendors.VendorBrandContactLens, error) {
	var items []vendors.VendorBrandContactLens
	return items, r.DB.Where("id_vendor = ?", vendorID).Find(&items).Error
}

func (r *VendorBrandContactLensRepo) Add(vendorID, brandContactLensID int) error {
	return r.DB.Create(&vendors.VendorBrandContactLens{
		IDVendor: vendorID, IDBrandContactLens: brandContactLensID,
	}).Error
}

func (r *VendorBrandContactLensRepo) Remove(id int) error {
	return r.DB.Delete(&vendors.VendorBrandContactLens{}, id).Error
}

func (r *VendorBrandContactLensRepo) RemoveByVendorAndBrand(vendorID, brandID int) error {
	return r.DB.Where("id_vendor = ? AND id_brand_contact_lens = ?", vendorID, brandID).
		Delete(&vendors.VendorBrandContactLens{}).Error
}
