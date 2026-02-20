package lenses

import "fmt"

type LensSafetyThickness struct {
	IDLensSafetyThickness int    `gorm:"column:id_lens_safety_thickness;primaryKey" json:"id_lens_safety_thickness"`
	SafetyThicknessName   string `gorm:"column:safety_thickness_name;type:varchar(100);not null" json:"safety_thickness_name"`
	Description           string `gorm:"column:description;type:text" json:"description"`
}

func (LensSafetyThickness) TableName() string {
	return "lens_safety_thickness"
}

func (l *LensSafetyThickness) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_lens_safety_thickness": l.IDLensSafetyThickness,
		"safety_thickness_name":    l.SafetyThicknessName,
		"description":              l.Description,
	}
}

func (l *LensSafetyThickness) String() string {
	return fmt.Sprintf("<LensSafetyThickness %s>", l.SafetyThicknessName)
}
