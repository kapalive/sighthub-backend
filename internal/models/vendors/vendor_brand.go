// internal/models/vendors/vendor_brand.go
package vendors

import "fmt"

type VendorBrand struct {
	IDVendorBrand int `gorm:"column:id_vendor_brand;primaryKey" json:"id_vendor_brand"`
	IDVendor      int `gorm:"column:id_vendor;not null"       json:"-"`
	IDBrand       int `gorm:"column:id_brand;not null"        json:"-"`

	Vendor *Vendor `gorm:"foreignKey:IDVendor;references:IDVendor" json:"-"`
	Brand  *Brand  `gorm:"foreignKey:IDBrand;references:IDBrand"   json:"-"`
}

func (VendorBrand) TableName() string { return "vendor_brand" }

func (vb *VendorBrand) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_vendor_brand": vb.IDVendorBrand,
	}
	if vb.Vendor != nil {
		m["vendor"] = vb.Vendor.ToMap()
	} else {
		m["vendor"] = nil
	}
	if vb.Brand != nil {
		m["brand"] = vb.Brand.ToMap()
	} else {
		m["brand"] = nil
	}
	return m
}

func (vb *VendorBrand) String() string {
	return fmt.Sprintf("<VendorBrand Vendor ID: %d, Brand ID: %d>", vb.IDVendor, vb.IDBrand)
}
