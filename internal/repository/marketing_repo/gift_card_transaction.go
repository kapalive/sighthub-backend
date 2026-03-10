package marketing_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/marketing"
)

type GiftCardTransactionRepo struct{ DB *gorm.DB }

func NewGiftCardTransactionRepo(db *gorm.DB) *GiftCardTransactionRepo {
	return &GiftCardTransactionRepo{DB: db}
}

func (r *GiftCardTransactionRepo) GetByID(id int) (*marketing.GiftCardTransaction, error) {
	var t marketing.GiftCardTransaction
	if err := r.DB.First(&t, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *GiftCardTransactionRepo) GetByGiftCardID(giftCardID int) ([]marketing.GiftCardTransaction, error) {
	var txs []marketing.GiftCardTransaction
	if err := r.DB.Where("gift_card_id = ?", giftCardID).Find(&txs).Error; err != nil {
		return nil, err
	}
	return txs, nil
}

func (r *GiftCardTransactionRepo) Create(t *marketing.GiftCardTransaction) error {
	return r.DB.Create(t).Error
}
