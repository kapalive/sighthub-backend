package assessment

import (
	"errors"

	"gorm.io/gorm"
	a "sighthub-backend/internal/models/medical/vision_exam/assessment"
)

type AssessmentDiagnosisRepo struct{ DB *gorm.DB }

func NewAssessmentDiagnosisRepo(db *gorm.DB) *AssessmentDiagnosisRepo {
	return &AssessmentDiagnosisRepo{DB: db}
}

func (r *AssessmentDiagnosisRepo) GetByID(id int64) (*a.AssessmentDiagnosis, error) {
	var v a.AssessmentDiagnosis
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *AssessmentDiagnosisRepo) Create(v *a.AssessmentDiagnosis) error {
	return r.DB.Create(v).Error
}

func (r *AssessmentDiagnosisRepo) Save(v *a.AssessmentDiagnosis) error {
	return r.DB.Save(v).Error
}

func (r *AssessmentDiagnosisRepo) Delete(id int64) error {
	return r.DB.Delete(&a.AssessmentDiagnosis{}, id).Error
}
