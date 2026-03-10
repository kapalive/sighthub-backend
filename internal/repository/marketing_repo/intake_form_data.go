package marketing_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/marketing"
)

type IntakeFormDataRepo struct{ DB *gorm.DB }

func NewIntakeFormDataRepo(db *gorm.DB) *IntakeFormDataRepo { return &IntakeFormDataRepo{DB: db} }

func (r *IntakeFormDataRepo) GetByID(id int64) (*marketing.IntakeFormData, error) {
	var form marketing.IntakeFormData
	if err := r.DB.
		Preload("MedicalHistory").
		Preload("Medications").
		Preload("Allergies").
		First(&form, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &form, nil
}

func (r *IntakeFormDataRepo) GetByAppointmentID(appointmentID int64) (*marketing.IntakeFormData, error) {
	var form marketing.IntakeFormData
	if err := r.DB.
		Preload("MedicalHistory").
		Preload("Medications").
		Preload("Allergies").
		Where("appointment_id = ?", appointmentID).First(&form).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &form, nil
}

func (r *IntakeFormDataRepo) Create(form *marketing.IntakeFormData) error {
	return r.DB.Create(form).Error
}

func (r *IntakeFormDataRepo) Save(form *marketing.IntakeFormData) error {
	return r.DB.Save(form).Error
}
