package invoice_service

import (
	"fmt"
	"time"

	invModel "sighthub-backend/internal/models/inventory"
	"sighthub-backend/internal/models/invoices"
	"sighthub-backend/internal/models/location"
	"sighthub-backend/internal/models/vendors"
)

// ─── Create Vendor Invoice ─────────────────────────────────────────────────────

type CreateVendorInvoiceRequest struct {
	InvoiceID        int64    `json:"invoice_id"`
	VendorID         int64    `json:"vendor_id"`
	InvoiceNo        string   `json:"invoice_no"`
	InvoiceDate      *string  `json:"invoice_date"`
	Quantity         int      `json:"quantity"`
	SubTotal         float64  `json:"sub_total"`
	ShippingHandling float64  `json:"shipping_handling"`
	Tax              float64  `json:"tax"`
	InvoiceTotal     float64  `json:"invoice_total"`
	OrderRef         string   `json:"order_ref"`
	DiscountReceived float64  `json:"discount_received"`
	Notes            *string  `json:"notes"`
}

type VendorInvoiceData struct {
	IDVendorInvoice  int64   `json:"id_vendor_invoice"`
	InvoiceNo        string  `json:"invoice_no"`
	InvoiceDate      *string `json:"invoice_date"`
	Quantity         int     `json:"quantity"`
	SubTotal         string  `json:"sub_total"`
	ShippingHandling string  `json:"shipping_handling"`
	Tax              string  `json:"tax"`
	InvoiceTotal     string  `json:"invoice_total"`
	OrderRef         string  `json:"order_ref"`
	Note             *string `json:"note"`
	DiscountReceived string  `json:"discount_received"`
	Vendor           map[string]interface{} `json:"vendor"`
	RelatedInvoice   map[string]interface{} `json:"related_invoice"`
	Shipment         *ShipmentData          `json:"shipment"`
}

type ShipmentData struct {
	IDShipment   int64   `json:"id_shipment"`
	QtyOk        int     `json:"qty_ok"`
	QtyHold      int     `json:"qty_hold"`
	QtyShort     int     `json:"qty_short"`
	QtyOver      int     `json:"qty_over"`
	Cost         string  `json:"cost"`
	DateReceived *string `json:"date_received"`
	Status       *string `json:"status"`
	EmployeeName string  `json:"employee_name"`
	Notes        *string `json:"notes"`
}

func (s *Service) CreateVendorInvoice(el *EmpLocation, req CreateVendorInvoiceRequest) (map[string]interface{}, error) {
	if req.InvoiceID == 0 {
		return nil, fmt.Errorf("%w: invoice_id is required", ErrBadRequest)
	}
	if req.VendorID == 0 {
		return nil, fmt.Errorf("%w: vendor_id is required", ErrBadRequest)
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, req.InvoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: invoice not found", ErrNotFound)
	}

	var v vendors.Vendor
	if err := s.db.First(&v, req.VendorID).Error; err != nil {
		return nil, fmt.Errorf("%w: vendor not found", ErrNotFound)
	}

	invoiceNo := req.InvoiceNo
	if invoiceNo == "" {
		invoiceNo = "Temp"
	}
	orderRef := req.OrderRef
	if orderRef == "" {
		orderRef = "Temporary"
	}

	vi := invModel.VendorInvoice{
		InvoiceNo:        invoiceNo,
		Quantity:         req.Quantity,
		SubTotal:         &req.SubTotal,
		ShippingHandling: &req.ShippingHandling,
		Tax:              &req.Tax,
		InvoiceTotal:     &req.InvoiceTotal,
		OrderRef:         orderRef,
		VendorID:         req.VendorID,
		InvoiceID:        req.InvoiceID,
	}
	if req.Notes != nil {
		vi.Note = req.Notes
	}

	if req.InvoiceDate != nil {
		t, err := time.Parse("2006-01-02", *req.InvoiceDate)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid date format", ErrBadRequest)
		}
		vi.InvoiceDate = &t
	}

	if err := s.db.Create(&vi).Error; err != nil {
		return nil, err
	}

	// Tally qty by status from receipts items
	qtyOk, qtyHold, qtyShort, qtyOver := s.tallyReceiptQty(req.InvoiceID)

	empID := int64(el.Employee.IDEmployee)
	status := "Received"
	shipment := invModel.Shipment{
		VendorID:          req.VendorID,
		LocationID:        int64(el.Location.IDLocation),
		// brand_id — Python uses invoice.brand_id but Invoice model doesn't have BrandID
		QtyOk:             qtyOk,
		QtyHold:           qtyHold,
		QtyShort:          qtyShort,
		QtyOver:           qtyOver,
		Cost:              0,
		EmployeeIDPrepBy:  empID,
		EmployeeIDCreated: empID,
		DateReceived:      inv.DateCreate,
		Status:            &status,
		Notes:             req.Notes,
		VendorInvoiceID:   &vi.IDVendorInvoice,
	}
	s.db.Create(&shipment)

	empName := fmt.Sprintf("%s %s", el.Employee.FirstName, el.Employee.LastName)
	return map[string]interface{}{
		"message":          "Vendor invoice and Shipment successfully created",
		"vendor_invoice":   vi.ToMap(),
		"shipment":         shipment.ToMap(),
		"employee_name":    empName,
	}, nil
}

// ─── Update Vendor Invoice ─────────────────────────────────────────────────────

type UpdateVendorInvoiceRequest struct {
	InvoiceID        *int64  `json:"invoice_id"`
	VendorID         *int64  `json:"vendor_id"`
	InvoiceNo        *string `json:"invoice_no"`
	InvoiceDate      *string `json:"invoice_date"`
	Quantity         *int    `json:"quantity"`
	SubTotal         *float64 `json:"sub_total"`
	ShippingHandling *float64 `json:"shipping_handling"`
	Tax              *float64 `json:"tax"`
	InvoiceTotal     *float64 `json:"invoice_total"`
	OrderRef         *string  `json:"order_ref"`
	DiscountReceived *float64 `json:"discount_received"`
	Note             *string  `json:"note"`
	QtyHold          *int     `json:"qty_hold"`
}

func (s *Service) UpdateVendorInvoice(el *EmpLocation, vendorInvoiceID int64, req UpdateVendorInvoiceRequest) (*VendorInvoiceData, error) {
	var vi invModel.VendorInvoice
	if err := s.db.First(&vi, vendorInvoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: vendor invoice not found", ErrNotFound)
	}

	invoiceID := vi.InvoiceID
	if req.InvoiceID != nil {
		invoiceID = *req.InvoiceID
	}
	vendorID := vi.VendorID
	if req.VendorID != nil {
		vendorID = *req.VendorID
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: invoice not found", ErrNotFound)
	}
	var v vendors.Vendor
	if err := s.db.First(&v, vendorID).Error; err != nil {
		return nil, fmt.Errorf("%w: vendor not found", ErrNotFound)
	}

	if req.InvoiceDate != nil {
		t, err := time.Parse("2006-01-02", *req.InvoiceDate)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid date format", ErrBadRequest)
		}
		vi.InvoiceDate = &t
	}

	if req.InvoiceNo != nil {
		vi.InvoiceNo = *req.InvoiceNo
	}
	if req.Quantity != nil {
		vi.Quantity = *req.Quantity
	}
	if req.SubTotal != nil {
		vi.SubTotal = req.SubTotal
	}
	if req.ShippingHandling != nil {
		vi.ShippingHandling = req.ShippingHandling
	}
	if req.Tax != nil {
		vi.Tax = req.Tax
	}
	if req.InvoiceTotal != nil {
		vi.InvoiceTotal = req.InvoiceTotal
	}
	if req.OrderRef != nil {
		vi.OrderRef = *req.OrderRef
	}
	if req.Note != nil {
		vi.Note = req.Note
	}
	vi.InvoiceID = invoiceID
	vi.VendorID = vendorID

	qtyOk, qtyHold, qtyShort, qtyOver := s.tallyReceiptQty(invoiceID)

	// Override qty_hold from request
	if req.QtyHold != nil {
		qtyHold = *req.QtyHold
	}

	// Check qty_over
	var receiptCount int64
	s.db.Model(&invModel.ReceiptsItems{}).Where("invoice_id = ?", invoiceID).Count(&receiptCount)
	if int(receiptCount) > vi.Quantity {
		qtyOver = int(receiptCount) - vi.Quantity
	}

	s.db.Save(&vi)

	empID := int64(el.Employee.IDEmployee)
	var shipment invModel.Shipment
	if err := s.db.Where("vendor_invoice_id = ?", vendorInvoiceID).First(&shipment).Error; err != nil {
		// Create new shipment
		cost := inv.FinalAmount
		dateRcv := inv.DateCreate
		status := "Received"
		shipment = invModel.Shipment{
			VendorID:          vi.VendorID,
			LocationID:        inv.LocationID,
			QtyOk:             qtyOk,
			QtyHold:           qtyHold,
			QtyShort:          qtyShort,
			QtyOver:           qtyOver,
			Cost:              cost,
			EmployeeIDPrepBy:  empID,
			EmployeeIDCreated: empID,
			DateReceived:      dateRcv,
			Status:            &status,
			VendorInvoiceID:   &vi.IDVendorInvoice,
		}
		s.db.Create(&shipment)
	} else {
		shipment.QtyOk = qtyOk
		shipment.QtyHold = qtyHold
		shipment.QtyShort = qtyShort
		shipment.QtyOver = qtyOver
		shipment.Cost = inv.FinalAmount
		shipment.DateReceived = inv.DateCreate
		shipment.VendorID = vi.VendorID
		shipment.LocationID = inv.LocationID
		s.db.Save(&shipment)
	}

	empName := fmt.Sprintf("%s %s", el.Employee.FirstName, el.Employee.LastName)

	var invDate *string
	if vi.InvoiceDate != nil {
		d := vi.InvoiceDate.Format("2006-01-02")
		invDate = &d
	}
	d := shipment.DateReceived.Format(time.RFC3339)
	result := &VendorInvoiceData{
		IDVendorInvoice:  vi.IDVendorInvoice,
		InvoiceNo:        vi.InvoiceNo,
		InvoiceDate:      invDate,
		Quantity:         vi.Quantity,
		SubTotal:         fmtFloatPtr(vi.SubTotal),
		ShippingHandling: fmtFloatPtr(vi.ShippingHandling),
		Tax:              fmtFloatPtr(vi.Tax),
		InvoiceTotal:     fmtFloatPtr(vi.InvoiceTotal),
		OrderRef:         vi.OrderRef,
		Note:             vi.Note,
		DiscountReceived: "0.00",
		Vendor:           map[string]interface{}{"vendor_name": v.VendorName, "vendor_id": v.IDVendor},
		RelatedInvoice: map[string]interface{}{
			"number_invoice": inv.NumberInvoice,
			"date_create":    inv.DateCreate.Format(time.RFC3339),
			"total_amount":   fmtFloat(inv.TotalAmount),
			"final_amount":   fmtFloat(inv.FinalAmount),
		},
		Shipment: &ShipmentData{
			IDShipment:   shipment.IDShipment,
			QtyOk:        qtyOk,
			QtyHold:      qtyHold,
			QtyShort:     qtyShort,
			QtyOver:      qtyOver,
			Cost:         fmtFloat(shipment.Cost),
			DateReceived: &d,
			Status:       shipment.Status,
			EmployeeName: empName,
			Notes:        shipment.Notes,
		},
	}
	return result, nil
}

// ─── Get Vendor Invoice ────────────────────────────────────────────────────────

func (s *Service) GetVendorInvoice(el *EmpLocation, vendorInvoiceID int64) (*VendorInvoiceData, error) {
	var vi invModel.VendorInvoice
	if err := s.db.First(&vi, vendorInvoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: vendor invoice not found", ErrNotFound)
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, vi.InvoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: related invoice not found", ErrNotFound)
	}
	var v vendors.Vendor
	if err := s.db.First(&v, vi.VendorID).Error; err != nil {
		return nil, fmt.Errorf("%w: vendor not found", ErrNotFound)
	}

	var shipment *invModel.Shipment
	var sh invModel.Shipment
	if err := s.db.Where("vendor_invoice_id = ?", vendorInvoiceID).First(&sh).Error; err == nil {
		shipment = &sh
	}

	empName := fmt.Sprintf("%s %s", el.Employee.FirstName, el.Employee.LastName)

	var invDate *string
	if vi.InvoiceDate != nil {
		d := vi.InvoiceDate.Format("2006-01-02")
		invDate = &d
	}

	result := &VendorInvoiceData{
		IDVendorInvoice:  vi.IDVendorInvoice,
		InvoiceNo:        vi.InvoiceNo,
		InvoiceDate:      invDate,
		Quantity:         vi.Quantity,
		SubTotal:         fmtFloatPtr(vi.SubTotal),
		ShippingHandling: fmtFloatPtr(vi.ShippingHandling),
		Tax:              fmtFloatPtr(vi.Tax),
		InvoiceTotal:     fmtFloatPtr(vi.InvoiceTotal),
		OrderRef:         vi.OrderRef,
		Note:             vi.Note,
		DiscountReceived: "0.00",
		Vendor:           map[string]interface{}{"vendor_name": v.VendorName, "vendor_id": v.IDVendor},
		RelatedInvoice: map[string]interface{}{
			"number_invoice": inv.NumberInvoice,
			"date_create":    inv.DateCreate.Format(time.RFC3339),
			"total_amount":   fmtFloat(inv.TotalAmount),
			"final_amount":   fmtFloat(inv.FinalAmount),
		},
	}

	if shipment != nil {
		d := shipment.DateReceived.Format(time.RFC3339)
		result.Shipment = &ShipmentData{
			IDShipment:   shipment.IDShipment,
			QtyOk:        shipment.QtyOk,
			QtyHold:      shipment.QtyHold,
			QtyShort:     shipment.QtyShort,
			QtyOver:      shipment.QtyOver,
			Cost:         fmtFloat(shipment.Cost),
			DateReceived: &d,
			Status:       shipment.Status,
			EmployeeName: empName,
			Notes:        shipment.Notes,
		}
	}
	return result, nil
}

// ─── Vendor Contacts ──────────────────────────────────────────────────────────

func (s *Service) GetVendorContacts(vendorID int) (map[string]interface{}, error) {
	var v vendors.Vendor
	if err := s.db.First(&v, vendorID).Error; err != nil {
		return nil, fmt.Errorf("%w: vendor not found", ErrNotFound)
	}
	var rep vendors.RepContactVendor
	_ = s.db.Where("vendor_id = ?", vendorID).First(&rep)

	return map[string]interface{}{
		"website": v.Website,
		"fax":     rep.Fax,
		"phone":   rep.Phone,
		"email":   rep.Email,
		"name":    rep.Name,
	}, nil
}

// ─── Location Contacts ────────────────────────────────────────────────────────

func (s *Service) GetLocationContacts(locationID int) (map[string]interface{}, error) {
	var loc location.Location
	if err := s.db.First(&loc, locationID).Error; err != nil {
		return nil, fmt.Errorf("%w: location not found", ErrNotFound)
	}
	return map[string]interface{}{
		"phone":   loc.Phone,
		"fax":     loc.Fax,
		"email":   loc.Email,
		"website": loc.Website,
	}, nil
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// tallyReceiptQty counts inventory statuses among receipts for an invoice.
func (s *Service) tallyReceiptQty(invoiceID int64) (ok, hold, short, over int) {
	var receipts []invModel.ReceiptsItems
	s.db.Where("invoice_id = ?", invoiceID).Find(&receipts)
	for _, r := range receipts {
		var item invModel.Inventory
		if err := s.db.First(&item, r.InventoryID).Error; err != nil {
			continue
		}
		switch string(item.StatusItemsInventory) {
		case "Ready for Sale":
			ok++
		case "HOLD":
			hold++
		case "SHORT":
			short++
		case "OVER":
			over++
		}
	}
	return
}
