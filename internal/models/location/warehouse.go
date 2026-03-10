// internal/models/location/warehouse.go
package location

// Warehouse ⇄ table: warehouse
type Warehouse struct {
	IDWarehouse   int     `gorm:"column:id_warehouse;primaryKey;autoIncrement" json:"id_warehouse"`
	FullName      *string `gorm:"column:full_name;type:varchar(100)"           json:"full_name,omitempty"`
	ShortName     *string `gorm:"column:short_name;type:varchar(2)"            json:"short_name,omitempty"`
	StreetAddress *string `gorm:"column:street_address;type:varchar(100)"      json:"street_address,omitempty"`
	AddressLine2  *string `gorm:"column:address_line_2;type:varchar(100)"      json:"address_line_2,omitempty"`
	City          *string `gorm:"column:city;type:varchar(100)"                json:"city,omitempty"`
	State         *string `gorm:"column:state;type:varchar(2)"                 json:"state,omitempty"`
	PostalCode    *string `gorm:"column:postal_code;type:varchar(20)"          json:"postal_code,omitempty"`
	Country       *string `gorm:"column:country;type:varchar(50)"              json:"country,omitempty"`
	Phone         *string `gorm:"column:phone;type:varchar(20)"                json:"phone,omitempty"`
	TimeZone      *string `gorm:"column:time_zone;type:varchar(200)"           json:"time_zone,omitempty"`
	Fax           *string `gorm:"column:fax;type:varchar(20)"                  json:"fax,omitempty"`
}

func (Warehouse) TableName() string { return "warehouse" }

func (w *Warehouse) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_warehouse":   w.IDWarehouse,
		"full_name":      w.FullName,
		"short_name":     w.ShortName,
		"street_address": w.StreetAddress,
		"address_line_2": w.AddressLine2,
		"city":           w.City,
		"state":          w.State,
		"postal_code":    w.PostalCode,
		"country":        w.Country,
		"phone":          w.Phone,
		"time_zone":      w.TimeZone,
		"fax":            w.Fax,
	}
}
