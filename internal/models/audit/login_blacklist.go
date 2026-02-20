// internal/models/audit/login_blacklist.go
package audit

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// LoginBlacklist ⇄ login_blacklist
type LoginBlacklist struct {
	ID             int       `gorm:"column:id;primaryKey"                                       json:"id"`
	Username       string    `gorm:"column:username;type:varchar(150);not null;uniqueIndex"     json:"username"`
	ExpirationTime time.Time `gorm:"column:expiration_time;type:timestamptz;not null"           json:"expiration_time"`
}

func (LoginBlacklist) TableName() string { return "login_blacklist" }

// IsExpired — проверяем, истёк ли срок блокировки
func (b *LoginBlacklist) IsExpired() bool {
	return time.Now().UTC().After(b.ExpirationTime)
}

// AddToBlacklist — аналог @classmethod add_to_blacklist(username, duration_minutes=1)
func AddToBlacklist(ctx context.Context, db *gorm.DB, username string, durationMinutes int) (*LoginBlacklist, error) {
	if durationMinutes <= 0 {
		durationMinutes = 1
	}
	entry := &LoginBlacklist{
		Username:       username,
		ExpirationTime: time.Now().UTC().Add(time.Duration(durationMinutes) * time.Minute),
	}
	if err := db.WithContext(ctx).Create(entry).Error; err != nil {
		return nil, err
	}
	return entry, nil
}

// CleanupExpired — аналог @classmethod cleanup_expired()
func CleanupExpired(ctx context.Context, db *gorm.DB) (int64, error) {
	res := db.WithContext(ctx).
		Where("expiration_time < ?", time.Now().UTC()).
		Delete(&LoginBlacklist{})
	return res.RowsAffected, res.Error
}

func (b *LoginBlacklist) String() string {
	return fmt.Sprintf("<LoginBlacklist %s until %s>", b.Username, b.ExpirationTime.Format(time.RFC3339))
}
