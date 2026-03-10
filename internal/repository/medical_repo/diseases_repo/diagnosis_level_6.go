// internal/repository/medical_repo/diseases_repo/diagnosis_level_6.go
package diseases_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/diseases"
)

type DiagnosisLevel6Repo struct{ DB *gorm.DB }

func NewDiagnosisLevel6Repo(db *gorm.DB) *DiagnosisLevel6Repo {
	return &DiagnosisLevel6Repo{DB: db}
}

func (r *DiagnosisLevel6Repo) GetByDiagnosisID(diagnosisID int64) ([]diseases.DiagnosisLevel6, error) {
	var list []diseases.DiagnosisLevel6
	if err := r.DB.Where("diagnosis_id = ?", diagnosisID).
		Order("code").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *DiagnosisLevel6Repo) GetByID(id int64) (*diseases.DiagnosisLevel6, error) {
	var d diseases.DiagnosisLevel6
	if err := r.DB.First(&d, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *DiagnosisLevel6Repo) GetByCode(code string) (*diseases.DiagnosisLevel6, error) {
	var d diseases.DiagnosisLevel6
	if err := r.DB.Where("code = ?", code).First(&d).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}
