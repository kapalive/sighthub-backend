// internal/repository/invoices_repo/shipping_customer.go
package invoices_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/invoices"
)

type ShippingCustomerRepo struct{ DB *gorm.DB }

func NewShippingCustomerRepo(db *gorm.DB) *ShippingCustomerRepo {
	return &ShippingCustomerRepo{DB: db}
}

func (r *ShippingCustomerRepo) GetByInvoiceID(invoiceID int64) (*invoices.ShippingCustomer, error) {
	var row invoices.ShippingCustomer
	err := r.DB.Where("invoice_id = ?", invoiceID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// Upsert создаёт или обновляет запись об отправке для инвойса.
func (r *ShippingCustomerRepo) Upsert(sc *invoices.ShippingCustomer) error {
	if sc.IDShippingCustomer == 0 {
		return r.DB.Create(sc).Error
	}
	return r.DB.Save(sc).Error
}

func (r *ShippingCustomerRepo) Delete(id int64) error {
	return r.DB.Delete(&invoices.ShippingCustomer{}, id).Error
}

func (r *ShippingCustomerRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
