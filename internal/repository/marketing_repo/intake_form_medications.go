package marketing_repo

import (
	"gorm.io/gorm"
	"sighthub-backend/internal/models/marketing"
)

type IntakeFormMedicationsRepo struct{ DB *gorm.DB }

func NewIntakeFormMedicationsRepo(db *gorm.DB) *IntakeFormMedicationsRepo {
	return &IntakeFormMedicationsRepo{DB: db}
}

func (r *IntakeFormMedicationsRepo) GetByRequestID(requestID int64) ([]marketing.IntakeFormMedications, error) {
	var items []marketing.IntakeFormMedications
	if err := r.DB.Where("request_id = ?", requestID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *IntakeFormMedicationsRepo) Create(item *marketing.IntakeFormMedications) error {
	return r.DB.Create(item).Error
}

func (r *IntakeFormMedicationsRepo) Delete(id int64) error {
	return r.DB.Delete(&marketing.IntakeFormMedications{}, id).Error
}
