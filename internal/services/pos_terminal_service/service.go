package pos_terminal_service

import (
	"errors"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	"sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/general"
	"sighthub-backend/internal/models/invoices"
	"sighthub-backend/internal/models/location"
	"sighthub-backend/internal/models/patients"
	pkgActivity "sighthub-backend/pkg/activitylog"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrForbidden  = errors.New("forbidden")
	ErrBadRequest = errors.New("bad request")
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service { return &Service{db: db} }

type EmpLocation struct {
	Employee *employees.Employee
	Location *location.Location
}

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

func dec2(v float64) float64 {
	return math.Round(v*100) / 100
}

// ─── Terminal CRUD ───────────────────────────────────────────────────────────

func terminalDict(t *general.PaymentTerminal) map[string]interface{} {
	active := true
	if t.Active != nil {
		active = *t.Active
	}
	return map[string]interface{}{
		"terminal_id":      t.IDPaymentTerminal,
		"location_id":      t.LocationID,
		"title":            t.Title,
		"serial_number":    t.SerialNumber,
		"spin_register_id": t.SpinRegisterID,
		"spin_tpn":         t.SpinTPN,
		"active":           active,
		"is_default":       t.IsDefault,
	}
}

func (s *Service) ListTerminals(el *EmpLocation) ([]map[string]interface{}, error) {
	locID := int64(el.Location.IDLocation)
	var terms []general.PaymentTerminal
	s.db.Where("location_id = ? AND (active IS NULL OR active = true)", locID).Find(&terms)

	result := make([]map[string]interface{}, len(terms))
	for i, t := range terms {
		result[i] = terminalDict(&t)
	}
	return result, nil
}

type CreateTerminalRequest struct {
	Title          *string `json:"title"`
	SerialNumber   *string `json:"serial_number"`
	SpinRegisterID string  `json:"spin_register_id"`
	SpinTPN        string  `json:"spin_tpn"`
	Active         *bool   `json:"active"`
	SetAsDefault   bool    `json:"set_as_default"`
}

func (s *Service) CreateTerminal(el *EmpLocation, req CreateTerminalRequest) (map[string]interface{}, error) {
	if req.SpinRegisterID == "" {
		return nil, fmt.Errorf("%w: spin_register_id is required", ErrBadRequest)
	}
	if req.SpinTPN == "" {
		return nil, fmt.Errorf("%w: spin_tpn is required", ErrBadRequest)
	}

	locID := int64(el.Location.IDLocation)
	active := true
	if req.Active != nil {
		active = *req.Active
	}

	t := general.PaymentTerminal{
		LocationID:     &locID,
		Title:          req.Title,
		SerialNumber:   req.SerialNumber,
		SpinRegisterID: &req.SpinRegisterID,
		SpinTPN:        &req.SpinTPN,
		Active:         &active,
	}
	if err := s.db.Create(&t).Error; err != nil {
		return nil, err
	}

	if req.SetAsDefault {
		s.setDefaultTerminal(locID, &t)
	}

	pkgActivity.Log(s.db, "pos", "terminal_create",
		pkgActivity.WithEntity(int64(t.IDPaymentTerminal)),
		pkgActivity.WithDetails(map[string]interface{}{"location_id": locID}),
	)
	s.db.Save(&t)

	return map[string]interface{}{
		"ok":       true,
		"terminal": terminalDict(&t),
	}, nil
}

type UpdateTerminalRequest struct {
	Title          *string `json:"title"`
	SerialNumber   *string `json:"serial_number"`
	SpinRegisterID *string `json:"spin_register_id"`
	SpinTPN        *string `json:"spin_tpn"`
	Active         *bool   `json:"active"`
	SetAsDefault   bool    `json:"set_as_default"`
}

func (s *Service) UpdateTerminal(el *EmpLocation, terminalID int, req UpdateTerminalRequest) (map[string]interface{}, error) {
	var t general.PaymentTerminal
	if err := s.db.First(&t, terminalID).Error; err != nil {
		return nil, fmt.Errorf("%w: terminal not found", ErrNotFound)
	}

	locID := int64(el.Location.IDLocation)
	if t.LocationID == nil || *t.LocationID != locID {
		return nil, fmt.Errorf("%w: you can update terminals only in your location", ErrForbidden)
	}

	if req.Title != nil {
		t.Title = req.Title
	}
	if req.SerialNumber != nil {
		t.SerialNumber = req.SerialNumber
	}
	if req.SpinRegisterID != nil {
		if *req.SpinRegisterID == "" {
			return nil, fmt.Errorf("%w: spin_register_id cannot be empty", ErrBadRequest)
		}
		t.SpinRegisterID = req.SpinRegisterID
	}
	if req.SpinTPN != nil {
		if *req.SpinTPN == "" {
			return nil, fmt.Errorf("%w: spin_tpn cannot be empty", ErrBadRequest)
		}
		t.SpinTPN = req.SpinTPN
	}
	if req.Active != nil {
		t.Active = req.Active
	}

	if req.SetAsDefault {
		if t.Active != nil && !*t.Active {
			return nil, fmt.Errorf("%w: terminal is inactive", ErrBadRequest)
		}
		s.setDefaultTerminal(locID, &t)
	}

	s.db.Save(&t)
	pkgActivity.Log(s.db, "pos", "terminal_update",
		pkgActivity.WithEntity(int64(terminalID)),
	)

	return map[string]interface{}{
		"ok":       true,
		"terminal": terminalDict(&t),
	}, nil
}

func (s *Service) SetDefaultTerminal(el *EmpLocation, terminalID int) (map[string]interface{}, error) {
	var t general.PaymentTerminal
	if err := s.db.First(&t, terminalID).Error; err != nil {
		return nil, fmt.Errorf("%w: terminal not found", ErrNotFound)
	}

	locID := int64(el.Location.IDLocation)
	if t.LocationID == nil || *t.LocationID != locID {
		return nil, fmt.Errorf("%w: you can set default only for terminals in your location", ErrForbidden)
	}
	if t.Active != nil && !*t.Active {
		return nil, fmt.Errorf("%w: terminal is inactive", ErrBadRequest)
	}

	s.setDefaultTerminal(locID, &t)
	s.db.Save(&t)
	pkgActivity.Log(s.db, "pos", "terminal_default",
		pkgActivity.WithEntity(int64(terminalID)),
	)

	return map[string]interface{}{
		"ok":                  true,
		"default_terminal_id": t.IDPaymentTerminal,
	}, nil
}

func (s *Service) DeleteTerminal(el *EmpLocation, terminalID int) (map[string]interface{}, error) {
	var t general.PaymentTerminal
	if err := s.db.First(&t, terminalID).Error; err != nil {
		return nil, fmt.Errorf("%w: terminal not found", ErrNotFound)
	}

	locID := int64(el.Location.IDLocation)
	if t.LocationID == nil || *t.LocationID != locID {
		return nil, fmt.Errorf("%w: you can delete terminals only in your location", ErrForbidden)
	}

	// Soft delete: deactivate
	if t.Active != nil {
		f := false
		t.Active = &f
		t.IsDefault = false
		s.db.Save(&t)
		pkgActivity.Log(s.db, "pos", "terminal_deactivate",
			pkgActivity.WithEntity(int64(terminalID)),
		)
		return map[string]interface{}{
			"ok":          true,
			"message":     "Terminal deactivated",
			"terminal_id": t.IDPaymentTerminal,
		}, nil
	}

	// Hard delete only if not used
	var count int64
	s.db.Model(&general.PaymentTransaction{}).Where("payment_terminal_id = ?", terminalID).Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("%w: terminal has transactions; cannot delete (use soft delete)", ErrBadRequest)
	}

	s.db.Delete(&t)
	pkgActivity.Log(s.db, "pos", "terminal_delete",
		pkgActivity.WithEntity(int64(terminalID)),
	)

	return map[string]interface{}{
		"ok":          true,
		"message":     "Terminal deleted",
		"terminal_id": terminalID,
	}, nil
}

func (s *Service) setDefaultTerminal(locationID int64, t *general.PaymentTerminal) {
	s.db.Model(&general.PaymentTerminal{}).
		Where("location_id = ? AND is_default = true", locationID).
		Update("is_default", false)
	t.IsDefault = true
}

// ─── POS Start ───────────────────────────────────────────────────────────────

type PosStartRequest struct {
	Amount          float64 `json:"amount"`
	PaymentMethodID int64   `json:"payment_method_id"`
	TerminalID      *int    `json:"terminal_id"`
}

func (s *Service) PosStart(el *EmpLocation, invoiceID int64, req PosStartRequest) (map[string]interface{}, error) {
	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: invoice not found", ErrNotFound)
	}
	if inv.Finalized {
		return nil, fmt.Errorf("%w: invoice is finalized (locked)", ErrForbidden)
	}
	locID := int64(el.Location.IDLocation)
	if inv.LocationID != locID {
		return nil, fmt.Errorf("%w: POS payments can only be started at invoice location", ErrForbidden)
	}

	if req.Amount <= 0 {
		return nil, fmt.Errorf("%w: amount must be > 0", ErrBadRequest)
	}
	if req.PaymentMethodID == 0 {
		return nil, fmt.Errorf("%w: payment_method_id is required", ErrBadRequest)
	}

	terminal, err := s.getTerminalForInvoice(&inv, req.TerminalID)
	if err != nil {
		return nil, err
	}

	cfg, err := s.getSpinConfig(inv.LocationID)
	if err != nil {
		return nil, err
	}
	_ = cfg // will be used when SPIn integration is ready

	refID := fmt.Sprintf("%d%d", time.Now().UnixNano(), inv.IDInvoice)
	if len(refID) > 20 {
		refID = refID[:20]
	}

	now := time.Now()
	invNum := fmt.Sprintf("%d", inv.IDInvoice)
	pmID := req.PaymentMethodID
	tx := general.PaymentTransaction{
		PaymentTerminalID: terminal.IDPaymentTerminal,
		InvoiceID:         &inv.IDInvoice,
		PaymentMethodID:   &pmID,
		TransactionDate:   now,
		CreatedAt:         now,
		Amount:            dec2(req.Amount),
		Currency:          "USD",
		Status:            "pending",
		SpinRefID:         refID,
		SpinRegisterID:    terminal.SpinRegisterID,
		SpinTPN:           terminal.SpinTPN,
		SpinInvNum:        &invNum,
	}
	if err := s.db.Create(&tx).Error; err != nil {
		return nil, err
	}

	// TODO: SPIn XML call will go here when tokens/keys are available
	// For now, transaction is created with status="pending"

	pkgActivity.Log(s.db, "pos", "payment_start",
		pkgActivity.WithEntity(invoiceID),
		pkgActivity.WithDetails(map[string]interface{}{
			"amount":      fmt.Sprintf("%.2f", req.Amount),
			"terminal_id": terminal.IDPaymentTerminal,
		}),
	)

	return map[string]interface{}{
		"ok":                     true,
		"payment_transaction_id": tx.IDPaymentTransaction,
		"status":                 tx.Status,
		"message":                "Transaction created (SPIn integration pending)",
		"spin_ref_id":            tx.SpinRefID,
	}, nil
}

// ─── POS Commit ──────────────────────────────────────────────────────────────

type PosCommitRequest struct {
	PaymentTransactionID int64 `json:"payment_transaction_id"`
}

func (s *Service) PosCommit(el *EmpLocation, invoiceID int64, req PosCommitRequest) (map[string]interface{}, error) {
	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: invoice not found", ErrNotFound)
	}
	if inv.Finalized {
		return nil, fmt.Errorf("%w: invoice is finalized (locked)", ErrForbidden)
	}
	locID := int64(el.Location.IDLocation)
	if inv.LocationID != locID {
		return nil, fmt.Errorf("%w: POS commit can only be done at invoice location", ErrForbidden)
	}

	if req.PaymentTransactionID == 0 {
		return nil, fmt.Errorf("%w: payment_transaction_id is required", ErrBadRequest)
	}

	var tx general.PaymentTransaction
	if err := s.db.First(&tx, req.PaymentTransactionID).Error; err != nil {
		return nil, fmt.Errorf("%w: transaction not found", ErrNotFound)
	}
	if tx.InvoiceID == nil || *tx.InvoiceID != inv.IDInvoice {
		return nil, fmt.Errorf("%w: transaction not found", ErrNotFound)
	}
	if tx.Status != "approved" {
		return nil, fmt.Errorf("transaction not approved, status: %s", tx.Status)
	}

	// Check if already committed
	var existingPH patients.PaymentHistory
	if s.db.Where("payment_transaction_id = ?", tx.IDPaymentTransaction).First(&existingPH).Error == nil {
		return map[string]interface{}{
			"ok":      true,
			"message": "Already committed",
		}, nil
	}

	amount := dec2(tx.Amount)
	oldDue := dec2(inv.Due)
	oldPtBal := dec2(inv.PTBal)

	inv.PTBal = math.Max(dec2(oldPtBal-amount), 0)
	newDue := dec2(oldDue - amount)
	leftover := 0.0
	if newDue < 0 {
		leftover = dec2(-newDue)
		newDue = 0
	}
	inv.Due = newDue

	empID := int64(el.Employee.IDEmployee)
	ph := patients.PaymentHistory{
		PatientID:            &inv.PatientID,
		InvoiceID:            inv.IDInvoice,
		Amount:               amount,
		PaymentTimestamp:      time.Now(),
		PaymentMethodID:      tx.PaymentMethodID,
		EmployeeID:           &empID,
		PaymentTransactionID: &tx.IDPaymentTransaction,
	}
	s.db.Create(&ph)
	s.db.Save(&inv)

	if leftover > 0 {
		var cb patients.ClientBalance
		if s.db.Where("patient_id = ? AND location_id = ?", inv.PatientID, inv.LocationID).First(&cb).Error != nil {
			cb = patients.ClientBalance{
				PatientID:  inv.PatientID,
				Credit:     0,
				LocationID: int(inv.LocationID),
			}
			s.db.Create(&cb)
		}
		cb.Credit = dec2(cb.Credit + leftover)
		s.db.Save(&cb)
	}

	pkgActivity.Log(s.db, "pos", "payment_commit",
		pkgActivity.WithEntity(invoiceID),
		pkgActivity.WithDetails(map[string]interface{}{"amount": fmt.Sprintf("%.2f", amount)}),
	)

	return map[string]interface{}{
		"ok":                     true,
		"amount_paid":            fmt.Sprintf("%.2f", amount),
		"old_due":                fmt.Sprintf("%.2f", oldDue),
		"new_due":                fmt.Sprintf("%.2f", inv.Due),
		"credit_added":           fmt.Sprintf("%.2f", leftover),
		"payment_transaction_id": tx.IDPaymentTransaction,
	}, nil
}

// ─── SPIn Config ─────────────────────────────────────────────────────────────

type ProvisionSpinConfigRequest struct {
	SpinURL    string `json:"spin_url"`
	AuthKey    string `json:"auth_key"`
	TimeoutSec int    `json:"timeout_sec"`
	Active     *bool  `json:"active"`
	MerchantID string `json:"merchant_id"`
}

func (s *Service) ProvisionSpinConfig(el *EmpLocation, req ProvisionSpinConfigRequest) (map[string]interface{}, error) {
	if req.MerchantID == "" {
		return nil, fmt.Errorf("%w: merchant_id is required", ErrBadRequest)
	}
	if req.SpinURL == "" {
		return nil, fmt.Errorf("%w: spin_url is required", ErrBadRequest)
	}
	if req.AuthKey == "" {
		return nil, fmt.Errorf("%w: auth_key is required", ErrBadRequest)
	}
	if req.TimeoutSec <= 0 {
		req.TimeoutSec = 125
	}

	locID := int64(el.Location.IDLocation)
	active := true
	if req.Active != nil {
		active = *req.Active
	}

	var cfg general.SpinConfig
	if s.db.Where("location_id = ?", locID).First(&cfg).Error != nil {
		cfg = general.SpinConfig{LocationID: locID}
	}

	cfg.SpinURL = &req.SpinURL
	cfg.AuthKey = req.AuthKey
	cfg.TimeoutSec = req.TimeoutSec
	cfg.MerchantID = &req.MerchantID
	cfg.Active = active
	cfg.UpdatedAt = time.Now()

	s.db.Save(&cfg)

	pkgActivity.Log(s.db, "pos", "spin_config_provision",
		pkgActivity.WithEntity(locID),
	)

	return map[string]interface{}{
		"ok":          true,
		"location_id": locID,
		"active":      cfg.Active,
		"timeout_sec": cfg.TimeoutSec,
		"spin_url":    cfg.SpinURL,
		"merchant_id": cfg.MerchantID,
	}, nil
}

func (s *Service) GetSpinConfig(el *EmpLocation) (map[string]interface{}, error) {
	locID := int64(el.Location.IDLocation)
	var cfg general.SpinConfig
	if err := s.db.Where("location_id = ?", locID).First(&cfg).Error; err != nil {
		return nil, fmt.Errorf("%w: SPIn config not found for this location", ErrNotFound)
	}

	maskedKey := maskSecret(cfg.AuthKey, 4)

	return map[string]interface{}{
		"ok":          true,
		"location_id": locID,
		"active":      cfg.Active,
		"timeout_sec": cfg.TimeoutSec,
		"spin_url":    cfg.SpinURL,
		"merchant_id": cfg.MerchantID,
		"auth_key":    maskedKey,
		"updated_at":  cfg.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// ─── GET /pos/tx/{id} ────────────────────────────────────────────────────────

func (s *Service) GetTransaction(el *EmpLocation, txID int64) (map[string]interface{}, error) {
	var tx general.PaymentTransaction
	if err := s.db.First(&tx, txID).Error; err != nil {
		return nil, fmt.Errorf("%w: transaction not found", ErrNotFound)
	}

	if tx.InvoiceID != nil {
		var inv invoices.Invoice
		if s.db.First(&inv, *tx.InvoiceID).Error == nil {
			if inv.LocationID != int64(el.Location.IDLocation) {
				return nil, fmt.Errorf("%w: access denied", ErrForbidden)
			}
		}
	}

	return map[string]interface{}{
		"ok":                     true,
		"payment_transaction_id": tx.IDPaymentTransaction,
		"status":                 tx.Status,
		"message":                tx.SpinRespMsg,
		"result_code":            tx.SpinResultCode,
		"spin_ref_id":            tx.SpinRefID,
		"amount":                 fmt.Sprintf("%.2f", tx.Amount),
		"created_at":             tx.CreatedAt.Format(time.RFC3339),
	}, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (s *Service) getTerminalForInvoice(inv *invoices.Invoice, terminalID *int) (*general.PaymentTerminal, error) {
	var terminal general.PaymentTerminal
	if terminalID == nil || *terminalID == 0 {
		if err := s.db.Where("location_id = ? AND is_default = true", inv.LocationID).First(&terminal).Error; err != nil {
			return nil, fmt.Errorf("%w: terminal_id is required (no default terminal for this location)", ErrBadRequest)
		}
	} else {
		if err := s.db.First(&terminal, *terminalID).Error; err != nil {
			return nil, fmt.Errorf("%w: payment terminal not found", ErrNotFound)
		}
	}

	if terminal.LocationID == nil {
		return nil, fmt.Errorf("%w: terminal has no location assigned", ErrBadRequest)
	}
	if *terminal.LocationID != inv.LocationID {
		return nil, fmt.Errorf("%w: this terminal belongs to a different location", ErrForbidden)
	}
	if terminal.SpinRegisterID == nil || *terminal.SpinRegisterID == "" {
		return nil, fmt.Errorf("%w: terminal spin_register_id is not set", ErrBadRequest)
	}
	if terminal.SpinTPN == nil || *terminal.SpinTPN == "" {
		return nil, fmt.Errorf("%w: terminal spin_tpn is not set", ErrBadRequest)
	}
	if terminal.Active != nil && !*terminal.Active {
		return nil, fmt.Errorf("%w: terminal is inactive", ErrBadRequest)
	}
	return &terminal, nil
}

func (s *Service) getSpinConfig(locationID int64) (*general.SpinConfig, error) {
	var cfg general.SpinConfig
	if err := s.db.Where("location_id = ? AND active = true", locationID).First(&cfg).Error; err != nil {
		return nil, fmt.Errorf("%w: SPIn config not found for this location", ErrBadRequest)
	}
	if cfg.SpinURL == nil || *cfg.SpinURL == "" {
		return nil, fmt.Errorf("%w: SPIn spin_url is empty for this location", ErrBadRequest)
	}
	if cfg.AuthKey == "" {
		return nil, fmt.Errorf("%w: SPIn auth_key is empty for this location", ErrBadRequest)
	}
	if cfg.TimeoutSec <= 0 {
		return nil, fmt.Errorf("%w: SPIn timeout_sec is invalid", ErrBadRequest)
	}
	if cfg.MerchantID == nil || *cfg.MerchantID == "" {
		return nil, fmt.Errorf("%w: SPIn merchant_id is empty for this location", ErrBadRequest)
	}
	return &cfg, nil
}

func maskSecret(s string, keepLast int) string {
	if s == "" {
		return ""
	}
	if len(s) <= keepLast {
		masked := ""
		for range s {
			masked += "*"
		}
		return masked
	}
	masked := ""
	for i := 0; i < len(s)-keepLast; i++ {
		masked += "*"
	}
	return masked + s[len(s)-keepLast:]
}
