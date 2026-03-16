package appointment_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/appointment"
)

type AppointmentRepo struct{ DB *gorm.DB }

func NewAppointmentRepo(db *gorm.DB) *AppointmentRepo {
	return &AppointmentRepo{DB: db}
}

func (r *AppointmentRepo) GetByID(id int64) (*appointment.Appointment, error) {
	var v appointment.Appointment
	if err := r.DB.
		Preload("StatusAppointment").
		Preload("Schedule").
		Preload("Patient").
		Preload("Location").
		Preload("InsurancePolicy").
		Preload("Reason").
		First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *AppointmentRepo) GetByPatientID(patientID int64) ([]appointment.Appointment, error) {
	var items []appointment.Appointment
	return items, r.DB.
		Preload("StatusAppointment").
		Preload("Reason").
		Where("patient_id = ?", patientID).
		Order("appointment_date DESC, start_time DESC").
		Find(&items).Error
}

func (r *AppointmentRepo) GetByLocationAndDate(locationID int, date time.Time) ([]appointment.Appointment, error) {
	var items []appointment.Appointment
	return items, r.DB.
		Preload("StatusAppointment").
		Preload("Patient").
		Preload("Reason").
		Where("location_id = ? AND appointment_date = ?", locationID, date.Format("2006-01-02")).
		Order("start_time").
		Find(&items).Error
}

func (r *AppointmentRepo) GetByDateRange(locationID int, from, to time.Time) ([]appointment.Appointment, error) {
	var items []appointment.Appointment
	return items, r.DB.
		Preload("StatusAppointment").
		Preload("Patient").
		Preload("Reason").
		Where("location_id = ? AND appointment_date BETWEEN ? AND ?",
			locationID, from.Format("2006-01-02"), to.Format("2006-01-02")).
		Order("appointment_date, start_time").
		Find(&items).Error
}

func (r *AppointmentRepo) GetByScheduleID(scheduleID int64) ([]appointment.Appointment, error) {
	var items []appointment.Appointment
	return items, r.DB.
		Preload("StatusAppointment").
		Preload("Patient").
		Where("schedule_id = ?", scheduleID).
		Order("appointment_date, start_time").
		Find(&items).Error
}

type CreateAppointmentInput struct {
	ScheduleID                         *int64
	PatientID                          int64
	LocationID                         int
	AppointmentDate                    time.Time
	StartTime                          string
	EndTime                            string
	StatusAppointmentID                int
	Notes                              *string
	InsurancePolicyID                  *int64
	ReasonsVisionProviderAppointmentID *int
}

func (r *AppointmentRepo) Create(inp CreateAppointmentInput) (*appointment.Appointment, error) {
	v := appointment.Appointment{
		ScheduleID:                         inp.ScheduleID,
		PatientID:                          inp.PatientID,
		LocationID:                         inp.LocationID,
		AppointmentDate:                    inp.AppointmentDate,
		StartTime:                          inp.StartTime,
		EndTime:                            inp.EndTime,
		StatusAppointmentID:                inp.StatusAppointmentID,
		Notes:                              inp.Notes,
		InsurancePolicyID:                  inp.InsurancePolicyID,
		ReasonsVisionProviderAppointmentID: inp.ReasonsVisionProviderAppointmentID,
	}
	if err := r.DB.Create(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *AppointmentRepo) Save(v *appointment.Appointment) error {
	return r.DB.Save(v).Error
}

func (r *AppointmentRepo) UpdateStatus(id int64, statusID int) error {
	return r.DB.Model(&appointment.Appointment{}).
		Where("id_appointment = ?", id).
		Update("status_appointment_id", statusID).Error
}

func (r *AppointmentRepo) Delete(id int64) error {
	return r.DB.Delete(&appointment.Appointment{}, id).Error
}
