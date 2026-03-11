package reports

import "time"

type DailyClosePayment struct {
	IDDailyClosePayment int64      `gorm:"column:id_daily_close_payment;primaryKey;autoIncrement" json:"id_daily_close_payment"`
	PaymentMethodID     int64      `gorm:"column:payment_method_id;not null"                      json:"payment_method_id"`
	Date                time.Time  `gorm:"column:date;type:date;not null"                         json:"-"`
	Amount              float64    `gorm:"column:amount;type:numeric(10,2);not null"              json:"amount"`
	LocationID          int64      `gorm:"column:location_id;not null"                            json:"location_id"`
	CreatedAt           *time.Time `gorm:"column:created_at;type:timestamptz"                     json:"created_at,omitempty"`
	Note                *string    `gorm:"column:note;size:255"                                   json:"note,omitempty"`
}

func (DailyClosePayment) TableName() string { return "daily_close_payment" }
