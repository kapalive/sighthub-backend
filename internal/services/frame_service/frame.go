package frame_service

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/frames"
	"sighthub-backend/internal/models/lenses"
	"sighthub-backend/internal/models/types"
	"sighthub-backend/internal/models/vendors"
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ─── Result / Input types ─────────────────────────────────────────────────────

type ModelSearchResult struct {
	IDModel      int64   `json:"id_model"`
	TitleVariant string  `json:"title_variant"`
	UPC          *string `json:"upc"`
	GTIN         *string `json:"gtin"`
	ProductID    int64   `json:"product_id"`
	TitleProduct string  `json:"title_product"`
	BrandID      *int64  `json:"brand_id"`
	BrandName    *string `json:"brand_name"`
	VendorID     *int64  `json:"vendor_id"`
	VendorName   string  `json:"vendor_name"`
}

type BrandResult struct {
	BrandID          int                    `json:"brand_id"`
	BrandName        *string                `json:"brand_name"`
	ShortName        *string                `json:"short_name"`
	Description      *string                `json:"description"`
	ReturnPolicy     *string                `json:"return_policy"`
	Note             *string                `json:"note"`
	PrintModelOnTag  bool                   `json:"print_model_on_tag"`
	PrintPriceOnTag  bool                   `json:"print_price_on_tag"`
	Discount         *int                   `json:"discount"`
	CanLookup        bool                   `json:"can_lookup"`
	TypeItemsOfBrand map[string]interface{} `json:"type_items_of_brand"`
}

type VendorBrandComboResult struct {
	VendorID   int    `json:"vendor_id"`
	VendorName string `json:"vendor_name"`
	BrandID    int    `json:"brand_id"`
	BrandName  string `json:"brand_name"`
}

type ProductResult struct {
	ProductID    int64  `json:"product_id"`
	TitleProduct string `json:"title_product"`
}

type ModelDetailResult struct {
	ProductID        int64   `json:"product_id"`
	TitleProduct     string  `json:"title_product"`
	IDModel          int64   `json:"id_model"`
	TitleVariant     string  `json:"title_variant"`
	LensColor        *string `json:"lens_color"`
	SizeLensWidth    *string `json:"size_lens_width"`
	SizeBridgeWidth  *string `json:"size_bridge_width"`
	SizeTempleLength *string `json:"size_temple_length"`
	Sunglass         *bool   `json:"sunglass"`
	Photo            *bool   `json:"photo"`
	Polor            *bool   `json:"polor"`
	UPC              *string `json:"upc"`
	MfgNumber        *string `json:"mfg_number"`
	EAN              *string `json:"ean"`
	MfrSerialNumber  *string `json:"mfr_serial_number"`
	Mirror           bool    `json:"mirror"`
	BacksideAR       bool    `json:"backside_ar"`
	LensMaterial     *string `json:"lens_material"`
	Accessories      *string `json:"accessories"`
}

type SearchFilters struct {
	BrandID      *int64
	BrandName    *string
	VendorID     *int64
	VendorName   *string
	ProductID    *int64
	TitleProduct *string
	IDModel      *int64
	UPC          *string
	GTIN         *string
}

type UpdateModelInput struct {
	TitleVariant     *string
	LensColor        *string
	SizeLensWidth    *string
	SizeBridgeWidth  *string
	SizeTempleLength *string
	Sunglass         *bool
	Photo            *bool
	Polor            *bool
	Mirror           *bool
	BacksideAR       *bool
	LensMaterial     *string
	UPC              *string
	EAN              *string
	MfgNumber        *string
	MfrSerialNumber  *string
	Accessories      *string
	MaterialsFrame   *string
	MaterialsTemple  *string
	Color            *string
	ColorTemplate    *string
	Shape            *string
}

type AddProductInput struct {
	VendorID     int64
	BrandID      int64
	TitleProduct string
	TypeProduct  string
}

type AddVariantInput struct {
	ProductID        int64
	TitleVariant     string
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
	EAN              *string
	MfgNumber        *string
	MfrSerialNumber  *string
	Accessories      *string
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
	BrandID          *int64
}

// ─── Methods ──────────────────────────────────────────────────────────────────

func (s *Service) SearchModels(f SearchFilters) ([]ModelSearchResult, error) {
	type row struct {
		IDModel      int64
		TitleVariant string
		UPC          *string
		GTIN         *string
		ProductID    int64
		TitleProduct string
		BrandID      *int64
		BrandName    *string
		VendorID     *int64
		VendorName   string
	}

	q := s.db.Table("model m").
		Select(`m.id_model, m.title_variant, m.upc, m.gtin,
			p.id_product AS product_id, p.title_product,
			b.id_brand AS brand_id, b.brand_name,
			v.id_vendor AS vendor_id, v.vendor_name`).
		Joins("JOIN product p ON p.id_product = m.product_id").
		Joins("JOIN brand b ON b.id_brand = p.brand_id").
		Joins("JOIN vendor v ON v.id_vendor = p.vendor_id")

	if f.BrandID != nil {
		q = q.Where("b.id_brand = ?", *f.BrandID)
	}
	if f.BrandName != nil {
		q = q.Where("b.brand_name ILIKE ?", "%"+*f.BrandName+"%")
	}
	if f.VendorID != nil {
		q = q.Where("v.id_vendor = ?", *f.VendorID)
	}
	if f.VendorName != nil {
		q = q.Where("v.vendor_name ILIKE ?", "%"+*f.VendorName+"%")
	}
	if f.ProductID != nil {
		q = q.Where("p.id_product = ?", *f.ProductID)
	}
	if f.TitleProduct != nil {
		q = q.Where("p.title_product ILIKE ?", "%"+*f.TitleProduct+"%")
	}
	if f.IDModel != nil {
		q = q.Where("m.id_model = ?", *f.IDModel)
	}
	if f.UPC != nil {
		q = q.Where("m.upc ILIKE ?", "%"+*f.UPC+"%")
	}
	if f.GTIN != nil {
		q = q.Where("m.gtin ILIKE ?", "%"+*f.GTIN+"%")
	}

	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]ModelSearchResult, len(rows))
	for i, r := range rows {
		result[i] = ModelSearchResult{
			IDModel:      r.IDModel,
			TitleVariant: r.TitleVariant,
			UPC:          r.UPC,
			GTIN:         r.GTIN,
			ProductID:    r.ProductID,
			TitleProduct: r.TitleProduct,
			BrandID:      r.BrandID,
			BrandName:    r.BrandName,
			VendorID:     r.VendorID,
			VendorName:   r.VendorName,
		}
	}
	return result, nil
}

func (s *Service) GetVendorBrands(vendorID int) ([]BrandResult, error) {
	var vendor vendors.Vendor
	if err := s.db.First(&vendor, vendorID).Error; err != nil {
		return []BrandResult{}, nil
	}

	var vbs []vendors.VendorBrand
	s.db.Where("id_vendor = ?", vendorID).
		Preload("Brand").
		Find(&vbs)

	// preload TypeItemsOfBrand for each brand
	result := make([]BrandResult, 0, len(vbs))
	for _, vb := range vbs {
		if vb.Brand == nil {
			continue
		}
		b := vb.Brand

		var tiob map[string]interface{}
		if b.TypeItemsOfBrandID != nil {
			var t vendors.TypeItemsOfBrand
			if err := s.db.First(&t, *b.TypeItemsOfBrandID).Error; err == nil {
				tiob = t.ToMap()
			}
		}

		result = append(result, BrandResult{
			BrandID:          b.IDBrand,
			BrandName:        b.BrandName,
			ShortName:        b.ShortName,
			Description:      b.Description,
			ReturnPolicy:     b.ReturnPolicy,
			Note:             b.Note,
			PrintModelOnTag:  b.PrintModelOnTag,
			PrintPriceOnTag:  b.PrintPriceOnTag,
			Discount:         b.Discount,
			CanLookup:        b.CanLookup,
			TypeItemsOfBrand: tiob,
		})
	}
	return result, nil
}

func (s *Service) GetVendorBrandCombinations() ([]VendorBrandComboResult, error) {
	type row struct {
		IDVendor   int
		VendorName string
		IDBrand    int
		BrandName  string
	}
	var rows []row
	err := s.db.Table("vendor v").
		Select("v.id_vendor, v.vendor_name, b.id_brand, b.brand_name").
		Joins("JOIN product p ON p.vendor_id = v.id_vendor").
		Joins("JOIN brand b ON b.id_brand = p.brand_id").
		Where("v.frames = true").
		Distinct("v.id_vendor, v.vendor_name, b.id_brand, b.brand_name").
		Order("b.brand_name").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]VendorBrandComboResult, len(rows))
	for i, r := range rows {
		result[i] = VendorBrandComboResult{
			VendorID:   r.IDVendor,
			VendorName: r.VendorName,
			BrandID:    r.IDBrand,
			BrandName:  r.BrandName,
		}
	}
	return result, nil
}

func (s *Service) GetProducts(vendorID, brandID *int64) ([]ProductResult, error) {
	q := s.db.Model(&frames.Product{})
	if vendorID != nil {
		q = q.Where("vendor_id = ?", *vendorID)
	}
	if brandID != nil {
		q = q.Where("brand_id = ?", *brandID)
	}

	var products []frames.Product
	if err := q.Find(&products).Error; err != nil {
		return nil, err
	}

	result := make([]ProductResult, len(products))
	for i, p := range products {
		result[i] = ProductResult{ProductID: p.IDProduct, TitleProduct: p.TitleProduct}
	}
	return result, nil
}

func (s *Service) GetFrameMaterials() ([]string, error) {
	type row struct{ MaterialsFrame *string }
	var rows []row
	if err := s.db.Model(&frames.Model{}).
		Distinct("materials_frame").
		Where("materials_frame IS NOT NULL").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]string, 0, len(rows))
	for _, r := range rows {
		if r.MaterialsFrame != nil {
			result = append(result, *r.MaterialsFrame)
		}
	}
	return result, nil
}

func (s *Service) GetModelsByProduct(productID int, materialFilter *string) ([]ModelDetailResult, error) {
	type row struct {
		IDModel          int64
		ProductID        int64
		TitleProduct     string
		TitleVariant     string
		LensColor        *string
		SizeLensWidth    *string
		SizeBridgeWidth  *string
		SizeTempleLength *string
		Sunglass         *bool
		Photo            *bool
		Polor            *bool
		Mirror           bool
		BacksideAR       bool
		LensMaterial     *string
		UPC              *string
		EAN              *string
		MfgNumber        *string
		MfrSerialNumber  *string
		Accessories      *string
	}

	q := s.db.Table("model m").
		Select(`m.id_model, m.product_id, p.title_product,
			m.title_variant, m.lens_color,
			m.size_lens_width, m.size_bridge_width, m.size_temple_length,
			m.sunglass, m.photo, m.polor, m.mirror, m.backside_ar,
			m.lens_material, m.upc, m.ean, m.mfg_number, m.mfr_serial_number,
			m.accessories`).
		Joins("JOIN product p ON p.id_product = m.product_id").
		Where("m.product_id = ?", productID).
		Order("m.title_variant ASC")

	if materialFilter != nil {
		q = q.Where("m.materials_frame = ?", *materialFilter)
	}

	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]ModelDetailResult, len(rows))
	for i, r := range rows {
		result[i] = ModelDetailResult{
			ProductID:        r.ProductID,
			TitleProduct:     r.TitleProduct,
			IDModel:          r.IDModel,
			TitleVariant:     r.TitleVariant,
			LensColor:        r.LensColor,
			SizeLensWidth:    r.SizeLensWidth,
			SizeBridgeWidth:  r.SizeBridgeWidth,
			SizeTempleLength: r.SizeTempleLength,
			Sunglass:         r.Sunglass,
			Photo:            r.Photo,
			Polor:            r.Polor,
			UPC:              r.UPC,
			MfgNumber:        r.MfgNumber,
			EAN:              r.EAN,
			MfrSerialNumber:  r.MfrSerialNumber,
			Mirror:           r.Mirror,
			BacksideAR:       r.BacksideAR,
			LensMaterial:     r.LensMaterial,
			Accessories:      r.Accessories,
		}
	}
	return result, nil
}

func (s *Service) UpdateModel(id int, in UpdateModelInput) (*frames.Model, error) {
	var m frames.Model
	if err := s.db.First(&m, id).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if in.TitleVariant != nil {
		updates["title_variant"] = *in.TitleVariant
	}
	if in.LensColor != nil {
		updates["lens_color"] = *in.LensColor
	}
	if in.SizeLensWidth != nil {
		updates["size_lens_width"] = *in.SizeLensWidth
	}
	if in.SizeBridgeWidth != nil {
		updates["size_bridge_width"] = *in.SizeBridgeWidth
	}
	if in.SizeTempleLength != nil {
		updates["size_temple_length"] = *in.SizeTempleLength
	}
	if in.Sunglass != nil {
		updates["sunglass"] = *in.Sunglass
	}
	if in.Photo != nil {
		updates["photo"] = *in.Photo
	}
	if in.Polor != nil {
		updates["polor"] = *in.Polor
	}
	if in.Mirror != nil {
		updates["mirror"] = *in.Mirror
	}
	if in.BacksideAR != nil {
		updates["backside_ar"] = *in.BacksideAR
	}
	if in.LensMaterial != nil {
		lm := types.LensMaterial(*in.LensMaterial)
		updates["lens_material"] = lm
	}
	if in.UPC != nil {
		updates["upc"] = *in.UPC
	}
	if in.EAN != nil {
		updates["ean"] = *in.EAN
	}
	if in.MfgNumber != nil {
		updates["mfg_number"] = *in.MfgNumber
	}
	if in.MfrSerialNumber != nil {
		updates["mfr_serial_number"] = *in.MfrSerialNumber
	}
	if in.Accessories != nil {
		updates["accessories"] = *in.Accessories
	}
	if in.MaterialsFrame != nil {
		updates["materials_frame"] = *in.MaterialsFrame
	}
	if in.MaterialsTemple != nil {
		updates["materials_temple"] = *in.MaterialsTemple
	}
	if in.Color != nil {
		updates["color"] = *in.Color
	}
	if in.ColorTemplate != nil {
		updates["color_template"] = *in.ColorTemplate
	}
	if in.Shape != nil {
		updates["shape"] = *in.Shape
	}

	if len(updates) > 0 {
		if err := s.db.Model(&m).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	s.db.First(&m, id)
	return &m, nil
}

var allowedTypeProducts = map[string]struct{}{
	"eyeglasses": {},
	"sunglasses": {},
}

func (s *Service) AddProduct(in AddProductInput) (*frames.Product, error) {
	tp := strings.ToLower(strings.TrimSpace(in.TypeProduct))
	if _, ok := allowedTypeProducts[tp]; !ok {
		return nil, errors.New("type_product must be 'eyeglasses' or 'sunglasses'")
	}

	var vendor vendors.Vendor
	if err := s.db.First(&vendor, in.VendorID).Error; err != nil {
		return nil, fmt.Errorf("vendor not found")
	}

	var brand vendors.Brand
	if err := s.db.First(&brand, in.BrandID).Error; err != nil {
		return nil, fmt.Errorf("brand not found")
	}

	var existing frames.Product
	err := s.db.Where("brand_id = ? AND LOWER(title_product) = LOWER(?)", in.BrandID, in.TitleProduct).
		First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("product with this title already exists for this brand")
	}

	product := frames.Product{
		TitleProduct: in.TitleProduct,
		BrandID:      &in.BrandID,
		VendorID:     &in.VendorID,
		TypeProduct:  tp,
	}
	if err := s.db.Create(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (s *Service) AddVariant(in AddVariantInput) (*frames.Model, error) {
	var product frames.Product
	if err := s.db.First(&product, in.ProductID).Error; err != nil {
		return nil, fmt.Errorf("product not found")
	}

	var existing frames.Model
	err := s.db.Where("product_id = ? AND LOWER(title_variant) = LOWER(?)", in.ProductID, in.TitleVariant).
		First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("variant with this title already exists for this product")
	}

	model := frames.Model{
		ProductID:        in.ProductID,
		TitleVariant:     in.TitleVariant,
		LensColor:        in.LensColor,
		SizeLensWidth:    in.SizeLensWidth,
		SizeBridgeWidth:  in.SizeBridgeWidth,
		SizeTempleLength: in.SizeTempleLength,
		Sunglass:         in.Sunglass,
		Photo:            in.Photo,
		Polor:            in.Polor,
		UPC:              in.UPC,
		EAN:              in.EAN,
		MfgNumber:        in.MfgNumber,
		MfrSerialNumber:  in.MfrSerialNumber,
		Accessories:      in.Accessories,
	}
	if in.LensMaterial != nil {
		lm := types.LensMaterial(*in.LensMaterial)
		model.LensMaterial = &lm
	}
	if in.Mirror != nil {
		model.Mirror = *in.Mirror
	}
	if in.BacksideAR != nil {
		model.BacksideAR = *in.BacksideAR
	}

	if err := s.db.Create(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (s *Service) CreateCustomGlasses(id int, in CustomGlassesInput) (*frames.Model, error) {
	var existing frames.Model
	if err := s.db.First(&existing, id).Error; err != nil {
		return nil, fmt.Errorf("model not found")
	}

	// handle brand change: find or create product with new brand_id
	targetProductID := existing.ProductID
	if in.BrandID != nil {
		var oldProduct frames.Product
		if err := s.db.First(&oldProduct, existing.ProductID).Error; err != nil {
			return nil, fmt.Errorf("source product not found")
		}
		if oldProduct.BrandID == nil || *in.BrandID != *oldProduct.BrandID {
			var existingProd frames.Product
			err := s.db.Where("title_product = ? AND brand_id = ? AND vendor_id = ? AND type_product = ?",
				oldProduct.TitleProduct, *in.BrandID, oldProduct.VendorID, oldProduct.TypeProduct).
				First(&existingProd).Error
			if err == nil {
				targetProductID = existingProd.IDProduct
			} else {
				np := frames.Product{
					TitleProduct: oldProduct.TitleProduct,
					BrandID:      in.BrandID,
					VendorID:     oldProduct.VendorID,
					TypeProduct:  oldProduct.TypeProduct,
				}
				if err := s.db.Create(&np).Error; err != nil {
					return nil, err
				}
				targetProductID = np.IDProduct
			}
		}
	}

	newModel := frames.Model{
		ProductID:        targetProductID,
		TitleVariant:     existing.TitleVariant,
		LensColor:        existing.LensColor,
		LensMaterial:     existing.LensMaterial,
		SizeLensWidth:    existing.SizeLensWidth,
		SizeBridgeWidth:  existing.SizeBridgeWidth,
		SizeTempleLength: existing.SizeTempleLength,
		Sunglass:         existing.Sunglass,
		Photo:            existing.Photo,
		Polor:            existing.Polor,
		Mirror:           existing.Mirror,
		BacksideAR:       existing.BacksideAR,
		UPC:              existing.UPC,
		Accessories:      existing.Accessories,
	}

	// apply overrides
	if in.TitleVariant != nil {
		newModel.TitleVariant = *in.TitleVariant
	}
	if in.LensColor != nil {
		newModel.LensColor = in.LensColor
	}
	if in.LensMaterial != nil {
		lm := types.LensMaterial(*in.LensMaterial)
		newModel.LensMaterial = &lm
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

	// build " - CUSTOM[accessories]" suffix
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

	if err := s.db.Create(&newModel).Error; err != nil {
		return nil, err
	}
	return &newModel, nil
}

func (s *Service) GetFrameTypeMaterials() ([]map[string]interface{}, error) {
	var items []frames.FrameTypeMaterial
	if err := s.db.Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = it.ToMap()
	}
	return out, nil
}

func (s *Service) GetFrameShapes() ([]map[string]interface{}, error) {
	var items []frames.FrameShape
	if err := s.db.Order("title_frame_shape").Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = it.ToMap()
	}
	return out, nil
}

func (s *Service) GetLensMaterials() ([]map[string]interface{}, error) {
	var items []lenses.LensesMaterial
	if err := s.db.Order("material_name").Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = map[string]interface{}{
			"id_lenses_materials": it.IDLensesMaterials,
			"material_name":       it.MaterialName,
			"index":               it.Index,
			"description":         it.Description,
		}
	}
	return out, nil
}
