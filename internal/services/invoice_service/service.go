package invoice_service

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	"sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/frames"
	"sighthub-backend/internal/models/location"
	"sighthub-backend/internal/models/vendors"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrForbidden       = errors.New("forbidden")
	ErrBadRequest      = errors.New("bad request")
	ErrInternalTransfer = errors.New("internal transfer error")
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ─── Common helpers ───────────────────────────────────────────────────────────

type EmpLocation struct {
	Employee *employees.Employee
	Location *location.Location
}

// GetEmpLocation resolves the current user's employee + location from JWT username.
func (s *Service) GetEmpLocation(username string) (*EmpLocation, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, fmt.Errorf("login not found")
	}

	var emp employees.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, fmt.Errorf("employee not found")
	}

	if emp.LocationID == nil {
		return nil, fmt.Errorf("employee not assigned to a location")
	}

	var loc location.Location
	if err := s.db.First(&loc, *emp.LocationID).Error; err != nil {
		return nil, fmt.Errorf("location not found")
	}

	return &EmpLocation{Employee: &emp, Location: &loc}, nil
}

// fmtFloat formats a float64 to "%.2f" string.
func fmtFloat(f float64) string { return fmt.Sprintf("%.2f", f) }

// fmtFloatPtr formats a *float64 or returns "0.00".
func fmtFloatPtr(f *float64) string {
	if f == nil {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", *f)
}

// enrichItemMeta fills ProductTitle, VariantTitle, BrandName on a ReceiptItemRow
// using the inventory item's ModelID.
func (s *Service) enrichItemMeta(row *ReceiptItemRow, modelID *int64) {
	if modelID == nil {
		return
	}
	var m frames.Model
	if err := s.db.First(&m, *modelID).Error; err != nil {
		return
	}
	row.VariantTitle = &m.TitleVariant

	var p frames.Product
	if err := s.db.First(&p, m.ProductID).Error; err != nil {
		return
	}
	row.ProductTitle = &p.TitleProduct

	if p.BrandID != nil {
		var b vendors.Brand
		if err := s.db.First(&b, *p.BrandID).Error; err == nil {
			row.BrandName = b.BrandName
		}
	}
}
