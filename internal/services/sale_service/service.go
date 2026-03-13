package sale_service

import (
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func ptrStr(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// ── types ───────────────────────────────────────────────────────────────────

type LocationItem struct {
	LocationID   int    `json:"location_id"`
	LocationName string `json:"location_name"`
}

type VendorItem struct {
	VendorID   int    `json:"vendor_id"`
	VendorName string `json:"vendor_name"`
}

type BrandItem struct {
	IDBrand   int    `json:"id_brand"`
	BrandName string `json:"brand_name"`
}

type EmployeeItem struct {
	EmployeeID int    `json:"employee_id"`
	Name       string `json:"name"`
}

type SaleItem struct {
	InvoiceNumber    string  `json:"invoice_number"`
	Date             *string `json:"date"`
	SaleKey          *string `json:"sale_key"`
	Description      string  `json:"description"`
	ItemType         *string `json:"item_type"`
	Quantity         string  `json:"quantity"`
	Price            string  `json:"price"`
	TotalAmount      string  `json:"total_amount"`
	PtBalance        string  `json:"pt_balance"`
	InsuranceBalance string  `json:"insurance_balance"`
	FinalAmount      string  `json:"final_amount"`
	DueAmount        string  `json:"due_amount"`
	PbCost           *string `json:"pb_cost"`
	PbSellingPrice   *string `json:"pb_selling_price"`
}

type YearlyRepItem struct {
	EmpID        int                `json:"emp_id"`
	EmployeeName string             `json:"employee_name"`
	Months       map[string]float64 `json:"months,omitempty"`
	Total        float64            `json:"Total"`
}

type ProfCodeItem struct {
	Desc     string `json:"desc"`
	MfrNbr   string `json:"mfr_nbr"`
	ItemNbr  string `json:"item_nbr"`
	Qty      int    `json:"qty"`
	UnitCost string `json:"unit_cost"`
	TCost    string `json:"t_cost"`
	TSales   string `json:"t_sales"`
}

type ProfCodeSummary struct {
	TotalQty   int    `json:"total_qty"`
	TotalSales string `json:"total_sales"`
	TotalCost  string `json:"total_cost"`
}

type InsuranceCompanyItem struct {
	InsuranceCompanyID int    `json:"insurance_company_id"`
	CompanyName        string `json:"company_name"`
}

type InsuranceReportItem struct {
	InvoiceID            int64   `json:"invoice_id"`
	InvoiceNumber        string  `json:"invoice_number"`
	DateCreate           *string `json:"date_create"`
	FinalAmount          string  `json:"final_amount"`
	InsBalance           string  `json:"ins_balance"`
	InsuranceCompanyName string  `json:"insurance_company_name"`
	InsurancePolicyID    int64   `json:"insurance_policy_id"`
	MemberNumber         *string `json:"member_number"`
}

type CommissionInvoiceItem struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type CommissionInvoice struct {
	InvoiceID string                  `json:"invoice_id"`
	Items     []CommissionInvoiceItem `json:"items"`
	Total     float64                 `json:"total"`
}

type CommissionEmployee struct {
	Employee          string              `json:"employee"`
	Locations         []string            `json:"locations"`
	Invoices          []CommissionInvoice `json:"invoices"`
	CommissionPercent float64             `json:"commission_percent"`
	CommissionTotal   float64             `json:"commission_total"`
}

type SalesReportEmp struct {
	Employee string  `json:"employee"`
	Total    float64 `json:"total"`
}

type ReferralItem struct {
	IDQuestionnaireReferral int64   `json:"id_questionnaire_referral"`
	CreatedAt               string  `json:"created_at"`
	PatientID               *int64  `json:"patient_id"`
	PatientName             string  `json:"patient_name"`
	LocationID              *int    `json:"location_id"`
	Location                *string `json:"location"`
	EmployeeID              *int64  `json:"employee_id"`
	EmployeeName            string  `json:"employee_name"`
	VisitReasonID           *int    `json:"visit_reason_id"`
	VisitReason             *string `json:"visit_reason"`
	ReferralSourceID        *int    `json:"referral_source_id"`
	ReferralSource          *string `json:"referral_source"`
}

// ── methods ─────────────────────────────────────────────────────────────────

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

func (s *Service) GetLocations() ([]LocationItem, error) {
	rows, err := s.db.Raw(`SELECT id_location, full_name FROM location ORDER BY full_name`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []LocationItem
	for rows.Next() {
		var it LocationItem
		if err := rows.Scan(&it.LocationID, &it.LocationName); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

func (s *Service) GetVendors() ([]VendorItem, error) {
	rows, err := s.db.Raw(`SELECT id_vendor, vendor_name FROM vendor ORDER BY vendor_name`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []VendorItem
	for rows.Next() {
		var it VendorItem
		if err := rows.Scan(&it.VendorID, &it.VendorName); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

func (s *Service) GetVendorBrands(vendorID int) ([]BrandItem, error) {
	query := `
		SELECT b.id_brand, b.brand_name
		FROM vendor_brand vb
		JOIN brand b ON b.id_brand = vb.id_brand
		WHERE vb.id_vendor = ?
		UNION ALL
		SELECT bl.id_brand_lens, bl.brand_name
		FROM vendor_brand_lens vbl
		JOIN brand_lens bl ON bl.id_brand_lens = vbl.id_brand_lens
		WHERE vbl.id_vendor = ?
		UNION ALL
		SELECT bcl.id_brand_contact_lens, bcl.brand_name
		FROM vendor_brand_contact_lens vbcl
		JOIN brand_contact_lens bcl ON bcl.id_brand_contact_lens = vbcl.id_brand_contact_lens
		WHERE vbcl.id_vendor = ?`

	rows, err := s.db.Raw(query, vendorID, vendorID, vendorID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []BrandItem
	for rows.Next() {
		var it BrandItem
		if err := rows.Scan(&it.IDBrand, &it.BrandName); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

func (s *Service) GetEmployees() ([]EmployeeItem, error) {
	rows, err := s.db.Raw(`
		SELECT id_employee, TRIM(CONCAT(first_name,' ',last_name))
		FROM employee ORDER BY first_name, last_name`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []EmployeeItem
	for rows.Next() {
		var it EmployeeItem
		if err := rows.Scan(&it.EmployeeID, &it.Name); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

// PB_KEY_MAP normalizes item_type strings
var PB_KEY_MAP = map[string]string{
	"frame": "Frames", "frames": "Frames",
	"lens": "Lens", "lenses": "Lens",
	"contact lens": "Contact Lens", "contact lenses": "Contact Lens",
	"contact_lenses": "Contact Lens", "contact_lens": "Contact Lens",
	"prof. service": "Prof. service", "prof_service": "Prof. service",
	"professional service": "Prof. service",
	"treatment": "Treatment", "treatments": "Treatment",
	"add service": "Add service", "add_service": "Add service",
	"misc": "Misc",
}

func (s *Service) GetSaleItems(
	locationID int, dateStart, dateEnd time.Time,
	employeeID *int, showInterCompany bool,
	pbKey *string, invoiceContains *string,
	saleKeyFilter *string, saleKeyContains *string,
	sunOnly bool, totalBy string,
	minBal, maxBal *float64,
	vendorID, brandID *int,
) ([]SaleItem, error) {

	query := `
		SELECT i.number_invoice,
		       i.date_create,
		       iis.sale_key,
		       iis.description,
		       iis.item_type,
		       iis.quantity,
		       iis.price,
		       COALESCE(i.total_amount,0),
		       COALESCE(i.pt_bal,0),
		       COALESCE(i.ins_bal,0),
		       COALESCE(i.final_amount,0),
		       COALESCE(i.due,0),
		       pb.pb_cost,
		       pb.pb_selling_price
		FROM invoice i
		JOIN invoice_item_sale iis ON iis.invoice_id = i.id_invoice
		LEFT JOIN inventory inv
		       ON LOWER(TRIM(CAST(iis.item_type AS TEXT))) = 'frames'
		      AND inv.id_inventory = iis.item_id
		LEFT JOIN price_book pb ON pb.inventory_id = inv.id_inventory`

	// additional joins for vendor/brand filtering
	needProductJoin := false
	needLensJoin := false
	needCLJoin := false

	effectivePbKey := ""
	if pbKey != nil {
		effectivePbKey = *pbKey
	}

	if vendorID != nil || brandID != nil {
		if effectivePbKey == "" {
			effectivePbKey = "Frames"
		}
		switch effectivePbKey {
		case "Lens":
			needLensJoin = true
		case "Contact Lens":
			needCLJoin = true
		default:
			needProductJoin = true
		}
	}

	if needProductJoin {
		query += `
		JOIN model m ON m.id_model = inv.model_id
		JOIN product p ON p.id_product = m.product_id`
	}
	if needLensJoin {
		query += `
		JOIN lenses le ON le.id_lenses = iis.item_id`
	}
	if needCLJoin {
		query += `
		JOIN contact_lens_item cli ON cli.id_contact_lens_item = iis.item_id`
	}

	query += `
		WHERE i.location_id = ?
		  AND i.date_create BETWEEN ? AND ?`

	args := []interface{}{locationID, dateStart, dateEnd}

	if employeeID != nil {
		query += ` AND i.employee_id = ?`
		args = append(args, *employeeID)
	}

	if !showInterCompany {
		query += ` AND i.number_invoice NOT ILIKE 'I%'`
	}

	if effectivePbKey != "" {
		query += ` AND LOWER(TRIM(CAST(iis.item_type AS TEXT))) = ?`
		args = append(args, strings.ToLower(effectivePbKey))
	}

	if invoiceContains != nil {
		query += ` AND iis.description ILIKE ?`
		args = append(args, "%"+*invoiceContains+"%")
	}

	if saleKeyFilter != nil {
		query += ` AND LOWER(TRIM(CAST(iis.sale_key AS TEXT))) = ?`
		args = append(args, strings.ToLower(*saleKeyFilter))
	}

	if saleKeyContains != nil {
		query += ` AND LOWER(TRIM(CAST(iis.sale_key AS TEXT))) LIKE ?`
		args = append(args, "%"+strings.ToLower(*saleKeyContains)+"%")
	}

	if sunOnly {
		query += ` AND iis.description ILIKE '%SUN%'`
	}

	if totalBy == "pt_bal" {
		query += ` AND i.pt_bal > 0`
	} else if totalBy == "ins_bal" {
		query += ` AND i.ins_bal > 0`
	}

	if minBal != nil {
		query += ` AND i.final_amount >= ?`
		args = append(args, *minBal)
	}
	if maxBal != nil {
		query += ` AND i.final_amount <= ?`
		args = append(args, *maxBal)
	}

	if vendorID != nil {
		switch {
		case needProductJoin:
			query += ` AND p.vendor_id = ?`
		case needLensJoin:
			query += ` AND le.vendor_id = ?`
		case needCLJoin:
			query += ` AND cli.vendor_id = ?`
		}
		args = append(args, *vendorID)
	}

	if brandID != nil {
		switch {
		case needProductJoin:
			query += ` AND p.brand_id = ?`
		case needLensJoin:
			query += ` AND le.brand_lens_id = ?`
		case needCLJoin:
			query += ` AND cli.brand_contact_lens_id = ?`
		}
		args = append(args, *brandID)
	}

	query += ` ORDER BY i.date_create DESC`

	rows, err := s.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []SaleItem
	for rows.Next() {
		var (
			invNum, desc                  string
			dateCreate                    *time.Time
			saleKey, itemType             *string
			qty                           int
			price                         float64
			totalAmt, ptBal, insBal       float64
			finalAmt, due                 float64
			pbCost, pbSelling             *float64
		)
		if err := rows.Scan(&invNum, &dateCreate, &saleKey, &desc, &itemType,
			&qty, &price, &totalAmt, &ptBal, &insBal, &finalAmt, &due,
			&pbCost, &pbSelling); err != nil {
			return nil, err
		}

		var dateStr *string
		if dateCreate != nil {
			d := dateCreate.Format("2006-01-02T15:04:05")
			dateStr = &d
		}

		var trimmedSaleKey *string
		if saleKey != nil {
			s := strings.TrimSpace(*saleKey)
			trimmedSaleKey = &s
		}

		var trimmedItemType *string
		if itemType != nil {
			s := strings.TrimSpace(*itemType)
			trimmedItemType = &s
		}

		var pbCostStr, pbSellingStr *string
		if pbCost != nil {
			s := fmt.Sprintf("%.2f", *pbCost)
			pbCostStr = &s
		}
		if pbSelling != nil {
			s := fmt.Sprintf("%.2f", *pbSelling)
			pbSellingStr = &s
		}

		items = append(items, SaleItem{
			InvoiceNumber:    invNum,
			Date:             dateStr,
			SaleKey:          trimmedSaleKey,
			Description:      desc,
			ItemType:         trimmedItemType,
			Quantity:         fmt.Sprintf("%d", qty),
			Price:            fmt.Sprintf("%.2f", price),
			TotalAmount:      fmt.Sprintf("%.2f", totalAmt),
			PtBalance:        fmt.Sprintf("%.2f", ptBal),
			InsuranceBalance: fmt.Sprintf("%.2f", insBal),
			FinalAmount:      fmt.Sprintf("%.2f", finalAmt),
			DueAmount:        fmt.Sprintf("%.2f", due),
			PbCost:           pbCostStr,
			PbSellingPrice:   pbSellingStr,
		})
	}
	return items, nil
}

func (s *Service) YearlyComparisonByRep(
	locationID int, yearStart, yearEnd, startMonth, endMonth int,
) (active []YearlyRepItem, inactive []YearlyRepItem, err error) {

	startDate := time.Date(yearStart, time.Month(startMonth), 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(yearEnd, time.Month(endMonth), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, -1)

	query := `
		SELECT e.id_employee,
		       e.first_name, e.last_name, e.active,
		       EXTRACT(YEAR FROM i.date_create)  AS yr,
		       EXTRACT(MONTH FROM i.date_create) AS mo,
		       COALESCE(SUM(iis.quantity),0)      AS qty_sold
		FROM employee e
		JOIN invoice i             ON i.employee_id  = e.id_employee
		JOIN invoice_item_sale iis ON iis.invoice_id = i.id_invoice
		WHERE i.location_id = ?
		  AND i.date_create >= ? AND i.date_create <= ?
		GROUP BY e.id_employee, e.first_name, e.last_name, e.active,
		         EXTRACT(YEAR FROM i.date_create), EXTRACT(MONTH FROM i.date_create)
		ORDER BY e.last_name, e.first_name`

	rows, err := s.db.Raw(query, locationID, startDate, endDate).Rows()
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	type empData struct {
		empID     int
		firstName string
		lastName  string
		isActive  bool
		sales     map[string]float64 // "Jan 23" -> qty
	}
	empMap := make(map[int]*empData)

	for rows.Next() {
		var id int
		var first, last string
		var act bool
		var yr, mo int
		var qty float64
		if err := rows.Scan(&id, &first, &last, &act, &yr, &mo, &qty); err != nil {
			return nil, nil, err
		}
		ed, ok := empMap[id]
		if !ok {
			ed = &empData{empID: id, firstName: first, lastName: last, isActive: act, sales: make(map[string]float64)}
			empMap[id] = ed
		}
		label := monthLabel(yr, mo)
		ed.sales[label] = qty
	}

	// generate month labels
	ymPairs := generateYMPairs(yearStart, startMonth, yearEnd, endMonth)

	for _, ed := range empMap {
		item := YearlyRepItem{
			EmpID:        ed.empID,
			EmployeeName: ed.firstName + " " + ed.lastName,
			Months:       make(map[string]float64),
		}
		var total float64
		for _, ym := range ymPairs {
			label := monthLabel(ym[0], ym[1])
			val := ed.sales[label]
			item.Months[label] = val
			total += val
		}
		item.Total = total
		if ed.isActive {
			active = append(active, item)
		} else {
			inactive = append(inactive, item)
		}
	}
	return active, inactive, nil
}

func (s *Service) ProfessionalCodes(locationID int, startDate, endDate time.Time) ([]ProfCodeItem, ProfCodeSummary, error) {
	query := `
		SELECT COALESCE(ps.invoice_desc,'') AS description,
		       COALESCE(ps.item_number,'')  AS item_number,
		       COALESCE(ps.mfr_number,'')   AS mfr_number,
		       COALESCE(ps.price,0)         AS unit_cost,
		       COALESCE(SUM(iis.quantity),0) AS qty_sold,
		       COALESCE(SUM(iis.total),0)    AS total_sold
		FROM professional_service ps
		JOIN invoice_item_sale iis ON iis.item_id = ps.id_professional_service
		JOIN invoice i ON i.id_invoice = iis.invoice_id
		WHERE iis.item_type = 'Prof. service'
		  AND i.location_id = ?
		  AND i.date_create BETWEEN ? AND ?
		GROUP BY ps.invoice_desc, ps.item_number, ps.mfr_number, ps.price
		ORDER BY qty_sold DESC`

	rows, err := s.db.Raw(query, locationID, startDate, endDate).Rows()
	if err != nil {
		return nil, ProfCodeSummary{}, err
	}
	defer rows.Close()

	var items []ProfCodeItem
	var totalQty int
	var totalSales, totalCost float64

	for rows.Next() {
		var desc, itemNbr, mfrNbr string
		var unitCost float64
		var qty int
		var tSales float64
		if err := rows.Scan(&desc, &itemNbr, &mfrNbr, &unitCost, &qty, &tSales); err != nil {
			return nil, ProfCodeSummary{}, err
		}
		tCost := round2(unitCost * float64(qty))
		items = append(items, ProfCodeItem{
			Desc:     desc,
			MfrNbr:   mfrNbr,
			ItemNbr:  itemNbr,
			Qty:      qty,
			UnitCost: fmt.Sprintf("%.2f", unitCost),
			TCost:    fmt.Sprintf("%.2f", tCost),
			TSales:   fmt.Sprintf("%.2f", tSales),
		})
		totalQty += qty
		totalSales += tSales
		totalCost += tCost
	}

	summary := ProfCodeSummary{
		TotalQty:   totalQty,
		TotalSales: fmt.Sprintf("%.2f", totalSales),
		TotalCost:  fmt.Sprintf("%.2f", totalCost),
	}
	return items, summary, nil
}

func (s *Service) YearlyComparisonByBrand(
	locationID *int, vendorID, brandID *int,
	yearStart, yearEnd int,
) ([]map[string]interface{}, error) {

	startDate := time.Date(yearStart, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(yearEnd, 12, 31, 0, 0, 0, 0, time.UTC)

	// determine brand type for filtering
	var brandType string
	if brandID != nil {
		var exists bool
		s.db.Raw(`SELECT EXISTS(SELECT 1 FROM brand WHERE id_brand = ?)`, *brandID).Row().Scan(&exists)
		if exists {
			brandType = "frames"
		} else {
			s.db.Raw(`SELECT EXISTS(SELECT 1 FROM brand_lens WHERE id_brand_lens = ?)`, *brandID).Row().Scan(&exists)
			if exists {
				brandType = "lens"
			} else {
				s.db.Raw(`SELECT EXISTS(SELECT 1 FROM brand_contact_lens WHERE id_brand_contact_lens = ?)`, *brandID).Row().Scan(&exists)
				if exists {
					brandType = "contact_lens"
				}
			}
		}
	}

	// Build UNION ALL query
	parts := []string{}
	args := []interface{}{}

	// --- frames ---
	if brandID == nil || brandType == "frames" {
		fq := `
		SELECT b.id_brand AS brand_id, b.brand_name,
		       EXTRACT(YEAR FROM i.date_create) AS yr,
		       EXTRACT(MONTH FROM i.date_create) AS mo,
		       COALESCE(SUM(iis.quantity),0) AS qty_sold
		FROM invoice i
		JOIN invoice_item_sale iis ON iis.invoice_id = i.id_invoice
		JOIN inventory inv ON inv.id_inventory = iis.item_id
		JOIN model m ON m.id_model = inv.model_id
		JOIN product p ON p.id_product = m.product_id
		JOIN brand b ON b.id_brand = p.brand_id
		WHERE i.date_create BETWEEN ? AND ?
		  AND iis.item_type = 'Frames'`
		fArgs := []interface{}{startDate, endDate}
		if locationID != nil {
			fq += ` AND i.location_id = ?`
			fArgs = append(fArgs, *locationID)
		}
		if vendorID != nil {
			fq += ` AND p.vendor_id = ?`
			fArgs = append(fArgs, *vendorID)
		}
		if brandID != nil && brandType == "frames" {
			fq += ` AND b.id_brand = ?`
			fArgs = append(fArgs, *brandID)
		}
		fq += ` GROUP BY b.id_brand, b.brand_name, EXTRACT(YEAR FROM i.date_create), EXTRACT(MONTH FROM i.date_create)`
		parts = append(parts, fq)
		args = append(args, fArgs...)
	}

	// --- lenses ---
	if brandID == nil || brandType == "lens" {
		lq := `
		SELECT bl.id_brand_lens AS brand_id, bl.brand_name,
		       EXTRACT(YEAR FROM i.date_create) AS yr,
		       EXTRACT(MONTH FROM i.date_create) AS mo,
		       COALESCE(SUM(iis.quantity),0) AS qty_sold
		FROM invoice i
		JOIN invoice_item_sale iis ON iis.invoice_id = i.id_invoice
		JOIN lenses le ON le.id_lenses = iis.item_id
		JOIN brand_lens bl ON bl.id_brand_lens = le.brand_lens_id
		WHERE i.date_create BETWEEN ? AND ?
		  AND iis.item_type = 'Lens'`
		lArgs := []interface{}{startDate, endDate}
		if locationID != nil {
			lq += ` AND i.location_id = ?`
			lArgs = append(lArgs, *locationID)
		}
		if vendorID != nil {
			lq += ` AND le.vendor_id = ?`
			lArgs = append(lArgs, *vendorID)
		}
		if brandID != nil && brandType == "lens" {
			lq += ` AND bl.id_brand_lens = ?`
			lArgs = append(lArgs, *brandID)
		}
		lq += ` GROUP BY bl.id_brand_lens, bl.brand_name, EXTRACT(YEAR FROM i.date_create), EXTRACT(MONTH FROM i.date_create)`
		parts = append(parts, lq)
		args = append(args, lArgs...)
	}

	// --- contact lenses ---
	if brandID == nil || brandType == "contact_lens" {
		cq := `
		SELECT bcl.id_brand_contact_lens AS brand_id, bcl.brand_name,
		       EXTRACT(YEAR FROM i.date_create) AS yr,
		       EXTRACT(MONTH FROM i.date_create) AS mo,
		       COALESCE(SUM(iis.quantity),0) AS qty_sold
		FROM invoice i
		JOIN invoice_item_sale iis ON iis.invoice_id = i.id_invoice
		JOIN contact_lens_item cli ON cli.id_contact_lens_item = iis.item_id
		JOIN brand_contact_lens bcl ON bcl.id_brand_contact_lens = cli.brand_contact_lens_id
		WHERE i.date_create BETWEEN ? AND ?
		  AND iis.item_type = 'Contact Lens'`
		cArgs := []interface{}{startDate, endDate}
		if locationID != nil {
			cq += ` AND i.location_id = ?`
			cArgs = append(cArgs, *locationID)
		}
		if vendorID != nil {
			cq += ` AND cli.vendor_id = ?`
			cArgs = append(cArgs, *vendorID)
		}
		if brandID != nil && brandType == "contact_lens" {
			cq += ` AND bcl.id_brand_contact_lens = ?`
			cArgs = append(cArgs, *brandID)
		}
		cq += ` GROUP BY bcl.id_brand_contact_lens, bcl.brand_name, EXTRACT(YEAR FROM i.date_create), EXTRACT(MONTH FROM i.date_create)`
		parts = append(parts, cq)
		args = append(args, cArgs...)
	}

	if len(parts) == 0 {
		return []map[string]interface{}{}, nil
	}

	unionQuery := `SELECT brand_id, brand_name, yr, mo, SUM(qty_sold) AS qty_sold FROM (` +
		strings.Join(parts, " UNION ALL ") +
		`) AS unioned GROUP BY brand_id, brand_name, yr, mo ORDER BY brand_name`

	rows, err := s.db.Raw(unionQuery, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// pivot
	type brandData struct {
		brandID   int
		brandName string
		sales     map[string]float64
	}
	brandMap := make(map[int]*brandData)
	var brandOrder []int

	for rows.Next() {
		var bID, yr, mo int
		var bName string
		var qty float64
		if err := rows.Scan(&bID, &bName, &yr, &mo, &qty); err != nil {
			return nil, err
		}
		bd, ok := brandMap[bID]
		if !ok {
			bd = &brandData{brandID: bID, brandName: bName, sales: make(map[string]float64)}
			brandMap[bID] = bd
			brandOrder = append(brandOrder, bID)
		}
		label := monthLabel(yr, mo)
		bd.sales[label] = qty
	}

	ymPairs := generateYMPairs(yearStart, 1, yearEnd, 12)
	var result []map[string]interface{}

	for _, bID := range brandOrder {
		bd := brandMap[bID]
		obj := map[string]interface{}{
			"brand_id":   bd.brandID,
			"brand_name": bd.brandName,
		}
		var total float64
		for _, ym := range ymPairs {
			label := monthLabel(ym[0], ym[1])
			qty := bd.sales[label]
			obj[label] = qty
			total += qty
		}
		obj["Total"] = total
		result = append(result, obj)
	}
	return result, nil
}

func (s *Service) GetInsuranceCompanies() ([]InsuranceCompanyItem, error) {
	rows, err := s.db.Raw(`SELECT id_insurance_company, company_name FROM insurance_company ORDER BY company_name`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []InsuranceCompanyItem
	for rows.Next() {
		var it InsuranceCompanyItem
		if err := rows.Scan(&it.InsuranceCompanyID, &it.CompanyName); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

func (s *Service) InsuranceReport(locationID *int, startDate, endDate time.Time, insuranceID *int) ([]InsuranceReportItem, error) {
	query := `
		SELECT i.id_invoice, i.number_invoice, i.date_create,
		       COALESCE(i.final_amount,0), COALESCE(i.ins_bal,0),
		       ic.company_name,
		       ip.id_insurance_policy, ip.group_number
		FROM invoice i
		JOIN invoice_insurance_policy iip ON iip.invoice_id = i.id_invoice
		JOIN insurance_policy ip ON ip.id_insurance_policy = iip.insurance_policy_id
		JOIN insurance_company ic ON ic.id_insurance_company = ip.insurance_company_id
		WHERE i.date_create >= ? AND i.date_create <= ?`

	args := []interface{}{startDate, endDate}
	if locationID != nil {
		query += ` AND i.location_id = ?`
		args = append(args, *locationID)
	}
	if insuranceID != nil {
		query += ` AND ip.id_insurance_policy = ?`
		args = append(args, *insuranceID)
	}
	query += ` ORDER BY i.date_create`

	rows, err := s.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []InsuranceReportItem
	for rows.Next() {
		var invID, policyID int64
		var invNum, companyName string
		var dateCreate *time.Time
		var finalAmt, insBal float64
		var groupNum *string

		if err := rows.Scan(&invID, &invNum, &dateCreate, &finalAmt, &insBal,
			&companyName, &policyID, &groupNum); err != nil {
			return nil, err
		}

		var dateStr *string
		if dateCreate != nil {
			d := dateCreate.Format("2006-01-02T15:04:05")
			dateStr = &d
		}

		items = append(items, InsuranceReportItem{
			InvoiceID:            invID,
			InvoiceNumber:        invNum,
			DateCreate:           dateStr,
			FinalAmount:          fmt.Sprintf("%.2f", finalAmt),
			InsBalance:           fmt.Sprintf("%.2f", insBal),
			InsuranceCompanyName: companyName,
			InsurancePolicyID:    policyID,
			MemberNumber:         groupNum,
		})
	}
	return items, nil
}

func (s *Service) CommissionReport(locationID *int, startDate, endDate time.Time) ([]CommissionEmployee, error) {
	// Get invoices
	query := `
		SELECT i.id_invoice, i.number_invoice, COALESCE(i.final_amount,0),
		       e.id_employee, e.first_name, e.last_name,
		       COALESCE(l.full_name,'Unknown')
		FROM invoice i
		JOIN employee e ON e.id_employee = i.employee_id
		LEFT JOIN location l ON l.id_location = i.location_id
		WHERE i.number_invoice LIKE 'S%'
		  AND i.created_at >= ? AND i.created_at <= ?`

	args := []interface{}{startDate, endDate}
	if locationID != nil {
		query += ` AND i.location_id = ?`
		args = append(args, *locationID)
	}
	query += ` ORDER BY i.created_at`

	rows, err := s.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type empAcc struct {
		empID      int
		name       string
		locations  map[string]struct{}
		invoices   []CommissionInvoice
		commPct    float64
		commTotal  float64
	}
	empMap := make(map[int]*empAcc)
	var empOrder []int

	for rows.Next() {
		var invID int64
		var invNum string
		var finalAmt float64
		var empID int
		var first, last, locName string

		if err := rows.Scan(&invID, &invNum, &finalAmt, &empID, &first, &last, &locName); err != nil {
			return nil, err
		}

		acc, ok := empMap[empID]
		if !ok {
			// get commission percent
			var commPct float64
			s.db.Raw(`
				SELECT COALESCE(commission_percent,0) FROM employee_commissions
				WHERE employee_id = ? ORDER BY created_at DESC LIMIT 1`, empID).Row().Scan(&commPct)

			acc = &empAcc{
				empID:     empID,
				name:      first + " " + last,
				locations: make(map[string]struct{}),
				commPct:   commPct,
			}
			empMap[empID] = acc
			empOrder = append(empOrder, empID)
		}
		acc.locations[locName] = struct{}{}

		// get invoice items
		itemRows, err := s.db.Raw(`
			SELECT description, COALESCE(total,0) FROM invoice_item_sale WHERE invoice_id = ?`, invID).Rows()
		if err != nil {
			continue
		}
		var invItems []CommissionInvoiceItem
		for itemRows.Next() {
			var desc string
			var amt float64
			if err := itemRows.Scan(&desc, &amt); err != nil {
				continue
			}
			invItems = append(invItems, CommissionInvoiceItem{Name: desc, Amount: amt})
		}
		itemRows.Close()

		acc.invoices = append(acc.invoices, CommissionInvoice{
			InvoiceID: invNum,
			Items:     invItems,
			Total:     finalAmt,
		})
		rate := acc.commPct / 100.0
		acc.commTotal += finalAmt * rate
	}

	var result []CommissionEmployee
	for _, empID := range empOrder {
		acc := empMap[empID]
		var locs []string
		for l := range acc.locations {
			locs = append(locs, l)
		}
		result = append(result, CommissionEmployee{
			Employee:          acc.name,
			Locations:         locs,
			Invoices:          acc.invoices,
			CommissionPercent: acc.commPct,
			CommissionTotal:   round2(acc.commTotal),
		})
	}
	return result, nil
}

func (s *Service) SalesReport(locationID int) (map[string][]SalesReportEmp, error) {
	today := time.Now().UTC()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	startOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	startOfYear := time.Date(today.Year(), 1, 1, 0, 0, 0, 0, time.UTC)

	fetchSummary := func(since time.Time) ([]SalesReportEmp, error) {
		rows, err := s.db.Raw(`
			SELECT TRIM(CONCAT(e.first_name,' ',e.last_name)) AS emp,
			       COALESCE(SUM(i.final_amount),0) AS total
			FROM invoice i
			JOIN employee e ON e.id_employee = i.employee_id
			WHERE i.number_invoice LIKE 'S%'
			  AND i.location_id = ?
			  AND i.created_at >= ?
			GROUP BY e.first_name, e.last_name
			ORDER BY emp`, locationID, since).Rows()
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var items []SalesReportEmp
		for rows.Next() {
			var it SalesReportEmp
			if err := rows.Scan(&it.Employee, &it.Total); err != nil {
				return nil, err
			}
			items = append(items, it)
		}
		return items, nil
	}

	todayData, err := fetchSummary(todayDate)
	if err != nil {
		return nil, err
	}
	monthData, err := fetchSummary(startOfMonth)
	if err != nil {
		return nil, err
	}
	yearData, err := fetchSummary(startOfYear)
	if err != nil {
		return nil, err
	}

	return map[string][]SalesReportEmp{
		"today":         todayData,
		"current_month": monthData,
		"current_year":  yearData,
	}, nil
}

func (s *Service) ReferralReport(
	allowedIDs []int,
	startDate, endDate *time.Time,
	visitReasonID, referralSourceID *int,
) ([]ReferralItem, error) {

	query := `
		SELECT qr.id_questionnaire_referral,
		       qr.datetime_created,
		       qr.patient_id,
		       qr.location_id,
		       qr.employee_id,
		       COALESCE(p.first_name,'') AS patient_first,
		       COALESCE(p.last_name,'')  AS patient_last,
		       COALESCE(e.first_name,'') AS emp_first,
		       COALESCE(e.last_name,'')  AS emp_last,
		       l.full_name               AS location_name,
		       vr.id_visit_reasons       AS visit_reason_id,
		       vr.title                  AS visit_reason_title,
		       rs.id_referral_sources    AS referral_source_id,
		       rs.title                  AS referral_source_title
		FROM questionnaire_referral qr
		LEFT JOIN patient p  ON p.id_patient   = qr.patient_id
		LEFT JOIN employee e ON e.id_employee  = qr.employee_id
		LEFT JOIN location l ON l.id_location  = qr.location_id
		LEFT JOIN visit_reasons vr ON vr.id_visit_reasons = qr.visit_reasons_id
		LEFT JOIN referral_sources rs ON rs.id_referral_sources = qr.referral_sources_id
		WHERE qr.location_id IN (?)`

	args := []interface{}{allowedIDs}

	if startDate != nil {
		query += ` AND qr.datetime_created >= ?`
		args = append(args, *startDate)
	}
	if endDate != nil {
		query += ` AND qr.datetime_created <= ?`
		args = append(args, *endDate)
	}
	if visitReasonID != nil {
		query += ` AND qr.visit_reasons_id = ?`
		args = append(args, *visitReasonID)
	}
	if referralSourceID != nil {
		query += ` AND qr.referral_sources_id = ?`
		args = append(args, *referralSourceID)
	}

	query += ` ORDER BY qr.datetime_created DESC`

	rows, err := s.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ReferralItem
	for rows.Next() {
		var (
			id                                                           int64
			created                                                      *time.Time
			patientID, employeeID                                        *int64
			locationID, visitReasonIDVal, referralSourceIDVal             *int
			patientFirst, patientLast, empFirst, empLast                 string
			locationName, visitReasonTitle, referralSourceTitle           *string
		)
		if err := rows.Scan(&id, &created, &patientID, &locationID, &employeeID,
			&patientFirst, &patientLast, &empFirst, &empLast,
			&locationName, &visitReasonIDVal, &visitReasonTitle,
			&referralSourceIDVal, &referralSourceTitle); err != nil {
			return nil, err
		}

		createdStr := ""
		if created != nil {
			createdStr = created.Format(time.RFC3339)
		}

		items = append(items, ReferralItem{
			IDQuestionnaireReferral: id,
			CreatedAt:               createdStr,
			PatientID:               patientID,
			PatientName:             strings.TrimSpace(patientFirst + " " + patientLast),
			LocationID:              locationID,
			Location:                locationName,
			EmployeeID:              employeeID,
			EmployeeName:            strings.TrimSpace(empFirst + " " + empLast),
			VisitReasonID:           visitReasonIDVal,
			VisitReason:             visitReasonTitle,
			ReferralSourceID:        referralSourceIDVal,
			ReferralSource:          referralSourceTitle,
		})
	}
	return items, nil
}

// ── helpers ─────────────────────────────────────────────────────────────────

var monthNames = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

var MONTH_MAP = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

func ParseMonth(s string, def int) int {
	if s == "" {
		return def
	}
	if m, ok := MONTH_MAP[strings.ToLower(s)]; ok {
		return m
	}
	return def
}

func ParseYearRange(s string) (int, int) {
	if s == "" {
		y := time.Now().Year()
		return y, y
	}
	parts := strings.Split(s, "-")
	if len(parts) == 1 {
		y := 0
		fmt.Sscanf(parts[0], "%d", &y)
		return y, y
	}
	var y1, y2 int
	fmt.Sscanf(parts[0], "%d", &y1)
	fmt.Sscanf(parts[1], "%d", &y2)
	return y1, y2
}

func monthLabel(y, m int) string {
	return fmt.Sprintf("%s %s", monthNames[m-1], fmt.Sprintf("%d", y)[2:])
}

func generateYMPairs(yearStart, monthStart, yearEnd, monthEnd int) [][2]int {
	var pairs [][2]int
	y, m := yearStart, monthStart
	for {
		if y > yearEnd || (y == yearEnd && m > monthEnd) {
			break
		}
		pairs = append(pairs, [2]int{y, m})
		m++
		if m > 12 {
			m = 1
			y++
		}
	}
	return pairs
}
