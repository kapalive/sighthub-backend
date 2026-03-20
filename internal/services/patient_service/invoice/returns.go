package invoice

import (
	"errors"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"

	invModel "sighthub-backend/internal/models/inventory"
	"sighthub-backend/internal/models/invoices"
	patModel "sighthub-backend/internal/models/patients"
)

// ─── Input DTOs ───────────────────────────────────────────────────────────────

type ReturnItemInput struct {
	ItemSaleID int64 `json:"item_sale_id"`
}

type ProcessReturnInput struct {
	ReturnReason string            `json:"return_reason"`
	Items        []ReturnItemInput `json:"items"`
}

// ─── ProcessReturn ────────────────────────────────────────────────────────────

type ProcessReturnResult struct {
	Message            string  `json:"message"`
	ReturnID           int64   `json:"return_id"`
	TotalReturnAmount  float64 `json:"total_return_amount"`
}

func (s *Service) ProcessReturn(username string, invoiceID int64, input ProcessReturnInput) (*ProcessReturnResult, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, errors.New("invoice not found")
	}
	if inv.LocationID != int64(loc.IDLocation) {
		return nil, errors.New("invoice was created in another location. returns can only be processed in the same location")
	}
	if inv.Due > 0 {
		return nil, fmt.Errorf("return is not allowed. invoice is not fully paid. due: %.2f", inv.Due)
	}

	if len(input.Items) == 0 {
		return nil, errors.New("no items provided for return")
	}

	reason := input.ReturnReason
	if reason == "" {
		reason = "No reason specified"
	}

	var result *ProcessReturnResult
	err = s.db.Transaction(func(tx *gorm.DB) error {
		ri := invoices.ReturnInvoice{
			InvoiceID:    invoiceID,
			ReturnReason: &reason,
			Status:       "Initialized",
		}
		if err := tx.Create(&ri).Error; err != nil {
			return err
		}

		totalReturnAmount := 0.0

		for _, item := range input.Items {
			var invItem invoices.InvoiceItemSale
			if err := tx.Where("id_invoice_sale = ? AND invoice_id = ? AND item_type != 'service'",
				item.ItemSaleID, invoiceID).First(&invItem).Error; err != nil {
				return fmt.Errorf("item with ID %d not found or is not a product", item.ItemSaleID)
			}

			returnAmount := math.Round(invItem.Price*100) / 100
			totalReturnAmount += returnAmount

			returnItem := invoices.ReturnItem{
				ReturnID:     ri.ReturnID,
				ItemSaleID:   item.ItemSaleID,
				ReturnAmount: &returnAmount,
			}
			if err := tx.Create(&returnItem).Error; err != nil {
				return err
			}
		}

		ri.ReturnAmount = totalReturnAmount
		if err := tx.Save(&ri).Error; err != nil {
			return err
		}

		result = &ProcessReturnResult{
			Message:           "Return request created successfully. It needs to be confirmed to finalize the return.",
			ReturnID:          ri.ReturnID,
			TotalReturnAmount: totalReturnAmount,
		}
		return nil
	})
	return result, err
}

// ─── DenyReturn ───────────────────────────────────────────────────────────────

type DenyReturnResult struct {
	Message string        `json:"message"`
	Skipped []interface{} `json:"skipped,omitempty"`
}

func (s *Service) DenyReturn(username string, returnID int64) (*DenyReturnResult, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var ri invoices.ReturnInvoice
	if err := s.db.First(&ri, returnID).Error; err != nil {
		return nil, errors.New("return record not found")
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, ri.InvoiceID).Error; err != nil {
		return nil, errors.New("invoice not found")
	}
	if inv.LocationID != int64(loc.IDLocation) {
		return nil, errors.New("invoice was created in another location. return denial can only be processed in the same location")
	}
	if ri.Status != "Initialized" {
		return nil, errors.New("return request is already processed or in invalid status")
	}

	var skipped []interface{}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		ri.Status = "Request Denied"
		if err := tx.Save(&ri).Error; err != nil {
			return err
		}

		var returnItems []invoices.ReturnItem
		tx.Where("return_id = ?", returnID).Find(&returnItems)

		for _, ritem := range returnItems {
			var invItem invoices.InvoiceItemSale
			if err := tx.First(&invItem, ritem.ItemSaleID).Error; err != nil {
				skipped = append(skipped, map[string]interface{}{
					"item_sale_id": ritem.ItemSaleID,
					"reason":       "invoice_item_sale not found",
				})
				continue
			}

			if invItem.ItemID == nil || *invItem.ItemID == 0 {
				skipped = append(skipped, map[string]interface{}{
					"item_sale_id": invItem.IDInvoiceSale,
					"item_id":      0,
					"reason":       "skip: item_id=0 (placeholder / not inventory)",
				})
				continue
			}

			var inventory invModel.Inventory
			if err := tx.First(&inventory, *invItem.ItemID).Error; err != nil {
				skipped = append(skipped, map[string]interface{}{
					"item_sale_id": invItem.IDInvoiceSale,
					"item_id":      *invItem.ItemID,
					"reason":       "skip: inventory row not found (FK protection)",
				})
				continue
			}

			fromLocID := int64(loc.IDLocation)
			rbID := ri.InvoiceID
			note := fmt.Sprintf("deny_return return_id=%d item_sale_id=%d", ri.ReturnID, invItem.IDInvoiceSale)
			_ = note
			if err := s.addInventoryTxWithNotes(tx, inventory.IDInventory, &fromLocID, nil,
				int64(emp.IDEmployee), rbID, &returnID, "SOLD", "Return Denied"); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	result := &DenyReturnResult{
		Message: "Return request denied; items kept as sold. No credit adjustments made.",
	}
	if len(skipped) > 0 {
		result.Skipped = skipped
	}
	return result, nil
}

// ─── ConfirmReturn ────────────────────────────────────────────────────────────

type ConfirmReturnResult struct {
	Message           string        `json:"message"`
	TotalReturnAmount float64       `json:"total_return_amount"`
	SkippedInventory  []interface{} `json:"skipped_inventory,omitempty"`
}

func (s *Service) ConfirmReturn(username string, returnID int64) (*ConfirmReturnResult, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var ri invoices.ReturnInvoice
	if err := s.db.First(&ri, returnID).Error; err != nil {
		return nil, errors.New("return record not found")
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, ri.InvoiceID).Error; err != nil {
		return nil, errors.New("invoice not found")
	}
	if ri.Status != "Initialized" {
		return nil, errors.New("return already confirmed or in invalid status")
	}

	var skippedInventory []interface{}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		ri.Status = "Confirmed"
		if err := tx.Save(&ri).Error; err != nil {
			return err
		}

		var returnItems []invoices.ReturnItem
		tx.Where("return_id = ?", returnID).Find(&returnItems)

		for _, ritem := range returnItems {
			var invItem invoices.InvoiceItemSale
			if err := tx.First(&invItem, ritem.ItemSaleID).Error; err != nil {
				return fmt.Errorf("invoiceItemSale with ID %d not found", ritem.ItemSaleID)
			}

			qty := invItem.Quantity
			if qty <= 0 {
				return fmt.Errorf("invalid quantity for invoice_item_sale=%d: %d", invItem.IDInvoiceSale, qty)
			}

			perUnitTax := invItem.TotalTax / float64(qty)
			oldTotal := invItem.Total

			newQty := qty - 1
			invItem.Quantity = newQty
			newPrice := invItem.Price * float64(newQty)
			invItem.Total = math.Round(newPrice*100) / 100
			invItem.TotalTax = math.Round(perUnitTax*float64(newQty)*10000) / 10000

			newTotal := invItem.Total
			oldIns := 0.0
			if invItem.InsBalance != nil {
				oldIns = *invItem.InsBalance
			}
			if oldTotal > 0 && oldIns > 0 {
				ratio := oldIns / oldTotal
				newIns := math.Round(newTotal*ratio*100) / 100
				newPt := math.Round((newTotal-newIns)*100) / 100
				invItem.InsBalance = &newIns
				invItem.PtBalance = &newPt
			} else {
				invItem.InsBalance = float64Ptr(0)
				invItem.PtBalance = &newTotal
			}

			if err := tx.Save(&invItem).Error; err != nil {
				return err
			}

			// Handle inventory
			if invItem.ItemID == nil || *invItem.ItemID == 0 {
				skippedInventory = append(skippedInventory, map[string]interface{}{
					"item_sale_id": invItem.IDInvoiceSale,
					"item_id":      0,
					"reason":       "skip inventory: item_id=0 (placeholder / not inventory)",
				})
				continue
			}

			var inventory invModel.Inventory
			if err := tx.First(&inventory, *invItem.ItemID).Error; err != nil {
				skippedInventory = append(skippedInventory, map[string]interface{}{
					"item_sale_id": invItem.IDInvoiceSale,
					"item_id":      *invItem.ItemID,
					"reason":       "skip inventory: inventory row not found (FK protection)",
				})
				continue
			}

			inventory.StatusItemsInventory = "Ready for Sale"
			if err := tx.Save(&inventory).Error; err != nil {
				return err
			}

			if err := s.addInventoryTx(tx, inventory.IDInventory, nil, nil,
				int64(emp.IDEmployee), inv.IDInvoice, &returnID, "Ready for Sale", "Return"); err != nil {
				return err
			}
		}

		if err := s.recalcInvoice(tx, &inv); err != nil {
			return err
		}

		totalReturnAmount := ri.ReturnAmount

		// Credit client balance
		var cb patModel.ClientBalance
		if err := tx.Where("patient_id = ? AND location_id = ?", *inv.PatientID, inv.LocationID).
			First(&cb).Error; err != nil {
			cb = patModel.ClientBalance{
				PatientID:  *inv.PatientID,
				LocationID: int(inv.LocationID),
				Credit:     0,
			}
			if err := tx.Create(&cb).Error; err != nil {
				return err
			}
		}
		cb.Credit += totalReturnAmount
		if err := tx.Save(&cb).Error; err != nil {
			return err
		}

		note := fmt.Sprintf("Confirmed Return -> added to client balance. Return ID %d", ri.ReturnID)
		tc := patModel.TransferCredit{
			InvoiceID: inv.IDInvoice,
			PatientID: inv.PatientID,
			Amount:    totalReturnAmount,
			Note:      &note,
		}
		return tx.Create(&tc).Error
	})
	if err != nil {
		return nil, err
	}

	result := &ConfirmReturnResult{
		Message:           "Return confirmed, invoice updated, client balance credited",
		TotalReturnAmount: ri.ReturnAmount,
	}
	if len(skippedInventory) > 0 {
		result.SkippedInventory = skippedInventory
	}
	return result, nil
}

// ─── GetReturnsByInvoice ──────────────────────────────────────────────────────

type ReturnSummary struct {
	ReturnID   int64   `json:"return_id"`
	Status     string  `json:"status"`
	ReturnDate *string `json:"return_date"`
}

func (s *Service) GetReturnsByInvoice(username string, invoiceID int64) ([]ReturnSummary, error) {
	_, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, errors.New("invoice not found")
	}

	var returns []invoices.ReturnInvoice
	if err := s.db.Where("invoice_id = ?", invoiceID).Find(&returns).Error; err != nil {
		return nil, err
	}

	result := make([]ReturnSummary, 0, len(returns))
	for _, r := range returns {
		summary := ReturnSummary{
			ReturnID: r.ReturnID,
			Status:   r.Status,
		}
		d := r.ReturnDate.Format("2006-01-02 15:04:05")
		summary.ReturnDate = &d
		result = append(result, summary)
	}
	return result, nil
}

// ─── GetReturn ────────────────────────────────────────────────────────────────

type ReturnItemDetail struct {
	ItemID      interface{} `json:"item_id"`
	Description string      `json:"description"`
	Quantity    int         `json:"quantity"`
	Price       string      `json:"price"`
	Total       string      `json:"total"`
}

type ReturnDetail struct {
	ReturnID          int64              `json:"return_id"`
	InvoiceID         int64              `json:"invoice_id"`
	ReturnReason      *string            `json:"return_reason"`
	ReturnDate        *string            `json:"return_date"`
	Status            string             `json:"status"`
	ReturnedQty       int                `json:"returned_quantity"`
	TotalReturnAmount float64            `json:"total_return_amount"`
	Items             []ReturnItemDetail `json:"items"`
}

func (s *Service) GetReturn(username string, returnID int64) (*ReturnDetail, error) {
	_, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var ri invoices.ReturnInvoice
	if err := s.db.First(&ri, returnID).Error; err != nil {
		return nil, errors.New("return record not found")
	}

	var returnItems []invoices.ReturnItem
	s.db.Preload("InvoiceItemSale").Where("return_id = ?", returnID).Find(&returnItems)

	var items []ReturnItemDetail
	if len(returnItems) > 0 {
		for _, ritem := range returnItems {
			if ritem.InvoiceItemSale != nil {
				amount := 0.0
				if ritem.ReturnAmount != nil {
					amount = *ritem.ReturnAmount
				}
				items = append(items, ReturnItemDetail{
					ItemID:      ritem.InvoiceItemSale.ItemID,
					Description: ritem.InvoiceItemSale.Description,
					Quantity:    1,
					Price:       fmt.Sprintf("%.2f", amount),
					Total:       fmt.Sprintf("%.2f", amount),
				})
			}
		}
	} else {
		// Fallback to inventory transactions
		type txRow struct {
			InventoryID int64
		}
		var txRows []txRow
		s.db.Raw(`SELECT inventory_id FROM inventory_transaction
			WHERE invoice_id = ? AND transaction_type = 'Return'`, ri.InvoiceID).Scan(&txRows)
		for _, row := range txRows {
			items = append(items, ReturnItemDetail{
				ItemID:      row.InventoryID,
				Description: "Inventory Transaction",
				Quantity:    1,
				Price:       "0.00",
				Total:       "0.00",
			})
		}
	}

	var returnDateStr *string
	d := ri.ReturnDate.Format("2006-01-02 15:04:05")
	returnDateStr = &d

	return &ReturnDetail{
		ReturnID:          ri.ReturnID,
		InvoiceID:         ri.InvoiceID,
		ReturnReason:      ri.ReturnReason,
		ReturnDate:        returnDateStr,
		Status:            ri.Status,
		ReturnedQty:       ri.ReturnedQty,
		TotalReturnAmount: ri.ReturnAmount,
		Items:             items,
	}, nil
}

// ─── DeleteReturn ─────────────────────────────────────────────────────────────

func (s *Service) DeleteReturn(username string, returnID int64) error {
	_, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}

	var ri invoices.ReturnInvoice
	if err := s.db.First(&ri, returnID).Error; err != nil {
		return errors.New("return record not found")
	}
	if ri.Status != "Initialized" {
		return errors.New("cannot delete a return that is already confirmed or denied")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete return items
		if err := tx.Where("return_id = ?", returnID).Delete(&invoices.ReturnItem{}).Error; err != nil {
			return err
		}

		// Delete inventory transactions with old_invoice_id = return_id and type=Return
		if err := tx.Exec(`DELETE FROM inventory_transaction WHERE old_invoice_id = ? AND transaction_type = 'Return'`, returnID).Error; err != nil {
			return err
		}

		return tx.Delete(&ri).Error
	})
}

// ─── addInventoryTxWithNotes ──────────────────────────────────────────────────
// Wraps addInventoryTx, notes param ignored for now (not in schema)

func (s *Service) addInventoryTxWithNotes(tx *gorm.DB, inventoryID int64, fromLocID *int64, toLocID *int64,
	transferredBy, invoiceID int64, oldInvoiceID *int64, statusItems, txType string) error {
	return s.addInventoryTx(tx, inventoryID, fromLocID, toLocID, transferredBy, invoiceID, oldInvoiceID, statusItems, txType)
}

// ─── unused time import guard ─────────────────────────────────────────────────

var _ = time.Now
