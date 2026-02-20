// internal/models/lenses/price_book_lens.go
package lenses

// NAhuy ne nujno
import (
	"fmt"

	vendormodel "sighthub-backend/internal/models/vendors"
)

type PriceBookLens struct {
	IDPriceBookLens int64    `gorm:"column:id_price_book_lens;primaryKey"          json:"id_price_book_lens"`
	LensesID        int      `gorm:"column:lenses_id;not null"                     json:"lenses_id"`
	MaterialID      *int     `gorm:"column:material_id"                            json:"material_id,omitempty"`
	BrandLensID     *int     `gorm:"column:brand_lens_id"                          json:"brand_lens_id,omitempty"`
	LensSeriesID    *int     `gorm:"column:lens_series_id"                         json:"lens_series_id,omitempty"`
	LensTypeID      *int     `gorm:"column:lens_type_id"                           json:"lens_type_id,omitempty"`
	VCodesLensID    *int     `gorm:"column:v_codes_lens_id"                        json:"v_codes_lens_id,omitempty"`
	BasePrice       float64  `gorm:"column:base_price;type:numeric(10,2);not null" json:"base_price"`
	Discount        *float64 `gorm:"column:discount;type:numeric(10,2)"            json:"discount,omitempty"`
	FinalPrice      *float64 `gorm:"column:final_price;type:numeric(10,2)"         json:"final_price,omitempty"`

	// --- relations (preload when нужно) ---
	Lenses     *Lenses                `gorm:"foreignKey:LensesID;references:IDLenses"                     json:"-"`
	Material   *LensesMaterial        `gorm:"foreignKey:MaterialID;references:IDLensesMaterials"          json:"-"`
	Brand      *vendormodel.BrandLens `gorm:"foreignKey:BrandLensID;references:IDBrandLens"               json:"-"`
	LensSeries *LensSeries            `gorm:"foreignKey:LensSeriesID;references:IDLensSeries"             json:"-"`
	LensType   *LensType              `gorm:"foreignKey:LensTypeID;references:IDLensType"                 json:"-"`
	VCode      *VCodesLens            `gorm:"foreignKey:VCodesLensID;references:IDVCodesLens"             json:"-"`
}

func (PriceBookLens) TableName() string { return "price_book_lens" }

func (p *PriceBookLens) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_price_book_lens": p.IDPriceBookLens,
		"lenses_id":          p.LensesID,
		"material_id":        p.MaterialID,
		"brand_lens_id":      p.BrandLensID,
		"lens_series_id":     p.LensSeriesID,
		"lens_type_id":       p.LensTypeID,
		"v_codes_lens_id":    p.VCodesLensID,
		"base_price":         p.BasePrice,
		"discount":           p.Discount,
		"final_price":        p.FinalPrice,
	}
}

func (p *PriceBookLens) String() string {
	return fmt.Sprintf("<PriceBookLens %d>", p.IDPriceBookLens)
}
