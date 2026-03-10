// internal/repository/medical_repo/known_allergies.go
package medical_repo

import (
	"gorm.io/gorm"
	"sighthub-backend/internal/models/medical"
)

type KnownAllergiesRepo struct{ DB *gorm.DB }

func NewKnownAllergiesRepo(db *gorm.DB) *KnownAllergiesRepo {
	return &KnownAllergiesRepo{DB: db}
}

func (r *KnownAllergiesRepo) GetByEyeExamID(eyeExamID int64) ([]medical.KnownAllergies, error) {
	var list []medical.KnownAllergies
	if err := r.DB.Where("eye_exam_id = ?", eyeExamID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *KnownAllergiesRepo) Create(title string, eyeExamID int64) (*medical.KnownAllergies, error) {
	ka := medical.KnownAllergies{
		Title:     title,
		EyeExamID: eyeExamID,
	}
	if err := r.DB.Create(&ka).Error; err != nil {
		return nil, err
	}
	return &ka, nil
}

func (r *KnownAllergiesRepo) Delete(id int64) error {
	return r.DB.Delete(&medical.KnownAllergies{}, id).Error
}

func (r *KnownAllergiesRepo) DeleteByEyeExamID(eyeExamID int64) error {
	return r.DB.Where("eye_exam_id = ?", eyeExamID).Delete(&medical.KnownAllergies{}).Error
}
