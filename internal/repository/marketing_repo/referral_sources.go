package marketing_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/marketing"
)

type ReferralSourceRepo struct{ DB *gorm.DB }

func NewReferralSourceRepo(db *gorm.DB) *ReferralSourceRepo { return &ReferralSourceRepo{DB: db} }

func (r *ReferralSourceRepo) GetAll() ([]marketing.ReferralSource, error) {
	var items []marketing.ReferralSource
	if err := r.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ReferralSourceRepo) GetByID(id int) (*marketing.ReferralSource, error) {
	var item marketing.ReferralSource
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *ReferralSourceRepo) Create(item *marketing.ReferralSource) error {
	return r.DB.Create(item).Error
}

func (r *ReferralSourceRepo) Save(item *marketing.ReferralSource) error {
	return r.DB.Save(item).Error
}

func (r *ReferralSourceRepo) Delete(id int) error {
	return r.DB.Delete(&marketing.ReferralSource{}, id).Error
}
