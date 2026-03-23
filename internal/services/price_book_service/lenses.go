package price_book_service

import (
	"fmt"
	"regexp"
	"strings"

	"sighthub-backend/internal/models/lenses"
	vendormodel "sighthub-backend/internal/models/vendors"
)

// ─── Result types ─────────────────────────────────────────────────────────────

type LensBrandVendorResult struct {
	IDBrandLens int     `json:"id_brand_lens"`
	BrandName   string  `json:"brand_name"`
	IDVendor    *int    `json:"id_vendor"`
	VendorName  *string `json:"vendor_name"`
}

type LensTypeResult struct {
	IDLensType int    `json:"id_lens_type"`
	TypeName   string `json:"type_name"`
}

type LensMaterialResult struct {
	IDLensesMaterials int    `json:"id_lenses_materials"`
	MaterialName      string `json:"material_name"`
}

type LensSeriesResult struct {
	IDLensSeries int    `json:"id_lens_series"`
	SeriesName   string `json:"series_name"`
}

type LensSpecialResult struct {
	IDLensSpecialFeatures int    `json:"id_lens_special_features"`
	FeatureName           string `json:"feature_name"`
}

type VCodeResult struct {
	IDVCodesLens int    `json:"id_v_codes_lens"`
	Code         string `json:"code"`
}

type LensListItem struct {
	ItemID         int      `json:"item_id"`
	ItemName       string   `json:"item_name"`
	BrandName      *string  `json:"brand_name"`
	TypeName       *string  `json:"type_name"`
	MaterialName   *string  `json:"material_name"`
	Description    *string  `json:"description"`
	SeriesName     *string  `json:"series_name"`
	VCodes         string   `json:"v_codes"`
	SpecialFeatures []string `json:"special_features"`
	Price          *string  `json:"price"`
	Cost           *string  `json:"cost"`
	Source         *string  `json:"source"`
	PbKey          string   `json:"pb_key"`
}

type LensDetail struct {
	IDLenses    int                    `json:"id_lenses"`
	LensName    string                 `json:"lens_name"`
	Brand       map[string]interface{} `json:"brand"`
	Type        map[string]interface{} `json:"type"`
	Material    map[string]interface{} `json:"material"`
	Series      map[string]interface{} `json:"series"`
	Vendor      map[string]interface{} `json:"vendor"`
	Description *string                `json:"description"`
	VCodes      []map[string]interface{} `json:"v_codes"`
	SpecialFeatures []map[string]interface{} `json:"special_features"`
	Price       string                 `json:"price"`
	Cost        string                 `json:"cost"`
	MfrNumber   *string                `json:"mfr_number"`
	CanLookup   bool                   `json:"can_lookup"`
}

type AddLensInput struct {
	LensName          string
	BrandLensID       int
	LensTypeID        *int
	LensesMaterialsID *int
	LensSeriesID      *int
	Description       *string
	VendorID          *int
	Price             float64
	Cost              float64
	VCodes            []int
	SpecialFeatures   []int
}

type UpdateLensInput struct {
	LensName          *string
	BrandLensID       *int
	LensTypeID        *int
	LensesMaterialsID *int
	LensSeriesID      *int // -1 = clear
	Description       *string
	VendorID          *int // -1 = clear
	Price             *float64
	Cost              *float64
	VCodes            *[]int
	SpecialFeatures   *[]int
}

// ─── Methods ──────────────────────────────────────────────────────────────────

func (s *Service) GetLensBrandsVendors() ([]LensBrandVendorResult, error) {
	type row struct {
		IDBrandLens int
		BrandName   string
		IDVendor    *int
		VendorName  *string
	}
	var rows []row
	err := s.db.Table("brand_lens bl").
		Select("DISTINCT bl.id_brand_lens, bl.brand_name, v.id_vendor, v.vendor_name").
		Joins("LEFT JOIN lenses l ON l.brand_lens_id = bl.id_brand_lens").
		Joins("LEFT JOIN vendor v ON v.id_vendor = l.vendor_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]LensBrandVendorResult, len(rows))
	for i, r := range rows {
		result[i] = LensBrandVendorResult{r.IDBrandLens, r.BrandName, r.IDVendor, r.VendorName}
	}
	return result, nil
}

func (s *Service) GetLensTypes() ([]LensTypeResult, error) {
	var rows []lenses.LensType
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]LensTypeResult, len(rows))
	for i, r := range rows {
		result[i] = LensTypeResult{r.IDLensType, r.TypeName}
	}
	return result, nil
}

func (s *Service) GetLensMaterials() ([]LensMaterialResult, error) {
	var rows []lenses.LensesMaterial
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]LensMaterialResult, len(rows))
	for i, r := range rows {
		result[i] = LensMaterialResult{int(r.IDLensesMaterials), r.MaterialName}
	}
	return result, nil
}

func (s *Service) AddLensMaterial(name string) (int, error) {
	m := lenses.LensesMaterial{MaterialName: name}
	if err := s.db.Create(&m).Error; err != nil {
		return 0, err
	}
	return int(m.IDLensesMaterials), nil
}

func (s *Service) GetLensSpecialFeatures() ([]LensSpecialResult, error) {
	var rows []lenses.LensSpecialFeature
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]LensSpecialResult, len(rows))
	for i, r := range rows {
		result[i] = LensSpecialResult{r.IDLensSpecialFeatures, r.FeatureName}
	}
	return result, nil
}

func (s *Service) AddLensSpecialFeature(name string) (int, error) {
	f := lenses.LensSpecialFeature{FeatureName: name}
	if err := s.db.Create(&f).Error; err != nil {
		return 0, err
	}
	return f.IDLensSpecialFeatures, nil
}

func (s *Service) GetLensSeries() ([]LensSeriesResult, error) {
	var rows []lenses.LensSeries
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]LensSeriesResult, len(rows))
	for i, r := range rows {
		result[i] = LensSeriesResult{r.IDLensSeries, r.SeriesName}
	}
	return result, nil
}

func (s *Service) AddLensSeries(name string) (int, error) {
	sr := lenses.LensSeries{SeriesName: name}
	if err := s.db.Create(&sr).Error; err != nil {
		return 0, err
	}
	return sr.IDLensSeries, nil
}

func (s *Service) GetVCodes() ([]VCodeResult, error) {
	var rows []lenses.VCodesLens
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]VCodeResult, len(rows))
	for i, r := range rows {
		result[i] = VCodeResult{r.IDVCodesLens, r.Code}
	}
	return result, nil
}

func (s *Service) AddVCode(code string) (int, error) {
	vc := lenses.VCodesLens{Code: code}
	if err := s.db.Create(&vc).Error; err != nil {
		return 0, err
	}
	return vc.IDVCodesLens, nil
}

type LensFilters struct {
	BrandID          *int
	VendorID         *int
	TypeID           *int
	MaterialID       *int
	SpecialFeatureID *int
	SeriesID         *int
	Source           *string
	Search           *string
	Page             int
	PerPage          int
}

type LensListResponse struct {
	Items      []LensListItem `json:"items"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	TotalPages int            `json:"total_pages"`
}

func (s *Service) GetLensList(f LensFilters) (*LensListResponse, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 {
		f.PerPage = 25
	}

	q := s.db.Model(&lenses.Lenses{})

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
	if f.Source != nil {
		q = q.Where("source = ?", *f.Source)
	}
	if f.SpecialFeatureID != nil {
		q = q.Joins("JOIN lenses_feature_relation lfr ON lfr.lenses_id = lenses.id_lenses").
			Where("lfr.lens_special_features_id = ?", *f.SpecialFeatureID)
	}
	if f.Search != nil && *f.Search != "" {
		words := strings.Fields(strings.TrimSpace(*f.Search))
		for _, w := range words {
			// each word must match start of any word in description (case-insensitive)
			safe := regexp.QuoteMeta(w)
			q = q.Where("description ~* ?", `(^|\s)`+safe)
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (f.Page - 1) * f.PerPage

	var lens []lenses.Lenses
	if err := q.
		Preload("BrandLens").
		Preload("LensType").
		Preload("LensesMaterial").
		Preload("LensSeries").
		Offset(offset).Limit(f.PerPage).
		Find(&lens).Error; err != nil {
		return nil, err
	}

	result := make([]LensListItem, 0, len(lens))
	for _, l := range lens {
		// v-codes
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

		// special features
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

		item := LensListItem{
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
		result = append(result, item)
	}

	totalPages := int(total) / f.PerPage
	if int(total)%f.PerPage != 0 {
		totalPages++
	}

	return &LensListResponse{
		Items:      result,
		Total:      total,
		Page:       f.Page,
		PerPage:    f.PerPage,
		TotalPages: totalPages,
	}, nil
}

func (s *Service) GetLens(id int) (*LensDetail, error) {
	var l lenses.Lenses
	if err := s.db.
		Preload("BrandLens").
		Preload("LensType").
		Preload("LensesMaterial").
		Preload("LensSeries").
		Preload("Vendor").
		First(&l, id).Error; err != nil {
		return nil, err
	}

	// v-codes with IDs
	type vcRow struct {
		IDVCodesLens int
		Code         string
	}
	var vcRows []vcRow
	s.db.Table("v_codes_lens vc").
		Select("vc.id_v_codes_lens, vc.code").
		Joins("JOIN lenses_v_codes_relation lvr ON lvr.v_codes_lens_id = vc.id_v_codes_lens").
		Where("lvr.lenses_id = ?", l.IDLenses).
		Scan(&vcRows)
	vcList := make([]map[string]interface{}, len(vcRows))
	for i, v := range vcRows {
		vcList[i] = map[string]interface{}{"id_v_code": v.IDVCodesLens, "v_code": v.Code}
	}

	// special features (array of objects with id + name)
	type sfRow struct {
		ID          int    `gorm:"column:id_lens_special_features"`
		FeatureName string `gorm:"column:feature_name"`
	}
	var sfRows []sfRow
	s.db.Table("lens_special_features sf").
		Select("sf.id_lens_special_features, sf.feature_name").
		Joins("JOIN lenses_feature_relation lfr ON lfr.lens_special_features_id = sf.id_lens_special_features").
		Where("lfr.lenses_id = ?", l.IDLenses).
		Scan(&sfRows)
	sfList := make([]map[string]interface{}, len(sfRows))
	for i, sf := range sfRows {
		sfList[i] = map[string]interface{}{"id_lens_special_features": sf.ID, "feature_name": sf.FeatureName}
	}

	brand := map[string]interface{}{"id_brand_lens": nil, "brand_name": nil}
	if l.BrandLens != nil {
		brand["id_brand_lens"] = l.BrandLens.IDBrandLens
		brand["brand_name"] = l.BrandLens.BrandName
	}
	ltype := map[string]interface{}{"id_lens_type": nil, "type_name": nil}
	if l.LensType != nil {
		ltype["id_lens_type"] = l.LensType.IDLensType
		ltype["type_name"] = l.LensType.TypeName
	}
	mat := map[string]interface{}{"id_lenses_materials": nil, "material_name": nil}
	if l.LensesMaterial != nil {
		mat["id_lenses_materials"] = l.LensesMaterial.IDLensesMaterials
		mat["material_name"] = l.LensesMaterial.MaterialName
	}
	ser := map[string]interface{}{"id_lens_series": nil, "series_name": nil}
	if l.LensSeries != nil {
		ser["id_lens_series"] = l.LensSeries.IDLensSeries
		ser["series_name"] = l.LensSeries.SeriesName
	}
	vendor := map[string]interface{}{"id_vendor": nil, "vendor_name": nil}
	if l.Vendor != nil {
		vendor["id_vendor"] = l.Vendor.IDVendor
		vendor["vendor_name"] = l.Vendor.VendorName
	}

	price := "0.00"
	if l.Price != nil {
		price = fmt.Sprintf("%.2f", *l.Price)
	}
	cost := "0.00"
	if l.Cost != nil {
		cost = fmt.Sprintf("%.2f", *l.Cost)
	}

	return &LensDetail{
		IDLenses:        l.IDLenses,
		LensName:        l.LensName,
		Brand:           brand,
		Type:            ltype,
		Material:        mat,
		Series:          ser,
		Vendor:          vendor,
		Description:     l.Description,
		VCodes:          vcList,
		SpecialFeatures: sfList,
		Price:           price,
		Cost:            cost,
		MfrNumber:       l.MFRNumber,
		CanLookup:       l.CanLookup,
	}, nil
}

func (s *Service) AddLens(in AddLensInput) (int, error) {
	var brand vendormodel.BrandLens
	if err := s.db.First(&brand, in.BrandLensID).Error; err != nil {
		return 0, fmt.Errorf("brand_lens not found")
	}

	customSource := "custom"
	l := lenses.Lenses{
		LensName:          in.LensName,
		BrandLensID:       &in.BrandLensID,
		LensTypeID:        in.LensTypeID,
		LensesMaterialsID: in.LensesMaterialsID,
		LensSeriesID:      in.LensSeriesID,
		Description:       in.Description,
		VendorID:          in.VendorID,
		Price:             &in.Price,
		Cost:              &in.Cost,
		Source:            &customSource,
	}
	if err := s.db.Create(&l).Error; err != nil {
		return 0, err
	}

	for _, vcID := range in.VCodes {
		var vc lenses.VCodesLens
		if s.db.First(&vc, vcID).Error != nil {
			s.db.Delete(&lenses.Lenses{}, l.IDLenses)
			return 0, fmt.Errorf("v_code %d not found", vcID)
		}
		s.db.Create(&lenses.LensesVCodesRelation{LensesID: l.IDLenses, VCodesLensID: vcID})
	}

	for _, sfID := range in.SpecialFeatures {
		s.db.Create(&lenses.LensesFeatureRelation{LensesID: l.IDLenses, LensSpecialFeaturesID: sfID})
	}

	return l.IDLenses, nil
}

func (s *Service) UpdateLens(id int, in UpdateLensInput) error {
	var l lenses.Lenses
	if err := s.db.First(&l, id).Error; err != nil {
		return fmt.Errorf("lens not found")
	}

	updates := map[string]interface{}{}
	if in.LensName != nil {
		updates["lens_name"] = *in.LensName
	}
	if in.BrandLensID != nil {
		updates["brand_lens_id"] = *in.BrandLensID
	}
	if in.LensTypeID != nil {
		updates["lens_type_id"] = *in.LensTypeID
	}
	if in.LensesMaterialsID != nil {
		updates["lenses_materials_id"] = *in.LensesMaterialsID
	}
	if in.LensSeriesID != nil {
		if *in.LensSeriesID == -1 {
			updates["lens_series_id"] = nil
		} else {
			updates["lens_series_id"] = *in.LensSeriesID
		}
	}
	if in.Description != nil {
		updates["description"] = *in.Description
	}
	if in.VendorID != nil {
		if *in.VendorID == -1 {
			updates["vendor_id"] = nil
		} else {
			updates["vendor_id"] = *in.VendorID
		}
	}
	if in.Price != nil {
		updates["price"] = *in.Price
	}
	if in.Cost != nil {
		updates["cost"] = *in.Cost
	}

	if len(updates) > 0 {
		if err := s.db.Model(&l).Updates(updates).Error; err != nil {
			return err
		}
	}

	if in.VCodes != nil {
		s.db.Where("lenses_id = ?", id).Delete(&lenses.LensesVCodesRelation{})
		for _, vcID := range *in.VCodes {
			if vcID == 0 {
				continue
			}
			s.db.Create(&lenses.LensesVCodesRelation{LensesID: id, VCodesLensID: vcID})
		}
	}

	if in.SpecialFeatures != nil {
		s.db.Where("lenses_id = ?", id).Delete(&lenses.LensesFeatureRelation{})
		for _, sfID := range *in.SpecialFeatures {
			s.db.Create(&lenses.LensesFeatureRelation{LensesID: id, LensSpecialFeaturesID: sfID})
		}
	}

	return nil
}

func (s *Service) DeleteLens(id int) error {
	var l lenses.Lenses
	if err := s.db.First(&l, id).Error; err != nil {
		return fmt.Errorf("lens not found")
	}

	// check if in use
	var countSale, countService int64
	s.db.Table("invoice_item_sale").Where("item_type = 'Lens' AND item_id = ?", id).Count(&countSale)
	s.db.Table("invoice_services_item isi").
		Joins("JOIN orders_lens ol ON ol.id_orders_lens = isi.lens_order_id").
		Where("ol.lens_id = ?", id).Count(&countService)
	if countSale > 0 || countService > 0 {
		return fmt.Errorf("lens is used in invoices")
	}

	s.db.Where("lenses_id = ?", id).Delete(&lenses.LensesVCodesRelation{})
	return s.db.Delete(&l).Error
}
