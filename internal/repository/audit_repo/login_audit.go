package audit_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/audit"
)

type LoginAuditRepo struct{ DB *gorm.DB }

func NewLoginAuditRepo(db *gorm.DB) *LoginAuditRepo {
	return &LoginAuditRepo{DB: db}
}

func (r *LoginAuditRepo) GetByID(id int) (*audit.LoginAudit, error) {
	var v audit.LoginAudit
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *LoginAuditRepo) GetByUsername(username string, limit int) ([]audit.LoginAudit, error) {
	var items []audit.LoginAudit
	q := r.DB.Where("user_name = ?", username).Order("log_time DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	return items, q.Find(&items).Error
}

func (r *LoginAuditRepo) GetByEmployeeID(employeeID int) ([]audit.LoginAudit, error) {
	var items []audit.LoginAudit
	return items, r.DB.Where("employee_id = ?", employeeID).Order("log_time DESC").Find(&items).Error
}

func (r *LoginAuditRepo) GetByDateRange(from, to time.Time) ([]audit.LoginAudit, error) {
	var items []audit.LoginAudit
	return items, r.DB.Where("log_time BETWEEN ? AND ?", from, to).Order("log_time DESC").Find(&items).Error
}

func (r *LoginAuditRepo) Create(v *audit.LoginAudit) error {
	return r.DB.Create(v).Error
}
