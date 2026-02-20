// internal/models/frames/product.go
package frames

import (
	"fmt"
	"sighthub-backend/internal/models/interfaces" // Import the interfaces package
)

type Product struct {
	IDProduct      int64                       `gorm:"column:id_product;primaryKey"                           json:"id_product"`
	TitleProduct   string                      `gorm:"column:title_product;type:varchar(150);not null"        json:"title_product"`
	BrandID        *int64                      `gorm:"column:brand_id"                                        json:"brand_id,omitempty"`
	VendorID       *int64                      `gorm:"column:vendor_id"                                       json:"vendor_id,omitempty"`
	ManufacturerID *int64                      `gorm:"column:manufacturer_id"                                 json:"manufacturer_id,omitempty"`
	TypeProduct    string                      `gorm:"column:type_product;type:varchar(50);not null;default:eyeglasses" json:"type_product"`
	Brand          interfaces.ProductInterface `gorm:"-" json:"brand,omitempty"`  // Use the interface here
	Vendor         interfaces.ProductInterface `gorm:"-" json:"vendor,omitempty"` // Use the interface here
}

func (Product) TableName() string { return "product" }

func (p *Product) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_product":    p.IDProduct,
		"title_product": p.TitleProduct,
		"type_product":  p.TypeProduct,
	}

	if p.BrandID != nil {
		m["brand_id"] = *p.BrandID
	} else {
		m["brand_id"] = nil
	}
	if p.VendorID != nil {
		m["vendor_id"] = *p.VendorID
	} else {
		m["vendor_id"] = nil
	}
	if p.ManufacturerID != nil {
		m["manufacturer_id"] = *p.ManufacturerID
	} else {
		m["manufacturer_id"] = nil
	}

	if p.Brand != nil {
		m["brand"] = p.Brand.ToMap()
	} else {
		m["brand"] = nil
	}
	if p.Vendor != nil {
		m["vendor"] = p.Vendor.ToMap()
	} else {
		m["vendor"] = nil
	}
	m["manufacturer"] = nil

	return m
}

func (p *Product) String() string {
	return fmt.Sprintf("<Product %s | BrandID:%v | VendorID:%v>", p.TitleProduct, p.BrandID, p.VendorID)
}
