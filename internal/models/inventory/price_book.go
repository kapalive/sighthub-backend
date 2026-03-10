// internal/models/inventory/price_book.go
package inventory

// PriceBook ⇄ table: price_book
// Хранит ценообразование для единицы инвентаря (1-к-1 с Inventory).
type PriceBook struct {
	IDPriceBook       int64    `gorm:"column:id_price_book;primaryKey;autoIncrement"                    json:"id_price_book"`
	InventoryID       int64    `gorm:"column:inventory_id;not null;uniqueIndex"                         json:"inventory_id"`
	ItemListCost      *float64 `gorm:"column:item_list_cost;type:numeric(10,2)"                         json:"item_list_cost,omitempty"`
	ItemDiscount      *float64 `gorm:"column:item_discount;type:numeric(10,2)"                          json:"item_discount,omitempty"`
	ItemNet           *float64 `gorm:"column:item_net;type:numeric(10,2)"                               json:"item_net,omitempty"`
	PbListCost        *float64 `gorm:"column:pb_list_cost;type:numeric(10,2)"                           json:"pb_list_cost,omitempty"`
	PbDiscount        *float64 `gorm:"column:pb_discount;type:numeric(10,2)"                            json:"pb_discount,omitempty"`
	PbCost            *float64 `gorm:"column:pb_cost;type:numeric(10,2)"                                json:"pb_cost,omitempty"`
	PbSellingPrice    *float64 `gorm:"column:pb_selling_price;type:numeric(10,2)"                       json:"pb_selling_price,omitempty"`
	PbStoreTierPrice  *float64 `gorm:"column:pb_store_tier_price;type:numeric(10,2)"                    json:"pb_store_tier_price,omitempty"`
	Note              *string  `gorm:"column:note;type:text"                                            json:"note,omitempty"`
	LensCost          *float64 `gorm:"column:lens_cost;type:numeric(10,2)"                              json:"lens_cost,omitempty"`
	AccessoriesCost   *float64 `gorm:"column:accessories_cost;type:numeric(10,2)"                       json:"accessories_cost,omitempty"`
}

func (PriceBook) TableName() string { return "price_book" }

func (p *PriceBook) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_price_book":       p.IDPriceBook,
		"inventory_id":        p.InventoryID,
		"item_list_cost":      p.ItemListCost,
		"item_discount":       p.ItemDiscount,
		"item_net":            p.ItemNet,
		"pb_list_cost":        p.PbListCost,
		"pb_discount":         p.PbDiscount,
		"pb_cost":             p.PbCost,
		"pb_selling_price":    p.PbSellingPrice,
		"pb_store_tier_price": p.PbStoreTierPrice,
		"note":                p.Note,
		"lens_cost":           p.LensCost,
		"accessories_cost":    p.AccessoriesCost,
	}
}
