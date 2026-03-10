// internal/repository/medical_repo/diseases_repo/diagnosis.go
package diseases_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/diseases"
)

type DiagnosisRepo struct{ DB *gorm.DB }

func NewDiagnosisRepo(db *gorm.DB) *DiagnosisRepo {
	return &DiagnosisRepo{DB: db}
}

func (r *DiagnosisRepo) GetByLevel4ID(level4ID int64) ([]diseases.Diagnosis, error) {
	var list []diseases.Diagnosis
	if err := r.DB.Where("level_4_id = ?", level4ID).
		Order("code").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *DiagnosisRepo) GetByID(id int64) (*diseases.Diagnosis, error) {
	var d diseases.Diagnosis
	if err := r.DB.First(&d, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *DiagnosisRepo) GetByCode(code string) (*diseases.Diagnosis, error) {
	var d diseases.Diagnosis
	if err := r.DB.Where("code = ?", code).First(&d).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *DiagnosisRepo) Search(q string) ([]diseases.Diagnosis, error) {
	var list []diseases.Diagnosis
	if err := r.DB.Where("full_name ILIKE ? OR code ILIKE ?", "%"+q+"%", q+"%").
		Order("code").Limit(30).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
