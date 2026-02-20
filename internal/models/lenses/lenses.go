// internal/models/lenses/lenses.go
package lenses

import (
	"fmt"

	vendormodel "sighthub-backend/internal/models/vendors"
)

type Lenses struct {
	IDLenses          int      `gorm:"column:id_lenses;primaryKey"                          json:"id_lenses"`
	LensName          string   `gorm:"column:lens_name;type:varchar(255);not null"          json:"lens_name"`
	LensSeriesID      *int     `gorm:"column:lens_series_id"                                json:"lens_series_id,omitempty"`
	LensTypeID        *int     `gorm:"column:lens_type_id"                                  json:"lens_type_id,omitempty"`
	LensesMaterialsID *int     `gorm:"column:lenses_materials_id"                           json:"lenses_materials_id,omitempty"`
	BrandLensID       *int     `gorm:"column:brand_lens_id"                                 json:"brand_lens_id,omitempty"`
	VendorID          *int     `gorm:"column:vendor_id"                                     json:"vendor_id,omitempty"`
	Description       *string  `gorm:"column:description;type:text"                         json:"description,omitempty"`
	Price             *float64 `gorm:"column:price;type:numeric(10,2)"                      json:"price,omitempty"`
	Cost              *float64 `gorm:"column:cost;type:numeric(10,2)"                       json:"cost,omitempty"`
	MFRNumber         *string  `gorm:"column:mfr_number;type:varchar(255)"                  json:"mfr_number,omitempty"`
	CanLookup         bool     `gorm:"column:can_lookup;not null;default:true"              json:"can_lookup"`

	// --- relations (load via Preload when нужно) ---
	LensSeries     *LensSeries            `gorm:"foreignKey:LensSeriesID;references:IDLensSeries"               json:"-"`
	LensType       *LensType              `gorm:"foreignKey:LensTypeID;references:IDLensType"                   json:"-"`
	LensesMaterial *LensesMaterial        `gorm:"foreignKey:LensesMaterialsID;references:IDLensesMaterials"     json:"-"`
	BrandLens      *vendormodel.BrandLens `gorm:"foreignKey:BrandLensID;references:IDBrandLens"                 json:"-"`
	Vendor         *vendormodel.Vendor    `gorm:"foreignKey:VendorID;references:IDVendor"                       json:"-"`
	// SpecialFeatures []LensSpecialFeature // TODO: many-to-many через `lenses_feature_relation` — назови точные колонки связки, добавлю.
}

func (Lenses) TableName() string { return "lenses" }

func (l *Lenses) ToMap() map[string]interface{} {
	out := map[string]interface{}{
		"id_lenses":           l.IDLenses,
		"lens_name":           l.LensName,
		"description":         l.Description,
		"price":               l.Price,
		"cost":                l.Cost,
		"mfr_number":          l.MFRNumber,
		"can_lookup":          l.CanLookup,
		"lens_series_id":      l.LensSeriesID,
		"lens_type_id":        l.LensTypeID,
		"lenses_materials_id": l.LensesMaterialsID,
		"brand_lens_id":       l.BrandLensID,
		"vendor_id":           l.VendorID,
	}

	// embedded dicts как в твоём Python .to_dict():
	if l.LensSeries != nil {
		out["lens_series"] = l.LensSeries.ToMap()
	} else {
		out["lens_series"] = nil
	}
	if l.LensType != nil {
		out["lens_type"] = l.LensType.ToMap()
	} else {
		out["lens_type"] = nil
	}
	if l.LensesMaterial != nil {
		out["lenses_material"] = l.LensesMaterial.ToMap()
	} else {
		out["lenses_material"] = nil
	}
	if l.BrandLens != nil {
		out["brand_lens"] = l.BrandLens.ToMap()
	} else {
		out["brand_lens"] = nil
	}
	if l.Vendor != nil {
		out["vendor"] = l.Vendor.ToMap()
	} else {
		out["vendor"] = nil
	}

	return out
}

func (l *Lenses) String() string {
	return fmt.Sprintf("<Lenses %s>", l.LensName)
}

// --- доп. метод для LensesMaterial (в исходнике его не было, но нужен для вложенного to_dict) ---
func (m *LensesMaterial) ToMap() map[string]interface{} {
	if m == nil {
		return nil
	}
	return map[string]interface{}{
		"id_lenses_materials": m.IDLensesMaterials,
		"material_name":       m.MaterialName,
		"index":               m.Index,
		"description":         m.Description,
	}
}
