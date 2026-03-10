package patients_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/patients"
)

type PatientCommunicationHistoryRepo struct{ DB *gorm.DB }

func NewPatientCommunicationHistoryRepo(db *gorm.DB) *PatientCommunicationHistoryRepo {
	return &PatientCommunicationHistoryRepo{DB: db}
}

func (r *PatientCommunicationHistoryRepo) GetByID(id int64) (*patients.PatientCommunicationHistory, error) {
	var item patients.PatientCommunicationHistory
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *PatientCommunicationHistoryRepo) GetByPatientID(patientID int64) ([]patients.PatientCommunicationHistory, error) {
	var items []patients.PatientCommunicationHistory
	if err := r.DB.Where("patient_id = ?", patientID).
		Order("communication_datetime DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PatientCommunicationHistoryRepo) GetByDateRange(locationID int, from, to time.Time) ([]patients.PatientCommunicationHistory, error) {
	var items []patients.PatientCommunicationHistory
	if err := r.DB.Where("location_id = ? AND communication_datetime BETWEEN ? AND ?", locationID, from, to).
		Order("communication_datetime DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PatientCommunicationHistoryRepo) Create(item *patients.PatientCommunicationHistory) error {
	return r.DB.Create(item).Error
}

func (r *PatientCommunicationHistoryRepo) Delete(id int64) error {
	return r.DB.Delete(&patients.PatientCommunicationHistory{}, id).Error
}
