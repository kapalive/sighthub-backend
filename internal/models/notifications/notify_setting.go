package notifications

// NotifySetting stores global notification channel toggles per action type.
type NotifySetting struct {
	ID         int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ActionName string `gorm:"column:action_name;type:varchar(50);not null;uniqueIndex" json:"action_name"`
	Email      bool   `gorm:"column:email;not null;default:true" json:"email"`
	SMS        bool   `gorm:"column:sms;not null;default:true" json:"sms"`
}

func (NotifySetting) TableName() string { return "notify_setting" }
