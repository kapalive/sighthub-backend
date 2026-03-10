package patients_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/patients"
)

type TransferCreditRepo struct{ DB *gorm.DB }

func NewTransferCreditRepo(db *gorm.DB) *TransferCreditRepo {
	return &TransferCreditRepo{DB: db}
}

func (r *TransferCreditRepo) GetByID(id int64) (*patients.TransferCredit, error) {
	var item patients.TransferCredit
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *TransferCreditRepo) GetByInvoiceID(invoiceID int64) ([]patients.TransferCredit, error) {
	var items []patients.TransferCredit
	if err := r.DB.Where("invoice_id = ?", invoiceID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *TransferCreditRepo) GetByPatientID(patientID int64) ([]patients.TransferCredit, error) {
	var items []patients.TransferCredit
	if err := r.DB.Where("patient_id = ?", patientID).
		Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *TransferCreditRepo) Create(item *patients.TransferCredit) error {
	return r.DB.Create(item).Error
}

func (r *TransferCreditRepo) Delete(id int64) error {
	return r.DB.Delete(&patients.TransferCredit{}, id).Error
}
