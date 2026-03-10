package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AccessTokenTTL  = 6 * time.Hour
	RefreshTokenTTL = 9 * time.Hour
)

type contextKey string

const usernameCtxKey contextKey = "username"

// Claims — JWT claims, соответствует Python payload структуре
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateAccessToken создаёт access token с уникальным jti.
// Возвращает (tokenString, jti, error).
func GenerateAccessToken(username, secret string) (string, string, error) {
	jti := uuid.NewString()
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(AccessTokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, err := token.SignedString([]byte(secret))
	return str, jti, err
}

// GenerateRefreshToken создаёт refresh token с уникальным jti.
// Возвращает (tokenString, jti, error).
func GenerateRefreshToken(username, secret string) (string, string, error) {
	jti := uuid.NewString()
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(RefreshTokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, err := token.SignedString([]byte(secret))
	return str, jti, err
}

// ParseToken проверяет и парсит JWT токен.
func ParseToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// ParseJTINoVerify извлекает jti и username без проверки подписи.
// Используется при logout/blacklist старых или просроченных токенов.
func ParseJTINoVerify(tokenStr string) (jti, username string, err error) {
	p := jwt.NewParser()
	claims := &Claims{}
	_, _, err = p.ParseUnverified(tokenStr, claims)
	if err != nil {
		return "", "", err
	}
	if claims.ID == "" {
		return "", "", errors.New("token missing jti")
	}
	return claims.ID, claims.Username, nil
}

// ExtractTokenFromHeader извлекает Bearer токен из заголовка Authorization.
func ExtractTokenFromHeader(r *http.Request) string {
	h := r.Header.Get("Authorization")
	parts := strings.SplitN(h, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}

// UsernameFromContext возвращает username, установленный JWTMiddleware.
func UsernameFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(usernameCtxKey).(string); ok {
		return v
	}
	return ""
}

// WithUsername устанавливает username в context (используется middleware).
func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameCtxKey, username)
}
