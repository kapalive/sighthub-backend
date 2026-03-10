// pkg/activitylog/activitylog.go
// Аналог utils_activity_log.py — лёгкая запись действий сотрудников
package activitylog

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// Entry — минимальная структура для вставки в activity_log.
// Не импортируем internal/models чтобы избежать циклических зависимостей.
type Entry struct {
	Timestamp  time.Time       `gorm:"column:timestamp"`
	EmployeeID *int64          `gorm:"column:employee_id"`
	LocationID *int            `gorm:"column:location_id"`
	EntityType string          `gorm:"column:entity_type"`
	EntityID   *int64          `gorm:"column:entity_id"`
	Action     string          `gorm:"column:action"`
	Details    json.RawMessage `gorm:"column:details;type:jsonb"`
}

func (Entry) TableName() string { return "activity_log" }

// Log создаёт запись в activity_log в рамках переданной транзакции/сессии.
// НЕ вызывает Commit — ответственность на вызывающей стороне.
func Log(db *gorm.DB, entityType, action string, opts ...Option) error {
	cfg := &config{}
	for _, o := range opts {
		o(cfg)
	}

	entry := Entry{
		Timestamp:  time.Now().UTC(),
		EmployeeID: cfg.employeeID,
		LocationID: cfg.locationID,
		EntityType: entityType,
		Action:     action,
		EntityID:   cfg.entityID,
		Details:    cfg.details,
	}

	if err := db.Create(&entry).Error; err != nil {
		log.Printf("activitylog.Log failed: %v", err)
		return err
	}
	return nil
}

// DiffFields возвращает map[field]{"old": ..., "new": ...} для изменённых полей.
// Аналог diff_fields из Python.
func DiffFields(old, new map[string]interface{}) map[string]interface{} {
	changed := make(map[string]interface{})
	for key, newVal := range new {
		oldVal := old[key]
		oldStr := fmt.Sprintf("%v", oldVal)
		newStr := fmt.Sprintf("%v", newVal)
		if oldStr != newStr {
			changed[key] = map[string]interface{}{
				"old": nilIfNone(oldVal, oldStr),
				"new": nilIfNone(newVal, newStr),
			}
		}
	}
	return changed
}

func nilIfNone(v interface{}, s string) interface{} {
	if v == nil {
		return nil
	}
	return s
}

// --- Option pattern ---

type config struct {
	employeeID *int64
	locationID *int
	entityID   *int64
	details    json.RawMessage
}

type Option func(*config)

func WithEmployee(id int64) Option {
	return func(c *config) { c.employeeID = &id }
}

func WithLocation(id int) Option {
	return func(c *config) { c.locationID = &id }
}

func WithEntity(id int64) Option {
	return func(c *config) { c.entityID = &id }
}

func WithDetails(d map[string]interface{}) Option {
	return func(c *config) {
		b, err := json.Marshal(d)
		if err == nil {
			c.details = b
		}
	}
}

func WithDetailsRaw(raw json.RawMessage) Option {
	return func(c *config) { c.details = raw }
}
