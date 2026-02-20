// internal/models/lenses/lenses_v_codes_relation.go
package lenses

import "fmt"

type LensesVCodesRelation struct {
	LensesID     int `gorm:"column:lenses_id;primaryKey"      json:"lenses_id"`
	VCodesLensID int `gorm:"column:v_codes_lens_id;primaryKey" json:"v_codes_lens_id"`

	// optional preload relations
	Lens  *Lenses     `gorm:"foreignKey:LensesID;references:IDLenses"        json:"-"`
	VCode *VCodesLens `gorm:"foreignKey:VCodesLensID;references:IDVCodesLens" json:"-"`
}

func (LensesVCodesRelation) TableName() string { return "lenses_v_codes_relation" }

func (r *LensesVCodesRelation) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"lenses_id":       r.LensesID,
		"v_codes_lens_id": r.VCodesLensID,
	}
}

func (r *LensesVCodesRelation) String() string {
	return fmt.Sprintf("<LensesVCodesRelation LensID:%d VCodeID:%d>", r.LensesID, r.VCodesLensID)
}
