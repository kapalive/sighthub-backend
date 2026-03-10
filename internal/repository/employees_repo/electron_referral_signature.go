package employees_repo

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/employees"
)

type SignatureRepo struct {
	DB *gorm.DB
}

func NewSignatureRepo(db *gorm.DB) *SignatureRepo {
	return &SignatureRepo{DB: db}
}

// GetByEmployeeID returns the ElectronReferralSignature for the given employee.
// Returns gorm.ErrRecordNotFound when no record exists.
func (r *SignatureRepo) GetByEmployeeID(employeeID int) (*employees.ElectronReferralSignature, error) {
	var sig employees.ElectronReferralSignature
	if err := r.DB.Where("employee_id = ?", employeeID).First(&sig).Error; err != nil {
		return nil, err
	}
	return &sig, nil
}

// Exists returns true when there is at least one signature record for the employee.
func (r *SignatureRepo) Exists(employeeID int) (bool, error) {
	var count int64
	if err := r.DB.Model(&employees.ElectronReferralSignature{}).
		Where("employee_id = ?", employeeID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Upsert creates the signature when none exists, or updates the existing one.
func (r *SignatureRepo) Upsert(sig *employees.ElectronReferralSignature) error {
	var existing employees.ElectronReferralSignature
	err := r.DB.Where("employee_id = ?", sig.EmployeeID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.DB.Create(sig).Error
	} else if err != nil {
		return err
	}
	updates := map[string]interface{}{
		"description":   sig.Description,
		"path_link_img": sig.PathLinkImg,
	}
	return r.DB.Model(&existing).Updates(updates).Error
}

// Delete removes all signature records for the given employee.
func (r *SignatureRepo) Delete(employeeID int) error {
	return r.DB.Where("employee_id = ?", employeeID).
		Delete(&employees.ElectronReferralSignature{}).Error
}
