package assessment

import (
	"errors"

	"gorm.io/gorm"
	a "sighthub-backend/internal/models/medical/vision_exam/assessment"
)

type AssessmentEyeRepo struct{ DB *gorm.DB }

func NewAssessmentEyeRepo(db *gorm.DB) *AssessmentEyeRepo {
	return &AssessmentEyeRepo{DB: db}
}

func (r *AssessmentEyeRepo) GetByEyeExamID(eyeExamID int64) (*a.AssessmentEye, error) {
	var v a.AssessmentEye
	if err := r.DB.
		Preload("Diagnoses").
		Preload("PQRSItems").
		Where("eye_exam_id = ?", eyeExamID).
		First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *AssessmentEyeRepo) GetByID(id int64) (*a.AssessmentEye, error) {
	var v a.AssessmentEye
	if err := r.DB.
		Preload("Diagnoses").
		Preload("PQRSItems").
		First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *AssessmentEyeRepo) Create(eyeExamID int64) (*a.AssessmentEye, error) {
	v := a.AssessmentEye{EyeExamID: eyeExamID}
	if err := r.DB.Create(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *AssessmentEyeRepo) Save(v *a.AssessmentEye) error {
	return r.DB.Save(v).Error
}
