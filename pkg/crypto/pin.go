// pkg/crypto/pin.go
// Аналог pin_crypto.py + hash helpers из utils.py
package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
)

// PinDigest вычисляет HMAC-SHA256 PIN-кода с pepper из env PIN_PEPPER.
// Аналог Python: hmac.new(pepper, pin, sha256).hexdigest()
func PinDigest(pin string) (string, error) {
	pepper := os.Getenv("PIN_PEPPER")
	if pepper == "" {
		return "", fmt.Errorf("PIN_PEPPER is not set")
	}
	mac := hmac.New(sha256.New, []byte(pepper))
	mac.Write([]byte(pin))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// HashString возвращает SHA-256 hex-дайджест строки.
func HashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// VerifyHash сравнивает строку с её SHA-256 хешем.
func VerifyHash(s, hash string) bool {
	return HashString(s) == hash
}
