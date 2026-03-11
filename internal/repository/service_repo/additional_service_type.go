package service_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/service"
)

type AdditionalServiceTypeRepo struct{ DB *gorm.DB }

func NewAdditionalServiceTypeRepo(db *gorm.DB) *AdditionalServiceTypeRepo {
	return &AdditionalServiceTypeRepo{DB: db}
}

func (r *AdditionalServiceTypeRepo) GetAll() ([]service.AdditionalServiceType, error) {
	var items []service.AdditionalServiceType
	return items, r.DB.Find(&items).Error
}

func (r *AdditionalServiceTypeRepo) GetByID(id int) (*service.AdditionalServiceType, error) {
	var item service.AdditionalServiceType
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}
