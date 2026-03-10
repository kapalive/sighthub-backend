// internal/models/vision_exam/social_history.go
package vision_exam

// SocialHistory ↔ table: social_history
type SocialHistory struct {
	IDSocialHistory int64   `gorm:"column:id_social_history;primaryKey;autoIncrement" json:"id_social_history"`
	AlcoholUse      *string `gorm:"column:alcohol_use;type:varchar(20)"               json:"alcohol_use,omitempty"`
	TobaccoUse      *string `gorm:"column:tobacco_use;type:varchar(20)"               json:"tobacco_use,omitempty"`
	AlertToTime     *bool   `gorm:"column:alert_to_time"                              json:"alert_to_time,omitempty"`
	AlertToPlace    *bool   `gorm:"column:alert_to_place"                             json:"alert_to_place,omitempty"`
	AwareOfSelf     *bool   `gorm:"column:aware_of_self"                              json:"aware_of_self,omitempty"`
	SitsUpright     *bool   `gorm:"column:sits_upright"                               json:"sits_upright,omitempty"`
}

func (SocialHistory) TableName() string { return "social_history" }

func (s *SocialHistory) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_social_history": s.IDSocialHistory,
		"alcohol_use":       s.AlcoholUse,
		"tobacco_use":       s.TobaccoUse,
		"alert_to_time":     s.AlertToTime,
		"alert_to_place":    s.AlertToPlace,
		"aware_of_self":     s.AwareOfSelf,
		"sits_upright":      s.SitsUpright,
	}
}
