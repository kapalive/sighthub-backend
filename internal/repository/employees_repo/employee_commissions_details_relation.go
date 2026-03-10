package employees_repo

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/employees"
)

type EmployeeCommissionsDetailsRelationRepo struct{ DB *gorm.DB }

func NewEmployeeCommissionsDetailsRelationRepo(db *gorm.DB) *EmployeeCommissionsDetailsRelationRepo {
	return &EmployeeCommissionsDetailsRelationRepo{DB: db}
}

func (r *EmployeeCommissionsDetailsRelationRepo) GetByCommissionsID(commissionsID int) ([]employees.EmployeeCommissionsDetailsRelation, error) {
	var items []employees.EmployeeCommissionsDetailsRelation
	return items, r.DB.Where("employee_commissions_id = ?", commissionsID).Find(&items).Error
}

func (r *EmployeeCommissionsDetailsRelationRepo) GetByID(id int) (*employees.EmployeeCommissionsDetailsRelation, error) {
	var v employees.EmployeeCommissionsDetailsRelation
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *EmployeeCommissionsDetailsRelationRepo) Create(v *employees.EmployeeCommissionsDetailsRelation) error {
	return r.DB.Create(v).Error
}

func (r *EmployeeCommissionsDetailsRelationRepo) Delete(id int) error {
	return r.DB.Delete(&employees.EmployeeCommissionsDetailsRelation{}, id).Error
}

func (r *EmployeeCommissionsDetailsRelationRepo) DeleteByCommissionsID(commissionsID int) error {
	return r.DB.Where("employee_commissions_id = ?", commissionsID).
		Delete(&employees.EmployeeCommissionsDetailsRelation{}).Error
}
