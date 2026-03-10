package patients_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/patients"
)

type PatientNotesRepo struct{ DB *gorm.DB }

func NewPatientNotesRepo(db *gorm.DB) *PatientNotesRepo { return &PatientNotesRepo{DB: db} }

func (r *PatientNotesRepo) GetByID(id int64) (*patients.PatientNotes, error) {
	var note patients.PatientNotes
	if err := r.DB.First(&note, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &note, nil
}

func (r *PatientNotesRepo) GetByPatientID(patientID int64) ([]patients.PatientNotes, error) {
	var notes []patients.PatientNotes
	if err := r.DB.Where("patient_id = ?", patientID).
		Order("top DESC, id_patient_notes DESC").Find(&notes).Error; err != nil {
		return nil, err
	}
	return notes, nil
}

func (r *PatientNotesRepo) Create(note *patients.PatientNotes) error {
	return r.DB.Create(note).Error
}

func (r *PatientNotesRepo) Save(note *patients.PatientNotes) error {
	return r.DB.Save(note).Error
}

func (r *PatientNotesRepo) Delete(id int64) error {
	return r.DB.Delete(&patients.PatientNotes{}, id).Error
}
