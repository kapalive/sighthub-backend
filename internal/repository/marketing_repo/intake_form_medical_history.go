package marketing_repo

import (
	"gorm.io/gorm"
	"sighthub-backend/internal/models/marketing"
)

type IntakeFormMedicalHistoryRepo struct{ DB *gorm.DB }

func NewIntakeFormMedicalHistoryRepo(db *gorm.DB) *IntakeFormMedicalHistoryRepo {
	return &IntakeFormMedicalHistoryRepo{DB: db}
}

func (r *IntakeFormMedicalHistoryRepo) GetByRequestID(requestID int64) ([]marketing.IntakeFormMedicalHistory, error) {
	var items []marketing.IntakeFormMedicalHistory
	if err := r.DB.Where("request_id = ?", requestID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *IntakeFormMedicalHistoryRepo) Create(item *marketing.IntakeFormMedicalHistory) error {
	return r.DB.Create(item).Error
}

func (r *IntakeFormMedicalHistoryRepo) Delete(id int64) error {
	return r.DB.Delete(&marketing.IntakeFormMedicalHistory{}, id).Error
}
