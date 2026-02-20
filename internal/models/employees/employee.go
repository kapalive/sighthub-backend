package employees

import (
	"time"
)

// Employee ⇄ employee
type Employee struct {
	IDEmployee      int        `gorm:"column:id_employee;primaryKey;autoIncrement"   json:"id_employee"`
	FirstName       string     `gorm:"column:first_name;type:varchar(20);not null"   json:"first_name"`
	MiddleName      *string    `gorm:"column:middle_name;type:varchar(20)"          json:"middle_name,omitempty"`
	LastName        string     `gorm:"column:last_name;type:varchar(20);not null"   json:"last_name"`
	Suffix          *string    `gorm:"column:suffix;type:varchar(4)"                json:"suffix,omitempty"`
	DOB             *time.Time `gorm:"column:dob;type:date"                       json:"-"`
	Phone           *string    `gorm:"column:phone;type:varchar(10)"               json:"phone,omitempty"`
	Email           *string    `gorm:"column:email;type:varchar(100)"              json:"email,omitempty"`
	StreetAddress   *string    `gorm:"column:street_address;type:varchar(100)"    json:"street_address,omitempty"`
	AddressLine2    *string    `gorm:"column:address_line_2;type:varchar(100)"    json:"address_line_2,omitempty"`
	City            *string    `gorm:"column:city;type:varchar(100)"              json:"city,omitempty"`
	State           *string    `gorm:"column:state;type:varchar(100)"             json:"state,omitempty"`
	SSN             *string    `gorm:"column:ssn;type:varchar(11)"                json:"ssn,omitempty"`
	StartDate       *time.Time `gorm:"column:start_date;type:date"             json:"-"`
	TerminationDate *time.Time `gorm:"column:termination_date;type:date"     json:"-"`
	Active          bool       `gorm:"column:active;not null;default:true"       json:"active"`
	Signature       *string    `gorm:"column:signature;type:varchar(255)"        json:"signature,omitempty"`
	Zip             *string    `gorm:"column:zip;type:varchar(20)"               json:"zip,omitempty"`
	Country         *string    `gorm:"column:country;type:varchar(50)"           json:"country,omitempty"`
	Prefix          *string    `gorm:"column:prefix;type:varchar(10)"            json:"prefix,omitempty"`

	EmployeeLoginID int64  `gorm:"column:employee_login_id;not null"         json:"employee_login_id"`
	JobTitleID      *int64 `gorm:"column:job_title_id"                        json:"job_title_id,omitempty"`
	WorkShiftID     *int64 `gorm:"column:work_shift_id"                       json:"work_shift_id,omitempty"`
	LocationID      *int64 `gorm:"column:location_id"                         json:"location_id,omitempty"`
	StorePayrollID  *int64 `gorm:"column:store_payroll_id"                    json:"store_payroll_id,omitempty"`
	RatePerHourID   *int64 `gorm:"column:rate_per_hour_id"                    json:"rate_per_hour_id,omitempty"`
	PayrollTypeID   *int64 `gorm:"column:payroll_type_id"                     json:"payroll_type_id,omitempty"`
}

func (Employee) TableName() string { return "employee" }

// ToMap — аналог Python to_dict()
func (e *Employee) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_employee":    e.IDEmployee,
		"first_name":     e.FirstName,
		"middle_name":    e.MiddleName,
		"last_name":      e.LastName,
		"suffix":         e.Suffix,
		"prefix":         e.Prefix,
		"phone":          e.Phone,
		"email":          e.Email,
		"street_address": e.StreetAddress,
		"address_line_2": e.AddressLine2,
		"city":           e.City,
		"state":          e.State,
		"zip":            e.Zip,
		"country":        e.Country,
		"ssn":            e.SSN,
		"signature":      e.Signature,
		"active":         e.Active,

		"location_id": e.LocationID,
		// как в Python: если 0 — вернуть nil
		"employee_login_id": func() interface{} {
			if e.EmployeeLoginID > 0 {
				return e.EmployeeLoginID
			}
			return nil
		}(),

		// вложенные объекты оставляем nil (одна модель в файле)
		"job_title":      nil,
		"work_shift":     nil,
		"assigned_store": nil,
		"payroll_store":  nil,
		"rate_per_hour":  nil,
		"payroll_type":   nil,
	}

	if e.DOB != nil && !e.DOB.IsZero() {
		m["dob"] = e.DOB.Format("2006-01-02")
	} else {
		m["dob"] = nil
	}
	if e.StartDate != nil && !e.StartDate.IsZero() {
		m["start_date"] = e.StartDate.Format("2006-01-02")
	} else {
		m["start_date"] = nil
	}
	if e.TerminationDate != nil && !e.TerminationDate.IsZero() {
		m["termination_date"] = e.TerminationDate.Format("2006-01-02")
	} else {
		m["termination_date"] = nil
	}

	return m
}
