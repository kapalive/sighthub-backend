package appointment_repo

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/appointment"
)

type ReasonsVisionProviderAppointmentRepo struct{ DB *gorm.DB }

func NewReasonsVisionProviderAppointmentRepo(db *gorm.DB) *ReasonsVisionProviderAppointmentRepo {
	return &ReasonsVisionProviderAppointmentRepo{DB: db}
}

func (r *ReasonsVisionProviderAppointmentRepo) GetAll() ([]appointment.ReasonsVisionProviderAppointment, error) {
	var items []appointment.ReasonsVisionProviderAppointment
	return items, r.DB.Find(&items).Error
}

func (r *ReasonsVisionProviderAppointmentRepo) GetByID(id int) (*appointment.ReasonsVisionProviderAppointment, error) {
	var v appointment.ReasonsVisionProviderAppointment
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *ReasonsVisionProviderAppointmentRepo) Create(v *appointment.ReasonsVisionProviderAppointment) error {
	return r.DB.Create(v).Error
}

func (r *ReasonsVisionProviderAppointmentRepo) Save(v *appointment.ReasonsVisionProviderAppointment) error {
	return r.DB.Save(v).Error
}

func (r *ReasonsVisionProviderAppointmentRepo) Delete(id int) error {
	return r.DB.Delete(&appointment.ReasonsVisionProviderAppointment{}, id).Error
}
