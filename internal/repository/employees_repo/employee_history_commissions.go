package employees_repo

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/employees"
)

type EmployeeHistoryCommissionsRepo struct{ DB *gorm.DB }

func NewEmployeeHistoryCommissionsRepo(db *gorm.DB) *EmployeeHistoryCommissionsRepo {
	return &EmployeeHistoryCommissionsRepo{DB: db}
}

func (r *EmployeeHistoryCommissionsRepo) GetByEmployeeID(employeeID int) ([]employees.EmployeeHistoryCommissions, error) {
	var items []employees.EmployeeHistoryCommissions
	return items, r.DB.Where("employee_id = ?", employeeID).Order("start_date DESC").Find(&items).Error
}

func (r *EmployeeHistoryCommissionsRepo) GetByCommissionsID(commissionsID int) ([]employees.EmployeeHistoryCommissions, error) {
	var items []employees.EmployeeHistoryCommissions
	return items, r.DB.Where("employee_commissions_id = ?", commissionsID).Order("start_date DESC").Find(&items).Error
}

func (r *EmployeeHistoryCommissionsRepo) GetByID(id int) (*employees.EmployeeHistoryCommissions, error) {
	var v employees.EmployeeHistoryCommissions
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *EmployeeHistoryCommissionsRepo) Create(v *employees.EmployeeHistoryCommissions) error {
	return r.DB.Create(v).Error
}

func (r *EmployeeHistoryCommissionsRepo) Save(v *employees.EmployeeHistoryCommissions) error {
	return r.DB.Save(v).Error
}
