package schedule_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/schedule"
)

type CalendarRepo struct{ DB *gorm.DB }

func NewCalendarRepo(db *gorm.DB) *CalendarRepo { return &CalendarRepo{DB: db} }

func (r *CalendarRepo) GetByID(id int64) (*schedule.Calendar, error) {
	var item schedule.Calendar
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *CalendarRepo) GetByDate(date time.Time) ([]schedule.Calendar, error) {
	var items []schedule.Calendar
	return items, r.DB.Where("date = ?", date.Format("2006-01-02")).Find(&items).Error
}

func (r *CalendarRepo) GetByEmployeeAndDate(employeeID int64, date time.Time) (*schedule.Calendar, error) {
	var item schedule.Calendar
	if err := r.DB.Where("employee_id = ? AND date = ?", employeeID, date.Format("2006-01-02")).
		First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *CalendarRepo) GetByEmployeeAndDateRange(employeeID int64, from, to time.Time) ([]schedule.Calendar, error) {
	var items []schedule.Calendar
	return items, r.DB.Where("employee_id = ? AND date BETWEEN ? AND ?",
		employeeID, from.Format("2006-01-02"), to.Format("2006-01-02")).
		Order("date").Find(&items).Error
}

func (r *CalendarRepo) GetHolidaysByDateRange(from, to time.Time) ([]schedule.Calendar, error) {
	var items []schedule.Calendar
	return items, r.DB.Where("is_holiday = true AND date BETWEEN ? AND ?",
		from.Format("2006-01-02"), to.Format("2006-01-02")).
		Order("date").Find(&items).Error
}

func (r *CalendarRepo) Create(item *schedule.Calendar) error {
	return r.DB.Create(item).Error
}

func (r *CalendarRepo) Save(item *schedule.Calendar) error {
	return r.DB.Save(item).Error
}

func (r *CalendarRepo) Delete(id int64) error {
	return r.DB.Delete(&schedule.Calendar{}, id).Error
}
