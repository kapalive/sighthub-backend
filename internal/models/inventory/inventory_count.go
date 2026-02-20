// internal/models/inventory/inventory_count.go
package inventory

import (
	"fmt"
	"sighthub-backend/internal/models/interfaces" // Импортируем интерфейс
	"time"
)

// Модель для подсчета инвентаря (InventoryCount)
type InventoryCount struct {
	IDInventoryCount  int64     `gorm:"column:id_inventory_count;primaryKey;autoIncrement" json:"id_inventory_count"`
	BrandID           int64     `gorm:"column:brand_id;not null" json:"brand_id"`
	LocationID        int64     `gorm:"column:location_id;not null" json:"location_id"`
	Status            bool      `gorm:"column:status;not null" json:"status"` // TRUE for OPEN, FALSE for CLOSED
	PrepByDate        time.Time `gorm:"column:prep_by_date;not null" json:"prep_by_date"`
	PrepByEmployeeID  int64     `gorm:"column:prep_by_employee_id;not null" json:"prep_by_employee_id"`
	CreatedDate       time.Time `gorm:"column:created_date;not null;default:CURRENT_TIMESTAMP" json:"created_date"`
	CreatedEmployeeID int64     `gorm:"column:created_employee_id;not null" json:"created_employee_id"`
	UpdatedDate       time.Time `gorm:"column:updated_date;not null;default:CURRENT_TIMESTAMP;onUpdate:CURRENT_TIMESTAMP" json:"updated_date"`
	UpdatedEmployeeID int64     `gorm:"column:updated_employee_id;not null" json:"updated_employee_id"`
	Quantity          int       `gorm:"column:quantity;not null" json:"quantity"`
	Cost              float64   `gorm:"column:cost;type:numeric(10,2)" json:"cost"`
	Notes             *string   `gorm:"column:notes" json:"notes,omitempty"`

	// Связь с интерфейсом BrandInterface для работы с брендом
	Brand interfaces.BrandInterface `gorm:"-" json:"brand,omitempty"`
}

// TableName задаёт имя таблицы в БД
func (InventoryCount) TableName() string {
	return "inventory_count"
}

// ToMap превращает объект в карту для удобства работы с данными
func (i *InventoryCount) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_inventory_count":  i.IDInventoryCount,
		"brand_id":            i.BrandID,
		"location_id":         i.LocationID,
		"status":              i.Status,
		"prep_by_date":        i.PrepByDate,
		"prep_by_employee_id": i.PrepByEmployeeID,
		"created_date":        i.CreatedDate,
		"created_employee_id": i.CreatedEmployeeID,
		"updated_date":        i.UpdatedDate,
		"updated_employee_id": i.UpdatedEmployeeID,
		"quantity":            i.Quantity,
		"cost":                i.Cost,
		"notes":               i.Notes,
	}
}

// String метод для печати объекта
func (i *InventoryCount) String() string {
	return fmt.Sprintf("<InventoryCount %d | BrandID: %d | LocationID: %d>", i.IDInventoryCount, i.BrandID, i.LocationID)
}

// Получить бренд по ID через интерфейс
func (i *InventoryCount) GetBrand(brandVendor interfaces.BrandInterface) (map[string]interface{}, error) {
	// Используем интерфейс для получения бренда
	return brandVendor.ToMap(), nil
}
