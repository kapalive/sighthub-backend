package posterior

type FindingsPosterior struct {
	IDFindingsPosterior int64   `gorm:"column:id_findings_posterior;primaryKey" json:"id_findings_posterior"`
	OdView              *string `gorm:"column:od_view;size:255" json:"od_view"`
	OsView              *string `gorm:"column:os_view;size:255" json:"os_view"`
	OdVitreous          *string `gorm:"column:od_vitreous;size:255" json:"od_vitreous"`
	OsVitreous          *string `gorm:"column:os_vitreous;size:255" json:"os_vitreous"`
	OdMacula            *string `gorm:"column:od_macula;size:255" json:"od_macula"`
	OsMacula            *string `gorm:"column:os_macula;size:255" json:"os_macula"`
	OdBackground        *string `gorm:"column:od_background;size:255" json:"od_background"`
	OsBackground        *string `gorm:"column:os_background;size:255" json:"os_background"`
	OdVessels           *string `gorm:"column:od_vessels;size:255" json:"od_vessels"`
	OsVessels           *string `gorm:"column:os_vessels;size:255" json:"os_vessels"`
	OdPeripheralFundus  *string `gorm:"column:od_peripheral_fundus;size:255" json:"od_peripheral_fundus"`
	OsPeripheralFundus  *string `gorm:"column:os_peripheral_fundus;size:255" json:"os_peripheral_fundus"`
	OdOpticsNerve       *string `gorm:"column:od_optics_nerve;size:255" json:"od_optics_nerve"`
	OsOpticsNerve       *string `gorm:"column:os_optics_nerve;size:255" json:"os_optics_nerve"`
}

func (FindingsPosterior) TableName() string { return "findings_posterior" }
