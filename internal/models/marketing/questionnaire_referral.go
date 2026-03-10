package marketing

import "time"

type QuestionnaireReferral struct {
	IDQuestionnaireReferral int64      `gorm:"column:id_questionnaire_referral;primaryKey" json:"id_questionnaire_referral"`
	VisitReasonsID          *int       `gorm:"column:visit_reasons_id"                     json:"visit_reasons_id,omitempty"`
	ReferralSourcesID       *int       `gorm:"column:referral_sources_id"                  json:"referral_sources_id,omitempty"`
	PatientID               *int64     `gorm:"column:patient_id"                           json:"patient_id,omitempty"`
	DatetimeCreated         *time.Time `gorm:"column:datetime_created;type:timestamptz"    json:"datetime_created,omitempty"`
	LocationID              *int       `gorm:"column:location_id"                          json:"location_id,omitempty"`
	EmployeeID              *int64     `gorm:"column:employee_id"                          json:"employee_id,omitempty"`

	VisitReason     *VisitReason     `gorm:"foreignKey:VisitReasonsID"     json:"visit_reason,omitempty"`
	ReferralSource  *ReferralSource  `gorm:"foreignKey:ReferralSourcesID"  json:"referral_source,omitempty"`
}

func (QuestionnaireReferral) TableName() string { return "questionnaire_referral" }
