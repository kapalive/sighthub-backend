// internal/models/lenses/lenses_feature_relation.go
package lenses

import "fmt"

type LensesFeatureRelation struct {
	LensesID              int `gorm:"column:lenses_id;primaryKey"                    json:"lenses_id"`
	LensSpecialFeaturesID int `gorm:"column:lens_special_features_id;primaryKey"     json:"lens_special_features_id"`

	// (опционально) связи для Preload
	Lenses             *Lenses             `gorm:"foreignKey:LensesID;references:IDLenses"                          json:"-"`
	LensSpecialFeature *LensSpecialFeature `gorm:"foreignKey:LensSpecialFeaturesID;references:IDLensSpecialFeatures" json:"-"`
}

func (LensesFeatureRelation) TableName() string { return "lenses_feature_relation" }

func (r *LensesFeatureRelation) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"lenses_id":                r.LensesID,
		"lens_special_features_id": r.LensSpecialFeaturesID,
	}
}

func (r *LensesFeatureRelation) String() string {
	return fmt.Sprintf("<LensesFeatureRelation %d, %d>", r.LensesID, r.LensSpecialFeaturesID)
}
