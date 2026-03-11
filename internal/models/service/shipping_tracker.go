package service

type ShippingTracker struct {
	IDTracker         int64   `gorm:"column:id_tracker;primaryKey;autoIncrement" json:"id_tracker"`
	Tracker           string  `gorm:"column:tracker;size:30;not null"            json:"tracker"`
	ShippingLabelPath *string `gorm:"column:shipping_label_path;size:255"        json:"shipping_label_path,omitempty"`
}

func (ShippingTracker) TableName() string { return "shipping_tracker" }
