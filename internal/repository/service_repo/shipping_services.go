package service_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/service"
)

type ShippingServicesRepo struct{ DB *gorm.DB }

func NewShippingServicesRepo(db *gorm.DB) *ShippingServicesRepo {
	return &ShippingServicesRepo{DB: db}
}

func (r *ShippingServicesRepo) GetAll() ([]service.ShippingServices, error) {
	var items []service.ShippingServices
	return items, r.DB.Find(&items).Error
}

func (r *ShippingServicesRepo) GetByID(id int) (*service.ShippingServices, error) {
	var item service.ShippingServices
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}
