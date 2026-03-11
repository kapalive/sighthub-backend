package reports

import "time"

type CashCountSheet struct {
	IDCashCountSheet    int64      `gorm:"column:id_cash_count_sheet;primaryKey;autoIncrement" json:"id_cash_count_sheet"`
	DailyClosePaymentID *int64     `gorm:"column:daily_close_payment_id"                       json:"daily_close_payment_id,omitempty"`
	Date                time.Time  `gorm:"column:date;type:date;not null"                      json:"-"`
	Cent1               int        `gorm:"column:cent_1;default:0"                             json:"cent_1"`
	Cent5               int        `gorm:"column:cent_5;default:0"                             json:"cent_5"`
	Cent10              int        `gorm:"column:cent_10;default:0"                            json:"cent_10"`
	Cent25              int        `gorm:"column:cent_25;default:0"                            json:"cent_25"`
	Cent50              int        `gorm:"column:cent_50;default:0"                            json:"cent_50"`
	Dollar1             int        `gorm:"column:dollar_1;default:0"                           json:"dollar_1"`
	Dollar2             int        `gorm:"column:dollar_2;default:0"                           json:"dollar_2"`
	Dollar5             int        `gorm:"column:dollar_5;default:0"                           json:"dollar_5"`
	Dollar10            int        `gorm:"column:dollar_10;default:0"                          json:"dollar_10"`
	Dollar20            int        `gorm:"column:dollar_20;default:0"                          json:"dollar_20"`
	Dollar50            int        `gorm:"column:dollar_50;default:0"                          json:"dollar_50"`
	Dollar100           int        `gorm:"column:dollar_100;default:0"                         json:"dollar_100"`
}

func (CashCountSheet) TableName() string { return "cash_count_sheet" }
