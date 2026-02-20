package lenses

// Таблица: lenses_materials
// Поля:
// - id_lenses_materials: PK
// - material_name: NOT NULL varchar(100)
// - index: numeric(3,2) (nullable)
// - description: text (nullable)

type LensesMaterial struct {
	IDLensesMaterials int64    `gorm:"column:id_lenses_materials;primaryKey;autoIncrement" json:"id_lenses_materials"`
	MaterialName      string   `gorm:"column:material_name;type:varchar(100);not null"     json:"material_name"`
	Index             *float64 `gorm:"column:index;type:numeric(3,2)"                      json:"index,omitempty"`
	Description       *string  `gorm:"column:description;type:text"                        json:"description,omitempty"`
}

// TableName явно задаёт имя таблицы
func (LensesMaterial) TableName() string { return "lenses_materials" }
