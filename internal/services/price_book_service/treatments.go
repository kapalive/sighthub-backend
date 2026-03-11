package price_book_service

import (
	"fmt"
	"time"

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
	IDLensTreatments int64   `json:"id_lens_treatments"`
	ItemNbr          string  `json:"item_nbr"`
	Vendor           *string `json:"vendor"`
	VendorID         *int    `json:"vendor_id"`
	Description      *string `json:"description"`
	Price            *string `json:"price"`
	Cost             *string `json:"cost"`
	CanLookup        bool    `json:"can_lookup"`
	SRCoat           bool    `json:"sr_coat"`
	UV               bool    `json:"uv"`
	AR               bool    `json:"ar"`
	Tint             bool    `json:"tint"`
	Photo            bool    `json:"photo"`
	Polar            bool    `json:"polar"`
	Drill            bool    `json:"drill"`
	HighIndex        bool    `json:"high_index"`
	VCodesLensID     *int    `json:"v_codes_lens_id"`
	VCode            *string `json:"v_code"`
	CreatedAt        string  `json:"created_at"`
	ModifiedAt       string  `json:"modified_at"`
}

type CreateTreatmentInput struct {
	ItemNbr      string
	VendorID     int
	Price        float64
	Description  *string
	Cost         *float64
	VCodesLensID *int
	CanLookup    bool
	SRCoat       bool
	UV           bool
	AR           bool
	Tint         bool
	Photo        bool
	Polar        bool
	Drill        bool
	HighIndex    bool
}

type UpdateTreatmentInput struct {
	ItemNbr      *string
	Description  *string
	VendorID     *int
	VCodesLensID *int  // -1 = clear
	Price        *float64
	Cost         *float64 // nil = don't change; use ClearCost = true to set NULL
	ClearCost    bool
	CanLookup    *bool
	SRCoat       *bool
	UV           *bool
	AR           *bool
	Tint         *bool
	Photo        *bool
	Polar        *bool
	Drill        *bool
	HighIndex    *bool
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
		Preload("Vendor").
		Preload("VCodesLens").
		First(&t, id).Error; err != nil {
		return nil, fmt.Errorf("lens treatment not found")
	}

	var vendorName *string
	var vendorID *int
	if t.Vendor != nil {
		type vendorInfo interface {
			GetVendorName() string
			GetVendorID() int
		}
		// Access via raw query to avoid interface complexities
	}
	// simpler: query vendor separately
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

	return &TreatmentDetail{
		IDLensTreatments: t.IDLensTreatments,
		ItemNbr:          t.ItemNbr,
		Vendor:           vendorName,
		VendorID:         vendorID,
		Description:      t.Description,
		Price:            priceStr,
		Cost:             costStr,
		CanLookup:        t.CanLookup,
		SRCoat:           t.SRCoat,
		UV:               t.UV,
		AR:               t.AR,
		Tint:             t.Tint,
		Photo:            t.Photo,
		Polar:            t.Polar,
		Drill:            t.Drill,
		HighIndex:        t.HighIndex,
		VCodesLensID:     t.VCodesLensID,
		VCode:            vCode,
		CreatedAt:        t.CreatedAt.Format(time.RFC3339),
		ModifiedAt:       t.ModifiedAt.Format(time.RFC3339),
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
		SRCoat:       in.SRCoat,
		UV:           in.UV,
		AR:           in.AR,
		Tint:         in.Tint,
		Photo:        in.Photo,
		Polar:        in.Polar,
		Drill:        in.Drill,
		HighIndex:    in.HighIndex,
	}
	if err := s.db.Create(&t).Error; err != nil {
		return 0, err
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
	if in.SRCoat != nil {
		updates["sr_coat"] = *in.SRCoat
	}
	if in.UV != nil {
		updates["uv"] = *in.UV
	}
	if in.AR != nil {
		updates["ar"] = *in.AR
	}
	if in.Tint != nil {
		updates["tint"] = *in.Tint
	}
	if in.Photo != nil {
		updates["photo"] = *in.Photo
	}
	if in.Polar != nil {
		updates["polar"] = *in.Polar
	}
	if in.Drill != nil {
		updates["drill"] = *in.Drill
	}
	if in.HighIndex != nil {
		updates["high_index"] = *in.HighIndex
	}

	if len(updates) > 0 {
		return s.db.Model(&t).Updates(updates).Error
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
