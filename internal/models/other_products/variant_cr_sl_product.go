// internal/models/other_products/variant_cr_sl_product.go
package other_products

import (
	"fmt"
)

type VariantCrSlProduct struct {
	IDVariantCrSlProduct int64    `gorm:"column:id_variant_cr_sl_product;primaryKey;autoIncrement" json:"id_variant_cr_sl_product"`
	Title                string   `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Size                 *string  `gorm:"column:size;type:varchar(50)" json:"size,omitempty"`
	Color                *string  `gorm:"column:color;type:varchar(50)" json:"color,omitempty"`
	Materials            *string  `gorm:"column:materials;type:varchar(255)" json:"materials,omitempty"`
	Gender               *string  `gorm:"column:gender;type:varchar(50)" json:"gender,omitempty"`
	UPC                  *string  `gorm:"column:upc;type:varchar(100)" json:"upc,omitempty"`
	Weight               *float64 `gorm:"column:weight;type:numeric" json:"weight,omitempty"`
	CrossSellProductsID  int64    `gorm:"column:cross_sell_products_id;not null" json:"cross_sell_products_id"`
}

func (VariantCrSlProduct) TableName() string { return "variant_cr_sl_product" }

func (v *VariantCrSlProduct) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_variant_cr_sl_product": v.IDVariantCrSlProduct,
		"title":                    v.Title,
		"size":                     v.Size,
		"color":                    v.Color,
		"materials":                v.Materials,
		"gender":                   v.Gender,
		"upc":                      v.UPC,
		"weight":                   v.Weight,
		"cross_sell_products_id":   v.CrossSellProductsID,
	}
}

func (v *VariantCrSlProduct) String() string {
	return fmt.Sprintf("<VariantCrSlProduct %s>", v.Title)
}
