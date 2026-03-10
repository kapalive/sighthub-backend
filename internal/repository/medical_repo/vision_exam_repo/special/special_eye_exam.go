package special

import (
	"errors"

	"gorm.io/gorm"
	sp "sighthub-backend/internal/models/medical/vision_exam/special"
)

type SpecialEyeExamRepo struct{ DB *gorm.DB }

func NewSpecialEyeExamRepo(db *gorm.DB) *SpecialEyeExamRepo {
	return &SpecialEyeExamRepo{DB: db}
}

func (r *SpecialEyeExamRepo) GetByEyeExamID(eyeExamID int64) (*sp.SpecialEyeExam, error) {
	var v sp.SpecialEyeExam
	if err := r.DB.
		Preload("Files").
		Where("eye_exam_id = ?", eyeExamID).
		First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *SpecialEyeExamRepo) GetByID(id int64) (*sp.SpecialEyeExam, error) {
	var v sp.SpecialEyeExam
	if err := r.DB.
		Preload("Files").
		First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *SpecialEyeExamRepo) Create(eyeExamID int64) (*sp.SpecialEyeExam, error) {
	v := sp.SpecialEyeExam{EyeExamID: eyeExamID}
	if err := r.DB.Create(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *SpecialEyeExamRepo) Save(v *sp.SpecialEyeExam) error {
	return r.DB.Save(v).Error
}
