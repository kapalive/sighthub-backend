package patients_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/patients"
)

type InsuranceHolderPatientsRepo struct{ DB *gorm.DB }

func NewInsuranceHolderPatientsRepo(db *gorm.DB) *InsuranceHolderPatientsRepo {
	return &InsuranceHolderPatientsRepo{DB: db}
}

func (r *InsuranceHolderPatientsRepo) GetByID(id int) (*patients.InsuranceHolderPatients, error) {
	var item patients.InsuranceHolderPatients
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *InsuranceHolderPatientsRepo) GetByPatientID(patientID int64) ([]patients.InsuranceHolderPatients, error) {
	var items []patients.InsuranceHolderPatients
	if err := r.DB.Where("patient_id = ?", patientID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *InsuranceHolderPatientsRepo) GetByPolicyID(policyID int64) ([]patients.InsuranceHolderPatients, error) {
	var items []patients.InsuranceHolderPatients
	if err := r.DB.Where("insurance_policy_id = ?", policyID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *InsuranceHolderPatientsRepo) Create(item *patients.InsuranceHolderPatients) error {
	return r.DB.Create(item).Error
}

func (r *InsuranceHolderPatientsRepo) Save(item *patients.InsuranceHolderPatients) error {
	return r.DB.Save(item).Error
}

func (r *InsuranceHolderPatientsRepo) Delete(id int) error {
	return r.DB.Delete(&patients.InsuranceHolderPatients{}, id).Error
}
