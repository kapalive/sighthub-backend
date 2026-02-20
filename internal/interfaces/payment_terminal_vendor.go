// internal/interfaces/payment_terminal_vendor.go
package interfaces

// Интерфейс для работы с платёжными терминалами
type PaymentTerminalVendorInterface interface {
	GetPaymentTerminalByID(paymentTerminalID int) (map[string]interface{}, error)
}
