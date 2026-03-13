package report_daily_service

import (
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ─── helpers ─────────────────────────────────────────────────────────────────

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func pctChange(current, lastYear float64) float64 {
	if lastYear == 0 {
		return 0
	}
	return round2((current - lastYear) / lastYear * 100)
}

func showcaseIDs(permitted []int, db *gorm.DB) ([]int, error) {
	var ids []int
	if err := db.Raw(`SELECT id_location FROM location WHERE showcase = true AND id_location IN ?`, permitted).Scan(&ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// ─── 1. GET /sales ───────────────────────────────────────────────────────────

type CompRow struct {
	Location      string  `json:"location"`
	ShortName     *string `json:"short_name"`
	CurrentTotal  float64 `json:"current_total"`
	LastYearTotal float64 `json:"last_year_total"`
	Difference    float64 `json:"difference"`
	PctChange     float64 `json:"percentage_change"`
}

type DailySalesResult struct {
	CurrentDate string    `json:"current_date"`
	Weekday     string    `json:"weekday"`
	DayData     []CompRow `json:"day_data"`
	MonthData   []CompRow `json:"month_data"`
	YearData    []CompRow `json:"year_data"`
}

func (s *Service) DailySalesSummary(permittedIDs []int, targetDate time.Time) (*DailySalesResult, error) {
	locIDs, err := showcaseIDs(permittedIDs, s.db)
	if err != nil {
		return nil, err
	}
	if len(locIDs) == 0 {
		return nil, fmt.Errorf("no showcase locations found")
	}

	lastYearDate := targetDate.AddDate(-1, 0, 0)
	lastYearMonth := time.Date(lastYearDate.Year(), lastYearDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	currentMonth := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastYearStart := time.Date(lastYearDate.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	currentYearStart := time.Date(targetDate.Year(), 1, 1, 0, 0, 0, 0, time.UTC)

	type row struct {
		LocationName      string
		ShortName         *string
		TodayTotal        float64
		LastYearTodayTotal float64
		MtdTotal          float64
		LastYearMtd       float64
		YtdTotal          float64
		LastYearYtd       float64
	}

	q := `
		SELECT l.full_name AS location_name, l.short_name,
			SUM(CASE WHEN i.created_at::date = ? THEN i.final_amount ELSE 0 END) AS today_total,
			SUM(CASE WHEN i.created_at::date = ? THEN i.final_amount ELSE 0 END) AS last_year_today_total,
			SUM(CASE WHEN date_trunc('month', i.created_at) = ? THEN i.final_amount ELSE 0 END) AS mtd_total,
			SUM(CASE WHEN date_trunc('month', i.created_at) = ? THEN i.final_amount ELSE 0 END) AS last_year_mtd,
			SUM(CASE WHEN date_trunc('year', i.created_at) = ? THEN i.final_amount ELSE 0 END) AS ytd_total,
			SUM(CASE WHEN date_trunc('year', i.created_at) = ? THEN i.final_amount ELSE 0 END) AS last_year_ytd
		FROM invoice i
		JOIN location l ON l.id_location = i.location_id
		WHERE i.created_at::date <= ? AND i.location_id IN ?
		GROUP BY l.full_name, l.short_name
	`

	var rows []row
	if err := s.db.Raw(q,
		targetDate.Format("2006-01-02"),
		lastYearDate.Format("2006-01-02"),
		currentMonth, lastYearMonth,
		currentYearStart, lastYearStart,
		targetDate.Format("2006-01-02"), locIDs,
	).Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := &DailySalesResult{
		CurrentDate: targetDate.Format("2006-01-02"),
		Weekday:     targetDate.Format("Mon"),
	}

	for _, r := range rows {
		dayDiff := r.TodayTotal - r.LastYearTodayTotal
		result.DayData = append(result.DayData, CompRow{
			Location: r.LocationName, ShortName: r.ShortName,
			CurrentTotal: r.TodayTotal, LastYearTotal: r.LastYearTodayTotal,
			Difference: dayDiff, PctChange: pctChange(r.TodayTotal, r.LastYearTodayTotal),
		})

		mtdDiff := r.MtdTotal - r.LastYearMtd
		result.MonthData = append(result.MonthData, CompRow{
			Location: r.LocationName, ShortName: r.ShortName,
			CurrentTotal: r.MtdTotal, LastYearTotal: r.LastYearMtd,
			Difference: mtdDiff, PctChange: pctChange(r.MtdTotal, r.LastYearMtd),
		})

		ytdDiff := r.YtdTotal - r.LastYearYtd
		result.YearData = append(result.YearData, CompRow{
			Location: r.LocationName, ShortName: r.ShortName,
			CurrentTotal: r.YtdTotal, LastYearTotal: r.LastYearYtd,
			Difference: ytdDiff, PctChange: pctChange(r.YtdTotal, r.LastYearYtd),
		})
	}

	return result, nil
}

// ─── 2. GET /monthly_sales_summary ───────────────────────────────────────────

type MonthlySalesRow struct {
	Location     string  `json:"location"`
	GrossSales   float64 `json:"gross_sales"`
	NetSales     float64 `json:"net_sales"`
	TotalTax     float64 `json:"total_tax"`
	IntercoSales float64 `json:"interco_sales"`
}

func (s *Service) MonthlySalesSummary(permittedIDs []int, month, year int) ([]MonthlySalesRow, error) {
	locIDs, err := showcaseIDs(permittedIDs, s.db)
	if err != nil {
		return nil, err
	}
	if len(locIDs) == 0 {
		return nil, fmt.Errorf("no showcase locations found")
	}

	q := `
		SELECT l.full_name AS location,
			COALESCE(SUM(CASE WHEN EXTRACT(MONTH FROM i.created_at) = ? THEN i.total_amount ELSE 0 END), 0) AS gross_sales,
			COALESCE(SUM(i.final_amount - COALESCE(ri.return_amount, 0) - i.tax_amount), 0) AS net_sales,
			COALESCE(SUM(i.tax_amount), 0) AS total_tax,
			COALESCE(SUM(CASE WHEN i.number_invoice LIKE 'I%' THEN i.final_amount ELSE 0 END), 0) AS interco_sales
		FROM invoice i
		JOIN location l ON l.id_location = i.location_id
		LEFT JOIN return_invoice ri ON ri.invoice_id = i.id_invoice
		WHERE EXTRACT(MONTH FROM i.created_at) = ?
		  AND EXTRACT(YEAR FROM i.created_at) = ?
		  AND i.location_id IN ?
		GROUP BY l.full_name
	`

	var rows []MonthlySalesRow
	if err := s.db.Raw(q, month, month, year, locIDs).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// ─── 3. GET /ytd_sales_summary ───────────────────────────────────────────────

type YTDSalesRow struct {
	Date       string  `json:"date"`
	Location   string  `json:"location"`
	GrossSales float64 `json:"gross_sales"`
}

func (s *Service) YTDSalesSummary(locationID int, startDate, endDate time.Time, employeeID *int) ([]YTDSalesRow, error) {
	// Verify location is showcase
	var count int64
	s.db.Raw(`SELECT COUNT(*) FROM location WHERE id_location = ? AND showcase = true`, locationID).Scan(&count)
	if count == 0 {
		return nil, fmt.Errorf("showcase location not found")
	}

	q := `
		SELECT TO_CHAR(date_trunc('month', i.created_at), 'Mon YYYY') AS date,
			   l.full_name AS location,
			   COALESCE(SUM(i.total_amount), 0) AS gross_sales
		FROM invoice i
		JOIN location l ON l.id_location = i.location_id
		WHERE i.created_at BETWEEN ? AND ?
		  AND i.location_id = ?
		  AND (i.number_invoice LIKE 'S%' OR i.patient_id IS NOT NULL)
	`
	args := []interface{}{startDate, endDate, locationID}

	if employeeID != nil {
		q += ` AND i.employee_id = ?`
		args = append(args, *employeeID)
	}

	q += ` GROUP BY date_trunc('month', i.created_at), l.full_name ORDER BY date_trunc('month', i.created_at)`

	var rows []YTDSalesRow
	if err := s.db.Raw(q, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// ─── 4. GET /sales_cash ─────────────────────────────────────────────────────

type DailySalesCashResult struct {
	CurrentDate string    `json:"current_date"`
	Weekday     string    `json:"weekday"`
	DayData     []CompRow `json:"day_data"`
	MonthData   []CompRow `json:"month_data"`
	QuarterData []CompRow `json:"quarter_data"`
	YearData    []CompRow `json:"year_data"`
}

func (s *Service) DailySalesCash(permittedIDs []int, targetDate time.Time) (*DailySalesCashResult, error) {
	locIDs, err := showcaseIDs(permittedIDs, s.db)
	if err != nil {
		return nil, err
	}
	if len(locIDs) == 0 {
		return nil, fmt.Errorf("no showcase locations found")
	}

	lastYear := targetDate.Year() - 1
	lastYearDate := targetDate.AddDate(-1, 0, 0)

	// Quarter calculations
	currentQuarter := (int(targetDate.Month()) - 1) / 3
	quarterStart := time.Date(targetDate.Year(), time.Month(currentQuarter*3+1), 1, 0, 0, 0, 0, time.UTC)
	lastYearQuarterStart := time.Date(lastYear, time.Month(currentQuarter*3+1), 1, 0, 0, 0, 0, time.UTC)

	lastYearMonth := time.Date(lastYear, targetDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	currentMonth := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastYearStart := time.Date(lastYear, 1, 1, 0, 0, 0, 0, time.UTC)
	currentYearStart := time.Date(targetDate.Year(), 1, 1, 0, 0, 0, 0, time.UTC)

	// sum_expr = final_amount - discount - due
	type row struct {
		LocationName         string
		ShortName            *string
		TodayTotal           float64
		LastYearTodayTotal   float64
		MtdTotal             float64
		LastYearMtd          float64
		QtdTotal             float64
		LastYearQtd          float64
		YtdTotal             float64
		LastYearYtd          float64
	}

	q := `
		SELECT l.full_name AS location_name, l.short_name,
			SUM(CASE WHEN i.created_at::date = ? THEN COALESCE(i.final_amount,0) - COALESCE(i.discount,0) - COALESCE(i.due,0) ELSE 0 END) AS today_total,
			SUM(CASE WHEN i.created_at::date = ? THEN COALESCE(i.final_amount,0) - COALESCE(i.discount,0) - COALESCE(i.due,0) ELSE 0 END) AS last_year_today_total,
			SUM(CASE WHEN date_trunc('month', i.created_at) = ? THEN COALESCE(i.final_amount,0) - COALESCE(i.discount,0) - COALESCE(i.due,0) ELSE 0 END) AS mtd_total,
			SUM(CASE WHEN date_trunc('month', i.created_at) = ? THEN COALESCE(i.final_amount,0) - COALESCE(i.discount,0) - COALESCE(i.due,0) ELSE 0 END) AS last_year_mtd,
			SUM(CASE WHEN i.created_at::date >= ? AND i.created_at::date <= ? THEN COALESCE(i.final_amount,0) - COALESCE(i.discount,0) - COALESCE(i.due,0) ELSE 0 END) AS qtd_total,
			SUM(CASE WHEN i.created_at::date >= ? AND i.created_at::date <= ? THEN COALESCE(i.final_amount,0) - COALESCE(i.discount,0) - COALESCE(i.due,0) ELSE 0 END) AS last_year_qtd,
			SUM(CASE WHEN date_trunc('year', i.created_at) = ? THEN COALESCE(i.final_amount,0) - COALESCE(i.discount,0) - COALESCE(i.due,0) ELSE 0 END) AS ytd_total,
			SUM(CASE WHEN date_trunc('year', i.created_at) = ? THEN COALESCE(i.final_amount,0) - COALESCE(i.discount,0) - COALESCE(i.due,0) ELSE 0 END) AS last_year_ytd
		FROM invoice i
		JOIN location l ON l.id_location = i.location_id
		WHERE i.created_at <= ? AND i.location_id IN ?
		GROUP BY l.full_name, l.short_name
	`

	var rows []row
	if err := s.db.Raw(q,
		targetDate.Format("2006-01-02"),
		lastYearDate.Format("2006-01-02"),
		currentMonth, lastYearMonth,
		quarterStart.Format("2006-01-02"), targetDate.Format("2006-01-02"),
		lastYearQuarterStart.Format("2006-01-02"), lastYearDate.Format("2006-01-02"),
		currentYearStart, lastYearStart,
		targetDate.Format("2006-01-02"), locIDs,
	).Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := &DailySalesCashResult{
		CurrentDate: targetDate.Format("2006-01-02"),
		Weekday:     targetDate.Format("Mon"),
	}

	for _, r := range rows {
		result.DayData = append(result.DayData, CompRow{
			Location: r.LocationName, ShortName: r.ShortName,
			CurrentTotal: r.TodayTotal, LastYearTotal: r.LastYearTodayTotal,
			Difference: r.TodayTotal - r.LastYearTodayTotal,
			PctChange: pctChange(r.TodayTotal, r.LastYearTodayTotal),
		})
		result.MonthData = append(result.MonthData, CompRow{
			Location: r.LocationName, ShortName: r.ShortName,
			CurrentTotal: r.MtdTotal, LastYearTotal: r.LastYearMtd,
			Difference: r.MtdTotal - r.LastYearMtd,
			PctChange: pctChange(r.MtdTotal, r.LastYearMtd),
		})
		result.QuarterData = append(result.QuarterData, CompRow{
			Location: r.LocationName, ShortName: r.ShortName,
			CurrentTotal: r.QtdTotal, LastYearTotal: r.LastYearQtd,
			Difference: r.QtdTotal - r.LastYearQtd,
			PctChange: pctChange(r.QtdTotal, r.LastYearQtd),
		})
		result.YearData = append(result.YearData, CompRow{
			Location: r.LocationName, ShortName: r.ShortName,
			CurrentTotal: r.YtdTotal, LastYearTotal: r.LastYearYtd,
			Difference: r.YtdTotal - r.LastYearYtd,
			PctChange: pctChange(r.YtdTotal, r.LastYearYtd),
		})
	}

	return result, nil
}

// ─── 5. GET /journal_report ──────────────────────────────────────────────────

type JournalReportResult struct {
	Data   []map[string]interface{} `json:"data"`
	Totals map[string]float64       `json:"totals"`
}

func (s *Service) JournalReport(locationID int, startDate, endDate time.Time, summary bool) (*JournalReportResult, error) {
	// Load payment methods
	type pm struct {
		ID        int
		ShortName *string
		Name      string
	}
	var methods []pm
	s.db.Raw(`SELECT id_payment_method AS id, short_name, method_name AS name FROM payment_method`).Scan(&methods)

	// Build dynamic payment columns
	pmCols := ""
	for _, m := range methods {
		key := m.Name
		if m.ShortName != nil && *m.ShortName != "" {
			key = *m.ShortName
		}
		pmCols += fmt.Sprintf(`,COALESCE(SUM(CASE WHEN ph.payment_method_id = %d THEN ph.paid ELSE 0 END), 0) AS "%s"`, m.ID, key)
	}

	groupBy := "i.created_at, i.number_invoice"
	orderBy := "i.created_at, i.number_invoice"
	dateExpr := "i.created_at"
	if summary {
		groupBy = "i.created_at::date"
		orderBy = "i.created_at::date"
		dateExpr = "i.created_at::date"
	}

	q := fmt.Sprintf(`
		SELECT %s AS date, i.number_invoice AS invoice
			%s
			,COALESCE(SUM(ins.ins_paid), 0) AS ins_pmt
			,SUM(COALESCE(i.discount, 0)) AS ins_adj
			,SUM(COALESCE(i.gift_card_bal, 0)) AS gift_pmt
		FROM invoice i
		LEFT JOIN (
			SELECT invoice_id, payment_method_id, SUM(amount) AS paid
			FROM payment_history
			WHERE payment_method_id != 14
			GROUP BY invoice_id, payment_method_id
		) ph ON ph.invoice_id = i.id_invoice
		LEFT JOIN (
			SELECT invoice_id, COALESCE(SUM(amount::numeric), 0) AS ins_paid
			FROM insurance_payment
			GROUP BY invoice_id
		) ins ON ins.invoice_id = i.id_invoice
		WHERE i.location_id = ?
		  AND i.created_at BETWEEN ? AND ?
		  AND (i.number_invoice LIKE 'S%%' OR i.patient_id IS NOT NULL)
		GROUP BY %s
		ORDER BY %s
	`, dateExpr, pmCols, groupBy, orderBy)

	var rawRows []map[string]interface{}
	if err := s.db.Raw(q, locationID, startDate, endDate).Scan(&rawRows).Error; err != nil {
		return nil, err
	}

	// Format dates and ensure floats
	data := make([]map[string]interface{}, 0, len(rawRows))
	for _, raw := range rawRows {
		row := make(map[string]interface{})
		for k, v := range raw {
			if k == "date" {
				if t, ok := v.(time.Time); ok {
					row[k] = t.Format("01/02")
				} else {
					row[k] = v
				}
			} else if k == "invoice" {
				row[k] = v
			} else {
				row[k] = toFloat(v)
			}
		}
		data = append(data, row)
	}

	// Calculate totals
	totals := make(map[string]float64)
	for _, m := range methods {
		key := m.Name
		if m.ShortName != nil && *m.ShortName != "" {
			key = *m.ShortName
		}
		totals[key] = 0
	}
	totals["total_ins_pmt"] = 0
	totals["total_ins_adj"] = 0
	totals["total_gift_pmt"] = 0

	for _, row := range data {
		for _, m := range methods {
			key := m.Name
			if m.ShortName != nil && *m.ShortName != "" {
				key = *m.ShortName
			}
			totals[key] += toFloat(row[key])
		}
		totals["total_ins_pmt"] += toFloat(row["ins_pmt"])
		totals["total_ins_adj"] += toFloat(row["ins_adj"])
		totals["total_gift_pmt"] += toFloat(row["gift_pmt"])
	}

	return &JournalReportResult{Data: data, Totals: totals}, nil
}

// ─── 6. GET /journal_transfer ────────────────────────────────────────────────

type JournalTransferResult struct {
	Data  []map[string]interface{} `json:"data"`
	Total float64                  `json:"total"`
}

func (s *Service) JournalTransfer(locationID int, startDate, endDate time.Time) (*JournalTransferResult, error) {
	q := `
		SELECT i.created_at::date AS date,
			   lf.full_name AS from_location,
			   i.location_id AS acct_from,
			   lt.full_name AS to_location,
			   i.to_location_id AS acct_to,
			   i.number_invoice AS invoice,
			   i.total_amount AS amount
		FROM invoice i
		JOIN location lf ON lf.id_location = i.location_id
		LEFT JOIN location lt ON lt.id_location = i.to_location_id
		WHERE i.location_id = ?
		  AND i.created_at BETWEEN ? AND ?
		  AND (i.number_invoice LIKE 'I%' OR i.to_location_id IS NOT NULL)
		ORDER BY i.created_at
	`

	type row struct {
		Date         time.Time
		FromLocation string
		AcctFrom     int
		ToLocation   *string
		AcctTo       *int
		Invoice      string
		Amount       float64
	}

	var rows []row
	if err := s.db.Raw(q, locationID, startDate, endDate).Scan(&rows).Error; err != nil {
		return nil, err
	}

	data := make([]map[string]interface{}, 0, len(rows))
	var total float64
	for _, r := range rows {
		toLocation := "Unknown"
		if r.ToLocation != nil {
			toLocation = *r.ToLocation
		}
		var acctTo interface{} = "Unknown"
		if r.AcctTo != nil {
			acctTo = *r.AcctTo
		}
		data = append(data, map[string]interface{}{
			"date":          r.Date.Format("01/02"),
			"from_location": r.FromLocation,
			"acct_from":     r.AcctFrom,
			"to_location":   toLocation,
			"acct_to":       acctTo,
			"invoice":       r.Invoice,
			"amount":        r.Amount,
		})
		total += r.Amount
	}

	return &JournalTransferResult{Data: data, Total: total}, nil
}

// ─── 7. GET /journal_receipts ────────────────────────────────────────────────

type JournalReceiptsResult struct {
	Data  []map[string]interface{} `json:"data"`
	Total float64                  `json:"total"`
}

func (s *Service) JournalReceipts(locationID int, startDate, endDate time.Time) (*JournalReceiptsResult, error) {
	q := `
		SELECT i.created_at::date AS date,
			   i.number_invoice AS receipt_nbr,
			   v.id_vendor AS vendor_id,
			   v.vendor_name,
			   vi.invoice_no AS vendor_invoice,
			   COALESCE(vi.invoice_date, i.created_at) AS vendor_date,
			   vi.invoice_total AS vendor_total,
			   vi.sub_total AS vendor_subtotal,
			   vi.shipping_handling AS vendor_shipping,
			   vi.tax AS vendor_tax,
			   i.final_amount AS actual_total,
			   (i.final_amount - COALESCE(vi.tax, 0)) AS act_f_cost,
			   (i.final_amount - COALESCE(vi.shipping_handling, 0)) AS act_part_cost
		FROM invoice i
		JOIN vendor v ON v.id_vendor = i.vendor_id
		JOIN vendor_invoice vi ON vi.invoice_id = i.id_invoice
		WHERE i.location_id = ?
		  AND i.created_at BETWEEN ? AND ?
		  AND i.number_invoice LIKE 'V%'
		ORDER BY i.created_at
	`

	type row struct {
		Date            time.Time
		ReceiptNbr      string
		VendorID        int
		VendorName      string
		VendorInvoice   *string
		VendorDate      time.Time
		VendorTotal     float64
		VendorSubtotal  float64
		VendorShipping  float64
		VendorTax       float64
		ActualTotal     float64
		ActFCost        float64
		ActPartCost     float64
	}

	var rows []row
	if err := s.db.Raw(q, locationID, startDate, endDate).Scan(&rows).Error; err != nil {
		return nil, err
	}

	data := make([]map[string]interface{}, 0, len(rows))
	var total float64
	for _, r := range rows {
		vinv := "Unknown"
		if r.VendorInvoice != nil {
			vinv = *r.VendorInvoice
		}
		data = append(data, map[string]interface{}{
			"receipt_nbr":    r.ReceiptNbr,
			"vendor_id":      r.VendorID,
			"vendor_name":    r.VendorName,
			"vendor_invoice":  vinv,
			"vendor_date":    r.VendorDate.Format("01/02/2006"),
			"vendor_total":   r.VendorTotal,
			"vendor_subtotal": r.VendorSubtotal,
			"vendor_shipping": r.VendorShipping,
			"vendor_tax":     r.VendorTax,
			"actual_total":   r.ActualTotal,
			"act_f_cost":     r.ActFCost,
			"act_part_cost":  r.ActPartCost,
		})
		total += r.VendorTotal
	}

	return &JournalReceiptsResult{Data: data, Total: total}, nil
}

// ─── 8. GET /all_reports ─────────────────────────────────────────────────────

func AllReports() map[string]interface{} {
	return map[string]interface{}{
		"Sales": []map[string]string{
			{"label": "Invoice Summary", "path": "/invoice-summary"},
			{"label": "Invoice Classification", "path": "/invoice-classification"},
			{"label": "Vendor/Brand Margin Report", "path": "/vendor-brand-margin-report"},
			{"label": "Sales by location", "path": "/sales-by-location"},
			{"label": "Sales by Frame", "path": "/sales-by-frame"},
			{"label": "Sales Average", "path": "/sales-average"},
			{"label": "Sales Breakdown by Product", "path": "/sales-breakdown-by-product"},
			{"label": "Sales by Employee - Detailed (Fully)", "path": "/sales-by-emp-detailed"},
			{"label": "Gift Card Balance", "path": "/gift-card-balance"},
			{"label": "Gift Card Activities", "path": "/gift-card-activities"},
			{"label": "SMS Purchases", "path": "/sms-purchases"},
		},
		"Inventory": []map[string]string{
			{"label": "List of Orders Placed", "path": "/list-of-orders-placed"},
			{"label": "List of Receipts", "path": "/list-of-receipts"},
			{"label": "Receipt by Brand", "path": "/receipt-by-brand"},
			{"label": "Missing Inventory", "path": "/missing-inventory"},
			{"label": "Inventory Work Flow", "path": "/inventory-work-flow"},
			{"label": "Inventory Analysis", "path": "/inventory-analysis"},
			{"label": "WOS Lens Order", "path": "/wos-lens-order"},
		},
		"Audit Logs": []map[string]string{
			{"label": "Audit Logins", "path": "/audit-logins"},
			{"label": "Audit Invoices", "path": "/audit-invoices"},
			{"label": "Audit Invoice Lines", "path": "/audit-invoice-lines"},
			{"label": "Audit Doctor Exams", "path": "/audit-dr-exams"},
			{"label": "Audit Reports", "path": "/audit-reports"},
			{"label": "Audit Files", "path": "/audit-files"},
			{"label": "Email Log", "path": "/email-log"},
			{"label": "SMS Log", "path": "/sms-log"},
			{"label": "Fax Log", "path": "/fax-log"},
			{"label": "Phone Log", "path": "/phone-log"},
			{"label": "Patient Communication Log", "path": "/patient-communication-log"},
		},
		"Performance": []map[string]string{
			{"label": "Lab Margin", "path": "/lab-margin"},
			{"label": "Discounts", "path": "/discounts"},
			{"label": "Inventory Turns (by Brand)", "path": "/inventory-turns"},
			{"label": "Score Card", "path": "/score-card"},
			{"label": "Elite Report", "path": "/elite-report"},
			{"label": "Appointment Search", "path": "/appointment-search"},
			{"label": "Appointment by Type", "path": "/appointment-by-type"},
			{"label": "Appointment by Insurance", "path": "/appointment-by-insurance"},
			{"label": "Invoice Time Analysis", "path": "/invoice-time-analysis"},
		},
		"Doctor Reports": []map[string]string{
			{"label": "Revenue by Doctor", "path": "/revenue-by-doctor"},
			{"label": "Professional Fees by Doctor/Payments Received", "path": "/prof-fees-by-dr"},
			{"label": "Appointment Stats", "path": "/appointment-stats"},
			{"label": "Sales (Doctor location) - Exams", "path": "/sales-doctor-location-exams"},
			{"label": "Appointment Sales", "path": "/appointment-sales"},
			{"label": "Referral Source", "path": "/referral-source"},
		},
		"Insurance": []map[string]string{
			{"label": "Insurance Statistics", "path": "/insurance-statistics"},
		},
		"Marketing": []map[string]string{
			{"label": "Live Survey Results", "path": "/live-survey-results"},
			{"label": "Mailing List", "path": "/mailing-list"},
			{"label": "List of Birthdays", "path": "/list-of-birthdays"},
			{"label": "List of All Patient/Customers", "path": "/list-of-patients-customers"},
			{"label": "Nonprofit Report", "path": "/nonprofit-report"},
		},
		"Accounting": []map[string]string{
			{"label": "Monthly Summary", "path": "/monthly-summary"},
			{"label": "Accounts Receivable Aging Report", "path": "/accounts-receivable-aging-report"},
			{"label": "Accounts Receivable as of a Specific Date", "path": "/accounts-receivable-specific-date"},
			{"label": "Accounts Receivable Analysis", "path": "/accounts-receivable-analysis"},
			{"label": "List of Transfer Credit", "path": "/list-of-transfer-credit"},
			{"label": "Credit Card Payments", "path": "/credit-card-payments"},
			{"label": "Payment Summary", "path": "/payment-summary"},
			{"label": "Payment Details", "path": "/payment-details"},
			{"label": "Payment Categories", "path": "/payment-categories"},
			{"label": "Payment Overview", "path": "/payment-overview"},
			{"label": "Payment Report", "path": "/payment-report"},
			{"label": "Accounts Payable Due By Month", "path": "/accounts-payable-due-by-month"},
		},
	}
}

// ─── 9. GET /locations ───────────────────────────────────────────────────────

type LocationItem struct {
	LocationID   int    `json:"location_id"`
	LocationName string `json:"location_name"`
}

func (s *Service) ShowcaseLocations(permittedIDs []int) ([]LocationItem, error) {
	var items []LocationItem
	if err := s.db.Raw(`
		SELECT id_location AS location_id, full_name AS location_name
		FROM location
		WHERE showcase = true AND id_location IN ?
	`, permittedIDs).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// ─── 10. GET /employees ──────────────────────────────────────────────────────

type EmployeeItem struct {
	EmployeeID int    `json:"employee_id"`
	Name       string `json:"name"`
}

func (s *Service) AllEmployees() ([]EmployeeItem, error) {
	var items []EmployeeItem
	if err := s.db.Raw(`
		SELECT id_employee AS employee_id,
			   CONCAT(first_name, ' ', last_name) AS name
		FROM employee
	`).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// ─── util ────────────────────────────────────────────────────────────────────

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int64:
		return float64(n)
	case int:
		return float64(n)
	case nil:
		return 0
	default:
		return 0
	}
}
