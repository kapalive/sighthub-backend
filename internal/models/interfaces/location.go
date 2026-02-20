// internal/models/interfaces/location.go
package interfaces

// LocationInterface - интерфейс для работы с локациями
type LocationInterface interface {
	GetLocationByID(locationID int64) (map[string]interface{}, error)
}
