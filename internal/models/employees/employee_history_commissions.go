// internal/models/employees/employee_history_commissions.go
package employees

import (
	"fmt"
	"time"
)

// EmployeeHistoryCommissions ⇄ employee_history_commissions
type EmployeeHistoryCommissions struct {
	IDEmployeeHistoryCommissions int        `gorm:"column:id_employee_history_commissions;primaryKey"        json:"id_employee_history_commissions"`
	EmployeeCommissionsID        int        `gorm:"column:employee_commissions_id;not null;index"            json:"employee_commissions_id"`
	EmployeeID                   int        `gorm:"column:employee_id;not null;index"                        json:"employee_id"`
	StartDate                    time.Time  `gorm:"column:start_date;type:date;not null"                     json:"-"`
	EndDate                      *time.Time `gorm:"column:end_date;type:date"                                 json:"-"`
	CommissionPercent            float64    `gorm:"column:commission_percent;type:numeric(5,2);not null"     json:"commission_percent"`
	CreatedAt                    *time.Time `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"-"`

	// Optional relations
	Employee           *Employee            `gorm:"foreignKey:EmployeeID;references:IDEmployee"                       json:"-"`
	EmployeeCommission *EmployeeCommissions `gorm:"foreignKey:EmployeeCommissionsID;references:IDEmployeeCommissions" json:"-"`
}

func (EmployeeHistoryCommissions) TableName() string { return "employee_history_commissions" }

func (e *EmployeeHistoryCommissions) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_employee_history_commissions": e.IDEmployeeHistoryCommissions,
		"employee_commissions_id":         e.EmployeeCommissionsID,
		"employee_id":                     e.EmployeeID,
		"commission_percent":              e.CommissionPercent,
	}
	m["start_date"] = e.StartDate.Format("2006-01-02")
	if e.EndDate != nil && !e.EndDate.IsZero() {
		m["end_date"] = e.EndDate.Format("2006-01-02")
	} else {
		m["end_date"] = nil
	}
	if e.CreatedAt != nil && !e.CreatedAt.IsZero() {
		m["created_at"] = e.CreatedAt.Format(time.RFC3339)
	} else {
		m["created_at"] = nil
	}
	return m
}

func (e *EmployeeHistoryCommissions) String() string {
	return fmt.Sprintf("<EmployeeHistoryCommissions %d>", e.IDEmployeeHistoryCommissions)
}
