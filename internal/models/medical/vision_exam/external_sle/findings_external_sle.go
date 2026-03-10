package external_sle

type FindingsExternalSle struct {
	IDFindingsExternalSle int64   `gorm:"column:id_findings_external_sle;primaryKey" json:"id_findings_external_sle"`
	Externals             *string `gorm:"column:externals;size:255" json:"externals"`
	OdLidsLashes          *string `gorm:"column:od_lids_lashes;size:255" json:"od_lids_lashes"`
	OsLidsLashes          *string `gorm:"column:os_lids_lashes;size:255" json:"os_lids_lashes"`
	OdConjunctivaSclera   *string `gorm:"column:od_conjunctiva_sclera;size:255" json:"od_conjunctiva_sclera"`
	OsConjunctivaSclera   *string `gorm:"column:os_conjunctiva_sclera;size:255" json:"os_conjunctiva_sclera"`
	OdCornea              *string `gorm:"column:od_cornea;size:255" json:"od_cornea"`
	OsCornea              *string `gorm:"column:os_cornea;size:255" json:"os_cornea"`
	OdTearFilm            *string `gorm:"column:od_tear_film;size:255" json:"od_tear_film"`
	OsTearFilm            *string `gorm:"column:os_tear_film;size:255" json:"os_tear_film"`
	OdAnteriorChamber     *string `gorm:"column:od_anterior_chamber;size:255" json:"od_anterior_chamber"`
	OsAnteriorChamber     *string `gorm:"column:os_anterior_chamber;size:255" json:"os_anterior_chamber"`
	OdIris                *string `gorm:"column:od_iris;size:255" json:"od_iris"`
	OsIris                *string `gorm:"column:os_iris;size:255" json:"os_iris"`
	OdLens                *string `gorm:"column:od_lens;size:255" json:"od_lens"`
	OsLens                *string `gorm:"column:os_lens;size:255" json:"os_lens"`
}

func (FindingsExternalSle) TableName() string { return "findings_external_sle" }
