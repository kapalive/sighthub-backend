package interfaces

// Interface for Product
type ProductInterface interface {
	ToMap() map[string]interface{}
}

// Interface for Model
type ModelInterface interface {
	ToMap() map[string]interface{}
	GetModelByID(modelID int64) (map[string]interface{}, error)
}
