package employees

import "fmt"

// ElectronReferralSignature ⇄ electron_referral_signature
type ElectronReferralSignature struct {
	IDElectronReferralSignature int     `gorm:"column:id_electron_referral_signature;primaryKey;autoIncrement" json:"id_electron_referral_signature"`
	Description                 *string `gorm:"column:description;type:text"                                    json:"description,omitempty"`
	PathLinkImg                 *string `gorm:"column:path_link_img;type:varchar(255)"                          json:"path_link_img,omitempty"`
	EmployeeID                  int     `gorm:"column:employee_id;not null"                                     json:"employee_id"`

	// Если надо прелоадить сотрудника — раскомментируй:
	// Employee *Employee `gorm:"foreignKey:EmployeeID;references:IDEmployee" json:"-"`
}

func (ElectronReferralSignature) TableName() string { return "electron_referral_signature" }

func (e *ElectronReferralSignature) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_electron_referral_signature": e.IDElectronReferralSignature,
		"description":                    e.Description,
		"path_link_img":                  e.PathLinkImg,
		"employee_id":                    e.EmployeeID,
	}
}

func (e *ElectronReferralSignature) String() string {
	return fmt.Sprintf("<ElectronReferralSignature %d - Employee %d>", e.IDElectronReferralSignature, e.EmployeeID)
}
