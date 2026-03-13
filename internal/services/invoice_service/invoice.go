package invoice_service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	invModel "sighthub-backend/internal/models/inventory"
	"sighthub-backend/internal/models/invoices"
	"sighthub-backend/internal/models/location"
	"sighthub-backend/internal/models/types"
)

// ─── Create Invoice ───────────────────────────────────────────────────────────

type CreateInvoiceRequest struct {
	VendorID     *int64   `json:"vendor_id"`
	ToLocationID *int64   `json:"location_id"`
	Items        []ItemIn `json:"items"`
	Discount     float64  `json:"discount"`
}

type ItemIn struct {
	SKU         string `json:"sku"`
	InventoryID *int64 `json:"inventory_id"`
}

type CreateInvoiceResult struct {
	Message         string  `json:"message"`
	NumberInvoice   string  `json:"number_invoice"`
	InvoiceID       int64   `json:"invoice_id"`
	TotalCost       float64 `json:"total_cost"`
	FinalAmount     float64 `json:"final_amount"`
	Discount        float64 `json:"discount"`
	Employee        string  `json:"employee"`
	TotalQuantity   int     `json:"total_quantity"`
	ToLocationID    *int64  `json:"to_location_id,omitempty"`
	ToLocationName  *string `json:"to_location_name,omitempty"`
	VendorInvoiceID *int64  `json:"vendor_invoice_id,omitempty"`
	ShipmentID      *int64  `json:"shipment_id,omitempty"`
}

func (s *Service) CreateInvoice(el *EmpLocation, req CreateInvoiceRequest) (*CreateInvoiceResult, error) {
	if (req.VendorID == nil || *req.VendorID == 0) && (req.ToLocationID == nil || *req.ToLocationID == 0) {
		return nil, fmt.Errorf("%w: either vendor_id or location_id must be provided", ErrBadRequest)
	}

	invoiceType := "V"
	if req.VendorID == nil || *req.VendorID == 0 {
		invoiceType = "I"
	}

	shortName := ""
	if el.Location.ShortName != nil {
		shortName = *el.Location.ShortName
	}
	invoiceNumber, err := s.createInvoiceNumber(invoiceType, shortName)
	if err != nil {
		return nil, err
	}

	totalAmount, inventoryItems, err := s.calculateTotal(invoiceType, req.Items)
	if err != nil {
		return nil, err
	}
	finalAmount := totalAmount - req.Discount

	if invoiceType == "I" {
		for _, item := range inventoryItems {
			if item.LocationID != int64(el.Location.IDLocation) {
				return nil, fmt.Errorf("%w: item %d does not belong to current location", ErrBadRequest, item.IDInventory)
			}
			st := string(item.StatusItemsInventory)
			if st != "Ready for Sale" && st != "Defective" {
				return nil, fmt.Errorf("%w: item %d has status '%s' and cannot be added", ErrBadRequest, item.IDInventory, st)
			}
			var cnt int64
			s.db.Model(&invModel.Missing{}).Where("inventory_id = ?", item.IDInventory).Count(&cnt)
			if cnt > 0 {
				return nil, fmt.Errorf("%w: item %d is marked as missing", ErrBadRequest, item.IDInventory)
			}
		}
	}

	due := 0.0
	if invoiceType == "I" {
		due = totalAmount
	}
	var toLocID *int64
	var vendorID int64
	if invoiceType == "I" {
		toLocID = req.ToLocationID
	} else {
		vendorID = *req.VendorID
	}
	empID := int64(el.Employee.IDEmployee)

	inv := invoices.Invoice{
		NumberInvoice: invoiceNumber,
		DateCreate:    time.Now(),
		Discount:      &req.Discount,
		TotalAmount:   totalAmount,
		FinalAmount:   finalAmount,
		Due:           due,
		Quantity:      len(inventoryItems),
		EmployeeID:    &empID,
		LocationID:    int64(el.Location.IDLocation),
		ToLocationID:  toLocID,
		VendorID:      vendorID,
	}
	if err := s.db.Create(&inv).Error; err != nil {
		return nil, err
	}

	for i := range inventoryItems {
		inventoryItems[i].InvoiceID = inv.IDInvoice
		s.db.Save(&inventoryItems[i])
	}

	var vendorInvoiceID, shipmentID *int64

	if invoiceType == "I" {
		var targetLoc location.Location
		if err := s.db.First(&targetLoc, *req.ToLocationID).Error; err != nil {
			s.db.Delete(&inv)
			return nil, fmt.Errorf("target location not found")
		}
		effectiveToLocID := int64(targetLoc.IDLocation)

		canReceive := targetLoc.CanReceiveItems != nil && *targetLoc.CanReceiveItems
		if !canReceive {
			isCurrentWarehouse := targetLoc.WarehouseID != nil && int64(el.Location.IDLocation) == int64(*targetLoc.WarehouseID)
			if isCurrentWarehouse {
				// allowed: current IS the warehouse
			} else if targetLoc.WarehouseID != nil {
				effectiveToLocID = int64(*targetLoc.WarehouseID)
			} else {
				s.db.Delete(&inv)
				return nil, fmt.Errorf("%w: target location cannot receive items and has no warehouse", ErrBadRequest)
			}
			inv.ToLocationID = &effectiveToLocID
			s.db.Save(&inv)
		}

		for i := range inventoryItems {
			item := &inventoryItems[i]
			oldInvoiceID := item.InvoiceID

			isLocal := effectiveToLocID == int64(el.Location.IDLocation) ||
				(el.Location.WarehouseID != nil && effectiveToLocID == int64(*el.Location.WarehouseID))
			if isLocal {
				item.StatusItemsInventory = types.StatusInventoryReadyForSale
			} else {
				item.StatusItemsInventory = types.StatusInventoryICTSentAndNotReceived
			}
			item.LocationID = effectiveToLocID
			item.InvoiceID = inv.IDInvoice
			s.db.Save(item)

			s.db.Create(&invModel.InventoryTransfer{
				InventoryID:    item.IDInventory,
				FromLocationID: int64(el.Location.IDLocation),
				ToLocationID:   effectiveToLocID,
				TransferredBy:  empID,
				StatusItems:    item.StatusItemsInventory,
				InvoiceID:      inv.IDInvoice,
				InvoiceFrom:    &oldInvoiceID,
				InvoiceTo:      &inv.IDInvoice,
				SystemNote: strPtr(fmt.Sprintf("Transfer logged for item %d in invoice %s",
					item.IDInventory, inv.NumberInvoice)),
			})
			s.db.Create(&invModel.InventoryTransaction{
				InventoryID:     &item.IDInventory,
				FromLocationID:  int64Ptr(int64(el.Location.IDLocation)),
				ToLocationID:    &effectiveToLocID,
				TransferredBy:   empID,
				InvoiceID:       &inv.IDInvoice,
				OldInvoiceID:    &oldInvoiceID,
				TransactionType: "Internal Transfer",
				StatusItems:     item.StatusItemsInventory,
			})
		}
	} else {
		vid, sid, err := s.createVendorInvoiceAndShipment(&inv, vendorID, el, inventoryItems)
		if err != nil {
			return nil, err
		}
		vendorInvoiceID = &vid
		shipmentID = &sid

		for i := range inventoryItems {
			item := &inventoryItems[i]
			s.db.Create(&invModel.ReceiptsItems{
				InvoiceID: inv.IDInvoice, InventoryID: item.IDInventory, DateTime: time.Now(),
			})
			s.db.Create(&invModel.InventoryTransaction{
				InventoryID:     &item.IDInventory,
				ToLocationID:    int64Ptr(int64(el.Location.IDLocation)),
				TransferredBy:   empID,
				InvoiceID:       &inv.IDInvoice,
				TransactionType: "Vendor Shipment",
				StatusItems:     item.StatusItemsInventory,
			})
		}
	}

	result := &CreateInvoiceResult{
		Message:         "Invoice created successfully",
		NumberInvoice:   inv.NumberInvoice,
		InvoiceID:       inv.IDInvoice,
		TotalCost:       inv.TotalAmount,
		FinalAmount:     inv.FinalAmount,
		Discount:        req.Discount,
		Employee:        el.Employee.FirstName + " " + el.Employee.LastName,
		TotalQuantity:   inv.Quantity,
		VendorInvoiceID: vendorInvoiceID,
		ShipmentID:      shipmentID,
	}
	if invoiceType == "I" {
		result.ToLocationID = inv.ToLocationID
		if inv.ToLocationID != nil {
			var toLoc location.Location
			if s.db.First(&toLoc, *inv.ToLocationID).Error == nil {
				result.ToLocationName = &toLoc.FullName
			}
		}
	}
	return result, nil
}

// ─── Update Invoice ───────────────────────────────────────────────────────────

type UpdateInvoiceResult struct {
	Message        string        `json:"message"`
	DateCreate     string        `json:"date_create"`
	InvoiceID      int64         `json:"invoice_id"`
	Employee       string        `json:"employee"`
	NumberInvoice  string        `json:"number_invoice"`
	TotalCost      string        `json:"total_cost"`
	FinalAmount    string        `json:"final_amount"`
	TotalDiscount  string        `json:"total_discount"`
	TotalQuantity  int           `json:"total_quantity"`
	Due            *string       `json:"due,omitempty"`
	ToLocationID   *int64        `json:"to_location_id,omitempty"`
	ToLocationName *string       `json:"to_location_name,omitempty"`
	LocationID     *int64        `json:"location_id,omitempty"`
	LocationName   *string       `json:"location_name,omitempty"`
	Items          []interface{} `json:"items"`
}

func (s *Service) UpdateInvoice(el *EmpLocation, invoiceID int64, dateCreate *string, discount *float64, items []ItemIn) (*UpdateInvoiceResult, error) {
	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: invoice not found", ErrNotFound)
	}
	if inv.PatientID != 0 {
		return nil, fmt.Errorf("%w: patient invoices cannot be updated", ErrBadRequest)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("%w: no items provided", ErrBadRequest)
	}

	if dateCreate != nil {
		t, err := time.Parse("2006-01-02", *dateCreate)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid date_create format, use YYYY-MM-DD", ErrBadRequest)
		}
		inv.DateCreate = t
	}
	if discount != nil {
		inv.Discount = discount
	}

	totalAmount := inv.TotalAmount
	var totalQuantity int
	if strings.HasPrefix(inv.NumberInvoice, "I") {
		var txs []invModel.InventoryTransaction
		s.db.Where("invoice_id = ?", invoiceID).Find(&txs)
		seen := map[int64]bool{}
		for _, tx := range txs {
			if tx.InventoryID != nil {
				seen[*tx.InventoryID] = true
			}
		}
		totalQuantity = len(seen)
	} else {
		totalQuantity = inv.Quantity
	}

	empID := int64(el.Employee.IDEmployee)

	for _, it := range items {
		var item invModel.Inventory
		if it.SKU != "" {
			if err := s.db.Where("sku = ? AND location_id = ?", it.SKU, el.Location.IDLocation).First(&item).Error; err != nil {
				return nil, fmt.Errorf("%w: item with SKU %s not found in current location", ErrBadRequest, it.SKU)
			}
		} else if it.InventoryID != nil {
			if err := s.db.Where("id_inventory = ? AND location_id = ?", *it.InventoryID, el.Location.IDLocation).First(&item).Error; err != nil {
				return nil, fmt.Errorf("%w: item %d not found in current location", ErrBadRequest, *it.InventoryID)
			}
		} else {
			return nil, fmt.Errorf("%w: each item must have sku or inventory_id", ErrBadRequest)
		}

		if string(item.StatusItemsInventory) != "Ready for Sale" {
			return nil, fmt.Errorf("%w: item %s must be 'Ready for Sale' to be transferred", ErrBadRequest, item.SKU)
		}
		var existCnt int64
		s.db.Model(&invModel.InventoryTransaction{}).Where("inventory_id = ? AND invoice_id = ?", item.IDInventory, invoiceID).Count(&existCnt)
		if existCnt > 0 {
			return nil, fmt.Errorf("%w: item %s is already on this invoice", ErrBadRequest, item.SKU)
		}

		itemNet := 0.0
		var pb invModel.PriceBook
		if s.db.Where("inventory_id = ?", item.IDInventory).First(&pb).Error == nil {
			lc, d := 0.0, 0.0
			if pb.ItemListCost != nil {
				lc = *pb.ItemListCost
			}
			if pb.ItemDiscount != nil {
				d = *pb.ItemDiscount
			}
			itemNet = lc - d
		}
		totalAmount += itemNet
		totalQuantity++

		oldInvoiceID := item.InvoiceID

		var currentLoc location.Location
		s.db.First(&currentLoc, inv.LocationID)

		toLocID := int64(0)
		if inv.ToLocationID != nil {
			toLocID = *inv.ToLocationID
		}

		var txType string
		isLocal := currentLoc.WarehouseID != nil &&
			(toLocID == int64(*currentLoc.WarehouseID) || inv.LocationID == int64(*currentLoc.WarehouseID))

		if isLocal {
			item.LocationID = toLocID
			item.StatusItemsInventory = types.StatusInventoryReadyForSale
			txType = "Local"
		} else {
			var targetLoc location.Location
			canReceive := false
			if s.db.First(&targetLoc, toLocID).Error == nil {
				canReceive = targetLoc.CanReceiveItems != nil && *targetLoc.CanReceiveItems
			}
			if canReceive {
				item.StatusItemsInventory = types.StatusInventoryICTSentAndNotReceived
				txType = "Transfer"
			} else {
				if currentLoc.WarehouseID != nil {
					newToLoc := int64(*currentLoc.WarehouseID)
					inv.ToLocationID = &newToLoc
					toLocID = newToLoc
				}
				item.StatusItemsInventory = types.StatusInventoryICTSentAndNotReceived
				txType = "Transfer"
			}
		}
		s.db.Save(&item)

		s.db.Create(&invModel.InventoryTransaction{
			InventoryID:     &item.IDInventory,
			FromLocationID:  &inv.LocationID,
			ToLocationID:    &toLocID,
			TransferredBy:   empID,
			InvoiceID:       &inv.IDInvoice,
			OldInvoiceID:    &oldInvoiceID,
			StatusItems:     item.StatusItemsInventory,
			TransactionType: txType,
		})
		s.db.Create(&invModel.InventoryTransfer{
			InventoryID:    item.IDInventory,
			FromLocationID: inv.LocationID,
			ToLocationID:   toLocID,
			TransferredBy:  empID,
			StatusItems:    item.StatusItemsInventory,
			InvoiceID:      inv.IDInvoice,
			InvoiceFrom:    &oldInvoiceID,
			InvoiceTo:      &inv.IDInvoice,
			SystemNote: strPtr(fmt.Sprintf("%s transfer logged from location %d to %d",
				txType, inv.LocationID, toLocID)),
		})
	}

	disc := 0.0
	if inv.Discount != nil {
		disc = *inv.Discount
	}
	inv.TotalAmount = totalAmount
	inv.FinalAmount = totalAmount - disc
	inv.Quantity = totalQuantity
	if strings.HasPrefix(inv.NumberInvoice, "I") {
		inv.Due = totalAmount
	}
	s.db.Save(&inv)

	var txs []invModel.InventoryTransaction
	s.db.Where("invoice_id = ?", invoiceID).Find(&txs)
	var itemsResp []interface{}
	for _, tx := range txs {
		var item invModel.Inventory
		if s.db.First(&item, tx.InventoryID).Error != nil {
			continue
		}
		price := 0.0
		var pb invModel.PriceBook
		if s.db.Where("inventory_id = ?", item.IDInventory).First(&pb).Error == nil && pb.ItemNet != nil {
			price = *pb.ItemNet
		}
		itemsResp = append(itemsResp, map[string]interface{}{
			"sku":              item.SKU,
			"status":           string(item.StatusItemsInventory),
			"price":            fmtFloat(price),
			"date_transaction": tx.DateTransaction.Format(time.RFC3339),
		})
	}

	result := &UpdateInvoiceResult{
		Message:       "Invoice updated successfully",
		DateCreate:    inv.DateCreate.Format("2006-01-02"),
		InvoiceID:     inv.IDInvoice,
		Employee:      el.Employee.FirstName + " " + el.Employee.LastName,
		NumberInvoice: inv.NumberInvoice,
		TotalCost:     fmtFloat(inv.TotalAmount),
		FinalAmount:   fmtFloat(inv.FinalAmount),
		TotalDiscount: fmtFloatPtr(inv.Discount),
		TotalQuantity: inv.Quantity,
		Items:         itemsResp,
	}
	if strings.HasPrefix(inv.NumberInvoice, "I") {
		due := fmtFloat(inv.Due)
		result.Due = &due
		result.ToLocationID = inv.ToLocationID
		if inv.ToLocationID != nil {
			var toLoc location.Location
			if s.db.First(&toLoc, *inv.ToLocationID).Error == nil {
				result.ToLocationName = &toLoc.FullName
			}
		}
		locID := inv.LocationID
		result.LocationID = &locID
		var fromLoc location.Location
		if s.db.First(&fromLoc, inv.LocationID).Error == nil {
			result.LocationName = &fromLoc.FullName
		}
	}
	return result, nil
}

// ─── View Invoice ─────────────────────────────────────────────────────────────

func (s *Service) ViewInvoice(el *EmpLocation, invoiceID int64) (map[string]interface{}, error) {
	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: invoice not found", ErrNotFound)
	}
	var loc location.Location
	if err := s.db.First(&loc, inv.LocationID).Error; err != nil {
		return nil, fmt.Errorf("%w: location not found", ErrNotFound)
	}
	var toLocName *string
	if inv.ToLocationID != nil {
		var toLoc location.Location
		if s.db.First(&toLoc, *inv.ToLocationID).Error == nil {
			toLocName = &toLoc.FullName
		}
	}

	var txs []invModel.InventoryTransaction
	s.db.Where("invoice_id = ?", inv.IDInvoice).Find(&txs)

	latest := map[int64]invModel.InventoryTransaction{}
	for _, tx := range txs {
		if tx.InventoryID == nil {
			continue
		}
		if existing, ok := latest[*tx.InventoryID]; !ok || tx.DateTransaction.After(existing.DateTransaction) {
			latest[*tx.InventoryID] = tx
		}
	}

	var txData []interface{}
	for _, tx := range latest {
		var item invModel.Inventory
		if s.db.First(&item, tx.InventoryID).Error != nil {
			continue
		}
		var info struct {
			ProductTitle string `gorm:"column:product_title"`
			VariantTitle string `gorm:"column:variant_title"`
			BrandName    string `gorm:"column:brand_name"`
		}
		s.db.Raw(`SELECT p.title_product AS product_title, m.title_variant AS variant_title,
			COALESCE(b.brand_name,'') AS brand_name FROM model m
			JOIN product p ON p.id_product = m.product_id
			LEFT JOIN brand b ON b.id_brand = p.brand_id WHERE m.id_model = ?`, item.ModelID).Scan(&info)

		var pb invModel.PriceBook
		var price *string
		if s.db.Where("inventory_id = ?", item.IDInventory).First(&pb).Error == nil && pb.ItemNet != nil {
			p := fmtFloat(*pb.ItemNet)
			price = &p
		}
		txData = append(txData, map[string]interface{}{
			"inventory_id":     item.IDInventory,
			"sku":              item.SKU,
			"price":            price,
			"product_title":    info.ProductTitle,
			"variant_title":    info.VariantTitle,
			"brand_name":       info.BrandName,
			"from_location_id": tx.FromLocationID,
			"to_location_id":   tx.ToLocationID,
			"status":           string(tx.StatusItems),
			"date_transfer":    tx.DateTransaction.Format(time.RFC3339),
			"transferred_by":   tx.TransferredBy,
		})
	}

	return map[string]interface{}{
		"invoice_id":       inv.IDInvoice,
		"number_invoice":   inv.NumberInvoice,
		"employee":         el.Employee.FirstName + " " + el.Employee.LastName,
		"date_create":      inv.DateCreate.Format(time.RFC3339),
		"location_name":    loc.FullName,
		"to_location_name": toLocName,
		"total_quantity":   inv.Quantity,
		"total_cost":       fmtFloat(inv.TotalAmount),
		"final_amount":     fmtFloat(inv.FinalAmount),
		"due":              fmtFloat(inv.Due),
		"transactions":     txData,
	}, nil
}

// ─── View Invoice Items ───────────────────────────────────────────────────────

func (s *Service) ViewInvoiceItem(el *EmpLocation, invoiceID int64) (map[string]interface{}, error) {
	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: invoice not found", ErrNotFound)
	}
	var loc location.Location
	s.db.First(&loc, inv.LocationID)

	type itemRow struct {
		SKU          string
		ProductTitle string
		VariantTitle string
		BrandName    string
		VendorName   string
		Status       string
		ItemListCost *float64
		ItemDiscount *float64
		ItemNet      *float64
		PbSelling    *float64
		DateInfo     time.Time
	}
	var items []itemRow

	loadInfo := func(modelID interface{}) (productTitle, variantTitle, brandName, vendorName string) {
		var info struct {
			ProductTitle string `gorm:"column:product_title"`
			VariantTitle string `gorm:"column:variant_title"`
			BrandName    string `gorm:"column:brand_name"`
			VendorName   string `gorm:"column:vendor_name"`
		}
		s.db.Raw(`SELECT p.title_product AS product_title, m.title_variant AS variant_title,
			COALESCE(b.brand_name,'') AS brand_name, COALESCE(v.vendor_name,'') AS vendor_name
			FROM model m JOIN product p ON p.id_product = m.product_id
			LEFT JOIN brand b ON b.id_brand = p.brand_id
			LEFT JOIN vendor v ON v.id_vendor = p.vendor_id WHERE m.id_model = ?`, modelID).Scan(&info)
		return info.ProductTitle, info.VariantTitle, info.BrandName, info.VendorName
	}

	if strings.HasPrefix(inv.NumberInvoice, "V") {
		var receipts []invModel.ReceiptsItems
		s.db.Where("invoice_id = ?", invoiceID).Find(&receipts)
		for _, r := range receipts {
			var item invModel.Inventory
			if s.db.First(&item, r.InventoryID).Error != nil {
				continue
			}
			var pb invModel.PriceBook
			s.db.Where("inventory_id = ?", item.IDInventory).First(&pb)
			pt, vt, bn, vn := loadInfo(item.ModelID)
			items = append(items, itemRow{
				SKU: item.SKU, ProductTitle: pt, VariantTitle: vt, BrandName: bn, VendorName: vn,
				Status: string(item.StatusItemsInventory), ItemListCost: pb.ItemListCost,
				ItemDiscount: pb.ItemDiscount, ItemNet: pb.ItemNet, PbSelling: pb.PbSellingPrice,
				DateInfo: r.DateTime,
			})
		}
	} else if strings.HasPrefix(inv.NumberInvoice, "I") {
		var txs []invModel.InventoryTransaction
		s.db.Where("invoice_id = ?", invoiceID).Find(&txs)
		for _, tx := range txs {
			var item invModel.Inventory
			if s.db.First(&item, tx.InventoryID).Error != nil {
				continue
			}
			var pb invModel.PriceBook
			s.db.Where("inventory_id = ?", item.IDInventory).First(&pb)
			pt, vt, bn, _ := loadInfo(item.ModelID)
			items = append(items, itemRow{
				SKU: item.SKU, ProductTitle: pt, VariantTitle: vt, BrandName: bn,
				Status: string(tx.StatusItems), ItemListCost: pb.ItemListCost,
				ItemDiscount: pb.ItemDiscount, ItemNet: pb.ItemNet, PbSelling: pb.PbSellingPrice,
				DateInfo: tx.DateTransaction,
			})
		}
	} else {
		return nil, fmt.Errorf("%w: unknown invoice type", ErrBadRequest)
	}

	associatedName := "Internal location Transfer"
	var vendorInvoiceID *int64
	if strings.HasPrefix(inv.NumberInvoice, "V") {
		var vn struct{ VendorName string }
		s.db.Raw(`SELECT v.vendor_name FROM vendor v
			JOIN product p ON p.vendor_id = v.id_vendor JOIN model m ON m.product_id = p.id_product
			JOIN inventory i ON i.model_id = m.id_model
			JOIN receipts_items ri ON ri.inventory_id = i.id_inventory
			WHERE ri.invoice_id = ? LIMIT 1`, invoiceID).Scan(&vn)
		if vn.VendorName != "" {
			associatedName = vn.VendorName
		}
		var vi invModel.VendorInvoice
		if s.db.Where("invoice_id = ?", inv.IDInvoice).First(&vi).Error == nil {
			vendorInvoiceID = &vi.IDVendorInvoice
		}
	}

	var itemsOut []interface{}
	for _, it := range items {
		row := map[string]interface{}{
			"sku":              it.SKU,
			"product_title":    it.ProductTitle,
			"variant_title":    it.VariantTitle,
			"brand_name":       it.BrandName,
			"status":           it.Status,
			"item_list_cost":   fmtFloatPtr(it.ItemListCost),
			"item_discount":    fmtFloatPtr(it.ItemDiscount),
			"item_net":         fmtFloatPtr(it.ItemNet),
			"pb_selling_price": fmtFloatPtr(it.PbSelling),
		}
		if strings.HasPrefix(inv.NumberInvoice, "V") {
			row["vendor_name"] = it.VendorName
			row["datetime_received"] = it.DateInfo.Format(time.RFC3339)
		} else {
			row["date_transfer"] = it.DateInfo.Format(time.RFC3339)
		}
		itemsOut = append(itemsOut, row)
	}

	result := map[string]interface{}{
		"invoice_id":      inv.IDInvoice,
		"number_invoice":  inv.NumberInvoice,
		"employee":        el.Employee.FirstName + " " + el.Employee.LastName,
		"date_create":     inv.DateCreate.Format(time.RFC3339),
		"total_amount":    fmtFloat(inv.TotalAmount),
		"final_amount":    fmtFloat(inv.FinalAmount),
		"total_quantity":  inv.Quantity,
		"location_name":   loc.FullName,
		"associated_name": associatedName,
		"items":           itemsOut,
	}
	if vendorInvoiceID != nil {
		result["vendor_invoice_id"] = *vendorInvoiceID
	}
	if strings.HasPrefix(inv.NumberInvoice, "I") {
		result["due"] = fmtFloat(inv.Due)
	}
	return result, nil
}

// ─── Delete Item ──────────────────────────────────────────────────────────────

func (s *Service) DeleteItem(invoiceID int64, sku string, inventoryID *int64) (map[string]interface{}, error) {
	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: invoice not found", ErrNotFound)
	}

	var item invModel.Inventory
	if sku != "" {
		sku = strings.Trim(sku, "'")
		if err := s.db.Where("sku = ?", sku).First(&item).Error; err != nil {
			return nil, fmt.Errorf("%w: item not found", ErrNotFound)
		}
	} else if inventoryID != nil {
		if err := s.db.First(&item, *inventoryID).Error; err != nil {
			return nil, fmt.Errorf("%w: item not found", ErrNotFound)
		}
	} else {
		return nil, fmt.Errorf("%w: sku or inventory_id required", ErrBadRequest)
	}

	linked := item.InvoiceID == invoiceID
	if !linked {
		var cnt int64
		s.db.Model(&invModel.ReceiptsItems{}).Where("invoice_id = ? AND inventory_id = ?", invoiceID, item.IDInventory).Count(&cnt)
		linked = cnt > 0
	}
	if !linked {
		var cnt int64
		s.db.Model(&invModel.InventoryTransaction{}).Where("invoice_id = ? AND inventory_id = ?", invoiceID, item.IDInventory).Count(&cnt)
		linked = cnt > 0
	}
	if !linked {
		return nil, fmt.Errorf("%w: item is not associated with this invoice", ErrBadRequest)
	}

	if strings.HasPrefix(inv.NumberInvoice, "I") {
		var rcnt int64
		s.db.Model(&invModel.ReceiptsItems{}).Where("invoice_id = ? AND inventory_id = ?", invoiceID, item.IDInventory).Count(&rcnt)
		if rcnt > 0 {
			return nil, fmt.Errorf("%w: item was already received by target location", ErrBadRequest)
		}
		var oldTx invModel.InventoryTransaction
		if err := s.db.Where("inventory_id = ? AND invoice_id = ?", item.IDInventory, invoiceID).First(&oldTx).Error; err != nil {
			return nil, fmt.Errorf("%w: no transaction found for item on this invoice", ErrBadRequest)
		}
		if s.db.Where("inventory_id = ? AND invoice_id = ?", item.IDInventory, invoiceID).First(&invModel.InventoryTransfer{}).Error != nil {
			return nil, fmt.Errorf("%w: item not associated with this transfer", ErrBadRequest)
		}
		if oldTx.TransactionType == "Local" && oldTx.FromLocationID != nil {
			item.LocationID = *oldTx.FromLocationID
		}
		item.StatusItemsInventory = types.StatusInventoryReadyForSale
		if oldTx.OldInvoiceID != nil {
			item.InvoiceID = *oldTx.OldInvoiceID
		}
		s.db.Where("inventory_id = ? AND invoice_id = ?", item.IDInventory, invoiceID).Delete(&invModel.InventoryTransaction{})
		s.db.Where("inventory_id = ? AND invoice_id = ?", item.IDInventory, invoiceID).Delete(&invModel.InventoryTransfer{})
		s.db.Save(&item)

	} else if strings.HasPrefix(inv.NumberInvoice, "V") {
		var otherTx invModel.InventoryTransaction
		if s.db.Where("inventory_id = ? AND invoice_id != ? AND invoice_id != 0", item.IDInventory, invoiceID).First(&otherTx).Error == nil {
			return nil, fmt.Errorf("%w: item has transactions in other invoices", ErrBadRequest)
		}
		var rtvCnt int64
		s.db.Raw("SELECT COUNT(*) FROM return_to_vendor_item WHERE inventory_id = ?", item.IDInventory).Scan(&rtvCnt)
		if rtvCnt > 0 {
			return nil, fmt.Errorf("%w: item is part of a Return-to-Vendor invoice", ErrBadRequest)
		}
		var missingCnt int64
		s.db.Model(&invModel.Missing{}).Where("inventory_id = ?", item.IDInventory).Count(&missingCnt)
		if missingCnt > 0 {
			return nil, fmt.Errorf("%w: item is in inventory count/missing list", ErrBadRequest)
		}
		s.db.Where("inventory_id = ?", item.IDInventory).Delete(&invModel.InventoryTransaction{})
		s.db.Where("inventory_id = ?", item.IDInventory).Delete(&invModel.InventoryTransfer{})
		s.db.Where("inventory_id = ? AND invoice_id = ?", item.IDInventory, invoiceID).Delete(&invModel.ReceiptsItems{})
		s.db.Where("inventory_id = ?", item.IDInventory).Delete(&invModel.Missing{})
		s.db.Where("inventory_id = ?", item.IDInventory).Delete(&invModel.PriceBook{})
		s.db.Delete(&item)
	} else {
		return nil, fmt.Errorf("%w: deletion not implemented for this invoice type", ErrBadRequest)
	}

	var newTotal float64
	var newCount int
	if strings.HasPrefix(inv.NumberInvoice, "I") {
		var txs []invModel.InventoryTransaction
		s.db.Where("invoice_id = ?", invoiceID).Find(&txs)
		seen := map[int64]bool{}
		for _, tx := range txs {
			if tx.InventoryID != nil {
				seen[*tx.InventoryID] = true
			}
		}
		newCount = len(seen)
		for id := range seen {
			var pb invModel.PriceBook
			if s.db.Where("inventory_id = ?", id).First(&pb).Error == nil {
				lc, d := 0.0, 0.0
				if pb.ItemListCost != nil {
					lc = *pb.ItemListCost
				}
				if pb.ItemDiscount != nil {
					d = *pb.ItemDiscount
				}
				newTotal += lc - d
			}
		}
	} else {
		var remaining []invModel.Inventory
		s.db.Where("invoice_id = ?", invoiceID).Find(&remaining)
		newCount = len(remaining)
		for _, it := range remaining {
			var pb invModel.PriceBook
			if s.db.Where("inventory_id = ?", it.IDInventory).First(&pb).Error == nil {
				lc, d := 0.0, 0.0
				if pb.ItemListCost != nil {
					lc = *pb.ItemListCost
				}
				if pb.ItemDiscount != nil {
					d = *pb.ItemDiscount
				}
				newTotal += lc - d
			}
		}
	}
	disc := 0.0
	if inv.Discount != nil {
		disc = *inv.Discount
	}
	inv.TotalAmount = newTotal
	inv.FinalAmount = newTotal - disc
	inv.Quantity = newCount
	if strings.HasPrefix(inv.NumberInvoice, "I") {
		inv.Due = newTotal
	}
	s.db.Save(&inv)

	return map[string]interface{}{
		"message":              "Item successfully removed and invoice updated",
		"invoice_total_amount": fmtFloat(inv.TotalAmount),
		"invoice_final_amount": fmtFloat(inv.FinalAmount),
		"due":                  fmtFloat(inv.Due),
		"quantity":             inv.Quantity,
	}, nil
}

// ─── Delete Invoice ───────────────────────────────────────────────────────────

func (s *Service) DeleteInvoice(invoiceID int64) (string, error) {
	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("%w: invoice not found", ErrNotFound)
	}
	disc := 0.0
	if inv.Discount != nil {
		disc = *inv.Discount
	}
	if inv.Quantity != 0 || inv.TotalAmount != 0 || inv.FinalAmount != 0 || disc != 0 {
		return "", fmt.Errorf("%w: invoice %s contains items or non-zero total", ErrBadRequest, inv.NumberInvoice)
	}
	s.db.Where("invoice_id = ?", inv.IDInvoice).Delete(&invModel.InventoryTransfer{})
	var inventoryItems []invModel.Inventory
	s.db.Where("invoice_id = ?", inv.IDInvoice).Find(&inventoryItems)
	for i := range inventoryItems {
		inventoryItems[i].InvoiceID = 0
		s.db.Save(&inventoryItems[i])
	}
	var vendorInv invModel.VendorInvoice
	if s.db.Where("invoice_id = ?", inv.IDInvoice).First(&vendorInv).Error == nil {
		s.db.Where("vendor_invoice_id = ?", vendorInv.IDVendorInvoice).Delete(&invModel.Shipment{})
		s.db.Delete(&vendorInv)
	}
	s.db.Delete(&inv)
	return inv.NumberInvoice, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (s *Service) createInvoiceNumber(invoiceType, shortName string) (string, error) {
	var maxID int64
	s.db.Model(&invoices.Invoice{}).Select("COALESCE(MAX(id_invoice), 0)").Scan(&maxID)
	return fmt.Sprintf("%s%s%07d", invoiceType, shortName, maxID+1), nil
}

func (s *Service) calculateTotal(invoiceType string, items []ItemIn) (float64, []invModel.Inventory, error) {
	var total float64
	var inventoryItems []invModel.Inventory
	for _, it := range items {
		var item invModel.Inventory
		if it.SKU != "" {
			if err := s.db.Where("sku = ?", it.SKU).First(&item).Error; err != nil {
				return 0, nil, fmt.Errorf("%w: inventory with SKU %s not found", ErrBadRequest, it.SKU)
			}
		} else if it.InventoryID != nil {
			if err := s.db.First(&item, *it.InventoryID).Error; err != nil {
				return 0, nil, fmt.Errorf("%w: inventory %d not found", ErrBadRequest, *it.InventoryID)
			}
		}
		var pb invModel.PriceBook
		if err := s.db.Where("inventory_id = ?", item.IDInventory).First(&pb).Error; err != nil {
			return 0, nil, fmt.Errorf("%w: price not found for inventory %d", ErrBadRequest, item.IDInventory)
		}
		var price float64
		if invoiceType == "V" || invoiceType == "I" {
			if pb.ItemNet != nil {
				price = *pb.ItemNet
			} else {
				lc, d := 0.0, 0.0
				if pb.ItemListCost != nil {
					lc = *pb.ItemListCost
				}
				if pb.ItemDiscount != nil {
					d = *pb.ItemDiscount
				}
				price = lc - d
			}
		} else {
			if pb.PbSellingPrice != nil {
				price = *pb.PbSellingPrice
				if pb.PbDiscount != nil {
					price -= *pb.PbDiscount
				}
			}
		}
		total += price
		inventoryItems = append(inventoryItems, item)
	}
	return total, inventoryItems, nil
}

func (s *Service) createVendorInvoiceAndShipment(inv *invoices.Invoice, vendorID int64, el *EmpLocation, _ []invModel.Inventory) (int64, int64, error) {
	vi := invModel.VendorInvoice{
		InvoiceNo: fmt.Sprintf("Temp-%s", inv.NumberInvoice),
		OrderRef:  "Temporary",
		VendorID:  vendorID,
		InvoiceID: inv.IDInvoice,
	}
	if err := s.db.Create(&vi).Error; err != nil {
		return 0, 0, err
	}
	empID := int64(el.Employee.IDEmployee)
	locID := int64(el.Location.IDLocation)
	shipment := invModel.Shipment{
		VendorID:          vendorID,
		LocationID:        locID,
		EmployeeIDPrepBy:  empID,
		EmployeeIDCreated: empID,
		DateReceived:      inv.DateCreate,
		Status:            strPtr("Pending"),
		Notes:             strPtr("Temporary shipment for vendor invoice"),
		VendorInvoiceID:   &vi.IDVendorInvoice,
	}
	if err := s.db.Create(&shipment).Error; err != nil {
		return 0, 0, err
	}
	return vi.IDVendorInvoice, shipment.IDShipment, nil
}

func strPtr(s string) *string    { return &s }
func int64Ptr(v int64) *int64    { return &v }
