// internal/models/lenses/lens_treatments.go
package lenses

import (
	"fmt"
	"time"
)

type LensTreatments struct {
	IDLensTreatments int64     `gorm:"column:id_lens_treatments;primaryKey"         json:"id_lens_treatments"`
	ItemNbr          string    `gorm:"column:item_nbr;type:varchar(50);not null"     json:"item_nbr"`
	Description      *string   `gorm:"column:description;type:text"                  json:"description,omitempty"`
	Price            *float64  `gorm:"column:price;type:numeric(10,2)"               json:"price,omitempty"`
	Cost             *float64  `gorm:"column:cost;type:numeric(10,2)"                json:"cost,omitempty"`
	VendorID         int       `gorm:"column:vendor_id;not null"                     json:"vendor_id"`
	VCodesLensID     *int      `gorm:"column:v_codes_lens_id"                        json:"v_codes_lens_id,omitempty"`
	CanLookup        bool      `gorm:"column:can_lookup;not null;default:true"       json:"can_lookup"`
	Source           *string   `gorm:"column:source;type:varchar(50)"                json:"source,omitempty"`
	VwCode           *string   `gorm:"column:vw_code;type:varchar(100)"             json:"-"`
	VwAdtID          *string   `gorm:"column:vw_adt_id;type:varchar(50)"            json:"-"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"             json:"created_at"`
	ModifiedAt       time.Time `gorm:"column:modified_at;autoUpdateTime"            json:"modified_at"`

	// --- relations ---
	VCodesLens      *VCodesLens             `gorm:"foreignKey:VCodesLensID;references:IDVCodesLens"    json:"-"`
	SpecialFeatures []LensSpecialFeature    `gorm:"many2many:treatments_feature_relation;foreignKey:IDLensTreatments;joinForeignKey:lens_treatments_id;references:IDLensSpecialFeatures;joinReferences:lens_special_features_id" json:"-"`
	VCodes          []VCodesLens            `gorm:"many2many:treatments_v_codes_relation;foreignKey:IDLensTreatments;joinForeignKey:lens_treatments_id;references:IDVCodesLens;joinReferences:v_codes_lens_id" json:"-"`
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
		"source":             l.Source,
	}

	if l.VCodesLens != nil {
		out["v_codes_lens"] = l.VCodesLens.ToMap()
	} else {
		out["v_codes_lens"] = nil
	}

	if len(l.SpecialFeatures) > 0 {
		sf := make([]map[string]interface{}, 0, len(l.SpecialFeatures))
		for _, f := range l.SpecialFeatures {
			sf = append(sf, f.ToMap())
		}
		out["special_features"] = sf
	} else {
		out["special_features"] = []map[string]interface{}{}
	}

	if len(l.VCodes) > 0 {
		vc := make([]map[string]interface{}, 0, len(l.VCodes))
		for _, v := range l.VCodes {
			vc = append(vc, v.ToMap())
		}
		out["v_codes"] = vc
	} else {
		out["v_codes"] = []map[string]interface{}{}
	}

	return out
}

func (l *LensTreatments) String() string {
	return fmt.Sprintf("<LensTreatments %d>", l.IDLensTreatments)
}
