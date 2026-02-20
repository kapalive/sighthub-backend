// internal/models/lenses/lens_special_features.go
package lenses

import "fmt"

type LensSpecialFeature struct {
	IDLensSpecialFeatures int    `gorm:"column:id_lens_special_features;primaryKey" json:"id_lens_special_features"`
	FeatureName           string `gorm:"column:feature_name;type:varchar(100);not null" json:"feature_name"`
	Description           string `gorm:"column:description;type:text" json:"description"`
}

func (LensSpecialFeature) TableName() string {
	return "lens_special_features"
}

func (l *LensSpecialFeature) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_lens_special_features": l.IDLensSpecialFeatures,
		"feature_name":             l.FeatureName,
		"description":              l.Description,
	}
}

func (l *LensSpecialFeature) String() string {
	return fmt.Sprintf("<LensSpecialFeature %s>", l.FeatureName)
}
