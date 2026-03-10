package general_repo

import (
	"errors"
	"time"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/general"
)

type PaymentTransactionRepo struct{ DB *gorm.DB }

func NewPaymentTransactionRepo(db *gorm.DB) *PaymentTransactionRepo {
	return &PaymentTransactionRepo{DB: db}
}

func (r *PaymentTransactionRepo) GetByID(id int64) (*general.PaymentTransaction, error) {
	var v general.PaymentTransaction
	if err := r.DB.Preload("Terminal").First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *PaymentTransactionRepo) GetByTerminalID(terminalID int, from, to time.Time) ([]general.PaymentTransaction, error) {
	var items []general.PaymentTransaction
	return items, r.DB.
		Where("payment_terminal_id = ? AND transaction_date BETWEEN ? AND ?", terminalID, from, to).
		Order("transaction_date DESC").
		Find(&items).Error
}

func (r *PaymentTransactionRepo) GetByReferenceNumber(ref string) (*general.PaymentTransaction, error) {
	var v general.PaymentTransaction
	if err := r.DB.Where("reference_number = ?", ref).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *PaymentTransactionRepo) Create(v *general.PaymentTransaction) error { return r.DB.Create(v).Error }
func (r *PaymentTransactionRepo) Save(v *general.PaymentTransaction) error   { return r.DB.Save(v).Error }

func (r *PaymentTransactionRepo) UpdateStatus(id int64, status string) error {
	return r.DB.Model(&general.PaymentTransaction{}).
		Where("id_payment_transaction = ?", id).
		Update("status", status).Error
}
