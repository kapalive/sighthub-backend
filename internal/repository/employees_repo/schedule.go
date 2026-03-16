package employees_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"

	schedulemodels "sighthub-backend/internal/models/schedule"
)

// ─────────────────────────────────────────────
// DTO types
// ─────────────────────────────────────────────

type DaySchedule struct {
	Date                string  `json:"date"`
	IsWorkingDay        bool    `json:"is_working_day"`
	TimeStart           *string `json:"time_start"`
	TimeEnd             *string `json:"time_end"`
	AppointmentDuration *int    `json:"appointment_duration"`
	Message             *string `json:"message"`
}

// ─────────────────────────────────────────────
// ScheduleRepo
// ─────────────────────────────────────────────

type ScheduleRepo struct {
	DB *gorm.DB
}

func NewScheduleRepo(db *gorm.DB) *ScheduleRepo {
	return &ScheduleRepo{DB: db}
}

// GetWeeklySchedule returns the weekly schedule template for an employee
// as a map keyed by day name (Monday … Sunday).
func (r *ScheduleRepo) GetWeeklySchedule(employeeID int) (map[string]interface{}, error) {
	var schedules []schedulemodels.Schedule
	if err := r.DB.Where("employee_id = ?", employeeID).Find(&schedules).Error; err != nil {
		return nil, err
	}

	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	result := make(map[string]interface{}, len(days))
	for _, day := range days {
		result[day] = nil
	}

	for _, s := range schedules {
		entry := map[string]interface{}{
			"id_schedule":          s.IDSchedule,
			"start_time":           s.StartTime,
			"end_time":             s.EndTime,
			"appointment_duration": s.AppointmentDuration,
		}
		if s.LunchStart != nil && *s.LunchStart != "" {
			entry["lunch_start"] = *s.LunchStart
		} else {
			entry["lunch_start"] = nil
		}
		if s.LunchEnd != nil && *s.LunchEnd != "" {
			entry["lunch_end"] = *s.LunchEnd
		} else {
			entry["lunch_end"] = nil
		}
		result[s.DayOfWeek] = entry
	}
	return result, nil
}

// SaveWeeklySchedule creates or updates Schedule entries for each day in `data`.
// Keys in data must match day-of-week names (e.g. "Monday").
// Expected fields per day: start_time, end_time, appointment_duration, lunch_start, lunch_end (all strings "HH:MM:SS").
func (r *ScheduleRepo) SaveWeeklySchedule(employeeID int, data map[string]interface{}) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		for dayName, raw := range data {
			dayData, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}

			var existing schedulemodels.Schedule
			err := tx.Where("employee_id = ? AND day_of_week = ?", employeeID, dayName).First(&existing).Error

			parseTime := func(key string) (string, bool) {
				v, ok := dayData[key]
				if !ok || v == nil {
					return "", false
				}
				s, ok := v.(string)
				if !ok || s == "" {
					return "", false
				}
				return s, true
			}

			startTime, hasStart := parseTime("start_time")
			endTime, hasEnd := parseTime("end_time")
			if !hasStart || !hasEnd {
				continue
			}

			apptDur := 15
			if v, ok := dayData["appointment_duration"]; ok && v != nil {
				switch val := v.(type) {
				case int:
					apptDur = val
				case float64:
					apptDur = int(val)
				}
			}

			var lunchStart, lunchEnd *string
			if t, ok := parseTime("lunch_start"); ok {
				lunchStart = &t
			}
			if t, ok := parseTime("lunch_end"); ok {
				lunchEnd = &t
			}

			if errors.Is(err, gorm.ErrRecordNotFound) {
				s := schedulemodels.Schedule{
					EmployeeID:          int64(employeeID),
					DayOfWeek:           dayName,
					StartTime:           startTime,
					EndTime:             endTime,
					AppointmentDuration: apptDur,
					LunchStart:          lunchStart,
					LunchEnd:            lunchEnd,
				}
				if err := tx.Create(&s).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				updates := map[string]interface{}{
					"start_time":           startTime,
					"end_time":             endTime,
					"appointment_duration": apptDur,
					"lunch_start":          lunchStart,
					"lunch_end":            lunchEnd,
				}
				if err := tx.Model(&existing).Updates(updates).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// GetScheduleWithDates returns a DaySchedule for each date in [from, to] inclusive.
func (r *ScheduleRepo) GetScheduleWithDates(employeeID int, from, to time.Time) ([]DaySchedule, error) {
	// Pre-load the full weekly schedule map
	weeklyMap, err := r.loadWeeklyScheduleMap(employeeID)
	if err != nil {
		return nil, err
	}

	result := make([]DaySchedule, 0)
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		dayName := d.Weekday().String() // "Monday", "Tuesday", …

		day := DaySchedule{
			Date:         dateStr,
			IsWorkingDay: true,
		}

		// Check Calendar for an override (off-day / holiday)
		var cal schedulemodels.Calendar
		empID64 := int64(employeeID)
		calErr := r.DB.
			Where("employee_id = ? AND date = ?", empID64, d.Truncate(24*time.Hour)).
			First(&cal).Error

		if calErr == nil {
			if !cal.IsWorkingDay {
				day.IsWorkingDay = false
				if cal.IsHoliday {
					msg := "Holiday"
					day.Message = &msg
				}
			} else if cal.TimeStart != nil {
				day.TimeStart = cal.TimeStart
				if cal.TimeEnd != nil {
					day.TimeEnd = cal.TimeEnd
				}
			}
		} else if !errors.Is(calErr, gorm.ErrRecordNotFound) {
			return nil, calErr
		} else {
			// No calendar override — use weekly template
			if s, ok := weeklyMap[dayName]; ok {
				ts := s.StartTime
				te := s.EndTime
				day.TimeStart = &ts
				day.TimeEnd = &te
				day.AppointmentDuration = &s.AppointmentDuration
			} else {
				day.IsWorkingDay = false
			}
		}

		result = append(result, day)
	}
	return result, nil
}

// AddOffDay marks the given date as a non-working holiday for the employee
// by inserting or updating the Calendar entry.
func (r *ScheduleRepo) AddOffDay(employeeID int, date time.Time) error {
	empID64 := int64(employeeID)
	dateOnly := date.Truncate(24 * time.Hour)

	var existing schedulemodels.Calendar
	err := r.DB.Where("employee_id = ? AND date = ?", empID64, dateOnly).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		cal := schedulemodels.Calendar{
			Date:         dateOnly,
			IsHoliday:    true,
			EmployeeID:   &empID64,
			IsWorkingDay: false,
		}
		return r.DB.Create(&cal).Error
	} else if err != nil {
		return err
	}
	return r.DB.Model(&existing).Updates(map[string]interface{}{
		"is_holiday":     true,
		"is_working_day": false,
	}).Error
}

// ListOffDays returns all off-day dates for the employee as "YYYY-MM-DD" strings.
func (r *ScheduleRepo) ListOffDays(employeeID int) ([]string, error) {
	var cals []schedulemodels.Calendar
	empID64 := int64(employeeID)
	if err := r.DB.
		Where("employee_id = ? AND is_working_day = false", empID64).
		Find(&cals).Error; err != nil {
		return nil, err
	}

	dates := make([]string, 0, len(cals))
	for _, c := range cals {
		dates = append(dates, c.Date.Format("2006-01-02"))
	}
	return dates, nil
}

// RemoveOffDay removes the off-day for the given date and employee.
// If a weekly schedule exists for that weekday, the calendar entry is set to
// is_working_day=true; otherwise it is deleted entirely.
func (r *ScheduleRepo) RemoveOffDay(employeeID int, date time.Time) error {
	empID64 := int64(employeeID)
	dateOnly := date.Truncate(24 * time.Hour)

	var cal schedulemodels.Calendar
	if err := r.DB.Where("employee_id = ? AND date = ?", empID64, dateOnly).First(&cal).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // nothing to remove
		}
		return err
	}

	// Check whether there is a weekly schedule for that day-of-week
	dayName := date.Weekday().String()
	var weeklyEntry schedulemodels.Schedule
	hasWeekly := r.DB.Where("employee_id = ? AND day_of_week = ?", employeeID, dayName).
		First(&weeklyEntry).Error == nil

	if hasWeekly {
		return r.DB.Model(&cal).Updates(map[string]interface{}{
			"is_working_day": true,
			"is_holiday":     false,
		}).Error
	}
	return r.DB.Delete(&cal).Error
}

// loadWeeklyScheduleMap returns a map of day_of_week → Schedule for an employee.
func (r *ScheduleRepo) loadWeeklyScheduleMap(employeeID int) (map[string]schedulemodels.Schedule, error) {
	var schedules []schedulemodels.Schedule
	if err := r.DB.Where("employee_id = ?", employeeID).Find(&schedules).Error; err != nil {
		return nil, err
	}
	m := make(map[string]schedulemodels.Schedule, len(schedules))
	for _, s := range schedules {
		m[s.DayOfWeek] = s
	}
	return m, nil
}
