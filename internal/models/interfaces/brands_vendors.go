package interfaces

// BrandInterface - интерфейс для Brand
type BrandInterface interface {
	ToMap() map[string]interface{}
	GetBrandByID(brandID int64) (map[string]interface{}, error)
}

// VendorInterface defines the required methods for Vendor
type VendorInterface interface {
	ID() int64
	Name() string // Добавлен метод Name
	ToMap() map[string]interface{}
	GetBrandByID(brandID int64) (map[string]interface{}, error)
	GetVendorByID(vendorID int64) (map[string]interface{}, error)
}
