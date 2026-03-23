package price_book_service

import (
	"fmt"

	"sighthub-backend/internal/models/lenses"
)

// ─── Result types ─────────────────────────────────────────────────────────────

type TreatmentVendorResult struct {
	VendorID   int    `json:"vendor_id"`
	VendorName string `json:"vendor_name"`
}

type TreatmentListItem struct {
	ItemID      int64   `json:"item_id"`
	ItemName    string  `json:"item_name"`
	Description *string `json:"description"`
	Price       *string `json:"price"`
	VCode       *string `json:"v_code"`
	PbKey       string  `json:"pb_key"`
}

type TreatmentDetail struct {
	IDLensTreatments int64                    `json:"id_lens_treatments"`
	ItemNbr          string                   `json:"item_nbr"`
	Vendor           *string                  `json:"vendor"`
	VendorID         *int                     `json:"vendor_id"`
	Description      *string                  `json:"description"`
	Price            *string                  `json:"price"`
	Cost             *string                  `json:"cost"`
	CanLookup        bool                     `json:"can_lookup"`
	SpecialFeatures  []map[string]interface{} `json:"special_features"`
	VCodesLensID     *int                     `json:"v_codes_lens_id"`
	VCode            *string                  `json:"v_code"`
	Source           *string                  `json:"source"`
}

type CreateTreatmentInput struct {
	ItemNbr         string
	VendorID        int
	Price           float64
	Description     *string
	Cost            *float64
	VCodesLensID    *int
	CanLookup       bool
	SpecialFeatures []int
}

type UpdateTreatmentInput struct {
	ItemNbr         *string
	Description     *string
	VendorID        *int
	VCodesLensID    *int // -1 = clear
	Price           *float64
	Cost            *float64
	ClearCost       bool
	CanLookup       *bool
	SpecialFeatures *[]int
}

// ─── Methods ──────────────────────────────────────────────────────────────────

func (s *Service) GetTreatmentVendors() ([]TreatmentVendorResult, error) {
	type row struct {
		IDVendor   int
		VendorName string
	}
	var rows []row
	err := s.db.Table("vendor v").
		Select("DISTINCT v.id_vendor, v.vendor_name").
		Joins("JOIN lens_treatments lt ON lt.vendor_id = v.id_vendor").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]TreatmentVendorResult, len(rows))
	for i, r := range rows {
		result[i] = TreatmentVendorResult{r.IDVendor, r.VendorName}
	}
	return result, nil
}

func (s *Service) GetTreatments(vendorID *int) ([]TreatmentListItem, error) {
	q := s.db.Table("lens_treatments lt").
		Select("lt.id_lens_treatments, lt.item_nbr, lt.description, lt.price, vc.code AS v_code").
		Joins("LEFT JOIN v_codes_lens vc ON vc.id_v_codes_lens = lt.v_codes_lens_id")

	if vendorID != nil {
		q = q.Where("lt.vendor_id = ?", *vendorID)
	}

	type row struct {
		IDLensTreatments int64
		ItemNbr          string
		Description      *string
		Price            *float64
		VCode            *string
	}
	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]TreatmentListItem, len(rows))
	for i, r := range rows {
		// price in list = price * 2 (for pair of lenses)
		var priceStr *string
		if r.Price != nil {
			p := fmt.Sprintf("%.2f", *r.Price*2)
			priceStr = &p
		}
		result[i] = TreatmentListItem{
			ItemID:      r.IDLensTreatments,
			ItemName:    r.ItemNbr,
			Description: r.Description,
			Price:       priceStr,
			VCode:       r.VCode,
			PbKey:       "Treatment",
		}
	}
	return result, nil
}

func (s *Service) GetTreatment(id int) (*TreatmentDetail, error) {
	var t lenses.LensTreatments
	if err := s.db.
		Preload("VCodesLens").
		First(&t, id).Error; err != nil {
		return nil, fmt.Errorf("lens treatment not found")
	}

	var vendorName *string
	var vendorID *int
	type vendorRow struct {
		IDVendor   int
		VendorName string
	}
	var vr vendorRow
	s.db.Table("vendor").Select("id_vendor, vendor_name").Where("id_vendor = ?", t.VendorID).Scan(&vr)
	if vr.IDVendor != 0 {
		vendorID = &vr.IDVendor
		vendorName = &vr.VendorName
	}

	var priceStr, costStr *string
	if t.Price != nil {
		p := fmt.Sprintf("%.2f", *t.Price)
		priceStr = &p
	}
	if t.Cost != nil {
		c := fmt.Sprintf("%.2f", *t.Cost)
		costStr = &c
	}

	var vCode *string
	if t.VCodesLens != nil {
		vCode = &t.VCodesLens.Code
	}

	// special features
	type sfRow struct {
		ID          int    `gorm:"column:id_lens_special_features"`
		FeatureName string `gorm:"column:feature_name"`
	}
	var sfRows []sfRow
	s.db.Table("lens_special_features sf").
		Select("sf.id_lens_special_features, sf.feature_name").
		Joins("JOIN treatments_feature_relation tfr ON tfr.lens_special_features_id = sf.id_lens_special_features").
		Where("tfr.lens_treatments_id = ?", t.IDLensTreatments).
		Scan(&sfRows)
	sfList := make([]map[string]interface{}, len(sfRows))
	for i, sf := range sfRows {
		sfList[i] = map[string]interface{}{"id_lens_special_features": sf.ID, "feature_name": sf.FeatureName}
	}

	return &TreatmentDetail{
		IDLensTreatments: t.IDLensTreatments,
		ItemNbr:          t.ItemNbr,
		Vendor:           vendorName,
		VendorID:         vendorID,
		Description:      t.Description,
		Price:            priceStr,
		Cost:             costStr,
		CanLookup:        t.CanLookup,
		SpecialFeatures:  sfList,
		VCodesLensID:     t.VCodesLensID,
		VCode:            vCode,
		Source:           t.Source,
	}, nil
}

func (s *Service) CreateTreatment(in CreateTreatmentInput) (int64, error) {
	t := lenses.LensTreatments{
		ItemNbr:      in.ItemNbr,
		Description:  in.Description,
		Price:        &in.Price,
		Cost:         in.Cost,
		VendorID:     in.VendorID,
		VCodesLensID: in.VCodesLensID,
		CanLookup:    in.CanLookup,
	}
	if err := s.db.Create(&t).Error; err != nil {
		return 0, err
	}
	for _, sfID := range in.SpecialFeatures {
		s.db.Exec("INSERT INTO treatments_feature_relation (lens_treatments_id, lens_special_features_id) VALUES (?, ?) ON CONFLICT DO NOTHING", t.IDLensTreatments, sfID)
	}
	return t.IDLensTreatments, nil
}

func (s *Service) UpdateTreatment(id int, in UpdateTreatmentInput) error {
	var t lenses.LensTreatments
	if err := s.db.First(&t, id).Error; err != nil {
		return fmt.Errorf("lens treatment not found")
	}

	updates := map[string]interface{}{}
	if in.ItemNbr != nil {
		updates["item_nbr"] = *in.ItemNbr
	}
	if in.Description != nil {
		updates["description"] = *in.Description
	}
	if in.VendorID != nil {
		updates["vendor_id"] = *in.VendorID
	}
	if in.VCodesLensID != nil {
		if *in.VCodesLensID == -1 {
			updates["v_codes_lens_id"] = nil
		} else {
			updates["v_codes_lens_id"] = *in.VCodesLensID
		}
	}
	if in.Price != nil {
		updates["price"] = *in.Price
	}
	if in.ClearCost {
		updates["cost"] = nil
	} else if in.Cost != nil {
		updates["cost"] = *in.Cost
	}
	if in.CanLookup != nil {
		updates["can_lookup"] = *in.CanLookup
	}
	if len(updates) > 0 {
		if err := s.db.Model(&t).Updates(updates).Error; err != nil {
			return err
		}
	}
	if in.SpecialFeatures != nil {
		s.db.Exec("DELETE FROM treatments_feature_relation WHERE lens_treatments_id = ?", t.IDLensTreatments)
		for _, sfID := range *in.SpecialFeatures {
			s.db.Exec("INSERT INTO treatments_feature_relation (lens_treatments_id, lens_special_features_id) VALUES (?, ?) ON CONFLICT DO NOTHING", t.IDLensTreatments, sfID)
		}
	}
	return nil
}

func (s *Service) DeleteTreatment(id int) error {
	var t lenses.LensTreatments
	if err := s.db.First(&t, id).Error; err != nil {
		return fmt.Errorf("lens treatment not found")
	}

	var countSale int64
	s.db.Table("invoice_item_sale").Where("item_type = 'Treatment' AND item_id = ?", id).Count(&countSale)
	if countSale > 0 {
		return fmt.Errorf("lens treatment is used in invoices")
	}

	return s.db.Delete(&t).Error
}
