package slrp

import (
	"errors"

	"gorm.io/gorm"
	s "sighthub-backend/internal/models/medical/vision_exam/slrp"
)

type ObjectiveSLRPEyeExamRepo struct{ DB *gorm.DB }

func NewObjectiveSLRPEyeExamRepo(db *gorm.DB) *ObjectiveSLRPEyeExamRepo {
	return &ObjectiveSLRPEyeExamRepo{DB: db}
}

func (r *ObjectiveSLRPEyeExamRepo) GetByID(id int64) (*s.ObjectiveSLRPEyeExam, error) {
	var v s.ObjectiveSLRPEyeExam
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *ObjectiveSLRPEyeExamRepo) Create(v *s.ObjectiveSLRPEyeExam) error {
	return r.DB.Create(v).Error
}

func (r *ObjectiveSLRPEyeExamRepo) Save(v *s.ObjectiveSLRPEyeExam) error {
	return r.DB.Save(v).Error
}
