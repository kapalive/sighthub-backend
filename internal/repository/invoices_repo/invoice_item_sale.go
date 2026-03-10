// internal/repository/invoices_repo/invoice_item_sale.go
package invoices_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/invoices"
)

type InvoiceItemSaleRepo struct{ DB *gorm.DB }

func NewInvoiceItemSaleRepo(db *gorm.DB) *InvoiceItemSaleRepo {
	return &InvoiceItemSaleRepo{DB: db}
}

func (r *InvoiceItemSaleRepo) GetByInvoiceID(invoiceID int64) ([]invoices.InvoiceItemSale, error) {
	var rows []invoices.InvoiceItemSale
	return rows, r.DB.Where("invoice_id = ?", invoiceID).Find(&rows).Error
}

func (r *InvoiceItemSaleRepo) GetByID(id int64) (*invoices.InvoiceItemSale, error) {
	var row invoices.InvoiceItemSale
	err := r.DB.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

func (r *InvoiceItemSaleRepo) Create(item *invoices.InvoiceItemSale) error {
	return r.DB.Create(item).Error
}

// UpdateItem обновляет изменяемые поля позиции (цена, скидка, кол-во, описание, taxable).
type UpdateItemInput struct {
	Price       *float64
	Discount    *float64
	Quantity    *int
	Description *string
	Taxable     *bool
}

func (r *InvoiceItemSaleRepo) UpdateItem(id int64, inp UpdateItemInput) error {
	updates := map[string]interface{}{}
	if inp.Price != nil {
		updates["price"] = *inp.Price
	}
	if inp.Discount != nil {
		updates["discount"] = *inp.Discount
	}
	if inp.Quantity != nil {
		updates["quantity"] = *inp.Quantity
	}
	if inp.Description != nil {
		updates["description"] = *inp.Description
	}
	if inp.Taxable != nil {
		updates["taxable"] = *inp.Taxable
	}
	if len(updates) == 0 {
		return nil
	}
	return r.DB.Model(&invoices.InvoiceItemSale{}).Where("id_invoice_sale = ?", id).Updates(updates).Error
}

// SetLineBalance обновляет ins_balance у позиции и пересчитывает pt_bal/ins_bal в инвойсе.
// Вся логика выполняется в транзакции.
func (r *InvoiceItemSaleRepo) SetLineBalance(tx *gorm.DB, itemSaleID int64, insBalance float64) error {
	return tx.Model(&invoices.InvoiceItemSale{}).
		Where("id_invoice_sale = ?", itemSaleID).
		Update("ins_balance", insBalance).Error
}

func (r *InvoiceItemSaleRepo) Delete(id int64) error {
	return r.DB.Delete(&invoices.InvoiceItemSale{}, id).Error
}

func (r *InvoiceItemSaleRepo) DeleteByInvoiceID(invoiceID int64) error {
	return r.DB.Where("invoice_id = ?", invoiceID).Delete(&invoices.InvoiceItemSale{}).Error
}

// SumTotals возвращает агрегированные суммы по инвойсу: total, discount, tax.
type InvoiceSaleTotals struct {
	TotalAmount float64
	Discount    float64
	TaxAmount   float64
	Quantity    int
}

func (r *InvoiceItemSaleRepo) SumTotals(invoiceID int64) (InvoiceSaleTotals, error) {
	var res InvoiceSaleTotals
	err := r.DB.Model(&invoices.InvoiceItemSale{}).
		Select("COALESCE(SUM(total),0) AS total_amount, COALESCE(SUM(discount),0) AS discount, COALESCE(SUM(total_tax),0) AS tax_amount, COALESCE(SUM(quantity),0) AS quantity").
		Where("invoice_id = ?", invoiceID).
		Scan(&res).Error
	return res, err
}

func (r *InvoiceItemSaleRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
