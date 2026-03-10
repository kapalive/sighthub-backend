package patients

import "time"

type PatientNotes struct {
	IDPatientNotes int64      `gorm:"column:id_patient_notes;primaryKey;autoIncrement" json:"id_patient_notes"`
	Note           string     `gorm:"column:note;type:text;not null"                   json:"note"`
	Top            bool       `gorm:"column:top;default:false"                         json:"top"`
	AlertDate      *time.Time `gorm:"column:alert_date;type:date"                      json:"alert_date,omitempty"`
	PatientID      int64      `gorm:"column:patient_id;not null"                       json:"patient_id"`
}

func (PatientNotes) TableName() string { return "patient_notes" }
