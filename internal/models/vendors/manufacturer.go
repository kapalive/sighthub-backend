// internal/models/vendors/manufacturer.go
package vendors

// Manufacturer ⇄ manufacturer
type Manufacturer struct {
	IDManufacturer   int    `gorm:"column:id_manufacturer;primaryKey;autoIncrement" json:"id_manufacturer"`
	ManufacturerName string `gorm:"column:manufacturer_name;type:varchar(255);not null" json:"manufacturer_name"`
}

func (Manufacturer) TableName() string { return "manufacturer" }

func (m *Manufacturer) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_manufacturer":   m.IDManufacturer,
		"manufacturer_name": m.ManufacturerName,
	}
}
