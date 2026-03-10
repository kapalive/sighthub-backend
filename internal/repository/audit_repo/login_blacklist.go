package audit_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/audit"
)

type LoginBlacklistRepo struct{ DB *gorm.DB }

func NewLoginBlacklistRepo(db *gorm.DB) *LoginBlacklistRepo {
	return &LoginBlacklistRepo{DB: db}
}

func (r *LoginBlacklistRepo) GetByUsername(username string) (*audit.LoginBlacklist, error) {
	var v audit.LoginBlacklist
	if err := r.DB.Where("username = ?", username).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func (r *LoginBlacklistRepo) IsBlacklisted(username string) (bool, error) {
	v, err := r.GetByUsername(username)
	if err != nil || v == nil {
		return false, err
	}
	return !v.IsExpired(), nil
}

func (r *LoginBlacklistRepo) Add(username string, durationMinutes int) (*audit.LoginBlacklist, error) {
	if durationMinutes <= 0 {
		durationMinutes = 1
	}
	entry := &audit.LoginBlacklist{
		Username:       username,
		ExpirationTime: time.Now().UTC().Add(time.Duration(durationMinutes) * time.Minute),
	}
	if err := r.DB.Create(entry).Error; err != nil {
		return nil, err
	}
	return entry, nil
}

func (r *LoginBlacklistRepo) Remove(username string) error {
	return r.DB.Where("username = ?", username).Delete(&audit.LoginBlacklist{}).Error
}

func (r *LoginBlacklistRepo) CleanupExpired() (int64, error) {
	res := r.DB.Where("expiration_time < ?", time.Now().UTC()).Delete(&audit.LoginBlacklist{})
	return res.RowsAffected, res.Error
}
