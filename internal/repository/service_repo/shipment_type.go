package service_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/service"
)

type ShipmentTypeRepo struct{ DB *gorm.DB }

func NewShipmentTypeRepo(db *gorm.DB) *ShipmentTypeRepo { return &ShipmentTypeRepo{DB: db} }

func (r *ShipmentTypeRepo) GetAll() ([]service.ShipmentType, error) {
	var items []service.ShipmentType
	return items, r.DB.Find(&items).Error
}

func (r *ShipmentTypeRepo) GetByID(id int) (*service.ShipmentType, error) {
	var item service.ShipmentType
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}
