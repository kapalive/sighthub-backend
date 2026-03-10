// internal/repository/invoices_repo/invoice_services_item.go
package invoices_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/invoices"
)

type InvoiceServicesItemRepo struct{ DB *gorm.DB }

func NewInvoiceServicesItemRepo(db *gorm.DB) *InvoiceServicesItemRepo {
	return &InvoiceServicesItemRepo{DB: db}
}

func (r *InvoiceServicesItemRepo) GetByInvoiceID(invoiceID int64) ([]invoices.InvoiceServicesItem, error) {
	var rows []invoices.InvoiceServicesItem
	return rows, r.DB.Where("invoice_id = ?", invoiceID).Find(&rows).Error
}

func (r *InvoiceServicesItemRepo) Create(item *invoices.InvoiceServicesItem) error {
	return r.DB.Create(item).Error
}

func (r *InvoiceServicesItemRepo) Delete(id int64) error {
	return r.DB.Delete(&invoices.InvoiceServicesItem{}, id).Error
}

func (r *InvoiceServicesItemRepo) DeleteByInvoiceID(invoiceID int64) error {
	return r.DB.Where("invoice_id = ?", invoiceID).Delete(&invoices.InvoiceServicesItem{}).Error
}

func (r *InvoiceServicesItemRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
