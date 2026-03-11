package vendors_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vendors"
)

type AgreementRepo struct{ DB *gorm.DB }

func NewAgreementRepo(db *gorm.DB) *AgreementRepo { return &AgreementRepo{DB: db} }

func (r *AgreementRepo) GetByID(id int) (*vendors.Agreement, error) {
	var item vendors.Agreement
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *AgreementRepo) GetByVendorID(vendorID int) ([]vendors.Agreement, error) {
	var items []vendors.Agreement
	return items, r.DB.Where("vendor_id = ?", vendorID).Find(&items).Error
}

func (r *AgreementRepo) Create(item *vendors.Agreement) error {
	return r.DB.Create(item).Error
}

func (r *AgreementRepo) Save(item *vendors.Agreement) error {
	return r.DB.Save(item).Error
}

func (r *AgreementRepo) Delete(id int) error {
	return r.DB.Delete(&vendors.Agreement{}, id).Error
}
