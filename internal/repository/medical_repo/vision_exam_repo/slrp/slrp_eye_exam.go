package slrp

import (
	"errors"

	"gorm.io/gorm"
	s "sighthub-backend/internal/models/medical/vision_exam/slrp"
)

type SLRPEyeExamRepo struct{ DB *gorm.DB }

func NewSLRPEyeExamRepo(db *gorm.DB) *SLRPEyeExamRepo {
	return &SLRPEyeExamRepo{DB: db}
}

func (r *SLRPEyeExamRepo) GetByEyeExamID(eyeExamID int64) (*s.SLRPEyeExam, error) {
	var v s.SLRPEyeExam
	if err := r.DB.
		Preload("Subjective").
		Preload("Objective").
		Preload("Assessment").
		Preload("Plan").
		Where("eye_exam_id = ?", eyeExamID).
		First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *SLRPEyeExamRepo) GetByID(id int64) (*s.SLRPEyeExam, error) {
	var v s.SLRPEyeExam
	if err := r.DB.
		Preload("Subjective").
		Preload("Objective").
		Preload("Assessment").
		Preload("Plan").
		First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *SLRPEyeExamRepo) Create(eyeExamID int64) (*s.SLRPEyeExam, error) {
	v := s.SLRPEyeExam{EyeExamID: eyeExamID}
	if err := r.DB.Create(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *SLRPEyeExamRepo) Save(v *s.SLRPEyeExam) error {
	return r.DB.Save(v).Error
}
