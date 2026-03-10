package external_sle

type PachExternalSle struct {
	IDPachExternalSle int64   `gorm:"column:id_pach_external_sle;primaryKey" json:"id_pach_external_sle"`
	Od                *string `gorm:"column:od;size:255" json:"od"`
	Os                *string `gorm:"column:os;size:255" json:"os"`
}

func (PachExternalSle) TableName() string { return "pach_external_sle" }
