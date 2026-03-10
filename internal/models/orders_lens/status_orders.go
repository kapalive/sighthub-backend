// internal/models/orders_lens/status_orders.go
package orders_lens

// StatusOrdersLens ⇄ table: status_orders_lens
type StatusOrdersLens struct {
	IDStatusOrdersLens int    `gorm:"column:id_status_orders_lens;primaryKey;autoIncrement" json:"id_status_orders_lens"`
	StatusOrders       string `gorm:"column:status_orders;type:varchar(255);not null"        json:"status_orders"`
}

func (StatusOrdersLens) TableName() string { return "status_orders_lens" }

func (s *StatusOrdersLens) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_status_orders_lens": s.IDStatusOrdersLens,
		"status_orders":         s.StatusOrders,
	}
}
