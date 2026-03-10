// internal/models/vendors/vendor_ap_invoice.go
package vendors

import "time"

// VendorAPInvoice ⇄ vendor_ap_invoice
type VendorAPInvoice struct {
	IDVendorAPInvoice         int64      `gorm:"column:id_vendor_ap_invoice;primaryKey;autoIncrement"     json:"id_vendor_ap_invoice"`
	VendorID                  int        `gorm:"column:vendor_id;not null"                                json:"vendor_id"`
	LocationID                int64      `gorm:"column:location_id;not null"                              json:"location_id"`
	EmployeeID                int64      `gorm:"column:employee_id;not null"                              json:"employee_id"`
	InvoiceNumber             string     `gorm:"column:invoice_number;type:varchar(30);not null"          json:"invoice_number"`
	InvoiceDate               time.Time  `gorm:"column:invoice_date;type:date;not null"                   json:"-"`
	BillDueDate               time.Time  `gorm:"column:bill_due_date;type:date;not null"                  json:"-"`
	InvoiceAmount             string     `gorm:"column:invoice_amount;type:numeric(12,2);not null"        json:"invoice_amount"`
	OpenBalance               string     `gorm:"column:open_balance;type:numeric(12,2);not null"          json:"open_balance"`
	TaxTotal                  string     `gorm:"column:tax_total;type:numeric(12,2);not null;default:0"   json:"tax_total"`
	Status                    string     `gorm:"column:status;type:varchar(20);not null;default:'Open'"   json:"status"`
	AttachmentURL             *string    `gorm:"column:attachment_url;type:varchar(255)"                  json:"attachment_url,omitempty"`
	Note                      *string    `gorm:"column:note;type:varchar(255)"                            json:"note,omitempty"`
	Terms                     *int       `gorm:"column:terms"                                             json:"terms,omitempty"`
	VendorLocationAccountID   *int64     `gorm:"column:vendor_location_account_id"                        json:"vendor_location_account_id,omitempty"`
	CreatedAt                 *time.Time `gorm:"column:created_at;type:timestamptz;default:now()"         json:"created_at,omitempty"`
}

func (VendorAPInvoice) TableName() string { return "vendor_ap_invoice" }

func (v *VendorAPInvoice) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_vendor_ap_invoice":      v.IDVendorAPInvoice,
		"vendor_id":                 v.VendorID,
		"location_id":               v.LocationID,
		"employee_id":               v.EmployeeID,
		"invoice_number":            v.InvoiceNumber,
		"invoice_date":              v.InvoiceDate.Format("2006-01-02"),
		"bill_due_date":             v.BillDueDate.Format("2006-01-02"),
		"invoice_amount":            v.InvoiceAmount,
		"open_balance":              v.OpenBalance,
		"tax_total":                 v.TaxTotal,
		"status":                    v.Status,
		"attachment_url":            v.AttachmentURL,
		"note":                      v.Note,
		"terms":                     v.Terms,
		"vendor_location_account_id": v.VendorLocationAccountID,
	}
	return m
}
