package patients_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/patients"
)

type PatientRepo struct{ DB *gorm.DB }

func NewPatientRepo(db *gorm.DB) *PatientRepo { return &PatientRepo{DB: db} }

func (r *PatientRepo) GetByID(id int64) (*patients.Patient, error) {
	var p patients.Patient
	if err := r.DB.First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *PatientRepo) GetByLocationID(locationID int64) ([]patients.Patient, error) {
	var ps []patients.Patient
	if err := r.DB.Where("location_id = ?", locationID).Find(&ps).Error; err != nil {
		return nil, err
	}
	return ps, nil
}

func (r *PatientRepo) Search(query string) ([]patients.Patient, error) {
	var ps []patients.Patient
	q := "%" + query + "%"
	err := r.DB.Where(
		"first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ? OR phone ILIKE ?",
		q, q, q, q,
	).Find(&ps).Error
	return ps, err
}

func (r *PatientRepo) Create(p *patients.Patient) error {
	return r.DB.Create(p).Error
}

func (r *PatientRepo) Save(p *patients.Patient) error {
	return r.DB.Save(p).Error
}

func (r *PatientRepo) Delete(id int64) error {
	return r.DB.Delete(&patients.Patient{}, id).Error
}
