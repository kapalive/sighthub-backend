package marketing_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/marketing"
)

type VisitReasonRepo struct{ DB *gorm.DB }

func NewVisitReasonRepo(db *gorm.DB) *VisitReasonRepo { return &VisitReasonRepo{DB: db} }

func (r *VisitReasonRepo) GetAll() ([]marketing.VisitReason, error) {
	var items []marketing.VisitReason
	if err := r.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *VisitReasonRepo) GetByID(id int) (*marketing.VisitReason, error) {
	var item marketing.VisitReason
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *VisitReasonRepo) Create(item *marketing.VisitReason) error {
	return r.DB.Create(item).Error
}

func (r *VisitReasonRepo) Save(item *marketing.VisitReason) error {
	return r.DB.Save(item).Error
}

func (r *VisitReasonRepo) Delete(id int) error {
	return r.DB.Delete(&marketing.VisitReason{}, id).Error
}
