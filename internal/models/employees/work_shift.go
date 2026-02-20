package employees

import (
	"time"
)

// WorkShift ⇄ work_shift
type WorkShift struct {
	IDWorkShift int64   `gorm:"column:id_work_shift;primaryKey"                      json:"id_work_shift"`
	Title       *string `gorm:"column:title;type:varchar(100)"                       json:"title,omitempty"`

	Monday    bool `gorm:"column:monday;not null;default:true"                       json:"monday"`
	Tuesday   bool `gorm:"column:tuesday;not null;default:true"                      json:"tuesday"`
	Wednesday bool `gorm:"column:wednesday;not null;default:true"                    json:"wednesday"`
	Thursday  bool `gorm:"column:thursday;not null;default:true"                     json:"thursday"`
	Friday    bool `gorm:"column:friday;not null;default:true"                       json:"friday"`
	Saturday  bool `gorm:"column:saturday;not null;default:false"                    json:"saturday"`
	Sunday    bool `gorm:"column:sunday;not null;default:false"                      json:"sunday"`

	MondayTimeStart    time.Time  `gorm:"column:monday_time_start;type:time;not null;default:'10:00:00'"  json:"-"`
	MondayTimeEnd      time.Time  `gorm:"column:monday_time_end;type:time;not null;default:'19:00:00'"    json:"-"`
	TuesdayTimeStart   time.Time  `gorm:"column:tuesday_time_start;type:time;not null;default:'10:00:00'" json:"-"`
	TuesdayTimeEnd     time.Time  `gorm:"column:tuesday_time_end;type:time;not null;default:'19:00:00'"   json:"-"`
	WednesdayTimeStart time.Time  `gorm:"column:wednesday_time_start;type:time;not null;default:'10:00:00'" json:"-"`
	WednesdayTimeEnd   time.Time  `gorm:"column:wednesday_time_end;type:time;not null;default:'19:00:00'"  json:"-"`
	ThursdayTimeStart  time.Time  `gorm:"column:thursday_time_start;type:time;not null;default:'10:00:00'" json:"-"`
	ThursdayTimeEnd    time.Time  `gorm:"column:thursday_time_end;type:time;not null;default:'19:00:00'"   json:"-"`
	FridayTimeStart    time.Time  `gorm:"column:friday_time_start;type:time;not null;default:'10:00:00'"   json:"-"`
	FridayTimeEnd      time.Time  `gorm:"column:friday_time_end;type:time;not null;default:'19:00:00'"     json:"-"`
	SaturdayTimeStart  *time.Time `gorm:"column:saturday_time_start;type:time"                             json:"-"`
	SaturdayTimeEnd    *time.Time `gorm:"column:saturday_time_end;type:time"                               json:"-"`
	SundayTimeStart    *time.Time `gorm:"column:sunday_time_start;type:time"                               json:"-"`
	SundayTimeEnd      *time.Time `gorm:"column:sunday_time_end;type:time"                                 json:"-"`

	// Interval в PG. Проще всего держать как текст "HH:MM:SS" (или "00:30:00")
	LunchDuration *string `gorm:"column:lunch_duration;type:interval;default:'00:30:00'"                   json:"-"`
}

func (WorkShift) TableName() string { return "work_shift" }

// ToMap — эквивалент Python to_dict()
func (w *WorkShift) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_work_shift": w.IDWorkShift,
		"title":         w.Title,
		"monday":        w.Monday,
		"tuesday":       w.Tuesday,
		"wednesday":     w.Wednesday,
		"thursday":      w.Thursday,
		"friday":        w.Friday,
		"saturday":      w.Saturday,
		"sunday":        w.Sunday,

		"monday_time_start":    w.MondayTimeStart.Format("15:04:05"),
		"monday_time_end":      w.MondayTimeEnd.Format("15:04:05"),
		"tuesday_time_start":   w.TuesdayTimeStart.Format("15:04:05"),
		"tuesday_time_end":     w.TuesdayTimeEnd.Format("15:04:05"),
		"wednesday_time_start": w.WednesdayTimeStart.Format("15:04:05"),
		"wednesday_time_end":   w.WednesdayTimeEnd.Format("15:04:05"),
		"thursday_time_start":  w.ThursdayTimeStart.Format("15:04:05"),
		"thursday_time_end":    w.ThursdayTimeEnd.Format("15:04:05"),
		"friday_time_start":    w.FridayTimeStart.Format("15:04:05"),
		"friday_time_end":      w.FridayTimeEnd.Format("15:04:05"),
	}

	if w.SaturdayTimeStart != nil && !w.SaturdayTimeStart.IsZero() {
		m["saturday_time_start"] = w.SaturdayTimeStart.Format("15:04:05")
	} else {
		m["saturday_time_start"] = nil
	}
	if w.SaturdayTimeEnd != nil && !w.SaturdayTimeEnd.IsZero() {
		m["saturday_time_end"] = w.SaturdayTimeEnd.Format("15:04:05")
	} else {
		m["saturday_time_end"] = nil
	}
	if w.SundayTimeStart != nil && !w.SundayTimeStart.IsZero() {
		m["sunday_time_start"] = w.SundayTimeStart.Format("15:04:05")
	} else {
		m["sunday_time_start"] = nil
	}
	if w.SundayTimeEnd != nil && !w.SundayTimeEnd.IsZero() {
		m["sunday_time_end"] = w.SundayTimeEnd.Format("15:04:05")
	} else {
		m["sunday_time_end"] = nil
	}

	// строковое представление интервала, как в Python str(self.lunch_duration)
	if w.LunchDuration != nil {
		m["lunch_duration"] = *w.LunchDuration
	} else {
		m["lunch_duration"] = nil
	}

	return m
}
