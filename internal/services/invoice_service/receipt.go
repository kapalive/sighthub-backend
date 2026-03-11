package invoice_service

import (
	"fmt"
	"time"

	invModel "sighthub-backend/internal/models/inventory"
	"sighthub-backend/internal/models/invoices"
	"sighthub-backend/internal/models/location"
	"sighthub-backend/internal/models/patients"
	"sighthub-backend/internal/models/vendors"
)

// ─── Receipt List ─────────────────────────────────────────────────────────────

type ReceiptFilter struct {
	InvoiceType    string // "V" or "I"
	FromLocationID *int64
	VendorID       *int64
	DateFrom       *time.Time
	DateTo         *time.Time
}

type ReceiptInvoiceRow struct {
	IDInvoice       int64  `json:"id_invoice"`
	CreatedDate     string `json:"created_date"`
	NumberInvoice   string `json:"number_invoice"`
	TotalAmount     string `json:"total_amount"`
	FinalAmount     string `json:"final_amount"`
	Due             string `json:"due"`
	Quantity        int    `json:"quantity"`
	VendorName      string `json:"vendor_name"`
	VendorInvoiceID *int64 `json:"vendor_invoice_id"`
}

// GetReceipts returns invoices grouped by vendor or from-location.
func (s *Service) GetReceipts(el *EmpLocation, f ReceiptFilter) (map[string][]ReceiptInvoiceRow, error) {
	if f.InvoiceType != "V" && f.InvoiceType != "I" {
		return nil, fmt.Errorf("%w: invoice_type must be V or I", ErrBadRequest)
	}

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

	q := s.db.Model(&invoices.Invoice{}).
		Where("date_create >= ? AND date_create <= ?", dateFrom, dateTo)

	switch f.InvoiceType {
	case "V":
		q = q.Where("number_invoice LIKE 'V%' AND location_id = ?", locID)
		if f.VendorID != nil {
			q = q.Where("vendor_id = ?", *f.VendorID)
		}
	case "I":
		q = q.Where("number_invoice LIKE 'I%' AND to_location_id = ?", locID)
		warehouseID := el.Location.WarehouseID
		if warehouseID != nil {
			q = q.Where(
				"location_id != ? OR (location_id = ? AND to_location_id = ? AND location_id != ?)",
				locID, locID, locID, int64(*warehouseID),
			)
		} else {
			q = q.Where("location_id != ?", locID)
		}
		if f.FromLocationID != nil {
			q = q.Where("location_id = ?", *f.FromLocationID)
		}
	}

	var rows []invoices.Invoice
	if err := q.Order("date_create DESC").Find(&rows).Error; err != nil {
		return nil, err
	}

	grouped := make(map[string][]ReceiptInvoiceRow)

	for _, inv := range rows {
		var groupName string
		var vendorInvoiceID *int64

		if f.InvoiceType == "V" {
			var v vendors.Vendor
			if err := s.db.First(&v, inv.VendorID).Error; err == nil {
				groupName = v.VendorName
			} else {
				groupName = "Unknown Vendor"
			}
			var vi invModel.VendorInvoice
			if err := s.db.Where("invoice_id = ?", inv.IDInvoice).First(&vi).Error; err == nil {
				id := vi.IDVendorInvoice
				vendorInvoiceID = &id
			}
		} else {
			var loc location.Location
			if err := s.db.First(&loc, inv.LocationID).Error; err == nil {
				groupName = loc.FullName
			} else {
				groupName = "Unknown Location"
			}
		}

		row := ReceiptInvoiceRow{
			IDInvoice:       inv.IDInvoice,
			CreatedDate:     inv.DateCreate.UTC().Format(time.RFC3339),
			NumberInvoice:   inv.NumberInvoice,
			TotalAmount:     fmtFloat(inv.TotalAmount),
			FinalAmount:     fmtFloat(inv.FinalAmount),
			Due:             fmtFloat(inv.Due),
			Quantity:        inv.Quantity,
			VendorName:      groupName,
			VendorInvoiceID: vendorInvoiceID,
		}
		grouped[groupName] = append(grouped[groupName], row)
	}

	return grouped, nil
}

// ─── Receipt Items ────────────────────────────────────────────────────────────

type ReceiptItemRow struct {
	SKU             string  `json:"sku"`
	ProductTitle    *string `json:"product_title"`
	VariantTitle    *string `json:"variant_title"`
	BrandName       *string `json:"brand_name"`
	VendorName      *string `json:"vendor_name,omitempty"`
	Status          string  `json:"status"`
	ItemNet         *string `json:"item_net,omitempty"`
	DateTransaction *string `json:"date_transaction,omitempty"`
	DateReceipt     *string `json:"date_receipt,omitempty"`
}

type GetReceiptResult struct {
	InvoiceID     int64            `json:"invoice_id"`
	NumberInvoice string           `json:"number_invoice"`
	TotalAmount   string           `json:"total_amount"`
	Due           string           `json:"due"`
	Items         []ReceiptItemRow `json:"items"`
}

// GetReceipt returns items for a given invoice.
func (s *Service) GetReceipt(invoiceID int64) (*GetReceiptResult, error) {
	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, ErrNotFound
	}

	var items []ReceiptItemRow

	if len(inv.NumberInvoice) > 0 && inv.NumberInvoice[0] == 'V' {
		var receipts []invModel.ReceiptsItems
		s.db.Where("invoice_id = ?", invoiceID).Find(&receipts)
		for _, r := range receipts {
			if row := s.buildItemRow(r.InventoryID, "V", r.DateTime); row != nil {
				items = append(items, *row)
			}
		}
	} else if len(inv.NumberInvoice) > 0 && inv.NumberInvoice[0] == 'I' {
		var txns []invModel.InventoryTransaction
		s.db.Where("invoice_id = ?", invoiceID).Find(&txns)
		for _, t := range txns {
			if row := s.buildItemRow(t.InventoryID, "I", t.DateTransaction); row != nil {
				items = append(items, *row)
			}
		}
	} else {
		return nil, fmt.Errorf("%w: unknown invoice type", ErrBadRequest)
	}

	if items == nil {
		items = []ReceiptItemRow{}
	}
	return &GetReceiptResult{
		InvoiceID:     invoiceID,
		NumberInvoice: inv.NumberInvoice,
		TotalAmount:   fmtFloat(inv.TotalAmount),
		Due:           fmtFloat(inv.Due),
		Items:         items,
	}, nil
}

func (s *Service) buildItemRow(invID int64, iType string, ts time.Time) *ReceiptItemRow {
	var item invModel.Inventory
	if err := s.db.First(&item, invID).Error; err != nil {
		return nil
	}
	tsStr := ts.Format(time.RFC3339)
	row := &ReceiptItemRow{
		SKU:    item.SKU,
		Status: string(item.StatusItemsInventory),
	}
	if iType == "V" {
		row.DateReceipt = &tsStr
	} else {
		row.DateTransaction = &tsStr
		var pb invModel.PriceBook
		if err := s.db.Where("inventory_id = ?", invID).First(&pb).Error; err == nil && pb.ItemNet != nil {
			v := fmtFloat(*pb.ItemNet)
			row.ItemNet = &v
		}
	}
	s.enrichItemMeta(row, item.ModelID)
	return row
}

// ─── Confirm Receipt ──────────────────────────────────────────────────────────

type ConfirmReceiptRequest struct {
	InvoiceID   int64  `json:"invoice_id"`
	SKU         string `json:"sku"`
	InventoryID *int64 `json:"inventory_id"`
}

type ConfirmReceiptResult struct {
	Message string           `json:"message"`
	Items   []ReceiptItemRow `json:"items"`
}

func (s *Service) ConfirmReceipt(el *EmpLocation, req ConfirmReceiptRequest) (*ConfirmReceiptResult, error) {
	if req.InvoiceID == 0 {
		return nil, fmt.Errorf("%w: invoice_id required", ErrBadRequest)
	}
	if req.SKU == "" && req.InventoryID == nil {
		return nil, fmt.Errorf("%w: sku or inventory_id required", ErrBadRequest)
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, req.InvoiceID).Error; err != nil {
		return nil, ErrNotFound
	}

	var item invModel.Inventory
	if req.SKU != "" {
		if err := s.db.Where("sku = ?", req.SKU).First(&item).Error; err != nil {
			return nil, ErrNotFound
		}
	} else {
		if err := s.db.First(&item, *req.InventoryID).Error; err != nil {
			return nil, ErrNotFound
		}
	}

	status := string(item.StatusItemsInventory)
	if status != "ICT (sent and not received)" && status != "Vendor Shipment" {
		return nil, fmt.Errorf("%w: item is not in a transferable state", ErrBadRequest)
	}

	var txn invModel.InventoryTransaction
	if err := s.db.Where("inventory_id = ? AND to_location_id = ?",
		item.IDInventory, int64(el.Location.IDLocation)).
		Order("date_transaction DESC").First(&txn).Error; err != nil {
		return nil, fmt.Errorf("%w: no transfer record found for this item", ErrNotFound)
	}

	if txn.ToLocationID != int64(el.Location.IDLocation) {
		return nil, fmt.Errorf("%w: item is not meant to be received by the current location", ErrForbidden)
	}

	now := time.Now()
	locID := int64(el.Location.IDLocation)
	empID := int64(el.Employee.IDEmployee)

	item.StatusItemsInventory = "Ready for Sale"
	item.LocationID = locID

	if len(inv.NumberInvoice) > 0 && inv.NumberInvoice[0] == 'I' {
		txn.StatusItems = "Ready for Sale"
		txn.ToLocationID = locID
		s.db.Save(&txn)

		newTxn := invModel.InventoryTransaction{
			InventoryID:     item.IDInventory,
			FromLocationID:  txn.FromLocationID,
			ToLocationID:    locID,
			TransferredBy:   txn.TransferredBy,
			InvoiceID:       req.InvoiceID,
			StatusItems:     "Ready for Sale",
			TransactionType: "Received from location",
			DateTransaction: now,
		}
		s.db.Create(&newTxn)

		transfer := invModel.InventoryTransfer{
			InventoryID:    item.IDInventory,
			FromLocationID: txn.FromLocationID,
			ToLocationID:   locID,
			TransferredBy:  txn.TransferredBy,
			ReceivedBy:     empID,
			StatusItems:    "Ready for Sale",
			InvoiceID:      txn.InvoiceID,
			InvoiceFrom:    &txn.InvoiceID,
			InvoiceTo:      &req.InvoiceID,
		}
		s.db.Create(&transfer)

	} else if len(inv.NumberInvoice) > 0 && inv.NumberInvoice[0] == 'V' {
		var vi invModel.VendorInvoice
		if err := s.db.Where("invoice_id = ?", item.InvoiceID).First(&vi).Error; err != nil {
			return nil, fmt.Errorf("%w: vendor invoice not found for this item", ErrNotFound)
		}
		item.InvoiceID = vi.InvoiceID

		ri := invModel.ReceiptsItems{
			InvoiceID:   vi.InvoiceID,
			InventoryID: item.IDInventory,
			DateTime:    now,
		}
		s.db.Create(&ri)

		newTxn := invModel.InventoryTransaction{
			InventoryID:     item.IDInventory,
			ToLocationID:    locID,
			TransferredBy:   empID,
			InvoiceID:       vi.InvoiceID,
			StatusItems:     "Ready for Sale",
			TransactionType: "Received from Vendor",
			DateTransaction: now,
		}
		s.db.Create(&newTxn)
	}

	s.db.Save(&item)
	s.db.Save(&txn)

	var respItems []ReceiptItemRow
	if len(inv.NumberInvoice) > 0 && inv.NumberInvoice[0] == 'V' {
		var receipts []invModel.ReceiptsItems
		s.db.Where("invoice_id = ?", inv.IDInvoice).Find(&receipts)
		for _, r := range receipts {
			if row := s.buildItemRow(r.InventoryID, "V", r.DateTime); row != nil {
				respItems = append(respItems, *row)
			}
		}
	} else {
		var txns []invModel.InventoryTransaction
		s.db.Where("invoice_id = ?", inv.IDInvoice).Find(&txns)
		for _, t := range txns {
			if row := s.buildItemRow(t.InventoryID, "I", t.DateTransaction); row != nil {
				respItems = append(respItems, *row)
			}
		}
	}
	if respItems == nil {
		respItems = []ReceiptItemRow{}
	}

	msg := "Item confirmed received"
	if req.SKU != "" {
		msg = fmt.Sprintf("Item with SKU %s has been successfully received.", req.SKU)
	}
	return &ConfirmReceiptResult{Message: msg, Items: respItems}, nil
}

// ─── Pay Transfer ─────────────────────────────────────────────────────────────

type PayTransferResult struct {
	Message       string `json:"message"`
	InvoiceID     int64  `json:"invoice_id"`
	NumberInvoice string `json:"number_invoice"`
	TotalAmount   string `json:"total_amount"`
	Due           string `json:"due"`
}

func (s *Service) PayTransfer(el *EmpLocation, invoiceID int64) (*PayTransferResult, error) {
	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, ErrNotFound
	}
	if len(inv.NumberInvoice) == 0 || inv.NumberInvoice[0] != 'I' {
		return nil, fmt.Errorf("%w: only internal transfer invoices can be marked as paid", ErrBadRequest)
	}
	if inv.ToLocationID == nil || *inv.ToLocationID != int64(el.Location.IDLocation) {
		return nil, fmt.Errorf("%w: only the receiving location can mark this invoice as paid", ErrForbidden)
	}

	paidAmount := inv.TotalAmount
	if inv.Due == 0 {
		inv.Due = paidAmount
	}
	inv.Due = 0

	empID := int64(el.Employee.IDEmployee)
	pmID := int64(1)
	payment := patients.PaymentHistory{
		InvoiceID:       inv.IDInvoice,
		Amount:          paidAmount,
		PaymentMethodID: &pmID,
		EmployeeID:      &empID,
	}
	s.db.Create(&payment)
	s.db.Save(&inv)

	return &PayTransferResult{
		Message:       "Transfer invoice marked as paid",
		InvoiceID:     inv.IDInvoice,
		NumberInvoice: inv.NumberInvoice,
		TotalAmount:   fmtFloat(paidAmount),
		Due:           "0.00",
	}, nil
}
