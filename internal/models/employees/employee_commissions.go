// internal/models/employees/employee_commissions.go
package employees

import (
	"fmt"
	"time"
)

// EmployeeCommissions ⇄ employee_commissions
type EmployeeCommissions struct {
	IDEmployeeCommissions int       `gorm:"column:id_employee_commissions;primaryKey"                  json:"id_employee_commissions"`
	EmployeeID            int       `gorm:"column:employee_id;not null;index"                          json:"employee_id"`
	StartDate             time.Time `gorm:"column:start_date;type:date;not null"                       json:"-"`
	EndDate               time.Time `gorm:"column:end_date;type:date;not null"                         json:"-"`
	CommissionPercent     float64   `gorm:"column:commission_percent;type:numeric(5,2);not null"       json:"commission_percent"`

	CreatedAt *time.Time `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP"   json:"-"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP"   json:"-"`

	// Optional relation (same package)
	Employee *Employee `gorm:"foreignKey:EmployeeID;references:IDEmployee" json:"-"`
}

func (EmployeeCommissions) TableName() string { return "employee_commissions" }

func (e *EmployeeCommissions) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_employee_commissions": e.IDEmployeeCommissions,
		"employee_id":             e.EmployeeID,
		"commission_percent":      e.CommissionPercent,
	}
	m["start_date"] = e.StartDate.Format("2006-01-02")
	m["end_date"] = e.EndDate.Format("2006-01-02")

	if e.CreatedAt != nil && !e.CreatedAt.IsZero() {
		m["created_at"] = e.CreatedAt.Format(time.RFC3339)
	} else {
		m["created_at"] = nil
	}
	if e.UpdatedAt != nil && !e.UpdatedAt.IsZero() {
		m["updated_at"] = e.UpdatedAt.Format(time.RFC3339)
	} else {
		m["updated_at"] = nil
	}
	return m
}

func (e *EmployeeCommissions) String() string {
	return fmt.Sprintf("<EmployeeCommissions %d - Employee %d>", e.IDEmployeeCommissions, e.EmployeeID)
}
