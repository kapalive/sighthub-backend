package lenses

import "fmt"

type LensSampleColor struct {
	IDLensSampleColor   int    `gorm:"column:id_lens_sample_color;primaryKey" json:"id_lens_sample_color"`
	LensSampleColorName string `gorm:"column:lens_sample_color_name;type:varchar(100);not null" json:"lens_sample_color_name"`
}

func (LensSampleColor) TableName() string {
	return "lens_sample_color"
}

func (l *LensSampleColor) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_lens_sample_color":   l.IDLensSampleColor,
		"lens_sample_color_name": l.LensSampleColorName,
	}
}

func (l *LensSampleColor) String() string {
	return fmt.Sprintf("<LensSampleColor %s>", l.LensSampleColorName)
}
