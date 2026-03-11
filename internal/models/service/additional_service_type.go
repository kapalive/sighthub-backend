package service

// AdditionalServiceType ⇄ table: add_service_type
type AdditionalServiceType struct {
	IDAddServiceType int    `gorm:"column:id_add_service_type;primaryKey" json:"id_add_service_type"`
	Title            string `gorm:"column:title;size:50;not null"         json:"title"`
}

func (AdditionalServiceType) TableName() string { return "add_service_type" }
