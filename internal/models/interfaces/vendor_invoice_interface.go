package interfaces

// VendorInvoiceInterface определяет методы для работы с моделью VendorInvoice
type VendorInvoiceInterface interface {
	ID() int64
	ToMap() map[string]interface{}
	String() string
}
