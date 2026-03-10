package employees_repo

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/employees"
)

type EmployeeHistoryCommissionsDetailsRelationRepo struct{ DB *gorm.DB }

func NewEmployeeHistoryCommissionsDetailsRelationRepo(db *gorm.DB) *EmployeeHistoryCommissionsDetailsRelationRepo {
	return &EmployeeHistoryCommissionsDetailsRelationRepo{DB: db}
}

func (r *EmployeeHistoryCommissionsDetailsRelationRepo) GetByHistoryCommissionsID(historyID int) ([]employees.EmployeeHistoryCommissionsDetailsRelation, error) {
	var items []employees.EmployeeHistoryCommissionsDetailsRelation
	return items, r.DB.Where("history_commissions_id = ?", historyID).Find(&items).Error
}

func (r *EmployeeHistoryCommissionsDetailsRelationRepo) GetByID(id int) (*employees.EmployeeHistoryCommissionsDetailsRelation, error) {
	var v employees.EmployeeHistoryCommissionsDetailsRelation
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *EmployeeHistoryCommissionsDetailsRelationRepo) Create(v *employees.EmployeeHistoryCommissionsDetailsRelation) error {
	return r.DB.Create(v).Error
}

func (r *EmployeeHistoryCommissionsDetailsRelationRepo) Delete(id int) error {
	return r.DB.Delete(&employees.EmployeeHistoryCommissionsDetailsRelation{}, id).Error
}

func (r *EmployeeHistoryCommissionsDetailsRelationRepo) DeleteByHistoryCommissionsID(historyID int) error {
	return r.DB.Where("history_commissions_id = ?", historyID).
		Delete(&employees.EmployeeHistoryCommissionsDetailsRelation{}).Error
}
