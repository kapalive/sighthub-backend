// internal/models/schedule/calendar.go
package schedule

import "time"

// Calendar ⇄ table: calendar
// Хранит исключения из расписания: выходные, праздники, переработки.
type Calendar struct {
	IDCalendar   int64      `gorm:"column:id_calendar;primaryKey;autoIncrement" json:"id_calendar"`
	Date         time.Time  `gorm:"column:date;type:date;not null"              json:"date"`
	IsHoliday    bool       `gorm:"column:is_holiday;not null;default:false"    json:"is_holiday"`
	WorkShiftID  *int64     `gorm:"column:work_shift_id"                        json:"work_shift_id,omitempty"`
	EmployeeID   *int64     `gorm:"column:employee_id"                          json:"employee_id,omitempty"`
	TimeStart    *time.Time `gorm:"column:time_start;type:time"                 json:"-"`
	TimeEnd      *time.Time `gorm:"column:time_end;type:time"                   json:"-"`
	IsWorkingDay bool       `gorm:"column:is_working_day;not null;default:true" json:"is_working_day"`
}

func (Calendar) TableName() string { return "calendar" }

func (c *Calendar) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_calendar":    c.IDCalendar,
		"date":           c.Date.Format("2006-01-02"),
		"is_holiday":     c.IsHoliday,
		"work_shift_id":  c.WorkShiftID,
		"employee_id":    c.EmployeeID,
		"is_working_day": c.IsWorkingDay,
	}
	if c.TimeStart != nil && !c.TimeStart.IsZero() {
		m["time_start"] = c.TimeStart.Format("15:04:05")
	} else {
		m["time_start"] = nil
	}
	if c.TimeEnd != nil && !c.TimeEnd.IsZero() {
		m["time_end"] = c.TimeEnd.Format("15:04:05")
	} else {
		m["time_end"] = nil
	}
	return m
}
