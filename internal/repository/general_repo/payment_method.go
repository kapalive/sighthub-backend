package general_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/general"
)

type PaymentMethodRepo struct{ DB *gorm.DB }

func NewPaymentMethodRepo(db *gorm.DB) *PaymentMethodRepo { return &PaymentMethodRepo{DB: db} }

func (r *PaymentMethodRepo) GetAll() ([]general.PaymentMethod, error) {
	var items []general.PaymentMethod
	return items, r.DB.Find(&items).Error
}

func (r *PaymentMethodRepo) GetByID(id int) (*general.PaymentMethod, error) {
	var v general.PaymentMethod
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}
