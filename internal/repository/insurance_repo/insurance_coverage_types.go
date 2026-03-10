package insurance_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/insurance"
)

type InsuranceCoverageTypeRepo struct{ DB *gorm.DB }

func NewInsuranceCoverageTypeRepo(db *gorm.DB) *InsuranceCoverageTypeRepo {
	return &InsuranceCoverageTypeRepo{DB: db}
}

func (r *InsuranceCoverageTypeRepo) GetAll() ([]insurance.InsuranceCoverageType, error) {
	var items []insurance.InsuranceCoverageType
	return items, r.DB.Order("coverage_name").Find(&items).Error
}

func (r *InsuranceCoverageTypeRepo) GetByID(id int) (*insurance.InsuranceCoverageType, error) {
	var v insurance.InsuranceCoverageType
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *InsuranceCoverageTypeRepo) Create(v *insurance.InsuranceCoverageType) error {
	return r.DB.Create(v).Error
}
func (r *InsuranceCoverageTypeRepo) Save(v *insurance.InsuranceCoverageType) error {
	return r.DB.Save(v).Error
}
func (r *InsuranceCoverageTypeRepo) Delete(id int) error {
	return r.DB.Delete(&insurance.InsuranceCoverageType{}, id).Error
}
