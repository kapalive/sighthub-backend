package report_accounting_service

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"gorm.io/gorm"

	"sighthub-backend/pkg/csvutil"
)

// ─── Constants ───────────────────────────────────────────────────────────────

var CategoryOrder = []string{
	"card", "installment", "fintech", "cash", "check",
	"gift", "insurance", "credit", "internal", "other",
}

var CategoryLabels = map[string]string{
	"card":        "Cards",
	"installment": "Installments",
	"fintech":     "Fintech",
	"cash":        "Cash",
	"check":       "Checks",
	"gift":        "Gift Cards",
	"insurance":   "Insurance",
	"credit":      "Patient Credit",
	"internal":    "Internal Transfers",
	"other":       "Other",
}

var weekdays = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ─── Helper types ────────────────────────────────────────────────────────────

type PaymentMethodInfo struct {
	ID        int
	Name      string
	ShortName *string
	Category  string
}

func (s *Service) buildPaymentMethodMaps() ([]PaymentMethodInfo, map[string][]PaymentMethodInfo, map[int]PaymentMethodInfo) {
	type row struct {
		IDPaymentMethod int
		MethodName      string
		ShortName       *string
		Category        *string
	}
	var rows []row
	s.db.Table("payment_method").Scan(&rows)

	var all []PaymentMethodInfo
	catMap := map[string][]PaymentMethodInfo{}
	byID := map[int]PaymentMethodInfo{}

	for _, r := range rows {
		cat := "other"
		if r.Category != nil && *r.Category != "" {
			cat = *r.Category
		}
		info := PaymentMethodInfo{ID: r.IDPaymentMethod, Name: r.MethodName, ShortName: r.ShortName, Category: cat}
		all = append(all, info)
		catMap[cat] = append(catMap[cat], info)
		byID[r.IDPaymentMethod] = info
	}
	return all, catMap, byID
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }
func round2(v float64) float64 { return math.Round(v*100) / 100 }

func dayStart(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}
func dayEnd(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 23, 59, 59, 999999999, d.Location())
}

func pct(part, total float64) float64 {
	if total == 0 {
		return 0
	}
	return round1(part / total * 100)
}

// ─── 1. Monthly Summary ─────────────────────────────────────────────────────

type MonthlySummaryResult struct {
	Data         []map[string]interface{} `json:"data"`
	Totals       map[string]float64       `json:"totals"`
	Percentages  map[string]float64       `json:"percentages"`
	CatLabels    map[string]string        `json:"category_labels"`
	Month        int                      `json:"month"`
	Year         int                      `json:"year"`
	LocationID   int                      `json:"location_id"`
}

func (s *Service) MonthlySummary(locationID, month, year int) (*MonthlySummaryResult, error) {
	now := time.Now().UTC()
	if month == 0 {
		month = int(now.Month())
	}
	if year == 0 {
		year = now.Year()
	}

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDayOfMonth := startDate.AddDate(0, 1, -1)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if lastDayOfMonth.After(today) {
		lastDayOfMonth = today
	}

	_, _, methodByID := s.buildPaymentMethodMaps()

	tsStart := dayStart(startDate)
	tsEnd := dayEnd(lastDayOfMonth)

	// Query A: invoice aggregates per day
	type invRow struct {
		Day          time.Time
		GrossSales   float64
		Interco      float64
		Tax          float64
		FinalSales   float64
		Returns      float64
		ArPatient    float64
		ArInsurance  float64
		ArInterco    float64
	}
	var invRows []invRow
	s.db.Raw(`
		SELECT DATE(i.created_at) AS day,
			COALESCE(SUM(CASE WHEN i.number_invoice NOT LIKE 'I%' THEN i.total_amount ELSE 0 END), 0) AS gross_sales,
			COALESCE(SUM(CASE WHEN i.number_invoice LIKE 'I%' THEN i.final_amount ELSE 0 END), 0) AS interco,
			COALESCE(SUM(CASE WHEN i.number_invoice NOT LIKE 'I%' THEN i.tax_amount ELSE 0 END), 0) AS tax,
			COALESCE(SUM(CASE WHEN i.number_invoice NOT LIKE 'I%' THEN i.final_amount ELSE 0 END), 0) AS final_sales,
			COALESCE(SUM(CASE WHEN i.number_invoice NOT LIKE 'I%' THEN COALESCE(ri.return_amount, 0) ELSE 0 END), 0) AS returns,
			COALESCE(SUM(CASE WHEN i.number_invoice NOT LIKE 'I%' THEN i.pt_bal ELSE 0 END), 0) AS ar_patient,
			COALESCE(SUM(CASE WHEN i.number_invoice NOT LIKE 'I%' THEN i.ins_bal ELSE 0 END), 0) AS ar_insurance,
			COALESCE(SUM(CASE WHEN i.number_invoice LIKE 'I%' THEN i.due ELSE 0 END), 0) AS ar_interco
		FROM invoice i
		LEFT JOIN return_invoices ri ON ri.invoice_id = i.id_invoice
		WHERE i.location_id = ? AND i.created_at BETWEEN ? AND ?
		GROUP BY DATE(i.created_at)
	`, locationID, tsStart, tsEnd).Scan(&invRows)

	invByDay := map[string]invRow{}
	for _, r := range invRows {
		invByDay[r.Day.Format("2006-01-02")] = r
	}

	// Query B: PaymentHistory per day per method
	type phRow struct {
		Day             time.Time
		PaymentMethodID int
		Amount          float64
	}
	var phRows []phRow
	s.db.Raw(`
		SELECT DATE(ph.payment_timestamp) AS day, ph.payment_method_id, SUM(ph.amount) AS amount
		FROM payment_history ph
		JOIN invoice i ON i.id_invoice = ph.invoice_id
		WHERE i.location_id = ? AND ph.payment_timestamp BETWEEN ? AND ?
		GROUP BY DATE(ph.payment_timestamp), ph.payment_method_id
	`, locationID, tsStart, tsEnd).Scan(&phRows)

	dayCat := map[string]map[string]float64{}
	for _, r := range phRows {
		key := r.Day.Format("2006-01-02")
		if dayCat[key] == nil {
			dayCat[key] = map[string]float64{}
		}
		m := methodByID[r.PaymentMethodID]
		dayCat[key][m.Category] += r.Amount
	}

	// Query C: InsurancePayment per day
	type insRow struct {
		Day         time.Time
		Payments    float64
		Adjustments float64
	}
	var insRows []insRow
	s.db.Raw(`
		SELECT DATE(ip.created_at) AS day,
			COALESCE(SUM(CASE WHEN ip.payment_type_id != 3 THEN ip.amount::numeric ELSE 0 END), 0) AS payments,
			COALESCE(SUM(CASE WHEN ip.payment_type_id = 3 THEN ip.amount::numeric ELSE 0 END), 0) AS adjustments
		FROM insurance_payment ip
		JOIN invoice i ON i.id_invoice = ip.invoice_id
		WHERE i.location_id = ? AND ip.created_at BETWEEN ? AND ?
		GROUP BY DATE(ip.created_at)
	`, locationID, tsStart, tsEnd).Scan(&insRows)

	insPayByDay := map[string]float64{}
	insAdjByDay := map[string]float64{}
	for _, r := range insRows {
		key := r.Day.Format("2006-01-02")
		insPayByDay[key] = r.Payments
		insAdjByDay[key] = r.Adjustments
	}

	// Assemble daily rows
	data := []map[string]interface{}{}
	totals := map[string]float64{}

	numDays := lastDayOfMonth.Day()
	for d := 1; d <= numDays; d++ {
		curDay := time.Date(year, time.Month(month), d, 0, 0, 0, 0, time.UTC)
		key := curDay.Format("2006-01-02")

		inv := invByDay[key]
		cats := dayCat[key]
		if cats == nil {
			cats = map[string]float64{}
		}

		net := inv.FinalSales - inv.Returns - inv.Tax

		catVals := map[string]float64{}
		for _, c := range CategoryOrder {
			catVals[c] = cats[c]
		}

		insPay := insPayByDay[key]
		insAdj := insAdjByDay[key]

		allPayments := insPay + insAdj
		for _, v := range catVals {
			allPayments += v
		}
		allAR := inv.ArPatient + inv.ArInsurance + inv.ArInterco
		balanceDiff := (allPayments + allAR) - (net + inv.Interco)

		row := map[string]interface{}{
			"date":            curDay.Format("01/02"),
			"date_iso":        curDay.Format("2006-01-02"),
			"weekday":         weekdays[curDay.Weekday()%7],
			"gross_sales":     round2(inv.GrossSales),
			"interco":         round2(inv.Interco),
			"net":             round2(net),
			"tax":             round2(inv.Tax),
			"returns":         round2(inv.Returns),
			"ins_payments":    round2(insPay),
			"ins_adjustments": round2(insAdj),
			"ar_patient":      round2(inv.ArPatient),
			"ar_insurance":    round2(inv.ArInsurance),
			"ar_interco":      round2(inv.ArInterco),
		}
		for _, c := range CategoryOrder {
			row[c] = round2(catVals[c])
		}
		row["balance_diff"] = round2(balanceDiff)
		data = append(data, row)

		// accumulate totals
		for _, k := range []string{"gross_sales", "interco", "net", "tax", "returns",
			"ins_payments", "ins_adjustments", "ar_patient", "ar_insurance", "ar_interco", "balance_diff"} {
			totals[k] += row[k].(float64)
		}
		for _, c := range CategoryOrder {
			totals[c] += catVals[c]
		}
	}

	// weekday mapping: Go uses Sunday=0, Python Monday=0
	// fix weekday calculation
	for _, row := range data {
		dateISO := row["date_iso"].(string)
		t, _ := time.Parse("2006-01-02", dateISO)
		wd := int(t.Weekday())
		// convert Sunday=0..Saturday=6 to Monday=0..Sunday=6
		pyWD := (wd + 6) % 7
		row["weekday"] = weekdays[pyWD]
	}

	// percentages
	totalCollected := 0.0
	for _, c := range CategoryOrder {
		totalCollected += totals[c]
	}
	totalCollected += totals["ins_payments"] + totals["ins_adjustments"]

	percentages := map[string]float64{}
	for _, c := range CategoryOrder {
		percentages[c] = pct(totals[c], totalCollected)
	}
	percentages["ins_payments"] = pct(totals["ins_payments"], totalCollected)
	percentages["ins_adjustments"] = pct(totals["ins_adjustments"], totalCollected)

	// round totals
	for k, v := range totals {
		totals[k] = round2(v)
	}

	return &MonthlySummaryResult{
		Data:        data,
		Totals:      totals,
		Percentages: percentages,
		CatLabels:   CategoryLabels,
		Month:       month,
		Year:        year,
		LocationID:  locationID,
	}, nil
}

// ─── 2. Daily Detail ─────────────────────────────────────────────────────────

type DailyDetailResult struct {
	Date           string                   `json:"date"`
	DateDisplay    string                   `json:"date_display"`
	Weekday        string                   `json:"weekday"`
	LocationID     int                      `json:"location_id"`
	GrossSales     float64                  `json:"gross_sales"`
	Net            float64                  `json:"net"`
	Tax            float64                  `json:"tax"`
	InsPayments    float64                  `json:"ins_payments"`
	InsAdjustments float64                  `json:"ins_adjustments"`
	ArPatient      float64                  `json:"ar_patient"`
	ArInsurance    float64                  `json:"ar_insurance"`
	ArInterco      float64                  `json:"ar_interco"`
	Groups         []map[string]interface{} `json:"groups"`
	GrandTotal     float64                  `json:"grand_total"`
}

func (s *Service) DailyDetail(locationID int, targetDate time.Time) (*DailyDetailResult, error) {
	_, catMap, _ := s.buildPaymentMethodMaps()

	ds := dayStart(targetDate)
	de := dayEnd(targetDate)

	// PaymentHistory per method
	type phRow struct {
		PaymentMethodID int
		Amount          float64
	}
	var phRows []phRow
	s.db.Raw(`
		SELECT ph.payment_method_id, SUM(ph.amount) AS amount
		FROM payment_history ph
		JOIN invoice i ON i.id_invoice = ph.invoice_id
		WHERE i.location_id = ? AND ph.payment_timestamp BETWEEN ? AND ?
		GROUP BY ph.payment_method_id
	`, locationID, ds, de).Scan(&phRows)

	methodAmounts := map[int]float64{}
	for _, r := range phRows {
		methodAmounts[r.PaymentMethodID] = r.Amount
	}

	// Insurance payments + adjustments
	type insAgg struct {
		Payments    float64
		Adjustments float64
	}
	var insRow insAgg
	s.db.Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN ip.payment_type_id != 3 THEN ip.amount::numeric ELSE 0 END), 0) AS payments,
			COALESCE(SUM(CASE WHEN ip.payment_type_id = 3 THEN ip.amount::numeric ELSE 0 END), 0) AS adjustments
		FROM insurance_payment ip
		JOIN invoice i ON i.id_invoice = ip.invoice_id
		WHERE i.location_id = ? AND ip.created_at BETWEEN ? AND ?
	`, locationID, ds, de).Scan(&insRow)

	// Invoice totals
	type invAgg struct {
		GrossSales  float64
		Net         float64
		Tax         float64
		ArPatient   float64
		ArInsurance float64
		ArInterco   float64
	}
	var invRow invAgg
	s.db.Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN number_invoice NOT LIKE 'I%' THEN total_amount ELSE 0 END), 0) AS gross_sales,
			COALESCE(SUM(CASE WHEN number_invoice NOT LIKE 'I%' THEN final_amount ELSE 0 END), 0) AS net,
			COALESCE(SUM(CASE WHEN number_invoice NOT LIKE 'I%' THEN tax_amount ELSE 0 END), 0) AS tax,
			COALESCE(SUM(CASE WHEN number_invoice NOT LIKE 'I%' THEN pt_bal ELSE 0 END), 0) AS ar_patient,
			COALESCE(SUM(CASE WHEN number_invoice NOT LIKE 'I%' THEN ins_bal ELSE 0 END), 0) AS ar_insurance,
			COALESCE(SUM(CASE WHEN number_invoice LIKE 'I%' THEN due ELSE 0 END), 0) AS ar_interco
		FROM invoice
		WHERE location_id = ? AND created_at BETWEEN ? AND ?
	`, locationID, ds, de).Scan(&invRow)

	grandTotal := insRow.Payments + insRow.Adjustments
	for _, v := range methodAmounts {
		grandTotal += v
	}

	// Group by category
	groups := []map[string]interface{}{}
	for _, cat := range CategoryOrder {
		catMethods := catMap[cat]
		items := []map[string]interface{}{}
		catSubtotal := 0.0

		for _, m := range catMethods {
			amt := methodAmounts[m.ID]
			items = append(items, map[string]interface{}{
				"id_payment_method": m.ID,
				"method_name":       m.Name,
				"short_name":        m.ShortName,
				"amount":            round2(amt),
				"percentage":        pct(amt, grandTotal),
			})
			catSubtotal += amt
		}
		if len(items) > 0 {
			groups = append(groups, map[string]interface{}{
				"category":       cat,
				"category_label": CategoryLabels[cat],
				"items":          items,
				"subtotal":       round2(catSubtotal),
				"subtotal_pct":   pct(catSubtotal, grandTotal),
			})
		}
	}

	// Insurance payments group
	if insRow.Payments != 0 {
		groups = append(groups, map[string]interface{}{
			"category":       "ins_payments",
			"category_label": "Insurance Payments",
			"items": []map[string]interface{}{{
				"id_payment_method": nil,
				"method_name":       "Insurance Payments",
				"short_name":        nil,
				"amount":            round2(insRow.Payments),
				"percentage":        pct(insRow.Payments, grandTotal),
			}},
			"subtotal":     round2(insRow.Payments),
			"subtotal_pct": pct(insRow.Payments, grandTotal),
		})
	}

	// Insurance adjustments group
	if insRow.Adjustments != 0 {
		groups = append(groups, map[string]interface{}{
			"category":       "ins_adjustments",
			"category_label": "Ins. Adjustments",
			"items": []map[string]interface{}{{
				"id_payment_method": nil,
				"method_name":       "Insurance Adjustments",
				"short_name":        nil,
				"amount":            round2(insRow.Adjustments),
				"percentage":        pct(insRow.Adjustments, grandTotal),
			}},
			"subtotal":     round2(insRow.Adjustments),
			"subtotal_pct": pct(insRow.Adjustments, grandTotal),
		})
	}

	wd := (int(targetDate.Weekday()) + 6) % 7

	return &DailyDetailResult{
		Date:           targetDate.Format("2006-01-02"),
		DateDisplay:    targetDate.Format("01/02/2006"),
		Weekday:        weekdays[wd],
		LocationID:     locationID,
		GrossSales:     round2(invRow.GrossSales),
		Net:            round2(invRow.Net),
		Tax:            round2(invRow.Tax),
		InsPayments:    round2(insRow.Payments),
		InsAdjustments: round2(insRow.Adjustments),
		ArPatient:      round2(invRow.ArPatient),
		ArInsurance:    round2(invRow.ArInsurance),
		ArInterco:      round2(invRow.ArInterco),
		Groups:         groups,
		GrandTotal:     round2(grandTotal),
	}, nil
}

// ─── 3. Payment Summary ─────────────────────────────────────────────────────

type PaymentSummaryResult struct {
	Data        []map[string]interface{} `json:"data"`
	GrandTotal  float64                  `json:"grand_total"`
	Date        string                   `json:"date"`
	DateDisplay string                   `json:"date_display"`
}

func (s *Service) PaymentSummary(locationIDs []int, targetDate time.Time) (*PaymentSummaryResult, error) {
	_, _, methodByID := s.buildPaymentMethodMaps()

	ds := dayStart(targetDate)
	de := dayEnd(targetDate)

	// location short names
	type locRow struct {
		IDLocation int
		ShortName  string
	}
	var locRows []locRow
	s.db.Table("location").Select("id_location, short_name").Where("id_location IN ?", locationIDs).Scan(&locRows)
	locMap := map[int]string{}
	for _, r := range locRows {
		locMap[r.IDLocation] = r.ShortName
	}

	// PaymentHistory grouped by location + method
	type phRow struct {
		LocationID      int
		PaymentMethodID int
		Amount          float64
	}
	var phRows []phRow
	s.db.Raw(`
		SELECT i.location_id, ph.payment_method_id, SUM(ph.amount) AS amount
		FROM payment_history ph
		JOIN invoice i ON i.id_invoice = ph.invoice_id
		WHERE i.location_id IN ? AND ph.payment_timestamp BETWEEN ? AND ?
		GROUP BY i.location_id, ph.payment_method_id
	`, locationIDs, ds, de).Scan(&phRows)

	// InsurancePayment grouped by location
	type insRow struct {
		LocationID int
		Amount     float64
	}
	var insRows []insRow
	s.db.Raw(`
		SELECT i.location_id, SUM(ip.amount::numeric) AS amount
		FROM insurance_payment ip
		JOIN invoice i ON i.id_invoice = ip.invoice_id
		WHERE i.location_id IN ? AND ip.created_at BETWEEN ? AND ?
		GROUP BY i.location_id
	`, locationIDs, ds, de).Scan(&insRows)

	dateDisplay := targetDate.Format("01/02/2006")
	grandTotal := 0.0
	data := []map[string]interface{}{}

	for _, r := range phRows {
		m := methodByID[r.PaymentMethodID]
		loc := locMap[r.LocationID]
		if loc == "" {
			loc = "??"
		}
		data = append(data, map[string]interface{}{
			"location":     loc,
			"location_id":  r.LocationID,
			"date":         dateDisplay,
			"payment_type": m.Name,
			"total":        round2(r.Amount),
		})
		grandTotal += r.Amount
	}

	for _, r := range insRows {
		if r.Amount == 0 {
			continue
		}
		loc := locMap[r.LocationID]
		if loc == "" {
			loc = "??"
		}
		data = append(data, map[string]interface{}{
			"location":     loc,
			"location_id":  r.LocationID,
			"date":         dateDisplay,
			"payment_type": "Insurance Payment",
			"total":        round2(r.Amount),
		})
		grandTotal += r.Amount
	}

	sort.Slice(data, func(i, j int) bool {
		li := data[i]["location"].(string)
		lj := data[j]["location"].(string)
		if li != lj {
			return li < lj
		}
		return data[i]["payment_type"].(string) < data[j]["payment_type"].(string)
	})

	return &PaymentSummaryResult{
		Data:        data,
		GrandTotal:  round2(grandTotal),
		Date:        targetDate.Format("2006-01-02"),
		DateDisplay: dateDisplay,
	}, nil
}

// ─── 4. Payment Details ─────────────────────────────────────────────────────

type PaymentDetailsResult struct {
	Data       []map[string]interface{} `json:"data"`
	TotalPtBal float64                  `json:"total_pt_bal"`
	TotalInsBal float64                 `json:"total_ins_bal"`
	Start      string                   `json:"start"`
	End        string                   `json:"end"`
}

func (s *Service) PaymentDetails(locationIDs []int, startDate, endDate time.Time) (*PaymentDetailsResult, error) {
	_, _, methodByID := s.buildPaymentMethodMaps()

	tsStart := dayStart(startDate)
	tsEnd := dayEnd(endDate)

	// PaymentHistory rows
	type phRow struct {
		CreatedAt       time.Time
		PaymentTS       time.Time
		NumberInvoice   string
		IDInvoice       int64
		PaymentMethodID int
		Amount          float64
		PtBal           float64
		InsBal          float64
	}
	var phRows []phRow
	s.db.Raw(`
		SELECT i.created_at, ph.payment_timestamp AS payment_ts, i.number_invoice,
			i.id_invoice, ph.payment_method_id, ph.amount, i.pt_bal, i.ins_bal
		FROM payment_history ph
		JOIN invoice i ON i.id_invoice = ph.invoice_id
		WHERE i.location_id IN ? AND ph.payment_timestamp BETWEEN ? AND ?
	`, locationIDs, tsStart, tsEnd).Scan(&phRows)

	// InsurancePayment rows
	type insRow struct {
		InvCreatedAt  time.Time
		PaymentAt     time.Time
		NumberInvoice string
		IDInvoice     int64
		Amount        float64
		PtBal         float64
		InsBal        float64
	}
	var insRows []insRow
	s.db.Raw(`
		SELECT i.created_at AS inv_created_at, ip.created_at AS payment_at,
			i.number_invoice, i.id_invoice, ip.amount::numeric AS amount, i.pt_bal, i.ins_bal
		FROM insurance_payment ip
		JOIN invoice i ON i.id_invoice = ip.invoice_id
		WHERE i.location_id IN ? AND ip.created_at BETWEEN ? AND ?
	`, locationIDs, tsStart, tsEnd).Scan(&insRows)

	type sortableRow struct {
		paymentDT time.Time
		invNum    string
		data      map[string]interface{}
	}

	var rows []sortableRow
	totalPtBal := 0.0
	totalInsBal := 0.0

	for _, r := range phRows {
		m := methodByID[r.PaymentMethodID]
		totalPtBal += r.Amount
		rows = append(rows, sortableRow{
			paymentDT: r.PaymentTS,
			invNum:    r.NumberInvoice,
			data: map[string]interface{}{
				"create_date":         r.CreatedAt.Format("01/02/2006"),
				"payment_date":        r.PaymentTS.Format("01/02/2006"),
				"number_invoice":      r.NumberInvoice,
				"invoice_id":          r.IDInvoice,
				"payment_description": m.Name,
				"pt_bal":              round2(r.Amount),
				"ins_bal":             0.0,
			},
		})
	}

	for _, r := range insRows {
		amt, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", r.Amount), 64)
		totalInsBal += amt
		rows = append(rows, sortableRow{
			paymentDT: r.PaymentAt,
			invNum:    r.NumberInvoice,
			data: map[string]interface{}{
				"create_date":         r.InvCreatedAt.Format("01/02/2006"),
				"payment_date":        r.PaymentAt.Format("01/02/2006"),
				"number_invoice":      r.NumberInvoice,
				"invoice_id":          r.IDInvoice,
				"payment_description": "Insurance Payment",
				"pt_bal":              0.0,
				"ins_bal":             round2(amt),
			},
		})
	}

	// sort: payment_date desc, invoice desc
	sort.Slice(rows, func(i, j int) bool {
		if !rows[i].paymentDT.Equal(rows[j].paymentDT) {
			return rows[i].paymentDT.After(rows[j].paymentDT)
		}
		return rows[i].invNum > rows[j].invNum
	})

	data := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		data[i] = r.data
	}

	return &PaymentDetailsResult{
		Data:        data,
		TotalPtBal:  round2(totalPtBal),
		TotalInsBal: round2(totalInsBal),
		Start:       startDate.Format("2006-01-02"),
		End:         endDate.Format("2006-01-02"),
	}, nil
}

// ─── 5. Payment Categories ──────────────────────────────────────────────────

type PaymentCategoriesResult struct {
	Data                  []map[string]interface{} `json:"data"`
	TotalPmt              float64                  `json:"total_pmt"`
	TotalInsurancePayment float64                  `json:"total_insurance_payment"`
	TotalPtPmt            float64                  `json:"total_pt_pmt"`
	Start                 string                   `json:"start"`
	End                   string                   `json:"end"`
}

func (s *Service) PaymentCategories(locationIDs []int, startDate, endDate time.Time,
	paymentTypeID *int, insuranceFilter string, insuranceCompanyID *int) (*PaymentCategoriesResult, error) {

	_, _, methodByID := s.buildPaymentMethodMaps()

	tsStart := dayStart(startDate)
	tsEnd := dayEnd(endDate)

	type sortableRow struct {
		paymentDT time.Time
		invNum    string
		data      map[string]interface{}
	}

	var rawRows []sortableRow
	totalPmt := 0.0
	totalInsPmt := 0.0
	totalPtPmt := 0.0
	empIDs := map[int64]struct{}{}

	// PaymentHistory (skip when "all_insurance")
	if insuranceFilter != "all_insurance" {
		where := `i.location_id IN ? AND ph.payment_timestamp BETWEEN ? AND ?
			AND i.number_invoice NOT LIKE 'I%'`
		args := []interface{}{locationIDs, tsStart, tsEnd}

		if paymentTypeID != nil {
			where += " AND ph.payment_method_id = ?"
			args = append(args, *paymentTypeID)
		}
		if insuranceCompanyID != nil {
			where += " AND ip2.insurance_company_id = ?"
			args = append(args, *insuranceCompanyID)
		}

		type phRow struct {
			PaymentTimestamp time.Time
			NumberInvoice   string
			IDInvoice       int64
			PtLast          *string
			PtFirst         *string
			CompanyName     *string
			PaymentMethodID *int
			Amount          float64
			EmployeeID      *int64
			PtBal           float64
			InsBal          float64
		}
		var phRows []phRow
		s.db.Raw(`
			SELECT ph.payment_timestamp, i.number_invoice, i.id_invoice,
				p.last_name AS pt_last, p.first_name AS pt_first,
				ic.company_name, ph.payment_method_id, ph.amount,
				ph.employee_id, i.pt_bal, i.ins_bal
			FROM payment_history ph
			JOIN invoice i ON i.id_invoice = ph.invoice_id
			JOIN patient p ON p.id_patient = i.patient_id
			LEFT JOIN insurance_policy ip2 ON ip2.id_insurance_policy = i.insurance_policy_id
			LEFT JOIN insurance_company ic ON ic.id_insurance_company = ip2.insurance_company_id
			LEFT JOIN payment_method pm ON pm.id_payment_method = ph.payment_method_id
			WHERE `+where, args...).Scan(&phRows)

		for _, r := range phRows {
			totalPtPmt += r.Amount
			totalPmt += r.Amount
			if r.EmployeeID != nil {
				empIDs[*r.EmployeeID] = struct{}{}
			}
			name := ""
			if r.PtLast != nil {
				name = *r.PtLast
				if r.PtFirst != nil {
					name += ", " + *r.PtFirst
				}
			}
			mName := "Unknown"
			if r.PaymentMethodID != nil {
				if m, ok := methodByID[*r.PaymentMethodID]; ok {
					mName = m.Name
				}
			}
			rawRows = append(rawRows, sortableRow{
				paymentDT: r.PaymentTimestamp,
				invNum:    r.NumberInvoice,
				data: map[string]interface{}{
					"pmt_date":          r.PaymentTimestamp.Format("01/02/2006"),
					"number_invoice":    r.NumberInvoice,
					"invoice_id":        r.IDInvoice,
					"name":              name,
					"insurance":         strVal(r.CompanyName),
					"ins_ck":            "",
					"payment_type":      mName,
					"tot_pmt":           round2(r.Amount),
					"insurance_payment": 0.0,
					"pt_pmt":            round2(r.Amount),
					"pt_bal":            round2(r.PtBal),
					"ins_bal":           round2(r.InsBal),
					"_emp_id":           r.EmployeeID,
				},
			})
		}
	}

	// InsurancePayment (skip when specific payment_type requested)
	if paymentTypeID == nil {
		where := `i.location_id IN ? AND ip.created_at BETWEEN ? AND ?
			AND i.number_invoice NOT LIKE 'I%'`
		args := []interface{}{locationIDs, tsStart, tsEnd}

		if insuranceCompanyID != nil {
			where += " AND ip2.insurance_company_id = ?"
			args = append(args, *insuranceCompanyID)
		}

		type insRow struct {
			CreatedAt       time.Time
			NumberInvoice   string
			IDInvoice       int64
			PtLast          *string
			PtFirst         *string
			CompanyName     *string
			ReferenceNumber *string
			Amount          float64
			EmployeeID      *int64
			PtBal           float64
			InsBal          float64
		}
		var insRows []insRow
		s.db.Raw(`
			SELECT ip.created_at, i.number_invoice, i.id_invoice,
				p.last_name AS pt_last, p.first_name AS pt_first,
				ic.company_name, ip.reference_number, ip.amount::numeric AS amount,
				ip.employee_id, i.pt_bal, i.ins_bal
			FROM insurance_payment ip
			JOIN invoice i ON i.id_invoice = ip.invoice_id
			JOIN patient p ON p.id_patient = i.patient_id
			JOIN insurance_policy ip2 ON ip2.id_insurance_policy = ip.insurance_policy_id
			JOIN insurance_company ic ON ic.id_insurance_company = ip2.insurance_company_id
			WHERE `+where, args...).Scan(&insRows)

		for _, r := range insRows {
			totalInsPmt += r.Amount
			totalPmt += r.Amount
			if r.EmployeeID != nil {
				empIDs[*r.EmployeeID] = struct{}{}
			}
			name := ""
			if r.PtLast != nil {
				name = *r.PtLast
				if r.PtFirst != nil {
					name += ", " + *r.PtFirst
				}
			}
			rawRows = append(rawRows, sortableRow{
				paymentDT: r.CreatedAt,
				invNum:    r.NumberInvoice,
				data: map[string]interface{}{
					"pmt_date":          r.CreatedAt.Format("01/02/2006"),
					"number_invoice":    r.NumberInvoice,
					"invoice_id":        r.IDInvoice,
					"name":              name,
					"insurance":         strVal(r.CompanyName),
					"ins_ck":            strVal(r.ReferenceNumber),
					"payment_type":      "Insurance Payment",
					"tot_pmt":           round2(r.Amount),
					"insurance_payment": round2(r.Amount),
					"pt_pmt":            0.0,
					"pt_bal":            round2(r.PtBal),
					"ins_bal":           round2(r.InsBal),
					"_emp_id":           r.EmployeeID,
				},
			})
		}
	}

	// Resolve employee names
	empMap := map[int64]string{}
	if len(empIDs) > 0 {
		ids := make([]int64, 0, len(empIDs))
		for id := range empIDs {
			ids = append(ids, id)
		}
		type empRow struct {
			IDEmployee int64
			LastName   string
			FirstName  string
		}
		var empRows []empRow
		s.db.Table("employee").Select("id_employee, last_name, first_name").
			Where("id_employee IN ?", ids).Scan(&empRows)
		for _, e := range empRows {
			empMap[e.IDEmployee] = e.LastName + ", " + e.FirstName
		}
	}

	for _, r := range rawRows {
		empID, _ := r.data["_emp_id"].(*int64)
		delete(r.data, "_emp_id")
		seller := ""
		if empID != nil {
			seller = empMap[*empID]
		}
		r.data["seller"] = seller
	}

	// sort: payment_date desc, invoice desc
	sort.Slice(rawRows, func(i, j int) bool {
		if !rawRows[i].paymentDT.Equal(rawRows[j].paymentDT) {
			return rawRows[i].paymentDT.After(rawRows[j].paymentDT)
		}
		return rawRows[i].invNum > rawRows[j].invNum
	})

	data := make([]map[string]interface{}, len(rawRows))
	for i, r := range rawRows {
		data[i] = r.data
	}

	return &PaymentCategoriesResult{
		Data:                  data,
		TotalPmt:              round2(totalPmt),
		TotalInsurancePayment: round2(totalInsPmt),
		TotalPtPmt:            round2(totalPtPmt),
		Start:                 startDate.Format("2006-01-02"),
		End:                   endDate.Format("2006-01-02"),
	}, nil
}

// ─── 6. Insurance Companies (lookup) ─────────────────────────────────────────

func (s *Service) InsuranceCompanies() ([]map[string]interface{}, error) {
	type row struct {
		ID   int64
		Name string
	}
	var rows []row
	if err := s.db.Table("insurance_company").
		Select("id_insurance_company AS id, company_name AS name").
		Order("company_name").Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		result[i] = map[string]interface{}{"id": r.ID, "name": r.Name}
	}
	return result, nil
}

// ─── 7. Payment Types (lookup) ───────────────────────────────────────────────

func (s *Service) PaymentTypes() ([]map[string]interface{}, error) {
	type row struct {
		ID   int
		Name string
	}
	var rows []row
	if err := s.db.Table("payment_method").
		Select("id_payment_method AS id, method_name AS name").
		Order("method_name").Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		result[i] = map[string]interface{}{"id": r.ID, "name": r.Name}
	}
	return result, nil
}

// ─── 8. AR Insurance Aging ──────────────────────────────────────────────────

type ARInsuranceAgingResult struct {
	Data     []map[string]interface{} `json:"data"`
	Totals   map[string]float64       `json:"totals"`
	Start    string                   `json:"start"`
	End      string                   `json:"end"`
	SearchBy string                   `json:"search_by"`
}

func (s *Service) ARInsuranceAging(locationIDs []int, startDate, endDate time.Time,
	searchBy string, insuranceCompanyID *int) (*ARInsuranceAgingResult, error) {

	if searchBy == "" {
		searchBy = "invoice_date"
	}

	tsStart := dayStart(startDate)
	tsEnd := dayEnd(endDate)

	agingKeys := []string{"current", "30_59", "60_89", "90_119", "120_149", "150_plus"}

	// Query 1: current ins_bal per insurance company
	balWhere := `i.location_id IN ? AND i.number_invoice NOT LIKE 'I%' AND i.ins_bal > 0`
	balArgs := []interface{}{locationIDs}

	if searchBy == "payment_date" {
		balWhere += ` AND i.id_invoice IN (
			SELECT DISTINCT invoice_id FROM insurance_payment
			WHERE created_at BETWEEN ? AND ?
		)`
		balArgs = append(balArgs, tsStart, tsEnd)
	} else {
		balWhere += " AND i.created_at BETWEEN ? AND ?"
		balArgs = append(balArgs, tsStart, tsEnd)
	}
	if insuranceCompanyID != nil {
		balWhere += " AND ic.id_insurance_company = ?"
		balArgs = append(balArgs, *insuranceCompanyID)
	}

	type balRow struct {
		IDInsuranceCompany int64
		CompanyName        string
		CurrentBal         float64
	}
	var balRows []balRow
	s.db.Raw(`
		SELECT ic.id_insurance_company, ic.company_name, SUM(i.ins_bal) AS current_bal
		FROM invoice i
		JOIN insurance_policy ip ON ip.id_insurance_policy = i.insurance_policy_id
		JOIN insurance_company ic ON ic.id_insurance_company = ip.insurance_company_id
		WHERE `+balWhere+`
		GROUP BY ic.id_insurance_company, ic.company_name
	`, balArgs...).Scan(&balRows)

	type companyAging struct {
		name    string
		buckets map[string]float64
	}
	companyData := map[int64]*companyAging{}
	for _, r := range balRows {
		companyData[r.IDInsuranceCompany] = &companyAging{
			name: r.CompanyName,
			buckets: map[string]float64{
				"current":  r.CurrentBal,
				"30_59":    0,
				"60_89":    0,
				"90_119":   0,
				"120_149":  0,
				"150_plus": 0,
			},
		}
	}

	if len(companyData) == 0 {
		emptyTotals := map[string]float64{"total_credits": 0}
		for _, k := range agingKeys {
			emptyTotals[k] = 0
		}
		return &ARInsuranceAgingResult{
			Data:     []map[string]interface{}{},
			Totals:   emptyTotals,
			Start:    startDate.Format("2006-01-02"),
			End:      endDate.Format("2006-01-02"),
			SearchBy: searchBy,
		}, nil
	}

	// Query 2: insurance payments – bucket by age
	pmtWhere := `i.location_id IN ? AND i.number_invoice NOT LIKE 'I%' AND i.ins_bal > 0`
	pmtArgs := []interface{}{locationIDs}

	if searchBy == "payment_date" {
		pmtWhere += " AND ip2.created_at BETWEEN ? AND ?"
		pmtArgs = append(pmtArgs, tsStart, tsEnd)
	} else {
		pmtWhere += " AND i.created_at BETWEEN ? AND ?"
		pmtArgs = append(pmtArgs, tsStart, tsEnd)
	}
	if insuranceCompanyID != nil {
		pmtWhere += " AND ic.id_insurance_company = ?"
		pmtArgs = append(pmtArgs, *insuranceCompanyID)
	}

	type pmtRow struct {
		IDInsuranceCompany int64
		Amount             float64
		PmtAt              time.Time
		InvAt              time.Time
	}
	var pmtRows []pmtRow
	s.db.Raw(`
		SELECT ic.id_insurance_company, ip2.amount::numeric AS amount,
			ip2.created_at AS pmt_at, i.created_at AS inv_at
		FROM insurance_payment ip2
		JOIN invoice i ON i.id_invoice = ip2.invoice_id
		JOIN insurance_policy ip ON ip.id_insurance_policy = ip2.insurance_policy_id
		JOIN insurance_company ic ON ic.id_insurance_company = ip.insurance_company_id
		WHERE `+pmtWhere, pmtArgs...).Scan(&pmtRows)

	for _, r := range pmtRows {
		cd, ok := companyData[r.IDInsuranceCompany]
		if !ok {
			continue
		}
		age := int(r.PmtAt.Sub(r.InvAt).Hours() / 24)
		switch {
		case age >= 150:
			cd.buckets["150_plus"] += r.Amount
		case age >= 120:
			cd.buckets["120_149"] += r.Amount
		case age >= 90:
			cd.buckets["90_119"] += r.Amount
		case age >= 60:
			cd.buckets["60_89"] += r.Amount
		case age >= 30:
			cd.buckets["30_59"] += r.Amount
		}
	}

	// Build response
	type sortEntry struct {
		cid  int64
		name string
	}
	var sorted []sortEntry
	for cid, cd := range companyData {
		sorted = append(sorted, sortEntry{cid, cd.name})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].name < sorted[j].name })

	totals := map[string]float64{"total_credits": 0}
	for _, k := range agingKeys {
		totals[k] = 0
	}

	data := []map[string]interface{}{}
	for _, se := range sorted {
		cd := companyData[se.cid]
		totalCredits := 0.0
		row := map[string]interface{}{
			"insurance_company_id": se.cid,
			"insurance":            cd.name,
		}
		for _, k := range agingKeys {
			row[k] = round2(cd.buckets[k])
			totals[k] += cd.buckets[k]
			totalCredits += cd.buckets[k]
		}
		row["total_credits"] = round2(totalCredits)
		totals["total_credits"] += totalCredits
		data = append(data, row)
	}

	for k, v := range totals {
		totals[k] = round2(v)
	}

	return &ARInsuranceAgingResult{
		Data:     data,
		Totals:   totals,
		Start:    startDate.Format("2006-01-02"),
		End:      endDate.Format("2006-01-02"),
		SearchBy: searchBy,
	}, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ─── CSV Exports ─────────────────────────────────────────────────────────────

// MonthlySummaryCSV generates CSV for monthly summary report.
func MonthlySummaryCSV(r *MonthlySummaryResult) *csvutil.Writer {
	cw := csvutil.New()

	// header
	catHeaders := make([]string, len(CategoryOrder))
	for i, c := range CategoryOrder {
		catHeaders[i] = CategoryLabels[c]
	}
	header := append([]string{"Date", "Day", "Gross Sales", "Interco", "Net", "Tax", "Returns"}, catHeaders...)
	header = append(header, "Ins. Payments", "Ins. Adjust.", "AR Patient", "AR Insurance", "AR Interco", "+/-")
	cw.Row(header...)

	f := csvutil.F

	// data rows
	for _, row := range r.Data {
		vals := []string{
			row["date"].(string), row["weekday"].(string),
			f(row["gross_sales"].(float64)), f(row["interco"].(float64)),
			f(row["net"].(float64)), f(row["tax"].(float64)), f(row["returns"].(float64)),
		}
		for _, cat := range CategoryOrder {
			vals = append(vals, f(row[cat].(float64)))
		}
		vals = append(vals,
			f(row["ins_payments"].(float64)), f(row["ins_adjustments"].(float64)),
			f(row["ar_patient"].(float64)), f(row["ar_insurance"].(float64)),
			f(row["ar_interco"].(float64)), f(row["balance_diff"].(float64)),
		)
		cw.Row(vals...)
	}

	// totals row
	cw.EmptyRow()
	totVals := []string{
		"", "TOTAL",
		f(r.Totals["gross_sales"]), f(r.Totals["interco"]),
		f(r.Totals["net"]), f(r.Totals["tax"]), f(r.Totals["returns"]),
	}
	for _, cat := range CategoryOrder {
		totVals = append(totVals, f(r.Totals[cat]))
	}
	totVals = append(totVals,
		f(r.Totals["ins_payments"]), f(r.Totals["ins_adjustments"]),
		f(r.Totals["ar_patient"]), f(r.Totals["ar_insurance"]),
		f(r.Totals["ar_interco"]), f(r.Totals["balance_diff"]),
	)
	cw.Row(totVals...)

	// percentages row
	f1 := csvutil.F1
	pctVals := []string{"", "%", "", "", "", "", ""}
	for _, cat := range CategoryOrder {
		pctVals = append(pctVals, f1(r.Percentages[cat]))
	}
	pctVals = append(pctVals,
		f1(r.Percentages["ins_payments"]), f1(r.Percentages["ins_adjustments"]),
		"", "", "", "",
	)
	cw.Row(pctVals...)

	cw.Flush()
	return cw
}

// DailyDetailCSV generates CSV for daily detail report.
func DailyDetailCSV(r *DailyDetailResult) *csvutil.Writer {
	cw := csvutil.New()
	f := csvutil.F
	f1 := csvutil.F1

	cw.Row(fmt.Sprintf("Daily Detail — %s", r.DateDisplay))
	cw.Row(
		fmt.Sprintf("Gross Sales: %s", f(r.GrossSales)),
		fmt.Sprintf("Net: %s", f(r.Net)),
		fmt.Sprintf("Tax: %s", f(r.Tax)),
		fmt.Sprintf("Ins. Payments: %s", f(r.InsPayments)),
		fmt.Sprintf("Ins. Adjust.: %s", f(r.InsAdjustments)),
		fmt.Sprintf("AR Patient: %s", f(r.ArPatient)),
		fmt.Sprintf("AR Insurance: %s", f(r.ArInsurance)),
		fmt.Sprintf("AR Interco: %s", f(r.ArInterco)),
	)
	cw.EmptyRow()
	cw.Row("Category", "Method", "Amount", "%")

	for _, grp := range r.Groups {
		items := grp["items"].([]map[string]interface{})
		catLabel := grp["category_label"].(string)
		for _, item := range items {
			cw.Row(
				catLabel,
				item["method_name"].(string),
				f(item["amount"].(float64)),
				f1(item["percentage"].(float64)),
			)
		}
		cw.Row(
			fmt.Sprintf("  TOTAL %s", catLabel), "",
			f(grp["subtotal"].(float64)),
			f1(grp["subtotal_pct"].(float64)),
		)
	}

	cw.EmptyRow()
	cw.Row("GRAND TOTAL", "", f(r.GrandTotal), "100.0")

	cw.Flush()
	return cw
}
