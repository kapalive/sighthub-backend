package vendors_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vendors"
)

type TypeItemsOfBrandRepo struct{ DB *gorm.DB }

func NewTypeItemsOfBrandRepo(db *gorm.DB) *TypeItemsOfBrandRepo {
	return &TypeItemsOfBrandRepo{DB: db}
}

func (r *TypeItemsOfBrandRepo) GetAll() ([]vendors.TypeItemsOfBrand, error) {
	var items []vendors.TypeItemsOfBrand
	return items, r.DB.Find(&items).Error
}

func (r *TypeItemsOfBrandRepo) GetByID(id int) (*vendors.TypeItemsOfBrand, error) {
	var item vendors.TypeItemsOfBrand
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}
