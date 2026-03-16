package vendors

import (
	"encoding/json"
	"time"
)

type PricingRule struct {
	IDPricingRule    int              `gorm:"column:id_pricing_rule;primaryKey" json:"id_pricing_rule"`
	BrandType        string           `gorm:"column:brand_type;type:varchar(20);not null" json:"brand_type"`
	BrandID          int              `gorm:"column:brand_id;not null" json:"brand_id"`
	MinPrice         *string          `gorm:"column:min_price;type:numeric(10,2)" json:"min_price"`
	MaxPrice         *string          `gorm:"column:max_price;type:numeric(10,2)" json:"max_price"`
	Multiplier       *string          `gorm:"column:multiplier;type:numeric(5,2)" json:"multiplier"`
	LowerMultiplier  *string          `gorm:"column:lower_multiplier;type:numeric(5,2)" json:"lower_multiplier"`
	SellingPrice     *string          `gorm:"column:selling_price;type:numeric(10,2)" json:"selling_price"`
	MinSellingPrice  *string          `gorm:"column:min_selling_price;type:numeric(10,2)" json:"min_selling_price"`
	RoundingTargets  json.RawMessage  `gorm:"column:rounding_targets;type:json" json:"rounding_targets"`
	CreatedAt        *time.Time       `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        *time.Time       `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (PricingRule) TableName() string { return "pricing_rules" }

func (p *PricingRule) IsBase() bool {
	return p.SellingPrice != nil
}

func (p *PricingRule) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"id_pricing_rule":  p.IDPricingRule,
		"brand_type":       p.BrandType,
		"brand_id":         p.BrandID,
		"is_base":          p.IsBase(),
		"min_price":        p.MinPrice,
		"max_price":        p.MaxPrice,
		"multiplier":       p.Multiplier,
		"lower_multiplier": p.LowerMultiplier,
		"selling_price":    p.SellingPrice,
		"min_selling_price": p.MinSellingPrice,
		"rounding_targets": p.RoundingTargets,
	}
	if p.CreatedAt != nil {
		m["created_at"] = p.CreatedAt.Format(time.RFC3339)
	} else {
		m["created_at"] = nil
	}
	if p.UpdatedAt != nil {
		m["updated_at"] = p.UpdatedAt.Format(time.RFC3339)
	} else {
		m["updated_at"] = nil
	}
	return m
}
