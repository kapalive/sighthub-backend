package service

type ShipmentType struct {
	IDShipmentType int    `gorm:"column:id_shipment_type;primaryKey" json:"id_shipment_type"`
	ShipmentTypeV  string `gorm:"column:shipment_type;size:150;not null" json:"shipment_type"`
}

func (ShipmentType) TableName() string { return "shipment_type" }
