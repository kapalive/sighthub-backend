package interfaces

// InvoiceInterface определяет методы, которые требуются для работы с Invoice
type InvoiceInterface interface {
	ToMap() map[string]interface{}
	String() string
	CalculateTax() float64
	GetInvoiceByID(invoiceID int64) (map[string]interface{}, error) // Добавлен метод GetInvoiceByID
}
