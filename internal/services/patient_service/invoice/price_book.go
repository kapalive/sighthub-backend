package invoice

import (
	"fmt"
	"regexp"
	"strings"

	lensModel "sighthub-backend/internal/models/lenses"
	"sighthub-backend/internal/models/invoices"
)

// ─── Invoice-scoped Price Book ───────────────────────────────────────────────

// getInvoiceLensSource returns the source of lenses in the invoice ("custom", "vision_web", "zeiss_only")
// or "" if no lenses.
func (s *Service) getInvoiceLensSource(invoiceID int64) string {
	var items []invoices.InvoiceItemSale
	s.db.Where("invoice_id = ? AND item_type = ?", invoiceID, "Lens").Find(&items)
	if len(items) == 0 || items[0].ItemID == nil {
		return ""
	}
	var srcs []string
	s.db.Model(&lensModel.Lenses{}).
		Where("id_lenses = ?", *items[0].ItemID).
		Pluck("source", &srcs)
	if len(srcs) > 0 && srcs[0] != "" {
		return srcs[0]
	}
	return "custom"
}

// ─── Lens list ───────────────────────────────────────────────────────────────

type InvPBLensFilters struct {
	BrandID          *int
	VendorID         *int
	TypeID           *int
	MaterialID       *int
	SpecialFeatureID *int
	SeriesID         *int
	Search           *string
	Page             int
	PerPage          int
}

type InvPBLensItem struct {
	ItemID          int      `json:"item_id"`
	ItemName        string   `json:"item_name"`
	BrandName       *string  `json:"brand_name"`
	TypeName        *string  `json:"type_name"`
	MaterialName    *string  `json:"material_name"`
	Description     *string  `json:"description"`
	SeriesName      *string  `json:"series_name"`
	VCodes          string   `json:"v_codes"`
	SpecialFeatures []string `json:"special_features"`
	Price           *string  `json:"price"`
	Cost            *string  `json:"cost"`
	Source          *string  `json:"source"`
	PbKey           string   `json:"pb_key"`
}

type InvPBLensResponse struct {
	Items      []InvPBLensItem `json:"items"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PerPage    int             `json:"per_page"`
	TotalPages int             `json:"total_pages"`
	Source     string          `json:"source"`
}

func (s *Service) InvoicePBLenses(invoiceID int64, f InvPBLensFilters) (*InvPBLensResponse, error) {
	source := s.getInvoiceLensSource(invoiceID)

	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 {
		f.PerPage = 25
	}

	q := s.db.Model(&lensModel.Lenses{})

	// If additional services exist → only custom lenses
	var addSvcCount int64
	s.db.Model(&invoices.InvoiceItemSale{}).
		Where("invoice_id = ? AND item_type = ?", invoiceID, "Add service").
		Count(&addSvcCount)
	if addSvcCount > 0 {
		source = "custom"
	}

	// Auto-filter by invoice lens source
	if source != "" {
		q = q.Where("source = ?", source)
	}

	if f.BrandID != nil {
		q = q.Where("brand_lens_id = ?", *f.BrandID)
	}
	if f.VendorID != nil {
		q = q.Where("vendor_id = ?", *f.VendorID)
	}
	if f.TypeID != nil {
		q = q.Where("lens_type_id = ?", *f.TypeID)
	}
	if f.MaterialID != nil {
		q = q.Where("lenses_materials_id = ?", *f.MaterialID)
	}
	if f.SeriesID != nil {
		q = q.Where("lens_series_id = ?", *f.SeriesID)
	}
	if f.SpecialFeatureID != nil {
		q = q.Joins("JOIN lenses_feature_relation lfr ON lfr.lenses_id = lenses.id_lenses").
			Where("lfr.lens_special_features_id = ?", *f.SpecialFeatureID)
	}
	if f.Search != nil && *f.Search != "" {
		words := strings.Fields(strings.TrimSpace(*f.Search))
		for _, w := range words {
			safe := regexp.QuoteMeta(w)
			q = q.Where("description ~* ?", `(^|\s)`+safe)
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (f.Page - 1) * f.PerPage
	var rows []lensModel.Lenses
	if err := q.
		Preload("BrandLens").
		Preload("LensType").
		Preload("LensesMaterial").
		Preload("LensSeries").
		Offset(offset).Limit(f.PerPage).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]InvPBLensItem, 0, len(rows))
	for _, l := range rows {
		type vcRow struct{ Code string }
		var vcRows []vcRow
		s.db.Table("v_codes_lens vc").
			Select("vc.code").
			Joins("JOIN lenses_v_codes_relation lvr ON lvr.v_codes_lens_id = vc.id_v_codes_lens").
			Where("lvr.lenses_id = ?", l.IDLenses).
			Scan(&vcRows)
		vcCodes := make([]string, len(vcRows))
		for i, v := range vcRows {
			vcCodes[i] = v.Code
		}

		type sfRow struct{ FeatureName string }
		var sfRows []sfRow
		s.db.Table("lens_special_features sf").
			Select("sf.feature_name").
			Joins("JOIN lenses_feature_relation lfr ON lfr.lens_special_features_id = sf.id_lens_special_features").
			Where("lfr.lenses_id = ?", l.IDLenses).
			Scan(&sfRows)
		sfNames := make([]string, len(sfRows))
		for i, sf := range sfRows {
			sfNames[i] = sf.FeatureName
		}

		var price, cost *string
		if l.Price != nil {
			p := fmt.Sprintf("%.2f", *l.Price*2)
			price = &p
		}
		if l.Cost != nil {
			c := fmt.Sprintf("%.2f", *l.Cost*2)
			cost = &c
		}

		item := InvPBLensItem{
			ItemID:          l.IDLenses,
			ItemName:        l.LensName,
			Description:     l.Description,
			VCodes:          strings.Join(vcCodes, ", "),
			SpecialFeatures: sfNames,
			Price:           price,
			Cost:            cost,
			Source:          l.Source,
			PbKey:           "Lens",
		}
		if l.BrandLens != nil {
			n := l.BrandLens.BrandName
			item.BrandName = &n
		}
		if l.LensType != nil {
			item.TypeName = &l.LensType.TypeName
		}
		if l.LensesMaterial != nil {
			item.MaterialName = &l.LensesMaterial.MaterialName
		}
		if l.LensSeries != nil {
			item.SeriesName = &l.LensSeries.SeriesName
		}
		items = append(items, item)
	}

	totalPages := int(total) / f.PerPage
	if int(total)%f.PerPage != 0 {
		totalPages++
	}

	return &InvPBLensResponse{
		Items:      items,
		Total:      total,
		Page:       f.Page,
		PerPage:    f.PerPage,
		TotalPages: totalPages,
		Source:     source,
	}, nil
}

// ─── Treatment list ──────────────────────────────────────────────────────────

type InvPBTreatmentItem struct {
	ItemID      int64   `json:"item_id"`
	ItemName    string  `json:"item_name"`
	Description *string `json:"description"`
	Price       *string `json:"price"`
	VCode       *string `json:"v_code"`
	Source      *string `json:"source"`
	PbKey       string  `json:"pb_key"`
}

func (s *Service) InvoicePBTreatments(invoiceID int64, search *string, employeeID int64) ([]InvPBTreatmentItem, error) {
	source := s.getInvoiceLensSource(invoiceID)

	// No lenses or custom lenses → no treatments available
	if source == "" || source == "custom" {
		return []InvPBTreatmentItem{}, nil
	}

	// If additional services already in invoice → no treatments
	var addSvcCount int64
	s.db.Model(&invoices.InvoiceItemSale{}).
		Where("invoice_id = ? AND item_type = ?", invoiceID, "Add service").
		Count(&addSvcCount)
	if addSvcCount > 0 {
		return []InvPBTreatmentItem{}, nil
	}

	q := s.db.Table("lens_treatments lt").
		Select("lt.id_lens_treatments, lt.item_nbr, lt.description, lt.price, lt.source, vc.code AS v_code").
		Joins("LEFT JOIN v_codes_lens vc ON vc.id_v_codes_lens = lt.v_codes_lens_id").
		Where("lt.source = ?", source)

	// For zeiss_only — filter by PCAT allowed treatments for the specific lens
	if source == "zeiss_only" && s.zeissAllowedTreatments != nil && employeeID > 0 {
		lensCode := s.getInvoiceZeissLensCode(invoiceID)
		custNum := s.getZeissCustomerNumber(employeeID)
		if lensCode != "" && custNum != "" {
			allowedCodes, err := s.zeissAllowedTreatments(employeeID, lensCode, custNum)
			if err == nil && len(allowedCodes) > 0 {
				q = q.Where("lt.vw_code IN ?", allowedCodes)
			}
		}
	}

	if search != nil && *search != "" {
		words := strings.Fields(strings.TrimSpace(*search))
		for _, w := range words {
			safe := regexp.QuoteMeta(w)
			q = q.Where("lt.description ~* ?", `(^|\s)`+safe)
		}
	}

	type row struct {
		IDLensTreatments int64
		ItemNbr          string
		Description      *string
		Price            *float64
		Source           *string
		VCode            *string
	}
	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]InvPBTreatmentItem, len(rows))
	for i, r := range rows {
		var priceStr *string
		if r.Price != nil {
			p := fmt.Sprintf("%.2f", *r.Price*2)
			priceStr = &p
		}
		result[i] = InvPBTreatmentItem{
			ItemID:      r.IDLensTreatments,
			ItemName:    r.ItemNbr,
			Description: r.Description,
			Price:       priceStr,
			VCode:       r.VCode,
			Source:      r.Source,
			PbKey:       "Treatment",
		}
	}
	return result, nil
}

// getInvoiceZeissLensCode returns the Zeiss commercial code (lens_name) for the lens in invoice
func (s *Service) getInvoiceZeissLensCode(invoiceID int64) string {
	var code string
	s.db.Raw(`
		SELECT l.lens_name
		FROM invoice_item_sale iis
		JOIN lenses l ON l.id_lenses = iis.item_id
		WHERE iis.invoice_id = ? AND iis.item_type = 'Lens' AND l.source = 'zeiss_only'
		LIMIT 1
	`, invoiceID).Scan(&code)
	return code
}

// getZeissCustomerNumber returns zeiss customer_number from zeiss_token for employee
func (s *Service) getZeissCustomerNumber(employeeID int64) string {
	var num string
	s.db.Raw("SELECT customer_number FROM zeiss_token WHERE employee_id = ?", employeeID).Scan(&num)
	return num
}

// ─── Additional service list ─────────────────────────────────────────────────

type InvPBAddServiceItem struct {
	ItemID      int64   `json:"item_id"`
	ItemName    *string `json:"item_name"`
	ServiceType string  `json:"service_type"`
	Description string  `json:"description"`
	Price       string  `json:"price"`
	SrCost      *bool   `json:"sr_cost"`
	Tint        *bool   `json:"tint"`
	AR          *bool   `json:"ar"`
	UV          *bool   `json:"uv"`
	Drill       *bool   `json:"drill"`
	Send        *bool   `json:"send"`
	VCode       *string `json:"v_code"`
	PbKey       string  `json:"pb_key"`
}

func (s *Service) InvoicePBAddServices(invoiceID int64, typeID *int, search *string) ([]InvPBAddServiceItem, error) {
	source := s.getInvoiceLensSource(invoiceID)

	// Additional services only when no lenses or custom lenses
	if source == "vision_web" || source == "zeiss_only" {
		return []InvPBAddServiceItem{}, nil
	}

	q := s.db.Table("additional_service ads").
		Select("ads.id_additional_service, ads.item_number, ast.title AS service_type_title, ads.invoice_desc, ads.price, ads.sr_cost, ads.tint, ads.ar, ads.uv, ads.drill, ads.send, ads.ins_v_code").
		Joins("JOIN add_service_type ast ON ast.id_add_service_type = ads.add_service_type_id")

	if typeID != nil {
		q = q.Where("ads.add_service_type_id = ?", *typeID)
	}
	if search != nil && *search != "" {
		words := strings.Fields(strings.TrimSpace(*search))
		for _, w := range words {
			safe := regexp.QuoteMeta(w)
			q = q.Where("ads.invoice_desc ~* ?", `(^|\s)`+safe)
		}
	}

	type row struct {
		IDAdditionalService int64
		ItemNumber          *string
		ServiceTypeTitle    string
		InvoiceDesc         string
		Price               float64
		SrCost              *bool
		Tint                *bool
		AR                  *bool
		UV                  *bool
		Drill               *bool
		Send                *bool
		InsVCode            *string
	}
	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]InvPBAddServiceItem, len(rows))
	for i, r := range rows {
		result[i] = InvPBAddServiceItem{
			ItemID:      r.IDAdditionalService,
			ItemName:    r.ItemNumber,
			ServiceType: r.ServiceTypeTitle,
			Description: r.InvoiceDesc,
			Price:       fmt.Sprintf("%.2f", r.Price),
			SrCost:      r.SrCost,
			Tint:        r.Tint,
			AR:          r.AR,
			UV:          r.UV,
			Drill:       r.Drill,
			Send:        r.Send,
			VCode:       r.InsVCode,
			PbKey:       "Add service",
		}
	}
	return result, nil
}
