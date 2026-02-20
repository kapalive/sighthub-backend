package lenses

import "fmt"

type CategoryGlasses struct {
	IDCategoryGlasses int    `gorm:"column:id_category_glasses;primaryKey" json:"id_category_glasses"`
	TitleCategory     string `gorm:"column:title_category;type:varchar(100);not null" json:"title_category"`
}

func (CategoryGlasses) TableName() string {
	return "category_glasses"
}

func (c *CategoryGlasses) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_category_glasses": c.IDCategoryGlasses,
		"title_category":      c.TitleCategory,
	}
}

func (c *CategoryGlasses) String() string {
	return fmt.Sprintf("<CategoryGlasses %s>", c.TitleCategory)
}
