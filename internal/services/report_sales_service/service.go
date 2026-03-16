package report_sales_service

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

func round2(v float64) float64 { return math.Round(v*100) / 100 }

// ── types ───────────────────────────────────────────────────────────────────

type BreakdownRow struct {
	Date            string  `json:"date"`
	Invoice         string  `json:"invoice"`
	Name            string  `json:"name"`
	InvTotal        float64 `json:"inv_total"`
	Cost            float64 `json:"cost"`
	OphFrOnly       float64 `json:"oph_fr_only"`
	OphFrRx         float64 `json:"oph_fr_rx"`
	SunPlano        float64 `json:"sun_plano"`
	SunRx           float64 `json:"sun_rx"`
	LOnly           float64 `json:"l_only"`
	ProfFees        float64 `json:"prof_fees"`
	Contacts        float64 `json:"contacts"`
	Other           float64 `json:"other"`
	Discount        float64 `json:"discount"`
	Tax             float64 `json:"tax"`
	Employee        string  `json:"employee"`
	CostInfoMissing string  `json:"cost_info_missing"`
}

type BreakdownTotals struct {
	InvTotal  float64 `json:"inv_total"`
	Cost      float64 `json:"cost"`
	OphFrOnly float64 `json:"oph_fr_only"`
	OphFrRx   float64 `json:"oph_fr_rx"`
	SunPlano  float64 `json:"sun_plano"`
	SunRx     float64 `json:"sun_rx"`
	LOnly     float64 `json:"l_only"`
	ProfFees  float64 `json:"prof_fees"`
	Contacts  float64 `json:"contacts"`
	Other     float64 `json:"other"`
	Discount  float64 `json:"discount"`
	Tax       float64 `json:"tax"`
}

type BreakdownPercent struct {
	OphFrOnly float64 `json:"oph_fr_only"`
	OphFrRx   float64 `json:"oph_fr_rx"`
	SunPlano  float64 `json:"sun_plano"`
	SunRx     float64 `json:"sun_rx"`
	LOnly     float64 `json:"l_only"`
	ProfFees  float64 `json:"prof_fees"`
	Contacts  float64 `json:"contacts"`
	Other     float64 `json:"other"`
}

type BreakdownResult struct {
	Data            []BreakdownRow   `json:"data"`
	Totals          BreakdownTotals  `json:"totals"`
	PercentNetOfTax BreakdownPercent `json:"percent_net_of_tax"`
	RecordCount     int              `json:"record_count"`
}

// ── raw row from SQL ────────────────────────────────────────────────────────

type rawRow struct {
	IDInvoice     int64
	NumberInvoice string
	DateCreate    time.Time
	InvDiscount   float64
	TaxAmount     float64
	EmployeeID    int64

	PatientFirst *string
	PatientLast  *string
	EmpFirst     *string
	EmpLast      *string

	IDInvoiceSale int64
	ItemType      *string
	ItemID        *int64
	ItemTotal     float64
	ItemCost      float64

	IsSunglass   *bool
	HasLabTicket bool
}

// ── invoice accumulator ─────────────────────────────────────────────────────

type invoiceAcc struct {
	date         time.Time
	invoice      string
	name         string
	employee     string
	invTotal     float64
	cost         float64
	ophFrOnly    float64
	ophFrRx      float64
	sunPlano     float64
	sunRx        float64
	lOnly        float64
	profFees     float64
	contacts     float64
	other        float64
	discount     float64
	tax          float64
	hasLabTicket bool
	zeroCostCats map[string]struct{}
	order        int // preserve insertion order
}

// GetEmployeeLocationID returns employee's location_id from JWT username
func (s *Service) GetEmployeeLocationID(username string) (int, error) {
	var locID int
	err := s.db.Raw(`
		SELECT e.location_id
		FROM employee e
		JOIN employee_login el ON el.id_employee_login = e.employee_login_id
		WHERE el.employee_login = ?`, username).Row().Scan(&locID)
	if err != nil {
		return 0, fmt.Errorf("employee or location not found")
	}
	return locID, nil
}

// ── Breakdown ───────────────────────────────────────────────────────────────

func (s *Service) Breakdown(
	locationIDs []int,
	dateStart, dateEnd time.Time,
	employeeID *int,
	showInterCompany bool,
) (*BreakdownResult, error) {

	query := `
		SELECT
			i.id_invoice,
			i.number_invoice,
			i.date_create,
			COALESCE(i.discount, 0)   AS inv_discount,
			COALESCE(i.tax_amount, 0) AS tax_amount,
			i.employee_id,

			p.first_name AS patient_first,
			p.last_name  AS patient_last,
			e.first_name AS emp_first,
			e.last_name  AS emp_last,

			iis.id_invoice_sale,
			iis.item_type,
			iis.item_id,
			COALESCE(iis.total, 0)  AS item_total,
			COALESCE(iis.cost, 0)   AS item_cost,

			m.sunglass AS is_sunglass,

			EXISTS(SELECT 1 FROM lab_ticket lt WHERE lt.invoice_id = i.id_invoice) AS has_lab_ticket

		FROM invoice i
		JOIN invoice_item_sale iis ON iis.invoice_id = i.id_invoice
		LEFT JOIN patient p        ON p.id_patient   = i.patient_id
		JOIN employee e            ON e.id_employee  = i.employee_id
		LEFT JOIN inventory inv
		       ON LOWER(TRIM(CAST(iis.item_type AS TEXT))) = 'frames'
		      AND inv.id_inventory = iis.item_id
		LEFT JOIN model m          ON m.id_model = inv.model_id

		WHERE i.location_id IN (?)
		  AND i.date_create BETWEEN ? AND ?`

	args := []interface{}{locationIDs, dateStart, dateEnd}

	if employeeID != nil {
		query += ` AND i.employee_id = ?`
		args = append(args, *employeeID)
	}

	if !showInterCompany {
		query += ` AND i.number_invoice NOT ILIKE 'I%'`
	}

	query += ` ORDER BY i.date_create, i.number_invoice`

	rows, err := s.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// aggregate by invoice preserving order
	invMap := make(map[int64]*invoiceAcc)
	var invOrder []*invoiceAcc
	orderIdx := 0

	for rows.Next() {
		var r rawRow
		if err := rows.Scan(
			&r.IDInvoice, &r.NumberInvoice, &r.DateCreate,
			&r.InvDiscount, &r.TaxAmount, &r.EmployeeID,
			&r.PatientFirst, &r.PatientLast,
			&r.EmpFirst, &r.EmpLast,
			&r.IDInvoiceSale, &r.ItemType, &r.ItemID,
			&r.ItemTotal, &r.ItemCost,
			&r.IsSunglass, &r.HasLabTicket,
		); err != nil {
			return nil, err
		}

		acc, exists := invMap[r.IDInvoice]
		if !exists {
			// employee display: "FIRST L"
			empFirst := strings.ToUpper(ptrStr(r.EmpFirst))
			empLast := ptrStr(r.EmpLast)
			empDisplay := empFirst
			if empLast != "" {
				empDisplay = empFirst + " " + strings.ToUpper(empLast[:1])
			}

			// patient name
			patName := strings.TrimSpace(ptrStr(r.PatientFirst) + " " + ptrStr(r.PatientLast))

			acc = &invoiceAcc{
				date:         r.DateCreate,
				invoice:      r.NumberInvoice,
				name:         patName,
				employee:     empDisplay,
				discount:     r.InvDiscount,
				tax:          r.TaxAmount,
				hasLabTicket: r.HasLabTicket,
				zeroCostCats: make(map[string]struct{}),
				order:        orderIdx,
			}
			invMap[r.IDInvoice] = acc
			invOrder = append(invOrder, acc)
			orderIdx++
		}

		itemTotal := r.ItemTotal
		itemCost := r.ItemCost
		itemType := strings.ToLower(strings.TrimSpace(ptrStr(r.ItemType)))

		acc.invTotal += itemTotal
		acc.cost += itemCost

		hasRx := acc.hasLabTicket
		var catLabel string

		switch itemType {
		case "frames":
			isSun := r.IsSunglass != nil && *r.IsSunglass
			if isSun {
				if hasRx {
					acc.sunRx += itemTotal
					catLabel = "Sun Rx"
				} else {
					acc.sunPlano += itemTotal
					catLabel = "Sun Plano"
				}
			} else {
				if hasRx {
					acc.ophFrRx += itemTotal
					catLabel = "Oph Fr Rx"
				} else {
					acc.ophFrOnly += itemTotal
					catLabel = "Oph Fr Only"
				}
			}
		case "lens":
			acc.lOnly += itemTotal
			catLabel = "Lab"
		case "contact lens":
			acc.contacts += itemTotal
			catLabel = "Contacts"
		case "prof. service":
			acc.profFees += itemTotal
			catLabel = "Prof Fees"
		default:
			acc.other += itemTotal
			catLabel = "Other"
		}

		if itemCost == 0 && itemTotal != 0 {
			acc.zeroCostCats[catLabel] = struct{}{}
		}
	}

	// build result
	result := &BreakdownResult{}
	var totals BreakdownTotals

	for _, acc := range invOrder {
		// cost info missing
		var missingParts []string
		for k := range acc.zeroCostCats {
			missingParts = append(missingParts, k)
		}
		sort.Strings(missingParts)
		costInfoMissing := strings.Join(missingParts, ", ")

		dateStr := ""
		if !acc.date.IsZero() {
			dateStr = acc.date.Format("01/02/2006")
		}

		row := BreakdownRow{
			Date:            dateStr,
			Invoice:         acc.invoice,
			Name:            acc.name,
			InvTotal:        round2(acc.invTotal),
			Cost:            round2(acc.cost),
			OphFrOnly:       round2(acc.ophFrOnly),
			OphFrRx:         round2(acc.ophFrRx),
			SunPlano:        round2(acc.sunPlano),
			SunRx:           round2(acc.sunRx),
			LOnly:           round2(acc.lOnly),
			ProfFees:        round2(acc.profFees),
			Contacts:        round2(acc.contacts),
			Other:           round2(acc.other),
			Discount:        round2(acc.discount),
			Tax:             round2(acc.tax),
			Employee:        acc.employee,
			CostInfoMissing: costInfoMissing,
		}
		result.Data = append(result.Data, row)

		totals.InvTotal += acc.invTotal
		totals.Cost += acc.cost
		totals.OphFrOnly += acc.ophFrOnly
		totals.OphFrRx += acc.ophFrRx
		totals.SunPlano += acc.sunPlano
		totals.SunRx += acc.sunRx
		totals.LOnly += acc.lOnly
		totals.ProfFees += acc.profFees
		totals.Contacts += acc.contacts
		totals.Other += acc.other
		totals.Discount += acc.discount
		totals.Tax += acc.tax
	}

	result.Totals = totals
	result.RecordCount = len(result.Data)

	// percent net of tax
	catSum := totals.OphFrOnly + totals.OphFrRx + totals.SunPlano + totals.SunRx +
		totals.LOnly + totals.ProfFees + totals.Contacts + totals.Other

	if catSum > 0 {
		result.PercentNetOfTax = BreakdownPercent{
			OphFrOnly: math.Round(totals.OphFrOnly/catSum*1000) / 10,
			OphFrRx:   math.Round(totals.OphFrRx/catSum*1000) / 10,
			SunPlano:  math.Round(totals.SunPlano/catSum*1000) / 10,
			SunRx:     math.Round(totals.SunRx/catSum*1000) / 10,
			LOnly:     math.Round(totals.LOnly/catSum*1000) / 10,
			ProfFees:  math.Round(totals.ProfFees/catSum*1000) / 10,
			Contacts:  math.Round(totals.Contacts/catSum*1000) / 10,
			Other:     math.Round(totals.Other/catSum*1000) / 10,
		}
	}

	return result, nil
}

func ptrStr(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
