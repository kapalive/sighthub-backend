package appointment_repo

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/appointment"
)

type StatusAppointmentRepo struct{ DB *gorm.DB }

func NewStatusAppointmentRepo(db *gorm.DB) *StatusAppointmentRepo {
	return &StatusAppointmentRepo{DB: db}
}

func (r *StatusAppointmentRepo) GetAll() ([]appointment.StatusAppointment, error) {
	var items []appointment.StatusAppointment
	return items, r.DB.Find(&items).Error
}

func (r *StatusAppointmentRepo) GetByID(id int) (*appointment.StatusAppointment, error) {
	var v appointment.StatusAppointment
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}
