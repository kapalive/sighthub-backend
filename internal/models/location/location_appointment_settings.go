// internal/models/location/location_appointment_settings.go
package location

// LocationAppointmentSettings ⇄ table: location_appointment_settings
type LocationAppointmentSettings struct {
	IDLocationAppointmentSettings int  `gorm:"column:id_location_appointment_settings;primaryKey;autoIncrement" json:"id_location_appointment_settings"`
	LocationID                    int  `gorm:"column:location_id;not null;uniqueIndex"                          json:"location_id"`
	RequestAppointmentEnabled     bool `gorm:"column:request_appointment_enabled;not null;default:false"        json:"request_appointment_enabled"`
	IntakeFormEnabled             bool `gorm:"column:intake_form_enabled;not null;default:false"                json:"intake_form_enabled"`
	AppointmentDuration           *int `gorm:"column:appointment_duration"                                      json:"appointment_duration,omitempty"`
}

func (LocationAppointmentSettings) TableName() string { return "location_appointment_settings" }
