package appointment

import "fmt"

type StatusAppointment struct {
	IDStatusAppointment int    `gorm:"column:id_status_appointment;primaryKey"        json:"id_status_appointment"`
	StatusAppointment   string `gorm:"column:status_appointment;type:varchar(30);not null" json:"status_appointment"`
}

func (StatusAppointment) TableName() string { return "status_appointment" }

func (s *StatusAppointment) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_status_appointment": s.IDStatusAppointment,
		"status_appointment":    s.StatusAppointment,
	}
}

func (s *StatusAppointment) String() string {
	return fmt.Sprintf("<StatusAppointment %d %s>", s.IDStatusAppointment, s.StatusAppointment)
}
