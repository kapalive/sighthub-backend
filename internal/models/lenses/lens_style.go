// internal/models/lenses/lens_style.go
package lenses

import "fmt"

type LensStyle struct {
	IDLensStyle int    `gorm:"column:id_lens_style;primaryKey" json:"id_lens_style"`
	StyleName   string `gorm:"column:style_name;type:varchar(100);not null" json:"style_name"`
	Description string `gorm:"column:description;type:text" json:"description"`
}

func (LensStyle) TableName() string {
	return "lens_style"
}

func (l *LensStyle) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_lens_style": l.IDLensStyle,
		"style_name":    l.StyleName,
		"description":   l.Description,
	}
}

func (l *LensStyle) String() string {
	return fmt.Sprintf("<LensStyle %s>", l.StyleName)
}
