package claim_service

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	empModel "sighthub-backend/internal/models/employees"
	insuranceModel "sighthub-backend/internal/models/insurance"
	locModel "sighthub-backend/internal/models/location"
	marketingModel "sighthub-backend/internal/models/marketing"
	"sighthub-backend/pkg/activitylog"
	"sighthub-backend/pkg/pdfutil"
)

// fieldNames are the ClaimTemplate DB column keys that map to PDF fields.
var fieldNames = []string{
	"insurance_type",
	"insurance_name", "insurance_address", "insurance_address2",
	"insurance_city_state_zip",
	"condition_employment", "condition_auto_accident", "accident_state",
	"condition_other_accident", "patient_signature", "insured_signature",
	"service1_rendering_npi", "service2_rendering_npi", "service3_rendering_npi",
	"service4_rendering_npi", "service5_rendering_npi", "service6_rendering_npi",
	"federal_tax_id", "tax_id_type", "accept_assignment",
	"physician_signature", "physician_signature_date",
	"facility_name", "facility_street", "facility_city_state_zip", "facility_npi",
	"billing_provider_name", "billing_provider_street",
	"billing_provider_city_state_zip", "billing_provider_phone_area",
	"billing_provider_phone", "billing_provider_npi",
}

// templateFieldValue gets the pointer value of a named field from ClaimTemplate.
func templateFieldValue(t *insuranceModel.ClaimTemplate, field string) *string {
	switch field {
	case "insurance_type":
		return t.InsuranceType
	case "insurance_name":
		return t.InsuranceName
	case "insurance_address":
		return t.InsuranceAddress
	case "insurance_address2":
		return t.InsuranceAddress2
	case "insurance_city_state_zip":
		return t.InsuranceCityStateZip
	case "condition_employment":
		return t.ConditionEmployment
	case "condition_auto_accident":
		return t.ConditionAutoAccident
	case "accident_state":
		return t.AccidentState
	case "condition_other_accident":
		return t.ConditionOtherAccident
	case "patient_signature":
		return t.PatientSignature
	case "insured_signature":
		return t.InsuredSignature
	case "service1_rendering_npi":
		return t.Service1RenderingNpi
	case "service2_rendering_npi":
		return t.Service2RenderingNpi
	case "service3_rendering_npi":
		return t.Service3RenderingNpi
	case "service4_rendering_npi":
		return t.Service4RenderingNpi
	case "service5_rendering_npi":
		return t.Service5RenderingNpi
	case "service6_rendering_npi":
		return t.Service6RenderingNpi
	case "federal_tax_id":
		return t.FederalTaxID
	case "tax_id_type":
		return t.TaxIDType
	case "accept_assignment":
		return t.AcceptAssignment
	case "physician_signature":
		return t.PhysicianSignature
	case "physician_signature_date":
		return t.PhysicianSignatureDate
	case "facility_name":
		return t.FacilityName
	case "facility_street":
		return t.FacilityStreet
	case "facility_city_state_zip":
		return t.FacilityCityStateZip
	case "facility_npi":
		return t.FacilityNpi
	case "billing_provider_name":
		return t.BillingProviderName
	case "billing_provider_street":
		return t.BillingProviderStreet
	case "billing_provider_city_state_zip":
		return t.BillingProviderCSZ
	case "billing_provider_phone_area":
		return t.BillingProviderPhoneArea
	case "billing_provider_phone":
		return t.BillingProviderPhone
	case "billing_provider_npi":
		return t.BillingProviderNpi
	}
	return nil
}

// setTemplateField sets a named field on ClaimTemplate (uppercases non-radio values).
func setTemplateField(t *insuranceModel.ClaimTemplate, field, value string) {
	up := value
	if value != "" && !strings.HasPrefix(value, "/") {
		up = strings.ToUpper(value)
	}
	var v *string
	if up != "" {
		v = &up
	}
	switch field {
	case "insurance_type":
		t.InsuranceType = v
	case "insurance_name":
		t.InsuranceName = v
	case "insurance_address":
		t.InsuranceAddress = v
	case "insurance_address2":
		t.InsuranceAddress2 = v
	case "insurance_city_state_zip":
		t.InsuranceCityStateZip = v
	case "condition_employment":
		t.ConditionEmployment = v
	case "condition_auto_accident":
		t.ConditionAutoAccident = v
	case "accident_state":
		t.AccidentState = v
	case "condition_other_accident":
		t.ConditionOtherAccident = v
	case "patient_signature":
		t.PatientSignature = v
	case "insured_signature":
		t.InsuredSignature = v
	case "service1_rendering_npi":
		t.Service1RenderingNpi = v
	case "service2_rendering_npi":
		t.Service2RenderingNpi = v
	case "service3_rendering_npi":
		t.Service3RenderingNpi = v
	case "service4_rendering_npi":
		t.Service4RenderingNpi = v
	case "service5_rendering_npi":
		t.Service5RenderingNpi = v
	case "service6_rendering_npi":
		t.Service6RenderingNpi = v
	case "federal_tax_id":
		t.FederalTaxID = v
	case "tax_id_type":
		t.TaxIDType = v
	case "accept_assignment":
		t.AcceptAssignment = v
	case "physician_signature":
		t.PhysicianSignature = v
	case "physician_signature_date":
		t.PhysicianSignatureDate = v
	case "facility_name":
		t.FacilityName = v
	case "facility_street":
		t.FacilityStreet = v
	case "facility_city_state_zip":
		t.FacilityCityStateZip = v
	case "facility_npi":
		t.FacilityNpi = v
	case "billing_provider_name":
		t.BillingProviderName = v
	case "billing_provider_street":
		t.BillingProviderStreet = v
	case "billing_provider_city_state_zip":
		t.BillingProviderCSZ = v
	case "billing_provider_phone_area":
		t.BillingProviderPhoneArea = v
	case "billing_provider_phone":
		t.BillingProviderPhone = v
	case "billing_provider_npi":
		t.BillingProviderNpi = v
	}
}

// templateDefaultFields returns non-nil fields from the ClaimTemplate as a map.
func templateDefaultFields(t *insuranceModel.ClaimTemplate) map[string]string {
	result := make(map[string]string)
	for _, f := range fieldNames {
		if v := templateFieldValue(t, f); v != nil && *v != "" {
			result[f] = *v
		}
	}
	return result
}

// applyFields sets all default_fields onto the ClaimTemplate.
func applyFields(t *insuranceModel.ClaimTemplate, fields map[string]interface{}) {
	for _, f := range fieldNames {
		if v, ok := fields[f]; ok && v != nil {
			if str, ok := v.(string); ok {
				setTemplateField(t, f, str)
			}
		}
	}
}

// buildTemplateData populates a full field map from related models + ClaimTemplate defaults.
func (s *Service) buildTemplateData(t *insuranceModel.ClaimTemplate) map[string]string {
	data := make(map[string]string)

	if t.DoctorID != nil {
		var doc empModel.Employee
		if s.db.First(&doc, t.DoctorID).Error == nil {
			var npi empModel.DoctorNpiNumber
			s.db.Where("employee_id = ?", doc.IDEmployee).First(&npi)

			docName := fmt.Sprintf("Dr. %s %s", doc.FirstName, doc.LastName)
			if npi.PrintingName != nil && *npi.PrintingName != "" {
				docName = *npi.PrintingName
			}
			data["billing_provider_name"] = docName
			data["physician_signature"] = docName
			if npi.IDDoctorNPINumber != 0 {
				if npi.DRNPINumber != "" {
					data["billing_provider_npi"] = npi.DRNPINumber
				}
				if npi.EIN != nil && *npi.EIN != "" {
					data["federal_tax_id"] = *npi.EIN
					data["tax_id_type"] = "/EIN"
				}
			}
		}
	}

	if t.InsuranceCompanyID != nil {
		var ic insuranceModel.InsuranceCompany
		if s.db.First(&ic, t.InsuranceCompanyID).Error == nil {
			if ic.CompanyName != "" {
				data["insurance_name"] = ic.CompanyName
			}
			if ic.Address != nil && *ic.Address != "" {
				data["insurance_address"] = *ic.Address
			}
			if ic.AddressLine2 != nil && *ic.AddressLine2 != "" {
				data["insurance_address2"] = *ic.AddressLine2
			}
			parts := []string{}
			if ic.City != nil && *ic.City != "" {
				parts = append(parts, *ic.City)
			}
			if ic.State != nil && *ic.State != "" {
				parts = append(parts, *ic.State)
			}
			if ic.ZipCode != nil && *ic.ZipCode != "" {
				parts = append(parts, *ic.ZipCode)
			}
			if len(parts) > 0 {
				data["insurance_city_state_zip"] = strings.Join(parts, " ")
			}
		}
	}

	if t.LocationID != nil {
		var loc locModel.Location
		if s.db.First(&loc, t.LocationID).Error == nil {
			data["facility_name"] = loc.FullName
			if loc.StreetAddress != nil && *loc.StreetAddress != "" {
				data["facility_street"] = *loc.StreetAddress
			}
			parts := []string{}
			if loc.City != nil && *loc.City != "" {
				parts = append(parts, *loc.City)
			}
			if loc.State != nil && *loc.State != "" {
				parts = append(parts, *loc.State)
			}
			if loc.PostalCode != nil && *loc.PostalCode != "" {
				parts = append(parts, *loc.PostalCode)
			}
			if len(parts) > 0 {
				data["facility_city_state_zip"] = strings.Join(parts, " ")
			}
			if loc.Phone != nil && *loc.Phone != "" {
				digits := strings.Map(func(r rune) rune {
					if r >= '0' && r <= '9' {
						return r
					}
					return -1
				}, *loc.Phone)
				if len(digits) >= 10 {
					data["billing_provider_phone_area"] = digits[:3]
					data["billing_provider_phone"] = digits[3:10]
				}
			}
		}
	}

	// Override with stored template defaults
	for k, v := range templateDefaultFields(t) {
		data[k] = v
	}

	return data
}

// getCMS1500TemplatePath gets the PDF template path from the database.
func (s *Service) getCMS1500TemplatePath() (string, error) {
	var form marketingModel.Form
	if err := s.db.Where("id_form = ? AND is_active = ?", 1, true).First(&form).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("PDF template not found")
		}
		return "", err
	}
	return form.PathFile, nil
}

// ── GET /templates ────────────────────────────────────────────────────────────

func (s *Service) ListClaimTemplates() ([]map[string]interface{}, error) {
	var templates []insuranceModel.ClaimTemplate
	if err := s.db.Order("name").Find(&templates).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(templates))
	for _, t := range templates {
		item := map[string]interface{}{
			"template_id": t.IDClaimTemplate,
			"name":        t.Name,
			"description": t.Description,
			"location_name":    nil,
			"insurance_name":   nil,
			"doctor_name":      nil,
		}
		if t.LocationID != nil {
			var loc locModel.Location
			if s.db.First(&loc, t.LocationID).Error == nil {
				item["location_name"] = loc.FullName
			}
		}
		if t.InsuranceCompanyID != nil {
			var ic insuranceModel.InsuranceCompany
			if s.db.First(&ic, t.InsuranceCompanyID).Error == nil {
				item["insurance_name"] = ic.CompanyName
			}
		}
		if t.DoctorID != nil {
			var doc empModel.Employee
			if s.db.First(&doc, t.DoctorID).Error == nil {
				item["doctor_name"] = fmt.Sprintf("Dr. %s %s", doc.FirstName, doc.LastName)
			}
		}
		result = append(result, item)
	}
	return result, nil
}

// ── GET /templates/:id ────────────────────────────────────────────────────────

func (s *Service) GetClaimTemplate(templateID int) (map[string]interface{}, error) {
	var t insuranceModel.ClaimTemplate
	if err := s.db.First(&t, templateID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("template not found")
		}
		return nil, err
	}

	return map[string]interface{}{
		"template_id":         t.IDClaimTemplate,
		"name":                t.Name,
		"description":         t.Description,
		"location_id":         t.LocationID,
		"doctor_id":           t.DoctorID,
		"insurance_company_id": t.InsuranceCompanyID,
		"template_fields":     s.buildTemplateData(&t),
	}, nil
}

// ── POST /templates ───────────────────────────────────────────────────────────

func (s *Service) CreateClaimTemplate(data map[string]interface{}) (map[string]interface{}, error) {
	name := strings.TrimSpace(strVal(data["name"]))
	if name == "" {
		return nil, errors.New("name is required")
	}

	t := insuranceModel.ClaimTemplate{
		Name:        name,
		Description: strPtr(strVal(data["description"])),
	}
	if v := data["location_id"]; v != nil {
		if id, ok := toInt(v); ok {
			t.LocationID = &id
		}
	}
	if v := data["doctor_id"]; v != nil {
		if id, ok := toInt(v); ok {
			t.DoctorID = &id
		}
	}
	if v := data["insurance_company_id"]; v != nil {
		if id, ok := toInt(v); ok {
			t.InsuranceCompanyID = &id
		}
	}

	if fields, ok := data["default_fields"].(map[string]interface{}); ok {
		applyFields(&t, fields)
	}

	if err := s.db.Create(&t).Error; err != nil {
		return nil, err
	}

	activitylog.Log(s.db, "claim", "template_create", activitylog.WithEntity(int64(t.IDClaimTemplate)))

	return map[string]interface{}{
		"message":     "Template created",
		"template_id": t.IDClaimTemplate,
	}, nil
}

// ── PUT /templates/:id ────────────────────────────────────────────────────────

func (s *Service) UpdateClaimTemplate(templateID int, data map[string]interface{}) (map[string]interface{}, error) {
	var t insuranceModel.ClaimTemplate
	if err := s.db.First(&t, templateID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("template not found")
		}
		return nil, err
	}

	if v, ok := data["name"]; ok {
		name := strings.TrimSpace(strVal(v))
		if name == "" {
			return nil, errors.New("name cannot be empty")
		}
		t.Name = name
	}
	if v, ok := data["description"]; ok {
		t.Description = strPtr(strVal(v))
	}
	if v, ok := data["location_id"]; ok {
		if v == nil {
			t.LocationID = nil
		} else if id, ok2 := toInt(v); ok2 {
			t.LocationID = &id
		}
	}
	if v, ok := data["doctor_id"]; ok {
		if v == nil {
			t.DoctorID = nil
		} else if id, ok2 := toInt(v); ok2 {
			t.DoctorID = &id
		}
	}
	if v, ok := data["insurance_company_id"]; ok {
		if v == nil {
			t.InsuranceCompanyID = nil
		} else if id, ok2 := toInt(v); ok2 {
			t.InsuranceCompanyID = &id
		}
	}
	if fields, ok := data["default_fields"].(map[string]interface{}); ok {
		applyFields(&t, fields)
	}

	if err := s.db.Save(&t).Error; err != nil {
		return nil, err
	}

	activitylog.Log(s.db, "claim", "template_update", activitylog.WithEntity(int64(templateID)))

	return map[string]interface{}{
		"message":     "Template updated",
		"template_id": t.IDClaimTemplate,
	}, nil
}

// ── DELETE /templates/:id ─────────────────────────────────────────────────────

func (s *Service) DeleteClaimTemplate(templateID int) error {
	var t insuranceModel.ClaimTemplate
	if err := s.db.First(&t, templateID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("template not found")
		}
		return err
	}
	if err := s.db.Delete(&t).Error; err != nil {
		return err
	}
	activitylog.Log(s.db, "claim", "template_delete", activitylog.WithEntity(int64(templateID)))
	return nil
}

// ── POST /templates/:id/render ────────────────────────────────────────────────

func (s *Service) RenderClaimPDF(templateID int, templateFields, otherFields map[string]interface{}, color string) ([]byte, string, error) {
	var t insuranceModel.ClaimTemplate
	if err := s.db.First(&t, templateID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("template not found")
		}
		return nil, "", err
	}

	templatePath, err := s.getCMS1500TemplatePath()
	if err != nil {
		return nil, "", err
	}

	// Merge: other_fields override template_fields
	merged := make(map[string]string)
	for k, v := range templateFields {
		if str, ok := v.(string); ok && str != "" {
			merged[k] = str
		}
	}
	for k, v := range otherFields {
		if str, ok := v.(string); ok && str != "" {
			merged[k] = str
		}
	}

	pdfBytes, err := pdfutil.FillCMS1500Bytes(templatePath, merged)
	if err != nil {
		return nil, "", err
	}

	safeName := strings.NewReplacer(" ", "_", "/", "-").Replace(t.Name)
	filename := fmt.Sprintf("cms1500_%s_%s.pdf", safeName, color)
	return pdfBytes, filename, nil
}

// ── GET /templates/:id/pdf ────────────────────────────────────────────────────

func (s *Service) GetTemplatePDF(templateID int) ([]byte, string, error) {
	var t insuranceModel.ClaimTemplate
	if err := s.db.First(&t, templateID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("template not found")
		}
		return nil, "", err
	}

	templatePath, err := s.getCMS1500TemplatePath()
	if err != nil {
		return nil, "", err
	}

	data := s.buildTemplateData(&t)
	pdfBytes, err := pdfutil.FillCMS1500Bytes(templatePath, data)
	if err != nil {
		return nil, "", err
	}

	safeName := strings.NewReplacer(" ", "_", "/", "-").Replace(t.Name)
	return pdfBytes, fmt.Sprintf("cms1500_%s.pdf", safeName), nil
}

// ── GET /templates/:id/preview ────────────────────────────────────────────────

func (s *Service) PreviewTemplatePDF(templateID int) ([]byte, string, error) {
	var t insuranceModel.ClaimTemplate
	if err := s.db.First(&t, templateID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("template not found")
		}
		return nil, "", err
	}

	templatePath, err := s.getCMS1500TemplatePath()
	if err != nil {
		return nil, "", err
	}

	data := s.buildTemplateData(&t)
	pdfBytes, err := pdfutil.FillCMS1500Bytes(templatePath, data)
	if err != nil {
		return nil, "", err
	}

	safeName := strings.NewReplacer(" ", "_", "/", "-").Replace(t.Name)
	return pdfBytes, fmt.Sprintf("cms1500_%s_preview.pdf", safeName), nil
}

// ── small helpers ─────────────────────────────────────────────────────────────

func strVal(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	}
	return 0, false
}
