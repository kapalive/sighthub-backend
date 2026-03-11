package vendors_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vendors"
)

type VendorLabsRepo struct{ DB *gorm.DB }

func NewVendorLabsRepo(db *gorm.DB) *VendorLabsRepo { return &VendorLabsRepo{DB: db} }

func (r *VendorLabsRepo) GetByID(id int) (*vendors.VendorLabs, error) {
	var item vendors.VendorLabs
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *VendorLabsRepo) GetByVendorID(vendorID int) ([]vendors.VendorLabs, error) {
	var items []vendors.VendorLabs
	return items, r.DB.Where("vendor_id = ?", vendorID).Find(&items).Error
}

func (r *VendorLabsRepo) GetLabIDsByVendor(vendorID int) ([]int, error) {
	var ids []int
	return ids, r.DB.Model(&vendors.VendorLabs{}).
		Where("vendor_id = ?", vendorID).Pluck("lab_id", &ids).Error
}

func (r *VendorLabsRepo) Add(vendorID, labID int) error {
	return r.DB.Create(&vendors.VendorLabs{VendorID: vendorID, LabID: labID}).Error
}

func (r *VendorLabsRepo) Remove(id int) error {
	return r.DB.Delete(&vendors.VendorLabs{}, id).Error
}

func (r *VendorLabsRepo) RemoveByVendorAndLab(vendorID, labID int) error {
	return r.DB.Where("vendor_id = ? AND lab_id = ?", vendorID, labID).
		Delete(&vendors.VendorLabs{}).Error
}
