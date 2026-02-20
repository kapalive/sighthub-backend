// internal/models/vendors/vendor_brand_lens.go
package vendors

import "fmt"

type VendorBrandLens struct {
	IDVendorBrandLens int `gorm:"column:id_vendor_brand_lens;primaryKey" json:"id_vendor_brand_lens"`
	IDVendor          int `gorm:"column:id_vendor;not null"            json:"id_vendor"`
	IDBrandLens       int `gorm:"column:id_brand_lens;not null"        json:"id_brand_lens"`

	Vendor    *Vendor    `gorm:"foreignKey:IDVendor;references:IDVendor"          json:"-"`
	BrandLens *BrandLens `gorm:"foreignKey:IDBrandLens;references:IDBrandLens"    json:"-"`
}

func (VendorBrandLens) TableName() string { return "vendor_brand_lens" }

func (r *VendorBrandLens) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_vendor_brand_lens": r.IDVendorBrandLens,
		"vendor_id":            r.IDVendor,
	}
	if r.BrandLens != nil {
		m["brand_lens"] = r.BrandLens.ToMap()
	} else {
		m["brand_lens"] = nil
	}
	return m
}

func (r *VendorBrandLens) String() string {
	return fmt.Sprintf("<VendorBrandLens vendor=%d brand_lens=%d>", r.IDVendor, r.IDBrandLens)
}
