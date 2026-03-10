package auth_service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	authModel "sighthub-backend/internal/models/auth"
	"sighthub-backend/internal/models/audit"
	"sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/location"
	pkgAuth "sighthub-backend/pkg/auth"
	"sighthub-backend/pkg/crypto"
)

const (
	blacklistTTL      = 24 * time.Hour // TTL ключей blacklist в Redis
	lockDuration      = 1 * time.Minute
	maxFailedAttempts = 3
)

var (
	ErrBlacklisted    = errors.New("account is temporarily locked")
	ErrInactive       = errors.New("account is inactive")
	ErrInvalidCreds   = errors.New("invalid credentials")
	ErrInvalidPin     = errors.New("invalid PIN code")
	ErrNoRefreshToken = errors.New("refresh token missing")
	ErrExpiredRefresh = errors.New("expired refresh token")
	ErrRevoked        = errors.New("token has been revoked")
	ErrNotFound       = errors.New("user not found")
)

// BlacklistError — пользователь заблокирован, содержит оставшееся время в секундах
type BlacklistError struct {
	TimeRemaining int
}

func (e *BlacklistError) Error() string {
	return fmt.Sprintf("account locked for %d more seconds", e.TimeRemaining)
}

// LoginResult — результат успешного входа
type LoginResult struct {
	AccessToken  string
	RefreshToken string
}

// AuthService — сервис аутентификации
type AuthService struct {
	db  *gorm.DB
	rdb *redis.Client
	cfg *config.Config
}

func New(db *gorm.DB, rdb *redis.Client, cfg *config.Config) *AuthService {
	return &AuthService{db: db, rdb: rdb, cfg: cfg}
}

// ─── Public methods ───────────────────────────────────────────────────────────

// Login аутентифицирует пользователя по username + password.
func (s *AuthService) Login(ctx context.Context, username, password, ip, userAgent string) (*LoginResult, error) {
	username = strings.ToUpper(strings.TrimSpace(username))

	// Чистим просроченные блокировки
	audit.CleanupExpired(ctx, s.db)

	// Проверяем login blacklist
	var blEntry audit.LoginBlacklist
	blExists := s.db.WithContext(ctx).Where("username = ?", username).First(&blEntry).Error == nil
	if blExists {
		remaining := time.Until(blEntry.ExpirationTime)
		if remaining > 0 {
			return nil, &BlacklistError{TimeRemaining: int(remaining.Seconds())}
		}
		// Просрочена — удаляем
		s.db.WithContext(ctx).Delete(&blEntry)
		blExists = false
	}

	// Ищем пользователя
	var user authModel.EmployeeLogin
	if err := s.db.WithContext(ctx).Where("employee_login = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCreds
		}
		return nil, err
	}

	if !user.Active {
		return nil, ErrInactive
	}

	if user.CheckPassword(password) {
		// Успешный вход
		s.db.WithContext(ctx).Model(&user).Update("failed_attempts", 0)

		result, err := s.issueTokens(ctx, user.Username)
		if err != nil {
			return nil, err
		}

		s.writeAuditLog(ctx, &user, ip, userAgent, "MASTER", true)
		return result, nil
	}

	// Неверный пароль
	user.FailedAttempts++
	s.db.WithContext(ctx).Model(&user).Update("failed_attempts", user.FailedAttempts)

	if user.FailedAttempts >= maxFailedAttempts {
		exp := time.Now().UTC().Add(lockDuration)
		if blExists {
			s.db.WithContext(ctx).Model(&blEntry).Update("expiration_time", exp)
		} else {
			bl := audit.LoginBlacklist{Username: username, ExpirationTime: exp}
			s.db.WithContext(ctx).Create(&bl)
		}
	}

	s.writeAuditLog(ctx, &user, ip, userAgent, "MASTER", false)
	return nil, ErrInvalidCreds
}

// LoginWithPin аутентифицирует пользователя по 5-значному PIN.
func (s *AuthService) LoginWithPin(ctx context.Context, pin, ip, userAgent string) (*LoginResult, error) {
	audit.CleanupExpired(ctx, s.db)

	var user *authModel.EmployeeLogin

	// Быстрый путь: поиск по HMAC-digest (если PIN_PEPPER задан)
	if digest, err := crypto.PinDigest(pin); err == nil {
		var u authModel.EmployeeLogin
		if s.db.WithContext(ctx).Where("express_login_digest = ?", digest).First(&u).Error == nil {
			if u.CheckExpressLogin(pin) {
				user = &u
			}
		}
	}

	// Fallback: перебор всех пользователей с express_login
	if user == nil {
		var all []authModel.EmployeeLogin
		s.db.WithContext(ctx).Where("express_login IS NOT NULL").Find(&all)
		for i := range all {
			if all[i].CheckExpressLogin(pin) {
				user = &all[i]
				break
			}
		}
	}

	if user == nil {
		return nil, ErrInvalidPin
	}

	if !user.Active {
		return nil, ErrInactive
	}

	s.db.WithContext(ctx).Model(user).Update("failed_attempts", 0)

	result, err := s.issueTokens(ctx, user.Username)
	if err != nil {
		return nil, err
	}

	s.writeAuditLog(ctx, user, ip, userAgent, "EXPRESS PASS", true)
	return result, nil
}

// Refresh выдаёт новую пару токенов по refresh_token из cookie.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*LoginResult, error) {
	if refreshToken == "" {
		return nil, ErrNoRefreshToken
	}

	claims, err := pkgAuth.ParseToken(refreshToken, s.cfg.RefreshSecretKey)
	if err != nil {
		return nil, ErrExpiredRefresh
	}

	// Проверяем blacklist в Redis
	exists, _ := s.rdb.Exists(ctx, claims.ID).Result()
	if exists > 0 {
		return nil, ErrRevoked
	}

	return s.issueTokens(ctx, claims.Username)
}

// Logout добавляет оба токена в blacklist и удаляет запись из access_token.
func (s *AuthService) Logout(ctx context.Context, username string) error {
	var tokenEntry authModel.AccessToken
	if err := s.db.WithContext(ctx).Where("username = ?", username).First(&tokenEntry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	// Blacklist access token
	if tokenEntry.AccessToken != nil {
		if jti, _, err := pkgAuth.ParseJTINoVerify(*tokenEntry.AccessToken); err == nil {
			s.rdb.Set(ctx, jti, "blacklisted", blacklistTTL)
		}
	}

	// Blacklist refresh token
	if tokenEntry.RefreshToken != nil {
		if jti, _, err := pkgAuth.ParseJTINoVerify(*tokenEntry.RefreshToken); err == nil {
			s.rdb.Set(ctx, jti, "blacklisted", blacklistTTL)
		}
	}

	s.db.WithContext(ctx).Delete(&tokenEntry)
	return nil
}

// IsTokenBlacklisted проверяет, есть ли jti в Redis blacklist.
func (s *AuthService) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	exists, err := s.rdb.Exists(ctx, jti).Result()
	return exists > 0, err
}

// ─── Private helpers ──────────────────────────────────────────────────────────

// issueTokens генерирует новую пару токенов, блэклистит старые, обновляет access_token таблицу.
func (s *AuthService) issueTokens(ctx context.Context, username string) (*LoginResult, error) {
	// Блэклистим старые токены, если есть
	var old authModel.AccessToken
	if s.db.WithContext(ctx).Where("username = ?", username).First(&old).Error == nil {
		if old.AccessToken != nil {
			if jti, _, err := pkgAuth.ParseJTINoVerify(*old.AccessToken); err == nil {
				s.rdb.Set(ctx, jti, "blacklisted", blacklistTTL)
			}
		}
		if old.RefreshToken != nil {
			if jti, _, err := pkgAuth.ParseJTINoVerify(*old.RefreshToken); err == nil {
				s.rdb.Set(ctx, jti, "blacklisted", blacklistTTL)
			}
		}
	}

	accessToken, _, err := pkgAuth.GenerateAccessToken(username, s.cfg.JWTSecretKey)
	if err != nil {
		return nil, err
	}
	refreshToken, _, err := pkgAuth.GenerateRefreshToken(username, s.cfg.RefreshSecretKey)
	if err != nil {
		return nil, err
	}

	// Upsert в access_token таблице
	if old.IDAccessToken != 0 {
		old.UpdateAccessToken(accessToken, &refreshToken)
		s.db.WithContext(ctx).Save(&old)
	} else {
		entry := authModel.NewAccessToken(username, &accessToken, &refreshToken)
		s.db.WithContext(ctx).Create(entry)
	}

	return &LoginResult{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

// getEmployeeAndLocation возвращает сотрудника и локацию для данного login_id.
func (s *AuthService) getEmployeeAndLocation(ctx context.Context, loginID int) (*employees.Employee, *location.Location) {
	var emp employees.Employee
	if err := s.db.WithContext(ctx).Where("employee_login_id = ?", loginID).First(&emp).Error; err != nil {
		return nil, nil
	}
	if emp.LocationID == nil {
		return &emp, nil
	}
	var loc location.Location
	if s.db.WithContext(ctx).First(&loc, "id_location = ?", *emp.LocationID).Error != nil {
		return &emp, nil
	}
	return &emp, &loc
}

// writeAuditLog записывает попытку входа в login_audit.
func (s *AuthService) writeAuditLog(ctx context.Context, user *authModel.EmployeeLogin, ip, userAgent, method string, success bool) {
	emp, loc := s.getEmployeeAndLocation(ctx, user.IDEmployeeLogin)

	locationName := "Unknown Location"
	if loc != nil {
		locationName = loc.FullName
	}

	// Обрезаем до varchar(50)
	if len(userAgent) > 50 {
		userAgent = userAgent[:50]
	}
	if len(ip) > 50 {
		ip = ip[:50]
	}

	var empID *int
	if emp != nil {
		id := emp.IDEmployee
		empID = &id
	}

	log := &audit.LoginAudit{
		UserName:     user.Username,
		LocationName: &locationName,
		LoginMethod:  &method,
		IPAddress:    &ip,
		BrowserType:  &userAgent,
		LoginStatus:  success,
		PWStatus:     boolPtr(success),
		ActiveStatus: boolPtr(success),
		EmployeeID:   empID,
	}
	s.db.WithContext(ctx).Create(log)
}

func boolPtr(b bool) *bool { return &b }
