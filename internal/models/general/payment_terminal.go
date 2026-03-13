// internal/models/general/payment_terminal.go
package general

import (
	"fmt"
	"time"
)

type PaymentTerminal struct {
	IDPaymentTerminal int       `gorm:"column:id_payment_terminal;primaryKey"                        json:"id_payment_terminal"`
	LocationID        *int64    `gorm:"column:location_id;index"                                     json:"location_id,omitempty"`
	Title             *string   `gorm:"column:title;type:varchar(120)"                               json:"title,omitempty"`
	SerialNumber      *string   `gorm:"column:serial_namber;type:varchar(120)"                       json:"serial_number,omitempty"`
	SpinRegisterID    *string   `gorm:"column:spin_register_id;type:varchar(64);index"               json:"spin_register_id,omitempty"`
	SpinTPN           *string   `gorm:"column:spin_tpn;type:varchar(64);index"                       json:"spin_tpn,omitempty"`
	IsDefault         bool      `gorm:"column:is_default;not null;default:false"                     json:"is_default"`
	Active            *bool     `gorm:"column:active"                                                json:"active,omitempty"`
	CreatedAt         time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now()"    json:"created_at"`
}

func (PaymentTerminal) TableName() string { return "payment_terminal" }

func (t *PaymentTerminal) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_payment_terminal": t.IDPaymentTerminal,
		"location_id":         t.LocationID,
		"title":               t.Title,
		"serial_number":       t.SerialNumber,
		"spin_register_id":    t.SpinRegisterID,
		"spin_tpn":            t.SpinTPN,
		"is_default":          t.IsDefault,
		"active":              t.Active,
		"created_at":          t.CreatedAt.Format(time.RFC3339),
	}
}

func (t *PaymentTerminal) String() string {
	return fmt.Sprintf("<PaymentTerminal %d>", t.IDPaymentTerminal)
}
