package auth_repo

import (
	"errors"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/auth"
)

type EmployeeLoginRepo struct{ DB *gorm.DB }

func NewEmployeeLoginRepo(db *gorm.DB) *EmployeeLoginRepo {
	return &EmployeeLoginRepo{DB: db}
}

func (r *EmployeeLoginRepo) GetByUsername(username string) (*auth.EmployeeLogin, error) {
	var v auth.EmployeeLogin
	if err := r.DB.Where("employee_login = ?", username).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *EmployeeLoginRepo) GetByID(id int) (*auth.EmployeeLogin, error) {
	var v auth.EmployeeLogin
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

// GetByPinDigest — быстрый поиск по HMAC-digest PIN (для express login).
func (r *EmployeeLoginRepo) GetByPinDigest(digest string) (*auth.EmployeeLogin, error) {
	var v auth.EmployeeLogin
	if err := r.DB.Where("express_login_digest = ?", digest).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *EmployeeLoginRepo) Create(v *auth.EmployeeLogin) error {
	return r.DB.Create(v).Error
}

func (r *EmployeeLoginRepo) Save(v *auth.EmployeeLogin) error {
	return r.DB.Save(v).Error
}

func (r *EmployeeLoginRepo) IncrementFailedAttempts(username string) error {
	return r.DB.Model(&auth.EmployeeLogin{}).
		Where("employee_login = ?", username).
		UpdateColumn("failed_attempts", gorm.Expr("failed_attempts + 1")).Error
}

func (r *EmployeeLoginRepo) ResetFailedAttempts(username string) error {
	return r.DB.Model(&auth.EmployeeLogin{}).
		Where("employee_login = ?", username).
		Update("failed_attempts", 0).Error
}

func (r *EmployeeLoginRepo) SetActive(username string, active bool) error {
	return r.DB.Model(&auth.EmployeeLogin{}).
		Where("employee_login = ?", username).
		Update("active", active).Error
}
