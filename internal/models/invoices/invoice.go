// internal/models/invoices/invoice.go
package invoices

import (
	"fmt"
	"sighthub-backend/internal/models/general"
	"sighthub-backend/internal/models/insurance"
	"time"

	"sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/location"
	"sighthub-backend/internal/models/patients"
	"sighthub-backend/internal/models/types"
	"sighthub-backend/internal/models/vendors"
)

type Invoice struct {
	IDInvoice           int64                     `gorm:"column:id_invoice;primaryKey"                           json:"id_invoice"`
	NumberInvoice       string                    `gorm:"column:number_invoice;type:varchar(16);not null"         json:"number_invoice"`
	DateCreate          time.Time                 `gorm:"column:date_create;default:CURRENT_TIMESTAMP"             json:"date_create"`
	CreatedAt           time.Time                 `gorm:"column:created_at;default:CURRENT_TIMESTAMP"             json:"created_at"`
	PaymentMethodID     *int64                    `gorm:"column:payment_method_id"                                json:"payment_method_id,omitempty"`
	PaidInsuranceStatus *types.PaidInsuranceStatus `gorm:"column:paid_insurance_status"                           json:"paid_insurance_status,omitempty"`
	EmployeeID          *int64                    `gorm:"column:employee_id"                                      json:"employee_id,omitempty"`
	Notified            *string                   `gorm:"column:notified;type:text"                               json:"notes,omitempty"`
	PTBal               float64                   `gorm:"column:pt_bal;type:numeric(10,2)"                        json:"pt_bal"`
	Discount            *float64                  `gorm:"column:discount;type:numeric(10,2)"                      json:"discount,omitempty"`
	TotalAmount         float64                   `gorm:"column:total_amount;type:numeric(10,2)"                  json:"total_amount"`
	FinalAmount         float64                   `gorm:"column:final_amount;type:numeric(10,2)"                  json:"final_amount"`
	InsBal              float64                   `gorm:"column:ins_bal;type:numeric(10,2)"                       json:"ins_bal"`
	GiftCardBal         *float64                  `gorm:"column:gift_card_bal;type:numeric(10,2)"                 json:"gift_card_bal,omitempty"`
	Due                 float64                   `gorm:"column:due;type:numeric(10,2)"                           json:"due"`
	StatusInvoiceID     *int64                    `gorm:"column:status_invoice_id"                                json:"status_invoice_id,omitempty"`
	Quantity            int                       `gorm:"column:quantity"                                         json:"quantity"`
	Referral            *string                   `gorm:"column:referral;type:varchar(100)"                        json:"referral,omitempty"`
	ClassField          *string                   `gorm:"column:class;type:varchar(100)"                           json:"class_field,omitempty"`
	Reason              *string                   `gorm:"column:reason;type:varchar(100)"                          json:"reason,omitempty"`
	DoctorID            *int64                    `gorm:"column:doctor_id"                                        json:"doctor_id,omitempty"`
	LocationID          int64                     `gorm:"column:location_id"                                       json:"location_id"`
	ToLocationID        *int64                    `gorm:"column:to_location_id"                                   json:"to_location_id,omitempty"`
	PatientID           int64                     `gorm:"column:patient_id"                                        json:"patient_id"`
	VendorID            *int64                    `gorm:"column:vendor_id"                                         json:"vendor_id,omitempty"`
	InsurancePolicyID   *int64                    `gorm:"column:insurance_policy_id"                              json:"insurance_policy_id,omitempty"`
	Remake              bool                      `gorm:"column:remake;default:false"                              json:"remake"`
	TaxAmount           float64                   `gorm:"column:tax_amount;type:numeric(10,2);default:0.00"        json:"tax_amount"`
	Finalized           bool                      `gorm:"column:finalized;default:false"                           json:"finalized"`

	// --- relations (preload when needed) ---
	PaymentMethod   *general.PaymentMethod     `gorm:"foreignKey:PaymentMethodID;references:IDPaymentMethod" json:"-"`
	Employee        *employees.Employee        `gorm:"foreignKey:EmployeeID;references:IDEmployee"           json:"-"`
	Doctor          *employees.Employee        `gorm:"foreignKey:DoctorID;references:IDEmployee"             json:"-"`
	Patient         *patients.Patient          `gorm:"foreignKey:PatientID;references:IDPatient"             json:"-"`
	Location        *location.Location         `gorm:"foreignKey:LocationID;references:IDLocation"           json:"-"`
	Vendor          *vendors.Vendor            `gorm:"foreignKey:VendorID;references:IDVendor"               json:"-"`
	InsurancePolicy *insurance.InsurancePolicy `gorm:"foreignKey:InsurancePolicyID;references:IDInsurancePolicy" json:"-"`
	StatusInvoice   *StatusInvoice             `gorm:"foreignKey:StatusInvoiceID;references:IDStatusInvoice" json:"-"`
}

func (Invoice) TableName() string { return "invoice" }

func (i *Invoice) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_invoice":            i.IDInvoice,
		"number_invoice":        i.NumberInvoice,
		"date_create":           i.DateCreate,
		"created_at":            i.CreatedAt,
		"payment_method_id":     i.PaymentMethodID,
		"paid_insurance_status": i.PaidInsuranceStatus,
		"employee_id":           i.EmployeeID,
		"notified":              i.Notified,
		"pt_bal":                i.PTBal,
		"tax_amount":            i.TaxAmount,
		"discount":              i.Discount,
		"total_amount":          i.TotalAmount,
		"final_amount":          i.FinalAmount,
		"ins_bal":               i.InsBal,
		"gift_card_bal":         i.GiftCardBal,
		"due":                   i.Due,
		"status_invoice_id":     i.StatusInvoiceID,
		"quantity":              i.Quantity,
		"referral":              i.Referral,
		"class_field":           i.ClassField,
		"reason":                i.Reason,
		"doctor_id":             i.DoctorID,
		"location_id":           i.LocationID,
		"to_location_id":        i.ToLocationID,
		"patient_id":            i.PatientID,
		"vendor_id":             i.VendorID,
		"insurance_policy_id":   i.InsurancePolicyID,
		"remake":                i.Remake,
		"finalized":             i.Finalized,
	}
}

func (i *Invoice) String() string {
	return fmt.Sprintf("<Invoice %s>", i.NumberInvoice)
}

func (i *Invoice) CalculateTax() {
	// Example tax calculation (replace with actual business logic)
	taxRate := 0.10 // Assuming 10% tax rate for simplicity

	// Dereference i.Discount if it is not nil, otherwise use 0
	discount := 0.0
	if i.Discount != nil {
		discount = *i.Discount
	}

	// Calculate taxable total (assuming total amount is taxable minus any discount)
	taxableTotal := i.TotalAmount - discount

	// Calculate tax
	i.TaxAmount = taxableTotal * taxRate
	i.FinalAmount = i.TotalAmount + i.TaxAmount - discount
}
