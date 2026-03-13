package email_template_service

import (
	"errors"

	"gorm.io/gorm"

	mktModel "sighthub-backend/internal/models/marketing"
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ── DTOs ─────────────────────────────────────────────────────────────────────

type TemplateItem struct {
	IDEmailTemplate int    `json:"id_email_template"`
	Name            string `json:"name"`
	Category        string `json:"category"`
	IsDefault       bool   `json:"is_default"`
}

type CategoryResult struct {
	Category           string         `json:"category"`
	SelectedTemplateID *int           `json:"selected_template_id"`
	Templates          []TemplateItem `json:"templates"`
}

type OrgTemplateResult struct {
	IDOrgEmailTemplate int    `json:"id_org_email_template"`
	Category           string `json:"category"`
	TemplateID         int    `json:"template_id"`
}

// ── Methods ───────────────────────────────────────────────────────────────────

// GetAllTemplates returns templates grouped by category with the org-selected template id.
func (s *Service) GetAllTemplates() ([]CategoryResult, error) {
	var templates []mktModel.EmailTemplate
	if err := s.db.Order("category ASC, is_default DESC, name ASC").Find(&templates).Error; err != nil {
		return nil, err
	}

	var orgSelections []mktModel.OrgEmailTemplate
	s.db.Find(&orgSelections)

	selections := map[string]int{}
	for _, o := range orgSelections {
		selections[o.Category] = o.TemplateID
	}

	type catData struct {
		templates []TemplateItem
		defaultID *int
	}
	// preserve insertion order via slice of keys
	catOrder := []string{}
	cats := map[string]*catData{}

	for _, t := range templates {
		if _, exists := cats[t.Category]; !exists {
			cats[t.Category] = &catData{}
			catOrder = append(catOrder, t.Category)
		}
		item := TemplateItem{
			IDEmailTemplate: t.IDEmailTemplate,
			Name:            t.Name,
			Category:        t.Category,
			IsDefault:       t.IsDefault,
		}
		cats[t.Category].templates = append(cats[t.Category].templates, item)
		if t.IsDefault && cats[t.Category].defaultID == nil {
			id := t.IDEmailTemplate
			cats[t.Category].defaultID = &id
		}
	}

	result := make([]CategoryResult, 0, len(catOrder))
	for _, cat := range catOrder {
		data := cats[cat]
		var selectedID *int
		if id, ok := selections[cat]; ok {
			selectedID = &id
		} else {
			selectedID = data.defaultID
		}
		result = append(result, CategoryResult{
			Category:           cat,
			SelectedTemplateID: selectedID,
			Templates:          data.templates,
		})
	}
	return result, nil
}

// SetOrgTemplate upserts the org-selected template for the template's category.
func (s *Service) SetOrgTemplate(templateID int) (*OrgTemplateResult, error) {
	var tmpl mktModel.EmailTemplate
	if err := s.db.Where("id_email_template = ?", templateID).First(&tmpl).Error; err != nil {
		return nil, errors.New("template not found")
	}

	var org mktModel.OrgEmailTemplate
	err := s.db.Where("category = ?", tmpl.Category).First(&org).Error
	if err == nil {
		org.TemplateID = templateID
		if err := s.db.Save(&org).Error; err != nil {
			return nil, err
		}
	} else {
		org = mktModel.OrgEmailTemplate{Category: tmpl.Category, TemplateID: templateID}
		if err := s.db.Create(&org).Error; err != nil {
			return nil, err
		}
	}

	return &OrgTemplateResult{
		IDOrgEmailTemplate: org.IDOrgEmailTemplate,
		Category:           org.Category,
		TemplateID:         org.TemplateID,
	}, nil
}
