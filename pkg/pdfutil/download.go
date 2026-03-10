// pkg/pdfutil/download.go
// Аналог download_pdf + get_patient_dob_parts + get_patient_full_address из utils_form.py
package pdfutil

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const maxPDFSize = 24 * 1024 * 1024 // 24 MB

// DownloadPDF скачивает PDF по URL во временный файл.
// Возвращает путь к временному файлу — вызывающий должен удалить его сам (os.Remove).
// Аналог download_pdf из Python.
func DownloadPDF(url string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("pdf download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("pdf download: status %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "" && !strings.Contains(ct, "application/pdf") {
		return "", fmt.Errorf("pdf download: content-type is %q, not PDF", ct)
	}

	if resp.ContentLength > maxPDFSize {
		return "", fmt.Errorf("pdf download: file size %d exceeds 24 MB limit", resp.ContentLength)
	}

	tmp, err := os.CreateTemp("", "*.pdf")
	if err != nil {
		return "", fmt.Errorf("pdf download: create temp file: %w", err)
	}

	n, err := io.Copy(tmp, io.LimitReader(resp.Body, maxPDFSize+1))
	tmp.Close()
	if err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("pdf download: write temp file: %w", err)
	}
	if n > maxPDFSize {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("pdf download: file exceeds 24 MB limit")
	}

	return tmp.Name(), nil
}

// DOBParts разбивает дату рождения на части mm/dd/yyyy.
// Аналог get_patient_dob_parts из Python.
func DOBParts(dob *time.Time) map[string]interface{} {
	if dob == nil || dob.IsZero() {
		return map[string]interface{}{
			"birth_mm": nil,
			"birth_dd": nil,
			"birth_yy": nil,
		}
	}
	return map[string]interface{}{
		"birth_mm": dob.Format("01"),
		"birth_dd": dob.Format("02"),
		"birth_yy": dob.Format("2006"),
	}
}

// FullAddress объединяет поля адреса пациента в строку.
// Аналог get_patient_full_address из Python.
func FullAddress(streetAddress, addressLine2 *string) string {
	parts := make([]string, 0, 2)
	if streetAddress != nil && *streetAddress != "" {
		parts = append(parts, *streetAddress)
	}
	if addressLine2 != nil && *addressLine2 != "" {
		parts = append(parts, *addressLine2)
	}
	return strings.Join(parts, " ")
}
