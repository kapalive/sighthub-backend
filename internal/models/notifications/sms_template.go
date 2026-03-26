package notifications

import "time"

type SMSTemplate struct {
	IDSMSTemplate int       `gorm:"column:id_sms_template;primaryKey" json:"id_sms_template"`
	Category      string    `gorm:"column:category;type:varchar(50);not null" json:"category"`
	Name          string    `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Body          string    `gorm:"column:body;type:text;not null" json:"body"`
	IsSystem      bool      `gorm:"column:is_system;not null;default:false" json:"is_system"`
	Active        bool      `gorm:"column:active;not null;default:true" json:"active"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (SMSTemplate) TableName() string { return "sms_template" }
