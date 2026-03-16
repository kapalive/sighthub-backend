package employees

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

	MondayTimeStart    string  `gorm:"column:monday_time_start;type:time;not null;default:'10:00:00'"  json:"-"`
	MondayTimeEnd      string  `gorm:"column:monday_time_end;type:time;not null;default:'19:00:00'"    json:"-"`
	TuesdayTimeStart   string  `gorm:"column:tuesday_time_start;type:time;not null;default:'10:00:00'" json:"-"`
	TuesdayTimeEnd     string  `gorm:"column:tuesday_time_end;type:time;not null;default:'19:00:00'"   json:"-"`
	WednesdayTimeStart string  `gorm:"column:wednesday_time_start;type:time;not null;default:'10:00:00'" json:"-"`
	WednesdayTimeEnd   string  `gorm:"column:wednesday_time_end;type:time;not null;default:'19:00:00'"  json:"-"`
	ThursdayTimeStart  string  `gorm:"column:thursday_time_start;type:time;not null;default:'10:00:00'" json:"-"`
	ThursdayTimeEnd    string  `gorm:"column:thursday_time_end;type:time;not null;default:'19:00:00'"   json:"-"`
	FridayTimeStart    string  `gorm:"column:friday_time_start;type:time;not null;default:'10:00:00'"   json:"-"`
	FridayTimeEnd      string  `gorm:"column:friday_time_end;type:time;not null;default:'19:00:00'"     json:"-"`
	SaturdayTimeStart  *string `gorm:"column:saturday_time_start;type:time"                             json:"-"`
	SaturdayTimeEnd    *string `gorm:"column:saturday_time_end;type:time"                               json:"-"`
	SundayTimeStart    *string `gorm:"column:sunday_time_start;type:time"                               json:"-"`
	SundayTimeEnd      *string `gorm:"column:sunday_time_end;type:time"                                 json:"-"`

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

		"monday_time_start":    w.MondayTimeStart,
		"monday_time_end":      w.MondayTimeEnd,
		"tuesday_time_start":   w.TuesdayTimeStart,
		"tuesday_time_end":     w.TuesdayTimeEnd,
		"wednesday_time_start": w.WednesdayTimeStart,
		"wednesday_time_end":   w.WednesdayTimeEnd,
		"thursday_time_start":  w.ThursdayTimeStart,
		"thursday_time_end":    w.ThursdayTimeEnd,
		"friday_time_start":    w.FridayTimeStart,
		"friday_time_end":      w.FridayTimeEnd,
	}

	if w.SaturdayTimeStart != nil && *w.SaturdayTimeStart != "" {
		m["saturday_time_start"] = *w.SaturdayTimeStart
	} else {
		m["saturday_time_start"] = nil
	}
	if w.SaturdayTimeEnd != nil && *w.SaturdayTimeEnd != "" {
		m["saturday_time_end"] = *w.SaturdayTimeEnd
	} else {
		m["saturday_time_end"] = nil
	}
	if w.SundayTimeStart != nil && *w.SundayTimeStart != "" {
		m["sunday_time_start"] = *w.SundayTimeStart
	} else {
		m["sunday_time_start"] = nil
	}
	if w.SundayTimeEnd != nil && *w.SundayTimeEnd != "" {
		m["sunday_time_end"] = *w.SundayTimeEnd
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
