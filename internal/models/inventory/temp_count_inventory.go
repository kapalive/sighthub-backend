// internal/models/inventory/temp_count_inventory.go
package inventory

import "time"

// TempCountInventory ⇄ table: temp_count_inventory
// Временное хранение позиций во время процесса физического пересчёта инвентаря.
type TempCountInventory struct {
	IDTempCount      int       `gorm:"column:id_temp_count;primaryKey;autoIncrement"  json:"id_temp_count"`
	CountDate        time.Time `gorm:"column:count_date;not null;default:now()"       json:"count_date"`
	InventoryID      int64     `gorm:"column:inventory_id;not null"                   json:"inventory_id"`
	LocationID       int       `gorm:"column:location_id;not null"                    json:"location_id"`
	BrandID          *int      `gorm:"column:brand_id"                                json:"brand_id,omitempty"`
	VendorID         *int      `gorm:"column:vendor_id"                               json:"vendor_id,omitempty"`
	InStock          bool      `gorm:"column:in_stock;not null;default:false"         json:"in_stock"`
	InventoryCountID int64     `gorm:"column:inventory_count_id;not null"             json:"inventory_count_id"`
}

func (TempCountInventory) TableName() string { return "temp_count_inventory" }

func (t *TempCountInventory) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_temp_count":      t.IDTempCount,
		"count_date":         t.CountDate,
		"inventory_id":       t.InventoryID,
		"location_id":        t.LocationID,
		"brand_id":           t.BrandID,
		"vendor_id":          t.VendorID,
		"in_stock":           t.InStock,
		"inventory_count_id": t.InventoryCountID,
	}
}
