package general_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/general"
)

type ServiceTaxByStateRepo struct{ DB *gorm.DB }

func NewServiceTaxByStateRepo(db *gorm.DB) *ServiceTaxByStateRepo {
	return &ServiceTaxByStateRepo{DB: db}
}

func (r *ServiceTaxByStateRepo) GetAll() ([]general.ServiceTaxByState, error) {
	var items []general.ServiceTaxByState
	return items, r.DB.Order("state_code").Find(&items).Error
}

func (r *ServiceTaxByStateRepo) GetByID(id int) (*general.ServiceTaxByState, error) {
	var v general.ServiceTaxByState
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *ServiceTaxByStateRepo) GetActiveByStateCode(stateCode string) (*general.ServiceTaxByState, error) {
	var v general.ServiceTaxByState
	if err := r.DB.
		Where("state_code = ? AND tax_active = true", stateCode).
		Order("effective_date DESC").
		First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *ServiceTaxByStateRepo) Create(v *general.ServiceTaxByState) error { return r.DB.Create(v).Error }
func (r *ServiceTaxByStateRepo) Save(v *general.ServiceTaxByState) error   { return r.DB.Save(v).Error }
