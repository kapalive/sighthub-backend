package schedule

import (
	"sighthub-backend/internal/models/employees"
	"time"
)

type Schedule struct {
	IDSchedule int    `gorm:"column:id_schedule;primaryKey"             json:"id_schedule"`
	EmployeeID int64  `gorm:"column:employee_id;not null"               json:"employee_id"`
	DayOfWeek  string `gorm:"column:day_of_week;type:varchar(10);not null" json:"day_of_week"`

	StartTime           time.Time `gorm:"column:start_time;type:time;not null"   json:"-"`
	EndTime             time.Time `gorm:"column:end_time;type:time;not null"     json:"-"`
	AppointmentDuration int       `gorm:"column:appointment_duration;default:15" json:"appointment_duration"`

	LunchStart *time.Time `gorm:"column:lunch_start;type:time"           json:"-"`
	LunchEnd   *time.Time `gorm:"column:lunch_end;type:time"             json:"-"`

	// Relation
	Employee *employees.Employee `gorm:"foreignKey:EmployeeID;references:IDEmployee" json:"employee,omitempty"`
}

func (Schedule) TableName() string { return "schedule" }

// ToMap — эквивалент Python to_dict()
func (s *Schedule) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_schedule":          s.IDSchedule,
		"employee_id":          s.EmployeeID,
		"day_of_week":          s.DayOfWeek,
		"appointment_duration": s.AppointmentDuration,
	}

	// Формат времени как "%H:%M:%S"
	m["start_time"] = s.StartTime.Format("15:04:05")
	m["end_time"] = s.EndTime.Format("15:04:05")

	if s.LunchStart != nil && !s.LunchStart.IsZero() {
		m["lunch_start"] = s.LunchStart.Format("15:04:05")
	} else {
		m["lunch_start"] = nil
	}
	if s.LunchEnd != nil && !s.LunchEnd.IsZero() {
		m["lunch_end"] = s.LunchEnd.Format("15:04:05")
	} else {
		m["lunch_end"] = nil
	}

	// Вложенный employee
	if s.Employee != nil {
		// Если у тебя собственный тип с ToMap — поддержим оба варианта.
		if tm, ok := any(s.Employee).(interface{ ToMap() map[string]interface{} }); ok {
			m["employee"] = tm.ToMap()
		} else {
			m["employee"] = s.Employee
		}
	} else {
		m["employee"] = nil
	}

	return m
}
