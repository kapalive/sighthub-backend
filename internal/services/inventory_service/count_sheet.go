package inventory_service

import (
	"errors"
	"fmt"
	"time"

	invModel "sighthub-backend/internal/models/inventory"
	"sighthub-backend/pkg/activitylog"
	"sighthub-backend/pkg/sku"
)

// ── GetCountSheets ──────────────────────────────────────────────────────────

func (s *Service) GetCountSheets(username string, brandID, vendorID *int64, dateFrom, dateTo *string) ([]map[string]interface{}, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	locationID := *emp.LocationID

	q := `
		SELECT DISTINCT ic.id_inventory_count, ic.created_date, ic.status,
		       b.brand_name, v.vendor_name,
		       (SELECT COUNT(*) FROM temp_count_inventory tci
		        WHERE tci.inventory_count_id = ic.id_inventory_count
		          AND tci.location_id = ic.location_id AND tci.in_stock = true) AS quantity_found,
		       (SELECT COALESCE(SUM(ms.quantity), 0) FROM missing ms
		        WHERE ms.inventory_count_id = ic.id_inventory_count
		          AND ms.location_id = ic.location_id) AS quantity_missing
		FROM inventory_count ic
		LEFT JOIN brand b ON ic.brand_id = b.id_brand
		LEFT JOIN vendor v ON ic.vendor_id = v.id_vendor
		LEFT JOIN temp_count_inventory tci2 ON tci2.inventory_count_id = ic.id_inventory_count
		WHERE ic.location_id = ?
	`
	args := []interface{}{locationID}

	if brandID != nil {
		q += ` AND tci2.brand_id = ?`
		args = append(args, *brandID)
	}
	if vendorID != nil {
		q += ` AND ic.vendor_id = ?`
		args = append(args, *vendorID)
	}
	if dateFrom != nil {
		q += ` AND ic.created_date >= ?`
		args = append(args, *dateFrom)
	}
	if dateTo != nil {
		q += ` AND ic.created_date <= ?`
		args = append(args, *dateTo)
	}

	type csRow struct {
		IDInventoryCount int64  `gorm:"column:id_inventory_count"`
		CreatedDate      string `gorm:"column:created_date"`
		Status           bool   `gorm:"column:status"`
		BrandName        *string `gorm:"column:brand_name"`
		VendorName       *string `gorm:"column:vendor_name"`
		QuantityFound    int    `gorm:"column:quantity_found"`
		QuantityMissing  int    `gorm:"column:quantity_missing"`
	}
	var rows []csRow
	if err := s.db.Raw(q, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		total := r.QuantityFound + r.QuantityMissing
		out[i] = map[string]interface{}{
			"id_count_sheet": r.IDInventoryCount,
			"created_date":   r.CreatedDate,
			"brand":          r.BrandName,
			"vendor":         r.VendorName,
			"status":         r.Status,
			"quantity":       fmt.Sprintf("%d/%d", r.QuantityFound, total),
		}
	}
	return out, nil
}

// ── CreateCountSheet ────────────────────────────────────────────────────────

type CreateCountSheetInput struct {
	BrandID  *int64  `json:"brand_id"`
	VendorID *int64  `json:"vendor_id"`
	Notes    string  `json:"notes"`
}

func (s *Service) CreateCountSheet(username string, input CreateCountSheetInput) (map[string]interface{}, error) {
	if input.BrandID == nil && input.VendorID == nil {
		return nil, errors.New("either brand_id or vendor_id is required")
	}
	if input.BrandID != nil && input.VendorID != nil {
		return nil, errors.New("provide either brand_id or vendor_id, not both")
	}

	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	locationID := int64(loc.IDLocation)
	employeeID := int64(emp.IDEmployee)

	// Check for existing open count sheet
	existQ := s.db.Where("location_id = ? AND status = true", locationID)
	if input.BrandID != nil {
		existQ = existQ.Where("brand_id = ?", *input.BrandID)
	} else {
		existQ = existQ.Where("vendor_id = ?", *input.VendorID)
	}
	var existing invModel.InventoryCount
	if err := existQ.First(&existing).Error; err == nil {
		label := "brand"
		if input.VendorID != nil {
			label = "vendor"
		}
		return nil, fmt.Errorf("an active count sheet already exists for this %s (id: %d)", label, existing.IDInventoryCount)
	}

	// Get items
	var itemQuery string
	var qArgs []interface{}
	if input.BrandID != nil {
		itemQuery = `
			SELECT i.id_inventory, i.model_id, p.brand_id
			FROM inventory i
			JOIN model m ON i.model_id = m.id_model
			JOIN product p ON m.product_id = p.id_product
			WHERE i.location_id = ? AND i.status_items_inventory = 'Ready for Sale' AND p.brand_id = ?
		`
		qArgs = []interface{}{locationID, *input.BrandID}
	} else {
		itemQuery = `
			SELECT i.id_inventory, i.model_id, p.brand_id
			FROM inventory i
			JOIN model m ON i.model_id = m.id_model
			JOIN product p ON m.product_id = p.id_product
			JOIN invoice inv ON i.invoice_id = inv.id_invoice
			WHERE i.location_id = ? AND i.status_items_inventory = 'Ready for Sale' AND inv.vendor_id = ?
		`
		qArgs = []interface{}{locationID, *input.VendorID}
	}

	type invItem struct {
		IDInventory int64  `gorm:"column:id_inventory"`
		ModelID     int64  `gorm:"column:model_id"`
		BrandID     *int64 `gorm:"column:brand_id"`
	}
	var items []invItem
	if err := s.db.Raw(itemQuery, qArgs...).Scan(&items).Error; err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, errors.New("no products found for the specified filter in this location or they are not Ready for Sale")
	}

	now := time.Now().UTC()
	notes := input.Notes
	cs := invModel.InventoryCount{
		LocationID:        locationID,
		BrandID:           input.BrandID,
		VendorID:          input.VendorID,
		Status:            true,
		PrepByDate:        now,
		PrepByEmployeeID:  employeeID,
		CreatedDate:       now,
		CreatedEmployeeID: employeeID,
		UpdatedDate:       now,
		UpdatedEmployeeID: employeeID,
		Quantity:          0,
		Cost:              0,
		Notes:             &notes,
	}
	if err := s.db.Create(&cs).Error; err != nil {
		return nil, err
	}

	activitylog.Log(s.db, "inventory", "count_sheet_create", activitylog.WithEntity(cs.IDInventoryCount))

	// Create Missing entries for each item
	for _, item := range items {
		// Get item_net from price_book
		var itemNet float64
		s.db.Raw(`SELECT COALESCE(item_net, 0) FROM price_book WHERE inventory_id = ?`, item.IDInventory).Scan(&itemNet)

		s.db.Create(&invModel.Missing{
			InventoryCountID: cs.IDInventoryCount,
			InventoryID:      item.IDInventory,
			LocationID:       locationID,
			BrandID:          item.BrandID,
			VendorID:         input.VendorID,
			ModelID:          item.ModelID,
			Quantity:         1,
			Cost:             itemNet,
			ReportedDate:     now,
			Notes:            strPtr(""),
		})
	}

	cs.Quantity = len(items)
	s.db.Save(&cs)

	// Log transaction
	s.db.Create(&invModel.InventoryTransaction{
		FromLocationID:   &locationID,
		TransferredBy:    employeeID,
		InventoryCountID: &cs.IDInventoryCount,
		StatusItems:      "Initiated",
		TransactionType:  "Count Sheet Opened",
		DateTransaction:  now,
	})

	return map[string]interface{}{
		"id_count_sheet": cs.IDInventoryCount,
		"created_date":   cs.CreatedDate.Format(time.RFC3339),
		"status":         cs.Status,
		"quantity":       fmt.Sprintf("0/%d", len(items)),
		"notes":          cs.Notes,
	}, nil
}

// ── DeleteCountSheet ────────────────────────────────────────────────────────

func (s *Service) DeleteCountSheet(username string, idCountSheet int64) error {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}
	locationID := int64(loc.IDLocation)

	var cs invModel.InventoryCount
	if err := s.db.Where("id_inventory_count = ? AND location_id = ?", idCountSheet, locationID).First(&cs).Error; err != nil {
		return errors.New("count sheet not found or does not belong to your location")
	}

	// Delete temp_count_inventory
	s.db.Where("inventory_count_id = ?", idCountSheet).Delete(&invModel.TempCountInventory{})
	// Delete missing
	s.db.Where("inventory_count_id = ?", idCountSheet).Delete(&invModel.Missing{})
	// Delete related transactions
	s.db.Where("inventory_count_id = ?", idCountSheet).Delete(&invModel.InventoryTransaction{})

	now := time.Now().UTC()
	noteStr := fmt.Sprintf("Deleted CountSheet with ID=%d", idCountSheet)
	s.db.Create(&invModel.InventoryTransaction{
		FromLocationID:  &locationID,
		TransferredBy:   int64(emp.IDEmployee),
		StatusItems:     "Ready for Sale",
		TransactionType: "Count Sheet: Deleted",
		DateTransaction: now,
		Notes:           &noteStr,
	})

	s.db.Delete(&cs)
	activitylog.Log(s.db, "inventory", "count_sheet_delete", activitylog.WithEntity(idCountSheet))
	return nil
}

// ── GetCountSheetInfo ───────────────────────────────────────────────────────

func (s *Service) GetCountSheetInfo(username string, idCountSheet int64) (map[string]interface{}, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	locationID := *emp.LocationID

	var row struct {
		IDInventoryCount int64   `gorm:"column:id_inventory_count"`
		LocationName     string  `gorm:"column:location_name"`
		Quantity         int     `gorm:"column:quantity"`
		Status           bool    `gorm:"column:status"`
		Notes            *string `gorm:"column:notes"`
		CreatedDate      string  `gorm:"column:created_date"`
		CreatedBy        string  `gorm:"column:created_by"`
		VendorID         *int64  `gorm:"column:vendor_id"`
		VendorName       *string `gorm:"column:vendor_name"`
		BrandID          *int64  `gorm:"column:brand_id"`
		BrandName        *string `gorm:"column:brand_name"`
	}
	err = s.db.Raw(`
		SELECT ic.id_inventory_count, l.full_name AS location_name, ic.quantity, ic.status, ic.notes,
		       ic.created_date::text, CONCAT(e.first_name, ' ', e.last_name) AS created_by,
		       ic.vendor_id, v.vendor_name, ic.brand_id, b.brand_name
		FROM inventory_count ic
		JOIN location l ON ic.location_id = l.id_location
		JOIN employee e ON ic.created_employee_id = e.id_employee
		LEFT JOIN vendor v ON ic.vendor_id = v.id_vendor
		LEFT JOIN brand b ON ic.brand_id = b.id_brand
		WHERE ic.id_inventory_count = ? AND ic.location_id = ?
	`, idCountSheet, locationID).Scan(&row).Error
	if err != nil || row.IDInventoryCount == 0 {
		return nil, errors.New("count sheet not found or does not belong to your location")
	}

	// Count items found
	var quantityCounted int
	s.db.Raw(`SELECT COUNT(*) FROM temp_count_inventory WHERE inventory_count_id = ? AND location_id = ? AND in_stock = true`,
		idCountSheet, locationID).Scan(&quantityCounted)

	return map[string]interface{}{
		"id_count_sheet": row.IDInventoryCount,
		"location":       row.LocationName,
		"quantity":       fmt.Sprintf("%d/%d", quantityCounted, row.Quantity),
		"brand_vendor": map[string]interface{}{
			"vendor_id":   row.VendorID,
			"vendor_name": row.VendorName,
			"brand_id":    row.BrandID,
			"brand_name":  row.BrandName,
		},
		"status":      row.Status,
		"notes":       row.Notes,
		"created_date": row.CreatedDate,
		"created_by":  row.CreatedBy,
	}, nil
}

// ── UpdateCountSheetNotes ───────────────────────────────────────────────────

func (s *Service) UpdateCountSheetNotes(username string, idCountSheet int64, notes string) error {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}

	var cs invModel.InventoryCount
	if err := s.db.First(&cs, idCountSheet).Error; err != nil {
		return fmt.Errorf("count sheet with id=%d not found", idCountSheet)
	}
	cs.Notes = &notes
	cs.UpdatedEmployeeID = int64(emp.IDEmployee)
	activitylog.Log(s.db, "inventory", "count_sheet_update", activitylog.WithEntity(idCountSheet))
	return s.db.Save(&cs).Error
}

// ── GetCountSheetItems ──────────────────────────────────────────────────────

func (s *Service) GetCountSheetItems(username string, idCountSheet int64) (map[string]interface{}, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	locationID := *emp.LocationID

	// Verify count sheet exists
	var csExists int64
	s.db.Raw(`SELECT COUNT(*) FROM inventory_count WHERE id_inventory_count = ?`, idCountSheet).Scan(&csExists)
	if csExists == 0 {
		return nil, errors.New("count sheet not found")
	}

	// Counted items
	type itemRow struct {
		InventoryID int64  `gorm:"column:id_inventory"`
		SKU         string `gorm:"column:sku"`
		Description string `gorm:"column:description"`
	}
	var counted []itemRow
	s.db.Raw(`
		SELECT i.id_inventory, i.sku,
		       CONCAT(b.brand_name, ' ', p.title_product, ' ', m.title_variant) AS description
		FROM temp_count_inventory tci
		JOIN inventory i ON tci.inventory_id = i.id_inventory
		JOIN model m ON i.model_id = m.id_model
		JOIN product p ON m.product_id = p.id_product
		JOIN brand b ON p.brand_id = b.id_brand
		WHERE tci.inventory_count_id = ? AND tci.location_id = ?
	`, idCountSheet, locationID).Scan(&counted)

	// Missing items
	var missing []itemRow
	s.db.Raw(`
		SELECT i.id_inventory, i.sku,
		       CONCAT(b.brand_name, ' ', p.title_product, ' ', m.title_variant) AS description
		FROM missing ms
		JOIN inventory i ON ms.inventory_id = i.id_inventory
		JOIN model m ON i.model_id = m.id_model
		JOIN product p ON m.product_id = p.id_product
		JOIN brand b ON p.brand_id = b.id_brand
		WHERE ms.inventory_count_id = ? AND ms.location_id = ?
	`, idCountSheet, locationID).Scan(&missing)

	countedList := make([]map[string]interface{}, len(counted))
	for i, c := range counted {
		countedList[i] = map[string]interface{}{
			"inventory_id": c.InventoryID,
			"sku":          c.SKU,
			"description":  c.Description,
		}
	}
	missingList := make([]map[string]interface{}, len(missing))
	for i, m := range missing {
		missingList[i] = map[string]interface{}{
			"inventory_id": m.InventoryID,
			"sku":          m.SKU,
			"description":  m.Description,
		}
	}

	// Location name
	var locationName string
	s.db.Raw(`SELECT full_name FROM location WHERE id_location = ?`, locationID).Scan(&locationName)

	// Brand/vendor info from first counted item
	var brandInfo, vendorInfo interface{}
	s.db.Raw(`
		SELECT b.id_brand, b.brand_name FROM temp_count_inventory tci
		JOIN inventory i ON tci.inventory_id = i.id_inventory
		JOIN model m ON i.model_id = m.id_model
		JOIN product p ON m.product_id = p.id_product
		JOIN brand b ON p.brand_id = b.id_brand
		WHERE tci.inventory_count_id = ? AND tci.location_id = ?
		LIMIT 1
	`, idCountSheet, locationID).Scan(&brandInfo)

	total := len(counted) + len(missing)
	return map[string]interface{}{
		"location":       locationName,
		"id_count_sheet": idCountSheet,
		"brand":          brandInfo,
		"vendor":         vendorInfo,
		"quantity":       fmt.Sprintf("%d/%d", len(counted), total),
		"counted_items":  countedList,
		"missing_items":  missingList,
	}, nil
}

// ── AddItemToCountSheet ─────────────────────────────────────────────────────

func (s *Service) AddItemToCountSheet(username string, idCountSheet int64, rawSKU string) (map[string]interface{}, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	locationID := *emp.LocationID

	// Verify count sheet exists
	var cs invModel.InventoryCount
	if err := s.db.First(&cs, idCountSheet).Error; err != nil {
		return nil, errors.New("count sheet not found")
	}

	// Find inventory item
	normalized := sku.Normalize(rawSKU)
	var inv invModel.Inventory
	if err := s.db.Where("sku = ? AND location_id = ?", normalized, locationID).First(&inv).Error; err != nil {
		return nil, errors.New("item not found or does not belong to your location")
	}

	// Check if already counted
	var tempExists int64
	s.db.Raw(`SELECT COUNT(*) FROM temp_count_inventory WHERE inventory_id = ? AND location_id = ? AND inventory_count_id = ?`,
		inv.IDInventory, locationID, idCountSheet).Scan(&tempExists)
	if tempExists > 0 {
		return nil, errors.New("this item has already been found and added to the inventory sheet")
	}

	// Check if in missing
	var missingItem invModel.Missing
	if err := s.db.Where("inventory_count_id = ? AND inventory_id = ? AND location_id = ?",
		idCountSheet, inv.IDInventory, locationID).First(&missingItem).Error; err != nil {
		return nil, errors.New("product not found in the out of stock list for this Count Sheet")
	}

	now := time.Now().UTC()

	// Get brand_id from product
	var brandID *int64
	s.db.Raw(`SELECT p.brand_id FROM model m JOIN product p ON m.product_id = p.id_product WHERE m.id_model = ?`, inv.ModelID).Scan(&brandID)

	brandIDVal := 0
	if brandID != nil {
		brandIDVal = int(*brandID)
	}

	// Move from Missing to TempCountInventory
	s.db.Create(&invModel.TempCountInventory{
		InventoryID:      inv.IDInventory,
		LocationID:       int(locationID),
		BrandID:          brandIDVal,
		InStock:          true,
		InventoryCountID: cs.IDInventoryCount,
		CountDate:        now,
	})

	inventoryID := inv.IDInventory
	s.db.Create(&invModel.InventoryTransaction{
		InventoryID:      &inventoryID,
		FromLocationID:   &locationID,
		TransferredBy:    int64(emp.IDEmployee),
		InventoryCountID: &cs.IDInventoryCount,
		StatusItems:      "Ready for Sale",
		TransactionType:  "Count Sheet: Counted",
		DateTransaction:  now,
	})

	s.db.Delete(&missingItem)

	// Return updated items
	return s.GetCountSheetItems(username, idCountSheet)
}

// ── DeleteItemFromCountSheet ────────────────────────────────────────────────

func (s *Service) DeleteItemFromCountSheet(username string, idCountSheet int64, itemID int64) (map[string]interface{}, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	locationID := *emp.LocationID

	var cs invModel.InventoryCount
	if err := s.db.First(&cs, idCountSheet).Error; err != nil {
		return nil, errors.New("count sheet not found")
	}

	var inv invModel.Inventory
	if err := s.db.Where("id_inventory = ? AND location_id = ?", itemID, locationID).First(&inv).Error; err != nil {
		return nil, errors.New("item not found or does not belong to your location")
	}

	var tempItem invModel.TempCountInventory
	if err := s.db.Where("inventory_count_id = ? AND inventory_id = ? AND location_id = ?",
		idCountSheet, itemID, locationID).First(&tempItem).Error; err != nil {
		return nil, errors.New("item not found in counted items for this count sheet")
	}

	// Get brand_id
	var brandID *int64
	s.db.Raw(`SELECT p.brand_id FROM model m JOIN product p ON m.product_id = p.id_product WHERE m.id_model = ?`, inv.ModelID).Scan(&brandID)

	// Get item_net
	var itemNet float64
	s.db.Raw(`SELECT COALESCE(item_net, 0) FROM price_book WHERE inventory_id = ?`, itemID).Scan(&itemNet)

	now := time.Now().UTC()
	emptyNotes := ""

	// Move to missing
	s.db.Create(&invModel.Missing{
		InventoryCountID: idCountSheet,
		InventoryID:      inv.IDInventory,
		LocationID:       locationID,
		BrandID:          brandID,
		VendorID:         cs.VendorID,
		ModelID:          safeInt64(inv.ModelID),
		Quantity:         1,
		Cost:             itemNet,
		ReportedDate:     now,
		Notes:            &emptyNotes,
	})

	inventoryID := inv.IDInventory
	s.db.Create(&invModel.InventoryTransaction{
		InventoryID:      &inventoryID,
		FromLocationID:   &locationID,
		TransferredBy:    int64(emp.IDEmployee),
		InventoryCountID: &cs.IDInventoryCount,
		StatusItems:      inv.StatusItemsInventory,
		TransactionType:  "Count Sheet: Removed",
		DateTransaction:  now,
	})

	s.db.Delete(&tempItem)

	return s.GetCountSheetItems(username, idCountSheet)
}

// ── CloseCountSheet ─────────────────────────────────────────────────────────

func (s *Service) CloseCountSheet(username string, idCountSheet int64) (map[string]interface{}, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	locationID := int64(loc.IDLocation)

	var cs invModel.InventoryCount
	if err := s.db.Where("id_inventory_count = ? AND location_id = ?", idCountSheet, locationID).First(&cs).Error; err != nil {
		return nil, errors.New("count sheet not found or does not belong to your location")
	}
	if !cs.Status {
		return nil, errors.New("this count sheet is already closed")
	}

	cs.Status = false
	activitylog.Log(s.db, "inventory", "count_sheet_close", activitylog.WithEntity(idCountSheet))
	s.db.Save(&cs)

	// Mark remaining missing items as 'Missing' in inventory
	var missingItems []invModel.Missing
	s.db.Where("inventory_count_id = ? AND location_id = ?", idCountSheet, locationID).Find(&missingItems)

	now := time.Now().UTC()
	for _, mi := range missingItems {
		s.db.Model(&invModel.Inventory{}).Where("id_inventory = ?", mi.InventoryID).
			Update("status_items_inventory", "Missing")

		invID := mi.InventoryID
		s.db.Create(&invModel.InventoryTransaction{
			InventoryID:      &invID,
			FromLocationID:   &locationID,
			TransferredBy:    int64(emp.IDEmployee),
			InventoryCountID: &cs.IDInventoryCount,
			StatusItems:      "Missing",
			TransactionType:  "Marked as Missing",
			DateTransaction:  now,
		})
	}

	return map[string]interface{}{
		"message":        "Count sheet successfully closed",
		"id_count_sheet": idCountSheet,
	}, nil
}
