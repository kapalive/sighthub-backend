package report_inventory_service

import (
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

func round2(v float64) float64 { return math.Round(v*100) / 100 }

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
	}
	return 0
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ─── Types ──────────────────────────────────────────────────────────────────────

type FrameInteractionItem struct {
	Serial        int64   `json:"serial"`
	ItemNbr       string  `json:"item_nbr"`
	ModelID       int64   `json:"model_id"`
	Description   string  `json:"description"`
	Status        string  `json:"status"`
	WorkFlow      string  `json:"work_flow"`
	BeginDate     *string `json:"begin_date"`
	EndDate       *string `json:"end_date"`
	Stock         string  `json:"stock"`
	EmployeeLogin string  `json:"employee_login"`
	EmployeeName  string  `json:"employee_name"`
	Cost          float64 `json:"cost"`
}

type MissingItem struct {
	Store       string  `json:"store"`
	Vendor      string  `json:"vendor"`
	Brand       string  `json:"brand"`
	Model       string  `json:"model"`
	Cost        float64 `json:"cost"`
	FNumber     string  `json:"f_number"`
	Serial      int64   `json:"serial"`
	InStockDate *string `json:"in_stock_date"`
}

type ReceiptByBrandItem struct {
	Vendor string  `json:"vendor"`
	Brand  string  `json:"brand"`
	Qty    int     `json:"qty"`
	Cost   float64 `json:"cost"`
	Price  float64 `json:"price"`
}

type ReceiptItem struct {
	Date             *string  `json:"date"`
	ReceiptNo        string   `json:"receipt_no"`
	Vendor           string   `json:"vendor"`
	Location         string   `json:"location"`
	InvoiceDate      *string  `json:"invoice_date"`
	PackInv          *string  `json:"pack_inv"`
	Qty              int      `json:"qty"`
	SubTotal         float64  `json:"sub_total"`
	ShippingHandling float64  `json:"shipping_handling"`
	Tax              float64  `json:"tax"`
	Total            float64  `json:"total"`
	OrderRef         string   `json:"order_ref"`
}

type ReceiptTotals struct {
	Qty              int     `json:"qty"`
	SubTotal         float64 `json:"sub_total"`
	ShippingHandling float64 `json:"shipping_handling"`
	Tax              float64 `json:"tax"`
	Total            float64 `json:"total"`
}

type TransferItem struct {
	Date            *string `json:"date"`
	Serial          int64   `json:"serial"`
	FromLocation    string  `json:"from_location"`
	ToLocation      string  `json:"to_location"`
	Brand           string  `json:"brand"`
	Model           string  `json:"model"`
	TransactionType string  `json:"transaction_type"`
	Cost            float64 `json:"cost"`
	Price           float64 `json:"price"`
}

type LocationItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ─── Location group helpers ─────────────────────────────────────────────────────

func (s *Service) resolveLocationGroup(locationID int, allowedIDs []int) []int {
	type locRow struct {
		IDLocation      int
		WarehouseID     *int
		CanReceiveItems *bool
	}
	var loc locRow
	if err := s.db.Table("location").
		Select("id_location, warehouse_id, can_receive_items").
		Where("id_location = ?", locationID).
		Scan(&loc).Error; err != nil || loc.IDLocation == 0 {
		return nil
	}

	if loc.WarehouseID != nil {
		var ids []int
		s.db.Table("location").
			Where("warehouse_id = ? AND can_receive_items = true AND id_location IN ?",
				*loc.WarehouseID, allowedIDs).
			Pluck("id_location", &ids)
		return ids
	}

	if loc.CanReceiveItems != nil && *loc.CanReceiveItems {
		for _, id := range allowedIDs {
			if id == locationID {
				return []int{locationID}
			}
		}
	}
	return nil
}

func (s *Service) resolveLocationGroupAll(locationID int, allowedIDs []int) []int {
	type locRow struct {
		IDLocation  int
		WarehouseID *int
	}
	var loc locRow
	if err := s.db.Table("location").
		Select("id_location, warehouse_id").
		Where("id_location = ?", locationID).
		Scan(&loc).Error; err != nil || loc.IDLocation == 0 {
		return nil
	}

	if loc.WarehouseID != nil {
		var ids []int
		s.db.Table("location").
			Where("warehouse_id = ? AND id_location IN ?", *loc.WarehouseID, allowedIDs).
			Pluck("id_location", &ids)
		return ids
	}

	for _, id := range allowedIDs {
		if id == locationID {
			return []int{locationID}
		}
	}
	return nil
}

// ResolveReceiptLocations returns location IDs that can_receive_items, grouped by warehouse.
func (s *Service) ResolveReceiptLocations(locationIDStr string, allowedIDs []int, empLocationID *int64) ([]int, string) {
	if len(allowedIDs) == 0 {
		return nil, "No permitted locations"
	}

	if locationIDStr == "" {
		if empLocationID != nil {
			lid := int(*empLocationID)
			for _, id := range allowedIDs {
				if id == lid {
					ids := s.resolveLocationGroup(lid, allowedIDs)
					if len(ids) > 0 {
						return ids, ""
					}
					return []int{lid}, ""
				}
			}
		}
		var ids []int
		s.db.Table("location").
			Where("id_location IN ? AND can_receive_items = true", allowedIDs).
			Pluck("id_location", &ids)
		return ids, ""
	}

	if locationIDStr == "all" {
		var ids []int
		s.db.Table("location").
			Where("id_location IN ? AND can_receive_items = true", allowedIDs).
			Pluck("id_location", &ids)
		return ids, ""
	}

	var lid int
	if _, err := fmt.Sscanf(locationIDStr, "%d", &lid); err != nil {
		return nil, "Invalid location_id"
	}

	found := false
	for _, id := range allowedIDs {
		if id == lid {
			found = true
			break
		}
	}
	if !found {
		return nil, "Permission denied for this location"
	}

	ids := s.resolveLocationGroup(lid, allowedIDs)
	if len(ids) == 0 {
		return nil, "Location cannot receive items"
	}
	return ids, ""
}

// ResolveAllLocations returns location IDs without can_receive_items restriction.
func (s *Service) ResolveAllLocations(locationIDStr string, allowedIDs []int, empLocationID *int64) ([]int, string) {
	if len(allowedIDs) == 0 {
		return nil, "No permitted locations"
	}

	if locationIDStr == "" {
		if empLocationID != nil {
			lid := int(*empLocationID)
			for _, id := range allowedIDs {
				if id == lid {
					ids := s.resolveLocationGroupAll(lid, allowedIDs)
					if len(ids) > 0 {
						return ids, ""
					}
					return []int{lid}, ""
				}
			}
		}
		return allowedIDs, ""
	}

	if locationIDStr == "all" {
		return allowedIDs, ""
	}

	var lid int
	if _, err := fmt.Sscanf(locationIDStr, "%d", &lid); err != nil {
		return nil, "Invalid location_id"
	}

	found := false
	for _, id := range allowedIDs {
		if id == lid {
			found = true
			break
		}
	}
	if !found {
		return nil, "Permission denied for this location"
	}

	ids := s.resolveLocationGroupAll(lid, allowedIDs)
	if len(ids) > 0 {
		return ids, ""
	}
	return []int{lid}, ""
}

// ─── 1. Frame Interaction ───────────────────────────────────────────────────────

type frameRow struct {
	IDTransaction     int64
	Serial            int64
	IDModel           int64
	ModelID           *int64
	BrandName         string
	TitleProduct      string
	TitleVariant      string
	StatusItems       string
	TransactionType   string
	DateTransaction   time.Time
	EmployeeLogin     string
	EmployeeFirstName *string
	EmployeeLastName  *string
	ItemListCost      *float64
	ItemDiscount      *float64
}

func (s *Service) FrameInteraction(
	locationIDs []int,
	dateStart, dateEnd time.Time,
	vendorIDs, brandIDs []int64,
	statuses, vendorNames, brandNames []string,
) ([]FrameInteractionItem, float64, error) {

	q := `SELECT
		it.id_transaction,
		i.id_inventory AS serial,
		m.id_model,
		i.model_id,
		b.brand_name,
		p.title_product,
		m.title_variant,
		it.status_items,
		it.transaction_type,
		it.date_transaction,
		el.employee_login,
		e.first_name AS employee_first_name,
		e.last_name  AS employee_last_name,
		pb.item_list_cost,
		pb.item_discount
	FROM inventory_transaction it
	JOIN inventory i   ON it.inventory_id = i.id_inventory
	JOIN model m       ON i.model_id = m.id_model
	JOIN product p     ON m.product_id = p.id_product
	JOIN brand b       ON p.brand_id = b.id_brand
	JOIN vendor v      ON p.vendor_id = v.id_vendor
	JOIN employee e    ON it.transferred_by = e.id_employee
	JOIN employee_login el ON e.employee_login_id = el.id_employee_login
	LEFT JOIN price_book pb ON pb.inventory_id = i.id_inventory
	WHERE it.date_transaction BETWEEN ? AND ?
	  AND (it.from_location_id IN ? OR it.to_location_id IN ?)`

	args := []interface{}{dateStart, dateEnd, locationIDs, locationIDs}

	if len(vendorIDs) > 0 {
		q += " AND v.id_vendor IN ?"
		args = append(args, vendorIDs)
	}
	if len(brandIDs) > 0 {
		q += " AND b.id_brand IN ?"
		args = append(args, brandIDs)
	}
	if len(statuses) > 0 {
		q += " AND it.status_items IN ?"
		args = append(args, statuses)
	}
	if len(vendorNames) > 0 {
		q += " AND v.vendor_name IN ?"
		args = append(args, vendorNames)
	}
	if len(brandNames) > 0 {
		q += " AND b.brand_name IN ?"
		args = append(args, brandNames)
	}

	q += " ORDER BY it.date_transaction DESC"

	var rows []frameRow
	if err := s.db.Raw(q, args...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	// Compute end_date per inventory_id group
	invTxnIDs := map[int64][]int{}
	for i, r := range rows {
		invTxnIDs[r.Serial] = append(invTxnIDs[r.Serial], i)
	}

	defaultEnd := time.Date(2210, 1, 1, 12, 0, 0, 0, time.UTC)
	endDateMap := make(map[int64]time.Time)

	for _, indices := range invTxnIDs {
		for i, idx := range indices {
			txnID := rows[idx].IDTransaction
			if i == 0 {
				endDateMap[txnID] = defaultEnd
			} else {
				endDateMap[txnID] = rows[indices[i-1]].DateTransaction
			}
		}
	}

	var totalCost float64
	items := make([]FrameInteractionItem, 0, len(rows))

	for _, r := range rows {
		listCost := toFloat(r.ItemListCost)
		discount := toFloat(r.ItemDiscount)
		cost := round2(listCost - discount)
		totalCost += cost

		itemNbr := fmt.Sprintf("F%dF", r.IDModel)
		desc := r.BrandName + " " + r.TitleProduct + " " + r.TitleVariant
		stock := "N"
		if r.StatusItems == "Ready for Sale" {
			stock = "Y"
		}
		empName := (ptrStr(r.EmployeeFirstName) + " " + ptrStr(r.EmployeeLastName))

		var beginStr, endStr *string
		bd := r.DateTransaction.Format("01/02/2006 03:04:05 PM")
		beginStr = &bd
		if ed, ok := endDateMap[r.IDTransaction]; ok {
			s := ed.Format("01/02/2006 03:04:05 PM")
			endStr = &s
		}

		items = append(items, FrameInteractionItem{
			Serial:        r.Serial,
			ItemNbr:       itemNbr,
			ModelID:       r.IDModel,
			Description:   desc,
			Status:        r.StatusItems,
			WorkFlow:      r.TransactionType,
			BeginDate:     beginStr,
			EndDate:       endStr,
			Stock:         stock,
			EmployeeLogin: r.EmployeeLogin,
			EmployeeName:  empName,
			Cost:          cost,
		})
	}

	return items, round2(totalCost), nil
}

// ─── 2. Missing Inventory ───────────────────────────────────────────────────────

func (s *Service) MissingInventory(locationIDs []int, startDate, endDate time.Time) ([]MissingItem, float64, error) {
	q := `SELECT
		l.full_name  AS store,
		v.vendor_name AS vendor,
		b.brand_name  AS brand,
		p.title_product,
		m.title_variant,
		mi.cost,
		m.id_model,
		i.id_inventory AS serial,
		mi.reported_date
	FROM missing mi
	JOIN inventory i ON mi.inventory_id = i.id_inventory
	JOIN model m     ON mi.model_id = m.id_model
	JOIN product p   ON m.product_id = p.id_product
	JOIN brand b     ON mi.brand_id = b.id_brand
	JOIN vendor v    ON p.vendor_id = v.id_vendor
	JOIN location l  ON mi.location_id = l.id_location
	WHERE mi.location_id IN ?
	  AND mi.reported_date BETWEEN ? AND ?
	ORDER BY l.full_name, v.vendor_name, b.brand_name`

	type row struct {
		Store        string
		Vendor       string
		Brand        string
		TitleProduct string
		TitleVariant string
		Cost         *float64
		IDModel      int64
		Serial       int64
		ReportedDate *time.Time
	}
	var rows []row
	if err := s.db.Raw(q, locationIDs, startDate, endDate).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	var totalCost float64
	items := make([]MissingItem, 0, len(rows))
	for _, r := range rows {
		cost := toFloat(r.Cost)
		totalCost += cost
		model := (r.TitleProduct + " " + r.TitleVariant)
		fNumber := fmt.Sprintf("F%dF", r.IDModel)

		var dateStr *string
		if r.ReportedDate != nil {
			s := r.ReportedDate.Format("01/02/2006 03:04:05 PM")
			dateStr = &s
		}

		items = append(items, MissingItem{
			Store:       r.Store,
			Vendor:      r.Vendor,
			Brand:       r.Brand,
			Model:       model,
			Cost:        round2(cost),
			FNumber:     fNumber,
			Serial:      r.Serial,
			InStockDate: dateStr,
		})
	}

	return items, round2(totalCost), nil
}

// ─── 3. Receipt By Brand ────────────────────────────────────────────────────────

func (s *Service) ReceiptByBrand(locationIDs []int, startDate, endDate time.Time) ([]ReceiptByBrandItem, int, float64, float64, error) {
	q := `SELECT
		v.vendor_name AS vendor,
		b.brand_name  AS brand,
		COUNT(i.id_inventory) AS qty,
		COALESCE(SUM(pb.item_net), 0) AS cost,
		COALESCE(SUM(pb.pb_selling_price), 0) AS price
	FROM inventory i
	JOIN invoice inv ON i.invoice_id = inv.id_invoice
	JOIN model m     ON i.model_id = m.id_model
	JOIN product p   ON m.product_id = p.id_product
	JOIN brand b     ON p.brand_id = b.id_brand
	JOIN vendor v    ON p.vendor_id = v.id_vendor
	LEFT JOIN price_book pb ON pb.inventory_id = i.id_inventory
	WHERE inv.number_invoice LIKE 'V%'
	  AND inv.location_id IN ?
	  AND inv.created_at BETWEEN ? AND ?
	GROUP BY v.vendor_name, b.brand_name
	ORDER BY v.vendor_name, b.brand_name`

	type row struct {
		Vendor string
		Brand  string
		Qty    int
		Cost   float64
		Price  float64
	}
	var rows []row
	if err := s.db.Raw(q, locationIDs, startDate, endDate).Scan(&rows).Error; err != nil {
		return nil, 0, 0, 0, err
	}

	var totalQty int
	var totalCost, totalPrice float64
	items := make([]ReceiptByBrandItem, 0, len(rows))
	for _, r := range rows {
		totalQty += r.Qty
		totalCost += r.Cost
		totalPrice += r.Price
		items = append(items, ReceiptByBrandItem{
			Vendor: r.Vendor,
			Brand:  r.Brand,
			Qty:    r.Qty,
			Cost:   round2(r.Cost),
			Price:  round2(r.Price),
		})
	}

	return items, totalQty, round2(totalCost), round2(totalPrice), nil
}

// ─── 4. List of Receipts ────────────────────────────────────────────────────────

func (s *Service) ListOfReceipts(locationIDs []int, startDate, endDate time.Time) ([]ReceiptItem, ReceiptTotals, error) {
	q := `SELECT
		inv.created_at     AS date,
		inv.number_invoice AS receipt_no,
		v.vendor_name      AS vendor,
		rl.full_name       AS location,
		vi.invoice_date,
		vi.quantity         AS qty,
		vi.sub_total,
		vi.shipping_handling,
		vi.tax,
		vi.invoice_total    AS total,
		vi.order_ref
	FROM vendor_invoice vi
	JOIN invoice inv  ON vi.invoice_id = inv.id_invoice
	JOIN vendor v     ON vi.vendor_id = v.id_vendor
	JOIN location rl  ON inv.location_id = rl.id_location
	WHERE inv.location_id IN ?
	  AND inv.created_at BETWEEN ? AND ?
	ORDER BY inv.created_at DESC`

	type row struct {
		Date             *time.Time
		ReceiptNo        string
		Vendor           string
		Location         string
		InvoiceDate      *time.Time
		Qty              int
		SubTotal         *float64
		ShippingHandling *float64
		Tax              *float64
		Total            *float64
		OrderRef         string
	}
	var rows []row
	if err := s.db.Raw(q, locationIDs, startDate, endDate).Scan(&rows).Error; err != nil {
		return nil, ReceiptTotals{}, err
	}

	var totals ReceiptTotals
	items := make([]ReceiptItem, 0, len(rows))

	for _, r := range rows {
		sub := toFloat(r.SubTotal)
		sh := toFloat(r.ShippingHandling)
		tax := toFloat(r.Tax)
		tot := toFloat(r.Total)

		totals.Qty += r.Qty
		totals.SubTotal += sub
		totals.ShippingHandling += sh
		totals.Tax += tax
		totals.Total += tot

		var dateStr, invDateStr *string
		if r.Date != nil {
			s := r.Date.Format("01/02/2006")
			dateStr = &s
		}
		if r.InvoiceDate != nil {
			s := r.InvoiceDate.Format("01/02/2006")
			invDateStr = &s
		}

		items = append(items, ReceiptItem{
			Date:             dateStr,
			ReceiptNo:        r.ReceiptNo,
			Vendor:           r.Vendor,
			Location:         r.Location,
			InvoiceDate:      invDateStr,
			PackInv:          nil,
			Qty:              r.Qty,
			SubTotal:         round2(sub),
			ShippingHandling: round2(sh),
			Tax:              round2(tax),
			Total:            round2(tot),
			OrderRef:         r.OrderRef,
		})
	}

	totals.SubTotal = round2(totals.SubTotal)
	totals.ShippingHandling = round2(totals.ShippingHandling)
	totals.Tax = round2(totals.Tax)
	totals.Total = round2(totals.Total)

	return items, totals, nil
}

// ─── 5. Internal Transfers ──────────────────────────────────────────────────────

func (s *Service) InternalTransfers(locationIDs []int, startDate, endDate time.Time) ([]TransferItem, float64, float64, error) {
	q := `SELECT
		it.date_transaction AS date,
		i.id_inventory      AS serial,
		fl.full_name        AS from_location,
		tl.full_name        AS to_location,
		b.brand_name        AS brand,
		p.title_product,
		m.title_variant,
		it.transaction_type,
		pb.item_net         AS cost,
		pb.pb_selling_price AS price
	FROM inventory_transaction it
	JOIN inventory i     ON it.inventory_id = i.id_inventory
	JOIN model m         ON i.model_id = m.id_model
	JOIN product p       ON m.product_id = p.id_product
	JOIN brand b         ON p.brand_id = b.id_brand
	LEFT JOIN location fl ON it.from_location_id = fl.id_location
	LEFT JOIN location tl ON it.to_location_id = tl.id_location
	LEFT JOIN price_book pb ON pb.inventory_id = i.id_inventory
	WHERE it.from_location_id IN ?
	  AND it.to_location_id IN ?
	  AND it.date_transaction BETWEEN ? AND ?
	ORDER BY it.date_transaction DESC`

	type row struct {
		Date            *time.Time
		Serial          int64
		FromLocation    string
		ToLocation      string
		Brand           string
		TitleProduct    string
		TitleVariant    string
		TransactionType string
		Cost            *float64
		Price           *float64
	}
	var rows []row
	if err := s.db.Raw(q, locationIDs, locationIDs, startDate, endDate).Scan(&rows).Error; err != nil {
		return nil, 0, 0, err
	}

	var totalCost, totalPrice float64
	items := make([]TransferItem, 0, len(rows))

	for _, r := range rows {
		cost := toFloat(r.Cost)
		price := toFloat(r.Price)
		totalCost += cost
		totalPrice += price

		modelDesc := (r.TitleProduct + " " + r.TitleVariant)

		var dateStr *string
		if r.Date != nil {
			s := r.Date.Format("01/02/2006 03:04:05 PM")
			dateStr = &s
		}

		items = append(items, TransferItem{
			Date:            dateStr,
			Serial:          r.Serial,
			FromLocation:    r.FromLocation,
			ToLocation:      r.ToLocation,
			Brand:           r.Brand,
			Model:           modelDesc,
			TransactionType: r.TransactionType,
			Cost:            round2(cost),
			Price:           round2(price),
		})
	}

	return items, round2(totalCost), round2(totalPrice), nil
}

// ─── 6. Can-Receive Locations ───────────────────────────────────────────────────

func (s *Service) CanReceiveLocations(allowedIDs []int) ([]LocationItem, error) {
	q := `SELECT id_location AS id, full_name AS name
	FROM location
	WHERE id_location IN ?
	  AND can_receive_items = true
	  AND store_active = true
	ORDER BY full_name`

	var items []LocationItem
	if err := s.db.Raw(q, allowedIDs).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// ─── 7. All Locations ───────────────────────────────────────────────────────────

func (s *Service) AllLocations(allowedIDs []int) ([]LocationItem, error) {
	q := `SELECT id_location AS id, full_name AS name
	FROM location
	WHERE id_location IN ?
	  AND store_active = true
	ORDER BY full_name`

	var items []LocationItem
	if err := s.db.Raw(q, allowedIDs).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
