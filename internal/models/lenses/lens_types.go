// internal/models/lenses/lens_types.go
package lenses

import "fmt"

type LensType struct {
	IDLensType  int    `gorm:"column:id_lens_type;primaryKey" json:"id_lens_type"`
	TypeName    string `gorm:"column:type_name;type:varchar(100);not null" json:"type_name"`
	Description string `gorm:"column:description;type:text" json:"description"`
}

func (LensType) TableName() string {
	return "lens_types"
}

func (l *LensType) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_lens_type": l.IDLensType,
		"type_name":    l.TypeName,
		"description":  l.Description,
	}
}

func (l *LensType) String() string {
	return fmt.Sprintf("<LensType %s>", l.TypeName)
}
