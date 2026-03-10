// pkg/sku/sku.go
// Аналог SKU-хелперов из utils.py
package sku

import (
	"fmt"
	"regexp"
	"strings"
)

var nonDigits = regexp.MustCompile(`[^0-9]`)

// Normalize убирает пробелы, точки и слеши, дополняет нулями до 6 символов.
// "051/184" -> "051184", "51184" -> "051184"
func Normalize(raw string) string {
	cleaned := nonDigits.ReplaceAllString(raw, "")
	// Дополняем нулями до 6 символов слева
	for len(cleaned) < 6 {
		cleaned = "0" + cleaned
	}
	return cleaned
}

// Format форматирует нормализованный SKU в "XXX/YYY".
// Аналог format_sku / format_sku_for_display
func Format(sku string, sep ...string) (string, error) {
	sku = Normalize(sku)
	if len(sku) != 6 {
		return "", fmt.Errorf("SKU must contain 6 digits after normalization, got: %q", sku)
	}
	s := "/"
	if len(sep) > 0 {
		s = sep[0]
	}
	return sku[:3] + s + sku[3:], nil
}

// FormatLegacy принимает уже нормализованный 6-значный SKU и форматирует его.
func FormatLegacy(sku string) string {
	sku = strings.TrimSpace(sku)
	if len(sku) < 6 {
		sku = fmt.Sprintf("%06s", sku)
	}
	return sku[:3] + "/" + sku[3:]
}
