package marketing_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/marketing"
)

type GiftCardRepo struct{ DB *gorm.DB }

func NewGiftCardRepo(db *gorm.DB) *GiftCardRepo { return &GiftCardRepo{DB: db} }

func (r *GiftCardRepo) GetByID(id int) (*marketing.GiftCard, error) {
	var gc marketing.GiftCard
	if err := r.DB.First(&gc, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &gc, nil
}

func (r *GiftCardRepo) GetByCode(code string) (*marketing.GiftCard, error) {
	var gc marketing.GiftCard
	if err := r.DB.Where("code = ?", code).First(&gc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &gc, nil
}

func (r *GiftCardRepo) GetByLocationID(locationID int) ([]marketing.GiftCard, error) {
	var gcs []marketing.GiftCard
	if err := r.DB.Where("location_id = ?", locationID).Find(&gcs).Error; err != nil {
		return nil, err
	}
	return gcs, nil
}

func (r *GiftCardRepo) Create(gc *marketing.GiftCard) error {
	return r.DB.Create(gc).Error
}

func (r *GiftCardRepo) Save(gc *marketing.GiftCard) error {
	return r.DB.Save(gc).Error
}
