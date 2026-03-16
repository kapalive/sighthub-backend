package accounting_service

import (
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/general"
	locModel "sighthub-backend/internal/models/location"
	vendorModel "sighthub-backend/internal/models/vendors"
)

const adjustmentPaymentMethodID = 22

var validAccountStatuses = map[string]bool{"Active": true, "Closed": true, "Blocked": true}

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ── helpers ─────────────────────────────────────────────────────────────────

func (s *Service) getEmployeeAndLocation(username string) (*empModel.Employee, *locModel.Location, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, nil, errors.New("employee or location not found")
	}
	var emp empModel.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, nil, errors.New("employee or location not found")
	}
	if emp.LocationID == nil {
		return nil, nil, errors.New("employee or location not found")
	}
	var loc locModel.Location
	if err := s.db.First(&loc, *emp.LocationID).Error; err != nil {
		return nil, nil, errors.New("employee or location not found")
	}
	return &emp, &loc, nil
}

// resolveAccountingLocation: if warehouse location → use showcase location of same store.
func (s *Service) resolveAccountingLocation(loc *locModel.Location) (*locModel.Location, error) {
	if loc == nil {
		return nil, errors.New("accounting location not found")
	}
	if loc.WarehouseID != nil {
		var showcase locModel.Location
		err := s.db.
			Where("store_id = ? AND showcase = ?", loc.StoreID, true).
			Order("id_location ASC").
			First(&showcase).Error
		if err != nil {
			return loc, nil // fallback to original
		}
		return &showcase, nil
	}
	return loc, nil
}

// ensureVendorAccountForLocation finds or creates a VendorLocationAccount.
func (s *Service) ensureVendorAccountForLocation(vendorID int, locID int64, data map[string]interface{}) (*vendorModel.VendorLocationAccount, error) {
	accNumber := ""
	if v, ok := data["account_number"]; ok && v != nil {
		accNumber = strings.TrimSpace(v.(string))
	}
	if accNumber == "" {
		return nil, errors.New("account_number is required")
	}

	accIDRaw := data["vendor_location_account_id"]

	var row vendorModel.VendorLocationAccount

	if accIDRaw != nil {
		var accID int64
		switch v := accIDRaw.(type) {
		case float64:
			accID = int64(v)
		case int64:
			accID = v
		case int:
			accID = int64(v)
		}
		err := s.db.Where("id_vendor_location_account = ? AND vendor_id = ? AND location_id = ?", accID, vendorID, locID).First(&row).Error
		if err != nil {
			return nil, errors.New("vendor_location_account_id not found for this vendor+location")
		}
		if row.AccountNumber != accNumber {
			return nil, errors.New("account_number does not match vendor_location_account_id")
		}
		return &row, nil
	}

	// find by account_number
	err := s.db.Where("vendor_id = ? AND location_id = ? AND account_number = ?", vendorID, locID, accNumber).
		Order("created_at DESC").First(&row).Error
	if err == nil {
		return &row, nil
	}

	// create
	isActive := true
	if v, ok := data["is_active"]; ok && v != nil {
		if b, ok := v.(bool); ok {
			isActive = b
		}
	}
	var statusStr *string
	if v, ok := data["status"]; ok && v != nil {
		if s2, ok := v.(string); ok && s2 != "" {
			statusStr = &s2
		}
	}
	var qbRef *string
	if v, ok := data["qb_vendor_ref"]; ok && v != nil {
		if s2, ok := v.(string); ok && s2 != "" {
			qbRef = &s2
		}
	}
	note := "auto-created by API"
	var notePtr *string
	if v, ok := data["note"]; ok && v != nil {
		if s2, ok := v.(string); ok && s2 != "" {
			note = s2
		}
	}
	notePtr = &note

	newRow := vendorModel.VendorLocationAccount{
		VendorID:      vendorID,
		LocationID:    locID,
		AccountNumber: accNumber,
		Status:        statusStr,
		IsActive:      isActive,
		QbVendorRef:   qbRef,
		Note:          notePtr,
	}
	if createErr := s.db.Create(&newRow).Error; createErr != nil {
		// retry find (race condition)
		if findErr := s.db.Where("vendor_id = ? AND location_id = ? AND account_number = ?", vendorID, locID, accNumber).
			Order("created_at DESC").First(&row).Error; findErr != nil {
			return nil, errors.New("failed to create/find vendor account")
		}
		return &row, nil
	}
	return &newRow, nil
}

func pmLabel(pm *general.PaymentMethod) *string {
	if pm == nil {
		return nil
	}
	if pm.ShortName != nil && *pm.ShortName != "" {
		return pm.ShortName
	}
	return &pm.MethodName
}

func utcToday() time.Time {
	return time.Now().UTC().Truncate(24 * time.Hour)
}

// apInvoiceAdapter wraps VendorAPInvoice to implement pkg/accounting.Invoice interface.
type apInvoiceAdapter struct {
	inv *vendorModel.VendorAPInvoice
}

func (a apInvoiceAdapter) GetInvoiceDate() time.Time { return a.inv.InvoiceDate }
func (a apInvoiceAdapter) GetTerms() *int            { return a.inv.Terms }
func (a apInvoiceAdapter) GetInvoiceAmount() decimal.Decimal {
	d, _ := decimal.NewFromString(a.inv.InvoiceAmount)
	return d
}
func (a apInvoiceAdapter) GetOpenBalance() decimal.Decimal {
	d, _ := decimal.NewFromString(a.inv.OpenBalance)
	return d
}
