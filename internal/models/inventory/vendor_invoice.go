package inventory

import (
	"fmt"
	"sighthub-backend/internal/models/interfaces"
	"time"
)

type VendorInvoice struct {
	IDVendorInvoice  int64      `gorm:"column:id_vendor_invoice;primaryKey" json:"id_vendor_invoice"`
	InvoiceNo        string     `gorm:"column:invoice_no;type:varchar(30)" json:"invoice_no"`
	InvoiceDate      *time.Time `gorm:"column:invoice_date" json:"invoice_date"`
	Quantity         int        `gorm:"column:quantity" json:"quantity"`
	SubTotal         *float64   `gorm:"column:sub_total;type:numeric(10,2)" json:"sub_total"`
	ShippingHandling *float64   `gorm:"column:shipping_handling;type:numeric(10,2)" json:"shipping_handling"`
	Tax              *float64   `gorm:"column:tax;type:numeric(10,2)" json:"tax"`
	InvoiceTotal     *float64   `gorm:"column:invoice_total;type:numeric(10,2)" json:"invoice_total"`
	OrderRef         string     `gorm:"column:order_ref;type:varchar(50)" json:"order_ref"`
	DiscountReceived *int       `gorm:"column:discount_received" json:"discount_received"`
	Note             *string    `gorm:"column:note;type:varchar(255)" json:"note"`

	// Relations via interfaces
	VendorID  int64                       `gorm:"column:vendor_id;not null" json:"vendor_id"`
	Vendor    interfaces.VendorInterface  `gorm:"-" json:"vendor,omitempty"` // Using VendorInterface
	InvoiceID int64                       `gorm:"column:invoice_id;not null" json:"invoice_id"`
	Invoice   interfaces.InvoiceInterface `gorm:"-" json:"invoice,omitempty"` // Using InvoiceInterface
}

func (VendorInvoice) TableName() string { return "vendor_invoice" }

func (v *VendorInvoice) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_vendor_invoice": v.IDVendorInvoice,
		"invoice_no":        v.InvoiceNo,
		"invoice_date":      v.InvoiceDate,
		"quantity":          v.Quantity,
		"sub_total":         v.SubTotal,
		"shipping_handling": v.ShippingHandling,
		"tax":               v.Tax,
		"invoice_total":     v.InvoiceTotal,
		"order_ref":         v.OrderRef,
		"discount_received": v.DiscountReceived,
		"note":              v.Note,
		"vendor_id":         v.VendorID,
		"invoice_id":        v.InvoiceID,
		"vendor":            v.Vendor.ToMap(),
		"invoice":           v.Invoice.ToMap(),
	}
}

func (v *VendorInvoice) String() string {
	return fmt.Sprintf("<VendorInvoice %s | %s>", v.InvoiceNo, v.Vendor.Name())
}
