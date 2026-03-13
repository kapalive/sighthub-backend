package inventory_service

import (
	"fmt"
	"math"
	"strings"
)

type FilterParams struct {
	LocationIDs []int64
	VendorIDs   []int64
	BrandIDs    []int64
	ProductIDs  []int64
	ModelIDs    []int64
	Statuses    []string
	VendorNames []string
	BrandNames  []string
	GroupBy     string // "sku" or "model"
	Output      string // "json", "csv", "csv_detail"
	Page        int
	PerPage     int
}

type FilterResult struct {
	TotalItems  int                      `json:"total_items"`
	TotalPages  int                      `json:"total_pages"`
	CurrentPage int                      `json:"current_page"`
	PerPage     int                      `json:"per_page"`
	Items       []map[string]interface{} `json:"items"`
}

type CSVRow struct {
	IDInventory          int64   `gorm:"column:id_inventory"`
	SKU                  string  `gorm:"column:sku"`
	CreatedDate          string  `gorm:"column:created_date"`
	StatusItemsInventory string  `gorm:"column:status_items_inventory"`
	IDLocation           int64   `gorm:"column:id_location"`
	LocationName         string  `gorm:"column:location_name"`
	IDVendor             int64   `gorm:"column:id_vendor"`
	VendorName           string  `gorm:"column:vendor_name"`
	IDBrand              int64   `gorm:"column:id_brand"`
	BrandName            string  `gorm:"column:brand_name"`
	IDProduct            int64   `gorm:"column:id_product"`
	TitleProduct         string  `gorm:"column:title_product"`
	IDModel              int64   `gorm:"column:id_model"`
	TitleVariant         string  `gorm:"column:title_variant"`
	Username             string  `gorm:"column:username"`
	EmployeeName         string  `gorm:"column:employee_name"`
	PbSellingPrice       *string `gorm:"column:pb_selling_price"`
	EmployeeID           *int64  `gorm:"column:id_employee"`
}

func (s *Service) buildFilterQuery(params FilterParams, effectiveLocationIDs []int64) (string, []interface{}) {
	q := `
		FROM inventory i
		JOIN location l ON i.location_id = l.id_location
		JOIN model m ON i.model_id = m.id_model
		JOIN product p ON m.product_id = p.id_product
		JOIN brand b ON p.brand_id = b.id_brand
		JOIN vendor v ON p.vendor_id = v.id_vendor
		JOIN employee e ON i.employee_id = e.id_employee
		JOIN employee_login el ON e.employee_login_id = el.id_employee_login
		JOIN price_book pb ON pb.inventory_id = i.id_inventory
		WHERE l.id_location IN ?
	`
	args := []interface{}{effectiveLocationIDs}

	if len(params.VendorIDs) > 0 {
		q += ` AND v.id_vendor IN ?`
		args = append(args, params.VendorIDs)
	}
	if len(params.BrandIDs) > 0 {
		q += ` AND b.id_brand IN ?`
		args = append(args, params.BrandIDs)
	}
	if len(params.ProductIDs) > 0 {
		q += ` AND p.id_product IN ?`
		args = append(args, params.ProductIDs)
	}
	if len(params.ModelIDs) > 0 {
		q += ` AND m.id_model IN ?`
		args = append(args, params.ModelIDs)
	}
	if len(params.Statuses) > 0 {
		q += ` AND i.status_items_inventory IN ?`
		args = append(args, params.Statuses)
	}
	if len(params.VendorNames) > 0 {
		q += ` AND v.vendor_name IN ?`
		args = append(args, params.VendorNames)
	}
	if len(params.BrandNames) > 0 {
		q += ` AND b.brand_name IN ?`
		args = append(args, params.BrandNames)
	}
	return q, args
}

// GetInventoryCSV returns all matching rows for CSV export.
func (s *Service) GetInventoryCSV(params FilterParams, effectiveLocationIDs []int64) ([]CSVRow, error) {
	fromWhere, args := s.buildFilterQuery(params, effectiveLocationIDs)
	q := `SELECT i.id_inventory, i.sku, i.created_date::text AS created_date, i.status_items_inventory,
	       l.id_location, l.full_name AS location_name,
	       v.id_vendor, v.vendor_name, b.id_brand, b.brand_name,
	       p.id_product, p.title_product, m.id_model, m.title_variant,
	       el.username, CONCAT(e.first_name, ' ', e.last_name) AS employee_name,
	       pb.pb_selling_price::text AS pb_selling_price ` + fromWhere

	var rows []CSVRow
	if err := s.db.Raw(q, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// GetInventoryByFilter returns paginated, optionally grouped inventory.
func (s *Service) GetInventoryByFilter(params FilterParams, effectiveLocationIDs []int64) (*FilterResult, error) {
	fromWhere, args := s.buildFilterQuery(params, effectiveLocationIDs)

	// Count total
	var total int64
	countQ := `SELECT COUNT(*) ` + fromWhere
	if err := s.db.Raw(countQ, args...).Scan(&total).Error; err != nil {
		return nil, err
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	perPage := params.PerPage
	if perPage < 1 {
		perPage = 25
	}

	selectQ := `SELECT i.id_inventory, i.sku, i.created_date::text AS created_date, i.status_items_inventory,
	       l.id_location, l.full_name AS location_name,
	       v.id_vendor, v.vendor_name, b.id_brand, b.brand_name,
	       p.id_product, p.title_product, m.id_model, m.title_variant,
	       e.id_employee, el.username, pb.pb_selling_price::text AS pb_selling_price ` +
		fromWhere + fmt.Sprintf(` LIMIT %d OFFSET %d`, perPage, (page-1)*perPage)

	var rows []CSVRow
	if err := s.db.Raw(selectQ, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	var items []map[string]interface{}
	if params.GroupBy == "model" {
		grouped := make(map[int64]*struct {
			TitleVariant string
			Qty          int
			Items        []map[string]interface{}
		})
		var order []int64
		for _, r := range rows {
			g, ok := grouped[r.IDModel]
			if !ok {
				g = &struct {
					TitleVariant string
					Qty          int
					Items        []map[string]interface{}
				}{TitleVariant: r.TitleVariant}
				grouped[r.IDModel] = g
				order = append(order, r.IDModel)
			}
			g.Qty++
			g.Items = append(g.Items, rowToMap(r))
		}
		for _, mid := range order {
			g := grouped[mid]
			items = append(items, map[string]interface{}{
				"variant_id":    mid,
				"title_variant": g.TitleVariant,
				"quantity":      g.Qty,
				"items":         g.Items,
			})
		}
	} else {
		for _, r := range rows {
			m := rowToMap(r)
			m["variant_id"] = r.IDModel
			items = append(items, m)
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	return &FilterResult{
		TotalItems:  int(total),
		TotalPages:  totalPages,
		CurrentPage: page,
		PerPage:     perPage,
		Items:       items,
	}, nil
}

func rowToMap(r CSVRow) map[string]interface{} {
	return map[string]interface{}{
		"id_inventory":           r.IDInventory,
		"sku":                    r.SKU,
		"created_date":           r.CreatedDate,
		"status_items_inventory": r.StatusItemsInventory,
		"location_id":            r.IDLocation,
		"location_name":          r.LocationName,
		"vendor_id":              r.IDVendor,
		"vendor_name":            r.VendorName,
		"brand_id":               r.IDBrand,
		"brand_name":             r.BrandName,
		"product_id":             r.IDProduct,
		"title_product":          r.TitleProduct,
		"title_variant":          r.TitleVariant,
		"employee_id":            r.EmployeeID,
		"username":               r.Username,
		"pb_selling_price":       r.PbSellingPrice,
	}
}

// FormatCSV converts rows to CSV string.
func FormatCSV(rows []CSVRow) string {
	var sb strings.Builder
	sb.WriteString("Inventory ID,SKU,Created Date,Status,Location ID,Location Name,Vendor ID,Vendor Name,Brand ID,Brand Name,Product ID,Product Title,Model ID,Model Title,Employee Login,Employee Name,Selling Price\n")
	for _, r := range rows {
		sp := ""
		if r.PbSellingPrice != nil {
			sp = *r.PbSellingPrice
		}
		sb.WriteString(fmt.Sprintf("%d,%s,%s,%s,%d,%s,%d,%s,%d,%s,%d,%s,%d,%s,%s,%s,%s\n",
			r.IDInventory, csvEscape(r.SKU), r.CreatedDate, csvEscape(string(r.StatusItemsInventory)),
			r.IDLocation, csvEscape(r.LocationName), r.IDVendor, csvEscape(r.VendorName),
			r.IDBrand, csvEscape(r.BrandName), r.IDProduct, csvEscape(r.TitleProduct),
			r.IDModel, csvEscape(r.TitleVariant), csvEscape(r.Username), csvEscape(r.EmployeeName), sp))
	}
	return sb.String()
}

func csvEscape(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}
