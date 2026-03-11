package reports_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/reports"
)

type ARCountRepo struct{ DB *gorm.DB }

func NewARCountRepo(db *gorm.DB) *ARCountRepo { return &ARCountRepo{DB: db} }

func (r *ARCountRepo) GetByID(id int) (*reports.ARCount, error) {
	var item reports.ARCount
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *ARCountRepo) GetByLocationID(locationID int) ([]reports.ARCount, error) {
	var items []reports.ARCount
	return items, r.DB.Where("location_id = ?", locationID).Order("created_date DESC").Find(&items).Error
}

func (r *ARCountRepo) GetOpen(locationID int) (*reports.ARCount, error) {
	var item reports.ARCount
	if err := r.DB.Where("location_id = ? AND status = true", locationID).
		Order("created_date DESC").First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *ARCountRepo) GetByDateRange(locationID int, from, to time.Time) ([]reports.ARCount, error) {
	var items []reports.ARCount
	return items, r.DB.Where("location_id = ? AND created_date BETWEEN ? AND ?", locationID, from, to).
		Order("created_date DESC").Find(&items).Error
}

func (r *ARCountRepo) Create(item *reports.ARCount) error {
	return r.DB.Create(item).Error
}

func (r *ARCountRepo) Save(item *reports.ARCount) error {
	return r.DB.Save(item).Error
}

func (r *ARCountRepo) Close(id int, employeeID int) error {
	return r.DB.Model(&reports.ARCount{}).Where("id_ar_count = ?", id).
		Updates(map[string]interface{}{
			"status":              false,
			"updated_employee_id": employeeID,
			"updated_date":        time.Now(),
		}).Error
}
