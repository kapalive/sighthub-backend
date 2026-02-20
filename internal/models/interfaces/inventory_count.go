// internal/models/interfaces/inventory_count.go
package interfaces

// InventoryCountInterface - интерфейс для работы с InventoryCount
type InventoryCountInterface interface {
	GetInventoryCountByID(inventoryCountID int64) (map[string]interface{}, error)
}
