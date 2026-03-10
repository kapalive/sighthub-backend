package assessment

import (
	"errors"

	"gorm.io/gorm"
	a "sighthub-backend/internal/models/medical/vision_exam/assessment"
)

type AssessmentPQRSRepo struct{ DB *gorm.DB }

func NewAssessmentPQRSRepo(db *gorm.DB) *AssessmentPQRSRepo {
	return &AssessmentPQRSRepo{DB: db}
}

func (r *AssessmentPQRSRepo) GetByID(id int64) (*a.AssessmentPQRS, error) {
	var v a.AssessmentPQRS
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *AssessmentPQRSRepo) Create(v *a.AssessmentPQRS) error {
	return r.DB.Create(v).Error
}

func (r *AssessmentPQRSRepo) Save(v *a.AssessmentPQRS) error {
	return r.DB.Save(v).Error
}

func (r *AssessmentPQRSRepo) Delete(id int64) error {
	return r.DB.Delete(&a.AssessmentPQRS{}, id).Error
}
