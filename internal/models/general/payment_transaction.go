// internal/models/general/payment_transaction.go
package general

import (
	"fmt"
	"time"
)

type PaymentTransaction struct {
	IDPaymentTransaction int64      `gorm:"column:id_payment_transaction;primaryKey"              json:"id_payment_transaction"`
	PaymentTerminalID    int        `gorm:"column:payment_terminal_id;not null;index"             json:"payment_terminal_id"`
	InvoiceID            *int64     `gorm:"column:invoice_id;index"                               json:"invoice_id,omitempty"`
	PaymentMethodID      *int64     `gorm:"column:payment_method_id;index"                        json:"payment_method_id,omitempty"`
	TransactionDate      time.Time  `gorm:"column:transaction_date;not null"                      json:"transaction_date"`
	CreatedAt            time.Time  `gorm:"column:created_at;not null"                            json:"created_at"`
	CompletedAt          *time.Time `gorm:"column:completed_at"                                   json:"completed_at,omitempty"`
	Amount               float64    `gorm:"column:amount;type:numeric(12,2);not null"             json:"amount"`
	Currency             string     `gorm:"column:currency;type:varchar(3);not null;default:USD"  json:"currency"`
	Status               string     `gorm:"column:status;type:varchar(20);not null;index"         json:"status"`
	PaymentMethod        *string    `gorm:"column:payment_method;type:varchar(50)"                json:"payment_method,omitempty"`
	CardMaskedNumber     *string    `gorm:"column:card_masked_number;type:varchar(20)"            json:"card_masked_number,omitempty"`
	AuthorizationCode    *string    `gorm:"column:authorization_code;type:varchar(50)"            json:"authorization_code,omitempty"`
	ReferenceNumber      *string    `gorm:"column:reference_number;type:varchar(50)"              json:"reference_number,omitempty"`

	// SPIn / iPOSpays fields
	SpinRefID       string  `gorm:"column:spin_ref_id;type:varchar(64);not null;uniqueIndex"  json:"spin_ref_id"`
	SpinResultCode  *string `gorm:"column:spin_result_code;type:varchar(8)"                   json:"spin_result_code,omitempty"`
	SpinRespMsg     *string `gorm:"column:spin_resp_msg;type:text"                            json:"spin_resp_msg,omitempty"`
	SpinPnRef       *string `gorm:"column:spin_pn_ref;type:varchar(64);index"                 json:"spin_pn_ref,omitempty"`
	SpinInvNum      *string `gorm:"column:spin_inv_num;type:varchar(64);index"                json:"spin_inv_num,omitempty"`
	SpinRegisterID  *string `gorm:"column:spin_register_id;type:varchar(64)"                  json:"spin_register_id,omitempty"`
	SpinTPN         *string `gorm:"column:spin_tpn;type:varchar(64)"                          json:"spin_tpn,omitempty"`
	SpinExtData     *string `gorm:"column:spin_ext_data;type:text"                            json:"spin_ext_data,omitempty"`
	SpinEmvData     *string `gorm:"column:spin_emv_data;type:text"                            json:"spin_emv_data,omitempty"`
	SpinSignB64     *string `gorm:"column:spin_sign_b64;type:text"                            json:"spin_sign_b64,omitempty"`
	SpinIposToken   *string `gorm:"column:spin_ipos_token;type:varchar(256)"                  json:"spin_ipos_token,omitempty"`
	SpinRequestXML  *string `gorm:"column:spin_request_xml;type:text"                         json:"spin_request_xml,omitempty"`
	SpinResponseXML *string `gorm:"column:spin_response_xml;type:text"                        json:"spin_response_xml,omitempty"`

	Terminal *PaymentTerminal `gorm:"foreignKey:PaymentTerminalID;references:IDPaymentTerminal" json:"-"`
}

func (PaymentTransaction) TableName() string { return "payment_transaction" }

func (t *PaymentTransaction) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_payment_transaction": t.IDPaymentTransaction,
		"payment_terminal_id":    t.PaymentTerminalID,
		"invoice_id":             t.InvoiceID,
		"transaction_date":       t.TransactionDate.Format(time.RFC3339),
		"amount":                 fmt.Sprintf("%.2f", t.Amount),
		"currency":               t.Currency,
		"status":                 t.Status,
		"payment_method":         t.PaymentMethod,
		"card_masked_number":     t.CardMaskedNumber,
		"authorization_code":     t.AuthorizationCode,
		"spin_ref_id":            t.SpinRefID,
		"spin_result_code":       t.SpinResultCode,
		"spin_resp_msg":          t.SpinRespMsg,
		"spin_pn_ref":            t.SpinPnRef,
		"created_at":             t.CreatedAt.Format(time.RFC3339),
	}
}

func (t *PaymentTransaction) String() string {
	return fmt.Sprintf("<PaymentTransaction %d>", t.IDPaymentTransaction)
}
