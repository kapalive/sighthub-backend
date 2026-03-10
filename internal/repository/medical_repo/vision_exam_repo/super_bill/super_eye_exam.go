package super_bill

import (
	"errors"

	"gorm.io/gorm"
	sb "sighthub-backend/internal/models/medical/vision_exam/super_bill"
)

type SuperEyeExamRepo struct{ DB *gorm.DB }

func NewSuperEyeExamRepo(db *gorm.DB) *SuperEyeExamRepo {
	return &SuperEyeExamRepo{DB: db}
}

func (r *SuperEyeExamRepo) GetByEyeExamID(eyeExamID int64) (*sb.SuperEyeExam, error) {
	var v sb.SuperEyeExam
	if err := r.DB.
		Preload("Diagnoses").
		Where("eye_exam_id = ?", eyeExamID).
		First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *SuperEyeExamRepo) GetByID(id int64) (*sb.SuperEyeExam, error) {
	var v sb.SuperEyeExam
	if err := r.DB.
		Preload("Diagnoses").
		First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *SuperEyeExamRepo) Create(eyeExamID int64) (*sb.SuperEyeExam, error) {
	v := sb.SuperEyeExam{EyeExamID: eyeExamID}
	if err := r.DB.Create(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *SuperEyeExamRepo) Save(v *sb.SuperEyeExam) error {
	return r.DB.Save(v).Error
}
