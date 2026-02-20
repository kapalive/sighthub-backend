package inventory

import (
	"fmt"
	"time"
)

// Модель для партии товара (Batch)
type Batch struct {
	IDBatch           int64     `gorm:"column:id_batch;primaryKey;autoIncrement" json:"id_batch"`
	LocationID        int64     `gorm:"column:location_id" json:"location_id"`
	BrandID           int64     `gorm:"column:brand_id" json:"brand_id"`
	Qty               int       `gorm:"column:qty" json:"qty"`
	Cost              float64   `gorm:"column:cost;type:numeric(10,2)" json:"cost"`
	EmployeeIDPrepBy  int64     `gorm:"column:employee_id_prep_by" json:"employee_id_prep_by"`
	EmployeeIDCreated int64     `gorm:"column:employee_id_created" json:"employee_id_created"`
	EmployeeIDUpdated int64     `gorm:"column:employee_id_updated" json:"employee_id_updated"`
	Notes             *string   `gorm:"column:notes" json:"notes,omitempty"`
	CreatedAt         time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName задает имя таблицы в базе данных
func (Batch) TableName() string {
	return "batch"
}

// ToMap превращает объект в карту для удобства работы с данными
func (b *Batch) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_batch":            b.IDBatch,
		"location_id":         b.LocationID,
		"brand_id":            b.BrandID,
		"qty":                 b.Qty,
		"cost":                b.Cost,
		"employee_id_prep_by": b.EmployeeIDPrepBy,
		"employee_id_created": b.EmployeeIDCreated,
		"employee_id_updated": b.EmployeeIDUpdated,
		"notes":               b.Notes,
		"created_at":          b.CreatedAt,
		"updated_at":          b.UpdatedAt,
	}
}

// String метод для печати объекта
func (b *Batch) String() string {
	return fmt.Sprintf("<Batch %d>", b.IDBatch)
}
