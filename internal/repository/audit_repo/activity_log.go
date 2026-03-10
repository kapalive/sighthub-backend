package audit_repo

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/audit"
)

type ActivityLogRepo struct{ DB *gorm.DB }

func NewActivityLogRepo(db *gorm.DB) *ActivityLogRepo {
	return &ActivityLogRepo{DB: db}
}

func (r *ActivityLogRepo) GetByEntityType(entityType string, limit int) ([]audit.ActivityLog, error) {
	var items []audit.ActivityLog
	q := r.DB.Where("entity_type = ?", entityType).Order("timestamp DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	return items, q.Find(&items).Error
}

func (r *ActivityLogRepo) GetByEntity(entityType string, entityID int64) ([]audit.ActivityLog, error) {
	var items []audit.ActivityLog
	return items, r.DB.
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("timestamp DESC").
		Find(&items).Error
}

func (r *ActivityLogRepo) GetByEmployeeID(employeeID int64, from, to time.Time) ([]audit.ActivityLog, error) {
	var items []audit.ActivityLog
	return items, r.DB.
		Where("employee_id = ? AND timestamp BETWEEN ? AND ?", employeeID, from, to).
		Order("timestamp DESC").
		Find(&items).Error
}

func (r *ActivityLogRepo) Create(v *audit.ActivityLog) error {
	return r.DB.Create(v).Error
}

func (r *ActivityLogRepo) Log(employeeID *int64, locationID *int, entityType string, entityID *int64, action string, details interface{}) error {
	var raw json.RawMessage
	if details != nil {
		b, err := json.Marshal(details)
		if err != nil {
			return err
		}
		raw = b
	}
	v := audit.ActivityLog{
		EmployeeID: employeeID,
		LocationID: locationID,
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		Details:    raw,
	}
	return r.DB.Create(&v).Error
}
