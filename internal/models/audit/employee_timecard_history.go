// internal/models/audit/employee_timecard_history.go
package audit

import "time"

// EmployeeTimecardHistory ⇄ table: employee_timecard_history
// Хранит checkin/checkout записи для timecard-пользователей.
type EmployeeTimecardHistory struct {
	IDEmployeeTimecardHistory int       `gorm:"column:id_employee_timecard_history;primaryKey;autoIncrement" json:"id_employee_timecard_history"`
	EmployeeTimecardLoginID   int       `gorm:"column:employee_timecard_login_id;not null;index"             json:"employee_timecard_login_id"`
	ActionType                string    `gorm:"column:action_type;type:varchar(10);not null"                 json:"action_type"` // "checkin" | "checkout"
	Timestamp                 time.Time `gorm:"column:timestamp;type:timestamptz;not null;default:now()"     json:"timestamp"`
	Note                      *string   `gorm:"column:note;type:text"                                        json:"note,omitempty"`
}

func (EmployeeTimecardHistory) TableName() string { return "employee_timecard_history" }

func (e *EmployeeTimecardHistory) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_employee_timecard_history": e.IDEmployeeTimecardHistory,
		"employee_timecard_login_id":   e.EmployeeTimecardLoginID,
		"action_type":                  e.ActionType,
		"timestamp":                    e.Timestamp.Format(time.RFC3339),
		"note":                         e.Note,
	}
	return m
}
