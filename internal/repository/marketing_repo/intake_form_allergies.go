package marketing_repo

import (
	"gorm.io/gorm"
	"sighthub-backend/internal/models/marketing"
)

type IntakeFormAllergiesRepo struct{ DB *gorm.DB }

func NewIntakeFormAllergiesRepo(db *gorm.DB) *IntakeFormAllergiesRepo {
	return &IntakeFormAllergiesRepo{DB: db}
}

func (r *IntakeFormAllergiesRepo) GetByRequestID(requestID int64) ([]marketing.IntakeFormAllergies, error) {
	var items []marketing.IntakeFormAllergies
	if err := r.DB.Where("request_id = ?", requestID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *IntakeFormAllergiesRepo) Create(item *marketing.IntakeFormAllergies) error {
	return r.DB.Create(item).Error
}

func (r *IntakeFormAllergiesRepo) Delete(id int64) error {
	return r.DB.Delete(&marketing.IntakeFormAllergies{}, id).Error
}
