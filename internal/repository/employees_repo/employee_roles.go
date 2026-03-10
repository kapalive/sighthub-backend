package employees_repo

import (
	"time"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/employees"
)

type EmployeeRolesRepo struct {
	DB *gorm.DB
}

func NewEmployeeRolesRepo(db *gorm.DB) *EmployeeRolesRepo {
	return &EmployeeRolesRepo{DB: db}
}

// GetForLogin returns all EmployeeRole records for the given login ID.
func (r *EmployeeRolesRepo) GetForLogin(loginID int) ([]employees.EmployeeRole, error) {
	var roles []employees.EmployeeRole
	if err := r.DB.Where("id_employee_login = ?", loginID).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// SetRoles replaces all roles for the given login with the supplied roleIDs.
// It deletes existing roles and inserts fresh ones within the provided transaction.
func (r *EmployeeRolesRepo) SetRoles(tx *gorm.DB, loginID, grantedBy int, roleIDs []int) error {
	// Delete existing roles for this login
	if err := tx.Where("id_employee_login = ?", loginID).Delete(&employees.EmployeeRole{}).Error; err != nil {
		return err
	}

	if len(roleIDs) == 0 {
		return nil
	}

	now := time.Now()
	grantedByPtr := &grantedBy
	for _, roleID := range roleIDs {
		role := employees.EmployeeRole{
			IDEmployeeLogin: loginID,
			RoleID:          roleID,
			GrantedBy:       grantedByPtr,
			GrantedAt:       &now,
		}
		if err := tx.Create(&role).Error; err != nil {
			return err
		}
	}
	return nil
}
