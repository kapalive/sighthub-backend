// internal/models/general/payment_transaction.go
package general

import (
	"fmt"
	"time"
)

type PaymentTransaction struct {
	IDPaymentTransaction int64     `gorm:"column:id_payment_transaction;primaryKey"      json:"id_payment_transaction"`
	PaymentTerminalID    int       `gorm:"column:payment_terminal_id;not null"           json:"payment_terminal_id"`
	TransactionDate      time.Time `gorm:"column:transaction_date;not null"              json:"transaction_date"`
	Amount               *string   `gorm:"column:amount;type:numeric(12,2);not null"     json:"amount,omitempty"`
	Currency             string    `gorm:"column:currency;type:varchar(3);not null"      json:"currency"`
	Status               string    `gorm:"column:status;type:varchar(20);not null"       json:"status"`
	PaymentMethod        *string   `gorm:"column:payment_method;type:varchar(50)"        json:"payment_method,omitempty"`
	CardMaskedNumber     *string   `gorm:"column:card_masked_number;type:varchar(20)"    json:"card_masked_number,omitempty"`
	AuthorizationCode    *string   `gorm:"column:authorization_code;type:varchar(50)"    json:"authorization_code,omitempty"`
	ReferenceNumber      *string   `gorm:"column:reference_number;type:varchar(50)"      json:"reference_number,omitempty"`
	CreatedAt            time.Time `gorm:"column:created_at;not null"                    json:"created_at"`

	Terminal *PaymentTerminal `gorm:"foreignKey:PaymentTerminalID;references:IDPaymentTerminal" json:"-"`
}

func (PaymentTransaction) TableName() string { return "payment_transaction" }

func (t *PaymentTransaction) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_payment_transaction": t.IDPaymentTransaction,
		"payment_terminal_id":    t.PaymentTerminalID,
		"transaction_date":       t.TransactionDate.Format(time.RFC3339),
		"currency":               t.Currency,
		"status":                 t.Status,
		"created_at":             t.CreatedAt.Format(time.RFC3339),
	}
	if t.Amount != nil {
		m["amount"] = *t.Amount
	} else {
		m["amount"] = nil
	}
	if t.PaymentMethod != nil {
		m["payment_method"] = *t.PaymentMethod
	} else {
		m["payment_method"] = nil
	}
	if t.CardMaskedNumber != nil {
		m["card_masked_number"] = *t.CardMaskedNumber
	} else {
		m["card_masked_number"] = nil
	}
	if t.AuthorizationCode != nil {
		m["authorization_code"] = *t.AuthorizationCode
	} else {
		m["authorization_code"] = nil
	}
	if t.ReferenceNumber != nil {
		m["reference_number"] = *t.ReferenceNumber
	} else {
		m["reference_number"] = nil
	}
	return m
}

func (t *PaymentTransaction) String() string {
	return fmt.Sprintf("<PaymentTransaction %d>", t.IDPaymentTransaction)
}
