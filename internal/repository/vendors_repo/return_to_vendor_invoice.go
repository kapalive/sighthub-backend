// internal/repository/vendors_repo/return_to_vendor_invoice.go
package vendors_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/vendors"
)

type ReturnToVendorInvoiceRepo struct{ DB *gorm.DB }

func NewReturnToVendorInvoiceRepo(db *gorm.DB) *ReturnToVendorInvoiceRepo {
	return &ReturnToVendorInvoiceRepo{DB: db}
}

// GetAll возвращает все возвраты поставщику (с Vendor и Employee).
func (r *ReturnToVendorInvoiceRepo) GetAll(locationID *int64) ([]vendors.ReturnToVendorInvoice, error) {
	var rows []vendors.ReturnToVendorInvoice
	q := r.DB.Preload("Vendor").Preload("Employee").Preload("Items")
	if locationID != nil {
		// фильтруем через JOIN с позициями, если требуется привязка к локации
		// в модели нет location_id напрямую — возвращаем все, если locationID не задан
		_ = locationID
	}
	return rows, q.Order("created_date DESC").Find(&rows).Error
}

// GetByID возвращает возврат поставщику с позициями.
func (r *ReturnToVendorInvoiceRepo) GetByID(id int64) (*vendors.ReturnToVendorInvoice, error) {
	var row vendors.ReturnToVendorInvoice
	err := r.DB.
		Preload("Vendor").
		Preload("Employee").
		Preload("Items").
		First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// CreateInput — данные для нового возврата поставщику.
type CreateReturnToVendorInput struct {
	VendorID      int64
	EmployeeID    *int64
	CreditAmount  *float64
	Items         []ReturnToVendorItemInput
}

type ReturnToVendorItemInput struct {
	InventoryID  int64
	ReasonReturn string
	PurchaseCost float64
}

// Create создаёт возврат поставщику с позициями в транзакции.
func (r *ReturnToVendorInvoiceRepo) Create(inp CreateReturnToVendorInput) (*vendors.ReturnToVendorInvoice, error) {
	var total float64
	for _, item := range inp.Items {
		total += item.PurchaseCost
	}
	inv := &vendors.ReturnToVendorInvoice{
		VendorID:      inp.VendorID,
		EmployeeID:    inp.EmployeeID,
		CreditAmount:  inp.CreditAmount,
		PurchaseTotal: total,
		Quantity:      len(inp.Items),
	}
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(inv).Error; err != nil {
			return err
		}
		for _, it := range inp.Items {
			item := vendors.ReturnToVendorItem{
				ReturnToVendorInvoiceID: inv.IDReturnToVendorInvoice,
				InventoryID:             it.InventoryID,
				ReasonReturn:            it.ReasonReturn,
				PurchaseCost:            it.PurchaseCost,
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return inv, err
}

// UpdateInput — изменяемые поля.
type UpdateReturnToVendorInput struct {
	CreditAmount *float64
	EmployeeID   *int64
}

// Update обновляет возврат поставщику.
func (r *ReturnToVendorInvoiceRepo) Update(id int64, inp UpdateReturnToVendorInput) error {
	updates := map[string]interface{}{}
	if inp.CreditAmount != nil { updates["credit_amount"] = *inp.CreditAmount }
	if inp.EmployeeID != nil   { updates["employee_id"]   = *inp.EmployeeID }
	if len(updates) == 0 {
		return nil
	}
	return r.DB.Model(&vendors.ReturnToVendorInvoice{}).
		Where("id_return_to_vendor_invoice = ?", id).
		Updates(updates).Error
}

// Delete удаляет возврат и его позиции в транзакции.
func (r *ReturnToVendorInvoiceRepo) Delete(id int64) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("return_to_vendor_invoice_id = ?", id).Delete(&vendors.ReturnToVendorItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&vendors.ReturnToVendorInvoice{}, id).Error
	})
}

// AddPayment записывает платёж поставщику.
func (r *ReturnToVendorInvoiceRepo) AddPayment(p *vendors.PaymentToVendorTransaction) error {
	return r.DB.Create(p).Error
}

// GetPayments возвращает платежи поставщику по vendor_id + location_id.
func (r *ReturnToVendorInvoiceRepo) GetPayments(vendorID int, locationID int64) ([]vendors.PaymentToVendorTransaction, error) {
	var rows []vendors.PaymentToVendorTransaction
	return rows, r.DB.
		Where("vendor_id = ? AND location_id = ?", vendorID, locationID).
		Order("payment_date DESC").
		Find(&rows).Error
}

// DeletePayment удаляет платёж поставщику.
func (r *ReturnToVendorInvoiceRepo) DeletePayment(id int64) error {
	return r.DB.Delete(&vendors.PaymentToVendorTransaction{}, id).Error
}

func (r *ReturnToVendorInvoiceRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
