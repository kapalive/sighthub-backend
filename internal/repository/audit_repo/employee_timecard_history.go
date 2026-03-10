package audit_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/audit"
)

type EmployeeTimecardHistoryRepo struct{ DB *gorm.DB }

func NewEmployeeTimecardHistoryRepo(db *gorm.DB) *EmployeeTimecardHistoryRepo {
	return &EmployeeTimecardHistoryRepo{DB: db}
}

func (r *EmployeeTimecardHistoryRepo) GetByTimecardLoginID(loginID int) ([]audit.EmployeeTimecardHistory, error) {
	var items []audit.EmployeeTimecardHistory
	return items, r.DB.
		Where("employee_timecard_login_id = ?", loginID).
		Order("timestamp DESC").
		Find(&items).Error
}

func (r *EmployeeTimecardHistoryRepo) GetByDateRange(loginID int, from, to time.Time) ([]audit.EmployeeTimecardHistory, error) {
	var items []audit.EmployeeTimecardHistory
	return items, r.DB.
		Where("employee_timecard_login_id = ? AND timestamp BETWEEN ? AND ?", loginID, from, to).
		Order("timestamp").
		Find(&items).Error
}

func (r *EmployeeTimecardHistoryRepo) GetByID(id int) (*audit.EmployeeTimecardHistory, error) {
	var v audit.EmployeeTimecardHistory
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *EmployeeTimecardHistoryRepo) Create(v *audit.EmployeeTimecardHistory) error {
	return r.DB.Create(v).Error
}
