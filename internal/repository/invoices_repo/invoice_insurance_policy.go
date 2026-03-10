// internal/repository/invoices_repo/invoice_insurance_policy.go
package invoices_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/insurance"
	"sighthub-backend/internal/models/invoices"
)

type InvoiceInsurancePolicyRepo struct{ DB *gorm.DB }

func NewInvoiceInsurancePolicyRepo(db *gorm.DB) *InvoiceInsurancePolicyRepo {
	return &InvoiceInsurancePolicyRepo{DB: db}
}

// GetByInvoiceID возвращает запись связи инвойса со страховым полисом.
func (r *InvoiceInsurancePolicyRepo) GetByInvoiceID(invoiceID int64) (*invoices.InvoiceInsurancePolicy, error) {
	var row invoices.InvoiceInsurancePolicy
	err := r.DB.Where("invoice_id = ?", invoiceID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// Add привязывает страховой полис к инвойсу и обновляет insurance_policy_id в самом инвойсе.
func (r *InvoiceInsurancePolicyRepo) Add(invoiceID, policyID int64) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// удалить предыдущую привязку, если была
		tx.Where("invoice_id = ?", invoiceID).Delete(&invoices.InvoiceInsurancePolicy{})

		link := invoices.InvoiceInsurancePolicy{InvoiceID: invoiceID, InsurancePolicyID: policyID}
		if err := tx.Create(&link).Error; err != nil {
			return err
		}
		return tx.Model(&invoices.Invoice{}).
			Where("id_invoice = ?", invoiceID).
			Update("insurance_policy_id", policyID).Error
	})
}

// Remove отвязывает страховку от инвойса.
func (r *InvoiceInsurancePolicyRepo) Remove(invoiceID int64) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("invoice_id = ?", invoiceID).Delete(&invoices.InvoiceInsurancePolicy{}).Error; err != nil {
			return err
		}
		return tx.Model(&invoices.Invoice{}).
			Where("id_invoice = ?", invoiceID).
			Update("insurance_policy_id", nil).Error
	})
}

// --- InsurancePayment operations ---

// GetInsurancePayments возвращает все страховые платежи по инвойсу.
func (r *InvoiceInsurancePolicyRepo) GetInsurancePayments(invoiceID int64) ([]insurance.InsurancePayment, error) {
	var rows []insurance.InsurancePayment
	return rows, r.DB.Where("invoice_id = ?", invoiceID).Find(&rows).Error
}

// AddInsurancePayment создаёт запись страхового платежа.
func (r *InvoiceInsurancePolicyRepo) AddInsurancePayment(p *insurance.InsurancePayment) error {
	return r.DB.Create(p).Error
}

// DeleteInsurancePayment удаляет страховой платёж.
func (r *InvoiceInsurancePolicyRepo) DeleteInsurancePayment(paymentID int64) error {
	return r.DB.Delete(&insurance.InsurancePayment{}, paymentID).Error
}

// GetInsurancePaymentTypes возвращает справочник типов страховых платежей.
func (r *InvoiceInsurancePolicyRepo) GetInsurancePaymentTypes() ([]insurance.InsurancePaymentType, error) {
	var rows []insurance.InsurancePaymentType
	return rows, r.DB.Where("active = true").Find(&rows).Error
}

func (r *InvoiceInsurancePolicyRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
