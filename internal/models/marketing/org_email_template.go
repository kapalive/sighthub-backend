// internal/models/marketing/org_email_template.go
package marketing

// OrgEmailTemplate ⇄ org_email_template
type OrgEmailTemplate struct {
	IDOrgEmailTemplate int    `gorm:"column:id_org_email_template;primaryKey;autoIncrement" json:"id_org_email_template"`
	Category           string `gorm:"column:category;type:varchar(50);not null;uniqueIndex" json:"category"`
	TemplateID         int    `gorm:"column:template_id;not null"                           json:"template_id"`
}

func (OrgEmailTemplate) TableName() string { return "org_email_template" }

func (o *OrgEmailTemplate) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id_org_email_template": o.IDOrgEmailTemplate,
		"category":              o.Category,
		"template_id":           o.TemplateID,
	}
}
