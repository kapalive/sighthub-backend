package invoice_service

import (
	"fmt"
	"time"

	"sighthub-backend/internal/models/frames"
	"sighthub-backend/internal/models/general"
	invModel "sighthub-backend/internal/models/inventory"
	"sighthub-backend/internal/models/service"
	"sighthub-backend/internal/models/vendors"
)

// ─── List Return Invoices ─────────────────────────────────────────────────────

type ReturnInvoiceFilter struct {
	DateFrom      *time.Time
	DateTo        *time.Time
	VendorID      *int64
	GroupByVendor bool
}

func (s *Service) GetReturnInvoices(el *EmpLocation, f ReturnInvoiceFilter) (interface{}, error) {
	now := time.Now()
	dateFrom := now.AddDate(0, 0, -30)
	dateTo := now
	if f.DateFrom != nil {
		dateFrom = *f.DateFrom
	}
	if f.DateTo != nil {
		dateTo = *f.DateTo
	}

	locID := int64(el.Location.IDLocation)

	var invoiceIDs []int64
	q := s.db.Model(&vendors.ReturnToVendorInvoice{}).
		Joins("JOIN return_to_vendor_item ON return_to_vendor_item.return_to_vendor_invoice_id = return_to_vendor_invoice.id_return_to_vendor_invoice").
		Joins("JOIN inventory ON inventory.id_inventory = return_to_vendor_item.inventory_id").
		Where("inventory.location_id = ?", locID).
		Where("return_to_vendor_invoice.created_date BETWEEN ? AND ?", dateFrom, dateTo).
		Distinct("return_to_vendor_invoice.id_return_to_vendor_invoice")

	if f.VendorID != nil {
		q = q.Where("return_to_vendor_invoice.vendor_id = ?", *f.VendorID)
	}
	if err := q.Pluck("return_to_vendor_invoice.id_return_to_vendor_invoice", &invoiceIDs).Error; err != nil {
		return nil, err
	}

	if len(invoiceIDs) == 0 {
		return []interface{}{}, nil
	}

	var rtvInvoices []vendors.ReturnToVendorInvoice
	s.db.Where("id_return_to_vendor_invoice IN ?", invoiceIDs).
		Preload("Vendor").Preload("Items").Find(&rtvInvoices)

	if f.GroupByVendor {
		grouped := make(map[int64]map[string]interface{})
		for _, inv := range rtvInvoices {
			vid := inv.VendorID
			if _, ok := grouped[vid]; !ok {
				vendorName := ""
				if inv.Vendor != nil {
					vendorName = inv.Vendor.VendorName
				}
				grouped[vid] = map[string]interface{}{
					"vendor_id":   vid,
					"vendor_name": vendorName,
					"invoices":    []map[string]interface{}{},
				}
			}
			var cd *string
			if !inv.CreatedDate.IsZero() {
				d := inv.CreatedDate.Format("2006-01-02")
				cd = &d
			}
			invList := grouped[vid]["invoices"].([]map[string]interface{})
			grouped[vid]["invoices"] = append(invList, map[string]interface{}{
				"id_return_to_vendor_invoice": inv.IDReturnToVendorInvoice,
				"created_date":               cd,
				"credit_amount":              fmtFloatPtr(inv.CreditAmount),
			})
		}
		var out []map[string]interface{}
		for _, v := range grouped {
			out = append(out, v)
		}
		if out == nil {
			out = []map[string]interface{}{}
		}
		return out, nil
	}

	var results []map[string]interface{}
	for _, inv := range rtvInvoices {
		vendorName := ""
		if inv.Vendor != nil {
			vendorName = inv.Vendor.VendorName
		}
		var cd *string
		if !inv.CreatedDate.IsZero() {
			d := inv.CreatedDate.Format(time.RFC3339)
			cd = &d
		}
		results = append(results, map[string]interface{}{
			"id_return_to_vendor_invoice": inv.IDReturnToVendorInvoice,
			"vendor_id":                  inv.VendorID,
			"quantity":                   inv.Quantity,
			"purchase_total":             fmtFloat(inv.PurchaseTotal),
			"vendor_name":                vendorName,
			"created_date":               cd,
			"credit_amount":              fmtFloatPtr(inv.CreditAmount),
		})
	}
	if results == nil {
		results = []map[string]interface{}{}
	}
	return results, nil
}

// ─── Create Return Invoice ────────────────────────────────────────────────────

type ReturnItemIn struct {
	InventoryID int64  `json:"inventory_id"`
	Reason      string `json:"reason"`
}

type CreateReturnInvoiceRequest struct {
	Items        []ReturnItemIn `json:"items"`
	CreditAmount *float64       `json:"credit_amount"`
}

func (s *Service) CreateReturnInvoice(el *EmpLocation, req CreateReturnInvoiceRequest) (map[string]interface{}, error) {
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("%w: items must be a non-empty array", ErrBadRequest)
	}

	empID := int64(el.Employee.IDEmployee)
	locID := int64(el.Location.IDLocation)

	type resolved struct {
		item         invModel.Inventory
		reason       string
		purchaseCost float64
	}

	var vendorID *int64
	var items []resolved

	for i, itemData := range req.Items {
		if itemData.InventoryID == 0 || itemData.Reason == "" {
			return nil, fmt.Errorf("%w: item %d: inventory_id and reason are required", ErrBadRequest, i)
		}
		var item invModel.Inventory
		if err := s.db.First(&item, itemData.InventoryID).Error; err != nil {
			return nil, fmt.Errorf("%w: item %d: inventory not found", ErrNotFound, i)
		}
		if item.LocationID != locID {
			return nil, fmt.Errorf("%w: item %d: inventory does not belong to your location", ErrForbidden, i)
		}
		itemVendorID, err := s.resolveVendorID(item.IDInventory)
		if err != nil {
			return nil, fmt.Errorf("%w: item %d: could not determine vendor", ErrBadRequest, i)
		}
		if vendorID == nil {
			vendorID = &itemVendorID
		} else if *vendorID != itemVendorID {
			return nil, fmt.Errorf("%w: item %d: all items must belong to the same vendor", ErrBadRequest, i)
		}

		var purchaseCost float64
		var pb invModel.PriceBook
		if err := s.db.Where("inventory_id = ?", item.IDInventory).First(&pb).Error; err == nil && pb.ItemListCost != nil {
			purchaseCost = *pb.ItemListCost
		}
		items = append(items, resolved{item: item, reason: itemData.Reason, purchaseCost: purchaseCost})
	}

	if vendorID == nil {
		return nil, fmt.Errorf("%w: could not determine vendor", ErrBadRequest)
	}

	rtv := vendors.ReturnToVendorInvoice{
		VendorID:      *vendorID,
		CreditAmount:  req.CreditAmount,
		EmployeeID:    &empID,
		PurchaseTotal: 0,
		Quantity:      0,
		Status:        "Pending",
	}
	if err := s.db.Create(&rtv).Error; err != nil {
		return nil, err
	}

	for _, r := range items {
		rtvItem := vendors.ReturnToVendorItem{
			ReturnToVendorInvoiceID: rtv.IDReturnToVendorInvoice,
			InventoryID:             r.item.IDInventory,
			ReasonReturn:            r.reason,
			PurchaseCost:            r.purchaseCost,
		}
		s.db.Create(&rtvItem)

		rtv.PurchaseTotal += r.purchaseCost
		rtv.Quantity++

		r.item.StatusItemsInventory = "On Return"
		s.db.Save(&r.item)

		txn := invModel.InventoryTransaction{
			InventoryID:     r.item.IDInventory,
			FromLocationID:  r.item.LocationID,
			TransferredBy:   empID,
			InvoiceID:       r.item.InvoiceID,
			OldInvoiceID:    &r.item.InvoiceID,
			StatusItems:     "On Return",
			TransactionType: "ReturnToVendor",
			Notes:           strPtr(fmt.Sprintf("Return reason: %s", r.reason)),
		}
		s.db.Create(&txn)
	}
	s.db.Save(&rtv)

	return map[string]interface{}{
		"id_return_to_vendor_invoice": rtv.IDReturnToVendorInvoice,
		"vendor_id":                  rtv.VendorID,
		"purchase_total":             fmtFloat(rtv.PurchaseTotal),
		"quantity":                   rtv.Quantity,
		"status":                     rtv.Status,
		"credit_amount":              fmtFloatPtr(rtv.CreditAmount),
		"created_date":               rtv.CreatedDate.Format("2006-01-02"),
	}, nil
}

// ─── Update Return Invoice ────────────────────────────────────────────────────

type UpdateReturnInvoiceRequest struct {
	CreditAmount      *float64       `json:"credit_amount"`
	Status            *string        `json:"status"`
	ARNumber          *string        `json:"ar_number"`
	ShippingServiceID *int           `json:"shipping_service_id"`
	TrackingNumber    *string        `json:"tracking_number"`
	ShippingNumber    *string        `json:"shipping_number"`
	Note              *string        `json:"note"`
	Items             []ReturnItemIn `json:"items"`
}

func (s *Service) UpdateReturnInvoice(el *EmpLocation, returnInvoiceID int64, req UpdateReturnInvoiceRequest) (map[string]interface{}, error) {
	var rtv vendors.ReturnToVendorInvoice
	if err := s.db.Preload("Items").First(&rtv, returnInvoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: return invoice not found", ErrNotFound)
	}

	if req.CreditAmount != nil {
		rtv.CreditAmount = req.CreditAmount
	}
	if req.Status != nil {
		allowed := map[string]bool{"Pending": true, "Approved": true, "Completed": true}
		if !allowed[*req.Status] {
			return nil, fmt.Errorf("%w: status must be Pending, Approved, or Completed", ErrBadRequest)
		}
		rtv.Status = *req.Status
	}
	if req.ARNumber != nil {
		rtv.ARNumber = req.ARNumber
	}
	if req.ShippingServiceID != nil {
		rtv.ShippingServiceID = req.ShippingServiceID
	}
	if req.TrackingNumber != nil {
		rtv.TrackingNumber = req.TrackingNumber
	}
	if req.ShippingNumber != nil {
		rtv.ShippingNumber = req.ShippingNumber
	}
	if req.Note != nil {
		rtv.Note = req.Note
	}

	var defaultReason string
	if len(rtv.Items) > 0 {
		defaultReason = rtv.Items[0].ReasonReturn
	}

	empID := int64(el.Employee.IDEmployee)

	for _, itemData := range req.Items {
		reason := itemData.Reason
		if reason == "" {
			reason = defaultReason
		}
		if reason == "" {
			return nil, fmt.Errorf("%w: missing reason for inventory %d", ErrBadRequest, itemData.InventoryID)
		}

		var item invModel.Inventory
		if err := s.db.First(&item, itemData.InventoryID).Error; err != nil {
			return nil, fmt.Errorf("%w: inventory %d not found", ErrNotFound, itemData.InventoryID)
		}

		var existing vendors.ReturnToVendorItem
		if err := s.db.Where("return_to_vendor_invoice_id = ? AND inventory_id = ?",
			rtv.IDReturnToVendorInvoice, itemData.InventoryID).First(&existing).Error; err == nil {
			continue
		}

		var purchaseCost float64
		var pb invModel.PriceBook
		if err := s.db.Where("inventory_id = ?", item.IDInventory).First(&pb).Error; err == nil && pb.ItemListCost != nil {
			purchaseCost = *pb.ItemListCost
		}

		rtvItem := vendors.ReturnToVendorItem{
			ReturnToVendorInvoiceID: rtv.IDReturnToVendorInvoice,
			InventoryID:             item.IDInventory,
			ReasonReturn:            reason,
			PurchaseCost:            purchaseCost,
		}
		s.db.Create(&rtvItem)

		rtv.PurchaseTotal += purchaseCost
		rtv.Quantity++

		item.StatusItemsInventory = "On Return"
		s.db.Save(&item)

		txn := invModel.InventoryTransaction{
			InventoryID:     item.IDInventory,
			FromLocationID:  item.LocationID,
			TransferredBy:   empID,
			InvoiceID:       item.InvoiceID,
			OldInvoiceID:    &item.InvoiceID,
			StatusItems:     "On Return",
			TransactionType: "ReturnToVendor",
			Notes:           strPtr(fmt.Sprintf("Added item to existing invoice %d; reason: %s", returnInvoiceID, reason)),
		}
		s.db.Create(&txn)
	}

	s.db.Save(&rtv)
	return s.buildReturnInvoiceMap(&rtv), nil
}

// ─── Get Return Invoice ────────────────────────────────────────────────────────

func (s *Service) GetReturnInvoice(returnInvoiceID int64) (map[string]interface{}, error) {
	var rtv vendors.ReturnToVendorInvoice
	if err := s.db.Preload("Vendor").Preload("Items").Preload("ShippingService").
		Preload("Employee").First(&rtv, returnInvoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: return invoice not found", ErrNotFound)
	}
	return s.buildReturnInvoiceMap(&rtv), nil
}

// ─── Delete Return Invoice ────────────────────────────────────────────────────

func (s *Service) DeleteReturnInvoice(returnInvoiceID int64) error {
	var rtv vendors.ReturnToVendorInvoice
	if err := s.db.Preload("Items").First(&rtv, returnInvoiceID).Error; err != nil {
		return fmt.Errorf("%w: return invoice not found", ErrNotFound)
	}

	for _, rtvItem := range rtv.Items {
		var item invModel.Inventory
		if err := s.db.First(&item, rtvItem.InventoryID).Error; err == nil {
			oldStatus := string(item.StatusItemsInventory)
			item.StatusItemsInventory = "Ready for Sale"
			s.db.Save(&item)

			txn := invModel.InventoryTransaction{
				InventoryID:     item.IDInventory,
				ToLocationID:    item.LocationID,
				TransferredBy:   1,
				InvoiceID:       item.InvoiceID,
				OldInvoiceID:    &item.InvoiceID,
				StatusItems:     "Ready for Sale",
				TransactionType: "UndoReturnToVendor",
				Notes:           strPtr(fmt.Sprintf("Undo vendor return invoice=%d. Old status was %s", returnInvoiceID, oldStatus)),
			}
			s.db.Create(&txn)
		}
		s.db.Delete(&rtvItem)
	}
	s.db.Delete(&rtv)
	return nil
}

// ─── Shipping Services ────────────────────────────────────────────────────────

func (s *Service) GetShippingServices() ([]map[string]interface{}, error) {
	var svcs []service.ShippingServices
	s.db.Order("name_company").Find(&svcs)
	var result []map[string]interface{}
	for _, sv := range svcs {
		result = append(result, map[string]interface{}{
			"shipping_service_id":   sv.IDShippingServices,
			"shipping_service_name": sv.NameCompany,
		})
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, nil
}

// ─── Payment Methods ──────────────────────────────────────────────────────────

func (s *Service) GetReturnPaymentMethods() ([]map[string]interface{}, error) {
	allowedIDs := []int{1, 2, 3, 4, 5, 6, 22}
	var methods []general.PaymentMethod
	s.db.Where("id_payment_method IN ?", allowedIDs).Order("method_name").Find(&methods)
	var result []map[string]interface{}
	for _, m := range methods {
		result = append(result, map[string]interface{}{
			"id_payment_method": m.IDPaymentMethod,
			"method_name":       m.MethodName,
			"short_name":        m.ShortName,
		})
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, nil
}

// ─── Add Payment ──────────────────────────────────────────────────────────────

type AddReturnPaymentRequest struct {
	Amount          float64 `json:"amount"`
	PaymentMethodID int     `json:"payment_method_id"`
	Notes           *string `json:"notes"`
}

func (s *Service) AddReturnPayment(el *EmpLocation, returnInvoiceID int64, req AddReturnPaymentRequest) (map[string]interface{}, error) {
	var rtv vendors.ReturnToVendorInvoice
	if err := s.db.First(&rtv, returnInvoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: return invoice not found", ErrNotFound)
	}
	if req.PaymentMethodID == 0 {
		return nil, fmt.Errorf("%w: payment_method_id is required", ErrBadRequest)
	}
	var pm general.PaymentMethod
	if err := s.db.First(&pm, req.PaymentMethodID).Error; err != nil {
		return nil, fmt.Errorf("%w: payment method not found", ErrNotFound)
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("%w: amount must be > 0", ErrBadRequest)
	}

	empID := int64(el.Employee.IDEmployee)
	payment := vendors.VendorReturnPayment{
		ReturnToVendorInvoiceID: returnInvoiceID,
		Amount:                  req.Amount,
		PaymentMethodID:         req.PaymentMethodID,
		EmployeeID:              &empID,
		Notes:                   req.Notes,
	}
	if err := s.db.Create(&payment).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": "Payment recorded",
		"payment": map[string]interface{}{
			"id_vendor_return_payment":    payment.IDVendorReturnPayment,
			"return_to_vendor_invoice_id": payment.ReturnToVendorInvoiceID,
			"amount":                     fmtFloat(payment.Amount),
			"payment_method_id":          payment.PaymentMethodID,
			"payment_timestamp":          payment.PaymentTimestamp.Format(time.RFC3339),
			"notes":                      payment.Notes,
		},
		"balance_due": fmtFloat(s.calcBalanceDue(&rtv)),
	}, nil
}

// ─── Get Payments ─────────────────────────────────────────────────────────────

func (s *Service) GetReturnPayments(returnInvoiceID int64) (map[string]interface{}, error) {
	var rtv vendors.ReturnToVendorInvoice
	if err := s.db.Preload("Payments.PaymentMethod").First(&rtv, returnInvoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: return invoice not found", ErrNotFound)
	}

	var payments []map[string]interface{}
	for _, p := range rtv.Payments {
		methodName := ""
		if p.PaymentMethod != nil {
			methodName = p.PaymentMethod.MethodName
		}
		payments = append(payments, map[string]interface{}{
			"id_vendor_return_payment":    p.IDVendorReturnPayment,
			"return_to_vendor_invoice_id": p.ReturnToVendorInvoiceID,
			"amount":                     fmtFloat(p.Amount),
			"payment_method_id":          p.PaymentMethodID,
			"payment_method":             methodName,
			"payment_timestamp":          p.PaymentTimestamp.Format(time.RFC3339),
			"notes":                      p.Notes,
		})
	}
	if payments == nil {
		payments = []map[string]interface{}{}
	}
	return map[string]interface{}{
		"payments":    payments,
		"balance_due": fmtFloat(s.calcBalanceDue(&rtv)),
	}, nil
}

// ─── Delete Payment ───────────────────────────────────────────────────────────

func (s *Service) DeleteReturnPayment(returnInvoiceID, paymentID int64) (map[string]interface{}, error) {
	var payment vendors.VendorReturnPayment
	if err := s.db.Where("id_vendor_return_payment = ? AND return_to_vendor_invoice_id = ?",
		paymentID, returnInvoiceID).First(&payment).Error; err != nil {
		return nil, fmt.Errorf("%w: payment not found", ErrNotFound)
	}
	if err := s.db.Delete(&payment).Error; err != nil {
		return nil, err
	}

	var rtv vendors.ReturnToVendorInvoice
	s.db.First(&rtv, returnInvoiceID)
	return map[string]interface{}{
		"message":     "Payment deleted",
		"balance_due": fmtFloat(s.calcBalanceDue(&rtv)),
	}, nil
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

func (s *Service) resolveVendorID(inventoryID int64) (int64, error) {
	var vendorID int64
	err := s.db.Raw(`
		SELECT p.vendor_id FROM inventory i
		JOIN model m ON m.id_model = i.model_id
		JOIN product p ON p.id_product = m.product_id
		WHERE i.id_inventory = ?
		LIMIT 1`, inventoryID).Scan(&vendorID).Error
	if err != nil || vendorID == 0 {
		return 0, fmt.Errorf("vendor not found")
	}
	return vendorID, nil
}

func (s *Service) calcBalanceDue(rtv *vendors.ReturnToVendorInvoice) float64 {
	var total float64
	s.db.Model(&vendors.VendorReturnPayment{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("return_to_vendor_invoice_id = ?", rtv.IDReturnToVendorInvoice).
		Scan(&total)
	credit := 0.0
	if rtv.CreditAmount != nil {
		credit = *rtv.CreditAmount
	}
	return credit - total
}

func (s *Service) buildReturnInvoiceMap(rtv *vendors.ReturnToVendorInvoice) map[string]interface{} {
	vendorName := ""
	if rtv.Vendor != nil {
		vendorName = rtv.Vendor.VendorName
	}
	var employeeName string
	if rtv.Employee != nil {
		employeeName = fmt.Sprintf("%s %s", rtv.Employee.FirstName, rtv.Employee.LastName)
	}
	var createdDate *string
	if !rtv.CreatedDate.IsZero() {
		d := rtv.CreatedDate.Format("2006-01-02")
		createdDate = &d
	}
	var shippingServiceName *string
	if rtv.ShippingService != nil && rtv.ShippingService.ShortName != nil {
		shippingServiceName = rtv.ShippingService.ShortName
	}

	return map[string]interface{}{
		"id_return_to_vendor_invoice": rtv.IDReturnToVendorInvoice,
		"vendor_id":                  rtv.VendorID,
		"vendor_name":                vendorName,
		"employee_id":                rtv.EmployeeID,
		"employee_name":              employeeName,
		"created_date":               createdDate,
		"status":                     rtv.Status,
		"credit_amount":              fmtFloatPtr(rtv.CreditAmount),
		"purchase_total":             fmtFloat(rtv.PurchaseTotal),
		"quantity":                   rtv.Quantity,
		"ar_number":                  rtv.ARNumber,
		"shipping_service_id":        rtv.ShippingServiceID,
		"shipping_service_name":      shippingServiceName,
		"tracking_number":            rtv.TrackingNumber,
		"shipping_number":            rtv.ShippingNumber,
		"note":                       rtv.Note,
		"items_count":                len(rtv.Items),
		"items":                      s.buildReturnItems(rtv.Items),
	}
}

func (s *Service) buildReturnItems(items []vendors.ReturnToVendorItem) []map[string]interface{} {
	var result []map[string]interface{}
	for _, item := range items {
		row := map[string]interface{}{
			"id_return_to_vendor_item": item.IDReturnToVendorItem,
			"inventory_id":            item.InventoryID,
			"reason_return":           item.ReasonReturn,
			"purchase_cost":           fmtFloat(item.PurchaseCost),
		}

		var inv invModel.Inventory
		if err := s.db.First(&inv, item.InventoryID).Error; err == nil {
			invData := map[string]interface{}{
				"sku":                    inv.SKU,
				"status_items_inventory": string(inv.StatusItemsInventory),
				"location_id":            inv.LocationID,
			}
			if inv.ModelID != nil {
				var m frames.Model
				if err := s.db.First(&m, *inv.ModelID).Error; err == nil {
					invData["model_name"] = m.TitleVariant
					var prod frames.Product
					if err := s.db.First(&prod, m.ProductID).Error; err == nil {
						invData["product_name"] = prod.TitleProduct
						if prod.BrandID != nil {
							var brand vendors.Brand
							if err := s.db.First(&brand, *prod.BrandID).Error; err == nil {
								invData["brand_name"] = brand.BrandName
							}
						}
					}
				}
			}
			row["inventory"] = invData
		}
		result = append(result, row)
	}
	if result == nil {
		result = []map[string]interface{}{}
	}
	return result
}
