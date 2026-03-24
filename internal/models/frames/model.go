// internal/models/frames/model.go
package frames

import (
	"fmt"
	"sighthub-backend/internal/models/lenses"
	"sighthub-backend/internal/models/types"
)

type Model struct {
	IDModel           int64               `gorm:"column:id_model;primaryKey"              json:"id_model"`
	ProductID         int64               `gorm:"column:product_id;not null"              json:"product_id"`
	TitleVariant      string              `gorm:"column:title_variant;type:varchar(100);not null" json:"title_variant"`
	UPC               *string             `gorm:"column:upc;type:varchar(16)"             json:"upc,omitempty"`
	EAN               *string             `gorm:"column:ean;type:varchar(16)"             json:"ean,omitempty"`
	GTIN              *string             `gorm:"column:gtin;type:varchar(16)"            json:"gtin,omitempty"`
	LensType          *types.LensType     `gorm:"column:lens_type;type:varchar(50)"       json:"lens_type,omitempty"`
	LensMaterial      *types.LensMaterial `gorm:"column:lens_material;type:varchar(50)"   json:"lens_material,omitempty"`
	Mirror            bool                `gorm:"column:mirror;default:false"            json:"mirror"`
	BacksideAR        bool                `gorm:"column:backside_ar;default:false"       json:"backside_ar"`
	Sunglass          *bool               `gorm:"column:sunglass"                        json:"sunglass,omitempty"`
	Photo             *bool               `gorm:"column:photo"                           json:"photo,omitempty"`
	Polor             *bool               `gorm:"column:polor"                           json:"polor,omitempty"`
	LensColor         *string             `gorm:"column:lens_color;type:varchar(100)"    json:"lens_color,omitempty"`
	CategoryGlassesID *int64              `gorm:"column:category_glasses_id"             json:"category_glasses_id,omitempty"`
	SizeLensWidth     *string             `gorm:"column:size_lens_width;type:varchar(3)" json:"size_lens_width,omitempty"`
	SizeBridgeWidth   *string             `gorm:"column:size_bridge_width;type:varchar(3)" json:"size_bridge_width,omitempty"`
	SizeTempleLength  *string             `gorm:"column:size_temple_length;type:varchar(4)" json:"size_temple_length,omitempty"`
	TypeProducts      *types.TypeProducts `gorm:"column:type_products;type:varchar(50)"  json:"type_products,omitempty"`
	MaterialsFrame    *string             `gorm:"column:materials_frame"                 json:"materials_frame,omitempty"`
	MaterialsTemple   *string             `gorm:"column:materials_temple"                json:"materials_temple,omitempty"`
	Color             *string             `gorm:"column:color"                           json:"color,omitempty"`
	MfgNumber         *string             `gorm:"column:mfg_number;type:varchar(50)"     json:"mfg_number,omitempty"`
	MfrSerialNumber   *string             `gorm:"column:mfr_serial_number;type:varchar(100)" json:"mfr_serial_number,omitempty"`
	Accessories       *string             `gorm:"column:accessories;type:text"           json:"accessories,omitempty"`
	ColorTemplate     *string             `gorm:"column:color_template"                  json:"color_template,omitempty"`
	Shape             *string             `gorm:"column:shape"                           json:"shape,omitempty"`

	// Relationships
	Product         *Product                `gorm:"foreignKey:ProductID;references:IDProduct" json:"product,omitempty"`
	CategoryGlasses *lenses.CategoryGlasses `gorm:"foreignKey:CategoryGlassesID;references:IDCategoryGlasses" json:"category_glasses,omitempty"`
}

func (Model) TableName() string { return "model" }

func (m *Model) ToMap() map[string]interface{} {
	mapped := map[string]interface{}{
		"id_model":            m.IDModel,
		"title_variant":       m.TitleVariant,
		"upc":                 m.UPC,
		"ean":                 m.EAN,
		"gtin":                m.GTIN,
		"lens_type":           m.LensType,
		"lens_material":       m.LensMaterial,
		"mirror":              m.Mirror,
		"backside_ar":         m.BacksideAR,
		"sunglass":            m.Sunglass,
		"photo":               m.Photo,
		"polor":               m.Polor,
		"lens_color":          m.LensColor,
		"category_glasses_id": m.CategoryGlassesID,
		"size_lens_width":     m.SizeLensWidth,
		"size_bridge_width":   m.SizeBridgeWidth,
		"size_temple_length":  m.SizeTempleLength,
		"type_products":       m.TypeProducts,
		"materials_frame":     m.MaterialsFrame,
		"materials_temple":    m.MaterialsTemple,
		"color":               m.Color,
		"mfg_number":          m.MfgNumber,
		"mfr_serial_number":   m.MfrSerialNumber,
		"accessories":         m.Accessories,
		"color_template":      m.ColorTemplate,
		"shape":               m.Shape,
	}

	if m.Product != nil {
		mapped["product"] = m.Product.ToMap()
	} else {
		mapped["product"] = nil
	}

	if m.CategoryGlasses != nil {
		mapped["category_glasses"] = m.CategoryGlasses.ToMap()
	} else {
		mapped["category_glasses"] = nil
	}

	return mapped
}

func (m *Model) String() string {
	return fmt.Sprintf("<Model %s | %s | %s>", m.TitleVariant, m.LensType, m.MaterialsFrame)
}

// Implementing ModelInterface (No need to redefine ToMap again)
