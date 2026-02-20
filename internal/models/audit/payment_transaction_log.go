package audit

import (
	"fmt"
	"time"
)

// Модель логов транзакции платежа
type PaymentTransactionLog struct {
	IDPaymentTransactionLog int64     `gorm:"column:id_payment_transaction_log;primaryKey;autoIncrement" json:"id_payment_transaction_log"`
	PaymentTransactionID    int64     `gorm:"column:payment_transaction_id;not null" json:"payment_transaction_id"`
	LogMessage              string    `gorm:"column:log_message;type:text;not null" json:"log_message"`
	LoggedAt                time.Time `gorm:"column:logged_at;not null;default:CURRENT_TIMESTAMP" json:"logged_at"`
}

// TableName задаёт имя таблицы в БД
func (PaymentTransactionLog) TableName() string { return "payment_transaction_log" }

// ToMap превращает объект в карту для удобства работы с данными
func (p *PaymentTransactionLog) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_payment_transaction_log": p.IDPaymentTransactionLog,
		"payment_transaction_id":     p.PaymentTransactionID,
		"log_message":                p.LogMessage,
		"logged_at":                  p.LoggedAt,
	}
}

// String метод для печати объекта
func (p *PaymentTransactionLog) String() string {
	return fmt.Sprintf("<PaymentTransactionLog %d>", p.IDPaymentTransactionLog)
}
