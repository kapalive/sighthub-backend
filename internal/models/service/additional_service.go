package service

type AdditionalService struct {
	IDAdditionalService int64    `gorm:"column:id_additional_service;primaryKey" json:"id_additional_service"`
	ItemNumber          *string  `gorm:"column:item_number;size:12"              json:"item_number,omitempty"`
	AddServiceTypeID    *int     `gorm:"column:add_service_type_id"              json:"add_service_type_id,omitempty"`
	SrCost              *bool    `gorm:"column:sr_cost"                          json:"sr_cost,omitempty"`
	UV                  *bool    `gorm:"column:uv"                               json:"uv,omitempty"`
	AR                  *bool    `gorm:"column:ar"                               json:"ar,omitempty"`
	Tint                *bool    `gorm:"column:tint"                             json:"tint,omitempty"`
	Drill               *bool    `gorm:"column:drill"                            json:"drill,omitempty"`
	Send                *bool    `gorm:"column:send"                             json:"send,omitempty"`
	InvoiceDesc         string   `gorm:"column:invoice_desc;size:100;not null"   json:"invoice_desc"`
	CostPrice           float64  `gorm:"column:cost_price;type:numeric(10,2);default:0.00" json:"cost_price"`
	Price               float64  `gorm:"column:price;type:numeric(10,2);default:0.00"      json:"price"`
	ReportOmit          *bool    `gorm:"column:report_omit"                      json:"report_omit,omitempty"`
	InsVCode            *string  `gorm:"column:ins_v_code;size:6"                json:"ins_v_code,omitempty"`
	ClassLevel          *string  `gorm:"column:class_level;size:10"              json:"class_level,omitempty"`
	InsVCodeAdd         *string  `gorm:"column:ins_v_code_add;size:10"           json:"ins_v_code_add,omitempty"`
	Sort1               *float64 `gorm:"column:sort1"                            json:"sort1,omitempty"`
	Sort2               *float64 `gorm:"column:sort2"                            json:"sort2,omitempty"`
	Visible             bool     `gorm:"column:visible;default:true"             json:"visible"`
	MfrNumber           *string  `gorm:"column:mfr_number;size:60"               json:"mfr_number,omitempty"`
	Photochromatic      *bool    `gorm:"column:photochromatic"                   json:"photochromatic,omitempty"`
	Polarized           *bool    `gorm:"column:polarized"                        json:"polarized,omitempty"`
	CanDrill            *bool    `gorm:"column:can_drill"                        json:"can_drill,omitempty"`
	HighIndex           *bool    `gorm:"column:high_index"                       json:"high_index,omitempty"`
	Digital             *bool    `gorm:"column:digital"                          json:"digital,omitempty"`

	AddServiceType *AdditionalServiceType `gorm:"foreignKey:AddServiceTypeID;references:IDAddServiceType" json:"add_service_type,omitempty"`
}

func (AdditionalService) TableName() string { return "additional_service" }
