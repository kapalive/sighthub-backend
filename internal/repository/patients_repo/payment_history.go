// internal/repository/patients_repo/payment_history.go
package patients_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/patients"
)

type PaymentHistoryRepo struct{ DB *gorm.DB }

func NewPaymentHistoryRepo(db *gorm.DB) *PaymentHistoryRepo {
	return &PaymentHistoryRepo{DB: db}
}

// GetByInvoiceID возвращает все платежи для данного инвойса.
type PaymentRow struct {
	PaymentID       int64   `json:"payment_id"`
	InvoiceID       int64   `json:"invoice_id"`
	PatientID       *int64  `json:"patient_id,omitempty"`
	Amount          float64 `json:"amount"`
	PaymentTimestamp string `json:"payment_timestamp"`
	TransactionHash *string `json:"transaction_hash,omitempty"`
	PaymentMethodID *int64  `json:"payment_method_id,omitempty"`
	MethodName      *string `json:"method_name,omitempty"`
	EmployeeID      *int64  `json:"employee_id,omitempty"`
	Note            *string `json:"note,omitempty"`
}

func (r *PaymentHistoryRepo) GetByInvoiceID(invoiceID int64) ([]PaymentRow, error) {
	var rows []PaymentRow
	err := r.DB.
		Table("payment_history ph").
		Select("ph.payment_id, ph.invoice_id, ph.patient_id, ph.amount, TO_CHAR(ph.payment_timestamp AT TIME ZONE 'UTC', 'YYYY-MM-DD\"T\"HH24:MI:SSZ') AS payment_timestamp, ph.transaction_hash, ph.payment_method_id, pm.method_name, ph.employee_id, ph.note").
		Joins("LEFT JOIN payment_method pm ON pm.id_payment_method = ph.payment_method_id").
		Where("ph.invoice_id = ?", invoiceID).
		Order("ph.payment_timestamp DESC").
		Scan(&rows).Error
	return rows, err
}

// GetByPatientID возвращает все кредитные платежи пациента.
func (r *PaymentHistoryRepo) GetByPatientID(patientID int64) ([]patients.PaymentHistory, error) {
	var rows []patients.PaymentHistory
	return rows, r.DB.Where("patient_id = ?", patientID).Order("payment_timestamp DESC").Find(&rows).Error
}

// Create записывает новый платёж.
func (r *PaymentHistoryRepo) Create(p *patients.PaymentHistory) error {
	return r.DB.Create(p).Error
}

// Update обновляет данные платежа (метод, сумма, хэш).
type UpdatePaymentInput struct {
	PaymentMethodID *int64
	Amount          *float64
	TransactionHash *string
}

func (r *PaymentHistoryRepo) Update(id int64, inp UpdatePaymentInput) error {
	updates := map[string]interface{}{}
	if inp.PaymentMethodID != nil { updates["payment_method_id"]  = *inp.PaymentMethodID }
	if inp.Amount != nil          { updates["amount"]             = *inp.Amount }
	if inp.TransactionHash != nil { updates["transaction_hash"]   = *inp.TransactionHash }
	if len(updates) == 0 {
		return nil
	}
	return r.DB.Model(&patients.PaymentHistory{}).Where("payment_id = ?", id).Updates(updates).Error
}

// Delete удаляет запись о платеже.
func (r *PaymentHistoryRepo) Delete(id int64) error {
	return r.DB.Delete(&patients.PaymentHistory{}, id).Error
}

// SumByInvoice считает суммарный оплаченный объём по инвойсу.
func (r *PaymentHistoryRepo) SumByInvoice(invoiceID int64) (float64, error) {
	var total float64
	err := r.DB.Model(&patients.PaymentHistory{}).
		Where("invoice_id = ?", invoiceID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

func (r *PaymentHistoryRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
