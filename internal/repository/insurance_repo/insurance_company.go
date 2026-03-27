package insurance_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/insurance"
)

type InsuranceCompanyRepo struct{ DB *gorm.DB }

func NewInsuranceCompanyRepo(db *gorm.DB) *InsuranceCompanyRepo {
	return &InsuranceCompanyRepo{DB: db}
}

func (r *InsuranceCompanyRepo) GetAll() ([]insurance.InsuranceCompany, error) {
	var items []insurance.InsuranceCompany
	return items, r.DB.Order("company_name").Find(&items).Error
}

func (r *InsuranceCompanyRepo) GetByID(id int) (*insurance.InsuranceCompany, error) {
	var v insurance.InsuranceCompany
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *InsuranceCompanyRepo) Search(query string) ([]insurance.InsuranceCompany, error) {
	var items []insurance.InsuranceCompany
	return items, r.DB.
		Where("company_name ILIKE ?", "%"+query+"%").
		Order("company_name").
		Find(&items).Error
}

func (r *InsuranceCompanyRepo) Create(v *insurance.InsuranceCompany) error { return r.DB.Create(v).Error }
func (r *InsuranceCompanyRepo) Save(v *insurance.InsuranceCompany) error   { return r.DB.Save(v).Error }
func (r *InsuranceCompanyRepo) Delete(id int) error {
	return r.DB.Delete(&insurance.InsuranceCompany{}, id).Error
}
