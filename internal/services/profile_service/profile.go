package profile_service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	authModel "sighthub-backend/internal/models/auth"
	auditModel "sighthub-backend/internal/models/audit"
	empModel "sighthub-backend/internal/models/employees"
	insModel "sighthub-backend/internal/models/insurance"
	invModel "sighthub-backend/internal/models/invoices"
	pkgAuth "sighthub-backend/pkg/auth"
)

const (
	blacklistTTL      = 24 * time.Hour
	lockDuration      = 1 * time.Minute
	maxFailedAttempts = 3
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrEmployeeNotFound = errors.New("employee not found")
	ErrInvoiceNotFound  = errors.New("invoice not found")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrLockedOut        = errors.New("too many incorrect password attempts. You have been logged out and temporarily locked.")
)

type Service struct {
	db  *gorm.DB
	rdb *redis.Client
	cfg *config.Config
}

func New(db *gorm.DB, rdb *redis.Client, cfg *config.Config) *Service {
	return &Service{db: db, rdb: rdb, cfg: cfg}
}

// ─── ExpressPass ──────────────────────────────────────────────────────────────

func (s *Service) ExpressPass(ctx context.Context, username, password string) (string, error) {
	var user authModel.EmployeeLogin
	if err := s.db.WithContext(ctx).Where("employee_login = ?", username).First(&user).Error; err != nil {
		return "", ErrUserNotFound
	}

	if !user.CheckPassword(password) {
		user.FailedAttempts++
		s.db.WithContext(ctx).Model(&user).Update("failed_attempts", user.FailedAttempts)

		if user.FailedAttempts >= maxFailedAttempts {
			s.blacklistUser(ctx, username)
			s.logoutUser(ctx, username)
			return "", ErrLockedOut
		}
		return "", ErrInvalidPassword
	}

	s.db.WithContext(ctx).Model(&user).Update("failed_attempts", 0)

	pin, err := s.generateUniquePin(ctx, &user)
	if err != nil {
		return "", err
	}
	return pin, nil
}

// ─── GetInfo ──────────────────────────────────────────────────────────────────

type EmployeeInfo struct {
	FirstName    string  `json:"first_name"`
	MiddleName   *string `json:"middle_name"`
	LastName     string  `json:"last_name"`
	Suffix       *string `json:"suffix"`
	Phone        *string `json:"phone"`
	Email        *string `json:"email"`
	PrintingName *string `json:"printing_name"`
}

func (s *Service) GetInfo(ctx context.Context, username string) (*EmployeeInfo, error) {
	var user authModel.EmployeeLogin
	if err := s.db.WithContext(ctx).Where("employee_login = ?", username).First(&user).Error; err != nil {
		return nil, ErrUserNotFound
	}

	var emp empModel.Employee
	if err := s.db.WithContext(ctx).Where("employee_login_id = ?", user.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, ErrEmployeeNotFound
	}

	info := &EmployeeInfo{
		FirstName:  emp.FirstName,
		MiddleName: emp.MiddleName,
		LastName:   emp.LastName,
		Suffix:     emp.Suffix,
		Phone:      emp.Phone,
		Email:      emp.Email,
	}

	var npi empModel.DoctorNpiNumber
	if s.db.WithContext(ctx).Where("employee_id = ?", emp.IDEmployee).First(&npi).Error == nil {
		info.PrintingName = npi.PrintingName
	}

	return info, nil
}

// ─── ChangePassword ───────────────────────────────────────────────────────────

func (s *Service) ChangePassword(ctx context.Context, username, currentPassword, newPassword string) error {
	var user authModel.EmployeeLogin
	if err := s.db.WithContext(ctx).Where("employee_login = ?", username).First(&user).Error; err != nil {
		return ErrUserNotFound
	}

	if !user.CheckPassword(currentPassword) {
		user.FailedAttempts++
		s.db.WithContext(ctx).Model(&user).Update("failed_attempts", user.FailedAttempts)

		if user.FailedAttempts >= maxFailedAttempts {
			s.blacklistUser(ctx, username)
			s.logoutUser(ctx, username)
			return ErrLockedOut
		}
		return errors.New("current password is incorrect")
	}

	user.FailedAttempts = 0
	if err := user.SetPassword(newPassword); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Save(&user).Error
}

// ─── GetInvoiceByID ───────────────────────────────────────────────────────────

type InvoiceInsurancePolicy struct {
	IDInsurancePolicy    int64   `json:"id_insurance_policy"`
	GroupNumber          *string `json:"group_number"`
	CoverageDetails      *string `json:"coverage_details"`
	InsuranceCompanyName *string `json:"insurance_company_name"`
}

type InvoiceResponse struct {
	IDInvoice           int64                   `json:"id_invoice"`
	NumberInvoice       string                  `json:"number_invoice"`
	DateCreate          *string                 `json:"date_create"`
	PaymentMethodID     *int64                  `json:"payment_method_id"`
	PaidInsuranceStatus string                  `json:"paid_insurance_status"`
	EmployeeID          *int64                  `json:"employee_id"`
	PTBal               float64                 `json:"pt_bal"`
	Discount            *float64                `json:"discount"`
	TotalAmount         float64                 `json:"total_amount"`
	FinalAmount         float64                 `json:"final_amount"`
	InsBal              float64                 `json:"ins_bal"`
	GiftCardBal         *float64                `json:"gift_card_bal"`
	Due                 float64                 `json:"due"`
	StatusInvoiceID     int64                   `json:"status_invoice_id"`
	Quantity            int                     `json:"quantity"`
	Referral            *string                 `json:"referral"`
	Class               *string                 `json:"class"`
	Reason              *string                 `json:"reason"`
	DoctorID            *int64                  `json:"doctor_id"`
	LocationID          int64                   `json:"location_id"`
	ToLocationID        *int64                  `json:"to_location_id"`
	PatientID           int64                   `json:"patient_id"`
	VendorID            int64                   `json:"vendor_id"`
	InsurancePolicyID   *int64                  `json:"insurance_policy_id"`
	Remake              bool                    `json:"remake"`
	TaxAmount           float64                 `json:"tax_amount"`
	Finalized           bool                    `json:"finalized"`
	InsurancePolicy     *InvoiceInsurancePolicy `json:"insurance_policy,omitempty"`
}

func (s *Service) GetInvoiceByID(ctx context.Context, invoiceID int64) (*InvoiceResponse, error) {
	var inv invModel.Invoice
	if err := s.db.WithContext(ctx).First(&inv, invoiceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvoiceNotFound
		}
		return nil, err
	}

	dateCreate := inv.DateCreate.Format(time.RFC3339)
	resp := &InvoiceResponse{
		IDInvoice:           inv.IDInvoice,
		NumberInvoice:       inv.NumberInvoice,
		DateCreate:          &dateCreate,
		PaymentMethodID:     inv.PaymentMethodID,
		PaidInsuranceStatus: func() string { if inv.PaidInsuranceStatus != nil { return string(*inv.PaidInsuranceStatus) }; return "" }(),
		EmployeeID:          inv.EmployeeID,
		PTBal:               inv.PTBal,
		Discount:            inv.Discount,
		TotalAmount:         inv.TotalAmount,
		FinalAmount:         inv.FinalAmount,
		InsBal:              inv.InsBal,
		GiftCardBal:         inv.GiftCardBal,
		Due:                 inv.Due,
		StatusInvoiceID:     func() int64 { if inv.StatusInvoiceID != nil { return *inv.StatusInvoiceID }; return 0 }(),
		Quantity:            inv.Quantity,
		Referral:            inv.Referral,
		Class:               inv.ClassField,
		Reason:              inv.Reason,
		DoctorID:            inv.DoctorID,
		LocationID:          inv.LocationID,
		ToLocationID:        inv.ToLocationID,
		PatientID:           inv.PatientID,
		VendorID:            func() int64 { if inv.VendorID != nil { return *inv.VendorID }; return 0 }(),
		InsurancePolicyID:   inv.InsurancePolicyID,
		Remake:              inv.Remake,
		TaxAmount:           inv.TaxAmount,
		Finalized:           inv.Finalized,
	}

	if inv.InsurancePolicyID != nil {
		var policy insModel.InsurancePolicy
		if s.db.WithContext(ctx).Preload("InsuranceCompany").First(&policy, *inv.InsurancePolicyID).Error == nil {
			ip := &InvoiceInsurancePolicy{
				IDInsurancePolicy: policy.IDInsurancePolicy,
				GroupNumber:       policy.GroupNumber,
				CoverageDetails:   policy.CoverageDetails,
			}
			if policy.InsuranceCompany != nil {
				ip.InsuranceCompanyName = &policy.InsuranceCompany.CompanyName
			}
			resp.InsurancePolicy = ip
		}
	}

	return resp, nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func (s *Service) blacklistUser(ctx context.Context, username string) {
	exp := time.Now().UTC().Add(lockDuration)
	var bl auditModel.LoginBlacklist
	if s.db.WithContext(ctx).Where("username = ?", username).First(&bl).Error == nil {
		s.db.WithContext(ctx).Model(&bl).Update("expiration_time", exp)
	} else {
		bl = auditModel.LoginBlacklist{Username: username, ExpirationTime: exp}
		s.db.WithContext(ctx).Create(&bl)
	}
}

func (s *Service) logoutUser(ctx context.Context, username string) {
	var token authModel.AccessToken
	if s.db.WithContext(ctx).Where("username = ?", username).First(&token).Error != nil {
		return
	}
	if token.AccessToken != nil {
		if jti, _, err := pkgAuth.ParseJTINoVerify(*token.AccessToken); err == nil {
			s.rdb.Set(ctx, jti, "blacklisted", blacklistTTL)
		}
	}
	if token.RefreshToken != nil {
		if jti, _, err := pkgAuth.ParseJTINoVerify(*token.RefreshToken); err == nil {
			s.rdb.Set(ctx, jti, "blacklisted", blacklistTTL)
		}
	}
	s.db.WithContext(ctx).Delete(&token)
}

func (s *Service) generateUniquePin(ctx context.Context, user *authModel.EmployeeLogin) (string, error) {
	const maxAttempts = 200
	for i := 0; i < maxAttempts; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(90000))
		if err != nil {
			return "", err
		}
		pin := fmt.Sprintf("%d", 10000+n.Int64())
		if err := user.SetExpressLogin(pin); err != nil {
			return "", err
		}
		if err := s.db.WithContext(ctx).Save(user).Error; err == nil {
			return pin, nil
		}
	}
	return "", errors.New("could not generate unique PIN (try again)")
}
