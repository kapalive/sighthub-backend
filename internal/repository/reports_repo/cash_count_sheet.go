package reports_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/reports"
)

type CashCountSheetRepo struct{ DB *gorm.DB }

func NewCashCountSheetRepo(db *gorm.DB) *CashCountSheetRepo {
	return &CashCountSheetRepo{DB: db}
}

func (r *CashCountSheetRepo) GetByID(id int64) (*reports.CashCountSheet, error) {
	var item reports.CashCountSheet
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *CashCountSheetRepo) GetByDailyClosePaymentID(dailyClosePaymentID int64) (*reports.CashCountSheet, error) {
	var item reports.CashCountSheet
	if err := r.DB.Where("daily_close_payment_id = ?", dailyClosePaymentID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *CashCountSheetRepo) GetByDate(date time.Time) ([]reports.CashCountSheet, error) {
	var items []reports.CashCountSheet
	return items, r.DB.Where("date = ?", date.Format("2006-01-02")).Find(&items).Error
}

func (r *CashCountSheetRepo) Create(item *reports.CashCountSheet) error {
	return r.DB.Create(item).Error
}

func (r *CashCountSheetRepo) Save(item *reports.CashCountSheet) error {
	return r.DB.Save(item).Error
}

func (r *CashCountSheetRepo) Delete(id int64) error {
	return r.DB.Delete(&reports.CashCountSheet{}, id).Error
}
