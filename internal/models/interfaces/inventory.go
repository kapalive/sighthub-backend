// internal/models/interfaces/inventory.go
package interfaces

// InventoryInterface определяет необходимые методы для модели Inventory
type InventoryInterface interface {
	ID() int64
	ToMap() map[string]interface{}
	GetInventoryByID(inventoryID int64) (map[string]interface{}, error)
}
