// internal/repository/medical_repo/vision_exam_repo/history/ros_medical_history.go
package history

import (
	"gorm.io/gorm"
	"sighthub-backend/internal/models/medical/vision_exam/history"
)

type ROSMedicalHistoryRepo struct{ DB *gorm.DB }

func NewROSMedicalHistoryRepo(db *gorm.DB) *ROSMedicalHistoryRepo {
	return &ROSMedicalHistoryRepo{DB: db}
}

func (r *ROSMedicalHistoryRepo) GetByID(id int64) (*history.ROSMedicalHistory, error) {
	var v history.ROSMedicalHistory
	if err := r.DB.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *ROSMedicalHistoryRepo) Create() (*history.ROSMedicalHistory, error) {
	v := history.ROSMedicalHistory{}
	if err := r.DB.Create(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *ROSMedicalHistoryRepo) Save(v *history.ROSMedicalHistory) error {
	return r.DB.Save(v).Error
}
