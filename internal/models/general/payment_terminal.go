// internal/models/general/payment_terminal.go
package general

import "fmt"

type PaymentTerminal struct {
	IDPaymentTerminal int     `gorm:"column:id_payment_terminal;primaryKey"   json:"id_payment_terminal"`
	SerialNamber      *string `gorm:"column:serial_namber;type:varchar(255)"  json:"serial_namber,omitempty"`
}

func (PaymentTerminal) TableName() string { return "payment_terminal" }

func (t *PaymentTerminal) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_payment_terminal": t.IDPaymentTerminal,
		"serial_namber":       t.SerialNamber,
	}
}

func (t *PaymentTerminal) String() string {
	return fmt.Sprintf("<PaymentTerminal %d>", t.IDPaymentTerminal)
}
