// internal/models/vendors/brand_contact_lens.go
package vendors

import "fmt"

type BrandContactLens struct {
	IDBrandContactLens int               `gorm:"column:id_brand_contact_lens;primaryKey" json:"id_brand_contact_lens"`
	BrandName          string            `gorm:"column:brand_name;type:varchar(255);not null" json:"brand_name"`
	ShortName          *string           `gorm:"column:short_name;type:varchar(2)"        json:"short_name,omitempty"`
	Description        *string           `gorm:"column:description;type:text"             json:"description,omitempty"`
	ReturnPolicy       *string           `gorm:"column:return_policy;type:text"           json:"return_policy,omitempty"`
	Note               *string           `gorm:"column:note;type:text"                    json:"note,omitempty"`
	PrintModelOnTag    bool              `gorm:"column:print_model_on_tag;default:true"   json:"print_model_on_tag"`
	PrintPriceOnTag    bool              `gorm:"column:print_price_on_tag;default:true"   json:"print_price_on_tag"`
	Discount           int               `gorm:"column:discount"                          json:"discount"`
	CanLookup          bool              `gorm:"column:can_lookup;default:true"           json:"can_lookup"`
	TypeItemsOfBrandID *int              `gorm:"column:type_items_of_brand_id"            json:"type_items_of_brand_id,omitempty"`
	TypeItemsOfBrand   *TypeItemsOfBrand `gorm:"foreignKey:TypeItemsOfBrandID;references:IDTypeItemsOfBrand" json:"-"`
}

func (BrandContactLens) TableName() string { return "brand_contact_lens" }

func (b *BrandContactLens) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_brand_contact_lens": b.IDBrandContactLens,
		"brand_name":            b.BrandName,
		"short_name":            b.ShortName,
		"description":           b.Description,
		"return_policy":         b.ReturnPolicy,
		"note":                  b.Note,
		"print_model_on_tag":    b.PrintModelOnTag,
		"print_price_on_tag":    b.PrintPriceOnTag,
		"discount":              b.Discount,
		"can_lookup":            b.CanLookup,
	}
	if b.TypeItemsOfBrand != nil {
		m["type_items_of_brand"] = b.TypeItemsOfBrand.ToMap()
	} else {
		m["type_items_of_brand"] = nil
	}
	return m
}

func (b *BrandContactLens) String() string {
	return fmt.Sprintf("<BrandContactLens %s>", b.BrandName)
}
