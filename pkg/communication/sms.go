// pkg/communication/sms.go
// Аналог utils_communication.py — отправка SMS через Twilio
package communication

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var nonDigitRe = regexp.MustCompile(`\D`)

// FormatPhone приводит номер к международному формату "+XXXXXXXXXXX".
func FormatPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if strings.HasPrefix(phone, "+") {
		return phone
	}
	digits := nonDigitRe.ReplaceAllString(phone, "")
	return "+" + digits
}

// SMSResult возвращается из SendSMS.
type SMSResult struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code,omitempty"`
	Error      string `json:"error,omitempty"`
}

// SendSMS отправляет SMS через Twilio REST API.
// Читает TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN, TWILIO_PHONE_NUMBER из env.
func SendSMS(to, message string) SMSResult {
	sid := os.Getenv("TWILIO_ACCOUNT_SID")
	token := os.Getenv("TWILIO_AUTH_TOKEN")
	from := os.Getenv("TWILIO_PHONE_NUMBER")

	if sid == "" || token == "" || from == "" {
		return SMSResult{Error: "missing Twilio credentials"}
	}

	to = FormatPhone(to)

	apiURL := fmt.Sprintf(
		"https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", sid,
	)

	data := url.Values{
		"To":   {to},
		"From": {from},
		"Body": {message},
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return SMSResult{Error: err.Error()}
	}
	req.SetBasicAuth(sid, token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return SMSResult{Error: err.Error()}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 300 {
		return SMSResult{
			Error:      fmt.Sprintf("Twilio API error: %s", string(body)),
			StatusCode: resp.StatusCode,
		}
	}
	return SMSResult{Status: "accepted", StatusCode: resp.StatusCode}
}
