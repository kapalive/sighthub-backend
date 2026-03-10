// internal/models/inventory/shipment.go
package inventory

import "time"

// Shipment ⇄ table: shipment
// Хранит данные об отгрузке/получении товаров от поставщика.
type Shipment struct {
	IDShipment        int64      `gorm:"column:id_shipment;primaryKey;autoIncrement"         json:"id_shipment"`
	VendorID          int64      `gorm:"column:vendor_id;not null"                           json:"vendor_id"`
	LocationID        int64      `gorm:"column:location_id;not null"                         json:"location_id"`
	BrandID           int64      `gorm:"column:brand_id"                                     json:"brand_id"`
	QtyOk             int        `gorm:"column:qty_ok;not null;default:0"                    json:"qty_ok"`
	QtyHold           int        `gorm:"column:qty_hold;not null;default:0"                  json:"qty_hold"`
	QtyShort          int        `gorm:"column:qty_short;not null;default:0"                 json:"qty_short"`
	QtyOver           int        `gorm:"column:qty_over;not null;default:0"                  json:"qty_over"`
	Cost              float64    `gorm:"column:cost;type:numeric(10,2);not null"             json:"cost"`
	EmployeeIDPrepBy  int64      `gorm:"column:employee_id_prep_by"                          json:"employee_id_prep_by"`
	EmployeeIDCreated int64      `gorm:"column:employee_id_created"                          json:"employee_id_created"`
	DateReceived      time.Time  `gorm:"column:date_received;not null;default:now()"         json:"date_received"`
	Status            *string    `gorm:"column:status;type:varchar(20)"                      json:"status,omitempty"`
	Notes             *string    `gorm:"column:notes;type:text"                              json:"notes,omitempty"`
	VendorInvoiceID   *int64     `gorm:"column:vendor_invoice_id"                            json:"vendor_invoice_id,omitempty"`
}

func (Shipment) TableName() string { return "shipment" }

func (s *Shipment) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_shipment":          s.IDShipment,
		"vendor_id":            s.VendorID,
		"location_id":          s.LocationID,
		"brand_id":             s.BrandID,
		"qty_ok":               s.QtyOk,
		"qty_hold":             s.QtyHold,
		"qty_short":            s.QtyShort,
		"qty_over":             s.QtyOver,
		"cost":                 s.Cost,
		"employee_id_prep_by":  s.EmployeeIDPrepBy,
		"employee_id_created":  s.EmployeeIDCreated,
		"date_received":        s.DateReceived,
		"status":               s.Status,
		"notes":                s.Notes,
		"vendor_invoice_id":    s.VendorInvoiceID,
	}
}
