// internal/models/reports/missing_ar.go
package reports

import "time"

// MissingAR ⇄ missing_ar
type MissingAR struct {
	IDMissingAR  int       `gorm:"column:id_missing_ar;primaryKey;autoIncrement"    json:"id_missing_ar"`
	ARCountID    int       `gorm:"column:ar_count_id;not null"                      json:"ar_count_id"`
	InvoiceID    int64     `gorm:"column:invoice_id;not null"                       json:"invoice_id"`
	LocationID   int       `gorm:"column:location_id;not null"                      json:"location_id"`
	ReportedDate time.Time `gorm:"column:reported_date;type:timestamp;not null;default:now()" json:"-"`
	Notes        *string   `gorm:"column:notes;type:varchar(255)"                   json:"notes,omitempty"`
}

func (MissingAR) TableName() string { return "missing_ar" }
