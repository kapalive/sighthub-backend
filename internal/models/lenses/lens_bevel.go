package lenses

import "fmt"

type LensBevel struct {
	IDLensBevel   int    `gorm:"column:id_lens_bevel;primaryKey" json:"id_lens_bevel"`
	LensBevelName string `gorm:"column:lens_bevel_name;type:varchar(255);not null" json:"lens_bevel_name"`
	Description   string `gorm:"column:description;type:text" json:"description"`
}

func (LensBevel) TableName() string {
	return "lens_bevel"
}

func (l *LensBevel) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_lens_bevel":   l.IDLensBevel,
		"lens_bevel_name": l.LensBevelName,
		"description":     l.Description,
	}
}

func (l *LensBevel) String() string {
	return fmt.Sprintf("<LensBevel %s>", l.LensBevelName)
}
