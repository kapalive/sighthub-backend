package insurance_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/insurance"
)

type InsurancePolicyRepo struct{ DB *gorm.DB }

func NewInsurancePolicyRepo(db *gorm.DB) *InsurancePolicyRepo {
	return &InsurancePolicyRepo{DB: db}
}

func (r *InsurancePolicyRepo) GetByID(id int64) (*insurance.InsurancePolicy, error) {
	var v insurance.InsurancePolicy
	if err := r.DB.
		Preload("InsuranceCompany").
		Preload("InsuranceCoverageType").
		First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *InsurancePolicyRepo) GetByCompanyID(companyID int) ([]insurance.InsurancePolicy, error) {
	var items []insurance.InsurancePolicy
	return items, r.DB.Where("insurance_company_id = ?", companyID).Find(&items).Error
}

func (r *InsurancePolicyRepo) Create(v *insurance.InsurancePolicy) error { return r.DB.Create(v).Error }
func (r *InsurancePolicyRepo) Save(v *insurance.InsurancePolicy) error   { return r.DB.Save(v).Error }
func (r *InsurancePolicyRepo) Delete(id int64) error {
	return r.DB.Delete(&insurance.InsurancePolicy{}, id).Error
}
