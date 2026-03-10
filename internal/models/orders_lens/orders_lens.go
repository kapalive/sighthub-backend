// internal/models/orders_lens/orders_lens.go
package orders_lens

import "time"

// OrdersLens ⇄ table: orders_lens
type OrdersLens struct {
	IDOrdersLens       int64      `gorm:"column:id_orders_lens;primaryKey;autoIncrement"            json:"id_orders_lens"`
	NumberOrder        string     `gorm:"column:number_order;type:varchar(16);not null"             json:"number_order"`
	DateCreate         time.Time  `gorm:"column:date_create;type:date;not null"                     json:"date_create"`
	PromisedDate       time.Time  `gorm:"column:promised_date;type:date;not null"                   json:"promised_date"`
	PromisedTimeBy     *string    `gorm:"column:promised_time_by;type:varchar(6)"                   json:"promised_time_by,omitempty"`
	StatusOrdersLensID int        `gorm:"column:status_orders_lens_id;not null"                     json:"status_orders_lens_id"`
	LensID             *int       `gorm:"column:lens_id"                                            json:"lens_id,omitempty"`
	PatientID          *int64     `gorm:"column:patient_id"                                         json:"patient_id,omitempty"`
	Note               *string    `gorm:"column:note;type:text"                                     json:"note,omitempty"`

	Status *StatusOrdersLens `gorm:"foreignKey:StatusOrdersLensID;references:IDStatusOrdersLens" json:"-"`
}

func (OrdersLens) TableName() string { return "orders_lens" }

func (o *OrdersLens) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_orders_lens":       o.IDOrdersLens,
		"number_order":         o.NumberOrder,
		"date_create":          o.DateCreate.Format("2006-01-02"),
		"promised_date":        o.PromisedDate.Format("2006-01-02"),
		"promised_time_by":     o.PromisedTimeBy,
		"status_orders_lens_id": o.StatusOrdersLensID,
		"lens_id":              o.LensID,
		"patient_id":           o.PatientID,
		"note":                 o.Note,
	}
	if o.Status != nil {
		m["status"] = o.Status.ToMap()
	}
	return m
}
