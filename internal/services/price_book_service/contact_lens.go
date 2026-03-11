package price_book_service

import (
	"fmt"

	contactlens "sighthub-backend/internal/models/contact_lens"
	"sighthub-backend/internal/models/vendors"
)

// ─── Result types ─────────────────────────────────────────────────────────────

type CLVendorResult struct {
	VendorID   int    `json:"vendor_id"`
	VendorName string `json:"vendor_name"`
}

type CLBrandResult struct {
	IDBrandContactLens int    `json:"id_brand_contact_lens"`
	BrandName          string `json:"brand_name"`
}

type CLListItem struct {
	ItemID      int     `json:"item_id"`
	ItemName    string  `json:"item_name"`
	BrandName   string  `json:"brand_name"`
	Description *string `json:"description"`
	VCode       *string `json:"v_code"`
	Price       *string `json:"price"`
	Cost        *string `json:"cost"`
	PbKey       string  `json:"pb_key"`
}

type CLDetail struct {
	IDContactLensItem int     `json:"id_contact_lens_item"`
	NameContact       string  `json:"name_contact"`
	BrandName         string  `json:"brand_name"`
	BrandID           int     `json:"brand_id"`
	InvoiceDesc       *string `json:"invoice_desc"`
	VCode             *string `json:"v_code"`
	SellingPrice      *string `json:"selling_price"`
	Cost              *string `json:"cost"`
	VendorID          int     `json:"vendor_id"`
	CanLookup         *bool   `json:"can_lookup"`
}

type CLListFilters struct {
	VendorID  *int
	BrandID   *int
	CanLookup *bool
}

type AddCLInput struct {
	NameContact  string
	BrandID      int
	InvoiceDesc  string
	SellingPrice float64
	Cost         float64
	VendorID     int
	InsVCode     *string
	CanLookup    bool
}

type UpdateCLInput struct {
	NameContact  *string
	BrandID      *int
	InvoiceDesc  *string
	SellingPrice *float64
	Cost         *float64
	VendorID     *int
	InsVCode     *string // empty string = clear
	SetVCode     bool    // true if InsVCode field was present in request
	CanLookup    *bool
}

// ─── Methods ──────────────────────────────────────────────────────────────────

func (s *Service) GetCLVendors() ([]CLVendorResult, error) {
	type row struct {
		IDVendor   int
		VendorName string
	}
	var rows []row
	err := s.db.Table("vendor v").
		Select("DISTINCT v.id_vendor, v.vendor_name").
		Joins("JOIN contact_lens_item cli ON cli.vendor_id = v.id_vendor").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]CLVendorResult, len(rows))
	for i, r := range rows {
		result[i] = CLVendorResult{r.IDVendor, r.VendorName}
	}
	return result, nil
}

func (s *Service) GetCLBrands() ([]CLBrandResult, error) {
	var rows []vendors.BrandContactLens
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]CLBrandResult, len(rows))
	for i, r := range rows {
		result[i] = CLBrandResult{r.IDBrandContactLens, r.BrandName}
	}
	return result, nil
}

func (s *Service) GetCLList(f CLListFilters) ([]CLListItem, error) {
	q := s.db.Table("contact_lens_item cli").
		Select("cli.id_contact_lens_item, cli.name_contact, bcl.brand_name, cli.invoice_desc, cli.ins_v_code, cli.selling_price, cli.cost").
		Joins("JOIN brand_contact_lens bcl ON bcl.id_brand_contact_lens = cli.brand_contact_lens_id")

	if f.VendorID != nil {
		q = q.Where("cli.vendor_id = ?", *f.VendorID)
	}
	if f.BrandID != nil {
		q = q.Where("cli.brand_contact_lens_id = ?", *f.BrandID)
	}
	if f.CanLookup != nil {
		q = q.Where("cli.can_lookup = ?", *f.CanLookup)
	}

	type row struct {
		IDContactLensItem int
		NameContact       string
		BrandName         string
		InvoiceDesc       *string
		InsVCode          *string
		SellingPrice      *float64
		Cost              *float64
	}
	var rows []row
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]CLListItem, len(rows))
	for i, r := range rows {
		result[i] = CLListItem{
			ItemID:      r.IDContactLensItem,
			ItemName:    r.NameContact,
			BrandName:   r.BrandName,
			Description: r.InvoiceDesc,
			VCode:       r.InsVCode,
			Price:       strOrNil(r.SellingPrice),
			Cost:        strOrNil(r.Cost),
			PbKey:       "Contact Lens",
		}
	}
	return result, nil
}

func (s *Service) GetCL(id int) (*CLDetail, error) {
	type row struct {
		IDContactLensItem  int
		NameContact        string
		IDBrandContactLens int
		BrandName          string
		InvoiceDesc        *string
		InsVCode           *string
		SellingPrice       *float64
		Cost               *float64
		VendorID           int
		CanLookup          *bool
	}
	var r row
	err := s.db.Table("contact_lens_item cli").
		Select("cli.id_contact_lens_item, cli.name_contact, bcl.id_brand_contact_lens, bcl.brand_name, cli.invoice_desc, cli.ins_v_code, cli.selling_price, cli.cost, cli.vendor_id, cli.can_lookup").
		Joins("JOIN brand_contact_lens bcl ON bcl.id_brand_contact_lens = cli.brand_contact_lens_id").
		Where("cli.id_contact_lens_item = ?", id).
		Scan(&r).Error
	if err != nil {
		return nil, err
	}
	if r.IDContactLensItem == 0 {
		return nil, fmt.Errorf("contact lens not found")
	}
	return &CLDetail{
		IDContactLensItem: r.IDContactLensItem,
		NameContact:       r.NameContact,
		BrandName:         r.BrandName,
		BrandID:           r.IDBrandContactLens,
		InvoiceDesc:       r.InvoiceDesc,
		VCode:             r.InsVCode,
		SellingPrice:      strOrNil(r.SellingPrice),
		Cost:              strOrNil(r.Cost),
		VendorID:          r.VendorID,
		CanLookup:         r.CanLookup,
	}, nil
}

func (s *Service) AddCL(in AddCLInput) (int, error) {
	item := contactlens.ContactLensItem{
		NameContact:        in.NameContact,
		BrandContactLensID: &in.BrandID,
		InvoiceDesc:        &in.InvoiceDesc,
		InsVCode:           in.InsVCode,
		SellingPrice:       &in.SellingPrice,
		Cost:               &in.Cost,
		VendorID:           in.VendorID,
		CanLookup:          &in.CanLookup,
	}
	if err := s.db.Create(&item).Error; err != nil {
		return 0, err
	}
	return item.IDContactLensItem, nil
}

func (s *Service) UpdateCL(id int, in UpdateCLInput) error {
	var item contactlens.ContactLensItem
	if err := s.db.First(&item, id).Error; err != nil {
		return fmt.Errorf("contact lens not found")
	}

	updates := map[string]interface{}{}
	if in.NameContact != nil {
		updates["name_contact"] = *in.NameContact
	}
	if in.BrandID != nil {
		updates["brand_contact_lens_id"] = *in.BrandID
	}
	if in.InvoiceDesc != nil {
		updates["invoice_desc"] = *in.InvoiceDesc
	}
	if in.SellingPrice != nil {
		updates["selling_price"] = *in.SellingPrice
	}
	if in.Cost != nil {
		updates["cost"] = *in.Cost
	}
	if in.VendorID != nil {
		updates["vendor_id"] = *in.VendorID
	}
	if in.SetVCode {
		if in.InsVCode == nil || *in.InsVCode == "" {
			updates["ins_v_code"] = nil
		} else {
			updates["ins_v_code"] = *in.InsVCode
		}
	}
	if in.CanLookup != nil {
		updates["can_lookup"] = *in.CanLookup
	}

	if len(updates) > 0 {
		return s.db.Model(&item).Updates(updates).Error
	}
	return nil
}

func (s *Service) DeleteCL(id int) error {
	var item contactlens.ContactLensItem
	if err := s.db.First(&item, id).Error; err != nil {
		return fmt.Errorf("contact lens not found")
	}

	var countSale, countService int64
	s.db.Table("invoice_item_sale").Where("item_type = 'Contact Lens' AND item_id = ?", id).Count(&countSale)
	s.db.Table("invoice_services_item").Where("contact_lens_item_id = ?", id).Count(&countService)
	if countSale > 0 || countService > 0 {
		return fmt.Errorf("contact lens is used in invoices")
	}

	return s.db.Delete(&item).Error
}
