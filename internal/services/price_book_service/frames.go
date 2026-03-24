package price_book_service

import (
	"fmt"
	"strings"

	"sighthub-backend/internal/models/frames"
	"sighthub-backend/internal/models/inventory"
)

// ─── Result types ─────────────────────────────────────────────────────────────

type VendorBrandResult struct {
	VendorID   int    `json:"vendor_id"`
	VendorName string `json:"vendor_name"`
	BrandID    int    `json:"brand_id"`
	BrandName  string `json:"brand_name"`
}

type PBProductResult struct {
	ProductID    int64  `json:"product_id"`
	TitleProduct string `json:"title_product"`
}

type PBModelResult struct {
	IDInventory      int64   `json:"id_inventory"`
	SKU              string  `json:"sku"`
	TitleProduct     string  `json:"title_product"`
	ProductID        int64   `json:"product_id"`
	TitleVariant     string  `json:"title_variant"`
	Sun              *bool   `json:"sun"`
	UPC              *string `json:"upc"`
	Color            *string `json:"color"`
	SizeLensWidth    *string `json:"size_lens_width"`
	SizeBridgeWidth  *string `json:"size_bridge_width"`
	SizeTempleLength *string `json:"size_temple_length"`
	MfgNumber        *string `json:"mfg_number"`
	MfrSerialNumber  *string `json:"mfr_serial_number"`
	Accessories      *string `json:"accessories"`
	ListCost         string  `json:"list_cost"`
	Discount         string  `json:"discount"`
	NetPrice         string  `json:"net_price"`
	SellingPrice     string  `json:"selling_price"`
	LensCost         string  `json:"lens_cost"`
	AccessoriesCost  string  `json:"accessories_cost"`
	LocationName     string  `json:"location_name"`
	PbKey            string  `json:"pb_key"`
}

type CustomGlassesInput struct {
	TitleVariant     *string
	LensColor        *string
	LensMaterial     *string
	SizeLensWidth    *string
	SizeBridgeWidth  *string
	SizeTempleLength *string
	Sunglass         *bool
	Photo            *bool
	Polor            *bool
	Mirror           *bool
	BacksideAR       *bool
	UPC              *string
	Accessories      *string
	TypeProduct      *string
	BrandID          *int64
	// PriceBook fields (optional)
	ItemListCost    *float64
	ItemDiscount    *float64
	ItemNet         *float64
	PbSellingPrice  *float64
	LensCost        *float64
	AccessoriesCost *float64
}

// ─── Methods ──────────────────────────────────────────────────────────────────

func (s *Service) GetPBVendorBrandCombinations() ([]VendorBrandResult, error) {
	type row struct {
		IDVendor   int
		VendorName string
		IDBrand    int
		BrandName  string
	}
	var rows []row
	err := s.db.Table("vendor_brand vb").
		Select("v.id_vendor, v.vendor_name, b.id_brand, b.brand_name").
		Joins("JOIN vendor v ON v.id_vendor = vb.id_vendor").
		Joins("JOIN brand b ON b.id_brand = vb.id_brand").
		Order("b.brand_name").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]VendorBrandResult, len(rows))
	for i, r := range rows {
		result[i] = VendorBrandResult{r.IDVendor, r.VendorName, r.IDBrand, r.BrandName}
	}
	return result, nil
}

func (s *Service) GetPBProducts(vendorID, brandID *int64) ([]PBProductResult, error) {
	q := s.db.Table("product p").
		Select("DISTINCT p.id_product, p.title_product").
		Joins("JOIN model m ON m.product_id = p.id_product").
		Joins("JOIN inventory i ON i.model_id = m.id_model")
	if vendorID != nil {
		q = q.Where("p.vendor_id = ?", *vendorID)
	}
	if brandID != nil {
		q = q.Where("p.brand_id = ?", *brandID)
	}
	type row struct {
		IDProduct    int64
		TitleProduct string
	}
	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]PBProductResult, len(rows))
	for i, r := range rows {
		result[i] = PBProductResult{r.IDProduct, r.TitleProduct}
	}
	return result, nil
}

func (s *Service) GetPBModels(productID int64) ([]PBModelResult, error) {
	type row struct {
		IDInventory      int64
		SKU              string
		IDProduct        int64
		TitleProduct     string
		TitleVariant     string
		Sunglass         *bool
		UPC              *string
		Color            *string
		SizeLensWidth    *string
		SizeBridgeWidth  *string
		SizeTempleLength *string
		MfgNumber        *string
		MfrSerialNumber  *string
		Accessories      *string
		ItemListCost     *float64
		ItemDiscount     *float64
		ItemNet          *float64
		PbSellingPrice   *float64
		LensCost         *float64
		AccessoriesCost  *float64
		LocationName     string
	}

	var rows []row
	err := s.db.Table("inventory inv").
		Select(`inv.id_inventory, inv.sku,
			p.id_product, p.title_product,
			m.title_variant, m.sunglass, m.upc, m.color,
			m.size_lens_width, m.size_bridge_width, m.size_temple_length,
			m.mfg_number, m.mfr_serial_number, m.accessories,
			pb.item_list_cost, pb.item_discount, pb.item_net,
			pb.pb_selling_price, pb.lens_cost, pb.accessories_cost,
			l.full_name AS location_name`).
		Joins("JOIN model m ON m.id_model = inv.model_id").
		Joins("JOIN product p ON p.id_product = m.product_id").
		Joins("JOIN location l ON l.id_location = inv.location_id").
		Joins("LEFT JOIN price_book pb ON pb.inventory_id = inv.id_inventory").
		Where("p.id_product = ?", productID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]PBModelResult, len(rows))
	for i, r := range rows {
		result[i] = PBModelResult{
			IDInventory:      r.IDInventory,
			SKU:              r.SKU,
			TitleProduct:     r.TitleProduct,
			ProductID:        r.IDProduct,
			TitleVariant:     r.TitleVariant,
			Sun:              r.Sunglass,
			UPC:              r.UPC,
			Color:            r.Color,
			SizeLensWidth:    r.SizeLensWidth,
			SizeBridgeWidth:  r.SizeBridgeWidth,
			SizeTempleLength: r.SizeTempleLength,
			MfgNumber:        r.MfgNumber,
			MfrSerialNumber:  r.MfrSerialNumber,
			Accessories:      r.Accessories,
			ListCost:         strOrZero(r.ItemListCost),
			Discount:         strOrZero(r.ItemDiscount),
			NetPrice:         strOrZero(r.ItemNet),
			SellingPrice:     strOrZero(r.PbSellingPrice),
			LensCost:         strOrZero(r.LensCost),
			AccessoriesCost:  strOrZero(r.AccessoriesCost),
			LocationName:     r.LocationName,
			PbKey:            "Frames",
		}
	}
	return result, nil
}

func (s *Service) CreateCustomGlasses(inventoryID int, employeeID int64, in CustomGlassesInput) (int64, string, error) {
	var inv inventory.Inventory
	if err := s.db.First(&inv, inventoryID).Error; err != nil {
		return 0, "", fmt.Errorf("inventory not found")
	}
	if inv.ModelID == nil {
		return 0, "", fmt.Errorf("no existing model for this inventory")
	}

	var oldModel frames.Model
	if err := s.db.First(&oldModel, *inv.ModelID).Error; err != nil {
		return 0, "", fmt.Errorf("model not found")
	}

	var oldProduct frames.Product
	if err := s.db.First(&oldProduct, oldModel.ProductID).Error; err != nil {
		return 0, "", fmt.Errorf("source product not found")
	}

	// resolve type_product
	allowedTypes := map[string]struct{}{"eyeglasses": {}, "sunglasses": {}, "safety_glasses": {}}
	typeAlias := map[string]string{"sun": "sunglasses", "optical": "eyeglasses", "rx": "eyeglasses", "safety": "safety_glasses"}

	requestedType := ""
	if in.TypeProduct != nil {
		rt := strings.ToLower(strings.TrimSpace(*in.TypeProduct))
		if alias, ok := typeAlias[rt]; ok {
			rt = alias
		}
		if _, ok := allowedTypes[rt]; !ok {
			return 0, "", fmt.Errorf("invalid type_product: allowed eyeglasses, sunglasses, safety_glasses")
		}
		requestedType = rt
	}

	// build new model from existing
	newModel := frames.Model{
		ProductID:        oldModel.ProductID,
		TitleVariant:     oldModel.TitleVariant,
		LensColor:        oldModel.LensColor,
		LensMaterial:     oldModel.LensMaterial,
		SizeLensWidth:    oldModel.SizeLensWidth,
		SizeBridgeWidth:  oldModel.SizeBridgeWidth,
		SizeTempleLength: oldModel.SizeTempleLength,
		Sunglass:         oldModel.Sunglass,
		Photo:            oldModel.Photo,
		Polor:            oldModel.Polor,
		Mirror:           oldModel.Mirror,
		BacksideAR:       oldModel.BacksideAR,
		UPC:              oldModel.UPC,
		Accessories:      oldModel.Accessories,
	}

	// apply overrides
	if in.TitleVariant != nil {
		newModel.TitleVariant = *in.TitleVariant
	}
	if in.LensColor != nil {
		newModel.LensColor = in.LensColor
	}
	if in.SizeLensWidth != nil {
		newModel.SizeLensWidth = in.SizeLensWidth
	}
	if in.SizeBridgeWidth != nil {
		newModel.SizeBridgeWidth = in.SizeBridgeWidth
	}
	if in.SizeTempleLength != nil {
		newModel.SizeTempleLength = in.SizeTempleLength
	}
	if in.Sunglass != nil {
		newModel.Sunglass = in.Sunglass
	}
	if in.Photo != nil {
		newModel.Photo = in.Photo
	}
	if in.Polor != nil {
		newModel.Polor = in.Polor
	}
	if in.Mirror != nil {
		newModel.Mirror = *in.Mirror
	}
	if in.BacksideAR != nil {
		newModel.BacksideAR = *in.BacksideAR
	}
	if in.UPC != nil {
		newModel.UPC = in.UPC
	}
	if in.Accessories != nil {
		newModel.Accessories = in.Accessories
	}

	// handle type/brand change: find or create product with new type and/or brand
	targetProductID := oldProduct.IDProduct
	linkedType := oldProduct.TypeProduct
	targetBrandID := oldProduct.BrandID
	targetType := oldProduct.TypeProduct

	if requestedType != "" && requestedType != oldProduct.TypeProduct {
		targetType = requestedType
		linkedType = requestedType
		sg := requestedType == "sunglasses"
		newModel.Sunglass = &sg
	}
	if in.BrandID != nil {
		targetBrandID = in.BrandID
	}

	brandChanged := in.BrandID != nil && (oldProduct.BrandID == nil || *in.BrandID != *oldProduct.BrandID)
	typeChanged := targetType != oldProduct.TypeProduct

	if brandChanged || typeChanged {
		var existing frames.Product
		err := s.db.Where("title_product = ? AND brand_id = ? AND vendor_id = ? AND type_product = ?",
			oldProduct.TitleProduct, targetBrandID, oldProduct.VendorID, targetType).
			First(&existing).Error
		if err == nil {
			targetProductID = existing.IDProduct
		} else {
			np := frames.Product{
				TitleProduct: oldProduct.TitleProduct,
				BrandID:      targetBrandID,
				VendorID:     oldProduct.VendorID,
				TypeProduct:  targetType,
			}
			if err := s.db.Create(&np).Error; err != nil {
				return 0, "", err
			}
			targetProductID = np.IDProduct
		}
	}
	newModel.ProductID = int64(targetProductID)

	// build CUSTOM suffix
	tv := strings.TrimSpace(newModel.TitleVariant)
	acc := ""
	if newModel.Accessories != nil {
		acc = strings.TrimSpace(*newModel.Accessories)
	}
	if acc != "" {
		newModel.TitleVariant = fmt.Sprintf("%s - CUSTOM %s", tv, acc)
	} else {
		newModel.TitleVariant = fmt.Sprintf("%s - CUSTOM", tv)
	}

	tx := s.db.Begin()
	if err := tx.Create(&newModel).Error; err != nil {
		tx.Rollback()
		return 0, "", err
	}

	// update inventory
	if err := tx.Model(&inv).Update("model_id", newModel.IDModel).Error; err != nil {
		tx.Rollback()
		return 0, "", err
	}

	// record transaction
	var lastTx inventory.InventoryTransaction
	hasLastTx := tx.Where("inventory_id = ?", inventoryID).
		Order("id_transaction DESC").First(&lastTx).Error == nil

	fromLoc := &inv.LocationID
	toLoc := &inv.LocationID
	invoiceID := &inv.InvoiceID
	if hasLastTx {
		fromLoc = lastTx.FromLocationID
		toLoc = lastTx.ToLocationID
		invoiceID = lastTx.InvoiceID
	}

	notes := fmt.Sprintf("old_model_id=%d;new_model_id=%d", *inv.ModelID, newModel.IDModel)
	itx := inventory.InventoryTransaction{
		InventoryID:     &inv.IDInventory,
		FromLocationID:  fromLoc,
		ToLocationID:    toLoc,
		TransferredBy:   employeeID,
		InvoiceID:       invoiceID,
		StatusItems:     inv.StatusItemsInventory,
		TransactionType: "CustomModel",
		Notes:           &notes,
	}
	if hasLastTx && lastTx.OldInvoiceID != nil {
		itx.OldInvoiceID = lastTx.OldInvoiceID
	}
	if err := tx.Create(&itx).Error; err != nil {
		tx.Rollback()
		return 0, "", err
	}

	// optionally update PriceBook prices
	var pb inventory.PriceBook
	if tx.Where("inventory_id = ?", inventoryID).First(&pb).Error == nil {
		updates := map[string]interface{}{}
		if in.ItemListCost != nil {
			updates["item_list_cost"] = *in.ItemListCost
		}
		if in.ItemDiscount != nil {
			updates["item_discount"] = *in.ItemDiscount
		}
		if in.ItemNet != nil {
			updates["item_net"] = *in.ItemNet
		}
		if in.PbSellingPrice != nil {
			updates["pb_selling_price"] = *in.PbSellingPrice
		}
		if in.LensCost != nil {
			updates["lens_cost"] = *in.LensCost
		}
		if in.AccessoriesCost != nil {
			updates["accessories_cost"] = *in.AccessoriesCost
		}
		if len(updates) > 0 {
			if err := tx.Model(&pb).Updates(updates).Error; err != nil {
				tx.Rollback()
				return 0, "", err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return 0, "", err
	}
	return newModel.IDModel, linkedType, nil
}

func (s *Service) RevertCustomGlasses(inventoryID int, employeeID int64, pbPrices map[string]*float64) (int64, int64, error) {
	var inv inventory.Inventory
	if err := s.db.First(&inv, inventoryID).Error; err != nil {
		return 0, 0, fmt.Errorf("inventory not found")
	}

	var customTx inventory.InventoryTransaction
	if err := s.db.Where("inventory_id = ? AND transaction_type = 'CustomModel'", inventoryID).
		Order("id_transaction DESC").First(&customTx).Error; err != nil {
		return 0, 0, nil // already regular — caller checks
	}

	// parse old_model_id from notes
	var oldModelID int64
	if customTx.Notes != nil {
		for _, part := range strings.Split(*customTx.Notes, ";") {
			if strings.HasPrefix(part, "old_model_id=") {
				fmt.Sscanf(strings.TrimPrefix(part, "old_model_id="), "%d", &oldModelID)
			}
		}
	}
	if oldModelID == 0 {
		return 0, 0, fmt.Errorf("cannot determine original model ID from history")
	}

	var oldModel frames.Model
	if err := s.db.First(&oldModel, oldModelID).Error; err != nil {
		return 0, 0, fmt.Errorf("original model (ID=%d) no longer exists", oldModelID)
	}

	currentModelID := int64(0)
	if inv.ModelID != nil {
		currentModelID = *inv.ModelID
	}

	tx := s.db.Begin()
	if err := tx.Model(&inv).Update("model_id", oldModelID).Error; err != nil {
		tx.Rollback()
		return 0, 0, err
	}

	var pb inventory.PriceBook
	if tx.Where("inventory_id = ?", inventoryID).First(&pb).Error == nil && len(pbPrices) > 0 {
		updates := map[string]interface{}{}
		for k, v := range pbPrices {
			if v != nil {
				updates[k] = *v
			}
		}
		if len(updates) > 0 {
			tx.Model(&pb).Updates(updates)
		}
	}

	var lastTx inventory.InventoryTransaction
	hasLastTx := tx.Where("inventory_id = ?", inventoryID).
		Order("id_transaction DESC").First(&lastTx).Error == nil

	notes := fmt.Sprintf("old_model_id=%d;restored_model_id=%d", currentModelID, oldModelID)
	revertTx := inventory.InventoryTransaction{
		InventoryID:     &inv.IDInventory,
		FromLocationID:  &inv.LocationID,
		ToLocationID:    &inv.LocationID,
		TransferredBy:   employeeID,
		InvoiceID:       &inv.InvoiceID,
		StatusItems:     inv.StatusItemsInventory,
		TransactionType: "RevertToRegular",
		Notes:           &notes,
	}
	if hasLastTx {
		revertTx.FromLocationID = lastTx.FromLocationID
		revertTx.ToLocationID = lastTx.ToLocationID
		revertTx.InvoiceID = lastTx.InvoiceID
		if lastTx.OldInvoiceID != nil {
			revertTx.OldInvoiceID = lastTx.OldInvoiceID
		}
	}
	if err := tx.Create(&revertTx).Error; err != nil {
		tx.Rollback()
		return 0, 0, err
	}

	if err := tx.Commit().Error; err != nil {
		return 0, 0, err
	}
	return oldModelID, currentModelID, nil
}
