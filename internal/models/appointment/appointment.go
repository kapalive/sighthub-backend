package appointment

import (
	"sighthub-backend/internal/models/insurance"
	"sighthub-backend/internal/models/location"
	"sighthub-backend/internal/models/patients"
	"sighthub-backend/internal/models/schedule"
	"time"
)

// Appointment ⇄ appointment
type Appointment struct {
	IDAppointment int64  `gorm:"column:id_appointment;primaryKey"              json:"id_appointment"`
	ScheduleID    *int64 `gorm:"column:schedule_id"                             json:"schedule_id,omitempty"`
	PatientID     int64  `gorm:"column:patient_id;not null"                     json:"patient_id"`
	LocationID    int    `gorm:"column:location_id;not null"                    json:"location_id"`

	// Отдельные date/time колонки
	AppointmentDate time.Time `gorm:"column:appointment_date;type:date;not null" json:"-"`
	StartTime       string `gorm:"column:start_time;type:time;not null"       json:"-"`
	EndTime         string `gorm:"column:end_time;type:time;not null"         json:"-"`

	StatusAppointmentID int     `gorm:"column:status_appointment_id;not null" json:"status_appointment_id"`
	Notes               *string `gorm:"column:notes;type:text"             json:"notes,omitempty"`

	InsurancePolicyID                  *int64 `gorm:"column:insurance_policy_id"                       json:"insurance_policy_id,omitempty"`
	ReasonsVisionProviderAppointmentID *int   `gorm:"column:reasons_vision_provider_appointment_id" json:"reasons_id,omitempty"`

	// ---------- Relations (preload по желанию) ----------
	Schedule          *schedule.Schedule                `gorm:"foreignKey:ScheduleID;references:IDSchedule"             json:"schedule,omitempty"`
	Patient           *patients.Patient                 `gorm:"foreignKey:PatientID;references:IDPatient"              json:"patient,omitempty"`
	Location          *location.Location                `gorm:"foreignKey:LocationID;references:IDLocation"             json:"location,omitempty"`
	StatusAppointment *StatusAppointment                `gorm:"foreignKey:StatusAppointmentID;references:IDStatusAppointment"    json:"status_appointment,omitempty"`
	InsurancePolicy   *insurance.InsurancePolicy        `gorm:"foreignKey:InsurancePolicyID;references:IDInsurancePolicy"      json:"insurance_policy,omitempty"`
	Reason            *ReasonsVisionProviderAppointment `gorm:"foreignKey:ReasonsVisionProviderAppointmentID;references:IDReasonsVisionProviderAppointment" json:"reason_obj,omitempty"`
}

func (Appointment) TableName() string { return "appointment" }

// ToMap — эквивалент Python to_dict()
func (a *Appointment) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_appointment":        a.IDAppointment,
		"schedule_id":           a.ScheduleID,
		"patient_id":            a.PatientID,
		"location_id":           a.LocationID,
		"status_appointment_id": a.StatusAppointmentID,
		"notes":                 a.Notes,
		"insurance_policy_id":   a.InsurancePolicyID,
		"reasons_id":            a.ReasonsVisionProviderAppointmentID,
	}

	// date/time формат как в Python: YYYY-MM-DD и HH:MM:SS
	if !a.AppointmentDate.IsZero() {
		m["appointment_date"] = a.AppointmentDate.Format("2006-01-02")
	} else {
		m["appointment_date"] = nil
	}
	if a.StartTime != "" {
		m["start_time"] = a.StartTime
	} else {
		m["start_time"] = nil
	}
	if a.EndTime != "" {
		m["end_time"] = a.EndTime
	} else {
		m["end_time"] = nil
	}

	// вложенные объекты — как в Python
	if a.StatusAppointment != nil {
		if tm, ok := any(a.StatusAppointment).(interface{ ToMap() map[string]interface{} }); ok {
			m["status_appointment"] = tm.ToMap()
		} else {
			m["status_appointment"] = a.StatusAppointment
		}
	}
	if a.Schedule != nil {
		if tm, ok := any(a.Schedule).(interface{ ToMap() map[string]interface{} }); ok {
			m["schedule"] = tm.ToMap()
		} else {
			m["schedule"] = a.Schedule
		}
	}
	if a.Patient != nil {
		if tm, ok := any(a.Patient).(interface{ ToMap() map[string]interface{} }); ok {
			m["patient"] = tm.ToMap()
		} else {
			m["patient"] = a.Patient
		}
	}
	if a.Location != nil {
		if tm, ok := any(a.Location).(interface{ ToMap() map[string]interface{} }); ok {
			m["location"] = tm.ToMap()
		} else {
			m["location"] = a.Location
		}
	}
	if a.InsurancePolicy != nil {
		if tm, ok := any(a.InsurancePolicy).(interface{ ToMap() map[string]interface{} }); ok {
			m["insurance_policy"] = tm.ToMap()
		} else {
			m["insurance_policy"] = a.InsurancePolicy
		}
	}
	// в Python: 'reason': self.reason.reason
	if a.Reason != nil {
		m["reason"] = a.Reason.Reason // поле Reason string в модельке причины
	} else {
		m["reason"] = nil
	}

	return m
}

// ------ Подсказки -------
// 1) Если PK у связанных сущностей НЕ 'ID', добавь references:…
//    напр.: Schedule *Schedule `gorm:"foreignKey:ScheduleID;references:IDSchedule"`
// 2) Якщо хочешь null в JSON вместо "", оставляй *string/*int64 и omitempty как выше.
// 3) В БД для времени/даты уже указаны типы через gorm:"type:date|time".
