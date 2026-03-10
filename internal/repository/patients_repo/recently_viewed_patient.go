package patients_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/patients"
)

type RecentlyViewedPatientRepo struct{ DB *gorm.DB }

func NewRecentlyViewedPatientRepo(db *gorm.DB) *RecentlyViewedPatientRepo {
	return &RecentlyViewedPatientRepo{DB: db}
}

func (r *RecentlyViewedPatientRepo) GetByLocationID(locationID int) ([]patients.RecentlyViewedPatient, error) {
	var items []patients.RecentlyViewedPatient
	if err := r.DB.Where("location_id = ?", locationID).
		Order("datetime_viewed DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *RecentlyViewedPatientRepo) Upsert(locationID int, patientID int64) error {
	return r.DB.Exec(
		`INSERT INTO recently_viewed_patient (location_id, patient_id, datetime_viewed)
		 VALUES (?, ?, NOW())
		 ON CONFLICT (location_id, patient_id) DO UPDATE SET datetime_viewed = NOW()`,
		locationID, patientID,
	).Error
}

func (r *RecentlyViewedPatientRepo) Delete(id int64) error {
	return r.DB.Delete(&patients.RecentlyViewedPatient{}, id).Error
}

func (r *RecentlyViewedPatientRepo) GetByID(id int64) (*patients.RecentlyViewedPatient, error) {
	var item patients.RecentlyViewedPatient
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}
