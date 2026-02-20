// internal/models/vendors/type_items_brand.go
package vendors

import "fmt"

type TypeItemsOfBrand struct {
	IDTypeItemsOfBrand int     `gorm:"column:id_type_items_of_brand;primaryKey" json:"id_type_items_of_brand"`
	TypeName           string  `gorm:"column:type_name;type:varchar(50);not null" json:"type_name"`
	Description        *string `gorm:"column:description;type:text"             json:"description,omitempty"`
}

func (TypeItemsOfBrand) TableName() string { return "type_items_of_brand" }

func (t *TypeItemsOfBrand) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_type_items_of_brand": t.IDTypeItemsOfBrand,
		"type_name":              t.TypeName,
		"description":            t.Description,
	}
}

func (t *TypeItemsOfBrand) String() string {
	return fmt.Sprintf("<TypeItemsOfBrand %s>", t.TypeName)
}
