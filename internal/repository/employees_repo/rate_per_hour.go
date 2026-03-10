package employees_repo

import (
	"gorm.io/gorm"

	"sighthub-backend/internal/models/employees"
)

type RatePerHourRepo struct {
	DB *gorm.DB
}

func NewRatePerHourRepo(db *gorm.DB) *RatePerHourRepo {
	return &RatePerHourRepo{DB: db}
}

// GetByID returns the RatePerHour record with the given id.
func (r *RatePerHourRepo) GetByID(id int) (*employees.RatePerHour, error) {
	var rate employees.RatePerHour
	if err := r.DB.First(&rate, "id_rate_per_hour = ?", id).Error; err != nil {
		return nil, err
	}
	return &rate, nil
}

// Create inserts a new RatePerHour record and populates its primary key.
func (r *RatePerHourRepo) Create(rate *employees.RatePerHour) error {
	return r.DB.Create(rate).Error
}

// Update updates the rate_per_hour and hours_per_week fields for the record
// with the given id.
func (r *RatePerHourRepo) Update(id int, rateVal float64, hoursPerWeek float64) error {
	return r.DB.Model(&employees.RatePerHour{}).
		Where("id_rate_per_hour = ?", id).
		Updates(map[string]interface{}{
			"rate_per_hour":  rateVal,
			"hours_per_week": hoursPerWeek,
		}).Error
}
