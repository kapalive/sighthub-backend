package employees

import (
	"fmt"
	"time"
)

// DoctorNpiNumber ⇄ doctor_npi_number
type DoctorNpiNumber struct {
	IDDoctorNPINumber int        `gorm:"column:id_doctor_npi_number;primaryKey;autoIncrement"                    json:"id_doctor_npi_number"`
	DRNPINumber       string     `gorm:"column:dr_npi_number;type:varchar(12);not null;uniqueIndex:uniq_dr_npi" json:"dr_npi_number"`
	EIN               *string    `gorm:"column:ein;type:varchar(20)"                                            json:"ein,omitempty"`
	DEA               *string    `gorm:"column:dea;type:varchar(30)"                                            json:"dea,omitempty"`
	DEAExpiration     *time.Time `gorm:"column:dea_expiration;type:date"                                     json:"-"`
	PrintingName      *string    `gorm:"column:printing_name;type:varchar(100)"                                 json:"printing_name,omitempty"`
	EmployeeID        *int64     `gorm:"column:employee_id;uniqueIndex:uniq_emp_npi"                            json:"employee_id,omitempty"`
}

func (DoctorNpiNumber) TableName() string { return "doctor_npi_number" }

// ToMap — аналог Python to_dict()
func (d *DoctorNpiNumber) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_doctor_npi_number": d.IDDoctorNPINumber,
		"dr_npi_number":        d.DRNPINumber,
		"ein":                  d.EIN,
		"dea":                  d.DEA,
		"printing_name":        d.PrintingName,
		"employee_id":          d.EmployeeID,
	}
	if d.DEAExpiration != nil && !d.DEAExpiration.IsZero() {
		m["dea_expiration"] = d.DEAExpiration.Format("2006-01-02")
	} else {
		m["dea_expiration"] = nil
	}
	return m
}

func (d *DoctorNpiNumber) String() string {
	emp := "nil"
	if d.EmployeeID != nil {
		emp = fmt.Sprintf("%d", *d.EmployeeID)
	}
	return fmt.Sprintf("<DoctorNpiNumber %s | Employee ID: %s>", d.DRNPINumber, emp)
}
