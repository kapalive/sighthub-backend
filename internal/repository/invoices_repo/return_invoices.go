// internal/repository/invoices_repo/return_invoices.go
package invoices_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/invoices"
)

type ReturnInvoiceRepo struct{ DB *gorm.DB }

func NewReturnInvoiceRepo(db *gorm.DB) *ReturnInvoiceRepo { return &ReturnInvoiceRepo{DB: db} }

// GetByInvoiceID возвращает все возвраты для данного инвойса.
func (r *ReturnInvoiceRepo) GetByInvoiceID(invoiceID int64) ([]invoices.ReturnInvoice, error) {
	var rows []invoices.ReturnInvoice
	return rows, r.DB.Where("invoice_id = ?", invoiceID).Find(&rows).Error
}

// GetByID возвращает конкретный возврат.
func (r *ReturnInvoiceRepo) GetByID(id int64) (*invoices.ReturnInvoice, error) {
	var row invoices.ReturnInvoice
	err := r.DB.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// GetPatientReturns возвращает список возвратов по пациенту через JOIN с invoice.
type PatientReturnRow struct {
	ReturnID     int64     `json:"return_id"`
	InvoiceID    int64     `json:"invoice_id"`
	NumberInvoice string   `json:"number_invoice"`
	ReturnReason *string   `json:"return_reason"`
	ReturnedQty  int       `json:"returned_quantity"`
	ReturnAmount float64   `json:"return_amount"`
	ReturnDate   time.Time `json:"return_date"`
	Status       string    `json:"status"`
}

func (r *ReturnInvoiceRepo) GetPatientReturns(patientID int64, locationID int64) ([]PatientReturnRow, error) {
	var rows []PatientReturnRow
	err := r.DB.
		Table("return_invoices ri").
		Select("ri.return_id, ri.invoice_id, i.number_invoice, ri.return_reason, ri.returned_quantity, ri.return_amount, ri.return_date, ri.status").
		Joins("JOIN invoice i ON i.id_invoice = ri.invoice_id").
		Where("i.patient_id = ? AND i.location_id = ?", patientID, locationID).
		Order("ri.return_date DESC").
		Scan(&rows).Error
	return rows, err
}

// Create создаёт запись о возврате.
func (r *ReturnInvoiceRepo) Create(ret *invoices.ReturnInvoice) error {
	return r.DB.Create(ret).Error
}

// UpdateStatus обновляет статус возврата.
func (r *ReturnInvoiceRepo) UpdateStatus(id int64, status string) error {
	return r.DB.Model(&invoices.ReturnInvoice{}).
		Where("return_id = ?", id).
		Update("status", status).Error
}

// Delete удаляет возврат и связанные позиции в транзакции.
func (r *ReturnInvoiceRepo) Delete(id int64) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("return_id = ?", id).Delete(&invoices.ReturnItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&invoices.ReturnInvoice{}, id).Error
	})
}

func (r *ReturnInvoiceRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
