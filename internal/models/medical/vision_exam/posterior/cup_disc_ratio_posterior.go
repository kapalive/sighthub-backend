package posterior

type CupDiscRatioPosterior struct {
	IDCupDiscRatioPosterior int64   `gorm:"column:id_cup_disc_ratio_posterior;primaryKey" json:"id_cup_disc_ratio_posterior"`
	OdV                     *string `gorm:"column:od_v;size:255" json:"od_v"`
	OsV                     *string `gorm:"column:os_v;size:255" json:"os_v"`
	OdH                     *string `gorm:"column:od_h;size:255" json:"od_h"`
	OsH                     *string `gorm:"column:os_h;size:255" json:"os_h"`
}

func (CupDiscRatioPosterior) TableName() string { return "cup_disc_ratio_posterior" }
