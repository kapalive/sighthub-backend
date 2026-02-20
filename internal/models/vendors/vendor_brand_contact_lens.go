// internal/models/vendors/vendor_brand_contact_lens.go
package vendors

import "fmt"

type VendorBrandContactLens struct {
	IDVendorBrandContactLens int `gorm:"column:id_vendor_brand_contact_lens;primaryKey" json:"id_vendor_brand_contact_lens"`
	IDVendor                 int `gorm:"column:id_vendor;not null"                      json:"id_vendor"`
	IDBrandContactLens       int `gorm:"column:id_brand_contact_lens;not null"          json:"id_brand_contact_lens"`

	Vendor           *Vendor           `gorm:"foreignKey:IDVendor;references:IDVendor"                          json:"-"`
	BrandContactLens *BrandContactLens `gorm:"foreignKey:IDBrandContactLens;references:IDBrandContactLens"      json:"-"`
}

func (VendorBrandContactLens) TableName() string { return "vendor_brand_contact_lens" }

func (r *VendorBrandContactLens) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_vendor_brand_contact_lens": r.IDVendorBrandContactLens,
		"vendor_id":                    r.IDVendor,
	}
	if r.BrandContactLens != nil {
		m["brand_contact_lens"] = r.BrandContactLens.ToMap()
	} else {
		m["brand_contact_lens"] = nil
	}
	return m
}

func (r *VendorBrandContactLens) String() string {
	return fmt.Sprintf("<VendorBrandContactLens vendor=%d brand_contact_lens=%d>", r.IDVendor, r.IDBrandContactLens)
}
