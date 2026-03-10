// internal/repository/invoices_repo/status_invoice.go
package invoices_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/invoices"
)

type StatusInvoiceRepo struct{ DB *gorm.DB }

func NewStatusInvoiceRepo(db *gorm.DB) *StatusInvoiceRepo { return &StatusInvoiceRepo{DB: db} }

func (r *StatusInvoiceRepo) GetAll() ([]invoices.StatusInvoice, error) {
	var rows []invoices.StatusInvoice
	return rows, r.DB.Find(&rows).Error
}

func (r *StatusInvoiceRepo) GetByID(id int) (*invoices.StatusInvoice, error) {
	var row invoices.StatusInvoice
	err := r.DB.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

func (r *StatusInvoiceRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
