package contact_lens_repo

import (
	"errors"

	"gorm.io/gorm"

	contactlens "sighthub-backend/internal/models/contact_lens"
)

type ContactLensItemRepo struct{ DB *gorm.DB }

func NewContactLensItemRepo(db *gorm.DB) *ContactLensItemRepo {
	return &ContactLensItemRepo{DB: db}
}

func (r *ContactLensItemRepo) GetByID(id int) (*contactlens.ContactLensItem, error) {
	var v contactlens.ContactLensItem
	if err := r.DB.Preload("Brand").First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *ContactLensItemRepo) GetByVendorID(vendorID int) ([]contactlens.ContactLensItem, error) {
	var items []contactlens.ContactLensItem
	return items, r.DB.Preload("Brand").Where("vendor_id = ?", vendorID).Find(&items).Error
}

func (r *ContactLensItemRepo) GetByBrandID(brandID int) ([]contactlens.ContactLensItem, error) {
	var items []contactlens.ContactLensItem
	return items, r.DB.Preload("Brand").Where("brand_contact_lens_id = ?", brandID).Find(&items).Error
}

func (r *ContactLensItemRepo) Search(query string) ([]contactlens.ContactLensItem, error) {
	var items []contactlens.ContactLensItem
	return items, r.DB.Preload("Brand").
		Where("name_contact ILIKE ? OR model ILIKE ? OR invoice_desc ILIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%").
		Find(&items).Error
}

func (r *ContactLensItemRepo) Create(v *contactlens.ContactLensItem) error {
	return r.DB.Create(v).Error
}

func (r *ContactLensItemRepo) Save(v *contactlens.ContactLensItem) error {
	return r.DB.Save(v).Error
}

func (r *ContactLensItemRepo) Delete(id int) error {
	return r.DB.Delete(&contactlens.ContactLensItem{}, id).Error
}
