// internal/models/general/payment_method.go
package general

import "fmt"

type PaymentMethod struct {
	IDPaymentMethod int     `gorm:"column:id_payment_method;primaryKey"          json:"id_payment_method"`
	MethodName      string  `gorm:"column:method_name;type:varchar(30);not null" json:"method_name"`
	ShortName       *string `gorm:"column:short_name;type:varchar(2)"            json:"short_name,omitempty"`
}

func (PaymentMethod) TableName() string { return "payment_method" }

func (m *PaymentMethod) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_payment_method": m.IDPaymentMethod,
		"method_name":       m.MethodName,
		"short_name":        m.ShortName,
	}
}

func (m *PaymentMethod) String() string {
	return fmt.Sprintf("<PaymentMethod %s>", m.MethodName)
}
