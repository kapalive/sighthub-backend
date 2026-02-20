package employees

import (
	"fmt"
	"time"
)

// EmployeeRole ⇄ employee_roles
type EmployeeRole struct {
	IDEmployeeRoles int        `gorm:"column:id_employee_roles;primaryKey;autoIncrement"                    json:"id_employee_roles"`
	IDEmployeeLogin int        `gorm:"column:id_employee_login;not null;uniqueIndex:uix_employee_role"      json:"id_employee_login"`
	RoleID          int        `gorm:"column:role_id;not null;uniqueIndex:uix_employee_role"                json:"role_id"`
	GrantedBy       *int       `gorm:"column:granted_by"                                                    json:"granted_by,omitempty"`
	GrantedAt       *time.Time `gorm:"column:granted_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"-"`
}

func (EmployeeRole) TableName() string { return "employee_roles" }

func (e *EmployeeRole) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_employee_roles": e.IDEmployeeRoles,
		"id_employee_login": e.IDEmployeeLogin,
		"role_id":           e.RoleID,
		"granted_by":        e.GrantedBy,
	}
	if e.GrantedAt != nil && !e.GrantedAt.IsZero() {
		m["granted_at"] = e.GrantedAt.Format(time.RFC3339)
	} else {
		m["granted_at"] = nil
	}
	return m
}

func (e *EmployeeRole) String() string {
	return fmt.Sprintf("<EmployeeRole EmployeeLogin=%d RoleID=%d>", e.IDEmployeeLogin, e.RoleID)
}
