package schedule

import (
	"sighthub-backend/internal/models/employees"
)

type Schedule struct {
	IDSchedule int    `gorm:"column:id_schedule;primaryKey"             json:"id_schedule"`
	EmployeeID int64  `gorm:"column:employee_id;not null"               json:"employee_id"`
	DayOfWeek  string `gorm:"column:day_of_week;type:varchar(10);not null" json:"day_of_week"`

	StartTime           string `gorm:"column:start_time;type:time;not null"   json:"-"`
	EndTime             string `gorm:"column:end_time;type:time;not null"     json:"-"`
	AppointmentDuration int    `gorm:"column:appointment_duration;default:15" json:"appointment_duration"`

	LunchStart *string `gorm:"column:lunch_start;type:time"           json:"-"`
	LunchEnd   *string `gorm:"column:lunch_end;type:time"             json:"-"`

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

	// Формат времени — уже строки из PostgreSQL time
	if s.StartTime != "" {
		m["start_time"] = s.StartTime
	} else {
		m["start_time"] = nil
	}
	if s.EndTime != "" {
		m["end_time"] = s.EndTime
	} else {
		m["end_time"] = nil
	}

	if s.LunchStart != nil && *s.LunchStart != "" {
		m["lunch_start"] = *s.LunchStart
	} else {
		m["lunch_start"] = nil
	}
	if s.LunchEnd != nil && *s.LunchEnd != "" {
		m["lunch_end"] = *s.LunchEnd
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
