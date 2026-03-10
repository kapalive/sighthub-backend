package misc

type MiscInvoiceItem struct {
	IDMiscItem  int64   `gorm:"column:id_misc_item;primaryKey;autoIncrement" json:"id_misc_item"`
	ItemNumber  string  `gorm:"column:item_number;size:50;not null;uniqueIndex" json:"item_number"`
	Description string  `gorm:"column:description;size:255;not null"           json:"description"`
	PbKey       string  `gorm:"column:pb_key;not null"                         json:"pb_key"`
	Cost        *string `gorm:"column:cost;type:numeric(10,2)"                 json:"cost,omitempty"`
	Price       *string `gorm:"column:price;type:numeric(10,2)"                json:"price,omitempty"`
	SaleKey     *string `gorm:"column:sale_key;size:50"                        json:"sale_key,omitempty"`
	CanLookup   bool    `gorm:"column:can_lookup;not null;default:true"        json:"can_lookup"`
	Active      bool    `gorm:"column:active;not null;default:true"            json:"active"`
}

func (MiscInvoiceItem) TableName() string { return "misc_invoice_item" }
