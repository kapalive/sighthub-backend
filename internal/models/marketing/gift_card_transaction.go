package marketing

import "time"

type GiftCardTransaction struct {
	IDTransaction          int        `gorm:"column:id_transaction;primaryKey;autoIncrement" json:"id_transaction"`
	GiftCardID             *int       `gorm:"column:gift_card_id"                            json:"gift_card_id,omitempty"`
	TransactionType        string     `gorm:"column:transaction_type;size:50;not null"       json:"transaction_type"`
	Amount                 string     `gorm:"column:amount;type:numeric(10,2);not null"      json:"amount"`
	CreatedAt              *time.Time `gorm:"column:created_at;type:timestamptz"             json:"created_at,omitempty"`
	ProcessedByPatientID   *int       `gorm:"column:processed_by_patient_id"                 json:"processed_by_patient_id,omitempty"`
	RelatedInvoiceID       *int       `gorm:"column:related_invoice_id"                      json:"related_invoice_id,omitempty"`
}

func (GiftCardTransaction) TableName() string { return "gift_card_transaction" }
