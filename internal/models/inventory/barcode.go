// internal/models/inventory/barcode.go
package inventory

import (
	"fmt"
	"time"
)

type Barcode struct {
	IDBarcode   int64     `gorm:"column:id_barcode;primaryKey"           json:"id_barcode"`
	InventoryID int64     `gorm:"column:inventory_id;not null"           json:"inventory_id"`
	BarcodePath *string   `gorm:"column:barcode_path;type:varchar(255)"  json:"barcode_path,omitempty"`
	DateUpload  time.Time `gorm:"column:date_upload;default:CURRENT_TIMESTAMP" json:"date_upload"`

	// --- relations (preload when needed) ---
	Inventory *Inventory `gorm:"foreignKey:InventoryID;references:IDInventory" json:"-"`
}

func (Barcode) TableName() string { return "barcode" }

func (b *Barcode) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_barcode":   b.IDBarcode,
		"inventory_id": b.InventoryID,
		"barcode_path": b.BarcodePath,
		"date_upload":  b.DateUpload,
	}
}

func (b *Barcode) String() string {
	return fmt.Sprintf("<Barcode %d>", b.IDBarcode)
}
