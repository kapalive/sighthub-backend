// pkg/captcha/captcha.go
// Аналог utils_captcha.py — верификация Google reCAPTCHA
package captcha

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

const verifyURL = "https://www.google.com/recaptcha/api/siteverify"

type verifyResponse struct {
	Success bool     `json:"success"`
	Errors  []string `json:"error-codes"`
}

// Verify проверяет токен reCAPTCHA через Google API.
// Ключ читается из env RECAPTCHA_SECRET_KEY.
func Verify(token string) (bool, error) {
	secret := os.Getenv("RECAPTCHA_SECRET_KEY")
	if secret == "" {
		return false, fmt.Errorf("RECAPTCHA_SECRET_KEY is not set")
	}

	resp, err := http.PostForm(verifyURL, url.Values{
		"secret":   {secret},
		"response": {token},
	})
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var result verifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return false, err
	}
	return result.Success, nil
}
