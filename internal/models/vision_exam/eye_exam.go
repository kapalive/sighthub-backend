// internal/models/vision_exam/eye_exam.go
package vision_exam

import (
	"time"

	"sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/location"
)

// EyeExam ↔ table: eye_exam
type EyeExam struct {
	IDEyeExam      int64     `gorm:"column:id_eye_exam;primaryKey;autoIncrement" json:"id_eye_exam"`
	EyeExamDate    time.Time `gorm:"column:eye_exam_date;type:timestamptz;not null" json:"eye_exam_date"`
	EmployeeID     int64     `gorm:"column:employee_id;not null"                 json:"employee_id"`
	EyeExamTypeID  int64     `gorm:"column:eye_exam_type_id;not null"            json:"eye_exam_type_id"`
	LocationID     int       `gorm:"column:location_id;not null"                 json:"location_id"`
	PatientID      int64     `gorm:"column:patient_id;not null"                  json:"patient_id"`
	Passed         bool      `gorm:"column:passed;not null;default:false"        json:"passed"`

	// preload relations
	Employee    *employees.Employee `gorm:"foreignKey:EmployeeID;references:IDEmployee"     json:"-"`
	EyeExamType *EyeExamType        `gorm:"foreignKey:EyeExamTypeID;references:IDEyeExamType" json:"-"`
	Location    *location.Location  `gorm:"foreignKey:LocationID;references:IDLocation"     json:"-"`
}

func (EyeExam) TableName() string { return "eye_exam" }

func (e *EyeExam) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_eye_exam":      e.IDEyeExam,
		"eye_exam_date":    e.EyeExamDate.Format(time.RFC3339),
		"employee_id":      e.EmployeeID,
		"eye_exam_type_id": e.EyeExamTypeID,
		"location_id":      e.LocationID,
		"patient_id":       e.PatientID,
		"passed":           e.Passed,
	}
	if e.Employee != nil {
		m["employee"] = map[string]interface{}{
			"id_employee": e.Employee.IDEmployee,
			"first_name":  e.Employee.FirstName,
			"last_name":   e.Employee.LastName,
		}
	}
	if e.EyeExamType != nil {
		m["eye_exam_type"] = e.EyeExamType.ToMap()
	}
	return m
}
