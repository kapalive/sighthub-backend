package vendors_repo

import (
	"gorm.io/gorm"
	"sighthub-backend/internal/models/vendors"
)

// LabRepo — labs are now vendors with lab=true
type LabRepo struct{ DB *gorm.DB }

func NewLabRepo(db *gorm.DB) *LabRepo { return &LabRepo{DB: db} }

func (r *LabRepo) GetAll() ([]vendors.Vendor, error) {
	var items []vendors.Vendor
	return items, r.DB.Where("lab = true AND visible = true").Order("vendor_name").Find(&items).Error
}

func (r *LabRepo) GetByID(id int) (*vendors.Vendor, error) {
	var item vendors.Vendor
	if err := r.DB.First(&item, "id_vendor = ? AND lab = true", id).Error; err != nil {
		return nil, nil
	}
	return &item, nil
}
