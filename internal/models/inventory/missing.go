// internal/models/inventory/missing.go
package inventory

import (
	"fmt"
	"sighthub-backend/internal/models/interfaces" // Импортируем интерфейсы
	"time"
)

// Модель для отсутствующего инвентаря (Missing)
type Missing struct {
	IDMissing        int64     `gorm:"column:id_missing;primaryKey;autoIncrement" json:"id_missing"`
	InventoryCountID int64     `gorm:"column:inventory_count_id;not null" json:"inventory_count_id"`
	InventoryID      int64     `gorm:"column:inventory_id;not null" json:"inventory_id"`
	LocationID       int64     `gorm:"column:location_id;not null" json:"location_id"`
	BrandID          int64     `gorm:"column:brand_id;not null" json:"brand_id"`
	ModelID          int64     `gorm:"column:model_id;not null" json:"model_id"`
	Quantity         int       `gorm:"column:quantity;not null" json:"quantity"`
	Cost             float64   `gorm:"column:cost;type:numeric(10,2)" json:"cost"`
	ReportedDate     time.Time `gorm:"column:reported_date;default:CURRENT_TIMESTAMP" json:"reported_date"`
	Notes            *string   `gorm:"column:notes" json:"notes,omitempty"`
}

// TableName задаёт имя таблицы в БД
func (Missing) TableName() string {
	return "missing"
}

// ToMap превращает объект в карту для удобства работы с данными
func (m *Missing) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_missing":         m.IDMissing,
		"inventory_count_id": m.InventoryCountID,
		"inventory_id":       m.InventoryID,
		"location_id":        m.LocationID,
		"brand_id":           m.BrandID,
		"model_id":           m.ModelID,
		"quantity":           m.Quantity,
		"cost":               m.Cost,
		"reported_date":      m.ReportedDate,
		"notes":              m.Notes,
	}
}

// String метод для печати объекта
func (m *Missing) String() string {
	return fmt.Sprintf("<Missing %d | InventoryCountID: %d | InventoryID: %d | LocationID: %d | BrandID: %d | ModelID: %d>",
		m.IDMissing, m.InventoryCountID, m.InventoryID, m.LocationID, m.BrandID, m.ModelID)
}

// Получить данные о подсчете инвентаря (InventoryCount) через интерфейс
func (m *Missing) GetInventoryCount(inventoryCountVendor interfaces.InventoryCountInterface) (map[string]interface{}, error) {
	return inventoryCountVendor.GetInventoryCountByID(m.InventoryCountID)
}

// Получить данные об инвентаре через интерфейс
func (m *Missing) GetInventory(inventoryVendor interfaces.InventoryInterface) (map[string]interface{}, error) {
	return inventoryVendor.GetInventoryByID(m.InventoryID)
}

// Получить данные о местоположении через интерфейс
func (m *Missing) GetLocation(locationVendor interfaces.LocationInterface) (map[string]interface{}, error) {
	return locationVendor.GetLocationByID(m.LocationID)
}

// Получить данные о бренде через интерфейс
func (m *Missing) GetBrand(brandVendor interfaces.BrandInterface) (map[string]interface{}, error) {
	return brandVendor.GetBrandByID(m.BrandID)
}

// Получить данные о модели через интерфейс
func (m *Missing) GetModel(modelVendor interfaces.ModelInterface) (map[string]interface{}, error) {
	return modelVendor.GetModelByID(m.ModelID)
}
