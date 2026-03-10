package patients_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/patients"
)

type DocumentsPatientRepo struct{ DB *gorm.DB }

func NewDocumentsPatientRepo(db *gorm.DB) *DocumentsPatientRepo {
	return &DocumentsPatientRepo{DB: db}
}

func (r *DocumentsPatientRepo) GetByID(id int64) (*patients.DocumentsPatient, error) {
	var doc patients.DocumentsPatient
	if err := r.DB.First(&doc, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &doc, nil
}

func (r *DocumentsPatientRepo) GetByPatientID(patientID int64) ([]patients.DocumentsPatient, error) {
	var docs []patients.DocumentsPatient
	if err := r.DB.Where("patient_id = ? AND is_hidden = false", patientID).
		Order("created_time DESC").Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *DocumentsPatientRepo) Create(doc *patients.DocumentsPatient) error {
	return r.DB.Create(doc).Error
}

func (r *DocumentsPatientRepo) Save(doc *patients.DocumentsPatient) error {
	return r.DB.Save(doc).Error
}

func (r *DocumentsPatientRepo) Delete(id int64) error {
	return r.DB.Delete(&patients.DocumentsPatient{}, id).Error
}

func (r *DocumentsPatientRepo) SetHidden(id int64, hidden bool) error {
	return r.DB.Model(&patients.DocumentsPatient{}).
		Where("id_documents_patient = ?", id).
		Update("is_hidden", hidden).Error
}
