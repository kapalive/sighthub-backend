package patients

import "time"

type TransferCredit struct {
	IDTransferCredit int64      `gorm:"column:id_transfer_credit;primaryKey;autoIncrement" json:"id_transfer_credit"`
	InvoiceID        int64      `gorm:"column:invoice_id;not null"                         json:"invoice_id"`
	PatientID        *int64     `gorm:"column:patient_id"                                  json:"patient_id,omitempty"`
	Amount           float64    `gorm:"column:amount;type:numeric(10,2);not null"          json:"amount"`
	Note             *string    `gorm:"column:note;type:text"                              json:"note,omitempty"`
	CreatedAt        *time.Time `gorm:"column:created_at;type:timestamptz"                 json:"created_at,omitempty"`
}

func (TransferCredit) TableName() string { return "transfer_credit" }
