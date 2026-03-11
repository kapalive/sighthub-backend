package reports_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/reports"
)

type DailyClosePaymentRepo struct{ DB *gorm.DB }

func NewDailyClosePaymentRepo(db *gorm.DB) *DailyClosePaymentRepo {
	return &DailyClosePaymentRepo{DB: db}
}

func (r *DailyClosePaymentRepo) GetByID(id int64) (*reports.DailyClosePayment, error) {
	var item reports.DailyClosePayment
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *DailyClosePaymentRepo) GetByLocationAndDate(locationID int64, date time.Time) ([]reports.DailyClosePayment, error) {
	var items []reports.DailyClosePayment
	return items, r.DB.Where("location_id = ? AND date = ?", locationID, date.Format("2006-01-02")).Find(&items).Error
}

func (r *DailyClosePaymentRepo) GetByDateRange(locationID int64, from, to time.Time) ([]reports.DailyClosePayment, error) {
	var items []reports.DailyClosePayment
	return items, r.DB.Where("location_id = ? AND date BETWEEN ? AND ?",
		locationID, from.Format("2006-01-02"), to.Format("2006-01-02")).
		Order("date DESC").Find(&items).Error
}

func (r *DailyClosePaymentRepo) Create(item *reports.DailyClosePayment) error {
	return r.DB.Create(item).Error
}

func (r *DailyClosePaymentRepo) Save(item *reports.DailyClosePayment) error {
	return r.DB.Save(item).Error
}

func (r *DailyClosePaymentRepo) Delete(id int64) error {
	return r.DB.Delete(&reports.DailyClosePayment{}, id).Error
}

func (r *DailyClosePaymentRepo) DeleteByLocationAndDate(locationID int64, date time.Time) error {
	return r.DB.Where("location_id = ? AND date = ?", locationID, date.Format("2006-01-02")).
		Delete(&reports.DailyClosePayment{}).Error
}
