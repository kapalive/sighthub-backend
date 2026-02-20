package employees

import (
	"fmt"
)

// RatePerHour ⇄ rate_per_hour
type RatePerHour struct {
	IDRatePerHour int     `gorm:"column:id_rate_per_hour;primaryKey"                     json:"id_rate_per_hour"`
	RatePerHour   float64 `gorm:"column:rate_per_hour;type:numeric(10,2);not null"      json:"rate_per_hour"`
	Currency      string  `gorm:"column:currency;type:varchar(3);not null;default:USD"  json:"currency"`
	HoursPerWeek  float64 `gorm:"column:hours_per_week;type:double precision;not null;default:40" json:"hours_per_week"`
}

func (RatePerHour) TableName() string { return "rate_per_hour" }

// OverTime — эквивалент hybrid_property (1.5×)
func (r *RatePerHour) OverTime() float64 {
	return r.RatePerHour * 1.5
}

// ToMap — как в Python: числовые rate/over_time — строками
func (r *RatePerHour) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_rate_per_hour": r.IDRatePerHour,
		"rate_per_hour":    fmt.Sprintf("%.2f", r.RatePerHour),
		"currency":         r.Currency,
		"hours_per_week":   r.HoursPerWeek,
		"over_time":        fmt.Sprintf("%.2f", r.OverTime()),
	}
}

func (r *RatePerHour) String() string {
	return fmt.Sprintf("<RatePerHour %.2f %s | Overtime: %.2f>", r.RatePerHour, r.Currency, r.OverTime())
}
