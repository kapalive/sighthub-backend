// internal/models/reports/ar_count.go
package reports

import "time"

// ARCount ⇄ ar_count
type ARCount struct {
	IDARCount           int       `gorm:"column:id_ar_count;primaryKey;autoIncrement" json:"id_ar_count"`
	LocationID          int       `gorm:"column:location_id;not null"                 json:"location_id"`
	Status              bool      `gorm:"column:status;not null"                      json:"status"` // true=OPEN, false=CLOSED
	PrepByDate          time.Time `gorm:"column:prep_by_date;type:timestamp;not null" json:"-"`
	PrepByEmployeeID    int       `gorm:"column:prep_by_employee_id;not null"         json:"prep_by_employee_id"`
	CreatedDate         time.Time `gorm:"column:created_date;type:timestamp;not null;default:now()" json:"-"`
	CreatedEmployeeID   int       `gorm:"column:created_employee_id;not null"         json:"created_employee_id"`
	UpdatedDate         time.Time `gorm:"column:updated_date;type:timestamp;not null;default:now()" json:"-"`
	UpdatedEmployeeID   int       `gorm:"column:updated_employee_id;not null"         json:"updated_employee_id"`
	Quantity            int       `gorm:"column:quantity;not null;default:0"          json:"quantity"`
	Notes               *string   `gorm:"column:notes;type:varchar(255)"              json:"notes,omitempty"`
}

func (ARCount) TableName() string { return "ar_count" }
