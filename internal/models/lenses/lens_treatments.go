// internal/models/lenses/lens_treatments.go
package lenses

import (
	"fmt"
	"sighthub-backend/internal/models/vendors" // Импортируем только интерфейс
	"time"
)

type LensTreatments struct {
	IDLensTreatments int64     `gorm:"column:id_lens_treatments;primaryKey"         json:"id_lens_treatments"`
	ItemNbr          string    `gorm:"column:item_nbr;type:varchar(50);not null"     json:"item_nbr"`
	Description      *string   `gorm:"column:description;type:text"                  json:"description,omitempty"`
	Price            *float64  `gorm:"column:price;type:numeric(10,2)"               json:"price,omitempty"` // per 1 lens
	Cost             *float64  `gorm:"column:cost;type:numeric(10,2)"                json:"cost,omitempty"`
	VendorID         int       `gorm:"column:vendor_id;not null"                     json:"vendor_id"`
	VCodesLensID     *int      `gorm:"column:v_codes_lens_id"                        json:"v_codes_lens_id,omitempty"`
	CanLookup        bool      `gorm:"column:can_lookup;not null;default:true"       json:"can_lookup"`
	SRCoat           bool      `gorm:"column:sr_coat;not null;default:false"         json:"sr_coat"`
	UV               bool      `gorm:"column:uv;not null;default:false"              json:"uv"`
	AR               bool      `gorm:"column:ar;not null;default:false"              json:"ar"`
	Tint             bool      `gorm:"column:tint;not null;default:false"            json:"tint"`
	Photo            bool      `gorm:"column:photo;not null;default:false"           json:"photo"`
	Polar            bool      `gorm:"column:polar;not null;default:false"           json:"polar"`
	Drill            bool      `gorm:"column:drill;not null;default:false"           json:"drill"`
	HighIndex        bool      `gorm:"column:high_index;not null;default:false"      json:"high_index"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"             json:"created_at"`
	ModifiedAt       time.Time `gorm:"column:modified_at;autoUpdateTime"            json:"modified_at"`

	// --- relations (preload when needed) ---
	Vendor     vendors.VendorInterface `gorm:"foreignKey:VendorID;references:IDVendor"            json:"-"`
	VCodesLens *VCodesLens             `gorm:"foreignKey:VCodesLensID;references:IDVCodesLens"    json:"-"`
}

func (LensTreatments) TableName() string { return "lens_treatments" }

func (l *LensTreatments) ToMap() map[string]interface{} {
	out := map[string]interface{}{
		"id_lens_treatments": l.IDLensTreatments,
		"item_nbr":           l.ItemNbr,
		"description":        l.Description,
		"price":              l.Price,
		"cost":               l.Cost,
		"vendor_id":          l.VendorID,
		"v_codes_lens_id":    l.VCodesLensID,
		"can_lookup":         l.CanLookup,
		"sr_coat":            l.SRCoat,
		"uv":                 l.UV,
		"ar":                 l.AR,
		"tint":               l.Tint,
		"photo":              l.Photo,
		"polar":              l.Polar,
		"drill":              l.Drill,
		"high_index":         l.HighIndex,
		"created_at":         l.CreatedAt,
		"modified_at":        l.ModifiedAt,
	}

	if l.Vendor != nil {
		out["vendor"] = l.Vendor.ToMap()
	} else {
		out["vendor"] = nil
	}
	if l.VCodesLens != nil {
		out["v_codes_lens"] = l.VCodesLens.ToMap()
	} else {
		out["v_codes_lens"] = nil
	}

	return out
}

func (l *LensTreatments) String() string {
	return fmt.Sprintf("<LensTreatments %d>", l.IDLensTreatments)
}
