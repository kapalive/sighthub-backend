package appointment_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/appointment"
)

type RequestAppointmentRepo struct{ DB *gorm.DB }

func NewRequestAppointmentRepo(db *gorm.DB) *RequestAppointmentRepo {
	return &RequestAppointmentRepo{DB: db}
}

func (r *RequestAppointmentRepo) GetByID(id int64) (*appointment.RequestAppointment, error) {
	var v appointment.RequestAppointment
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *RequestAppointmentRepo) GetPending() ([]appointment.RequestAppointment, error) {
	var items []appointment.RequestAppointment
	return items, r.DB.Where("processed = false").Order("requesting_date, requesting_time").Find(&items).Error
}

func (r *RequestAppointmentRepo) GetByDoctorID(doctorID int64) ([]appointment.RequestAppointment, error) {
	var items []appointment.RequestAppointment
	return items, r.DB.Where("doctor_id = ?", doctorID).Order("requesting_date DESC").Find(&items).Error
}

func (r *RequestAppointmentRepo) GetByDateRange(from, to time.Time) ([]appointment.RequestAppointment, error) {
	var items []appointment.RequestAppointment
	return items, r.DB.
		Where("requesting_date BETWEEN ? AND ?", from.Format("2006-01-02"), to.Format("2006-01-02")).
		Order("requesting_date, requesting_time").
		Find(&items).Error
}

func (r *RequestAppointmentRepo) Create(v *appointment.RequestAppointment) error {
	return r.DB.Create(v).Error
}

func (r *RequestAppointmentRepo) Save(v *appointment.RequestAppointment) error {
	return r.DB.Save(v).Error
}

func (r *RequestAppointmentRepo) SetProcessed(id int64, accept bool) error {
	return r.DB.Model(&appointment.RequestAppointment{}).
		Where("id_request_appointment = ?", id).
		Updates(map[string]interface{}{"processed": true, "accept": accept}).Error
}

func (r *RequestAppointmentRepo) Delete(id int64) error {
	return r.DB.Delete(&appointment.RequestAppointment{}, id).Error
}
