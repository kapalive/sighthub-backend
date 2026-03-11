package reports_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/reports"
)

type TempCountARRepo struct{ DB *gorm.DB }

func NewTempCountARRepo(db *gorm.DB) *TempCountARRepo { return &TempCountARRepo{DB: db} }

func (r *TempCountARRepo) GetByID(id int) (*reports.TempCountAR, error) {
	var item reports.TempCountAR
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *TempCountARRepo) GetByARCountID(arCountID int) ([]reports.TempCountAR, error) {
	var items []reports.TempCountAR
	return items, r.DB.Where("ar_count_id = ?", arCountID).Find(&items).Error
}

func (r *TempCountARRepo) GetByInvoiceAndARCount(invoiceID int64, arCountID int) (*reports.TempCountAR, error) {
	var item reports.TempCountAR
	if err := r.DB.Where("invoice_id = ? AND ar_count_id = ?", invoiceID, arCountID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *TempCountARRepo) Create(item *reports.TempCountAR) error {
	return r.DB.Create(item).Error
}

func (r *TempCountARRepo) Delete(id int) error {
	return r.DB.Delete(&reports.TempCountAR{}, id).Error
}

func (r *TempCountARRepo) DeleteByARCountID(arCountID int) error {
	return r.DB.Where("ar_count_id = ?", arCountID).Delete(&reports.TempCountAR{}).Error
}
