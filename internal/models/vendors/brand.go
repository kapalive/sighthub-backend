// internal/models/vendors/brand.go
package vendors

import "fmt"

type Brand struct {
	IDBrand            int     `gorm:"column:id_brand;primaryKey"                 json:"id_brand"`
	BrandName          *string `gorm:"column:brand_name;type:varchar(100)"        json:"brand_name,omitempty"`
	ShortName          *string `gorm:"column:short_name;type:varchar(2)"          json:"short_name,omitempty"`
	ReturnPolicy       *string `gorm:"column:return_policy;type:text"             json:"return_policy,omitempty"`
	Note               *string `gorm:"column:note;type:text"                      json:"note,omitempty"`
	PrintModelOnTag    bool    `gorm:"column:print_model_on_tag;not null;default:true"  json:"print_model_on_tag"`
	PrintPriceOnTag    bool    `gorm:"column:print_price_on_tag;not null;default:true"  json:"print_price_on_tag"`
	Discount           *int    `gorm:"column:discount"                            json:"discount,omitempty"`
	Description        *string `gorm:"column:description;type:text"               json:"description,omitempty"`
	CanLookup          bool    `gorm:"column:can_lookup;not null;default:true"    json:"can_lookup"`
	TypeItemsOfBrandID *int    `gorm:"column:type_items_of_brand_id"              json:"type_items_of_brand_id,omitempty"`
}

func (Brand) TableName() string { return "brand" }

func (b *Brand) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_brand":               b.IDBrand,
		"brand_name":             b.BrandName,
		"short_name":             b.ShortName,
		"return_policy":          b.ReturnPolicy,
		"note":                   b.Note,
		"print_model_on_tag":     b.PrintModelOnTag,
		"print_price_on_tag":     b.PrintPriceOnTag,
		"discount":               b.Discount,
		"description":            b.Description,
		"can_lookup":             b.CanLookup,
		"type_items_of_brand_id": b.TypeItemsOfBrandID,
	}
}

func (b *Brand) String() string {
	name := ""
	if b.BrandName != nil {
		name = *b.BrandName
	}
	short := ""
	if b.ShortName != nil {
		short = *b.ShortName
	}
	return fmt.Sprintf("<Brand %s | %s>", name, short)
}
