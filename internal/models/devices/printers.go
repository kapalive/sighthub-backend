package devices

type Printer struct {
	IDPrinter  int     `gorm:"column:id_printer;primaryKey;autoIncrement" json:"id_printer"`
	IdDevice   string  `gorm:"column:id_device;size:255;not null"          json:"id_device"`
	NameDevice string  `gorm:"column:name_device;size:100;not null"        json:"name_device"`
	LocationID int     `gorm:"column:location_id;not null"                json:"location_id"`
	Note       *string `gorm:"column:note;size:255"                        json:"note,omitempty"`
}

func (Printer) TableName() string { return "printers" }
