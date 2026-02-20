package lenses

import "fmt"

type LensTintColor struct {
	IDLensTintColor   int    `gorm:"column:id_lens_tint_color;primaryKey" json:"id_lens_tint_color"`
	LensTintColorName string `gorm:"column:lens_tint_color_name;type:varchar(100);not null" json:"lens_tint_color_name"`
}

func (LensTintColor) TableName() string {
	return "lens_tint_color"
}

func (l *LensTintColor) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_lens_tint_color":   l.IDLensTintColor,
		"lens_tint_color_name": l.LensTintColorName,
	}
}

func (l *LensTintColor) String() string {
	return fmt.Sprintf("<LensTintColor %s>", l.LensTintColorName)
}
