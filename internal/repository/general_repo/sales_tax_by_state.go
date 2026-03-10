package general_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/general"
)

type SalesTaxByStateRepo struct{ DB *gorm.DB }

func NewSalesTaxByStateRepo(db *gorm.DB) *SalesTaxByStateRepo { return &SalesTaxByStateRepo{DB: db} }

func (r *SalesTaxByStateRepo) GetAll() ([]general.SalesTaxByState, error) {
	var items []general.SalesTaxByState
	return items, r.DB.Order("state_code").Find(&items).Error
}

func (r *SalesTaxByStateRepo) GetByID(id int) (*general.SalesTaxByState, error) {
	var v general.SalesTaxByState
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *SalesTaxByStateRepo) GetActiveByStateCode(stateCode string) (*general.SalesTaxByState, error) {
	var v general.SalesTaxByState
	if err := r.DB.
		Where("state_code = ? AND tax_active = true", stateCode).
		Order("effective_date DESC").
		First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *SalesTaxByStateRepo) Create(v *general.SalesTaxByState) error { return r.DB.Create(v).Error }
func (r *SalesTaxByStateRepo) Save(v *general.SalesTaxByState) error   { return r.DB.Save(v).Error }
