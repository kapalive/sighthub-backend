// internal/repository/inventory_repo/vendor_invoice.go
package inventory_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/inventory"
)

type VendorInvoiceRepo struct{ DB *gorm.DB }

func NewVendorInvoiceRepo(db *gorm.DB) *VendorInvoiceRepo {
	return &VendorInvoiceRepo{DB: db}
}

// GetByID возвращает vendor invoice по ID.
func (r *VendorInvoiceRepo) GetByID(id int64) (*inventory.VendorInvoice, error) {
	var row inventory.VendorInvoice
	err := r.DB.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// GetByInvoiceID возвращает vendor invoice по invoice_id (тот инвойс в системе).
func (r *VendorInvoiceRepo) GetByInvoiceID(invoiceID int64) (*inventory.VendorInvoice, error) {
	var row inventory.VendorInvoice
	err := r.DB.Where("invoice_id = ?", invoiceID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// GetByVendorID возвращает все vendor invoice для вендора.
func (r *VendorInvoiceRepo) GetByVendorID(vendorID int64) ([]inventory.VendorInvoice, error) {
	var rows []inventory.VendorInvoice
	return rows, r.DB.Where("vendor_id = ?", vendorID).Order("invoice_date DESC").Find(&rows).Error
}

// CreateVendorInvoiceInput — данные для нового vendor invoice.
type CreateVendorInvoiceInput struct {
	InvoiceNo        string
	InvoiceDate      *time.Time
	Quantity         int
	SubTotal         *float64
	ShippingHandling *float64
	Tax              *float64
	InvoiceTotal     *float64
	OrderRef         string
	DiscountReceived *int
	Note             *string
	VendorID         int64
	InvoiceID        int64
}

// Create создаёт vendor invoice.
func (r *VendorInvoiceRepo) Create(inp CreateVendorInvoiceInput) (*inventory.VendorInvoice, error) {
	vi := &inventory.VendorInvoice{
		InvoiceNo:        inp.InvoiceNo,
		InvoiceDate:      inp.InvoiceDate,
		Quantity:         inp.Quantity,
		SubTotal:         inp.SubTotal,
		ShippingHandling: inp.ShippingHandling,
		Tax:              inp.Tax,
		InvoiceTotal:     inp.InvoiceTotal,
		OrderRef:         inp.OrderRef,
		DiscountReceived: inp.DiscountReceived,
		Note:             inp.Note,
		VendorID:         inp.VendorID,
		InvoiceID:        inp.InvoiceID,
	}
	return vi, r.DB.Create(vi).Error
}

// UpdateVendorInvoiceInput — изменяемые поля.
type UpdateVendorInvoiceInput struct {
	InvoiceNo        *string
	InvoiceDate      *time.Time
	Quantity         *int
	SubTotal         *float64
	ShippingHandling *float64
	Tax              *float64
	InvoiceTotal     *float64
	OrderRef         *string
	DiscountReceived *int
	Note             *string
}

// Update обновляет vendor invoice.
func (r *VendorInvoiceRepo) Update(id int64, inp UpdateVendorInvoiceInput) error {
	updates := map[string]interface{}{}
	if inp.InvoiceNo != nil        { updates["invoice_no"]         = *inp.InvoiceNo }
	if inp.InvoiceDate != nil      { updates["invoice_date"]       = *inp.InvoiceDate }
	if inp.Quantity != nil         { updates["quantity"]           = *inp.Quantity }
	if inp.SubTotal != nil         { updates["sub_total"]          = *inp.SubTotal }
	if inp.ShippingHandling != nil { updates["shipping_handling"]  = *inp.ShippingHandling }
	if inp.Tax != nil              { updates["tax"]                = *inp.Tax }
	if inp.InvoiceTotal != nil     { updates["invoice_total"]      = *inp.InvoiceTotal }
	if inp.OrderRef != nil         { updates["order_ref"]          = *inp.OrderRef }
	if inp.DiscountReceived != nil { updates["discount_received"]  = *inp.DiscountReceived }
	if inp.Note != nil             { updates["note"]               = *inp.Note }
	if len(updates) == 0 {
		return nil
	}
	return r.DB.Model(&inventory.VendorInvoice{}).Where("id_vendor_invoice = ?", id).Updates(updates).Error
}

func (r *VendorInvoiceRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
