// internal/repository/medical_repo/vision_exam_repo/eye_exam_type.go
package vision_exam_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vision_exam"
)

type EyeExamTypeRepo struct{ DB *gorm.DB }

func NewEyeExamTypeRepo(db *gorm.DB) *EyeExamTypeRepo {
	return &EyeExamTypeRepo{DB: db}
}

func (r *EyeExamTypeRepo) GetAll() ([]vision_exam.EyeExamType, error) {
	var list []vision_exam.EyeExamType
	if err := r.DB.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *EyeExamTypeRepo) GetByID(id int64) (*vision_exam.EyeExamType, error) {
	var t vision_exam.EyeExamType
	if err := r.DB.First(&t, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}
