// internal/repository/inventory_repo/price_book.go
package inventory_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/inventory"
)

type PriceBookRepo struct{ DB *gorm.DB }

func NewPriceBookRepo(db *gorm.DB) *PriceBookRepo { return &PriceBookRepo{DB: db} }

// GetByInventoryID возвращает price_book для единицы инвентаря.
func (r *PriceBookRepo) GetByInventoryID(inventoryID int64) (*inventory.PriceBook, error) {
	var row inventory.PriceBook
	err := r.DB.Where("inventory_id = ?", inventoryID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// UpdatePriceBookInput — изменяемые ценовые поля.
type UpdatePriceBookInput struct {
	ItemListCost     *float64
	ItemDiscount     *float64
	ItemNet          *float64
	PbListCost       *float64
	PbDiscount       *float64
	PbCost           *float64
	PbSellingPrice   *float64
	PbStoreTierPrice *float64
	LensCost         *float64
	AccessoriesCost  *float64
	Note             *string
}

// Update обновляет ценовые поля price_book для inventoryID.
func (r *PriceBookRepo) Update(inventoryID int64, inp UpdatePriceBookInput) error {
	updates := map[string]interface{}{}
	if inp.ItemListCost != nil     { updates["item_list_cost"]      = *inp.ItemListCost }
	if inp.ItemDiscount != nil     { updates["item_discount"]       = *inp.ItemDiscount }
	if inp.ItemNet != nil          { updates["item_net"]            = *inp.ItemNet }
	if inp.PbListCost != nil       { updates["pb_list_cost"]        = *inp.PbListCost }
	if inp.PbDiscount != nil       { updates["pb_discount"]         = *inp.PbDiscount }
	if inp.PbCost != nil           { updates["pb_cost"]             = *inp.PbCost }
	if inp.PbSellingPrice != nil   { updates["pb_selling_price"]    = *inp.PbSellingPrice }
	if inp.PbStoreTierPrice != nil { updates["pb_store_tier_price"] = *inp.PbStoreTierPrice }
	if inp.LensCost != nil         { updates["lens_cost"]           = *inp.LensCost }
	if inp.AccessoriesCost != nil  { updates["accessories_cost"]    = *inp.AccessoriesCost }
	if inp.Note != nil             { updates["note"]                = *inp.Note }
	if len(updates) == 0 {
		return nil
	}
	return r.DB.Model(&inventory.PriceBook{}).
		Where("inventory_id = ?", inventoryID).
		Updates(updates).Error
}

// Upsert создаёт или обновляет price_book для inventoryID.
func (r *PriceBookRepo) Upsert(pb *inventory.PriceBook) error {
	existing, err := r.GetByInventoryID(pb.InventoryID)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.DB.Create(pb).Error
	}
	pb.IDPriceBook = existing.IDPriceBook
	return r.DB.Save(pb).Error
}

// CalcSellingPrice вычисляет selling price по формуле:
// net = list_cost * (1 - discount/100), selling = net * (1 + margin/100).
func (r *PriceBookRepo) CalcSellingPrice(listCost, discountPct, marginPct float64) float64 {
	net := listCost * (1 - discountPct/100)
	return net * (1 + marginPct/100)
}

func (r *PriceBookRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
