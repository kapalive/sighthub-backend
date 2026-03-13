package accounting_service

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"

	vendorModel "sighthub-backend/internal/models/vendors"
)

// ── GET /ledger/:vendor_id ────────────────────────────────────────────────────

func (s *Service) GetVendorLedger(username string, vendorID int) (map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
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

	var entries []map[string]interface{}

	// ── 1. Bills ─────────────────────────────────────────────────────────────
	var bills []vendorModel.VendorAPInvoice
	s.db.Where("vendor_id = ? AND location_id = ?", vendorID, locID).Find(&bills)

	// load account numbers
	billAccIDs := map[int64]struct{}{}
	for _, b := range bills {
		if b.VendorLocationAccountID != nil {
			billAccIDs[*b.VendorLocationAccountID] = struct{}{}
		}
	}
	billAccNums := map[int64]string{}
	if len(billAccIDs) > 0 {
		ids := make([]int64, 0, len(billAccIDs))
		for id := range billAccIDs {
			ids = append(ids, id)
		}
		var accs []vendorModel.VendorLocationAccount
		s.db.Where("id_vendor_location_account IN ?", ids).Find(&accs)
		for _, a := range accs {
			billAccNums[a.IDVendorLocationAccount] = a.AccountNumber
		}
	}

	for _, inv := range bills {
		var accNum interface{}
		if inv.VendorLocationAccountID != nil {
			if num, ok := billAccNums[*inv.VendorLocationAccountID]; ok {
				accNum = num
			}
		}
		entries = append(entries, map[string]interface{}{
			"type":         "Bill",
			"id":           inv.IDVendorAPInvoice,
			"date":         fmtDate(inv.InvoiceDate),
			"created_at":   fmtTimePtr(inv.CreatedAt),
			"due_date":     fmtDate(inv.BillDueDate),
			"number":       inv.InvoiceNumber,
			"amount":       inv.InvoiceAmount,
			"open_balance": inv.OpenBalance,
			"status":       inv.Status,
			"account_number": accNum,
			"note":         inv.Note,
		})
	}

	// ── 2. Payments (skip adjustments) ───────────────────────────────────────
	type payWithPM struct {
		vendorModel.PaymentToVendorTransaction
		ShortName  *string `gorm:"column:short_name"`
		MethodName string  `gorm:"column:method_name"`
	}
	var pays []payWithPM
	s.db.Table("payment_to_vendor_transaction p").
		Select("p.*, pm.short_name, pm.method_name").
		Joins("LEFT JOIN payment_method pm ON pm.id_payment_method = p.payment_method_id").
		Where("p.vendor_id = ? AND p.location_id = ? AND p.payment_method_id != ?", vendorID, locID, adjustmentPaymentMethodID).
		Find(&pays)

	for _, pay := range pays {
		var account interface{}
		if pay.ShortName != nil && *pay.ShortName != "" {
			account = *pay.ShortName
		} else if pay.MethodName != "" {
			account = pay.MethodName
		}
		amt := "-" + toDecimal(pay.Amount).StringFixed(2)
		entries = append(entries, map[string]interface{}{
			"type":         "Payment",
			"id":           pay.IDPaymentVendorTransaction,
			"date":         fmtDate(pay.PaymentDate),
			"created_at":   fmtTimePtr(pay.CreatedAt),
			"due_date":     nil,
			"number":       fmt.Sprintf("%d", pay.IDPaymentVendorTransaction),
			"amount":       amt,
			"open_balance": nil,
			"status":       nil,
			"account_number": account,
			"note":         pay.Note,
		})
	}

	// ── 3. Return to vendor invoices (negative) ───────────────────────────────
	groupLocIDs := s.groupLocationIDs(accLoc.StoreID)
	rtvIDs := s.rtvInvoiceIDsForVendor(vendorID, groupLocIDs)

	var rtvInvoices []vendorModel.ReturnToVendorInvoice
	if len(rtvIDs) > 0 {
		s.db.Where("id_return_to_vendor_invoice IN ?", rtvIDs).Find(&rtvInvoices)
	}

	for _, rtv := range rtvInvoices {
		var amount float64
		if rtv.CreditAmount != nil {
			amount = *rtv.CreditAmount
		} else {
			amount = rtv.PurchaseTotal
		}
		entries = append(entries, map[string]interface{}{
			"type":         "Return",
			"id":           rtv.IDReturnToVendorInvoice,
			"date":         rtv.CreatedDate.Format("2006-01-02"),
			"created_at":   rtv.CreatedDate.Format(time.RFC3339),
			"due_date":     nil,
			"number":       fmt.Sprintf("%d", rtv.IDReturnToVendorInvoice),
			"amount":       fmt.Sprintf("-%.2f", amount),
			"open_balance": nil,
			"status":       rtv.Status,
			"account_number": nil,
			"note":         rtv.Note,
			"quantity":     rtv.Quantity,
			"ar_number":    rtv.ARNumber,
		})
	}

	// ── 4. Return credits (positive) ──────────────────────────────────────────
	if len(rtvIDs) > 0 {
		var returnPays []vendorModel.VendorReturnPayment
		s.db.Preload("PaymentMethod").
			Where("return_to_vendor_invoice_id IN ? AND payment_method_id != ?", rtvIDs, adjustmentPaymentMethodID).
			Find(&returnPays)

		for _, rp := range returnPays {
			account := pmLabel(rp.PaymentMethod)
			entries = append(entries, map[string]interface{}{
				"type":         "ReturnCredit",
				"id":           rp.IDVendorReturnPayment,
				"date":         rp.PaymentTimestamp.Format(time.RFC3339),
				"created_at":   rp.PaymentTimestamp.Format(time.RFC3339),
				"due_date":     nil,
				"number":       fmt.Sprintf("%d", rp.ReturnToVendorInvoiceID),
				"amount":       fmt.Sprintf("%.2f", rp.Amount),
				"open_balance": nil,
				"status":       nil,
				"account_number": account,
				"note":         rp.Notes,
			})
		}
	}

	// Sort by created_at desc (nil last)
	sort.SliceStable(entries, func(i, j int) bool {
		a := entryTime(entries[i])
		b := entryTime(entries[j])
		return a.After(b)
	})

	if entries == nil {
		entries = []map[string]interface{}{}
	}
	return map[string]interface{}{
		"vendor_id":   vendorID,
		"location_id": locID,
		"entries":     entries,
	}, nil
}

func entryTime(e map[string]interface{}) time.Time {
	v, _ := e["created_at"]
	if v == nil {
		return time.Time{}
	}
	s, ok := v.(string)
	if !ok {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t2, err2 := time.Parse("2006-01-02", s)
		if err2 != nil {
			return time.Time{}
		}
		return t2
	}
	return t
}

// ── GET /ledger/entry ─────────────────────────────────────────────────────────

func (s *Service) GetLedgerEntry(username string, entryType string, entryID int64) (map[string]interface{}, error) {
	_, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	base := map[string]interface{}{
		"type": entryType, "id": nil,
		"vendor_id": nil, "vendor_name": nil, "location_id": nil,
		"employee_id": nil, "employee_name": nil,
		"vendor_location_account_id": nil, "account_number": nil,
		"payment_method_id": nil, "payment_method": nil,
		"invoice_date": nil, "bill_due_date": nil,
		"payment_date": nil, "payment_timestamp": nil, "created_date": nil,
		"created_at": nil,
		"invoice_number": nil, "terms": nil,
		"invoice_amount": nil, "open_balance": nil, "tax_total": nil,
		"attachment_url": nil,
		"amount": nil, "status": nil, "note": nil,
		"purchase_total": nil, "credit_amount": nil,
		"amount_paid": nil, "balance_due": nil,
		"quantity": nil, "ar_number": nil,
		"shipping_service_id": nil, "shipping_service_name": nil,
		"tracking_number": nil, "shipping_number": nil,
		"return_to_vendor_invoice_id": nil, "return_invoice": nil,
		"items": nil, "payments": nil,
	}

	switch entryType {
	case "Bill":
		var inv vendorModel.VendorAPInvoice
		if err := s.db.First(&inv, entryID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("bill not found")
			}
			return nil, err
		}

		// load relations manually
		var accNum interface{}
		if inv.VendorLocationAccountID != nil {
			var acc vendorModel.VendorLocationAccount
			if s.db.First(&acc, *inv.VendorLocationAccountID).Error == nil {
				accNum = acc.AccountNumber
			}
		}
		var vendorName interface{}
		var vendor vendorModel.Vendor
		if s.db.First(&vendor, inv.VendorID).Error == nil {
			vendorName = vendor.VendorName
		}
		empName := s.empName(inv.EmployeeID)

		var apItems []vendorModel.VendorAPInvoiceItem
		s.db.Where("vendor_ap_invoice_id = ?", inv.IDVendorAPInvoice).Find(&apItems)
		itemMaps := make([]map[string]interface{}, len(apItems))
		for i, it := range apItems {
			itemMaps[i] = map[string]interface{}{
				"id_vendor_ap_invoice_item": it.IDVendorAPInvoiceItem,
				"vendor_ap_invoice_id":      it.VendorAPInvoiceID,
				"line_no":                   it.LineNo,
				"quantity":                  it.Quantity,
				"description":               it.Description,
				"price_each":                it.PriceEach,
				"amount":                    it.Amount,
				"tax":                       it.Tax,
			}
		}

		base["id"] = inv.IDVendorAPInvoice
		base["vendor_id"] = inv.VendorID
		base["vendor_name"] = vendorName
		base["location_id"] = inv.LocationID
		base["employee_id"] = inv.EmployeeID
		base["employee_name"] = empName
		base["vendor_location_account_id"] = inv.VendorLocationAccountID
		base["account_number"] = accNum
		base["invoice_number"] = inv.InvoiceNumber
		base["invoice_date"] = fmtDate(inv.InvoiceDate)
		base["bill_due_date"] = fmtDate(inv.BillDueDate)
		base["terms"] = inv.Terms
		base["invoice_amount"] = inv.InvoiceAmount
		base["open_balance"] = inv.OpenBalance
		base["tax_total"] = inv.TaxTotal
		base["status"] = inv.Status
		base["attachment_url"] = inv.AttachmentURL
		base["note"] = inv.Note
		base["created_at"] = fmtTimePtr(inv.CreatedAt)
		base["items"] = itemMaps

	case "Payment":
		var pay vendorModel.PaymentToVendorTransaction
		if err := s.db.First(&pay, entryID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("payment not found")
			}
			return nil, err
		}

		var pm vendorModel.VendorLocationAccount
		_ = pm
		var accNum interface{}
		if pay.VendorLocationAccountID != nil {
			var acc vendorModel.VendorLocationAccount
			if s.db.First(&acc, *pay.VendorLocationAccountID).Error == nil {
				accNum = acc.AccountNumber
			}
		}
		var vendorName interface{}
		var vendor vendorModel.Vendor
		if s.db.First(&vendor, pay.VendorID).Error == nil {
			vendorName = vendor.VendorName
		}
		empName := s.empName(pay.EmployeeID)

		type pmRow struct {
			ShortName  *string `gorm:"column:short_name"`
			MethodName string  `gorm:"column:method_name"`
		}
		var pmR pmRow
		var pmLabel2 interface{}
		if s.db.Raw("SELECT short_name, method_name FROM payment_method WHERE id_payment_method = ?", pay.PaymentMethodID).Scan(&pmR).Error == nil {
			if pmR.ShortName != nil && *pmR.ShortName != "" {
				pmLabel2 = *pmR.ShortName
			} else if pmR.MethodName != "" {
				pmLabel2 = pmR.MethodName
			}
		}

		base["id"] = pay.IDPaymentVendorTransaction
		base["vendor_id"] = pay.VendorID
		base["vendor_name"] = vendorName
		base["location_id"] = pay.LocationID
		base["employee_id"] = pay.EmployeeID
		base["employee_name"] = empName
		base["vendor_location_account_id"] = pay.VendorLocationAccountID
		base["account_number"] = accNum
		base["payment_method_id"] = pay.PaymentMethodID
		base["payment_method"] = pmLabel2
		base["payment_date"] = fmtDate(pay.PaymentDate)
		base["amount"] = pay.Amount
		base["note"] = pay.Note
		base["created_at"] = fmtTimePtr(pay.CreatedAt)

	case "Return":
		var rtv vendorModel.ReturnToVendorInvoice
		if err := s.db.
			Preload("Employee").
			Preload("Vendor").
			Preload("Items").
			Preload("Payments.PaymentMethod").
			Preload("ShippingService").
			First(&rtv, entryID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("return not found")
			}
			return nil, err
		}

		rvBase := rtv.PurchaseTotal
		if rtv.CreditAmount != nil {
			rvBase = *rtv.CreditAmount
		}
		var realPays []vendorModel.VendorReturnPayment
		for _, p := range rtv.Payments {
			if p.PaymentMethodID != adjustmentPaymentMethodID {
				realPays = append(realPays, p)
			}
		}
		var paid float64
		for _, p := range realPays {
			paid += p.Amount
		}
		balance := rvBase - paid

		itemMaps := make([]map[string]interface{}, len(rtv.Items))
		for i, it := range rtv.Items {
			// get sku from inventory via raw query
			var sku *string
			s.db.Raw("SELECT sku FROM inventory WHERE id_inventory = ?", it.InventoryID).Scan(&sku)
			itemMaps[i] = map[string]interface{}{
				"id_return_to_vendor_item": it.IDReturnToVendorItem,
				"inventory_id":             it.InventoryID,
				"sku":                      sku,
				"purchase_cost":            it.PurchaseCost,
				"reason_return":            it.ReasonReturn,
			}
		}

		payMaps := make([]map[string]interface{}, len(realPays))
		for i, p := range realPays {
			payMaps[i] = p.ToMap()
		}

		var vendorName interface{}
		if rtv.Vendor != nil {
			vendorName = rtv.Vendor.VendorName
		}
		var empName interface{}
		if rtv.Employee != nil {
			empName = fmt.Sprintf("%s %s", rtv.Employee.FirstName, rtv.Employee.LastName)
		}
		var svcName interface{}
		if rtv.ShippingService != nil {
			svcName = rtv.ShippingService.ShortName
		}

		base["id"] = rtv.IDReturnToVendorInvoice
		base["vendor_id"] = rtv.VendorID
		base["vendor_name"] = vendorName
		base["employee_id"] = rtv.EmployeeID
		base["employee_name"] = empName
		base["created_date"] = rtv.CreatedDate.Format("2006-01-02")
		base["created_at"] = rtv.CreatedDate.Format(time.RFC3339)
		base["status"] = rtv.Status
		base["purchase_total"] = fmt.Sprintf("%.2f", rtv.PurchaseTotal)
		if rtv.CreditAmount != nil {
			base["credit_amount"] = fmt.Sprintf("%.2f", *rtv.CreditAmount)
		}
		base["amount_paid"] = fmt.Sprintf("%.2f", paid)
		base["balance_due"] = fmt.Sprintf("%.2f", balance)
		base["quantity"] = rtv.Quantity
		base["ar_number"] = rtv.ARNumber
		base["shipping_service_id"] = rtv.ShippingServiceID
		base["shipping_service_name"] = svcName
		base["tracking_number"] = rtv.TrackingNumber
		base["shipping_number"] = rtv.ShippingNumber
		base["note"] = rtv.Note
		base["items"] = itemMaps
		base["payments"] = payMaps

	case "ReturnCredit":
		var rp vendorModel.VendorReturnPayment
		if err := s.db.
			Preload("PaymentMethod").
			Preload("Employee").
			First(&rp, entryID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("return credit not found")
			}
			return nil, err
		}

		// load return invoice + vendor
		var rtv vendorModel.ReturnToVendorInvoice
		var rtvMap interface{}
		var vendorName interface{}
		var vendorID interface{}
		if s.db.Preload("Vendor").First(&rtv, rp.ReturnToVendorInvoiceID).Error == nil {
			if rtv.Vendor != nil {
				vendorName = rtv.Vendor.VendorName
				vendorID = rtv.VendorID
			}
			var creditStr interface{}
			if rtv.CreditAmount != nil {
				creditStr = fmt.Sprintf("%.2f", *rtv.CreditAmount)
			}
			rtvMap = map[string]interface{}{
				"id":             rtv.IDReturnToVendorInvoice,
				"status":         rtv.Status,
				"ar_number":      rtv.ARNumber,
				"purchase_total": fmt.Sprintf("%.2f", rtv.PurchaseTotal),
				"credit_amount":  creditStr,
			}
		}

		empName := interface{}(nil)
		if rp.Employee != nil {
			empName = fmt.Sprintf("%s %s", rp.Employee.FirstName, rp.Employee.LastName)
		}

		base["id"] = rp.IDVendorReturnPayment
		base["vendor_id"] = vendorID
		base["vendor_name"] = vendorName
		base["employee_id"] = rp.EmployeeID
		base["employee_name"] = empName
		base["payment_method_id"] = rp.PaymentMethodID
		base["payment_method"] = pmLabel(rp.PaymentMethod)
		base["payment_timestamp"] = rp.PaymentTimestamp.Format(time.RFC3339)
		base["amount"] = fmt.Sprintf("%.2f", rp.Amount)
		base["note"] = rp.Notes
		base["created_at"] = rp.PaymentTimestamp.Format(time.RFC3339)
		base["return_to_vendor_invoice_id"] = rp.ReturnToVendorInvoiceID
		base["return_invoice"] = rtvMap

	default:
		return nil, fmt.Errorf("unknown type '%s'", entryType)
	}

	return base, nil
}

func (s *Service) empName(empID int64) interface{} {
	type empRow struct {
		FirstName string `gorm:"column:first_name"`
		LastName  string `gorm:"column:last_name"`
	}
	var row empRow
	if s.db.Raw("SELECT first_name, last_name FROM employee WHERE id_employee = ?", empID).Scan(&row).Error == nil && row.FirstName != "" {
		return fmt.Sprintf("%s %s", row.FirstName, row.LastName)
	}
	return nil
}
