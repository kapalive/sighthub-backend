package ordered_inventory_service

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	"sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/frames"
	invModel "sighthub-backend/internal/models/inventory"
	"sighthub-backend/internal/models/invoices"
	"sighthub-backend/internal/models/location"
	"sighthub-backend/internal/models/vendors"
	invoiceSvc "sighthub-backend/internal/services/invoice_service"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrBadRequest = errors.New("bad request")
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service { return &Service{db: db} }

// GetEmpLocation resolves the current user's employee + location from JWT username.
func (s *Service) GetEmpLocation(username string) (*invoiceSvc.EmpLocation, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, fmt.Errorf("login not found")
	}

	var emp employees.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, fmt.Errorf("employee not found")
	}

	if emp.LocationID == nil {
		return nil, fmt.Errorf("employee not assigned to a location")
	}

	var loc location.Location
	if err := s.db.First(&loc, *emp.LocationID).Error; err != nil {
		return nil, fmt.Errorf("location not found")
	}

	return &invoiceSvc.EmpLocation{Employee: &emp, Location: &loc}, nil
}

// ─── POST /add ───────────────────────────────────────────────────────────────

type AddOrderedRequest struct {
	ModelID        int64   `json:"model_id"`
	InvoiceID      int64   `json:"invoice_id"`
	PbSellingPrice float64 `json:"pb_selling_price"`
}

type AddOrderedResult struct {
	IDInventory int64                  `json:"id_inventory"`
	SKU         string                 `json:"sku"`
	Status      string                 `json:"status"`
	Model       map[string]interface{} `json:"model"`
	PriceBook   map[string]interface{} `json:"price_book"`
}

func (s *Service) AddOrderedItem(el *invoiceSvc.EmpLocation, req AddOrderedRequest) (*AddOrderedResult, error) {
	if req.ModelID == 0 || req.InvoiceID == 0 {
		return nil, fmt.Errorf("%w: model_id and invoice_id are required", ErrBadRequest)
	}

	var model frames.Model
	if err := s.db.First(&model, req.ModelID).Error; err != nil {
		return nil, fmt.Errorf("%w: model not found", ErrNotFound)
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, req.InvoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: invoice not found", ErrNotFound)
	}

	locID := int64(el.Location.IDLocation)
	empID := int64(el.Employee.IDEmployee)

	// Create placeholder inventory item
	item := invModel.Inventory{
		LocationID:           locID,
		ModelID:              &req.ModelID,
		InvoiceID:            req.InvoiceID,
		EmployeeID:           &empID,
		StatusItemsInventory: "Ordered",
	}

	if err := s.db.Create(&item).Error; err != nil {
		return nil, fmt.Errorf("failed to create inventory: %v", err)
	}

	// Generate SKU from id_inventory
	item.SKU = fmt.Sprintf("%06d", item.IDInventory)
	s.db.Save(&item)

	// PriceBook — only selling price known at this point
	zero := float64(0)
	pb := invModel.PriceBook{
		InventoryID:      item.IDInventory,
		ItemListCost:     &zero,
		ItemDiscount:     &zero,
		ItemNet:          &zero,
		PbSellingPrice:   &req.PbSellingPrice,
		PbListCost:       &zero,
		PbDiscount:       &zero,
		PbCost:           &zero,
		PbStoreTierPrice: &zero,
		LensCost:         &zero,
		AccessoriesCost:  &zero,
	}
	s.db.Create(&pb)

	// Transaction log — "PreSold"
	notes := fmt.Sprintf("Dropship order — sale invoice %s", inv.NumberInvoice)
	txn := invModel.InventoryTransaction{
		InventoryID:     &item.IDInventory,
		ToLocationID:    &locID,
		TransferredBy:   empID,
		InvoiceID:       &req.InvoiceID,
		StatusItems:     "Ordered",
		TransactionType: "PreSold",
		DateTransaction: time.Now(),
		Notes:           &notes,
	}
	s.db.Create(&txn)

	// Build response with model metadata
	modelInfo := map[string]interface{}{
		"id_model":      model.IDModel,
		"title_variant": model.TitleVariant,
	}

	var product frames.Product
	if err := s.db.First(&product, model.ProductID).Error; err == nil {
		modelInfo["product_title"] = product.TitleProduct
		if product.BrandID != nil {
			var brand vendors.Brand
			if s.db.First(&brand, *product.BrandID).Error == nil {
				modelInfo["brand_name"] = brand.BrandName
			}
		}
		if product.VendorID != nil {
			var vendor vendors.Vendor
			if s.db.First(&vendor, *product.VendorID).Error == nil {
				modelInfo["vendor_name"] = vendor.VendorName
				modelInfo["vendor_id"] = vendor.IDVendor
			}
		}
	}

	return &AddOrderedResult{
		IDInventory: item.IDInventory,
		SKU:         item.SKU,
		Status:      string(item.StatusItemsInventory),
		Model:       modelInfo,
		PriceBook: map[string]interface{}{
			"pb_selling_price": fmtFloat(req.PbSellingPrice),
		},
	}, nil
}

// ─── GET /pending ────────────────────────────────────────────────────────────

type PendingItem struct {
	IDInventory  int64                  `json:"id_inventory"`
	SKU          string                 `json:"sku"`
	Status       string                 `json:"status"`
	BrandName    *string                `json:"brand_name"`
	ProductTitle string                 `json:"product_title"`
	VariantTitle string                 `json:"variant_title"`
	VendorName   string                 `json:"vendor_name"`
	PriceBook    map[string]interface{} `json:"price_book"`
	Sale         interface{}            `json:"sale"`
}

func (s *Service) GetPendingItems(el *invoiceSvc.EmpLocation, vendorID int64) ([]PendingItem, error) {
	if vendorID == 0 {
		return nil, fmt.Errorf("%w: vendor_id is required", ErrBadRequest)
	}

	var vendor vendors.Vendor
	if err := s.db.First(&vendor, vendorID).Error; err != nil {
		return nil, fmt.Errorf("%w: vendor not found", ErrNotFound)
	}

	locID := int64(el.Location.IDLocation)

	// Find all "Ordered" items at this location whose product belongs to the vendor
	type row struct {
		invModel.Inventory
		TitleVariant string  `gorm:"column:title_variant"`
		TitleProduct string  `gorm:"column:title_product"`
		BrandName    *string `gorm:"column:brand_name"`
		ProductID    int64   `gorm:"column:product_id"`
	}

	var rows []row
	s.db.Table("inventory").
		Select("inventory.*, models.title_variant, products.title_product, brands.brand_name, products.product_id").
		Joins("JOIN models ON models.id_model = inventory.model_id").
		Joins("JOIN products ON products.id_product = models.product_id").
		Joins("LEFT JOIN brands ON brands.id_brand = products.brand_id").
		Where("inventory.status_items_inventory = ? AND inventory.location_id = ? AND products.vendor_id = ?",
			"Ordered", locID, vendorID).
		Find(&rows)

	result := make([]PendingItem, 0, len(rows))
	for _, r := range rows {
		// PriceBook
		var pb invModel.PriceBook
		pbMap := map[string]interface{}{
			"item_list_cost":   "0.00",
			"item_discount":    "0.00",
			"item_net":         "0.00",
			"pb_selling_price": "0.00",
		}
		if s.db.Where("inventory_id = ?", r.IDInventory).First(&pb).Error == nil {
			pbMap["item_list_cost"] = fmtFloatPtr(pb.ItemListCost)
			pbMap["item_discount"] = fmtFloatPtr(pb.ItemDiscount)
			pbMap["item_net"] = fmtFloatPtr(pb.ItemNet)
			pbMap["pb_selling_price"] = fmtFloatPtr(pb.PbSellingPrice)
		}

		// Sale info from InvoiceItemSale
		var saleInfo interface{}
		var saleItem invoices.InvoiceItemSale
		if s.db.Where("item_id = ? AND item_type = ?", r.IDInventory, "Frames").First(&saleItem).Error == nil {
			si := map[string]interface{}{
				"id_invoice_sale": saleItem.IDInvoiceSale,
				"invoice_id":     saleItem.InvoiceID,
				"selling_price":  fmtFloat(saleItem.Price),
				"description":    saleItem.Description,
			}
			var saleInv invoices.Invoice
			if s.db.First(&saleInv, saleItem.InvoiceID).Error == nil {
				si["invoice_number"] = saleInv.NumberInvoice
			}
			saleInfo = si
		}

		result = append(result, PendingItem{
			IDInventory:  r.IDInventory,
			SKU:          r.SKU,
			Status:       string(r.StatusItemsInventory),
			BrandName:    r.BrandName,
			ProductTitle: r.TitleProduct,
			VariantTitle: r.TitleVariant,
			VendorName:   vendor.VendorName,
			PriceBook:    pbMap,
			Sale:         saleInfo,
		})
	}

	return result, nil
}

// ─── POST /receive ───────────────────────────────────────────────────────────

type ReceiveOrderedRequest struct {
	InventoryID     int64   `json:"inventory_id"`
	VendorInvoiceID int64   `json:"vendor_invoice_id"`
	ItemListCost    float64 `json:"item_list_cost"`
	ItemDiscount    float64 `json:"item_discount"`
	PbSellingPrice  float64 `json:"pb_selling_price"`
}

type ReceiveOrderedResult struct {
	Message              string                 `json:"message"`
	IDInventory          int64                  `json:"id_inventory"`
	SKU                  string                 `json:"sku"`
	PriceBook            map[string]interface{} `json:"price_book"`
	Warning              string                 `json:"warning,omitempty"`
	OriginalSellingPrice string                 `json:"original_selling_price,omitempty"`
	SaleInvoiceNumber    string                 `json:"sale_invoice_number,omitempty"`
}

func (s *Service) ReceiveOrderedItem(el *invoiceSvc.EmpLocation, req ReceiveOrderedRequest) (*ReceiveOrderedResult, error) {
	if req.InventoryID == 0 || req.VendorInvoiceID == 0 {
		return nil, fmt.Errorf("%w: inventory_id and vendor_invoice_id are required", ErrBadRequest)
	}

	locID := int64(el.Location.IDLocation)
	empID := int64(el.Employee.IDEmployee)

	// Validate inventory item
	var item invModel.Inventory
	if err := s.db.First(&item, req.InventoryID).Error; err != nil {
		return nil, fmt.Errorf("%w: inventory item not found", ErrNotFound)
	}
	if string(item.StatusItemsInventory) != "Ordered" {
		return nil, fmt.Errorf("%w: item status is '%s', expected 'Ordered'", ErrBadRequest, item.StatusItemsInventory)
	}
	if item.LocationID != locID {
		return nil, fmt.Errorf("%w: item does not belong to your location", ErrBadRequest)
	}

	// Validate vendor invoice
	var vi invModel.VendorInvoice
	if err := s.db.First(&vi, req.VendorInvoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: vendor invoice not found", ErrNotFound)
	}

	var purchaseInv invoices.Invoice
	if err := s.db.First(&purchaseInv, vi.InvoiceID).Error; err != nil {
		return nil, fmt.Errorf("%w: purchase invoice linked to vendor invoice not found", ErrNotFound)
	}

	// Vendor must match model's product vendor
	if item.ModelID != nil {
		var model frames.Model
		if s.db.First(&model, *item.ModelID).Error == nil {
			var product frames.Product
			if s.db.First(&product, model.ProductID).Error == nil {
				if product.VendorID == nil || *product.VendorID != int64(vi.VendorID) {
					return nil, fmt.Errorf("%w: item vendor does not match vendor invoice vendor", ErrBadRequest)
				}
			}
		}
	}

	// Parse prices
	itemNet := req.ItemListCost - req.ItemDiscount

	// Find original selling price from sale
	var warning, origPrice, saleInvNum string
	var saleItem invoices.InvoiceItemSale
	if s.db.Where("item_id = ? AND item_type = ?", req.InventoryID, "Frames").First(&saleItem).Error == nil {
		originalSellingPrice := saleItem.Price
		var saleInv invoices.Invoice
		if s.db.First(&saleInv, saleItem.InvoiceID).Error == nil {
			saleInvNum = saleInv.NumberInvoice
		}
		if req.PbSellingPrice != originalSellingPrice {
			warning = fmt.Sprintf(
				"Item was sold at %.2f (invoice %s). You entered %.2f. "+
					"Please set the correct selling price or update the discount in invoice %s.",
				originalSellingPrice, saleInvNum, req.PbSellingPrice, saleInvNum,
			)
			origPrice = fmtFloat(originalSellingPrice)
		}
	}

	// Update PriceBook
	var pb invModel.PriceBook
	if s.db.Where("inventory_id = ?", req.InventoryID).First(&pb).Error == nil {
		pb.ItemListCost = &req.ItemListCost
		pb.ItemDiscount = &req.ItemDiscount
		pb.ItemNet = &itemNet
		pb.PbSellingPrice = &req.PbSellingPrice
		s.db.Save(&pb)
	} else {
		zero := float64(0)
		pb = invModel.PriceBook{
			InventoryID:      req.InventoryID,
			ItemListCost:     &req.ItemListCost,
			ItemDiscount:     &req.ItemDiscount,
			ItemNet:          &itemNet,
			PbSellingPrice:   &req.PbSellingPrice,
			PbListCost:       &zero,
			PbDiscount:       &zero,
			PbCost:           &zero,
			PbStoreTierPrice: &zero,
			LensCost:         &zero,
			AccessoriesCost:  &zero,
		}
		s.db.Create(&pb)
	}

	// Mark SOLD; invoice_id stays = sale invoice
	saleInvoiceID := item.InvoiceID
	item.StatusItemsInventory = "SOLD"
	s.db.Save(&item)

	// Link to vendor invoice in ReceiptsItems
	ri := invModel.ReceiptsItems{
		InvoiceID:   vi.InvoiceID,
		InventoryID: req.InventoryID,
		DateTime:    time.Now(),
	}
	s.db.Create(&ri)

	// Transaction log
	saleInvRef := saleInvNum
	if saleInvRef == "" {
		saleInvRef = fmt.Sprintf("%d", saleInvoiceID)
	}
	notes := fmt.Sprintf("Received ordered item via vendor invoice %d. Originally sold on invoice %s.",
		req.VendorInvoiceID, saleInvRef)
	txn := invModel.InventoryTransaction{
		InventoryID:     &req.InventoryID,
		ToLocationID:    &locID,
		TransferredBy:   empID,
		InvoiceID:       &vi.InvoiceID,
		OldInvoiceID:    &saleInvoiceID,
		StatusItems:     "SOLD",
		TransactionType: "ReceivedOrdered",
		DateTransaction: time.Now(),
		Notes:           &notes,
	}
	s.db.Create(&txn)

	return &ReceiveOrderedResult{
		Message:     "Ordered item received and marked as SOLD",
		IDInventory: req.InventoryID,
		SKU:         item.SKU,
		PriceBook: map[string]interface{}{
			"item_list_cost":   fmtFloat(req.ItemListCost),
			"item_discount":    fmtFloat(req.ItemDiscount),
			"item_net":         fmtFloat(itemNet),
			"pb_selling_price": fmtFloat(req.PbSellingPrice),
		},
		Warning:              warning,
		OriginalSellingPrice: origPrice,
		SaleInvoiceNumber:    saleInvNum,
	}, nil
}

func fmtFloat(f float64) string    { return fmt.Sprintf("%.2f", f) }
func fmtFloatPtr(f *float64) string {
	if f == nil {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", *f)
}
