package service

type ShippingServices struct {
	IDShippingServices int     `gorm:"column:id_shipping_services;primaryKey" json:"id_shipping_services"`
	NameCompany        string  `gorm:"column:name_company;size:150;not null"  json:"name_company"`
	ShortName          *string `gorm:"column:short_name;size:50"              json:"short_name,omitempty"`
}

func (ShippingServices) TableName() string { return "shipping_services" }
