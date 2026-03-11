package reports_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/reports"
)

type MissingARRepo struct{ DB *gorm.DB }

func NewMissingARRepo(db *gorm.DB) *MissingARRepo { return &MissingARRepo{DB: db} }

func (r *MissingARRepo) GetByID(id int) (*reports.MissingAR, error) {
	var item reports.MissingAR
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *MissingARRepo) GetByARCountID(arCountID int) ([]reports.MissingAR, error) {
	var items []reports.MissingAR
	return items, r.DB.Where("ar_count_id = ?", arCountID).Find(&items).Error
}

func (r *MissingARRepo) GetByLocationID(locationID int) ([]reports.MissingAR, error) {
	var items []reports.MissingAR
	return items, r.DB.Where("location_id = ?", locationID).Order("reported_date DESC").Find(&items).Error
}

func (r *MissingARRepo) Create(item *reports.MissingAR) error {
	return r.DB.Create(item).Error
}

func (r *MissingARRepo) Delete(id int) error {
	return r.DB.Delete(&reports.MissingAR{}, id).Error
}

func (r *MissingARRepo) DeleteByARCountAndInvoice(arCountID int, invoiceID int64) error {
	return r.DB.Where("ar_count_id = ? AND invoice_id = ?", arCountID, invoiceID).
		Delete(&reports.MissingAR{}).Error
}
