package employees_repo

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/employees"
)

type DoctorNPIRepo struct {
	DB *gorm.DB
}

func NewDoctorNPIRepo(db *gorm.DB) *DoctorNPIRepo {
	return &DoctorNPIRepo{DB: db}
}

// GetByEmployeeID returns the DoctorNpiNumber record for the given employee.
// Returns gorm.ErrRecordNotFound when no record exists.
func (r *DoctorNPIRepo) GetByEmployeeID(employeeID int) (*employees.DoctorNpiNumber, error) {
	var npi employees.DoctorNpiNumber
	if err := r.DB.Where("employee_id = ?", employeeID).First(&npi).Error; err != nil {
		return nil, err
	}
	return &npi, nil
}

// Upsert creates the record when it does not exist or updates it otherwise.
// The record is matched on employee_id.
func (r *DoctorNPIRepo) Upsert(npi *employees.DoctorNpiNumber) error {
	var existing employees.DoctorNpiNumber
	err := r.DB.Where("employee_id = ?", npi.EmployeeID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.DB.Create(npi).Error
	} else if err != nil {
		return err
	}
	updates := map[string]interface{}{
		"dr_npi_number": npi.DRNPINumber,
		"ein":           npi.EIN,
		"dea":           npi.DEA,
		"dea_expiration": npi.DEAExpiration,
		"printing_name": npi.PrintingName,
	}
	return r.DB.Model(&existing).Updates(updates).Error
}

// Delete removes the DoctorNpiNumber record for the given employee.
func (r *DoctorNPIRepo) Delete(employeeID int) error {
	return r.DB.Where("employee_id = ?", employeeID).Delete(&employees.DoctorNpiNumber{}).Error
}
