package employees

import (
	"fmt"
	"time"
)

// EmployeePermission ⇄ employee_permission
type EmployeePermission struct {
	IDEmployeePermission     int        `gorm:"column:id_employee_permission;primaryKey;autoIncrement"                        json:"id_employee_permission"`
	EmployeeLoginID          int        `gorm:"column:employee_login_id;not null;index;uniqueIndex:unique_employee_permission" json:"employee_login_id"`
	PermissionsCombinationID int        `gorm:"column:permissions_combination_id;not null;index;uniqueIndex:unique_employee_permission" json:"permissions_combination_id"`
	GrantedBy                int        `gorm:"column:granted_by;not null;index"                                              json:"granted_by"`
	GrantedAt                *time.Time `gorm:"column:granted_at;type:timestamp;not null;default:CURRENT_TIMESTAMP"           json:"-"`
	IsActive                 bool       `gorm:"column:is_active;not null;default:true"                                        json:"is_active"`
}

func (EmployeePermission) TableName() string { return "employee_permission" }

// ToMap — аналог Python to_dict()
func (e *EmployeePermission) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_employee_permission":     e.IDEmployeePermission,
		"employee_login_id":          e.EmployeeLoginID,
		"permissions_combination_id": e.PermissionsCombinationID,
		"granted_by":                 e.GrantedBy,
		"is_active":                  e.IsActive,
	}
	if e.GrantedAt != nil && !e.GrantedAt.IsZero() {
		m["granted_at"] = e.GrantedAt.Format(time.RFC3339)
	} else {
		m["granted_at"] = nil
	}
	return m
}

func (e *EmployeePermission) String() string {
	return fmt.Sprintf("<EmployeePermission perm=%d for employee_login=%d>", e.PermissionsCombinationID, e.EmployeeLoginID)
}
