package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Секретный ключ (нужно хранить в .env.go)
var JWTSecret = []byte(os.Getenv("JWT_SECRET_KEY"))

// JWTClaims структура для claims токена
type JWTClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Время жизни токенов
const TokenTTL = time.Hour * 8 // later set time.Minute * 20 // 20 minutes
const RefreshTokenTTL = time.Hour * 24 * 7

// generateTokens создает access и refresh токены
func GenerateTokens(email string) (string, string, error) {
	// Access token
	accessTokenClaims := jwt.MapClaims{
		"sub": email,
		"exp": time.Now().Add(TokenTTL).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString(JWTSecret)
	if err != nil {
		return "", "", err
	}

	// Refresh token
	refreshTokenClaims := jwt.MapClaims{
		"sub": email,
		"exp": time.Now().Add(RefreshTokenTTL).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString(JWTSecret)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// HashToken хеширует токен перед сохранением в БД
func HashToken(token string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	return string(hash), err
}

func HashEmail(email string) string {
	hash := sha256.Sum256([]byte(email))
	return hex.EncodeToString(hash[:])
}

// Вспомогательная функция извлечения токена
func ExtractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}

func ParseToken(tokenString string, secretKey []byte) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ExtractUserID достает userID (sub) из access token
func ExtractUserID(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		if claims.Subject == "" {
			return "", errors.New("token does not contain subject")
		}
		return claims.Subject, nil
	}

	return "", errors.New("invalid token")
}
