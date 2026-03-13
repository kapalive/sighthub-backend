package report_library_service

import (
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ── helpers ─────────────────────────────────────────────────────────────────

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

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func ptrStr(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// ── types ───────────────────────────────────────────────────────────────────

type InvoiceSummaryItem struct {
	Date               string  `json:"date"`
	Invoice            string  `json:"invoice"`
	Total              float64 `json:"total"`
	Final              float64 `json:"final"`
	Due                float64 `json:"due"`
	PtBalance          float64 `json:"pt_balance"`
	InsuranceBalance   float64 `json:"insurance_balance"`
	GiftCardBalance    float64 `json:"gift_card_balance"`
	Discount           float64 `json:"discount"`
	Tax                float64 `json:"tax"`
	NetSales           float64 `json:"net_sales"`
}

type InvoiceSummaryTotals struct {
	Count                  int     `json:"count"`
	TotalAmount            float64 `json:"total_amount"`
	TotalFinal             float64 `json:"total_final"`
	TotalDue               float64 `json:"total_due"`
	TotalPtBalance         float64 `json:"total_pt_balance"`
	TotalInsuranceBalance  float64 `json:"total_insurance_balance"`
	TotalGiftCardBalance   float64 `json:"total_gift_card_balance"`
	TotalDiscount          float64 `json:"total_discount"`
	TotalTax               float64 `json:"total_tax"`
}

type InvoiceClassificationItem struct {
	Source     string  `json:"source"`
	Count      float64 `json:"count"`
	TotalSales float64 `json:"total_sales"`
	AvgSale    float64 `json:"avg_sale"`
	Employee   string  `json:"employee,omitempty"`
}

type VendorBrandMarginItem struct {
	Vendor      string  `json:"vendor"`
	Brand       string  `json:"brand"`
	Quantity    int     `json:"quantity"`
	Cost        float64 `json:"cost"`
	Sales       float64 `json:"sales"`
	GrossMargin float64 `json:"gross_margin"`
}

type SalesByLocationItem struct {
	Location       string  `json:"location"`
	InsuranceSales float64 `json:"insurance_sales"`
	PatientSales   float64 `json:"patient_sales"`
}

type SalesByFrameItem struct {
	DateOfSale    string      `json:"date_of_sale"`
	InvoiceNumber string      `json:"invoice_number"`
	Vendor        string      `json:"vendor"`
	Brand         string      `json:"brand"`
	Model         string      `json:"model"`
	SerialNumber  string      `json:"serial_number"`
	UPC           string      `json:"upc"`
	Quantity      int         `json:"quantity"`
	Cost          float64     `json:"cost"`
	Price         float64     `json:"price"`
	Customer      interface{} `json:"customer"`
}

type SalesAverageItem struct {
	Employee      string  `json:"employee"`
	Invoices      int     `json:"invoices"`
	FrameAvgSales float64 `json:"frame_avg_sales"`
	LensAvgSales  float64 `json:"lens_avg_sales"`
	ARCount       int     `json:"a_r_count"`
	Transitions   int     `json:"transitions"`
	Polarized     int     `json:"polarized"`
}

type SalesBreakdownItem struct {
	Date          string      `json:"date"`
	InvoiceNumber string      `json:"invoice_number"`
	PatientName   interface{} `json:"patient_name"`
	InvoiceTotal  float64     `json:"invoice_total"`
	Cost          float64     `json:"cost"`
	OphFrOnly     float64     `json:"oph_fr_only"`
	SunPlano      float64     `json:"sun_plano"`
	LOnly         float64     `json:"l_only"`
	Bifocal       float64     `json:"bifocal"`
	Progressive   float64     `json:"progressive"`
	Contacts      float64     `json:"contacts"`
	OphFrRx       float64     `json:"oph_fr_rx"`
	SunRx         float64     `json:"sun_rx"`
	Perfume       float64     `json:"perfume"`
	Electronics   float64     `json:"electronics"`
	Other         float64     `json:"other"`
}

type SalesByEmployeeItem struct {
	Employee           string  `json:"employee"`
	GrossSales         float64 `json:"gross_sales"`
	PtPay              float64 `json:"pt_pay"`
	InsPay             float64 `json:"ins_pay"`
	NetSales           float64 `json:"net_sales"`
	Discount           float64 `json:"discount"`
	FramesSales        float64 `json:"frames_sales"`
	LensSales          float64 `json:"lens_sales"`
	PlanoSales         float64 `json:"plano_sales"`
	SingleVisionSales  float64 `json:"single_vision_sales"`
}

type GiftCardBalanceItem struct {
	IssuedDate string  `json:"issued_date"`
	CardCode   string  `json:"card_code"`
	Balance    float64 `json:"balance"`
}

type GiftCardDetailsTx struct {
	Date          string  `json:"date"`
	InvoiceNumber *string `json:"invoice_number"`
	PatientName   *string `json:"patient_name"`
	Action        string  `json:"action"`
	Amount        float64 `json:"amount"`
}

type GiftCardDetails struct {
	IDGiftCard            int              `json:"id_gift_card"`
	Code                  string           `json:"code"`
	Nominal               float64          `json:"nominal"`
	Balance               float64          `json:"balance"`
	PurchaseInvoiceNumber *string          `json:"purchase_invoice_number"`
	ExpirationDate        *string          `json:"expiration_date"`
	Transactions          []GiftCardDetailsTx `json:"transactions"`
}

type GiftCardActivityItem struct {
	CardCode    string  `json:"card_code"`
	PatientName *string `json:"patient_name"`
	Date        string  `json:"date"`
	InvoiceNumber *string `json:"invoice_number"`
	Action      string  `json:"action"`
	Adding      float64 `json:"adding"`
	Using       float64 `json:"using"`
}

type ReferralCountItem struct {
	Title string `json:"title"`
	Count int    `json:"count"`
}

// ── methods ─────────────────────────────────────────────────────────────────

func (s *Service) InvoiceSummary(locationIDs []int, startDate, endDate string) ([]InvoiceSummaryItem, InvoiceSummaryTotals, error) {
	query := `
		SELECT DATE(i.created_at) AS date,
		       i.number_invoice   AS invoice,
		       COALESCE(i.total_amount,0)  AS total,
		       COALESCE(i.final_amount,0)  AS final,
		       COALESCE(i.due,0)           AS due,
		       COALESCE(i.pt_bal,0)        AS pt_balance,
		       COALESCE(i.ins_bal,0)       AS insurance_balance,
		       COALESCE(i.gift_card_bal,0) AS gift_card_balance,
		       COALESCE(i.discount,0)      AS discount,
		       COALESCE(i.tax_amount,0)    AS tax
		FROM invoice i
		WHERE i.created_at BETWEEN ? AND ?
		  AND i.location_id IN (?)
		  AND (i.number_invoice LIKE 'S%' OR i.patient_id IS NOT NULL)
		ORDER BY i.created_at`

	rows, err := s.db.Raw(query, startDate, endDate, locationIDs).Rows()
	if err != nil {
		return nil, InvoiceSummaryTotals{}, err
	}
	defer rows.Close()

	var items []InvoiceSummaryItem
	var totals InvoiceSummaryTotals

	for rows.Next() {
		var date time.Time
		var invoice string
		var total, final_, due, ptBal, insBal, gcBal, discount, tax float64

		if err := rows.Scan(&date, &invoice, &total, &final_, &due, &ptBal, &insBal, &gcBal, &discount, &tax); err != nil {
			return nil, InvoiceSummaryTotals{}, err
		}

		netSales := final_ - tax

		items = append(items, InvoiceSummaryItem{
			Date:             date.Format("01/02/2006"),
			Invoice:          invoice,
			Total:            round2(total),
			Final:            round2(final_),
			Due:              round2(due),
			PtBalance:        round2(ptBal),
			InsuranceBalance: round2(insBal),
			GiftCardBalance:  round2(gcBal),
			Discount:         round2(discount),
			Tax:              round2(tax),
			NetSales:         round2(netSales),
		})

		totals.Count++
		totals.TotalAmount += total
		totals.TotalFinal += final_
		totals.TotalDue += due
		totals.TotalPtBalance += ptBal
		totals.TotalInsuranceBalance += insBal
		totals.TotalGiftCardBalance += gcBal
		totals.TotalDiscount += discount
		totals.TotalTax += tax
	}

	return items, totals, nil
}

func (s *Service) InvoiceClassification(locationID int, startDate, endDate, classificationType string) ([]InvoiceClassificationItem, error) {
	var query string
	if classificationType == "by_employee" {
		query = `
			SELECT COALESCE(i.source, '[absent]') AS source,
			       COUNT(i.id_invoice)             AS cnt,
			       COALESCE(SUM(i.total_amount),0) AS total_sales,
			       COALESCE(AVG(i.total_amount),0) AS avg_sale,
			       TRIM(CONCAT(e.first_name,' ',e.last_name)) AS employee
			FROM invoice i
			JOIN employee e ON e.id_employee = i.employee_id
			WHERE i.created_at BETWEEN ? AND ?
			  AND i.location_id = ?
			GROUP BY e.first_name, e.last_name, i.source
			ORDER BY source`
	} else {
		query = `
			SELECT COALESCE(i.source, '[absent]') AS source,
			       COUNT(i.id_invoice)             AS cnt,
			       COALESCE(SUM(i.total_amount),0) AS total_sales,
			       COALESCE(AVG(i.total_amount),0) AS avg_sale
			FROM invoice i
			WHERE i.created_at BETWEEN ? AND ?
			  AND i.location_id = ?
			GROUP BY i.source
			ORDER BY source`
	}

	rows, err := s.db.Raw(query, startDate, endDate, locationID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []InvoiceClassificationItem
	for rows.Next() {
		var item InvoiceClassificationItem
		if classificationType == "by_employee" {
			if err := rows.Scan(&item.Source, &item.Count, &item.TotalSales, &item.AvgSale, &item.Employee); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(&item.Source, &item.Count, &item.TotalSales, &item.AvgSale); err != nil {
				return nil, err
			}
		}
		items = append(items, item)
	}

	return items, nil
}

func (s *Service) VendorBrandMarginReport(locationIDs []int, startDate, endDate string) ([]VendorBrandMarginItem, error) {
	query := `
		SELECT v.vendor_name                            AS vendor,
		       b.brand_name                             AS brand,
		       COALESCE(SUM(iis.quantity),0)            AS quantity,
		       COALESCE(SUM(pb.item_net * iis.quantity),0) AS cost,
		       COALESCE(SUM(iis.total),0)               AS sales,
		       CASE WHEN SUM(iis.total) > 0
		            THEN ROUND((SUM(iis.total) - SUM(pb.item_net * iis.quantity)) / SUM(iis.total) * 100, 2)
		            ELSE 0 END                          AS gross_margin
		FROM invoice_item_sale iis
		JOIN invoice i         ON i.id_invoice    = iis.invoice_id
		JOIN inventory inv     ON inv.invoice_id  = i.id_invoice
		JOIN price_book pb     ON pb.inventory_id = inv.id_inventory
		JOIN vendor v          ON v.id_vendor     = i.vendor_id
		JOIN product p         ON p.id_product    = inv.product_id
		JOIN brand b           ON b.id_brand      = p.brand_id
		WHERE i.created_at BETWEEN ? AND ?
		  AND inv.status_items_inventory = 'SOLD'
		  AND (i.number_invoice LIKE 'S%' OR i.patient_id IS NOT NULL)
		  AND i.location_id IN (?)
		GROUP BY v.vendor_name, b.brand_name
		ORDER BY v.vendor_name, b.brand_name`

	rows, err := s.db.Raw(query, startDate, endDate, locationIDs).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []VendorBrandMarginItem
	for rows.Next() {
		var it VendorBrandMarginItem
		if err := rows.Scan(&it.Vendor, &it.Brand, &it.Quantity, &it.Cost, &it.Sales, &it.GrossMargin); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

func (s *Service) SalesByLocation(locationIDs []int, startDate, endDate string) ([]SalesByLocationItem, error) {
	query := `
		SELECT l.full_name                    AS location,
		       COALESCE(SUM(i.ins_bal),0)     AS insurance_sales,
		       COALESCE(SUM(i.pt_bal),0)      AS patient_sales
		FROM invoice i
		JOIN location l ON l.id_location = i.location_id
		WHERE i.created_at BETWEEN ? AND ?
		  AND (i.number_invoice LIKE 'S%' OR i.patient_id IS NOT NULL)
		  AND i.location_id IN (?)
		GROUP BY l.full_name
		ORDER BY l.full_name`

	rows, err := s.db.Raw(query, startDate, endDate, locationIDs).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []SalesByLocationItem
	for rows.Next() {
		var it SalesByLocationItem
		if err := rows.Scan(&it.Location, &it.InsuranceSales, &it.PatientSales); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

func (s *Service) SalesByFrame(locationIDs []int, startDate, endDate string, locationID *int) ([]SalesByFrameItem, error) {
	query := `
		SELECT i.created_at                        AS date_of_sale,
		       i.number_invoice                    AS invoice_number,
		       v.vendor_name                       AS vendor,
		       b.brand_name                        AS brand,
		       m.title_variant                     AS model,
		       inv.sku                             AS serial_number,
		       COALESCE(m.upc,'')                  AS upc,
		       COALESCE(SUM(iis.quantity),0)        AS quantity,
		       COALESCE(SUM(iis.price),0)           AS cost,
		       COALESCE(SUM(iis.total),0)           AS price,
		       i.patient_id                        AS customer
		FROM invoice i
		JOIN inventory inv     ON inv.invoice_id    = i.id_invoice
		JOIN invoice_item_sale iis ON iis.invoice_id = i.id_invoice
		JOIN model m           ON m.id_model        = inv.model_id
		JOIN product p         ON p.id_product      = m.product_id
		JOIN vendor v          ON v.id_vendor       = i.vendor_id
		JOIN brand b           ON b.id_brand        = p.brand_id
		WHERE i.created_at BETWEEN ? AND ?
		  AND inv.status_items_inventory = 'SOLD'
		  AND (i.number_invoice LIKE 'S%' OR i.patient_id IS NOT NULL)
		  AND i.location_id IN (?)`

	args := []interface{}{startDate, endDate, locationIDs}
	if locationID != nil {
		query += ` AND i.location_id = ?`
		args = append(args, *locationID)
	}

	query += `
		GROUP BY i.created_at, i.number_invoice, v.vendor_name, b.brand_name,
		         m.title_variant, inv.sku, m.upc, i.patient_id
		ORDER BY i.created_at`

	rows, err := s.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []SalesByFrameItem
	for rows.Next() {
		var dateOfSale time.Time
		var invoiceNum, vendor, brand, model, serial, upc string
		var qty int
		var cost, price float64
		var customer *int64

		if err := rows.Scan(&dateOfSale, &invoiceNum, &vendor, &brand, &model, &serial, &upc, &qty, &cost, &price, &customer); err != nil {
			return nil, err
		}

		items = append(items, SalesByFrameItem{
			DateOfSale:    dateOfSale.Format("01/02/2006"),
			InvoiceNumber: invoiceNum,
			Vendor:        vendor,
			Brand:         brand,
			Model:         model,
			SerialNumber:  serial,
			UPC:           upc,
			Quantity:      qty,
			Cost:          round2(cost),
			Price:         round2(price),
			Customer:      customer,
		})
	}
	return items, nil
}

func (s *Service) SalesAverage(locationIDs []int, startDate, endDate string) ([]SalesAverageItem, error) {
	query := `
		SELECT TRIM(CONCAT(e.first_name,' ',e.last_name)) AS employee,
		       COUNT(DISTINCT i.id_invoice)                AS invoices,
		       COALESCE(AVG(CASE WHEN iis.item_type='Frames' THEN iis.total ELSE 0 END),0) AS frame_avg,
		       COALESCE(AVG(CASE WHEN iis.item_type='Lens'   THEN iis.total ELSE 0 END),0) AS lens_avg,
		       COUNT(CASE WHEN i.due > 0 AND i.due < i.final_amount THEN 1 END)            AS ar_count,
		       COALESCE(SUM(CASE WHEN iis.item_type='Lens'   THEN iis.quantity ELSE 0 END),0) AS transitions,
		       COALESCE(SUM(CASE WHEN iis.item_type='Frames' THEN iis.quantity ELSE 0 END),0) AS polarized
		FROM invoice i
		JOIN employee e            ON e.id_employee  = i.employee_id
		JOIN invoice_item_sale iis ON iis.invoice_id = i.id_invoice
		WHERE i.created_at BETWEEN ? AND ?
		  AND i.location_id IN (?)
		GROUP BY e.first_name, e.last_name
		ORDER BY employee`

	rows, err := s.db.Raw(query, startDate, endDate, locationIDs).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []SalesAverageItem
	for rows.Next() {
		var it SalesAverageItem
		if err := rows.Scan(&it.Employee, &it.Invoices, &it.FrameAvgSales, &it.LensAvgSales, &it.ARCount, &it.Transitions, &it.Polarized); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

func (s *Service) SalesBreakdownByProductType(locationIDs []int, startDate, endDate string) ([]SalesBreakdownItem, error) {
	query := `
		SELECT i.date_create                                                   AS date,
		       i.number_invoice                                                AS invoice_number,
		       i.patient_id                                                    AS patient_name,
		       COALESCE(SUM(i.final_amount),0)                                 AS invoice_total,
		       COALESCE(SUM(iis.total),0)                                      AS cost,
		       COALESCE(SUM(CASE WHEN m.type_products='eyeglasses' THEN iis.total ELSE 0 END),0)      AS oph_fr_only,
		       COALESCE(SUM(CASE WHEN m.type_products='sunglasses' THEN iis.total ELSE 0 END),0)      AS sun_plano,
		       COALESCE(SUM(CASE WHEN m.lens_type='single_vision' THEN iis.total ELSE 0 END),0)       AS l_only,
		       COALESCE(SUM(CASE WHEN m.lens_type='bifocal'       THEN iis.total ELSE 0 END),0)       AS bifocal,
		       COALESCE(SUM(CASE WHEN m.lens_type='progressive'   THEN iis.total ELSE 0 END),0)       AS progressive,
		       COALESCE(SUM(CASE WHEN m.type_products='contacts'  THEN iis.total ELSE 0 END),0)       AS contacts,
		       COALESCE(SUM(CASE WHEN m.category_glasses_id=1     THEN iis.total ELSE 0 END),0)       AS oph_fr_rx,
		       COALESCE(SUM(CASE WHEN m.category_glasses_id=2     THEN iis.total ELSE 0 END),0)       AS sun_rx,
		       COALESCE(SUM(CASE WHEN m.category_glasses_id=4     THEN iis.total ELSE 0 END),0)       AS perfume,
		       COALESCE(SUM(CASE WHEN m.category_glasses_id=7     THEN iis.total ELSE 0 END),0)       AS electronics,
		       COALESCE(SUM(CASE WHEN m.category_glasses_id NOT IN (1,2,4,7) THEN iis.total ELSE 0 END),0) AS other
		FROM invoice i
		JOIN invoice_item_sale iis ON i.id_invoice   = iis.invoice_id
		JOIN inventory inv         ON inv.id_inventory = iis.item_id
		JOIN model m               ON m.id_model      = inv.model_id
		WHERE i.date_create BETWEEN ? AND ?
		  AND (i.number_invoice LIKE 'S%' OR i.patient_id IS NOT NULL)
		  AND i.location_id IN (?)
		GROUP BY i.date_create, i.number_invoice, i.patient_id
		ORDER BY i.date_create`

	rows, err := s.db.Raw(query, startDate, endDate, locationIDs).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []SalesBreakdownItem
	for rows.Next() {
		var date time.Time
		var invoiceNum string
		var patientID *int64
		var invTotal, cost, ophFr, sunPl, lOnly, bif, prog, contacts, ophRx, sunRx, perf, elec, other float64

		if err := rows.Scan(&date, &invoiceNum, &patientID, &invTotal, &cost,
			&ophFr, &sunPl, &lOnly, &bif, &prog, &contacts,
			&ophRx, &sunRx, &perf, &elec, &other); err != nil {
			return nil, err
		}

		items = append(items, SalesBreakdownItem{
			Date:          date.Format("2006-01-02"),
			InvoiceNumber: invoiceNum,
			PatientName:   patientID,
			InvoiceTotal:  round2(invTotal),
			Cost:          round2(cost),
			OphFrOnly:     round2(ophFr),
			SunPlano:      round2(sunPl),
			LOnly:         round2(lOnly),
			Bifocal:       round2(bif),
			Progressive:   round2(prog),
			Contacts:      round2(contacts),
			OphFrRx:       round2(ophRx),
			SunRx:         round2(sunRx),
			Perfume:       round2(perf),
			Electronics:   round2(elec),
			Other:         round2(other),
		})
	}
	return items, nil
}

func (s *Service) SalesByEmployee(locationIDs []int, startDate, endDate string, employeeID *int) ([]SalesByEmployeeItem, error) {
	query := `
		SELECT TRIM(CONCAT(e.first_name,' ',e.last_name)) AS employee,
		       COALESCE(SUM(iis.total),0)                  AS gross_sales,
		       COALESCE(SUM(i.pt_bal),0)                   AS pt_pay,
		       COALESCE(SUM(i.ins_bal),0)                  AS ins_pay,
		       COALESCE(SUM(i.final_amount),0)             AS net_sales,
		       COALESCE(SUM(i.discount),0)                 AS discount,
		       COALESCE(SUM(CASE WHEN iis.item_type='Frames' THEN iis.total ELSE 0 END),0) AS frames_sales,
		       COALESCE(SUM(CASE WHEN iis.item_type='Lens'   THEN iis.total ELSE 0 END),0) AS lens_sales,
		       COALESCE(SUM(CASE WHEN lt.type_name='PLANO'   THEN iis.total ELSE 0 END),0) AS plano_sales,
		       COALESCE(SUM(CASE WHEN lt.type_name='SINGLE'  THEN iis.total ELSE 0 END),0) AS sv_sales
		FROM employee e
		JOIN invoice i             ON i.employee_id  = e.id_employee
		JOIN invoice_item_sale iis ON iis.invoice_id = i.id_invoice
		LEFT JOIN lenses le        ON iis.item_id    = le.id_lenses
		LEFT JOIN lens_types lt    ON le.lens_type_id = lt.id_lens_type
		WHERE i.finalized = true
		  AND i.created_at BETWEEN ? AND ?
		  AND i.location_id IN (?)`

	args := []interface{}{startDate, endDate, locationIDs}
	if employeeID != nil {
		query += ` AND e.id_employee = ?`
		args = append(args, *employeeID)
	}

	query += `
		GROUP BY e.first_name, e.last_name
		ORDER BY employee`

	rows, err := s.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []SalesByEmployeeItem
	for rows.Next() {
		var it SalesByEmployeeItem
		if err := rows.Scan(&it.Employee, &it.GrossSales, &it.PtPay, &it.InsPay, &it.NetSales,
			&it.Discount, &it.FramesSales, &it.LensSales, &it.PlanoSales, &it.SingleVisionSales); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

func (s *Service) GiftCardBalance() ([]GiftCardBalanceItem, error) {
	query := `
		SELECT gc.created_at, gc.code, COALESCE(gc.balance,0)
		FROM gift_card gc
		ORDER BY gc.created_at`

	rows, err := s.db.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []GiftCardBalanceItem
	for rows.Next() {
		var createdAt *time.Time
		var code string
		var balance float64

		if err := rows.Scan(&createdAt, &code, &balance); err != nil {
			return nil, err
		}

		issuedDate := ""
		if createdAt != nil {
			issuedDate = createdAt.Format("2006-01-02")
		}

		items = append(items, GiftCardBalanceItem{
			IssuedDate: issuedDate,
			CardCode:   code,
			Balance:    round2(balance),
		})
	}
	return items, nil
}

func (s *Service) GiftCardDetailsInfo(cardCode string) (*GiftCardDetails, error) {
	// get card
	var card struct {
		IDGiftCard     int
		Code           string
		Nominal        float64
		Balance        float64
		InvoiceID      *int
		ExpirationDate *time.Time
	}
	err := s.db.Raw(`
		SELECT id_gift_card, code, COALESCE(nominal,0), COALESCE(balance,0),
		       invoice_id, expiration_date
		FROM gift_card WHERE code = ?`, cardCode).Row().Scan(
		&card.IDGiftCard, &card.Code, &card.Nominal, &card.Balance,
		&card.InvoiceID, &card.ExpirationDate)
	if err != nil {
		return nil, fmt.Errorf("gift card not found")
	}

	// purchase invoice number
	var purchaseInvNum *string
	if card.InvoiceID != nil {
		var num string
		if err := s.db.Raw(`SELECT number_invoice FROM invoice WHERE id_invoice = ?`, *card.InvoiceID).Row().Scan(&num); err == nil {
			purchaseInvNum = &num
		}
	}

	var expDate *string
	if card.ExpirationDate != nil {
		d := card.ExpirationDate.Format("2006-01-02")
		expDate = &d
	}

	details := &GiftCardDetails{
		IDGiftCard:            card.IDGiftCard,
		Code:                  card.Code,
		Nominal:               round2(card.Nominal),
		Balance:               round2(card.Balance),
		PurchaseInvoiceNumber: purchaseInvNum,
		ExpirationDate:        expDate,
	}

	// transactions
	txRows, err := s.db.Raw(`
		SELECT t.created_at, t.transaction_type, COALESCE(t.amount,0),
		       i.number_invoice,
		       TRIM(CONCAT(p.first_name,' ',p.last_name)) AS patient_name
		FROM gift_card_transaction t
		LEFT JOIN invoice i ON i.id_invoice = t.related_invoice_id
		LEFT JOIN patient p ON p.id_patient = t.processed_by_patient_id
		WHERE t.gift_card_id = ?
		ORDER BY t.created_at`, card.IDGiftCard).Rows()
	if err != nil {
		return details, nil
	}
	defer txRows.Close()

	for txRows.Next() {
		var createdAt *time.Time
		var txType string
		var amount float64
		var invNum, patientName *string

		if err := txRows.Scan(&createdAt, &txType, &amount, &invNum, &patientName); err != nil {
			continue
		}

		date := ""
		if createdAt != nil {
			date = createdAt.Format("2006-01-02")
		}

		details.Transactions = append(details.Transactions, GiftCardDetailsTx{
			Date:          date,
			InvoiceNumber: invNum,
			PatientName:   patientName,
			Action:        txType,
			Amount:        round2(amount),
		})
	}

	return details, nil
}

func (s *Service) GiftCardActivity(locationID *int, startDate, endDate time.Time) ([]GiftCardActivityItem, error) {
	query := `
		SELECT gc.code           AS card_code,
		       TRIM(CONCAT(p.first_name,' ',p.last_name)) AS patient_name,
		       t.created_at      AS date,
		       i.number_invoice  AS invoice_number,
		       t.transaction_type AS action,
		       COALESCE(t.amount,0) AS amount
		FROM gift_card_transaction t
		JOIN gift_card gc ON gc.id_gift_card = t.gift_card_id
		LEFT JOIN patient p  ON p.id_patient = t.processed_by_patient_id
		LEFT JOIN invoice i  ON i.id_invoice = t.related_invoice_id
		WHERE t.created_at >= ? AND t.created_at <= ?`

	args := []interface{}{startDate, endDate}
	if locationID != nil {
		query += ` AND gc.location_id = ?`
		args = append(args, *locationID)
	}
	query += ` ORDER BY t.created_at`

	rows, err := s.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []GiftCardActivityItem
	for rows.Next() {
		var cardCode string
		var patientName, invNum *string
		var date *time.Time
		var action string
		var amount float64

		if err := rows.Scan(&cardCode, &patientName, &date, &invNum, &action, &amount); err != nil {
			return nil, err
		}

		dateStr := ""
		if date != nil {
			dateStr = date.Format("2006-01-02")
		}

		var adding, using float64
		if action == "Adding Funds" {
			adding = round2(amount)
		} else if action == "GC Payment" {
			using = round2(amount)
		}

		items = append(items, GiftCardActivityItem{
			CardCode:      cardCode,
			PatientName:   patientName,
			Date:          dateStr,
			InvoiceNumber: invNum,
			Action:        action,
			Adding:        adding,
			Using:         using,
		})
	}
	return items, nil
}

func (s *Service) QuestionnaireReferral(locationID int, startDT, endDT time.Time) ([]ReferralCountItem, error) {
	// all referral sources with counts (LEFT JOIN)
	query := `
		SELECT rs.title,
		       COUNT(qr.id_questionnaire_referral) AS cnt
		FROM referral_sources rs
		LEFT JOIN questionnaire_referral qr
		       ON qr.referral_sources_id = rs.id_referral_sources
		      AND qr.location_id = ?
		      AND qr.datetime_created >= ?
		      AND qr.datetime_created <= ?
		GROUP BY rs.id_referral_sources, rs.title`

	rows, err := s.db.Raw(query, locationID, startDT, endDT).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ReferralCountItem
	for rows.Next() {
		var it ReferralCountItem
		if err := rows.Scan(&it.Title, &it.Count); err != nil {
			return nil, err
		}
		items = append(items, it)
	}

	// absent (referral_sources_id IS NULL)
	var absentCount int
	s.db.Raw(`
		SELECT COUNT(id_questionnaire_referral)
		FROM questionnaire_referral
		WHERE location_id = ?
		  AND datetime_created >= ?
		  AND datetime_created <= ?
		  AND referral_sources_id IS NULL`, locationID, startDT, endDT).Row().Scan(&absentCount)

	items = append(items, ReferralCountItem{Title: "[absent]", Count: absentCount})

	// sort by count desc, then title asc
	sortReferralItems(items)
	return items, nil
}

func (s *Service) QuestionnaireReasons(locationID int, startDT, endDT time.Time) ([]ReferralCountItem, error) {
	query := `
		SELECT vr.title,
		       COUNT(qr.id_questionnaire_referral) AS cnt
		FROM visit_reasons vr
		LEFT JOIN questionnaire_referral qr
		       ON qr.visit_reasons_id = vr.id_visit_reasons
		      AND qr.location_id = ?
		      AND qr.datetime_created >= ?
		      AND qr.datetime_created <= ?
		GROUP BY vr.id_visit_reasons, vr.title`

	rows, err := s.db.Raw(query, locationID, startDT, endDT).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ReferralCountItem
	for rows.Next() {
		var it ReferralCountItem
		if err := rows.Scan(&it.Title, &it.Count); err != nil {
			return nil, err
		}
		items = append(items, it)
	}

	var absentCount int
	s.db.Raw(`
		SELECT COUNT(id_questionnaire_referral)
		FROM questionnaire_referral
		WHERE location_id = ?
		  AND datetime_created >= ?
		  AND datetime_created <= ?
		  AND visit_reasons_id IS NULL`, locationID, startDT, endDT).Row().Scan(&absentCount)

	items = append(items, ReferralCountItem{Title: "[absent]", Count: absentCount})

	sortReferralItems(items)
	return items, nil
}

func sortReferralItems(items []ReferralCountItem) {
	// sort by count desc, then title asc (case-insensitive)
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].Count > items[i].Count ||
				(items[j].Count == items[i].Count &&
					strings.ToLower(items[j].Title) < strings.ToLower(items[i].Title)) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

// GetEmployeeLocationID returns the employee's location_id from JWT username
func (s *Service) GetEmployeeLocationID(username string) (int, error) {
	var locID int
	err := s.db.Raw(`
		SELECT e.location_id
		FROM employee e
		JOIN employee_login el ON el.id_employee_login = e.employee_login_id
		WHERE el.username = ?`, username).Row().Scan(&locID)
	if err != nil {
		return 0, fmt.Errorf("employee or location not found")
	}
	return locID, nil
}
