package marketing

type ReferralSource struct {
	IDReferralSources int     `gorm:"column:id_referral_sources;primaryKey" json:"id_referral_sources"`
	Title             string  `gorm:"column:title;not null"                 json:"title"`
	Description       *string `gorm:"column:description;type:text"          json:"description,omitempty"`
}

func (ReferralSource) TableName() string { return "referral_sources" }
