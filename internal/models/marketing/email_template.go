// internal/models/marketing/email_template.go
package marketing

// EmailTemplate ⇄ email_template
type EmailTemplate struct {
	IDEmailTemplate int    `gorm:"column:id_email_template;primaryKey;autoIncrement" json:"id_email_template"`
	Name            string `gorm:"column:name;type:varchar(100);not null;uniqueIndex" json:"name"`
	Category        string `gorm:"column:category;type:varchar(50);not null"          json:"category"`
	IsDefault       bool   `gorm:"column:is_default;not null;default:false"           json:"is_default"`
}

func (EmailTemplate) TableName() string { return "email_template" }

func (e *EmailTemplate) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_email_template": e.IDEmailTemplate,
		"name":              e.Name,
		"category":          e.Category,
		"is_default":        e.IsDefault,
	}
}
