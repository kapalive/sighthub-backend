package patients

import "time"

type PatientCommunicationHistory struct {
	IDPatientCommunicationHistory int64     `gorm:"column:id_patient_communication_history;primaryKey;autoIncrement" json:"id_patient_communication_history"`
	PatientID                     int64     `gorm:"column:patient_id;not null"                                       json:"patient_id"`
	CommunicationTypeID           int       `gorm:"column:communication_type_id;not null"                            json:"communication_type_id"`
	EmployeeID                    int64     `gorm:"column:employee_id;not null"                                      json:"employee_id"`
	LocationID                    int       `gorm:"column:location_id;not null"                                      json:"location_id"`
	Content                       string    `gorm:"column:content;type:text;not null"                                json:"content"`
	Description                   *string   `gorm:"column:description;type:text"                                     json:"description,omitempty"`
	CommunicationDatetime         time.Time `gorm:"column:communication_datetime;not null"                           json:"communication_datetime"`
}

func (PatientCommunicationHistory) TableName() string { return "patient_communication_history" }
