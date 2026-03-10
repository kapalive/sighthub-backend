// internal/repository/medical_repo/vision_exam_repo/history/ocular_history.go
package history

import (
	"gorm.io/gorm"
	"sighthub-backend/internal/models/medical/vision_exam/history"
)

type OcularHistoryRepo struct{ DB *gorm.DB }

func NewOcularHistoryRepo(db *gorm.DB) *OcularHistoryRepo {
	return &OcularHistoryRepo{DB: db}
}

func (r *OcularHistoryRepo) GetByID(id int64) (*history.OcularHistory, error) {
	var v history.OcularHistory
	if err := r.DB.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *OcularHistoryRepo) Create() (*history.OcularHistory, error) {
	v := history.OcularHistory{}
	if err := r.DB.Create(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *OcularHistoryRepo) Save(v *history.OcularHistory) error {
	return r.DB.Save(v).Error
}
