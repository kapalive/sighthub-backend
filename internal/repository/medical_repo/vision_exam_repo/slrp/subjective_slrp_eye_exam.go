package slrp

import (
	"errors"

	"gorm.io/gorm"
	s "sighthub-backend/internal/models/medical/vision_exam/slrp"
)

type SubjectiveSLRPEyeExamRepo struct{ DB *gorm.DB }

func NewSubjectiveSLRPEyeExamRepo(db *gorm.DB) *SubjectiveSLRPEyeExamRepo {
	return &SubjectiveSLRPEyeExamRepo{DB: db}
}

func (r *SubjectiveSLRPEyeExamRepo) GetByID(id int64) (*s.SubjectiveSLRPEyeExam, error) {
	var v s.SubjectiveSLRPEyeExam
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *SubjectiveSLRPEyeExamRepo) Create(v *s.SubjectiveSLRPEyeExam) error {
	return r.DB.Create(v).Error
}

func (r *SubjectiveSLRPEyeExamRepo) Save(v *s.SubjectiveSLRPEyeExam) error {
	return r.DB.Save(v).Error
}
