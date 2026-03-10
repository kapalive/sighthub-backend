package appointment

import "time"

type RequestAppointment struct {
	IDRequestAppointment       int64      `gorm:"column:id_request_appointment;primaryKey;autoIncrement" json:"id_request_appointment"`
	PatientID                  *int64     `gorm:"column:patient_id"                                      json:"patient_id,omitempty"`
	FirstName                  string     `gorm:"column:first_name;size:30;not null"                     json:"first_name"`
	LastName                   string     `gorm:"column:last_name;size:30;not null"                      json:"last_name"`
	Submitted                  *bool      `gorm:"column:submitted;default:false"                         json:"submitted,omitempty"`
	Processed                  *bool      `gorm:"column:processed;default:false"                         json:"processed,omitempty"`
	Accept                     *bool      `gorm:"column:accept"                                          json:"accept,omitempty"`
	Dob                        *time.Time `gorm:"column:dob;type:date"                                   json:"-"`
	Phone                      string     `gorm:"column:phone;size:18;not null"                          json:"phone"`
	Email                      *string    `gorm:"column:email;size:50"                                   json:"email,omitempty"`
	RequestingDate             time.Time  `gorm:"column:requesting_date;type:date;not null"              json:"-"`
	RequestingTime             time.Time  `gorm:"column:requesting_time;type:time;not null"              json:"-"`
	ProfessionalServiceTypeID  *int64     `gorm:"column:professional_service_type_id"                    json:"professional_service_type_id,omitempty"`
	InsuranceCompanyID         *int64     `gorm:"column:insurance_company_id"                            json:"insurance_company_id,omitempty"`
	InsurancePolicyID          *int64     `gorm:"column:insurance_policy_id"                             json:"insurance_policy_id,omitempty"`
	GroupNumber                *string    `gorm:"column:group_number;size:50"                            json:"group_number,omitempty"`
	MemberNumber               *string    `gorm:"column:member_number;size:50"                           json:"member_number,omitempty"`
	HolderType                 *string    `gorm:"column:holder_type;size:50"                             json:"holder_type,omitempty"`
	Note                       *string    `gorm:"column:note;type:text"                                  json:"note,omitempty"`
	DoctorID                   int64      `gorm:"column:doctor_id;not null"                              json:"doctor_id"`
}

func (RequestAppointment) TableName() string { return "request_appointment" }
