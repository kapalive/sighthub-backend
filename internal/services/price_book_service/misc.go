package price_book_service

import (
	"fmt"

	"sighthub-backend/internal/models/misc"
)

// ─── Result types ─────────────────────────────────────────────────────────────

type MiscItemResult struct {
	ItemID      int64   `json:"item_id"`
	ItemNumber  string  `json:"item_number"`
	Description string  `json:"description"`
	PbKey       string  `json:"pb_key"`
	SaleKey     *string `json:"sale_key"`
	Cost        *string `json:"cost"`
	Price       *string `json:"price"`
	CanLookup   bool    `json:"can_lookup"`
	Active      bool    `json:"active"`
}

type MiscListFilters struct {
	PbKey      *string
	Q          *string
	Active     *bool
	LookupOnly bool
}

type AddMiscItemInput struct {
	ItemNumber  string
	Description string
	SaleKey     *string
	Cost        *string
	Price       *string
	CanLookup   *bool
	Active      *bool
}

type UpdateMiscItemInput struct {
	ItemNumber  *string
	Description *string
	Price       *string
	Cost        *string
	SaleKey     *string
	SetSaleKey  bool
	CanLookup   *bool
	Active      *bool
}

// ─── Methods ──────────────────────────────────────────────────────────────────

func (s *Service) GetMiscItems(f MiscListFilters) ([]MiscItemResult, error) {
	q := s.db.Model(&misc.MiscInvoiceItem{})

	if f.PbKey != nil {
		q = q.Where("pb_key = ?", *f.PbKey)
	}
	if f.Q != nil && *f.Q != "" {
		like := "%" + *f.Q + "%"
		q = q.Where("item_number ILIKE ? OR description ILIKE ?", like, like)
	}
	if f.Active != nil {
		q = q.Where("active = ?", *f.Active)
	}
	if f.LookupOnly {
		q = q.Where("can_lookup = true")
	}
	q = q.Order("item_number ASC, id_misc_item ASC")

	var rows []misc.MiscInvoiceItem
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]MiscItemResult, len(rows))
	for i, r := range rows {
		result[i] = MiscItemResult{
			ItemID:      r.IDMiscItem,
			ItemNumber:  r.ItemNumber,
			Description: r.Description,
			PbKey:       r.PbKey,
			SaleKey:     r.SaleKey,
			Cost:        r.Cost,
			Price:       r.Price,
			CanLookup:   r.CanLookup,
			Active:      r.Active,
		}
	}
	return result, nil
}

func (s *Service) GetMiscItem(id int) (*MiscItemResult, error) {
	var row misc.MiscInvoiceItem
	if err := s.db.First(&row, id).Error; err != nil {
		return nil, fmt.Errorf("misc item not found")
	}
	return &MiscItemResult{
		ItemID:      row.IDMiscItem,
		ItemNumber:  row.ItemNumber,
		Description: row.Description,
		PbKey:       row.PbKey,
		SaleKey:     row.SaleKey,
		Cost:        row.Cost,
		Price:       row.Price,
		CanLookup:   row.CanLookup,
		Active:      row.Active,
	}, nil
}

func (s *Service) AddMiscItem(in AddMiscItemInput) (*MiscItemResult, error) {
	row := misc.MiscInvoiceItem{
		ItemNumber:  in.ItemNumber,
		Description: in.Description,
		PbKey:       "misc",
		SaleKey:     in.SaleKey,
		Cost:        in.Cost,
		Price:       in.Price,
	}
	if in.CanLookup != nil {
		row.CanLookup = *in.CanLookup
	}
	if in.Active != nil {
		row.Active = *in.Active
	}
	if err := s.db.Create(&row).Error; err != nil {
		return nil, err
	}
	return &MiscItemResult{
		ItemID:      row.IDMiscItem,
		ItemNumber:  row.ItemNumber,
		Description: row.Description,
		PbKey:       row.PbKey,
		SaleKey:     row.SaleKey,
		Cost:        row.Cost,
		Price:       row.Price,
		CanLookup:   row.CanLookup,
		Active:      row.Active,
	}, nil
}

func (s *Service) UpdateMiscItem(id int, in UpdateMiscItemInput) (*MiscItemResult, error) {
	var row misc.MiscInvoiceItem
	if err := s.db.First(&row, id).Error; err != nil {
		return nil, fmt.Errorf("misc item not found")
	}

	updates := map[string]interface{}{}
	if in.ItemNumber != nil {
		updates["item_number"] = *in.ItemNumber
	}
	if in.Description != nil {
		updates["description"] = *in.Description
	}
	if in.Price != nil {
		updates["price"] = *in.Price
	}
	if in.Cost != nil {
		updates["cost"] = *in.Cost
	}
	if in.CanLookup != nil {
		updates["can_lookup"] = *in.CanLookup
	}
	if in.Active != nil {
		updates["active"] = *in.Active
	}

	// sale_key logic
	if in.SetSaleKey {
		if in.SaleKey == nil || *in.SaleKey == "" {
			updates["sale_key"] = row.PbKey
		} else {
			updates["sale_key"] = *in.SaleKey
		}
	} else if row.SaleKey == nil || *row.SaleKey == "" {
		updates["sale_key"] = row.PbKey
	}

	if len(updates) > 0 {
		if err := s.db.Model(&row).Updates(updates).Error; err != nil {
			return nil, err
		}
		// reload
		s.db.First(&row, id)
	}

	return &MiscItemResult{
		ItemID:      row.IDMiscItem,
		ItemNumber:  row.ItemNumber,
		Description: row.Description,
		PbKey:       row.PbKey,
		SaleKey:     row.SaleKey,
		Cost:        row.Cost,
		Price:       row.Price,
		CanLookup:   row.CanLookup,
		Active:      row.Active,
	}, nil
}

func (s *Service) DeleteMiscItem(id int) error {
	var row misc.MiscInvoiceItem
	if err := s.db.First(&row, id).Error; err != nil {
		return fmt.Errorf("misc item not found")
	}
	if err := s.db.Delete(&row).Error; err != nil {
		return fmt.Errorf("cannot delete: referenced by other records")
	}
	return nil
}
