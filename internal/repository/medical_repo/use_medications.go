// internal/repository/medical_repo/use_medications.go
package medical_repo

import (
	"gorm.io/gorm"
	"sighthub-backend/internal/models/medical"
)

type UseMedicationsRepo struct{ DB *gorm.DB }

func NewUseMedicationsRepo(db *gorm.DB) *UseMedicationsRepo {
	return &UseMedicationsRepo{DB: db}
}

func (r *UseMedicationsRepo) GetByEyeExamID(eyeExamID int64) ([]medical.UseMedications, error) {
	var list []medical.UseMedications
	if err := r.DB.Where("eye_exam_id = ?", eyeExamID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

type CreateMedicationInput struct {
	Title           string
	FormulationType *string
	Strength        *string
	EyeExamID       int64
}

func (r *UseMedicationsRepo) Create(inp CreateMedicationInput) (*medical.UseMedications, error) {
	m := medical.UseMedications{
		Title:           inp.Title,
		FormulationType: inp.FormulationType,
		Strength:        inp.Strength,
		EyeExamID:       inp.EyeExamID,
	}
	if err := r.DB.Create(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *UseMedicationsRepo) Delete(id int64) error {
	return r.DB.Delete(&medical.UseMedications{}, id).Error
}

func (r *UseMedicationsRepo) DeleteByEyeExamID(eyeExamID int64) error {
	return r.DB.Where("eye_exam_id = ?", eyeExamID).Delete(&medical.UseMedications{}).Error
}
