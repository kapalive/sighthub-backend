package employees_repo

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/employees"
)

type EmployeeCommissionsDetailsRepo struct{ DB *gorm.DB }

func NewEmployeeCommissionsDetailsRepo(db *gorm.DB) *EmployeeCommissionsDetailsRepo {
	return &EmployeeCommissionsDetailsRepo{DB: db}
}

func (r *EmployeeCommissionsDetailsRepo) GetByID(id int) (*employees.EmployeeCommissionsDetails, error) {
	var v employees.EmployeeCommissionsDetails
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *EmployeeCommissionsDetailsRepo) GetAll() ([]employees.EmployeeCommissionsDetails, error) {
	var items []employees.EmployeeCommissionsDetails
	return items, r.DB.Find(&items).Error
}

func (r *EmployeeCommissionsDetailsRepo) Create(v *employees.EmployeeCommissionsDetails) error {
	return r.DB.Create(v).Error
}

func (r *EmployeeCommissionsDetailsRepo) Save(v *employees.EmployeeCommissionsDetails) error {
	return r.DB.Save(v).Error
}

func (r *EmployeeCommissionsDetailsRepo) Delete(id int) error {
	return r.DB.Delete(&employees.EmployeeCommissionsDetails{}, id).Error
}
