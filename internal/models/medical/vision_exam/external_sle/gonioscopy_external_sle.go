package external_sle

type GonioscopyExternalSle struct {
	IDGonioscopyExternalSle int64   `gorm:"column:id_gonioscopy_external_sle;primaryKey" json:"id_gonioscopy_external_sle"`
	OdSup                   *string `gorm:"column:od_sup;size:255" json:"od_sup"`
	OsSup                   *string `gorm:"column:os_sup;size:255" json:"os_sup"`
	OdInf                   *string `gorm:"column:od_inf;size:255" json:"od_inf"`
	OsInf                   *string `gorm:"column:os_inf;size:255" json:"os_inf"`
	OdNasal                 *string `gorm:"column:od_nasal;size:255" json:"od_nasal"`
	OsNasal                 *string `gorm:"column:os_nasal;size:255" json:"os_nasal"`
	OdTemp                  *string `gorm:"column:od_temp;size:255" json:"od_temp"`
	OsTemp                  *string `gorm:"column:os_temp;size:255" json:"os_temp"`
}

func (GonioscopyExternalSle) TableName() string { return "gonioscopy_external_sle" }
