// pkg/accounting/terms.go
// Аналог utils_accounting.py — расчёт периодов оплаты по terms
package accounting

import (
	"time"

	"github.com/shopspring/decimal"
)

var validTerms = map[int]bool{
	30: true, 60: true, 90: true, 120: true, 150: true, 180: true,
	210: true, 240: true, 270: true, 300: true, 330: true, 360: true,
}

// D2 округляет decimal до 2 знаков (ROUND_HALF_UP).
func D2(x decimal.Decimal) decimal.Decimal {
	return x.Round(2)
}

// BuildDueAmounts делит total на periods равных частей, последняя корректируется на остаток.
func BuildDueAmounts(total decimal.Decimal, periods int) []decimal.Decimal {
	if periods <= 1 {
		return []decimal.Decimal{D2(total)}
	}
	base := D2(total.Div(decimal.NewFromInt(int64(periods))))
	dues := make([]decimal.Decimal, periods)
	for i := 0; i < periods-1; i++ {
		dues[i] = base
	}
	last := D2(total.Sub(base.Mul(decimal.NewFromInt(int64(periods - 1)))))
	dues[periods-1] = last
	return dues
}

// AddMonths прибавляет N календарных месяцев к дате.
// 2025-01-31 + 1 → 2025-02-28
func AddMonths(d time.Time, months int) time.Time {
	return d.AddDate(0, months, 0)
}

// PeriodStatus описывает один период оплаты.
type PeriodStatus struct {
	PeriodNo            int
	PeriodsTotal        int
	DueDate             time.Time
	PeriodAmount        decimal.Decimal
	PaidInPeriod        decimal.Decimal
	RemainingInPeriod   decimal.Decimal
	HasAnyPaymentInPeriod bool
}

// Invoice — минимальный интерфейс для расчётов.
type Invoice interface {
	GetInvoiceDate() time.Time
	GetTerms() *int
	GetInvoiceAmount() decimal.Decimal
	GetOpenBalance() decimal.Decimal
}

// TermsPeriodStatuses возвращает срез периодов с данными о покрытии суммой.
func TermsPeriodStatuses(inv Invoice) []PeriodStatus {
	if inv == nil {
		return nil
	}
	terms := inv.GetTerms()
	if terms == nil {
		return nil
	}
	t := *terms
	if !validTerms[t] {
		return nil
	}
	periodsTotal := t / 30
	if periodsTotal <= 0 {
		return nil
	}

	total := D2(inv.GetInvoiceAmount())
	openBal := D2(inv.GetOpenBalance())
	if total.IsZero() || !openBal.IsPositive() {
		return nil
	}

	paidTotal := D2(total.Sub(openBal))
	dues := BuildDueAmounts(total, periodsTotal)

	remPaid := paidTotal
	out := make([]PeriodStatus, 0, periodsTotal)

	for i, due := range dues {
		periodNo := i + 1
		var paidInPeriod decimal.Decimal
		if remPaid.IsPositive() {
			paidInPeriod = D2(decimal.Min(remPaid, due))
		}
		remPaid = D2(remPaid.Sub(paidInPeriod))
		remaining := D2(due.Sub(paidInPeriod))

		out = append(out, PeriodStatus{
			PeriodNo:              periodNo,
			PeriodsTotal:          periodsTotal,
			DueDate:               AddMonths(inv.GetInvoiceDate(), periodNo),
			PeriodAmount:          D2(due),
			PaidInPeriod:          paidInPeriod,
			RemainingInPeriod:     remaining,
			HasAnyPaymentInPeriod: paidInPeriod.IsPositive(),
		})
	}
	return out
}

// NextTermsPayment возвращает первый период с непокрытым остатком.
func NextTermsPayment(inv Invoice) *PeriodStatus {
	for _, p := range TermsPeriodStatuses(inv) {
		if p.RemainingInPeriod.IsPositive() {
			cp := p
			return &cp
		}
	}
	return nil
}

// TermsNotifyInfo возвращает данные для уведомления (за days дней до due_date) или nil.
func TermsNotifyInfo(inv Invoice, days int, today time.Time) map[string]interface{} {
	if today.IsZero() {
		today = time.Now()
	}
	if days < 0 {
		days = 0
	}

	periods := TermsPeriodStatuses(inv)
	if len(periods) == 0 {
		return nil
	}

	total := D2(inv.GetInvoiceAmount())
	openBal := D2(inv.GetOpenBalance())
	if !total.IsPositive() || !openBal.IsPositive() {
		return nil
	}

	var current *PeriodStatus
	for i := range periods {
		if periods[i].RemainingInPeriod.IsPositive() {
			current = &periods[i]
			break
		}
	}
	if current == nil {
		return nil
	}

	notifyFrom := current.DueDate.AddDate(0, 0, -days)
	if today.Before(notifyFrom) {
		return nil
	}

	isLast := current.PeriodNo == current.PeriodsTotal
	var rule string

	if isLast {
		if !openBal.IsPositive() {
			return nil
		}
		rule = "FINAL_UNTIL_PAID_100"
	} else {
		if current.HasAnyPaymentInPeriod {
			return nil
		}
		rule = "NO_PAYMENT_IN_PERIOD"
	}

	return map[string]interface{}{
		"rule":                rule,
		"period_no":           current.PeriodNo,
		"periods_total":       current.PeriodsTotal,
		"due_date":            current.DueDate,
		"notify_from":         notifyFrom,
		"open_balance":        openBal,
		"period_amount":       current.PeriodAmount,
		"paid_in_period":      current.PaidInPeriod,
		"remaining_in_period": current.RemainingInPeriod,
	}
}
