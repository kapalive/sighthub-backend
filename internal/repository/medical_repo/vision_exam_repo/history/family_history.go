// internal/repository/medical_repo/vision_exam_repo/history/family_history.go
package history

import (
	"gorm.io/gorm"
	"sighthub-backend/internal/models/medical/vision_exam/history"
)

type FamilyHistoryRepo struct{ DB *gorm.DB }

func NewFamilyHistoryRepo(db *gorm.DB) *FamilyHistoryRepo {
	return &FamilyHistoryRepo{DB: db}
}

func (r *FamilyHistoryRepo) GetByID(id int64) (*history.FamilyHistory, error) {
	var v history.FamilyHistory
	if err := r.DB.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *FamilyHistoryRepo) Create() (*history.FamilyHistory, error) {
	v := history.FamilyHistory{}
	if err := r.DB.Create(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *FamilyHistoryRepo) Save(v *history.FamilyHistory) error {
	return r.DB.Save(v).Error
}
