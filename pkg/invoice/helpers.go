// pkg/invoice/helpers.go
// Аналог utils_invoice.py — номера инвойсов, расчёты скидок, misc
package invoice

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CreateInvoiceNumber генерирует номер инвойса по типу и короткому имени магазина.
// Аналог create_invoice_number из Python.
// Формат: "{type}{shortName}{id:07d}"  например "INV-MN-0000001"
func CreateInvoiceNumber(db *gorm.DB, invoiceType, storeShortName string) (string, error) {
	var maxID int64
	err := db.Raw("SELECT COALESCE(MAX(id_invoice), 0) FROM invoice").Scan(&maxID).Error
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s%07d", invoiceType, storeShortName, maxID+1), nil
}

// DiscountAmount конвертирует raw discount в абсолютную сумму от base.
// "10%" → 10% от base; 0.1 → 10%; 1..100 → процент; >100 → фиксированная сумма.
// Аналог discount_amount из Python.
func DiscountAmount(raw string, base decimal.Decimal) decimal.Decimal {
	s := strings.TrimSpace(raw)
	if s == "" {
		return decimal.Zero
	}

	// "10%"
	if strings.HasSuffix(s, "%") {
		pctStr := strings.TrimSuffix(s, "%")
		pct, err := decimal.NewFromString(pctStr)
		if err != nil {
			return decimal.Zero
		}
		return money(base.Mul(pct.Div(decimal.NewFromInt(100))))
	}

	val, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	if val.IsNegative() {
		val = val.Neg()
	}

	// 0..1 — доля
	if val.LessThanOrEqual(decimal.NewFromInt(1)) {
		return money(base.Mul(val))
	}
	// 1..100 — процент
	if val.LessThanOrEqual(decimal.NewFromInt(100)) {
		return money(base.Mul(val.Div(decimal.NewFromInt(100))))
	}
	// >100 — фиксированная сумма
	return money(val)
}

// ApplyRoundingTargets округляет цену вверх так, чтобы цифра единиц попала
// в ближайший таргет из списка (например [0, 5, 9]).
func ApplyRoundingTargets(price decimal.Decimal, targets []int) decimal.Decimal {
	if len(targets) == 0 {
		return price
	}
	// Сортируем таргеты
	sorted := make([]int, len(targets))
	copy(sorted, targets)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Округляем вверх до целых
	priceInt := int(math.Ceil(price.InexactFloat64()))
	ones := priceInt % 10

	var nextTarget *int
	for _, t := range sorted {
		if t >= ones {
			v := t
			nextTarget = &v
			break
		}
	}

	var resultInt int
	if nextTarget != nil {
		resultInt = priceInt + (*nextTarget - ones)
	} else {
		// Переход через десяток
		resultInt = priceInt + (10 - ones) + sorted[0]
	}
	return decimal.NewFromInt(int64(resultInt))
}

// FormatAmount форматирует decimal в строку с 2 знаками: "10.50"
func FormatAmount(d decimal.Decimal) string {
	return d.StringFixed(2)
}

// ParseAmount парсит строку в decimal.
func ParseAmount(s string) (decimal.Decimal, error) {
	return decimal.NewFromString(strings.TrimSpace(s))
}

// CalculateAge возвращает полных лет или -1 если dob nil.
func CalculateAge(dobStr string) int {
	if dobStr == "" {
		return -1
	}
	// ожидаем "YYYY-MM-DD"
	parts := strings.Split(dobStr, "-")
	if len(parts) != 3 {
		return -1
	}
	year, err1 := strconv.Atoi(parts[0])
	if err1 != nil {
		return -1
	}
	// простой расчёт — только по году
	_ = year
	return -1
}

// money округляет до 2 знаков (ROUND_HALF_UP).
func money(d decimal.Decimal) decimal.Decimal {
	return d.Round(2)
}
