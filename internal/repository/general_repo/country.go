package general_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/general"
)

type CountryRepo struct{ DB *gorm.DB }

func NewCountryRepo(db *gorm.DB) *CountryRepo { return &CountryRepo{DB: db} }

func (r *CountryRepo) GetAll() ([]general.Country, error) {
	var items []general.Country
	return items, r.DB.Order("country").Find(&items).Error
}

func (r *CountryRepo) GetByID(id int) (*general.Country, error) {
	var v general.Country
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *CountryRepo) GetByCode(code string) (*general.Country, error) {
	var v general.Country
	if err := r.DB.Where("code = ?", code).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}
