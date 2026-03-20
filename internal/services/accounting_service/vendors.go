package accounting_service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"sighthub-backend/internal/models/general"
	locModel "sighthub-backend/internal/models/location"
	invModel "sighthub-backend/internal/models/inventory"
	vendorModel "sighthub-backend/internal/models/vendors"
	pkgAccounting "sighthub-backend/pkg/accounting"
	pkgActivity "sighthub-backend/pkg/activitylog"
)

// ── GET /vendors ─────────────────────────────────────────────────────────────

func (s *Service) GetVendors() ([]map[string]interface{}, error) {
	var vendors []vendorModel.Vendor
	if err := s.db.Find(&vendors).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(vendors))
	for i, v := range vendors {
		out[i] = v.ToMap()
	}
	return out, nil
}

// ── GET /:vendor_id/quickbooks-header ────────────────────────────────────────

func (s *Service) GetVendorQuickbooksHeader(username string, vendorID int) (map[string]interface{}, error) {
	_, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	var vendor vendorModel.Vendor
	if err := s.db.First(&vendor, vendorID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("vendor not found")
		}
		return nil, err
	}

	var noteParts []string
	if vendor.RegionalManager != nil && *vendor.RegionalManager != "" {
		noteParts = append(noteParts, fmt.Sprintf("Regional manager: %s", *vendor.RegionalManager))
	}
	if vendor.RegionalMNo != nil && *vendor.RegionalMNo != "" {
		noteParts = append(noteParts, fmt.Sprintf("Phone: %s", *vendor.RegionalMNo))
	}
	var note *string
	if len(noteParts) > 0 {
		s2 := strings.Join(noteParts, " | ")
		note = &s2
	}

	billedFrom := vendor.VendorName
	if vendor.ShortName != nil && *vendor.ShortName != "" {
		billedFrom = *vendor.ShortName
	}

	result := map[string]interface{}{
		"company_name": vendor.VendorName,
		"full_name":    strVal(vendor.RegionalManager),
		"billed_from":  billedFrom,
		"address":      vendor.StreetAddress,
		"address_2":    vendor.AddressLine2,
		"city":         vendor.City,
		"state":        vendor.State,
		"zip":          vendor.ZipCode,
		"country":      vendor.Country,
		"phone":        vendor.Phone,
		"email":        vendor.Email,
		"website":      vendor.Website,
		"notes":        note,
	}
	return result, nil
}

func strVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ── GET /payments_methods ────────────────────────────────────────────────────

func (s *Service) GetPaymentMethods() ([]map[string]interface{}, error) {
	var methods []general.PaymentMethod
	if err := s.db.Find(&methods).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(methods))
	for i, m := range methods {
		out[i] = m.ToMap()
	}
	return out, nil
}

// ── GET /stores ──────────────────────────────────────────────────────────────

func (s *Service) GetStores() ([]map[string]interface{}, error) {
	var stores []locModel.Store
	if err := s.db.Find(&stores).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(stores))
	for i, st := range stores {
		out[i] = st.ToMap()
	}
	return out, nil
}

// ── GET /locations ───────────────────────────────────────────────────────────

func (s *Service) GetLocations() ([]map[string]interface{}, error) {
	var locs []locModel.Location
	if err := s.db.Find(&locs).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(locs))
	for i, l := range locs {
		out[i] = l.ToMap()
	}
	return out, nil
}

// ── GET /vendor-invoices/:vendor_id ──────────────────────────────────────────

func (s *Service) GetVendorInvoicesList(username string, vendorID, page, perPage int) (map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	accLoc, err := s.resolveAccountingLocation(loc)
	if err != nil {
		return nil, err
	}

	// Collect location IDs for this store group (showcase + warehouse)
	storeLocIDs := []int64{int64(accLoc.IDLocation)}
	if accLoc.WarehouseID != nil {
		storeLocIDs = append(storeLocIDs, int64(*accLoc.WarehouseID))
	}
	// Also find all locations with same store_id
	if accLoc.StoreID != 0 {
		var siblings []locModel.Location
		s.db.Where("store_id = ?", accLoc.StoreID).Find(&siblings)
		for _, sib := range siblings {
			found := false
			for _, id := range storeLocIDs {
				if id == int64(sib.IDLocation) {
					found = true
					break
				}
			}
			if !found {
				storeLocIDs = append(storeLocIDs, int64(sib.IDLocation))
			}
		}
	}

	// VendorInvoice → Invoice.location_id
	baseQuery := s.db.Model(&invModel.VendorInvoice{}).
		Joins("JOIN invoice ON invoice.id_invoice = vendor_invoice.invoice_id").
		Where("vendor_invoice.vendor_id = ? AND invoice.location_id IN ?", vendorID, storeLocIDs)

	var total int64
	baseQuery.Count(&total)

	var rows []invModel.VendorInvoice
	offset := (page - 1) * perPage
	if err := s.db.
		Joins("JOIN invoice ON invoice.id_invoice = vendor_invoice.invoice_id").
		Where("vendor_invoice.vendor_id = ? AND invoice.location_id IN ?", vendorID, storeLocIDs).
		Order("vendor_invoice.invoice_date DESC").
		Offset(offset).Limit(perPage).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		var invDateStr interface{}
		if r.InvoiceDate != nil {
			invDateStr = r.InvoiceDate.Format("2006-01-02")
		}
		var totalStr interface{}
		if r.InvoiceTotal != nil {
			totalStr = fmt.Sprintf("%.2f", *r.InvoiceTotal)
		}
		items[i] = map[string]interface{}{
			"vendor_id":      vendorID,
			"invoice_number": r.InvoiceNo,
			"invoice_date":   invDateStr,
			"invoice_amount": totalStr,
		}
	}

	pages := int((total + int64(perPage) - 1) / int64(perPage))
	return map[string]interface{}{
		"items":    items,
		"total":    total,
		"page":     page,
		"per_page": perPage,
		"pages":    pages,
	}, nil
}

// ── GET /invoices/:vendor_id ─────────────────────────────────────────────────

func (s *Service) GetInvoicesByVendor(username string, vendorID int) (map[string]interface{}, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	_ = emp
	if err != nil {
		return nil, err
	}
	accLoc, err := s.resolveAccountingLocation(loc)
	if err != nil {
		return nil, err
	}
	locID := int64(accLoc.IDLocation)

	var invoices []vendorModel.VendorAPInvoice
	if err := s.db.
		Where("vendor_id = ? AND location_id = ?", vendorID, locID).
		Order("invoice_date DESC, id_vendor_ap_invoice DESC").
		Find(&invoices).Error; err != nil {
		return nil, err
	}

	// load items per invoice
	invoiceIDs := make([]int64, len(invoices))
	for i, inv := range invoices {
		invoiceIDs[i] = inv.IDVendorAPInvoice
	}
	itemsByInv := make(map[int64][]map[string]interface{})
	if len(invoiceIDs) > 0 {
		var apItems []vendorModel.VendorAPInvoiceItem
		s.db.Where("vendor_ap_invoice_id IN ?", invoiceIDs).Find(&apItems)
		for _, it := range apItems {
			itemsByInv[it.VendorAPInvoiceID] = append(itemsByInv[it.VendorAPInvoiceID], map[string]interface{}{
				"id_vendor_ap_invoice_item": it.IDVendorAPInvoiceItem,
				"vendor_ap_invoice_id":      it.VendorAPInvoiceID,
				"line_no":                   it.LineNo,
				"quantity":                  it.Quantity,
				"description":               it.Description,
				"price_each":                it.PriceEach,
				"amount":                    it.Amount,
				"tax":                       it.Tax,
			})
		}
	}

	// load VendorLocationAccount IDs needed
	accIDs := make(map[int64]struct{})
	for _, inv := range invoices {
		if inv.VendorLocationAccountID != nil {
			accIDs[*inv.VendorLocationAccountID] = struct{}{}
		}
	}
	accMap := make(map[int64]string)
	if len(accIDs) > 0 {
		ids := make([]int64, 0, len(accIDs))
		for id := range accIDs {
			ids = append(ids, id)
		}
		var accs []vendorModel.VendorLocationAccount
		s.db.Where("id_vendor_location_account IN ?", ids).Find(&accs)
		for _, acc := range accs {
			accMap[acc.IDVendorLocationAccount] = acc.AccountNumber
		}
	}

	items := make([]map[string]interface{}, len(invoices))
	for i, inv := range invoices {
		var accNum interface{}
		if inv.VendorLocationAccountID != nil {
			if num, ok := accMap[*inv.VendorLocationAccountID]; ok {
				accNum = num
			}
		}
		invItems := itemsByInv[inv.IDVendorAPInvoice]
		if invItems == nil {
			invItems = []map[string]interface{}{}
		}
		items[i] = map[string]interface{}{
			"id_vendor_ap_invoice":      inv.IDVendorAPInvoice,
			"vendor_id":                 inv.VendorID,
			"location_id":               inv.LocationID,
			"vendor_location_account_id": inv.VendorLocationAccountID,
			"account_number":            accNum,
			"terms":                     inv.Terms,
			"employee_id":               inv.EmployeeID,
			"invoice_number":            inv.InvoiceNumber,
			"invoice_date":              fmtDate(inv.InvoiceDate),
			"bill_due_date":             fmtDate(inv.BillDueDate),
			"invoice_amount":            inv.InvoiceAmount,
			"open_balance":              inv.OpenBalance,
			"tax_total":                 inv.TaxTotal,
			"status":                    inv.Status,
			"attachment_url":            inv.AttachmentURL,
			"note":                      inv.Note,
			"items":                     invItems,
		}
	}

	return map[string]interface{}{
		"vendor_id":   vendorID,
		"location_id": locID,
		"items":       items,
	}, nil
}

func fmtDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func fmtDatePtr(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.Format("2006-01-02")
}

func fmtTimePtr(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.Format(time.RFC3339)
}

// ── POST /vendor-bills ────────────────────────────────────────────────────────

type CreateVendorBillInput struct {
	VendorID        *int                     `json:"vendor_id"`
	InvoiceNumber   *string                  `json:"invoice_number"`
	InvoiceDate     *string                  `json:"invoice_date"`
	BillDueDate     *string                  `json:"bill_due_date"`
	AccountNumber   *string                  `json:"account_number"`
	InvoiceItems    []VendorBillItemInput    `json:"invoice_items"`
	AttachmentURL   *string                  `json:"link_to_invoice"`
	Note            *string                  `json:"note"`
	Terms           interface{}              `json:"terms"`
	InvoiceAmount   interface{}              `json:"invoice_amount"`
	// for ensure_vendor_account
	VendorLocationAccountID interface{}      `json:"vendor_location_account_id"`
	Status                  *string          `json:"status"`
	QbVendorRef             *string          `json:"qb_vendor_ref"`
}

type VendorBillItemInput struct {
	Quantity    interface{} `json:"quantity"`
	PriceEach   interface{} `json:"price_each"`
	Amount      interface{} `json:"amount"`
	Tax         interface{} `json:"tax"`
	Description string      `json:"description"`
}

func (s *Service) CreateVendorBill(username string, data map[string]interface{}) (map[string]interface{}, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	vendorID, _ := toInt(data["vendor_id"])
	invoiceNumber, _ := data["invoice_number"].(string)
	invoiceDateS, _ := data["invoice_date"].(string)
	billDueDateS, _ := data["bill_due_date"].(string)

	accNum := ""
	if v, ok := data["account_number"]; ok && v != nil {
		accNum = strings.TrimSpace(v.(string))
	}

	if vendorID == 0 || invoiceNumber == "" || invoiceDateS == "" || billDueDateS == "" || accNum == "" {
		return nil, errors.New("vendor_id, account_number, invoice_number, invoice_date, bill_due_date are required")
	}

	var vendor vendorModel.Vendor
	if err := s.db.First(&vendor, vendorID).Error; err != nil {
		return nil, errors.New("vendor not found")
	}

	accLoc, err := s.resolveAccountingLocation(loc)
	if err != nil {
		return nil, err
	}
	locID := int64(accLoc.IDLocation)

	vla, err := s.ensureVendorAccountForLocation(vendorID, locID, data)
	if err != nil {
		return nil, err
	}

	invoiceDate, err := time.Parse("2006-01-02", invoiceDateS)
	if err != nil {
		return nil, errors.New("invalid invoice_date")
	}
	billDueDate, err := time.Parse("2006-01-02", billDueDateS)
	if err != nil {
		return nil, errors.New("invalid bill_due_date")
	}

	var termsVal *int
	if v, ok := data["terms"]; ok && v != nil {
		n, err := toInt(v)
		if err != nil || n < 0 {
			return nil, errors.New("terms must be >= 0 integer")
		}
		termsVal = &n
	}

	var bodyInvoiceAmount *decimal.Decimal
	if v, ok := data["invoice_amount"]; ok && v != nil {
		d := toDecimal(v)
		bodyInvoiceAmount = &d
	}

	var invoiceItems []map[string]interface{}
	if raw, ok := data["invoice_items"]; ok && raw != nil {
		if arr, ok := raw.([]interface{}); ok {
			for _, it := range arr {
				if m, ok := it.(map[string]interface{}); ok {
					invoiceItems = append(invoiceItems, m)
				}
			}
		}
	}

	var attachURL *string
	if v, ok := data["link_to_invoice"]; ok && v != nil {
		if s2, ok := v.(string); ok && s2 != "" {
			attachURL = &s2
		}
	}
	var note *string
	if v, ok := data["note"]; ok && v != nil {
		if s2, ok := v.(string); ok && s2 != "" {
			note = &s2
		}
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	bill := vendorModel.VendorAPInvoice{
		VendorID:                vendorID,
		LocationID:              locID,
		EmployeeID:              int64(emp.IDEmployee),
		VendorLocationAccountID: &vla.IDVendorLocationAccount,
		InvoiceNumber:           invoiceNumber,
		InvoiceDate:             invoiceDate,
		BillDueDate:             billDueDate,
		InvoiceAmount:           "0.00",
		OpenBalance:             "0.00",
		TaxTotal:                "0.00",
		Status:                  "Open",
		AttachmentURL:           attachURL,
		Note:                    note,
		Terms:                   termsVal,
	}
	if err := tx.Create(&bill).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	totalAmount := decimal.Zero
	totalTax := decimal.Zero

	for idx, item := range invoiceItems {
		qty := toDecimal(item["quantity"])
		if qty.IsZero() {
			qty = decimal.NewFromInt(1)
		}
		priceEach := toDecimal(item["price_each"])

		var lineAmount decimal.Decimal
		if item["amount"] != nil {
			lineAmount = toDecimal(item["amount"])
		} else {
			lineAmount = qty.Mul(priceEach)
		}
		tax := toDecimal(item["tax"])

		totalAmount = totalAmount.Add(lineAmount)
		totalTax = totalTax.Add(tax)

		desc := ""
		if v, ok := item["description"].(string); ok {
			desc = v
		}

		billItem := vendorModel.VendorAPInvoiceItem{
			VendorAPInvoiceID: bill.IDVendorAPInvoice,
			LineNo:            idx + 1,
			Quantity:          qty.StringFixed(2),
			Description:       desc,
			PriceEach:         priceEach.StringFixed(2),
			Amount:            lineAmount.StringFixed(2),
			Tax:               tax.StringFixed(2),
		}
		if err := tx.Create(&billItem).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	finalAmount := totalAmount.Add(totalTax)
	if bodyInvoiceAmount != nil {
		finalAmount = *bodyInvoiceAmount
	}

	if err := tx.Model(&bill).Updates(map[string]interface{}{
		"invoice_amount": finalAmount.StringFixed(2),
		"open_balance":   finalAmount.StringFixed(2),
		"tax_total":      totalTax.StringFixed(2),
	}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	pkgActivity.Log(s.db, "accounting", "vendor_bill_create",
		pkgActivity.WithEntity(bill.IDVendorAPInvoice),
		pkgActivity.WithDetails(map[string]interface{}{
			"vendor_id":      bill.VendorID,
			"invoice_amount": finalAmount.StringFixed(2),
			"invoice_date":   fmtDate(bill.InvoiceDate),
		}),
	)

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	bill.InvoiceAmount = finalAmount.StringFixed(2)
	bill.OpenBalance = finalAmount.StringFixed(2)
	bill.TaxTotal = totalTax.StringFixed(2)
	return bill.ToMap(), nil
}

// ── GET /transactions/:vendor_id ─────────────────────────────────────────────

func (s *Service) GetTransactionsByVendor(username string, vendorID int) ([]map[string]interface{}, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	_ = emp
	if err != nil {
		return nil, err
	}

	var vendor vendorModel.Vendor
	if err := s.db.First(&vendor, vendorID).Error; err != nil {
		return nil, errors.New("vendor not found")
	}

	accLoc, err := s.resolveAccountingLocation(loc)
	if err != nil {
		return nil, err
	}
	locID := int64(accLoc.IDLocation)

	// payments with payment_method
	type payRow struct {
		vendorModel.PaymentToVendorTransaction
		ShortName  *string
		MethodName string
	}
	var pays []payRow
	s.db.Table("payment_to_vendor_transaction p").
		Select("p.*, pm.short_name, pm.method_name").
		Joins("LEFT JOIN payment_method pm ON pm.id_payment_method = p.payment_method_id").
		Where("p.vendor_id = ? AND p.location_id = ?", vendorID, locID).
		Order("p.payment_date DESC").
		Scan(&pays)

	var transactions []map[string]interface{}
	for _, pay := range pays {
		var accountVal interface{}
		if pay.ShortName != nil && *pay.ShortName != "" {
			accountVal = *pay.ShortName
		} else if pay.MethodName != "" {
			accountVal = pay.MethodName
		}
		transactions = append(transactions, map[string]interface{}{
			"type":    "Payment",
			"number":  fmt.Sprintf("%d", pay.IDPaymentVendorTransaction),
			"date":    fmtDate(pay.PaymentDate),
			"account": accountVal,
			"amount":  toDecimal(pay.Amount).InexactFloat64(),
		})
	}

	// vendor return credits
	groupLocIDs := s.groupLocationIDs(accLoc.StoreID)

	rtvInvoiceIDs := s.rtvInvoiceIDsForVendor(vendorID, groupLocIDs)

	if len(rtvInvoiceIDs) > 0 {
		var returnPays []vendorModel.VendorReturnPayment
		s.db.Preload("PaymentMethod").
			Where("return_to_vendor_invoice_id IN ? AND payment_method_id != ?", rtvInvoiceIDs, adjustmentPaymentMethodID).
			Order("payment_timestamp DESC").
			Find(&returnPays)

		for _, rp := range returnPays {
			account := pmLabel(rp.PaymentMethod)
			transactions = append(transactions, map[string]interface{}{
				"type":    "VendorReturnCredit",
				"number":  fmt.Sprintf("%d", rp.ReturnToVendorInvoiceID),
				"date":    rp.PaymentTimestamp.Format(time.RFC3339),
				"account": account,
				"amount":  rp.Amount,
			})
		}
	}

	if transactions == nil {
		transactions = []map[string]interface{}{}
	}
	return transactions, nil
}

// ── POST /add_payment ────────────────────────────────────────────────────────

type AddPaymentInput struct {
	VendorID                int64         `json:"vendor_id"`
	Amount                  interface{}   `json:"amount"`
	Date                    string        `json:"date"`
	PaymentMethodID         *int          `json:"payment_method_id"`
	PayForInvoices          []interface{} `json:"pay_for_invoices"`
	CorrectedInvoiceID      *int64        `json:"corrected_invoice_id"`
	CorrectedInvoiceBalance interface{}   `json:"corrected_invoice_balance"`
	Note                    *string       `json:"note"`
}

func (s *Service) AddPaymentToVendor(username string, data map[string]interface{}) (map[string]interface{}, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	vendorIDRaw, _ := data["vendor_id"]
	amountRaw := data["amount"]
	dateStr, _ := data["date"].(string)
	pmIDRaw, _ := data["payment_method_id"]

	if vendorIDRaw == nil || amountRaw == nil || dateStr == "" || pmIDRaw == nil {
		return nil, errors.New("vendor_id, amount, date, payment_method_id are required")
	}

	vendorID, _ := toInt(vendorIDRaw)
	pmID, _ := toInt(pmIDRaw)

	payDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, errors.New("invalid date format, expected YYYY-MM-DD")
	}
	paymentAmount := toDecimal(amountRaw)

	var payForIDs []int64
	if raw, ok := data["pay_for_invoices"]; ok && raw != nil {
		if arr, ok := raw.([]interface{}); ok {
			for _, v := range arr {
				if n, err := toInt(v); err == nil {
					payForIDs = append(payForIDs, int64(n))
				}
			}
		}
	}

	var correctedID *int64
	if v, ok := data["corrected_invoice_id"]; ok && v != nil {
		n, _ := toInt(v)
		id := int64(n)
		correctedID = &id
	}

	if len(payForIDs) == 0 && correctedID == nil {
		return nil, errors.New("pay_for_invoices or corrected_invoice_id is required")
	}

	var vendor vendorModel.Vendor
	if err := s.db.First(&vendor, vendorID).Error; err != nil {
		return nil, errors.New("vendor not found")
	}

	accLoc, err := s.resolveAccountingLocation(loc)
	if err != nil {
		return nil, err
	}
	locID := int64(accLoc.IDLocation)

	// deduplicate & remove correctedID from payForIDs
	seen := map[int64]bool{}
	var dedupIDs []int64
	for _, id := range payForIDs {
		if !seen[id] && (correctedID == nil || id != *correctedID) {
			seen[id] = true
			dedupIDs = append(dedupIDs, id)
		}
	}

	// Load full-pay invoices
	var invoicesFull []vendorModel.VendorAPInvoice
	sumFullPaid := decimal.Zero
	if len(dedupIDs) > 0 {
		if err := s.db.Where("id_vendor_ap_invoice IN ? AND vendor_id = ? AND location_id = ?", dedupIDs, vendorID, locID).
			Find(&invoicesFull).Error; err != nil {
			return nil, err
		}
		foundIDs := map[int64]bool{}
		for _, inv := range invoicesFull {
			foundIDs[inv.IDVendorAPInvoice] = true
		}
		for _, id := range dedupIDs {
			if !foundIDs[id] {
				return nil, fmt.Errorf("VendorAPInvoice not found for this vendor+location: %d", id)
			}
		}
		for _, inv := range invoicesFull {
			sumFullPaid = sumFullPaid.Add(toDecimal(inv.OpenBalance))
		}
	}

	// Load partial invoice
	var partialInv *vendorModel.VendorAPInvoice
	sumPartialPaid := decimal.Zero
	var correctedBalance *decimal.Decimal
	if correctedID != nil {
		var inv vendorModel.VendorAPInvoice
		if err := s.db.Where("id_vendor_ap_invoice = ? AND vendor_id = ? AND location_id = ?", *correctedID, vendorID, locID).
			First(&inv).Error; err != nil {
			return nil, errors.New("corrected_invoice_id not found for this vendor+location")
		}
		partialInv = &inv

		if data["corrected_invoice_balance"] == nil {
			return nil, errors.New("corrected_invoice_balance is required when corrected_invoice_id is set")
		}
		cb := toDecimal(data["corrected_invoice_balance"])
		correctedBalance = &cb
		oldBal := toDecimal(inv.OpenBalance)
		if cb.IsNegative() {
			return nil, errors.New("corrected_invoice_balance cannot be negative")
		}
		if cb.GreaterThan(oldBal) {
			return nil, errors.New("corrected_invoice_balance cannot be greater than current open_balance")
		}
		sumPartialPaid = oldBal.Sub(cb)
	}

	// Validate single account
	accountIDs := map[int64]bool{}
	for _, inv := range invoicesFull {
		if inv.VendorLocationAccountID == nil {
			return nil, errors.New("some invoices have no vendor_location_account_id")
		}
		accountIDs[*inv.VendorLocationAccountID] = true
	}
	if partialInv != nil {
		if partialInv.VendorLocationAccountID == nil {
			return nil, errors.New("corrected invoice has no vendor_location_account_id")
		}
		accountIDs[*partialInv.VendorLocationAccountID] = true
	}
	if len(accountIDs) != 1 {
		return nil, errors.New("invoices belong to different vendor accounts")
	}
	var accountID int64
	for id := range accountIDs {
		accountID = id
	}

	var vla vendorModel.VendorLocationAccount
	if err := s.db.Where("id_vendor_location_account = ? AND vendor_id = ? AND location_id = ?", accountID, vendorID, locID).
		First(&vla).Error; err != nil {
		return nil, errors.New("vendor_location_account_id from invoice not found for this vendor+location")
	}

	Q := decimal.NewFromFloat(0.01)
	totalExpected := sumFullPaid.Add(sumPartialPaid).Round(2)
	paymentAmount = paymentAmount.Round(2)

	if !totalExpected.Equal(paymentAmount) {
		return nil, fmt.Errorf("payment amount mismatch: expected %s got %s", totalExpected.StringFixed(2), paymentAmount.StringFixed(2))
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	_ = Q

	var noteStr *string
	if v, ok := data["note"].(string); ok && v != "" {
		noteStr = &v
	}

	payTx := vendorModel.PaymentToVendorTransaction{
		VendorID:                vendorID,
		LocationID:              locID,
		EmployeeID:              int64(emp.IDEmployee),
		PaymentMethodID:         pmID,
		PaymentDate:             payDate,
		Amount:                  paymentAmount.StringFixed(2),
		Note:                    noteStr,
		VendorLocationAccountID: &accountID,
	}
	if err := tx.Create(&payTx).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	closedIDs := make([]int64, 0, len(invoicesFull))
	for _, inv := range invoicesFull {
		if err := tx.Model(&inv).Updates(map[string]interface{}{
			"open_balance": "0.00",
			"status":       "Paid",
		}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		closedIDs = append(closedIDs, inv.IDVendorAPInvoice)
	}

	if partialInv != nil {
		newStatus := "Paid"
		if correctedBalance.GreaterThan(decimal.Zero) {
			newStatus = "Partially Paid"
		}
		if err := tx.Model(partialInv).Updates(map[string]interface{}{
			"open_balance": correctedBalance.StringFixed(2),
			"status":       newStatus,
		}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	pkgActivity.Log(s.db, "accounting", "vendor_payment",
		pkgActivity.WithEntity(payTx.IDPaymentVendorTransaction),
		pkgActivity.WithDetails(map[string]interface{}{
			"vendor_id":             vendorID,
			"amount":                paymentAmount.StringFixed(2),
			"payment_method_id":     pmID,
			"paid_invoice_ids":      dedupIDs,
			"corrected_invoice_id":  correctedID,
		}),
	)

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	var corrBalStr interface{}
	if correctedBalance != nil {
		corrBalStr = correctedBalance.StringFixed(2)
	}

	return map[string]interface{}{
		"message":                          "Payment applied successfully",
		"payment_id":                       payTx.IDPaymentVendorTransaction,
		"vendor_location_account_id":       accountID,
		"account_number":                   vla.AccountNumber,
		"amount":                           paymentAmount.StringFixed(2),
		"closed_invoices":                  closedIDs,
		"corrected_invoice_id":             correctedID,
		"corrected_invoice_new_balance":    corrBalStr,
	}, nil
}

// ── GET /vendors-balances ────────────────────────────────────────────────────

func (s *Service) GetVendorsBalances(username string) ([]map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	accLoc, err := s.resolveAccountingLocation(loc)
	if err != nil {
		return nil, err
	}
	locID := int64(accLoc.IDLocation)

	type balanceRow struct {
		VendorID   int     `gorm:"column:vendor_id"`
		Name       string  `gorm:"column:name"`
		Balance    float64 `gorm:"column:balance"`
		Attach     *string `gorm:"column:attach"`
	}
	var rows []balanceRow
	s.db.Raw(`
		SELECT v.id_vendor AS vendor_id, v.vendor_name AS name,
		       COALESCE(SUM(i.open_balance::numeric), 0) AS balance,
		       MAX(i.attachment_url) AS attach
		FROM vendor v
		LEFT JOIN vendor_ap_invoice i
		       ON i.vendor_id = v.id_vendor AND i.location_id = ?
		GROUP BY v.id_vendor, v.vendor_name
		ORDER BY v.vendor_name ASC
	`, locID).Scan(&rows)

	// subtract return credits for this store group
	groupLocIDs := s.groupLocationIDs(accLoc.StoreID)
	rtvInvoiceIDs := s.rtvInvoiceIDsByGroupLocs(groupLocIDs)

	rtvPaidByVendor := map[int64]decimal.Decimal{}
	if len(rtvInvoiceIDs) > 0 {
		// map rtv_id → vendor_id
		type rtvRow struct {
			InvoiceID int64 `gorm:"column:id_return_to_vendor_invoice"`
			VendorID  int64 `gorm:"column:vendor_id"`
		}
		var rtvRows []rtvRow
		s.db.Table("return_to_vendor_invoice").
			Select("id_return_to_vendor_invoice, vendor_id").
			Where("id_return_to_vendor_invoice IN ?", rtvInvoiceIDs).
			Scan(&rtvRows)
		rtvIDToVendor := map[int64]int64{}
		for _, r := range rtvRows {
			rtvIDToVendor[r.InvoiceID] = r.VendorID
		}

		type paidRow struct {
			RtvID int64   `gorm:"column:return_to_vendor_invoice_id"`
			Total float64 `gorm:"column:total"`
		}
		var paidRows []paidRow
		s.db.Table("vendor_return_payment").
			Select("return_to_vendor_invoice_id, SUM(amount) AS total").
			Where("return_to_vendor_invoice_id IN ?", rtvInvoiceIDs).
			Group("return_to_vendor_invoice_id").
			Scan(&paidRows)
		for _, pr := range paidRows {
			vID := rtvIDToVendor[pr.RtvID]
			rtvPaidByVendor[vID] = rtvPaidByVendor[vID].Add(decimal.NewFromFloat(pr.Total))
		}
	}

	result := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		bal := decimal.NewFromFloat(r.Balance)
		if credit, ok := rtvPaidByVendor[int64(r.VendorID)]; ok {
			bal = bal.Sub(credit)
		}
		attach := ""
		if r.Attach != nil {
			attach = *r.Attach
		}
		result[i] = map[string]interface{}{
			"vendor_id": r.VendorID,
			"name":      r.Name,
			"balance":   bal.StringFixed(2),
			"attach":    attach,
		}
	}
	return result, nil
}

// ── Account number CRUD ───────────────────────────────────────────────────────

func (s *Service) ListVendorLocationAccounts(username string, vendorID int, statusFilter string, isActiveFilter *bool) (map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	var vendor vendorModel.Vendor
	if err := s.db.First(&vendor, vendorID).Error; err != nil {
		return nil, errors.New("vendor not found")
	}
	locID := int64(loc.IDLocation)

	q := s.db.Where("vendor_id = ? AND location_id = ?", vendorID, locID)
	if statusFilter != "" {
		q = q.Where("status = ?", statusFilter)
	}
	if isActiveFilter != nil {
		q = q.Where("is_active = ?", *isActiveFilter)
	}

	var accs []vendorModel.VendorLocationAccount
	q.Order("created_at DESC").Find(&accs)

	items := make([]map[string]interface{}, len(accs))
	for i, a := range accs {
		items[i] = a.ToMap()
	}
	return map[string]interface{}{
		"vendor_id":   vendorID,
		"location_id": locID,
		"items":       items,
	}, nil
}

func (s *Service) UpdateVendorAccountNumber(username string, vendorID int, accID int64, data map[string]interface{}) (map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	locID := int64(loc.IDLocation)

	var row vendorModel.VendorLocationAccount
	if err := s.db.Where("id_vendor_location_account = ? AND vendor_id = ? AND location_id = ?", accID, vendorID, locID).
		First(&row).Error; err != nil {
		return nil, errors.New("account mapping not found")
	}

	updates := map[string]interface{}{}

	if v, ok := data["account_number"]; ok && v != nil {
		num := strings.TrimSpace(v.(string))
		if num == "" {
			return nil, errors.New("account_number cannot be empty")
		}
		updates["account_number"] = num
	}

	if v, ok := data["status"]; ok && v != nil {
		st := strings.TrimSpace(v.(string))
		if !validAccountStatuses[st] {
			return nil, fmt.Errorf("status must be one of Active, Blocked, Closed")
		}
		updates["status"] = st
	}

	if _, ok := data["qb_vendor_ref"]; ok {
		updates["qb_vendor_ref"] = data["qb_vendor_ref"]
	}
	if _, ok := data["note"]; ok {
		updates["note"] = data["note"]
	}

	if len(updates) > 0 {
		if err := s.db.Model(&row).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	s.db.First(&row, accID)
	return row.ToMap(), nil
}

func (s *Service) DeleteVendorAccountNumber(username string, vendorID int, accID int64) (map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	locID := int64(loc.IDLocation)

	var row vendorModel.VendorLocationAccount
	if err := s.db.Where("id_vendor_location_account = ? AND vendor_id = ? AND location_id = ?", accID, vendorID, locID).
		First(&row).Error; err != nil {
		return nil, errors.New("account mapping not found")
	}

	var invCount int64
	s.db.Model(&vendorModel.VendorAPInvoice{}).
		Where("vendor_id = ? AND location_id = ? AND vendor_location_account_id = ?", vendorID, locID, accID).
		Count(&invCount)
	if invCount > 0 {
		return nil, fmt.Errorf("cannot delete: this account number is already used in %d AP invoices", invCount)
	}

	var payCount int64
	s.db.Model(&vendorModel.PaymentToVendorTransaction{}).
		Where("vendor_id = ? AND location_id = ? AND vendor_location_account_id = ?", vendorID, locID, accID).
		Count(&payCount)
	if payCount > 0 {
		return nil, fmt.Errorf("cannot delete: this account number is already used in %d payments", payCount)
	}

	if err := s.db.Delete(&row).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message":                  "Deleted",
		"vendor_id":                vendorID,
		"location_id":              locID,
		"id_vendor_location_account": accID,
	}, nil
}

func (s *Service) CreateVendorAccountNumber(username string, vendorID int, data map[string]interface{}) (map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	locID := int64(loc.IDLocation)

	accountNumber := strings.TrimSpace(stringVal(data["account_number"]))
	status := strings.TrimSpace(stringVal(data["status"]))
	if status == "" {
		status = "Active"
	}
	if accountNumber == "" {
		return nil, errors.New("account_number is required")
	}
	if !validAccountStatuses[status] {
		return nil, errors.New("status must be one of Active, Blocked, Closed")
	}

	var qbRef, note *string
	if v, ok := data["qb_vendor_ref"].(string); ok && v != "" {
		qbRef = &v
	}
	if v, ok := data["note"].(string); ok && v != "" {
		note = &v
	}

	// return existing if found
	var existing vendorModel.VendorLocationAccount
	if err := s.db.Where("vendor_id = ? AND location_id = ? AND account_number = ?", vendorID, locID, accountNumber).
		Order("created_at DESC").First(&existing).Error; err == nil {
		return existing.ToMap(), nil
	}

	today := utcToday()
	var validFrom, validTo *time.Time
	if status == "Active" {
		validFrom = &today
	} else {
		validTo = &today
	}
	_ = validFrom
	_ = validTo

	statusPtr := &status
	row := vendorModel.VendorLocationAccount{
		VendorID:      vendorID,
		LocationID:    locID,
		AccountNumber: accountNumber,
		Status:        statusPtr,
		IsActive:      status == "Active",
		QbVendorRef:   qbRef,
		Note:          note,
	}
	if err := s.db.Create(&row).Error; err != nil {
		return nil, err
	}
	return row.ToMap(), nil
}

// ── GET /notify/terms ────────────────────────────────────────────────────────

func (s *Service) GetTermsNotifyList(username string, days int, vendorID *int, limit int) (map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	accLoc, err := s.resolveAccountingLocation(loc)
	if err != nil {
		return nil, err
	}
	locID := int64(accLoc.IDLocation)

	if days < 0 {
		days = 0
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)

	q := s.db.Where("location_id = ? AND terms IS NOT NULL AND terms > 0 AND open_balance::numeric > 0", locID).
		Order("invoice_date ASC, id_vendor_ap_invoice ASC").
		Limit(limit)
	if vendorID != nil {
		q = q.Where("vendor_id = ?", *vendorID)
	}

	var invoices []vendorModel.VendorAPInvoice
	q.Find(&invoices)

	// load vendor names and accounts
	vendorIDs := map[int]struct{}{}
	accIDs2 := map[int64]struct{}{}
	for _, inv := range invoices {
		vendorIDs[inv.VendorID] = struct{}{}
		if inv.VendorLocationAccountID != nil {
			accIDs2[*inv.VendorLocationAccountID] = struct{}{}
		}
	}
	vendorNames := map[int]string{}
	if len(vendorIDs) > 0 {
		ids := make([]int, 0, len(vendorIDs))
		for id := range vendorIDs {
			ids = append(ids, id)
		}
		var vs []vendorModel.Vendor
		s.db.Where("id_vendor IN ?", ids).Find(&vs)
		for _, v := range vs {
			vendorNames[v.IDVendor] = v.VendorName
		}
	}
	accNumbers2 := map[int64]string{}
	if len(accIDs2) > 0 {
		ids := make([]int64, 0, len(accIDs2))
		for id := range accIDs2 {
			ids = append(ids, id)
		}
		var accs []vendorModel.VendorLocationAccount
		s.db.Where("id_vendor_location_account IN ?", ids).Find(&accs)
		for _, a := range accs {
			accNumbers2[a.IDVendorLocationAccount] = a.AccountNumber
		}
	}

	type vendorSummary struct {
		VendorID                int64
		VendorName              string
		ItemsCount              int
		TotalRemainingToPayNow  decimal.Decimal
		EarliestDueDate         time.Time
		OverdueCount            int
	}

	var items []map[string]interface{}
	vendorsMap := map[int64]*vendorSummary{}
	totalRemaining := decimal.Zero

	for _, inv := range invoices {
		if toDecimal(inv.OpenBalance).LessThanOrEqual(decimal.Zero) {
			continue
		}
		periods := pkgAccounting.TermsPeriodStatuses(apInvoiceAdapter{&inv})
		if len(periods) == 0 {
			continue
		}
		periodsTotal := periods[0].PeriodsTotal

		for _, p := range periods {
			daysLeft := int(p.DueDate.Sub(today).Hours() / 24)
			if daysLeft > days {
				continue
			}
			isLast := (p.PeriodNo == periodsTotal)

			var remainingToPayNow decimal.Decimal
			if !isLast {
				if p.HasAnyPaymentInPeriod {
					continue
				}
				remainingToPayNow = p.PeriodAmount
			} else {
				remainingToPayNow = pkgAccounting.D2(toDecimal(inv.OpenBalance))
			}

			notifyFrom := p.DueDate.AddDate(0, 0, -days)

			var accNum interface{}
			if inv.VendorLocationAccountID != nil {
				if num, ok := accNumbers2[*inv.VendorLocationAccountID]; ok {
					accNum = num
				}
			}

			item := map[string]interface{}{
				"vendor_id":                      inv.VendorID,
				"vendor_name":                    vendorNames[inv.VendorID],
				"invoice_id":                     inv.IDVendorAPInvoice,
				"invoice_number":                 inv.InvoiceNumber,
				"invoice_date":                   fmtDate(inv.InvoiceDate),
				"terms":                          inv.Terms,
				"period_no":                      p.PeriodNo,
				"periods_total":                  periodsTotal,
				"period_due_date":                p.DueDate.Format("2006-01-02"),
				"notify_from":                    notifyFrom.Format("2006-01-02"),
				"days_left":                      daysLeft,
				"is_overdue":                     daysLeft < 0,
				"paid_in_period":                 p.PaidInPeriod.StringFixed(2),
				"has_any_payment_in_period":      p.HasAnyPaymentInPeriod,
				"remaining_to_pay_now":           remainingToPayNow.StringFixed(2),
				"invoice_amount":                 pkgAccounting.D2(toDecimal(inv.InvoiceAmount)).StringFixed(2),
				"open_balance":                   pkgAccounting.D2(toDecimal(inv.OpenBalance)).StringFixed(2),
				"vendor_location_account_id":     inv.VendorLocationAccountID,
				"account_number":                 accNum,
				"attachment_url":                 inv.AttachmentURL,
			}
			items = append(items, item)
			totalRemaining = totalRemaining.Add(remainingToPayNow)

			vID := int64(inv.VendorID)
			vs := vendorsMap[vID]
			if vs == nil {
				vs = &vendorSummary{
					VendorID:       vID,
					VendorName:     vendorNames[inv.VendorID],
					EarliestDueDate: p.DueDate,
				}
				vendorsMap[vID] = vs
			}
			vs.ItemsCount++
			vs.TotalRemainingToPayNow = vs.TotalRemainingToPayNow.Add(remainingToPayNow)
			if p.DueDate.Before(vs.EarliestDueDate) {
				vs.EarliestDueDate = p.DueDate
			}
			if daysLeft < 0 {
				vs.OverdueCount++
			}
		}
	}

	// sort items by period_due_date
	sortByDate(items, "period_due_date")

	vendors := make([]map[string]interface{}, 0, len(vendorsMap))
	for _, vs := range vendorsMap {
		vendors = append(vendors, map[string]interface{}{
			"vendor_id":                     vs.VendorID,
			"vendor_name":                   vs.VendorName,
			"items_count":                   vs.ItemsCount,
			"total_remaining_to_pay_now":    vs.TotalRemainingToPayNow.StringFixed(2),
			"earliest_due_date":             vs.EarliestDueDate.Format("2006-01-02"),
			"overdue_count":                 vs.OverdueCount,
		})
	}
	sortByDate(vendors, "earliest_due_date")

	if items == nil {
		items = []map[string]interface{}{}
	}

	return map[string]interface{}{
		"today":                      today.Format("2006-01-02"),
		"days":                       days,
		"vendor_count":               len(vendors),
		"items_count":                len(items),
		"total_remaining_to_pay_now": totalRemaining.StringFixed(2),
		"vendors":                    vendors,
		"items":                      items,
	}, nil
}

// ── GET /return-to-vendor-invoices/:vendor_id ────────────────────────────────

func (s *Service) GetReturnToVendorInvoices(vendorID, page, perPage int) (map[string]interface{}, error) {
	var total int64
	s.db.Model(&vendorModel.ReturnToVendorInvoice{}).Where("vendor_id = ?", vendorID).Count(&total)

	var rows []vendorModel.ReturnToVendorInvoice
	offset := (page - 1) * perPage
	s.db.Preload("Items").
		Where("vendor_id = ?", vendorID).
		Order("created_date DESC").
		Offset(offset).Limit(perPage).
		Find(&rows)

	items := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		var creditStr, purchaseStr interface{}
		if r.CreditAmount != nil {
			creditStr = fmt.Sprintf("%.2f", *r.CreditAmount)
		}
		purchaseStr = fmt.Sprintf("%.2f", r.PurchaseTotal)

		rtvItems := make([]map[string]interface{}, len(r.Items))
		for j, it := range r.Items {
			rtvItems[j] = it.ToMap()
		}

		items[i] = map[string]interface{}{
			"id_return_to_vendor_invoice": r.IDReturnToVendorInvoice,
			"vendor_id":                  r.VendorID,
			"created_date":               r.CreatedDate.Format("2006-01-02"),
			"credit_amount":              creditStr,
			"purchase_total":             purchaseStr,
			"quantity":                   r.Quantity,
			"items":                      rtvItems,
		}
	}

	pages := int((total + int64(perPage) - 1) / int64(perPage))
	return map[string]interface{}{
		"items":    items,
		"total":    total,
		"page":     page,
		"per_page": perPage,
		"pages":    pages,
	}, nil
}

// ── PUT /return-to-vendor-invoices/:rtv_id/credit ────────────────────────────

func (s *Service) UpdateReturnToVendorCredit(rtvID int64, creditAmount float64) (map[string]interface{}, error) {
	var rtv vendorModel.ReturnToVendorInvoice
	if err := s.db.First(&rtv, rtvID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("return to vendor invoice not found")
		}
		return nil, err
	}
	rtv.CreditAmount = &creditAmount
	if err := s.db.Model(&rtv).Update("credit_amount", creditAmount).Error; err != nil {
		return nil, err
	}
	return rtv.ToMap(), nil
}

// ── internal helpers ──────────────────────────────────────────────────────────

func (s *Service) groupLocationIDs(storeID int) []int64 {
	var locs []locModel.Location
	s.db.Where("store_id = ?", storeID).Find(&locs)
	ids := make([]int64, len(locs))
	for i, l := range locs {
		ids[i] = int64(l.IDLocation)
	}
	return ids
}

func (s *Service) rtvInvoiceIDsForVendor(vendorID int, groupLocIDs []int64) []int64 {
	if len(groupLocIDs) == 0 {
		return nil
	}
	var ids []int64
	s.db.Table("return_to_vendor_invoice rtv").
		Select("DISTINCT rtv.id_return_to_vendor_invoice").
		Joins("JOIN return_to_vendor_item ri ON ri.return_to_vendor_invoice_id = rtv.id_return_to_vendor_invoice").
		Joins("JOIN inventory inv ON inv.id_inventory = ri.inventory_id").
		Where("rtv.vendor_id = ? AND inv.location_id IN ?", vendorID, groupLocIDs).
		Pluck("rtv.id_return_to_vendor_invoice", &ids)
	return ids
}

func (s *Service) rtvInvoiceIDsByGroupLocs(groupLocIDs []int64) []int64 {
	if len(groupLocIDs) == 0 {
		return nil
	}
	var ids []int64
	s.db.Table("return_to_vendor_invoice rtv").
		Select("DISTINCT rtv.id_return_to_vendor_invoice").
		Joins("JOIN return_to_vendor_item ri ON ri.return_to_vendor_invoice_id = rtv.id_return_to_vendor_invoice").
		Joins("JOIN inventory inv ON inv.id_inventory = ri.inventory_id").
		Where("inv.location_id IN ?", groupLocIDs).
		Pluck("rtv.id_return_to_vendor_invoice", &ids)
	return ids
}

func toDecimal(v interface{}) decimal.Decimal {
	if v == nil {
		return decimal.Zero
	}
	switch val := v.(type) {
	case string:
		d, _ := decimal.NewFromString(val)
		return d
	case float64:
		return decimal.NewFromFloat(val)
	case float32:
		return decimal.NewFromFloat32(val)
	case int:
		return decimal.NewFromInt(int64(val))
	case int64:
		return decimal.NewFromInt(val)
	default:
		d, _ := decimal.NewFromString(fmt.Sprintf("%v", val))
		return d
	}
}

func toInt(v interface{}) (int, error) {
	if v == nil {
		return 0, fmt.Errorf("nil value")
	}
	switch val := v.(type) {
	case float64:
		return int(val), nil
	case int:
		return val, nil
	case int64:
		return int(val), nil
	case string:
		var n int
		_, err := fmt.Sscanf(val, "%d", &n)
		return n, err
	default:
		return 0, fmt.Errorf("unsupported type")
	}
}

func stringVal(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func sortByDate(items []map[string]interface{}, key string) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0; j-- {
			a, _ := items[j-1][key].(string)
			b, _ := items[j][key].(string)
			if a > b {
				items[j-1], items[j] = items[j], items[j-1]
			} else {
				break
			}
		}
	}
}
