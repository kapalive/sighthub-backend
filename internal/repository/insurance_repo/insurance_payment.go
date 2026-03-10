package insurance_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/insurance"
)

type InsurancePaymentRepo struct{ DB *gorm.DB }

func NewInsurancePaymentRepo(db *gorm.DB) *InsurancePaymentRepo {
	return &InsurancePaymentRepo{DB: db}
}

func (r *InsurancePaymentRepo) GetByID(id int64) (*insurance.InsurancePayment, error) {
	var v insurance.InsurancePayment
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *InsurancePaymentRepo) GetByInvoiceID(invoiceID int64) ([]insurance.InsurancePayment, error) {
	var items []insurance.InsurancePayment
	return items, r.DB.Where("invoice_id = ?", invoiceID).Order("created_at").Find(&items).Error
}

func (r *InsurancePaymentRepo) GetByPolicyID(policyID int64) ([]insurance.InsurancePayment, error) {
	var items []insurance.InsurancePayment
	return items, r.DB.Where("insurance_policy_id = ?", policyID).Order("created_at DESC").Find(&items).Error
}

func (r *InsurancePaymentRepo) Create(v *insurance.InsurancePayment) error { return r.DB.Create(v).Error }
func (r *InsurancePaymentRepo) Save(v *insurance.InsurancePayment) error   { return r.DB.Save(v).Error }
func (r *InsurancePaymentRepo) Delete(id int64) error {
	return r.DB.Delete(&insurance.InsurancePayment{}, id).Error
}
