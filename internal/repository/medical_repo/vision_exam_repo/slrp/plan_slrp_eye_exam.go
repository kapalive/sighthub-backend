package slrp

import (
	"errors"

	"gorm.io/gorm"
	s "sighthub-backend/internal/models/medical/vision_exam/slrp"
)

type PlanSLRPEyeExamRepo struct{ DB *gorm.DB }

func NewPlanSLRPEyeExamRepo(db *gorm.DB) *PlanSLRPEyeExamRepo {
	return &PlanSLRPEyeExamRepo{DB: db}
}

func (r *PlanSLRPEyeExamRepo) GetByID(id int64) (*s.PlanSLRPEyeExam, error) {
	var v s.PlanSLRPEyeExam
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *PlanSLRPEyeExamRepo) Create(v *s.PlanSLRPEyeExam) error {
	return r.DB.Create(v).Error
}

func (r *PlanSLRPEyeExamRepo) Save(v *s.PlanSLRPEyeExam) error {
	return r.DB.Save(v).Error
}
