package general_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/general"
)

type EthnicityRepo struct{ DB *gorm.DB }

func NewEthnicityRepo(db *gorm.DB) *EthnicityRepo { return &EthnicityRepo{DB: db} }

func (r *EthnicityRepo) GetAll() ([]general.Ethnicity, error) {
	var items []general.Ethnicity
	return items, r.DB.Find(&items).Error
}

func (r *EthnicityRepo) GetByID(id int) (*general.Ethnicity, error) {
	var v general.Ethnicity
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}
