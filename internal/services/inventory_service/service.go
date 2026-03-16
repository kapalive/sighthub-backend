package inventory_service

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	locModel "sighthub-backend/internal/models/location"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

func (s *Service) getEmployeeAndLocation(username string) (*empModel.Employee, *locModel.Location, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, nil, errors.New("employee not found")
	}
	var emp empModel.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, nil, errors.New("employee not found")
	}
	if emp.LocationID == nil {
		return nil, nil, errors.New("employee has no location")
	}
	var loc locModel.Location
	if err := s.db.First(&loc, *emp.LocationID).Error; err != nil {
		return nil, nil, errors.New("location not found")
	}
	return &emp, &loc, nil
}

func int64Ptr(v int64) *int64 { return &v }

func strPtr(v string) *string { return &v }

func fmtPrice(v *float64) string {
	if v == nil {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", *v)
}
