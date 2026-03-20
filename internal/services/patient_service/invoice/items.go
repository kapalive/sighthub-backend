package invoice

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"gorm.io/gorm"

	clModel "sighthub-backend/internal/models/contact_lens"
	generalModel "sighthub-backend/internal/models/general"
	insModel "sighthub-backend/internal/models/insurance"
	invModel "sighthub-backend/internal/models/inventory"
	"sighthub-backend/internal/models/invoices"
	lensModel "sighthub-backend/internal/models/lenses"
	marketingModel "sighthub-backend/internal/models/marketing"
	patModel "sighthub-backend/internal/models/patients"
	svcModel "sighthub-backend/internal/models/service"
)

// ─── Input DTOs ───────────────────────────────────────────────────────────────

type UpdateItemInput struct {
	Price       *float64 `json:"price"`
	Discount    *float64 `json:"discount"`
	Quantity    *int     `json:"quantity"`
	Description *string  `json:"description"`
	Taxable     *bool    `json:"taxable"`
}

type SetLineBalanceInput struct {
	PtBalance  *float64 `json:"pt_balance"`
	InsBalance *float64 `json:"ins_balance"`
}

type AddInsurancePolicyInput struct {
	InsuranceID int64 `json:"insurance_id"`
}

type AddGiftCardInput struct {
	GiftCardID      int    `json:"gift_card_id"`
	GiftCardBalance string `json:"gift_card_balance"`
}

// ─── UpdateItem ───────────────────────────────────────────────────────────────

type UpdateItemResult struct {
	Message     string  `json:"message"`
	TotalAmount float64 `json:"total_amount"`
	TaxAmount   float64 `json:"tax_amount"`
	FinalAmount float64 `json:"final_amount"`
	PtBal       float64 `json:"pt_bal"`
	InsBal      float64 `json:"ins_bal"`
	Due         float64 `json:"due"`
}

func (s *Service) UpdateItem(username string, invoiceID, itemSaleID int64, input UpdateItemInput) (*UpdateItemResult, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, errors.New("invoice not found")
	}
	if inv.Finalized {
		return nil, errors.New("invoice is finalized")
	}

	var item invoices.InvoiceItemSale
	if err := s.db.Where("id_invoice_sale = ? AND invoice_id = ?", itemSaleID, invoiceID).First(&item).Error; err != nil {
		return nil, errors.New("invoice item not found")
	}

	var result *UpdateItemResult
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Apply updates
		price := item.Price
		if input.Price != nil {
			price = *input.Price
		}
		discount := item.Discount
		if input.Discount != nil {
			discount = *input.Discount
		}
		qty := item.Quantity
		if input.Quantity != nil {
			qty = *input.Quantity
		}
		if input.Description != nil {
			item.Description = *input.Description
		}
		if input.Taxable != nil {
			item.Taxable = input.Taxable
		}

		pbKey := item.ItemType
		if pbKey == "Misc" || pbKey == "misc" {
			qty = 1
			item.Taxable = boolPtr(false)
		}

		item.Price = price
		item.Discount = discount
		item.Quantity = qty

		qf := float64(qty)
		lineSubtotal := math.Round((qf*price-discount)*100) / 100

		// Recalc cost
		lineCost := 0.0
		if item.ItemID != nil {
			switch pbKey {
			case "Frames":
				var pb invModel.PriceBook
				if s.db.Where("inventory_id = ?", *item.ItemID).First(&pb).Error == nil && pb.ItemListCost != nil {
					lineCost = math.Round(*pb.ItemListCost*qf*100) / 100
				}
			case "Lens":
				var lens lensModel.Lenses
				if s.db.First(&lens, *item.ItemID).Error == nil && lens.Cost != nil {
					lineCost = math.Round(*lens.Cost/2*qf*100) / 100
				}
			case "Contact Lens":
				var cl clModel.ContactLensItem
				if s.db.First(&cl, *item.ItemID).Error == nil && cl.Cost != nil {
					lineCost = math.Round(*cl.Cost*qf*100) / 100
				}
			case "Treatment":
				var tr lensModel.LensTreatments
				if s.db.First(&tr, *item.ItemID).Error == nil && tr.Cost != nil {
					lineCost = math.Round(*tr.Cost/2*qf*100) / 100
				}
			case "Prof. service":
				var srv svcModel.ProfessionalService
				if s.db.First(&srv, *item.ItemID).Error == nil {
					lineCost = math.Round(srv.Cost*qf*100) / 100
				}
			case "Add service":
				var add svcModel.AdditionalService
				if s.db.First(&add, *item.ItemID).Error == nil {
					lineCost = math.Round(add.CostPrice*qf*100) / 100
				}
			case "Misc", "misc":
				// misc cost stored as string pointer
			}
		}
		item.Cost = lineCost

		// Recalc tax
		taxRate := 0.0
		if item.Taxable != nil && *item.Taxable {
			if pbKey == "Prof. service" || pbKey == "Add service" {
				if loc.State != nil {
					var svcTax generalModel.ServiceTaxByState
					if s.db.Where("state_code = ? AND tax_active = true", *loc.State).
						Order("effective_date desc").First(&svcTax).Error == nil {
						taxRate = svcTax.TaxPercent / 100
					}
				}
			} else if loc.SalesTaxID != nil {
				var st generalModel.SalesTaxByState
				if s.db.First(&st, *loc.SalesTaxID).Error == nil {
					taxRate = st.SalesTaxPercent / 100
				}
			}
		}

		lineTax := math.Round(lineSubtotal*taxRate*10000) / 10000
		oldTotal := item.Total
		newTotal := math.Round((lineSubtotal+lineTax)*100) / 100
		item.TotalTax = lineTax
		item.Total = newTotal

		// Preserve ins/pt ratio
		oldIns := 0.0
		if item.InsBalance != nil {
			oldIns = *item.InsBalance
		}
		if oldTotal > 0 && oldIns > 0 {
			ratio := oldIns / oldTotal
			newIns := math.Round(newTotal*ratio*100) / 100
			newPt := math.Round((newTotal-newIns)*100) / 100
			item.InsBalance = &newIns
			item.PtBalance = &newPt
		} else {
			item.InsBalance = float64Ptr(0)
			item.PtBalance = &newTotal
		}

		if err := tx.Save(&item).Error; err != nil {
			return err
		}

		// Remake mirror
		if inv.Remake {
			negIns := -(*item.InsBalance)
			negPt := -(*item.PtBalance)
			neg := invoices.InvoiceItemSale{
				InvoiceID:   invoiceID,
				ItemType:    item.ItemType,
				ItemID:      item.ItemID,
				Description: item.Description,
				Quantity:    -qty,
				Price:       price,
				Discount:    -discount,
				Total:       -newTotal,
				Taxable:     item.Taxable,
				TotalTax:    -lineTax,
				Cost:        -lineCost,
				InsBalance:  &negIns,
				PtBalance:   &negPt,
			}
			if err := tx.Create(&neg).Error; err != nil {
				return err
			}
		}

		if err := s.recalcInvoice(tx, &inv); err != nil {
			return err
		}

		result = &UpdateItemResult{
			Message:     "Invoice item updated",
			TotalAmount: inv.TotalAmount,
			TaxAmount:   inv.TaxAmount,
			FinalAmount: inv.FinalAmount,
			PtBal:       inv.PTBal,
			InsBal:      inv.InsBal,
			Due:         inv.Due,
		}
		return nil
	})
	return result, err
}

// ─── DeleteItem ───────────────────────────────────────────────────────────────

func (s *Service) DeleteItem(username string, invoiceID, itemSaleID int64) error {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return errors.New("invoice not found")
	}
	if inv.Finalized {
		return errors.New("invoice is finalized")
	}

	var item invoices.InvoiceItemSale
	if err := s.db.Where("id_invoice_sale = ? AND invoice_id = ?", itemSaleID, invoiceID).First(&item).Error; err != nil {
		return fmt.Errorf("item %d not found", itemSaleID)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Rollback frame inventory
		if item.ItemType == "Frames" && item.ItemID != nil {
			var inventory invModel.Inventory
			if err := tx.First(&inventory, *item.ItemID).Error; err == nil {
				rbID := s.pickRollbackInvoiceID(tx, inventory.IDInventory, invoiceID)
				inventory.InvoiceID = rbID
				inventory.StatusItemsInventory = "Ready for Sale"
				if err := tx.Save(&inventory).Error; err != nil {
					return err
				}
				empID := int64(0) // no employee available
				toLocID := int64(loc.IDLocation)
				if err := s.addInventoryTx(tx, inventory.IDInventory, nil, &toLocID, empID, rbID, &invoiceID, "Ready for Sale", "Removed from Invoice"); err != nil {
					return err
				}
			}
		}

		if err := tx.Delete(&item).Error; err != nil {
			return err
		}
		return s.recalcInvoice(tx, &inv)
	})
}

// pickRollbackInvoiceID finds the best invoice_id to roll back inventory to.
func (s *Service) pickRollbackInvoiceID(tx *gorm.DB, inventoryID, currentInvoiceID int64) int64 {
	type txRow struct {
		InvoiceID    *int64
		OldInvoiceID *int64
	}

	// Try last tx for current invoice → old_invoice_id
	var last txRow
	if tx.Raw(`SELECT invoice_id, old_invoice_id FROM inventory_transaction
		WHERE inventory_id = ? AND invoice_id = ?
		ORDER BY date_transaction DESC LIMIT 1`, inventoryID, currentInvoiceID).Scan(&last).Error == nil {
		if last.OldInvoiceID != nil {
			var cnt int64
			if tx.Table("invoice").Where("id_invoice = ?", *last.OldInvoiceID).Count(&cnt).Error == nil && cnt > 0 {
				return *last.OldInvoiceID
			}
		}
	}

	// Base tx: old_invoice_id IS NULL
	var base txRow
	if tx.Raw(`SELECT invoice_id, old_invoice_id FROM inventory_transaction
		WHERE inventory_id = ? AND invoice_id IS NOT NULL AND old_invoice_id IS NULL
		ORDER BY date_transaction DESC LIMIT 1`, inventoryID).Scan(&base).Error == nil {
		if base.InvoiceID != nil {
			var cnt int64
			if tx.Table("invoice").Where("id_invoice = ?", *base.InvoiceID).Count(&cnt).Error == nil && cnt > 0 {
				return *base.InvoiceID
			}
		}
	}

	// Any last invoice != current
	var alt txRow
	if tx.Raw(`SELECT invoice_id, old_invoice_id FROM inventory_transaction
		WHERE inventory_id = ? AND invoice_id IS NOT NULL AND invoice_id != ?
		ORDER BY date_transaction DESC LIMIT 1`, inventoryID, currentInvoiceID).Scan(&alt).Error == nil {
		if alt.InvoiceID != nil {
			var cnt int64
			if tx.Table("invoice").Where("id_invoice = ?", *alt.InvoiceID).Count(&cnt).Error == nil && cnt > 0 {
				return *alt.InvoiceID
			}
		}
	}

	return currentInvoiceID
}

// ─── SetLineBalance ───────────────────────────────────────────────────────────

type LineBalanceResult struct {
	Message      string  `json:"message"`
	ItemSaleID   int64   `json:"item_sale_id"`
	LineTotal    float64 `json:"line_total"`
	PtBalance    float64 `json:"pt_balance"`
	InsBalance   float64 `json:"ins_balance"`
	InvoicePtBal float64 `json:"invoice_pt_bal"`
	InvoiceInsBal float64 `json:"invoice_ins_bal"`
	Due          float64 `json:"due"`
}

func (s *Service) SetLineBalance(username string, invoiceID, itemSaleID int64, input SetLineBalanceInput) (*LineBalanceResult, error) {
	_, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, errors.New("invoice not found")
	}
	if inv.Finalized {
		return nil, errors.New("invoice is finalized")
	}

	var item invoices.InvoiceItemSale
	if err := s.db.Where("id_invoice_sale = ? AND invoice_id = ?", itemSaleID, invoiceID).First(&item).Error; err != nil {
		return nil, errors.New("invoice item not found")
	}

	if input.PtBalance == nil && input.InsBalance == nil {
		return nil, errors.New("provide 'pt_balance' and/or 'ins_balance'")
	}

	lineTotal := item.Total

	var newPt, newIns float64
	switch {
	case input.PtBalance != nil && input.InsBalance != nil:
		newPt = math.Round(*input.PtBalance*100) / 100
		newIns = math.Round(*input.InsBalance*100) / 100
		if newPt < 0 || newIns < 0 {
			return nil, errors.New("balances cannot be negative")
		}
		sum := math.Round((newPt+newIns)*100) / 100
		if sum != math.Round(lineTotal*100)/100 {
			return nil, fmt.Errorf("pt_balance (%.2f) + ins_balance (%.2f) = %.2f, but line total is %.2f",
				newPt, newIns, sum, lineTotal)
		}
	case input.InsBalance != nil:
		newIns = math.Round(*input.InsBalance*100) / 100
		if newIns < 0 {
			return nil, errors.New("ins_balance cannot be negative")
		}
		if newIns > lineTotal {
			return nil, fmt.Errorf("ins_balance (%.2f) exceeds line total (%.2f)", newIns, lineTotal)
		}
		newPt = math.Round((lineTotal-newIns)*100) / 100
	case input.PtBalance != nil:
		newPt = math.Round(*input.PtBalance*100) / 100
		if newPt < 0 {
			return nil, errors.New("pt_balance cannot be negative")
		}
		if newPt > lineTotal {
			return nil, fmt.Errorf("pt_balance (%.2f) exceeds line total (%.2f)", newPt, lineTotal)
		}
		newIns = math.Round((lineTotal-newPt)*100) / 100
	}

	item.PtBalance = &newPt
	item.InsBalance = &newIns

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&item).Error; err != nil {
			return err
		}
		return s.recalcInvoice(tx, &inv)
	})
	if err != nil {
		return nil, err
	}

	return &LineBalanceResult{
		Message:       "Line balance updated",
		ItemSaleID:    item.IDInvoiceSale,
		LineTotal:     lineTotal,
		PtBalance:     newPt,
		InsBalance:    newIns,
		InvoicePtBal:  inv.PTBal,
		InvoiceInsBal: inv.InsBal,
		Due:           inv.Due,
	}, nil
}

// ─── AddInsurancePolicy ───────────────────────────────────────────────────────

func (s *Service) AddInsurancePolicy(username string, invoiceID int64, input AddInsurancePolicyInput) error {
	_, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return errors.New("invoice not found")
	}
	if inv.Finalized {
		return errors.New("invoice is finalized (locked) and cannot be updated")
	}

	var policy insModel.InsurancePolicy
	if err := s.db.First(&policy, input.InsuranceID).Error; err != nil {
		return fmt.Errorf("insurance policy with id %d not found", input.InsuranceID)
	}

	// Verify patient is a holder
	var holder patModel.InsuranceHolderPatients
	if err := s.db.Where("insurance_policy_id = ? AND patient_id = ?", input.InsuranceID, inv.PatientID).
		First(&holder).Error; err != nil {
		return fmt.Errorf("patient %d is not an authorized holder for policy %d", inv.PatientID, input.InsuranceID)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var existing []invoices.InvoiceInsurancePolicy
		tx.Where("invoice_id = ?", invoiceID).Find(&existing)

		if len(existing) == 1 && existing[0].InsurancePolicyID == policy.IDInsurancePolicy {
			return nil // already attached
		}

		for _, link := range existing {
			if err := tx.Delete(&link).Error; err != nil {
				return err
			}
		}

		newLink := invoices.InvoiceInsurancePolicy{
			InvoiceID:         invoiceID,
			InsurancePolicyID: policy.IDInsurancePolicy,
		}
		if err := tx.Create(&newLink).Error; err != nil {
			return err
		}

		inv.InsurancePolicyID = &policy.IDInsurancePolicy
		return tx.Save(&inv).Error
	})
}

// ─── DeleteInsuranceFromInvoice ───────────────────────────────────────────────

func (s *Service) DeleteInsuranceFromInvoice(username string, invoiceID int64) error {
	_, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return errors.New("invoice not found")
	}
	if inv.Finalized {
		return errors.New("invoice is finalized (locked) and cannot be updated")
	}

	var link invoices.InvoiceInsurancePolicy
	if err := s.db.Where("invoice_id = ?", invoiceID).First(&link).Error; err != nil {
		return errors.New("no insurance policy associated with this invoice")
	}

	if inv.InsBal > 0 {
		return errors.New("cannot delete insurance because insurance balance is not zero")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&link).Error; err != nil {
			return err
		}
		inv.InsBal = 0
		return tx.Save(&inv).Error
	})
}

// ─── AddGiftCard ──────────────────────────────────────────────────────────────

func (s *Service) AddGiftCard(username string, invoiceID int64, input AddGiftCardInput) error {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return errors.New("invoice not found")
	}
	if inv.Finalized {
		return errors.New("invoice is finalized (locked) and cannot be updated")
	}

	var gc marketingModel.GiftCard
	if err := s.db.First(&gc, input.GiftCardID).Error; err != nil {
		return errors.New("gift card not found")
	}

	gcAmountStr := input.GiftCardBalance
	if gcAmountStr == "" {
		gcAmountStr = "0.00"
	}
	gcAmount, err := strconv.ParseFloat(gcAmountStr, 64)
	if err != nil || gcAmount <= 0 {
		return errors.New("invalid gift card balance")
	}

	gcBalance, err := strconv.ParseFloat(gc.Balance, 64)
	if err != nil {
		return errors.New("invalid gift card balance on record")
	}
	if gcBalance < gcAmount {
		return errors.New("not enough balance on gift card")
	}

	remainingPt := inv.PTBal
	if gcAmount > remainingPt {
		return fmt.Errorf("gift card overpayment is not allowed. Maximum applicable amount: %.2f", remainingPt)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		newBalance := gcBalance - gcAmount
		gc.Balance = fmt.Sprintf("%.2f", newBalance)
		if newBalance == 0 {
			gc.Status = "used"
		}
		if err := tx.Save(&gc).Error; err != nil {
			return err
		}

		if inv.GiftCardBal != nil {
			newGCBal := *inv.GiftCardBal + gcAmount
			inv.GiftCardBal = &newGCBal
		} else {
			newGCBal := gcAmount
			inv.GiftCardBal = &newGCBal
		}

		pmID := int64(14)
		empID := int64(emp.IDEmployee)
		ph := patModel.PaymentHistory{
			PatientID:        inv.PatientID,
			InvoiceID:        invoiceID,
			Amount:           gcAmount,
			PaymentTimestamp: time.Now(),
			PaymentMethodID:  &pmID,
			EmployeeID:       &empID,
		}
		if err := tx.Create(&ph).Error; err != nil {
			return err
		}

		gcID := input.GiftCardID
		invIDInt := int(invoiceID)
		patIDInt := func() int { if inv.PatientID != nil { return int(*inv.PatientID) }; return 0 }()
		gct := marketingModel.GiftCardTransaction{
			GiftCardID:           &gcID,
			TransactionType:      "usage",
			Amount:               fmt.Sprintf("%.2f", gcAmount),
			ProcessedByPatientID: &patIDInt,
			RelatedInvoiceID:     &invIDInt,
		}
		if err := tx.Create(&gct).Error; err != nil {
			return err
		}

		return s.recalcInvoice(tx, &inv)
	})
}

// ─── DeleteGiftCard ───────────────────────────────────────────────────────────

func (s *Service) DeleteGiftCard(username string, invoiceID int64) error {
	_, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return errors.New("invoice not found")
	}

	var gc marketingModel.GiftCard
	if err := s.db.Where("invoice_id = ?", invoiceID).First(&gc).Error; err != nil {
		return errors.New("no gift card associated with this invoice")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete transactions
		invIDInt := int(invoiceID)
		if err := tx.Where("related_invoice_id = ?", invIDInt).
			Delete(&marketingModel.GiftCardTransaction{}).Error; err != nil {
			return err
		}

		if err := tx.Delete(&gc).Error; err != nil {
			return err
		}

		zero := 0.0
		inv.GiftCardBal = &zero
		return tx.Save(&inv).Error
	})
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func boolPtr(b bool) *bool     { return &b }
func float64Ptr(f float64) *float64 { return &f }
