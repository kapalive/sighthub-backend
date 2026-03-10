// internal/models/vendors/vendor_ap_invoice_item.go
package vendors

// VendorAPInvoiceItem ⇄ vendor_ap_invoice_item
type VendorAPInvoiceItem struct {
	IDVendorAPInvoiceItem int64   `gorm:"column:id_vendor_ap_invoice_item;primaryKey;autoIncrement" json:"id_vendor_ap_invoice_item"`
	VendorAPInvoiceID     int64   `gorm:"column:vendor_ap_invoice_id;not null"                      json:"vendor_ap_invoice_id"`
	LineNo                int     `gorm:"column:line_no;not null"                                   json:"line_no"`
	Quantity              string  `gorm:"column:quantity;type:numeric(10,2);not null;default:1"     json:"quantity"`
	Description           string  `gorm:"column:description;type:varchar(255);not null"             json:"description"`
	PriceEach             string  `gorm:"column:price_each;type:numeric(12,2);not null"             json:"price_each"`
	Amount                string  `gorm:"column:amount;type:numeric(12,2);not null"                 json:"amount"`
	Tax                   string  `gorm:"column:tax;type:numeric(12,2);not null;default:0"          json:"tax"`
}

func (VendorAPInvoiceItem) TableName() string { return "vendor_ap_invoice_item" }

func (v *VendorAPInvoiceItem) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_vendor_ap_invoice_item": v.IDVendorAPInvoiceItem,
		"vendor_ap_invoice_id":      v.VendorAPInvoiceID,
		"line_no":                   v.LineNo,
		"quantity":                  v.Quantity,
		"description":               v.Description,
		"price_each":                v.PriceEach,
		"amount":                    v.Amount,
		"tax":                       v.Tax,
	}
}
