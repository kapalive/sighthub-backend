package audit_repo

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/audit"
)

type PaymentTransactionLogRepo struct{ DB *gorm.DB }

func NewPaymentTransactionLogRepo(db *gorm.DB) *PaymentTransactionLogRepo {
	return &PaymentTransactionLogRepo{DB: db}
}

func (r *PaymentTransactionLogRepo) GetByID(id int64) (*audit.PaymentTransactionLog, error) {
	var v audit.PaymentTransactionLog
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *PaymentTransactionLogRepo) GetByTransactionID(txID int64) ([]audit.PaymentTransactionLog, error) {
	var items []audit.PaymentTransactionLog
	return items, r.DB.Where("payment_transaction_id = ?", txID).Order("logged_at").Find(&items).Error
}

func (r *PaymentTransactionLogRepo) Create(v *audit.PaymentTransactionLog) error {
	return r.DB.Create(v).Error
}

func (r *PaymentTransactionLogRepo) Log(txID int64, message string) error {
	v := audit.PaymentTransactionLog{
		PaymentTransactionID: txID,
		LogMessage:           message,
	}
	return r.DB.Create(&v).Error
}
