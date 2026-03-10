// internal/models/reports/temp_count_ar.go
package reports

import "time"

// TempCountAR ⇄ temp_count_ar
type TempCountAR struct {
	IDTempCountAR int       `gorm:"column:id_temp_count_ar;primaryKey;autoIncrement"   json:"id_temp_count_ar"`
	CountDate     time.Time `gorm:"column:count_date;type:timestamp;not null;default:now()" json:"-"`
	InvoiceID     int64     `gorm:"column:invoice_id;not null"                          json:"invoice_id"`
	LocationID    int       `gorm:"column:location_id;not null"                         json:"location_id"`
	ARCountID     int       `gorm:"column:ar_count_id;not null"                         json:"ar_count_id"`
}

func (TempCountAR) TableName() string { return "temp_count_ar" }
