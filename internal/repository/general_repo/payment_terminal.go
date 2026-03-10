package general_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/general"
)

type PaymentTerminalRepo struct{ DB *gorm.DB }

func NewPaymentTerminalRepo(db *gorm.DB) *PaymentTerminalRepo { return &PaymentTerminalRepo{DB: db} }

func (r *PaymentTerminalRepo) GetAll() ([]general.PaymentTerminal, error) {
	var items []general.PaymentTerminal
	return items, r.DB.Find(&items).Error
}

func (r *PaymentTerminalRepo) GetByID(id int) (*general.PaymentTerminal, error) {
	var v general.PaymentTerminal
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *PaymentTerminalRepo) Create(v *general.PaymentTerminal) error { return r.DB.Create(v).Error }
func (r *PaymentTerminalRepo) Save(v *general.PaymentTerminal) error   { return r.DB.Save(v).Error }
func (r *PaymentTerminalRepo) Delete(id int) error {
	return r.DB.Delete(&general.PaymentTerminal{}, id).Error
}
