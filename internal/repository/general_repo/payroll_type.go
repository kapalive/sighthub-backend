package general_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/general"
)

type PayrollTypeRepo struct{ DB *gorm.DB }

func NewPayrollTypeRepo(db *gorm.DB) *PayrollTypeRepo { return &PayrollTypeRepo{DB: db} }

func (r *PayrollTypeRepo) GetAll() ([]general.PayrollType, error) {
	var items []general.PayrollType
	return items, r.DB.Find(&items).Error
}

func (r *PayrollTypeRepo) GetByID(id int) (*general.PayrollType, error) {
	var v general.PayrollType
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}
