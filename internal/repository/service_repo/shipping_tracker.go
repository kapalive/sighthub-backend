package service_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/service"
)

type ShippingTrackerRepo struct{ DB *gorm.DB }

func NewShippingTrackerRepo(db *gorm.DB) *ShippingTrackerRepo {
	return &ShippingTrackerRepo{DB: db}
}

func (r *ShippingTrackerRepo) GetByID(id int64) (*service.ShippingTracker, error) {
	var item service.ShippingTracker
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *ShippingTrackerRepo) GetByTracker(tracker string) (*service.ShippingTracker, error) {
	var item service.ShippingTracker
	if err := r.DB.Where("tracker = ?", tracker).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *ShippingTrackerRepo) Create(item *service.ShippingTracker) error {
	return r.DB.Create(item).Error
}

func (r *ShippingTrackerRepo) Save(item *service.ShippingTracker) error {
	return r.DB.Save(item).Error
}

func (r *ShippingTrackerRepo) Delete(id int64) error {
	return r.DB.Delete(&service.ShippingTracker{}, id).Error
}
