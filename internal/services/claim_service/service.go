package claim_service

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	locModel "sighthub-backend/internal/models/location"
	invoiceModel "sighthub-backend/internal/models/invoices"
	patientModel "sighthub-backend/internal/models/patients"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ── helpers ─────────────────────────────────────────────────────────────────

func (s *Service) getEmployeeAndLocation(username string) (*empModel.Employee, *locModel.Location, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("username = ?", username).First(&login).Error; err != nil {
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

func (s *Service) sumPatientPayments(invoiceID int64) float64 {
	var total float64
	s.db.Model(&patientModel.PaymentHistory{}).
		Where("invoice_id = ? AND (payment_method_id IS NULL OR payment_method_id != 14)", invoiceID).
		Select("COALESCE(SUM(amount), 0)").Scan(&total)
	return total
}

func (s *Service) sumInsurancePayments(invoiceID int64) float64 {
	var total float64
	s.db.Raw(`SELECT COALESCE(SUM(amount::numeric), 0) FROM insurance_payment WHERE invoice_id = ?`, invoiceID).Scan(&total)
	return total
}

func (s *Service) recalcInvoice(inv *invoiceModel.Invoice) {
	var items []invoiceModel.InvoiceItemSale
	s.db.Where("invoice_id = ?", inv.IDInvoice).Find(&items)

	var totalAmount, ptBal, insBal float64
	for _, item := range items {
		totalAmount += item.Total
		if item.PtBalance != nil {
			ptBal += *item.PtBalance
		} else {
			ptBal += item.Total
		}
		if item.InsBalance != nil {
			insBal += *item.InsBalance
		}
	}

	inv.TotalAmount = totalAmount
	inv.PTBal = ptBal
	inv.InsBal = insBal

	discount := 0.0
	if inv.Discount != nil {
		discount = *inv.Discount
	}
	inv.FinalAmount = totalAmount - discount

	patientPaid := s.sumPatientPayments(inv.IDInvoice)
	insurancePaid := s.sumInsurancePayments(inv.IDInvoice)
	giftCardPaid := 0.0
	if inv.GiftCardBal != nil {
		giftCardPaid = *inv.GiftCardBal
	}
	inv.Due = inv.FinalAmount - patientPaid - insurancePaid - giftCardPaid
}

// errorStatus maps error strings to HTTP status codes.
func errorStatus(err error) int {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "not found"):
		return 404
	case strings.Contains(msg, "required"), strings.Contains(msg, "invalid"),
		strings.Contains(msg, "cannot"):
		return 400
	case strings.Contains(msg, "forbidden"), strings.Contains(msg, "finalized"):
		return 403
	default:
		return 500
	}
}
