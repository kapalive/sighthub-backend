// internal/models/audit/activity_log.go
package audit

import (
	"encoding/json"
	"time"
)

// ActivityLog ⇄ activity_log
type ActivityLog struct {
	ID         int64           `gorm:"column:id;primaryKey;autoIncrement"                         json:"id"`
	Timestamp  time.Time       `gorm:"column:timestamp;type:timestamptz;not null;default:now()"   json:"timestamp"`
	EmployeeID *int64          `gorm:"column:employee_id"                                         json:"employee_id,omitempty"`
	LocationID *int            `gorm:"column:location_id"                                         json:"location_id,omitempty"`
	EntityType string          `gorm:"column:entity_type;type:varchar(50);not null"               json:"entity_type"`
	EntityID   *int64          `gorm:"column:entity_id"                                           json:"entity_id,omitempty"`
	Action     string          `gorm:"column:action;type:varchar(50);not null"                    json:"action"`
	Details    json.RawMessage `gorm:"column:details;type:jsonb"                                  json:"details,omitempty"`
}

func (ActivityLog) TableName() string { return "activity_log" }

func (a *ActivityLog) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":          a.ID,
		"timestamp":   a.Timestamp.Format(time.RFC3339),
		"employee_id": a.EmployeeID,
		"location_id": a.LocationID,
		"entity_type": a.EntityType,
		"entity_id":   a.EntityID,
		"action":      a.Action,
		"details":     a.Details,
	}
}
