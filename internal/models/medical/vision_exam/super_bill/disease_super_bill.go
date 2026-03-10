package super_bill

type DiseaseSuperBill struct {
	IDDiseaseSuperBill   int64   `gorm:"column:id_disease_super_bill;primaryKey;autoIncrement" json:"id_disease_super_bill"`
	LevelID              int64   `gorm:"column:level_id;not null" json:"level_id"`
	Type                 string  `gorm:"column:type;size:50;not null" json:"type"`
	Code                 string  `gorm:"column:code;size:20;not null" json:"code"`
	Title                string  `gorm:"column:title;size:255;not null" json:"title"`
	GroupSet             *string `gorm:"column:group_set;size:255" json:"group_set"`
	Default              *bool   `gorm:"column:default;default:false" json:"default"`
	SuperBillDiagnosisID *int64  `gorm:"column:super_bill_diagnosis_id" json:"super_bill_diagnosis_id"`
	Include              bool    `gorm:"column:include;not null;default:true" json:"include"`
}

func (DiseaseSuperBill) TableName() string { return "disease_super_bill" }
