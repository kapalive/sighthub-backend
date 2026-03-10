package insurance_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/insurance"
)

type InsurancePaymentTypeRepo struct{ DB *gorm.DB }

func NewInsurancePaymentTypeRepo(db *gorm.DB) *InsurancePaymentTypeRepo {
	return &InsurancePaymentTypeRepo{DB: db}
}

func (r *InsurancePaymentTypeRepo) GetAll() ([]insurance.InsurancePaymentType, error) {
	var items []insurance.InsurancePaymentType
	return items, r.DB.Where("active = true").Order("name").Find(&items).Error
}

func (r *InsurancePaymentTypeRepo) GetByID(id int) (*insurance.InsurancePaymentType, error) {
	var v insurance.InsurancePaymentType
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *InsurancePaymentTypeRepo) Create(v *insurance.InsurancePaymentType) error {
	return r.DB.Create(v).Error
}
func (r *InsurancePaymentTypeRepo) Save(v *insurance.InsurancePaymentType) error {
	return r.DB.Save(v).Error
}
func (r *InsurancePaymentTypeRepo) Delete(id int) error {
	return r.DB.Delete(&insurance.InsurancePaymentType{}, id).Error
}
