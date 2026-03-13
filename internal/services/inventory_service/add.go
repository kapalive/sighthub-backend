package inventory_service

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	invModel "sighthub-backend/internal/models/inventory"
	"sighthub-backend/pkg/activitylog"
)

type AddInventoryInput struct {
	LocationID      *int64   `json:"location_id"`
	ModelID         int64    `json:"model_id"`
	InvoiceID       int64    `json:"invoice_id"`
	Count           int      `json:"count"`
	ItemListCost    float64  `json:"item_list_cost"`
	ItemDiscount    float64  `json:"item_discount"`
	PbSellingPrice  float64  `json:"pb_selling_price"`
	LensCost        float64  `json:"lens_cost"`
	AccessoriesCost float64  `json:"accessories_cost"`
	Note            *string  `json:"note"`
}

func (s *Service) AddInventoryItem(username string, input AddInventoryInput) (map[string]interface{}, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	employeeID := int64(emp.IDEmployee)
	locationID := int64(loc.IDLocation)
	if input.LocationID != nil {
		locationID = *input.LocationID
	}

	if input.ModelID == 0 {
		return nil, errors.New("Model ID is required")
	}
	if input.InvoiceID == 0 {
		return nil, errors.New("Invoice ID is required")
	}
	if input.Count < 1 {
		input.Count = 1
	}

	// Verify invoice exists
	var invoiceExists int64
	s.db.Raw(`SELECT COUNT(*) FROM invoice WHERE id_invoice = ?`, input.InvoiceID).Scan(&invoiceExists)
	if invoiceExists == 0 {
		return nil, fmt.Errorf("invoice with id=%d not found", input.InvoiceID)
	}

	itemNet := input.ItemListCost - input.ItemDiscount

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	type createdItem struct {
		IDInventory int64
		SKU         string
	}
	var created []createdItem

	for i := 0; i < input.Count; i++ {
		inv := invModel.Inventory{
			LocationID:           locationID,
			ModelID:              &input.ModelID,
			InvoiceID:            input.InvoiceID,
			EmployeeID:           &employeeID,
			StatusItemsInventory: "Ready for Sale",
		}
		if err := tx.Create(&inv).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("error adding inventory: %w", err)
		}

		generatedSKU := fmt.Sprintf("%06d", inv.IDInventory)
		tx.Model(&inv).Update("sku", generatedSKU)

		created = append(created, createdItem{IDInventory: inv.IDInventory, SKU: generatedSKU})

		// PriceBook
		pbItemNet := itemNet
		pbLP := input.ItemListCost
		pbDisc := input.ItemDiscount
		pbSP := input.PbSellingPrice
		pbLC := input.LensCost
		pbAC := input.AccessoriesCost
		tx.Create(&invModel.PriceBook{
			InventoryID:    inv.IDInventory,
			ItemListCost:   &pbLP,
			ItemDiscount:   &pbDisc,
			ItemNet:        &pbItemNet,
			PbSellingPrice: &pbSP,
			LensCost:       &pbLC,
			AccessoriesCost: &pbAC,
			Note:           input.Note,
		})

		// InventoryTransaction
		tx.Create(&invModel.InventoryTransaction{
			InventoryID:     &inv.IDInventory,
			ToLocationID:    &locationID,
			TransferredBy:   employeeID,
			InvoiceID:       &input.InvoiceID,
			StatusItems:     "Ready for Sale",
			TransactionType: "Received",
			DateTransaction: time.Now().UTC(),
		})

		// ReceiptsItems
		tx.Create(&invModel.ReceiptsItems{
			InvoiceID:   input.InvoiceID,
			InventoryID: inv.IDInventory,
		})
	}

	// Update invoice totals
	updateInvoiceTotals(tx, input.InvoiceID)

	// Update shipment
	updateShipment(tx, input.InvoiceID)

	// Activity log
	for _, ci := range created {
		activitylog.Log(tx, "inventory", "add",
			activitylog.WithEntity(ci.IDInventory),
			activitylog.WithDetails(map[string]interface{}{"sku": ci.SKU}),
		)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("error adding inventory: %w", err)
	}

	// Build response — items from transactions
	var items []map[string]interface{}
	type txItem struct {
		SKU             string  `gorm:"column:sku"`
		ProductTitle    *string `gorm:"column:product_title"`
		VariantTitle    *string `gorm:"column:variant_title"`
		ItemListCost    *string `gorm:"column:item_list_cost"`
		ItemNet         *string `gorm:"column:item_net"`
		PbSellingPrice  *string `gorm:"column:pb_selling_price"`
		LensCost        *string `gorm:"column:lens_cost"`
		AccessoriesCost *string `gorm:"column:accessories_cost"`
		Accessories     *string `gorm:"column:accessories"`
		Status          string  `gorm:"column:status_items_inventory"`
		DateTransaction *string `gorm:"column:date_transaction"`
	}
	var txItems []txItem
	s.db.Raw(`
		SELECT i.sku, p.title_product AS product_title, m.title_variant AS variant_title,
		       pb.item_list_cost::text, pb.item_net::text, pb.pb_selling_price::text,
		       pb.lens_cost::text, pb.accessories_cost::text,
		       m.accessories, i.status_items_inventory,
		       it.date_transaction::text
		FROM inventory_transaction it
		JOIN inventory i ON it.inventory_id = i.id_inventory
		LEFT JOIN model m ON i.model_id = m.id_model
		LEFT JOIN product p ON m.product_id = p.id_product
		LEFT JOIN price_book pb ON pb.inventory_id = i.id_inventory
		WHERE it.invoice_id = ?
	`, input.InvoiceID).Scan(&txItems)

	for _, t := range txItems {
		items = append(items, map[string]interface{}{
			"sku":              t.SKU,
			"product_title":    t.ProductTitle,
			"variant_title":    t.VariantTitle,
			"item_list_cost":   t.ItemListCost,
			"item_net":         t.ItemNet,
			"pb_selling_price": t.PbSellingPrice,
			"lens_cost":        t.LensCost,
			"accessories_cost": t.AccessoriesCost,
			"accessories":      t.Accessories,
			"status":           t.Status,
			"date_transaction": t.DateTransaction,
		})
	}

	return map[string]interface{}{
		"invoice_id": input.InvoiceID,
		"items":      items,
	}, nil
}

func updateInvoiceTotals(tx *gorm.DB, invoiceID int64) {
	tx.Exec(`
		UPDATE invoice SET
			total_amount = COALESCE((
				SELECT SUM(COALESCE(pb.item_net, 0))
				FROM receipts_items ri
				JOIN price_book pb ON pb.inventory_id = ri.inventory_id
				WHERE ri.invoice_id = ?
			), 0),
			quantity = (SELECT COUNT(*) FROM receipts_items WHERE invoice_id = ?),
			final_amount = COALESCE((
				SELECT SUM(COALESCE(pb.item_net, 0))
				FROM receipts_items ri
				JOIN price_book pb ON pb.inventory_id = ri.inventory_id
				WHERE ri.invoice_id = ?
			), 0) - COALESCE(discount, 0)
		WHERE id_invoice = ?
	`, invoiceID, invoiceID, invoiceID, invoiceID)
}

func updateShipment(tx *gorm.DB, invoiceID int64) {
	tx.Exec(`
		UPDATE shipment SET
			qty_ok = (SELECT COUNT(*) FROM receipts_items WHERE invoice_id = ?),
			cost = COALESCE((
				SELECT SUM(COALESCE(pb.item_net, 0))
				FROM receipts_items ri
				JOIN price_book pb ON pb.inventory_id = ri.inventory_id
				WHERE ri.invoice_id = ?
			), 0)
		WHERE vendor_invoice_id = (
			SELECT id_vendor_invoice FROM vendor_invoice WHERE invoice_id = ? LIMIT 1
		)
	`, invoiceID, invoiceID, invoiceID)
}
