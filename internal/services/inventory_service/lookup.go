package inventory_service

import (
	"errors"
	"fmt"
	"time"

	invModel "sighthub-backend/internal/models/inventory"
	vendorModel "sighthub-backend/internal/models/vendors"
	"sighthub-backend/pkg/activitylog"
	"sighthub-backend/pkg/sku"
)

// ── LookupBySKU ─────────────────────────────────────────────────────────────

func (s *Service) LookupBySKU(rawSKU string) (map[string]interface{}, error) {
	normalized := sku.Normalize(rawSKU)

	var row struct {
		IDInventory          int64   `gorm:"column:id_inventory"`
		SKU                  string  `gorm:"column:sku"`
		StatusItemsInventory string  `gorm:"column:status_items_inventory"`
		LocationName         string  `gorm:"column:location_name"`
		Sunglass             *bool   `gorm:"column:sunglass"`
		Photo                *string `gorm:"column:photo"`
		BrandName            string  `gorm:"column:brand_name"`
		TitleProduct         string  `gorm:"column:title_product"`
		TitleVariant         string  `gorm:"column:title_variant"`
		Polor                *string `gorm:"column:polor"`
		SizeLensWidth        *string `gorm:"column:size_lens_width"`
		SizeBridgeWidth      *string `gorm:"column:size_bridge_width"`
		SizeTempleLength     *string `gorm:"column:size_temple_length"`
		Gtin                 *string `gorm:"column:gtin"`
		Upc                  *string `gorm:"column:upc"`
		Ean                  *string `gorm:"column:ean"`
		MfgNumber            *string `gorm:"column:mfg_number"`
		MfrSerialNumber      *string `gorm:"column:mfr_serial_number"`
		Accessories          *string `gorm:"column:accessories"`
		// PriceBook fields
		ItemListCost     *float64 `gorm:"column:item_list_cost"`
		ItemDiscount     *float64 `gorm:"column:item_discount"`
		ItemNet          *float64 `gorm:"column:item_net"`
		PbListCost       *float64 `gorm:"column:pb_list_cost"`
		PbDiscount       *float64 `gorm:"column:pb_discount"`
		PbCost           *float64 `gorm:"column:pb_cost"`
		PbSellingPrice   *float64 `gorm:"column:pb_selling_price"`
		PbStoreTierPrice *float64 `gorm:"column:pb_store_tier_price"`
		LensCost         *float64 `gorm:"column:lens_cost"`
		AccessoriesCost  *float64 `gorm:"column:accessories_cost"`
		Note             *string  `gorm:"column:note"`
	}
	err := s.db.Raw(`
		SELECT i.id_inventory, i.sku, i.status_items_inventory,
		       l.full_name AS location_name,
		       m.sunglass, m.photo, m.title_variant, m.polor,
		       m.size_lens_width, m.size_bridge_width, m.size_temple_length,
		       m.gtin, m.upc, m.ean, m.mfg_number, m.mfr_serial_number, m.accessories,
		       b.brand_name, p.title_product,
		       pb.item_list_cost, pb.item_discount, pb.item_net,
		       pb.pb_list_cost, pb.pb_discount, pb.pb_cost,
		       pb.pb_selling_price, pb.pb_store_tier_price,
		       pb.lens_cost, pb.accessories_cost, pb.note
		FROM inventory i
		JOIN location l ON i.location_id = l.id_location
		JOIN model m ON i.model_id = m.id_model
		JOIN product p ON m.product_id = p.id_product
		JOIN brand b ON p.brand_id = b.id_brand
		LEFT JOIN price_book pb ON pb.inventory_id = i.id_inventory
		WHERE i.sku = ?
	`, normalized).Scan(&row).Error
	if err != nil {
		return nil, errors.New("item not found")
	}
	if row.IDInventory == 0 {
		return nil, errors.New("item not found")
	}

	return map[string]interface{}{
		"id_inventory":       row.IDInventory,
		"sku":                row.SKU,
		"defective":          row.StatusItemsInventory == "Defective",
		"location":           row.LocationName,
		"status":             row.StatusItemsInventory,
		"sunglass":           row.Sunglass,
		"photo":              row.Photo,
		"brand_name":         row.BrandName,
		"title_product":      row.TitleProduct,
		"title_variant":      row.TitleVariant,
		"polor":              row.Polor,
		"size_lens_width":    row.SizeLensWidth,
		"size_bridge_width":  row.SizeBridgeWidth,
		"size_temple_length": row.SizeTempleLength,
		"item_list_cost":     row.ItemListCost,
		"item_discount":      row.ItemDiscount,
		"item_net":           row.ItemNet,
		"pb_list_cost":       row.PbListCost,
		"pb_discount":        row.PbDiscount,
		"pb_cost":            row.PbCost,
		"pb_selling_price":   row.PbSellingPrice,
		"pb_store_tier_price": row.PbStoreTierPrice,
		"lens_cost":          row.LensCost,
		"accessories_cost":   row.AccessoriesCost,
		"gtin":               row.Gtin,
		"upc":                row.Upc,
		"ean":                row.Ean,
		"mfg_number":         row.MfgNumber,
		"mfr_serial_number":  row.MfrSerialNumber,
		"accessories":        row.Accessories,
		"notes":              row.Note,
	}, nil
}

// ── GetInventoryHistory ─────────────────────────────────────────────────────

func (s *Service) GetInventoryHistory(inventoryID int64, rawSKU string) (map[string]interface{}, error) {
	// Resolve inventory_id from SKU if needed
	if rawSKU != "" {
		normalized := sku.Normalize(rawSKU)
		var inv invModel.Inventory
		if err := s.db.Where("sku = ?", normalized).First(&inv).Error; err != nil {
			return nil, errors.New("item with this SKU not found")
		}
		inventoryID = inv.IDInventory
	}
	if inventoryID == 0 {
		return nil, errors.New("specify inventory_id or sku")
	}

	// General info
	var inv invModel.Inventory
	if err := s.db.First(&inv, inventoryID).Error; err != nil {
		return nil, errors.New("item not found")
	}

	var locationName string
	s.db.Raw(`SELECT full_name FROM location WHERE id_location = ?`, inv.LocationID).Scan(&locationName)

	inStock := "N"
	if string(inv.StatusItemsInventory) == "Ready for Sale" {
		inStock = "Y"
	}
	defective := false
	if string(inv.StatusItemsInventory) == "Defective" {
		defective = true
	}

	// Receipt info
	var receiptInfo struct {
		NumberInvoice string `gorm:"column:number_invoice"`
		DateTime      string `gorm:"column:receipt_datetime"`
	}
	s.db.Raw(`
		SELECT inv.number_invoice, ri.datetime::text AS receipt_datetime
		FROM receipts_items ri
		JOIN invoice inv ON ri.invoice_id = inv.id_invoice
		WHERE ri.inventory_id = ?
		LIMIT 1
	`, inventoryID).Scan(&receiptInfo)

	// Patient info
	var patientInfo struct {
		Patient        *string `gorm:"column:patient_name"`
		PatientInvoice *string `gorm:"column:number_invoice"`
	}
	s.db.Raw(`
		SELECT CONCAT(pt.first_name, ' ', pt.last_name) AS patient_name,
		       inv.number_invoice
		FROM invoice inv
		JOIN patient pt ON inv.patient_id = pt.id_patient
		WHERE inv.id_invoice = (SELECT invoice_id FROM inventory WHERE id_inventory = ?)
		  AND inv.patient_id IS NOT NULL
		LIMIT 1
	`, inventoryID).Scan(&patientInfo)

	// Order info
	var orderNumber *string
	s.db.Raw(`
		SELECT ol.number_order
		FROM invoice_services_item isi
		JOIN orders_lens ol ON isi.lens_order_id = ol.id_orders_lens
		WHERE isi.invoice_id = (SELECT invoice_id FROM inventory WHERE id_inventory = ?)
		LIMIT 1
	`, inventoryID).Scan(&orderNumber)

	generalInfo := map[string]interface{}{
		"inventory_id": inv.IDInventory,
		"sku":          inv.SKU,
		"defective":    defective,
		"location":     locationName,
		"status":       inv.StatusItemsInventory,
		"in_stock":     inStock,
		"receipt_add": map[string]interface{}{
			"number": receiptInfo.NumberInvoice,
			"date":   receiptInfo.DateTime,
		},
		"patient": map[string]interface{}{
			"patient":         patientInfo.Patient,
			"patient_invoice": patientInfo.PatientInvoice,
			"order":           orderNumber,
			"request":         nil,
		},
	}

	// Transaction history
	type txRow struct {
		TransactionID    int64   `gorm:"column:id_transaction"`
		StatusItems      string  `gorm:"column:status_items"`
		TransactionType  string  `gorm:"column:transaction_type"`
		DateTransaction  string  `gorm:"column:date_transaction"`
		InvoiceID        *int64  `gorm:"column:invoice_id"`
		InventoryCountID *int64  `gorm:"column:inventory_count_id"`
		FromLocation     *string `gorm:"column:from_location"`
		ToLocation       *string `gorm:"column:to_location"`
		EmployeeName     string  `gorm:"column:employee_name"`
		VendorName       *string `gorm:"column:vendor_name"`
	}
	var txRows []txRow
	s.db.Raw(`
		SELECT it.id_transaction, it.status_items, it.transaction_type,
		       it.date_transaction::text AS date_transaction,
		       it.invoice_id, it.inventory_count_id,
		       fl.short_name AS from_location,
		       tl.short_name AS to_location,
		       CONCAT(e.first_name, ' ', e.last_name) AS employee_name,
		       v.vendor_name
		FROM inventory_transaction it
		JOIN employee e ON it.transferred_by = e.id_employee
		LEFT JOIN location fl ON it.from_location_id = fl.id_location
		LEFT JOIN location tl ON it.to_location_id = tl.id_location
		LEFT JOIN invoice inv ON it.invoice_id = inv.id_invoice
		LEFT JOIN vendor v ON inv.vendor_id = v.id_vendor
		WHERE it.inventory_id = ?
		ORDER BY it.date_transaction DESC
	`, inventoryID).Scan(&txRows)

	txHistory := make([]map[string]interface{}, len(txRows))
	for i, tx := range txRows {
		fromLoc := "Vendor"
		if tx.FromLocation != nil {
			fromLoc = *tx.FromLocation
		} else if tx.VendorName != nil {
			fromLoc = *tx.VendorName
		}
		toLoc := "N/A"
		if tx.ToLocation != nil {
			toLoc = *tx.ToLocation
		}

		extra := "DELETED"
		if tx.InventoryCountID != nil {
			extra = fmt.Sprintf("Count ID: %d", *tx.InventoryCountID)
		} else if tx.InvoiceID != nil {
			extra = fmt.Sprintf("Invoice ID: %d", *tx.InvoiceID)
		}

		txHistory[i] = map[string]interface{}{
			"transaction_id":     tx.TransactionID,
			"status":             tx.StatusItems,
			"action":             tx.TransactionType,
			"date":               tx.DateTransaction,
			"invoice_id":         tx.InvoiceID,
			"inventory_count_id": tx.InventoryCountID,
			"rep":                tx.EmployeeName,
			"location_info": map[string]interface{}{
				"from_location": fromLoc,
				"to_location":   toLoc,
			},
			"extra": extra,
			"stock": inStock,
		}
	}

	return map[string]interface{}{
		"general_info":        generalInfo,
		"transaction_history": txHistory,
	}, nil
}

// ── UpdatePrice ─────────────────────────────────────────────────────────────

type UpdatePriceInput struct {
	ItemListCost   *float64 `json:"item_list_cost"`
	ItemDiscount   *float64 `json:"item_discount"`
	ItemNet        *float64 `json:"item_net"`
	PbSellingPrice *float64 `json:"pb_selling_price"`
	LensCost       *float64 `json:"lens_cost"`
	AccessoriesCost *float64 `json:"accessories_cost"`
	Note           *string  `json:"note"`
}

func (s *Service) UpdatePrice(inventoryID int64, input UpdatePriceInput) (map[string]interface{}, error) {
	var pb invModel.PriceBook
	if err := s.db.Where("inventory_id = ?", inventoryID).First(&pb).Error; err != nil {
		return nil, errors.New("PriceBook entry not found for the provided inventory_id")
	}

	if input.ItemListCost != nil {
		pb.ItemListCost = input.ItemListCost
	}
	if input.ItemDiscount != nil {
		pb.ItemDiscount = input.ItemDiscount
	}
	if input.ItemNet != nil {
		pb.ItemNet = input.ItemNet
	}
	if input.PbSellingPrice != nil {
		pb.PbSellingPrice = input.PbSellingPrice
	}
	if input.LensCost != nil {
		pb.LensCost = input.LensCost
	}
	if input.AccessoriesCost != nil {
		pb.AccessoriesCost = input.AccessoriesCost
	}
	if input.Note != nil {
		pb.Note = input.Note
	}

	activitylog.Log(s.db, "inventory", "price_update",
		activitylog.WithEntity(inventoryID),
		activitylog.WithDetails(map[string]interface{}{
			"item_list_cost":   fmtPrice(pb.ItemListCost),
			"pb_selling_price": fmtPrice(pb.PbSellingPrice),
		}),
	)
	if err := s.db.Save(&pb).Error; err != nil {
		return nil, fmt.Errorf("failed to update prices: %w", err)
	}

	return map[string]interface{}{
		"message":          "Price updated successfully",
		"inventory_id":     inventoryID,
		"item_list_cost":   fmtPrice(pb.ItemListCost),
		"item_discount":    fmtPrice(pb.ItemDiscount),
		"item_net":         fmtPrice(pb.ItemNet),
		"pb_selling_price": fmtPrice(pb.PbSellingPrice),
		"lens_cost":        fmtPrice(pb.LensCost),
		"accessories_cost": fmtPrice(pb.AccessoriesCost),
		"note":             pb.Note,
	}, nil
}

// ── UpdateInventoryState ────────────────────────────────────────────────────

type StateChangeInput struct {
	InventoryID                    *int64  `json:"inventory_id"`
	SKU                            *string `json:"sku"`
	StateDefective                 bool    `json:"state_defective"`
	StateReturnVendor              bool    `json:"state_return_vendor"`
	StateReturnInStoreNoDefect     bool    `json:"state_return_in_store_no_defect"`
	StateRemove                    bool    `json:"state_remove"`
	StateRemoveOther               bool    `json:"state_remove_other"`
	StateReturnInStock             bool    `json:"state_return_in_stock"`
	StateInStock                   bool    `json:"state_in_stock"`
	StateReturnInLocationNoDefect  bool    `json:"state_return_in_location_no_defect"`
	InventoryCountID               *int64  `json:"inventory_count_id"`
	Reason                         *string `json:"reason"`
	CreditAmount                   *string `json:"credit_amount"`
	Note                           *string `json:"note"`
}

func (s *Service) UpdateInventoryState(username string, input StateChangeInput) (map[string]interface{}, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	employeeID := int64(emp.IDEmployee)
	currentLocationID := loc.IDLocation

	// Resolve inventory item
	var inv invModel.Inventory
	if input.SKU != nil && *input.SKU != "" {
		normalized := sku.Normalize(*input.SKU)
		if err := s.db.Where("sku = ?", normalized).First(&inv).Error; err != nil {
			return nil, errors.New("item with this SKU not found")
		}
	} else if input.InventoryID != nil {
		if err := s.db.First(&inv, *input.InventoryID).Error; err != nil {
			return nil, errors.New("item with this inventory_id not found")
		}
	} else {
		return nil, errors.New("specify either inventory_id or sku")
	}

	now := time.Now().UTC()
	inventoryID := inv.IDInventory

	// ── Defective ───────────────────────────────────────────────────────
	if input.StateDefective {
		prevStatus := string(inv.StatusItemsInventory)
		inv.StatusItemsInventory = "Defective"
		s.db.Save(&inv)

		noteStr := ""
		if input.Note != nil {
			noteStr = *input.Note
		} else {
			noteStr = fmt.Sprintf("Status changed from %s to Defective", prevStatus)
		}
		s.db.Create(&invModel.InventoryTransaction{
			InventoryID:     &inventoryID,
			FromLocationID:  &inv.LocationID,
			ToLocationID:    &inv.LocationID,
			TransferredBy:   employeeID,
			InvoiceID:       &inv.InvoiceID,
			OldInvoiceID:    &inv.InvoiceID,
			StatusItems:     "Defective",
			TransactionType: "Status Change",
			DateTransaction: now,
			Notes:           &noteStr,
		})
		activitylog.Log(s.db, "inventory", "state_update", activitylog.WithEntity(inventoryID))
		return map[string]interface{}{"message": "Item reclassified as defective"}, nil
	}

	// ── Return to vendor ────────────────────────────────────────────────
	if input.StateReturnVendor {
		reason := "Return to vendor"
		if input.Reason != nil {
			reason = *input.Reason
		}
		inv.StatusItemsInventory = "On Return"
		s.db.Save(&inv)

		noteStr := fmt.Sprintf("Return to vendor. Reason: %s", reason)
		if input.Note != nil {
			noteStr = *input.Note
		}

		// Create return invoice
		var vendorID *int64
		s.db.Raw(`
			SELECT p.vendor_id FROM inventory i
			JOIN model m ON i.model_id = m.id_model
			JOIN product p ON m.product_id = p.id_product
			WHERE i.id_inventory = ?
		`, inventoryID).Scan(&vendorID)

		if vendorID != nil {
			rtv := vendorModel.ReturnToVendorInvoice{VendorID: *vendorID}
			s.db.Create(&rtv)
			s.db.Create(&vendorModel.ReturnToVendorItem{
				ReturnToVendorInvoiceID: rtv.IDReturnToVendorInvoice,
				InventoryID:             inventoryID,
				ReasonReturn:            reason,
			})
		}

		s.db.Create(&invModel.InventoryTransaction{
			InventoryID:     &inventoryID,
			FromLocationID:  &inv.LocationID,
			TransferredBy:   employeeID,
			InvoiceID:       &inv.InvoiceID,
			OldInvoiceID:    &inv.InvoiceID,
			StatusItems:     "On Return",
			TransactionType: "ReturnToVendor",
			DateTransaction: now,
			Notes:           &noteStr,
		})
		activitylog.Log(s.db, "inventory", "state_update", activitylog.WithEntity(inventoryID))
		return map[string]interface{}{"message": "Item status set to 'On Return'"}, nil
	}

	// ── Return in store (no defect) ─────────────────────────────────────
	if input.StateReturnInStoreNoDefect {
		prevStatus := string(inv.StatusItemsInventory)
		inv.StatusItemsInventory = "Ready for Sale"
		s.db.Save(&inv)

		noteStr := fmt.Sprintf("Returned to store (no defect). Prev status: %s", prevStatus)
		if input.Note != nil {
			noteStr = *input.Note
		}
		s.db.Create(&invModel.InventoryTransaction{
			InventoryID:     &inventoryID,
			FromLocationID:  &inv.LocationID,
			ToLocationID:    &inv.LocationID,
			TransferredBy:   employeeID,
			InvoiceID:       &inv.InvoiceID,
			OldInvoiceID:    &inv.InvoiceID,
			StatusItems:     "Ready for Sale",
			TransactionType: "Status Change",
			DateTransaction: now,
			Notes:           &noteStr,
		})
		activitylog.Log(s.db, "inventory", "state_update", activitylog.WithEntity(inventoryID))
		return map[string]interface{}{"message": "Item returned to store and set to 'Ready for Sale'"}, nil
	}

	// ── Remove (requires open count sheet) ──────────────────────────────
	if input.StateRemove {
		invCountID := input.InventoryCountID
		if invCountID == nil {
			var csID *int64
			s.db.Raw(`
				SELECT ic.id_inventory_count
				FROM inventory_count ic
				JOIN temp_count_inventory tci ON tci.inventory_count_id = ic.id_inventory_count
				WHERE tci.inventory_id = ? AND tci.location_id = ? AND ic.status = true
				ORDER BY ic.id_inventory_count DESC LIMIT 1
			`, inventoryID, currentLocationID).Scan(&csID)
			invCountID = csID
		}
		if invCountID == nil || *invCountID == 0 {
			return nil, errors.New("open count sheet not found for this inventory at this location")
		}

		noteStr := "Item removed from inventory"
		if input.Note != nil {
			noteStr = *input.Note
		}
		s.db.Create(&invModel.Missing{
			InventoryCountID: *invCountID,
			InventoryID:      inventoryID,
			LocationID:       inv.LocationID,
			ModelID:          safeInt64(inv.ModelID),
			Quantity:         1,
			ReportedDate:     now,
			Notes:            &noteStr,
		})
		s.db.Create(&invModel.InventoryTransaction{
			InventoryID:      &inventoryID,
			FromLocationID:   &inv.LocationID,
			TransferredBy:    employeeID,
			InvoiceID:        &inv.InvoiceID,
			OldInvoiceID:     &inv.InvoiceID,
			StatusItems:      "Missing",
			TransactionType:  "Removal",
			InventoryCountID: invCountID,
			Notes:            &noteStr,
		})
		s.db.Delete(&inv)
		activitylog.Log(s.db, "inventory", "state_update", activitylog.WithEntity(inventoryID))
		return map[string]interface{}{
			"message":            "Item removed from inventory",
			"inventory_count_id": *invCountID,
		}, nil
	}

	// ── Remove other ────────────────────────────────────────────────────
	if input.StateRemoveOther {
		if string(inv.StatusItemsInventory) == "Removed" {
			return nil, errors.New("item is already marked as 'Removed'")
		}
		inv.StatusItemsInventory = "Removed"
		s.db.Save(&inv)
		s.db.Create(&invModel.InventoryTransaction{
			InventoryID:     &inventoryID,
			FromLocationID:  &inv.LocationID,
			TransferredBy:   employeeID,
			OldInvoiceID:    &inv.InvoiceID,
			StatusItems:     "Removed",
			TransactionType: "Status Change",
			DateTransaction: now,
		})
		activitylog.Log(s.db, "inventory", "state_update", activitylog.WithEntity(inventoryID))
		return map[string]interface{}{"message": "Item status updated to 'Removed' successfully"}, nil
	}

	// ── Return in stock ─────────────────────────────────────────────────
	if input.StateReturnInStock {
		// Clean Missing if exists
		s.db.Where("inventory_id = ? AND location_id = ?", inventoryID, inv.LocationID).Delete(&invModel.Missing{})

		prevStatus := string(inv.StatusItemsInventory)

		// Determine restore location
		var lastFromLoc *int64
		s.db.Raw(`
			SELECT from_location_id FROM inventory_transaction
			WHERE inventory_id = ? AND status_items = 'Removed'
			ORDER BY date_transaction DESC, id_transaction DESC LIMIT 1
		`, inventoryID).Scan(&lastFromLoc)
		restoreLoc := inv.LocationID
		if lastFromLoc != nil && *lastFromLoc > 0 {
			restoreLoc = *lastFromLoc
		}

		inv.StatusItemsInventory = "Ready for Sale"
		s.db.Save(&inv)

		noteStr := fmt.Sprintf("Returned from %s", prevStatus)
		if input.Note != nil {
			noteStr = *input.Note
		}
		s.db.Create(&invModel.InventoryTransaction{
			InventoryID:     &inventoryID,
			FromLocationID:  &restoreLoc,
			ToLocationID:    &restoreLoc,
			TransferredBy:   employeeID,
			InvoiceID:       &inv.InvoiceID,
			OldInvoiceID:    &inv.InvoiceID,
			StatusItems:     "Ready for Sale",
			TransactionType: "Returned to Stock",
			Notes:           &noteStr,
			DateTransaction: now,
		})
		activitylog.Log(s.db, "inventory", "state_update", activitylog.WithEntity(inventoryID))
		return map[string]interface{}{"message": "Item successfully returned to stock"}, nil
	}

	// ── In stock (transfer to warehouse) ────────────────────────────────
	if input.StateInStock {
		return s.transferToWarehouse(inv, employeeID, input.Note)
	}

	// ── Return in location no defect ────────────────────────────────────
	if input.StateReturnInLocationNoDefect {
		if string(inv.StatusItemsInventory) != "Defective" {
			return nil, errors.New("item is not defective")
		}
		return s.transferToWarehouse(inv, employeeID, input.Note)
	}

	return nil, errors.New("no valid state action specified")
}

func (s *Service) transferToWarehouse(inv invModel.Inventory, employeeID int64, note *string) (map[string]interface{}, error) {
	type locInfo struct {
		WarehouseID         *int64  `gorm:"column:warehouse_id"`
		WarehouseLocationID *int64  `gorm:"column:warehouse_location_id"`
		ShortName           string  `gorm:"column:short_name"`
		CanReceiveItems     *bool   `gorm:"column:can_receive_items"`
	}
	var loc locInfo
	s.db.Raw(`SELECT warehouse_id, warehouse_location_id, short_name, can_receive_items FROM location WHERE id_location = ?`, inv.LocationID).Scan(&loc)

	now := time.Now().UTC()
	inventoryID := inv.IDInventory

	if loc.WarehouseID != nil && *loc.WarehouseID > 0 {
		// Get item net price
		var itemNet *float64
		s.db.Raw(`SELECT item_net FROM price_book WHERE inventory_id = ?`, inventoryID).Scan(&itemNet)
		netVal := 0.00
		if itemNet != nil {
			netVal = *itemNet
		}

		// Create transfer invoice
		invNumber := fmt.Sprintf("I-%s-%s", loc.ShortName, time.Now().Format("20060102150405"))
		var noteStr *string
		if note != nil {
			noteStr = note
		}
		result := s.db.Exec(`
			INSERT INTO invoice (number_invoice, date_create, discount, total_amount, final_amount, employee_id, location_id, to_location_id, notes)
			VALUES (?, CURRENT_TIMESTAMP, 0.00, ?, ?, ?, ?, ?, ?)
		`, invNumber, netVal, netVal, employeeID, inv.LocationID, *loc.WarehouseID, noteStr)
		if result.Error != nil {
			return nil, result.Error
		}

		// Get the created invoice ID
		var newInvoiceID int64
		s.db.Raw(`SELECT id_invoice FROM invoice WHERE number_invoice = ? ORDER BY id_invoice DESC LIMIT 1`, invNumber).Scan(&newInvoiceID)

		// Log transfer
		s.db.Create(&invModel.InventoryTransfer{
			InventoryID:    inventoryID,
			FromLocationID: inv.LocationID,
			ToLocationID:   *loc.WarehouseID,
			TransferredBy:  &employeeID,
			StatusItems:    inv.StatusItemsInventory,
			InvoiceID:      newInvoiceID,
			SystemNote:     strPtr("Automatic transfer to warehouse"),
		})

		// Update inventory
		warehouseLocID := *loc.WarehouseID
		if loc.WarehouseLocationID != nil {
			warehouseLocID = *loc.WarehouseLocationID
		}
		s.db.Model(&inv).Updates(map[string]interface{}{
			"location_id":            warehouseLocID,
			"status_items_inventory": "Ready for Sale",
			"invoice_id":            newInvoiceID,
		})

		s.db.Create(&invModel.InventoryTransaction{
			InventoryID:     &inventoryID,
			FromLocationID:  &inv.LocationID,
			ToLocationID:    loc.WarehouseID,
			TransferredBy:   employeeID,
			InvoiceID:       &newInvoiceID,
			StatusItems:     "Ready for Sale",
			TransactionType: "Transfer to Warehouse",
			DateTransaction: now,
		})

		activitylog.Log(s.db, "inventory", "state_update", activitylog.WithEntity(inventoryID))
		return map[string]interface{}{"message": "Inventory state updated successfully"}, nil
	}

	// No warehouse — if can_receive_items, just change status
	if loc.CanReceiveItems != nil && *loc.CanReceiveItems {
		inv.StatusItemsInventory = "Ready for Sale"
		s.db.Save(&inv)
		s.db.Create(&invModel.InventoryTransaction{
			InventoryID:     &inventoryID,
			FromLocationID:  &inv.LocationID,
			ToLocationID:    &inv.LocationID,
			TransferredBy:   employeeID,
			StatusItems:     "Ready for Sale",
			TransactionType: "Status Change",
			DateTransaction: now,
		})
		activitylog.Log(s.db, "inventory", "state_update", activitylog.WithEntity(inventoryID))
		return map[string]interface{}{"message": "Inventory state updated successfully"}, nil
	}

	return nil, errors.New("item is already in stock or no warehouse available")
}

func safeInt64(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}
