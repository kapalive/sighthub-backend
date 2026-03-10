package slrp

import (
	"errors"

	"gorm.io/gorm"
	s "sighthub-backend/internal/models/medical/vision_exam/slrp"
)

type AssessmentSLRPEyeExamRepo struct{ DB *gorm.DB }

func NewAssessmentSLRPEyeExamRepo(db *gorm.DB) *AssessmentSLRPEyeExamRepo {
	return &AssessmentSLRPEyeExamRepo{DB: db}
}

func (r *AssessmentSLRPEyeExamRepo) GetByID(id int64) (*s.AssessmentSLRPEyeExam, error) {
	var v s.AssessmentSLRPEyeExam
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *AssessmentSLRPEyeExamRepo) Create(v *s.AssessmentSLRPEyeExam) error {
	return r.DB.Create(v).Error
}

func (r *AssessmentSLRPEyeExamRepo) Save(v *s.AssessmentSLRPEyeExam) error {
	return r.DB.Save(v).Error
}
