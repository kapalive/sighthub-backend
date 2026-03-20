package inventory_service

import (
	"fmt"
	"strings"
)

// GetVendorsWithBrands returns vendors that have frames=true and VendorBrand relationships.
func (s *Service) GetVendorsWithBrands() ([]map[string]interface{}, error) {
	var rows []struct {
		VendorID   int64  `gorm:"column:id_vendor"`
		VendorName string `gorm:"column:vendor_name"`
	}
	err := s.db.Raw(`
		SELECT DISTINCT v.id_vendor, v.vendor_name
		FROM vendor v
		JOIN vendor_brand vb ON v.id_vendor = vb.id_vendor
		WHERE v.frames = true
		ORDER BY v.vendor_name
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		out[i] = map[string]interface{}{
			"vendor_id":   r.VendorID,
			"vendor_name": r.VendorName,
		}
	}
	return out, nil
}

// GetBrands returns all frame brands.
func (s *Service) GetBrands() ([]map[string]interface{}, error) {
	var rows []struct {
		BrandID   int64  `gorm:"column:id_brand"`
		BrandName string `gorm:"column:brand_name"`
	}
	err := s.db.Raw(`SELECT id_brand, brand_name FROM brand`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		out[i] = map[string]interface{}{
			"brand_id":   r.BrandID,
			"brand_name": r.BrandName,
		}
	}
	return out, nil
}

// GetStores returns locations by permitted IDs.
func (s *Service) GetStores(locationIDs []int64) ([]map[string]interface{}, error) {
	var rows []struct {
		LocationID int64  `gorm:"column:id_location"`
		FullName   string `gorm:"column:full_name"`
	}
	err := s.db.Raw(`SELECT id_location, full_name FROM location WHERE id_location IN ?`, locationIDs).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		out[i] = map[string]interface{}{
			"location_id": r.LocationID,
			"location":    r.FullName,
		}
	}
	return out, nil
}

// GetProductsByBrand returns products for a brand.
func (s *Service) GetProductsByBrand(brandID int64) ([]map[string]interface{}, error) {
	var rows []struct {
		ProductID    int64  `gorm:"column:id_product"`
		TitleProduct string `gorm:"column:title_product"`
	}
	err := s.db.Raw(`SELECT id_product, title_product FROM product WHERE brand_id = ? ORDER BY title_product`, brandID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		out[i] = map[string]interface{}{
			"product_id":    r.ProductID,
			"title_product": r.TitleProduct,
		}
	}
	return out, nil
}

// GetVariantsByProduct returns model variants for a product.
func (s *Service) GetVariantsByProduct(productID int64) ([]map[string]interface{}, error) {
	var rows []struct {
		VariantID    int64  `gorm:"column:id_model"`
		TitleVariant string `gorm:"column:title_variant"`
	}
	err := s.db.Raw(`SELECT id_model, title_variant FROM model WHERE product_id = ? ORDER BY title_variant`, productID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		out[i] = map[string]interface{}{
			"variant_id":    r.VariantID,
			"title_variant": r.TitleVariant,
		}
	}
	return out, nil
}

// GetItemStatuses returns the available inventory item statuses.
func (s *Service) GetItemStatuses() []string {
	return []string{
		"Ready for Sale", "Defective", "On Return",
		"ICT (to receive in MN)", "ICT (sent and not received)",
		"SOLD", "Missing", "Removed",
	}
}

// SearchProducts searches products by title prefix and groups by location.
func (s *Service) SearchProducts(query string) ([]map[string]interface{}, error) {
	var rows []struct {
		IDInventory          int64   `gorm:"column:id_inventory"`
		SKU                  string  `gorm:"column:sku"`
		CreatedDate          string  `gorm:"column:created_date"`
		StatusItemsInventory string  `gorm:"column:status_items_inventory"`
		LocationName         string  `gorm:"column:location_name"`
		BrandName            string  `gorm:"column:brand_name"`
		TitleProduct         string  `gorm:"column:title_product"`
		TitleVariant         string  `gorm:"column:title_variant"`
		VendorName           string  `gorm:"column:vendor_name"`
		PbSellingPrice       *string `gorm:"column:pb_selling_price"`
	}
	err := s.db.Raw(`
		SELECT i.id_inventory, i.sku, i.created_date, i.status_items_inventory,
		       l.full_name AS location_name,
		       b.brand_name, p.title_product, m.title_variant, v.vendor_name,
		       pb.pb_selling_price::text AS pb_selling_price
		FROM inventory i
		JOIN location l ON i.location_id = l.id_location
		JOIN model m ON i.model_id = m.id_model
		JOIN product p ON m.product_id = p.id_product
		JOIN brand b ON p.brand_id = b.id_brand
		JOIN vendor v ON p.vendor_id = v.id_vendor
		LEFT JOIN price_book pb ON pb.inventory_id = i.id_inventory
		WHERE LOWER(p.title_product) LIKE ?
	`, strings.ToLower(query)+"%").Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]map[string]interface{})
	for _, r := range rows {
		item := map[string]interface{}{
			"brand_name":             r.BrandName,
			"created_date":           r.CreatedDate,
			"id_inventory":           r.IDInventory,
			"pb_selling_price":       r.PbSellingPrice,
			"sku":                    r.SKU,
			"status_items_inventory": r.StatusItemsInventory,
			"location":              r.LocationName,
			"title_product":          r.TitleProduct,
			"title_variant":          r.TitleVariant,
			"vendor_name":            r.VendorName,
		}
		grouped[r.LocationName] = append(grouped[r.LocationName], item)
	}

	out := make([]map[string]interface{}, 0, len(grouped))
	for loc, items := range grouped {
		out = append(out, map[string]interface{}{
			"location":  loc,
			"item_list": items,
		})
	}
	return out, nil
}

// SearchModel searches products by title and returns models with location stock info.
func (s *Service) SearchModel(query string) ([]map[string]interface{}, error) {
	var rows []struct {
		BrandName    string `gorm:"column:brand_name"`
		ProductID    int64  `gorm:"column:id_product"`
		TitleProduct string `gorm:"column:title_product"`
		LocationID   int64  `gorm:"column:location_id"`
		LocationName string `gorm:"column:full_name"`
		ShortName    string `gorm:"column:short_name"`
		Qty          int    `gorm:"column:qty"`
	}
	err := s.db.Raw(`
		SELECT b.brand_name, p.id_product, p.title_product,
		       l.id_location AS location_id, l.full_name, l.short_name,
		       COUNT(i.id_inventory) AS qty
		FROM product p
		JOIN brand b ON p.brand_id = b.id_brand
		JOIN model m ON m.product_id = p.id_product
		JOIN inventory i ON i.model_id = m.id_model
		JOIN location l ON i.location_id = l.id_location
		WHERE LOWER(p.title_product) LIKE ?
		GROUP BY b.brand_name, p.id_product, p.title_product, l.id_location, l.full_name, l.short_name
	`, strings.ToLower(query)+"%").Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	type productKey struct{ pid int64 }
	type locInfo struct {
		LocationID   int64  `json:"location_id"`
		LocationName string `json:"location_name"`
		ShortName    string `json:"short_name"`
		Quantity     int    `json:"quantity"`
	}
	type productInfo struct {
		BrandName    string
		TitleProduct string
		ProductID    int64
		Locations    []locInfo
	}
	m := make(map[int64]*productInfo)
	var order []int64
	for _, r := range rows {
		pi, ok := m[r.ProductID]
		if !ok {
			pi = &productInfo{BrandName: r.BrandName, TitleProduct: r.TitleProduct, ProductID: r.ProductID}
			m[r.ProductID] = pi
			order = append(order, r.ProductID)
		}
		pi.Locations = append(pi.Locations, locInfo{
			LocationID: r.LocationID, LocationName: r.LocationName, ShortName: r.ShortName, Quantity: r.Qty,
		})
	}

	out := make([]map[string]interface{}, 0, len(order))
	for _, pid := range order {
		pi := m[pid]
		locs := make([]map[string]interface{}, len(pi.Locations))
		for j, l := range pi.Locations {
			locs[j] = map[string]interface{}{
				"location_id":   l.LocationID,
				"location_name": l.LocationName,
				"short_name":    l.ShortName,
				"quantity":      l.Quantity,
			}
		}
		out = append(out, map[string]interface{}{
			"brand_name": pi.BrandName,
			"product":    pi.TitleProduct,
			"product_id": pi.ProductID,
			"locations":  locs,
		})
	}
	return out, nil
}

// GetStockByModel returns inventory items for a given model_id.
func (s *Service) GetStockByModel(modelID int64) ([]map[string]interface{}, error) {
	var rows []struct {
		SKU            string  `gorm:"column:sku"`
		LocationID     int64   `gorm:"column:location_id"`
		LocationName   string  `gorm:"column:full_name"`
		VariantID      int64   `gorm:"column:id_model"`
		VariantTitle   string  `gorm:"column:title_variant"`
		ProductID      int64   `gorm:"column:id_product"`
		ProductTitle   string  `gorm:"column:title_product"`
		BrandName      *string `gorm:"column:brand_name"`
		PbSellingPrice *string `gorm:"column:pb_selling_price"`
	}
	err := s.db.Raw(`
		SELECT i.sku, i.location_id, l.full_name, m.id_model, m.title_variant,
		       p.id_product, p.title_product, b.brand_name,
		       pb.pb_selling_price::text AS pb_selling_price
		FROM inventory i
		JOIN location l ON i.location_id = l.id_location
		JOIN model m ON i.model_id = m.id_model
		JOIN product p ON m.product_id = p.id_product
		LEFT JOIN brand b ON p.brand_id = b.id_brand
		LEFT JOIN price_book pb ON pb.inventory_id = i.id_inventory
		WHERE i.model_id = ?
	`, modelID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		out[i] = map[string]interface{}{
			"sku":            r.SKU,
			"location_id":    r.LocationID,
			"location_name":  r.LocationName,
			"variant_id":     r.VariantID,
			"variant_title":  r.VariantTitle,
			"product_id":     r.ProductID,
			"product_title":  r.ProductTitle,
			"brand":          r.BrandName,
			"selling_price":  r.PbSellingPrice,
		}
	}
	return out, nil
}

// GetInventoryReceipt returns receipt history for a model at a location.
func (s *Service) GetInventoryReceipt(productID, locationID int64) ([]map[string]interface{}, error) {
	var rows []struct {
		SKU          string `gorm:"column:sku"`
		IDInventory  int64  `gorm:"column:id_inventory"`
		StatusItems  string `gorm:"column:status_items_inventory"`
		LocationName string `gorm:"column:location_name"`
		ReceiveDate  string `gorm:"column:receive_date"`
	}
	err := s.db.Raw(`
		SELECT i.sku, i.id_inventory, i.status_items_inventory,
		       l.full_name AS location_name,
		       it.date_transaction::text AS receive_date
		FROM inventory i
		JOIN inventory_transaction it ON i.id_inventory = it.inventory_id
		JOIN location l ON i.location_id = l.id_location
		WHERE i.model_id = ?
		  AND it.to_location_id = ?
		  AND it.transaction_type = 'Received'
		ORDER BY it.date_transaction DESC
	`, productID, locationID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		out[i] = map[string]interface{}{
			"sku":           r.SKU,
			"id_inventory":  r.IDInventory,
			"status_items":  r.StatusItems,
			"location_name": r.LocationName,
			"receive_date":  r.ReceiveDate,
		}
	}
	return out, nil
}

// CalcSellPrice calculates selling price based on pricing rules.
func (s *Service) CalcSellPrice(brandType string, brandID int64, listPrice float64) (map[string]interface{}, error) {
	// Query the pricing rule for this brand type + brand_id
	var rule struct {
		Multiplier *float64 `gorm:"column:multiplier"`
	}
	err := s.db.Raw(`
		SELECT multiplier FROM pricing_rule
		WHERE brand_type = ? AND brand_id = ?
		LIMIT 1
	`, brandType, brandID).Scan(&rule).Error
	if err != nil || rule.Multiplier == nil {
		return nil, fmt.Errorf("no pricing rule found")
	}
	sellingPrice := listPrice * (*rule.Multiplier)
	return map[string]interface{}{
		"selling_price": fmt.Sprintf("%.2f", sellingPrice),
		"note":          nil,
	}, nil
}
