package referral

type ReferralDoctor struct {
	IDReferralDoctor int64   `gorm:"column:id_referral_doctor;primaryKey;autoIncrement" json:"id_referral_doctor"`
	Salutation       *string `gorm:"column:salutation;size:4" json:"salutation"`
	Npi              *string `gorm:"column:npi;size:20" json:"npi"`
	LastName         *string `gorm:"column:last_name" json:"last_name"`
	FirstName        *string `gorm:"column:first_name" json:"first_name"`
	Address          *string `gorm:"column:address;type:text" json:"address"`
	Address2         *string `gorm:"column:address2;size:20" json:"address2"`
	City             *string `gorm:"column:city;size:100" json:"city"`
	State            *string `gorm:"column:state;size:2" json:"state"`
	Zip              *string `gorm:"column:zip;size:6" json:"zip"`
	Phone            *string `gorm:"column:phone;size:16" json:"phone"`
	Fax              *string `gorm:"column:fax;size:16" json:"fax"`
	Email            *string `gorm:"column:email;size:50" json:"email"`
}

func (ReferralDoctor) TableName() string { return "referral_doctor" }
