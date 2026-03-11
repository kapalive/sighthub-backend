package schedule_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/schedule"
)

type ScheduleRepo struct{ DB *gorm.DB }

func NewScheduleRepo(db *gorm.DB) *ScheduleRepo { return &ScheduleRepo{DB: db} }

func (r *ScheduleRepo) GetByID(id int) (*schedule.Schedule, error) {
	var item schedule.Schedule
	if err := r.DB.Preload("Employee").First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *ScheduleRepo) GetByEmployeeID(employeeID int64) ([]schedule.Schedule, error) {
	var items []schedule.Schedule
	return items, r.DB.Where("employee_id = ?", employeeID).Find(&items).Error
}

func (r *ScheduleRepo) GetByEmployeeAndDay(employeeID int64, dayOfWeek string) (*schedule.Schedule, error) {
	var item schedule.Schedule
	if err := r.DB.Where("employee_id = ? AND day_of_week = ?", employeeID, dayOfWeek).
		First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *ScheduleRepo) Create(item *schedule.Schedule) error {
	return r.DB.Create(item).Error
}

func (r *ScheduleRepo) Save(item *schedule.Schedule) error {
	return r.DB.Save(item).Error
}

func (r *ScheduleRepo) Delete(id int) error {
	return r.DB.Delete(&schedule.Schedule{}, id).Error
}

func (r *ScheduleRepo) DeleteByEmployeeID(employeeID int64) error {
	return r.DB.Where("employee_id = ?", employeeID).Delete(&schedule.Schedule{}).Error
}
