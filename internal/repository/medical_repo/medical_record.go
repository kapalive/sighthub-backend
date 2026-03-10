// internal/repository/medical_repo/medical_record.go
package medical_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/medical"
)

type MedicalRecordRepo struct{ DB *gorm.DB }

func NewMedicalRecordRepo(db *gorm.DB) *MedicalRecordRepo {
	return &MedicalRecordRepo{DB: db}
}

func (r *MedicalRecordRepo) GetByID(id int64) (*medical.MedicalRecord, error) {
	var m medical.MedicalRecord
	if err := r.DB.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *MedicalRecordRepo) Create(occupation, note *string) (*medical.MedicalRecord, error) {
	m := medical.MedicalRecord{
		Occupation:            occupation,
		PersistingPatientNote: note,
	}
	if err := r.DB.Create(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MedicalRecordRepo) Save(m *medical.MedicalRecord) error {
	return r.DB.Save(m).Error
}
