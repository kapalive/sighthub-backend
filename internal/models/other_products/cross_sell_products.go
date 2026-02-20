package other_products

import (
	"fmt"
	"sighthub-backend/internal/models/interfaces" // Импортируем интерфейсы
)

type CrossSellProduct struct {
	IDCrossSellProducts int64  `gorm:"column:id_cross_sell_products;primaryKey;autoIncrement" json:"id_cross_sell_products"`
	Title               string `gorm:"column:title;type:varchar(255);not null" json:"title"`
	BrandID             *int64 `gorm:"column:brand_id" json:"brand_id,omitempty"`
	VendorID            *int64 `gorm:"column:vendor_id" json:"vendor_id,omitempty"`

	// Relationships через интерфейсы
	Brand  interfaces.BrandInterface  `gorm:"-" json:"brand,omitempty"`  // Используем интерфейс
	Vendor interfaces.VendorInterface `gorm:"-" json:"vendor,omitempty"` // Используем интерфейс
}

func (CrossSellProduct) TableName() string { return "cross_sell_products" }

func (c *CrossSellProduct) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_cross_sell_products": c.IDCrossSellProducts,
		"title":                  c.Title,
	}

	// Включаем Brand и Vendor в map, если они не nil
	if c.Brand != nil {
		m["brand"] = c.Brand.ToMap()
	} else {
		m["brand"] = nil
	}
	if c.Vendor != nil {
		m["vendor"] = c.Vendor.ToMap()
	} else {
		m["vendor"] = nil
	}

	return m
}

func (c *CrossSellProduct) String() string {
	return fmt.Sprintf("<CrossSellProduct %s>", c.Title)
}
