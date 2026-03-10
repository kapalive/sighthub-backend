// internal/repository/location_repo/location_appointment_settings.go
package location_repo

import (
	"gorm.io/gorm"
	locmodel "sighthub-backend/internal/models/location"
)

// LocationAppointmentSettingsRepo — репозиторий для location_appointment_settings.
type LocationAppointmentSettingsRepo struct {
	DB *gorm.DB
}

func NewLocationAppointmentSettingsRepo(db *gorm.DB) *LocationAppointmentSettingsRepo {
	return &LocationAppointmentSettingsRepo{DB: db}
}

// GetByLocationID возвращает настройки для конкретной локации.
func (r *LocationAppointmentSettingsRepo) GetByLocationID(locationID int) (*locmodel.LocationAppointmentSettings, error) {
	var s locmodel.LocationAppointmentSettings
	err := r.DB.Where("location_id = ?", locationID).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetByStoreID ищет настройки через join с location (берёт основную локацию стора).
func (r *LocationAppointmentSettingsRepo) GetByStoreID(storeID int) (*locmodel.LocationAppointmentSettings, error) {
	var s locmodel.LocationAppointmentSettings
	err := r.DB.
		Joins("JOIN location ON location.id_location = location_appointment_settings.location_id").
		Where("location.store_id = ? AND location.warehouse_id IS NULL", storeID).
		First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Upsert создаёт или обновляет настройки.
func (r *LocationAppointmentSettingsRepo) Upsert(s *locmodel.LocationAppointmentSettings) error {
	return r.DB.Save(s).Error
}
