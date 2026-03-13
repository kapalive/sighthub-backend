// internal/models/inventory/missing.go
package inventory

import (
	"fmt"
	"time"
)

type Missing struct {
	IDMissing        int64     `gorm:"column:id_missing;primaryKey;autoIncrement" json:"id_missing"`
	InventoryCountID int64     `gorm:"column:inventory_count_id;not null" json:"inventory_count_id"`
	InventoryID      int64     `gorm:"column:inventory_id;not null" json:"inventory_id"`
	LocationID       int64     `gorm:"column:location_id;not null" json:"location_id"`
	BrandID          *int64    `gorm:"column:brand_id" json:"brand_id,omitempty"`
	VendorID         *int64    `gorm:"column:vendor_id" json:"vendor_id,omitempty"`
	ModelID          int64     `gorm:"column:model_id;not null" json:"model_id"`
	Quantity         int       `gorm:"column:quantity;not null" json:"quantity"`
	Cost             float64   `gorm:"column:cost;type:numeric(10,2)" json:"cost"`
	ReportedDate     time.Time `gorm:"column:reported_date;default:CURRENT_TIMESTAMP" json:"reported_date"`
	Notes            *string   `gorm:"column:notes" json:"notes,omitempty"`
}

func (Missing) TableName() string {
	return "missing"
}

func (m *Missing) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_missing":         m.IDMissing,
		"inventory_count_id": m.InventoryCountID,
		"inventory_id":       m.InventoryID,
		"location_id":        m.LocationID,
		"brand_id":           m.BrandID,
		"vendor_id":          m.VendorID,
		"model_id":           m.ModelID,
		"quantity":           m.Quantity,
		"cost":               m.Cost,
		"reported_date":      m.ReportedDate,
		"notes":              m.Notes,
	}
}

func (m *Missing) String() string {
	return fmt.Sprintf("<Missing %d | InventoryCountID: %d | InventoryID: %d>",
		m.IDMissing, m.InventoryCountID, m.InventoryID)
}
