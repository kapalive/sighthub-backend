// internal/models/vendors/return_to_vendor_invoice.go
package vendors

import (
	"fmt"
	"time"

	"sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/service"
)

type ReturnToVendorInvoice struct {
	IDReturnToVendorInvoice int64     `gorm:"column:id_return_to_vendor_invoice;primaryKey" json:"id_return_to_vendor_invoice"`
	VendorID                int64     `gorm:"column:vendor_id;not null"                     json:"vendor_id"`
	CreatedDate             time.Time `gorm:"column:created_date;default:CURRENT_TIMESTAMP" json:"created_date"`
	CreditAmount            *float64  `gorm:"column:credit_amount;type:numeric(10,2)"       json:"credit_amount,omitempty"`
	EmployeeID              *int64    `gorm:"column:employee_id"                            json:"employee_id,omitempty"`
	PurchaseTotal           float64   `gorm:"column:purchase_total;type:numeric(10,2);default:0.00" json:"purchase_total"`
	Quantity                int       `gorm:"column:quantity;default:0;not null"            json:"quantity"`
	Status                  string    `gorm:"column:status;type:rtv_invoice_status;default:'Pending';not null" json:"status"`
	ARNumber                *string   `gorm:"column:ar_number;type:varchar(100)"            json:"ar_number,omitempty"`
	ShippingServiceID       *int      `gorm:"column:shipping_service_id"                    json:"shipping_service_id,omitempty"`
	TrackingNumber          *string   `gorm:"column:tracking_number;type:varchar(100)"      json:"tracking_number,omitempty"`
	ShippingNumber          *string   `gorm:"column:shipping_number;type:varchar(100)"      json:"shipping_number,omitempty"`
	Note                    *string   `gorm:"column:note;type:text"                         json:"note,omitempty"`

	// --- relationships ---
	Employee        *employees.Employee       `gorm:"foreignKey:EmployeeID;references:IDEmployee"                              json:"employee,omitempty"`
	Vendor          *Vendor                   `gorm:"foreignKey:VendorID;references:IDVendor"                                  json:"vendor,omitempty"`
	Items           []ReturnToVendorItem      `gorm:"foreignKey:ReturnToVendorInvoiceID;references:IDReturnToVendorInvoice"    json:"items"`
	Payments        []VendorReturnPayment     `gorm:"foreignKey:ReturnToVendorInvoiceID;references:IDReturnToVendorInvoice"    json:"payments,omitempty"`
	ShippingService *service.ShippingServices `gorm:"foreignKey:ShippingServiceID;references:IDShippingServices"               json:"shipping_service,omitempty"`
}

func (ReturnToVendorInvoice) TableName() string { return "return_to_vendor_invoice" }

func (r *ReturnToVendorInvoice) ToMap() map[string]interface{} {
	out := map[string]interface{}{
		"id_return_to_vendor_invoice": r.IDReturnToVendorInvoice,
		"vendor_id":                   r.VendorID,
		"purchase_total":              r.PurchaseTotal,
		"quantity":                    r.Quantity,
		"employee_id":                 r.EmployeeID,
		"created_date":                r.CreatedDate,
		"credit_amount":               r.CreditAmount,
		"status":                      r.Status,
		"ar_number":                   r.ARNumber,
		"shipping_service_id":         r.ShippingServiceID,
		"tracking_number":             r.TrackingNumber,
		"shipping_number":             r.ShippingNumber,
		"note":                        r.Note,
	}

	if r.Employee != nil {
		out["employee"] = r.Employee.ToMap()
	} else {
		out["employee"] = nil
	}
	if r.Vendor != nil {
		out["vendor"] = r.Vendor.ToMap()
	} else {
		out["vendor"] = nil
	}
	if len(r.Items) > 0 {
		out["items"] = r.Items
	}
	return out
}

func (r *ReturnToVendorInvoice) String() string {
	return fmt.Sprintf("<ReturnToVendorInvoice id=%d vendor_id=%d>", r.IDReturnToVendorInvoice, r.VendorID)
}
