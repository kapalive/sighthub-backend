package settings_service

import (
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"gorm.io/gorm"

	appointmentModel "sighthub-backend/internal/models/appointment"
	"sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/frames"
	"sighthub-backend/internal/models/general"
	"sighthub-backend/internal/models/insurance"
	"sighthub-backend/internal/models/lab_ticket"
	"sighthub-backend/internal/models/lenses"
	"sighthub-backend/internal/models/location"
	"sighthub-backend/internal/models/notifications"
	serviceModel "sighthub-backend/internal/models/service"
)

type Service struct{ DB *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{DB: db} }

// ══════════════════════════════════════════════════════════════════════════════
//  SMTP
// ══════════════════════════════════════════════════════════════════════════════

func (s *Service) ListSMTP() ([]map[string]interface{}, error) {
	var configs []general.SmtpClient
	if err := s.DB.Order("id_smtp_client").Find(&configs).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(configs))
	for i, c := range configs {
		m := c.ToMap()
		// add location_name
		if c.LocationID != nil {
			var name string
			s.DB.Raw(`SELECT full_name FROM location WHERE id_location = ?`, *c.LocationID).Row().Scan(&name)
			m["location_name"] = name
		}
		out[i] = m
	}
	return out, nil
}

func (s *Service) GetSMTP(id int) (map[string]interface{}, error) {
	var c general.SmtpClient
	if err := s.DB.First(&c, id).Error; err != nil {
		return nil, err
	}
	m := c.ToMap()
	if c.LocationID != nil {
		var name string
		s.DB.Raw(`SELECT full_name FROM location WHERE id_location = ?`, *c.LocationID).Row().Scan(&name)
		m["location_name"] = name
	}
	return m, nil
}

func (s *Service) CreateSMTP(data map[string]interface{}) (map[string]interface{}, error) {
	required := []string{"smtp_host", "smtp_port", "smtp_username", "smtp_password", "label", "name_key"}
	var missing []string
	for _, f := range required {
		if _, ok := data[f]; !ok {
			missing = append(missing, f)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("Missing required fields: %s", strings.Join(missing, ", "))
	}

	c := general.SmtpClient{
		Label:       fmt.Sprintf("%v", data["label"]),
		NameKey:     fmt.Sprintf("%v", data["name_key"]),
		SMTPHost:    fmt.Sprintf("%v", data["smtp_host"]),
		SMTPPort:    toInt(data["smtp_port"], 587),
		SMTPUsername: fmt.Sprintf("%v", data["smtp_username"]),
		SMTPPassword: fmt.Sprintf("%v", data["smtp_password"]),
		UseTLS:      toBool(data["use_tls"], true),
		UseSSL:      toBool(data["use_ssl"], false),
		Active:      toBool(data["active"], true),
	}
	if v, ok := data["location_id"]; ok && v != nil {
		id := toInt(v, 0)
		if id > 0 {
			c.LocationID = &id
		}
	}
	if v, ok := data["from_display"]; ok && v != nil {
		s := fmt.Sprintf("%v", v)
		c.FromDisplay = &s
	}
	if v, ok := data["sender_email"]; ok && v != nil {
		s := fmt.Sprintf("%v", v)
		c.SenderEmail = &s
	}

	if err := s.DB.Create(&c).Error; err != nil {
		return nil, err
	}
	return c.ToMap(), nil
}

func (s *Service) UpdateSMTP(id int, data map[string]interface{}) (map[string]interface{}, error) {
	var c general.SmtpClient
	if err := s.DB.First(&c, id).Error; err != nil {
		return nil, err
	}

	updatable := []string{"label", "name_key", "location_id", "smtp_host", "smtp_port",
		"smtp_username", "smtp_password", "from_display", "sender_email",
		"use_tls", "use_ssl", "active"}

	hasUpdate := false
	for _, f := range updatable {
		if _, ok := data[f]; ok {
			hasUpdate = true
			break
		}
	}
	if !hasUpdate {
		return nil, fmt.Errorf("No valid fields provided for update")
	}

	if v, ok := data["label"]; ok {
		c.Label = fmt.Sprintf("%v", v)
	}
	if v, ok := data["name_key"]; ok {
		c.NameKey = fmt.Sprintf("%v", v)
	}
	if v, ok := data["location_id"]; ok {
		if v == nil {
			c.LocationID = nil
		} else {
			id := toInt(v, 0)
			c.LocationID = &id
		}
	}
	if v, ok := data["smtp_host"]; ok {
		c.SMTPHost = fmt.Sprintf("%v", v)
	}
	if v, ok := data["smtp_port"]; ok {
		c.SMTPPort = toInt(v, c.SMTPPort)
	}
	if v, ok := data["smtp_username"]; ok {
		c.SMTPUsername = fmt.Sprintf("%v", v)
	}
	if v, ok := data["smtp_password"]; ok {
		c.SMTPPassword = fmt.Sprintf("%v", v)
	}
	if v, ok := data["from_display"]; ok {
		if v == nil {
			c.FromDisplay = nil
		} else {
			s := fmt.Sprintf("%v", v)
			c.FromDisplay = &s
		}
	}
	if v, ok := data["sender_email"]; ok {
		if v == nil {
			c.SenderEmail = nil
		} else {
			s := fmt.Sprintf("%v", v)
			c.SenderEmail = &s
		}
	}
	if v, ok := data["use_tls"]; ok {
		c.UseTLS = toBool(v, c.UseTLS)
	}
	if v, ok := data["use_ssl"]; ok {
		c.UseSSL = toBool(v, c.UseSSL)
	}
	if v, ok := data["active"]; ok {
		c.Active = toBool(v, c.Active)
	}

	if err := s.DB.Save(&c).Error; err != nil {
		return nil, err
	}
	return c.ToMap(), nil
}

func (s *Service) TestSMTP(id int) error {
	var c general.SmtpClient
	if err := s.DB.First(&c, id).Error; err != nil {
		return err
	}
	addr := fmt.Sprintf("%s:%d", c.SMTPHost, c.SMTPPort)
	auth := smtp.PlainAuth("", c.SMTPUsername, c.SMTPPassword, c.SMTPHost)

	conn, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("SMTP connection failed: %v", err)
	}
	defer conn.Close()

	if c.UseTLS {
		if err := conn.StartTLS(nil); err != nil {
			// TLS might not be supported, try auth anyway
		}
	}
	if err := conn.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %v", err)
	}
	return nil
}

// ══════════════════════════════════════════════════════════════════════════════
//  APPOINTMENT REASONS
// ══════════════════════════════════════════════════════════════════════════════

func (s *Service) ListAppointmentReasons() ([]map[string]interface{}, error) {
	var items []appointmentModel.ReasonsVisionProviderAppointment
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = it.ToMap()
	}
	return out, nil
}

func (s *Service) CreateAppointmentReason(reason string, description *string) (map[string]interface{}, error) {
	r := appointmentModel.ReasonsVisionProviderAppointment{Reason: reason, Description: description}
	if err := s.DB.Create(&r).Error; err != nil {
		return nil, err
	}
	return r.ToMap(), nil
}

func (s *Service) UpdateAppointmentReason(id int, data map[string]interface{}) (map[string]interface{}, error) {
	var r appointmentModel.ReasonsVisionProviderAppointment
	if err := s.DB.First(&r, id).Error; err != nil {
		return nil, err
	}
	if v, ok := data["reason"]; ok {
		r.Reason = fmt.Sprintf("%v", v)
	}
	if v, ok := data["description"]; ok {
		if v == nil {
			r.Description = nil
		} else {
			s := fmt.Sprintf("%v", v)
			r.Description = &s
		}
	}
	if err := s.DB.Save(&r).Error; err != nil {
		return nil, err
	}
	return r.ToMap(), nil
}

func (s *Service) DeleteAppointmentReason(id int) error {
	return s.DB.Delete(&appointmentModel.ReasonsVisionProviderAppointment{}, id).Error
}

// ══════════════════════════════════════════════════════════════════════════════
//  LOCATIONS (showcase)
// ══════════════════════════════════════════════════════════════════════════════

func (s *Service) GetShowcaseLocations() ([]map[string]interface{}, error) {
	var locs []location.Location
	if err := s.DB.Where("showcase = ? AND store_active = ?", true, true).Find(&locs).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(locs))
	for i, l := range locs {
		out[i] = map[string]interface{}{
			"location_id":   l.IDLocation,
			"location_name": l.FullName,
		}
	}
	return out, nil
}

// ══════════════════════════════════════════════════════════════════════════════
//  LOCATION APPOINTMENT SETTINGS
// ══════════════════════════════════════════════════════════════════════════════

func (s *Service) SetRequestAppointment(locationID int, enabled bool) (map[string]interface{}, error) {
	var setting location.LocationAppointmentSettings
	err := s.DB.Where("location_id = ?", locationID).First(&setting).Error
	if err != nil {
		setting = location.LocationAppointmentSettings{LocationID: locationID}
		s.DB.Create(&setting)
	}
	setting.RequestAppointmentEnabled = enabled
	s.DB.Save(&setting)

	var loc location.Location
	s.DB.First(&loc, locationID)

	return map[string]interface{}{
		"location_id":                  loc.IDLocation,
		"location_full_name":           loc.FullName,
		"short_name":                   loc.ShortName,
		"request_appointment_enabled":  setting.RequestAppointmentEnabled,
	}, nil
}

func (s *Service) GetRequestAppointmentSettings() ([]map[string]interface{}, error) {
	var locs []location.Location
	s.DB.Where("showcase = ?", true).Find(&locs)

	var result []map[string]interface{}
	for _, loc := range locs {
		var setting location.LocationAppointmentSettings
		enabled := false
		if err := s.DB.Where("location_id = ?", loc.IDLocation).First(&setting).Error; err == nil {
			enabled = setting.RequestAppointmentEnabled
		}
		result = append(result, map[string]interface{}{
			"location_id":                 loc.IDLocation,
			"location_full_name":          loc.FullName,
			"short_name":                  loc.ShortName,
			"request_appointment_enabled": enabled,
		})
	}
	return result, nil
}

func (s *Service) SetIntakeForm(locationID int, enabled bool) (map[string]interface{}, error) {
	var setting location.LocationAppointmentSettings
	err := s.DB.Where("location_id = ?", locationID).First(&setting).Error
	if err != nil {
		setting = location.LocationAppointmentSettings{LocationID: locationID}
		s.DB.Create(&setting)
	}
	setting.IntakeFormEnabled = enabled
	s.DB.Save(&setting)

	var loc location.Location
	s.DB.First(&loc, locationID)

	return map[string]interface{}{
		"location_id":          loc.IDLocation,
		"location_full_name":   loc.FullName,
		"short_name":           loc.ShortName,
		"intake_form_enabled":  setting.IntakeFormEnabled,
	}, nil
}

func (s *Service) GetIntakeFormSettings() ([]map[string]interface{}, error) {
	var locs []location.Location
	s.DB.Where("showcase = ?", true).Find(&locs)

	var result []map[string]interface{}
	for _, loc := range locs {
		var setting location.LocationAppointmentSettings
		enabled := false
		if err := s.DB.Where("location_id = ?", loc.IDLocation).First(&setting).Error; err == nil {
			enabled = setting.IntakeFormEnabled
		}
		result = append(result, map[string]interface{}{
			"location_id":         loc.IDLocation,
			"location_full_name":  loc.FullName,
			"short_name":          loc.ShortName,
			"intake_form_enabled": enabled,
		})
	}
	return result, nil
}

func (s *Service) SetAppointmentDuration(duration int) (int, error) {
	var locs []location.Location
	s.DB.Where("showcase = ?", true).Find(&locs)
	for _, loc := range locs {
		var setting location.LocationAppointmentSettings
		err := s.DB.Where("location_id = ?", loc.IDLocation).First(&setting).Error
		if err != nil {
			setting = location.LocationAppointmentSettings{LocationID: loc.IDLocation}
			s.DB.Create(&setting)
		}
		setting.AppointmentDuration = &duration
		s.DB.Save(&setting)
	}
	return duration, nil
}

func (s *Service) GetAppointmentDuration() (*int, error) {
	var loc location.Location
	if err := s.DB.Where("showcase = ?", true).First(&loc).Error; err != nil {
		return nil, fmt.Errorf("No appointment duration found")
	}
	var setting location.LocationAppointmentSettings
	if err := s.DB.Where("location_id = ?", loc.IDLocation).First(&setting).Error; err != nil {
		return nil, fmt.Errorf("No appointment duration found")
	}
	return setting.AppointmentDuration, nil
}

// ══════════════════════════════════════════════════════════════════════════════
//  LENS TYPES
// ══════════════════════════════════════════════════════════════════════════════

func (s *Service) ListLensTypes() ([]map[string]interface{}, error) {
	var items []lenses.LensType
	if err := s.DB.Select("DISTINCT id_lens_type, type_name").Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = it.ToMap()
	}
	return out, nil
}

func (s *Service) CreateLensType(typeName, desc string) (map[string]interface{}, error) {
	it := lenses.LensType{TypeName: typeName, Description: desc}
	if err := s.DB.Create(&it).Error; err != nil {
		return nil, err
	}
	return it.ToMap(), nil
}

func (s *Service) UpdateLensType(id int, data map[string]interface{}) (map[string]interface{}, error) {
	var it lenses.LensType
	if err := s.DB.First(&it, id).Error; err != nil {
		return nil, err
	}
	if v, ok := data["type_name"]; ok {
		it.TypeName = fmt.Sprintf("%v", v)
	}
	if v, ok := data["description"]; ok {
		it.Description = fmt.Sprintf("%v", v)
	}
	s.DB.Save(&it)
	return it.ToMap(), nil
}

func (s *Service) DeleteLensType(id int) error {
	// check if used by lenses
	var count int64
	s.DB.Model(&lenses.Lenses{}).Where("lens_type_id = ?", id).Count(&count)
	if count > 0 {
		return fmt.Errorf("Cannot delete: used by %d lenses", count)
	}
	return s.DB.Delete(&lenses.LensType{}, id).Error
}

// ══════════════════════════════════════════════════════════════════════════════
//  GENERIC CRUD for simple tables
// ══════════════════════════════════════════════════════════════════════════════

// LensMaterials
func (s *Service) ListLensMaterials() ([]map[string]interface{}, error) {
	rows, err := s.DB.Raw(`SELECT id_lenses_materials, material_name, COALESCE("index",0), COALESCE(description,'') FROM lenses_materials ORDER BY material_name`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]interface{}
	for rows.Next() {
		var id int64
		var name string
		var idx float64
		var desc string
		rows.Scan(&id, &name, &idx, &desc)
		out = append(out, map[string]interface{}{"id_lenses_materials": id, "material_name": name, "index": idx, "description": desc})
	}
	return out, nil
}

func (s *Service) CreateLensMaterial(name string, idx *float64, desc *string) (map[string]interface{}, error) {
	it := lenses.LensesMaterial{MaterialName: name, Index: idx, Description: desc}
	if err := s.DB.Create(&it).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"id_lenses_materials": it.IDLensesMaterials, "material_name": it.MaterialName, "index": it.Index, "description": it.Description}, nil
}

func (s *Service) UpdateLensMaterial(id int, data map[string]interface{}) error {
	delete(data, "id_lens_material")
	delete(data, "id_lenses_materials")
	return s.DB.Model(&lenses.LensesMaterial{}).Where("id_lenses_materials = ?", id).Updates(data).Error
}

func (s *Service) DeleteLensMaterial(id int) error {
	var count int64
	s.DB.Model(&lenses.Lenses{}).Where("lenses_materials_id = ?", id).Count(&count)
	if count > 0 {
		return fmt.Errorf("Cannot delete: used by %d lenses", count)
	}
	return s.DB.Delete(&lenses.LensesMaterial{}, id).Error
}

// LensSpecialFeatures
func (s *Service) ListLensSpecialFeatures() ([]map[string]interface{}, error) {
	var items []lenses.LensSpecialFeature
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = map[string]interface{}{"id_lens_special_features": it.IDLensSpecialFeatures, "feature_name": it.FeatureName, "description": it.Description}
	}
	return out, nil
}

func (s *Service) CreateLensSpecialFeature(name, desc string) (map[string]interface{}, error) {
	it := lenses.LensSpecialFeature{FeatureName: name, Description: desc}
	if err := s.DB.Create(&it).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"id_lens_special_features": it.IDLensSpecialFeatures, "feature_name": it.FeatureName, "description": it.Description}, nil
}

func (s *Service) UpdateLensSpecialFeature(id int, data map[string]interface{}) error {
	return s.DB.Model(&lenses.LensSpecialFeature{}).Where("id_lens_special_features = ?", id).Updates(data).Error
}

func (s *Service) DeleteLensSpecialFeature(id int) error {
	return s.DB.Delete(&lenses.LensSpecialFeature{}, id).Error
}

// LensSeries
func (s *Service) ListLensSeries() ([]map[string]interface{}, error) {
	var items []lenses.LensSeries
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = map[string]interface{}{"id_lens_series": it.IDLensSeries, "series_name": it.SeriesName, "description": it.Description}
	}
	return out, nil
}

func (s *Service) CreateLensSeries(name, desc string) (map[string]interface{}, error) {
	it := lenses.LensSeries{SeriesName: name, Description: desc}
	if err := s.DB.Create(&it).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"id_lens_series": it.IDLensSeries, "series_name": it.SeriesName, "description": it.Description}, nil
}

func (s *Service) UpdateLensSeries(id int, data map[string]interface{}) error {
	return s.DB.Model(&lenses.LensSeries{}).Where("id_lens_series = ?", id).Updates(data).Error
}

func (s *Service) DeleteLensSeries(id int) error {
	var count int64
	s.DB.Model(&lenses.Lenses{}).Where("lens_series_id = ?", id).Count(&count)
	if count > 0 {
		return fmt.Errorf("Cannot delete: used by %d lenses", count)
	}
	return s.DB.Delete(&lenses.LensSeries{}, id).Error
}

// VCodes
func (s *Service) ListVCodes() ([]map[string]interface{}, error) {
	var items []lenses.VCodesLens
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = map[string]interface{}{"id_v_codes_lens": it.IDVCodesLens, "code": it.Code, "description": it.Description}
	}
	return out, nil
}

func (s *Service) CreateVCode(code, desc string) (map[string]interface{}, error) {
	it := lenses.VCodesLens{Code: code, Description: desc}
	if err := s.DB.Create(&it).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"id_v_codes_lens": it.IDVCodesLens, "code": it.Code, "description": it.Description}, nil
}

func (s *Service) UpdateVCode(id int, data map[string]interface{}) error {
	return s.DB.Model(&lenses.VCodesLens{}).Where("id_v_codes_lens = ?", id).Updates(data).Error
}

func (s *Service) DeleteVCode(id int) error {
	return s.DB.Delete(&lenses.VCodesLens{}, id).Error
}

// LensStyle
func (s *Service) ListLensStyles() ([]map[string]interface{}, error) {
	var items []lenses.LensStyle
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = map[string]interface{}{"id_lens_style": it.IDLensStyle, "style_name": it.StyleName, "description": it.Description}
	}
	return out, nil
}

func (s *Service) CreateLensStyle(name, desc string) (map[string]interface{}, error) {
	it := lenses.LensStyle{StyleName: name, Description: desc}
	if err := s.DB.Create(&it).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"id_lens_style": it.IDLensStyle, "style_name": it.StyleName, "description": it.Description}, nil
}

func (s *Service) UpdateLensStyle(id int, data map[string]interface{}) error {
	return s.DB.Model(&lenses.LensStyle{}).Where("id_lens_style = ?", id).Updates(data).Error
}

func (s *Service) DeleteLensStyle(id int) error {
	return s.DB.Delete(&lenses.LensStyle{}, id).Error
}

// LensTintColor
func (s *Service) ListLensTintColors() ([]map[string]interface{}, error) {
	var items []lenses.LensTintColor
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = map[string]interface{}{"id_lens_tint_color": it.IDLensTintColor, "lens_tint_color_name": it.LensTintColorName}
	}
	return out, nil
}

func (s *Service) CreateLensTintColor(name string) (map[string]interface{}, error) {
	it := lenses.LensTintColor{LensTintColorName: name}
	if err := s.DB.Create(&it).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"id_lens_tint_color": it.IDLensTintColor, "lens_tint_color_name": it.LensTintColorName}, nil
}

func (s *Service) UpdateLensTintColor(id int, name string) error {
	return s.DB.Model(&lenses.LensTintColor{}).Where("id_lens_tint_color = ?", id).Update("lens_tint_color_name", name).Error
}

func (s *Service) DeleteLensTintColor(id int) error {
	return s.DB.Delete(&lenses.LensTintColor{}, id).Error
}

// LensSampleColor
func (s *Service) ListLensSampleColors() ([]map[string]interface{}, error) {
	var items []lenses.LensSampleColor
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = map[string]interface{}{"id_lens_sample_color": it.IDLensSampleColor, "lens_sample_color_name": it.LensSampleColorName}
	}
	return out, nil
}

func (s *Service) CreateLensSampleColor(name string) (map[string]interface{}, error) {
	it := lenses.LensSampleColor{LensSampleColorName: name}
	if err := s.DB.Create(&it).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"id_lens_sample_color": it.IDLensSampleColor, "lens_sample_color_name": it.LensSampleColorName}, nil
}

func (s *Service) UpdateLensSampleColor(id int, name string) error {
	return s.DB.Model(&lenses.LensSampleColor{}).Where("id_lens_sample_color = ?", id).Update("lens_sample_color_name", name).Error
}

func (s *Service) DeleteLensSampleColor(id int) error {
	return s.DB.Delete(&lenses.LensSampleColor{}, id).Error
}

// LensSafetyThickness
func (s *Service) ListLensSafetyThickness() ([]map[string]interface{}, error) {
	var items []lenses.LensSafetyThickness
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = map[string]interface{}{"id_lens_safety_thickness": it.IDLensSafetyThickness, "safety_thickness_name": it.SafetyThicknessName, "description": it.Description}
	}
	return out, nil
}

func (s *Service) CreateLensSafetyThickness(name, desc string) (map[string]interface{}, error) {
	it := lenses.LensSafetyThickness{SafetyThicknessName: name, Description: desc}
	if err := s.DB.Create(&it).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"id_lens_safety_thickness": it.IDLensSafetyThickness, "safety_thickness_name": it.SafetyThicknessName, "description": it.Description}, nil
}

func (s *Service) UpdateLensSafetyThickness(id int, data map[string]interface{}) error {
	return s.DB.Model(&lenses.LensSafetyThickness{}).Where("id_lens_safety_thickness = ?", id).Updates(data).Error
}

func (s *Service) DeleteLensSafetyThickness(id int) error {
	return s.DB.Delete(&lenses.LensSafetyThickness{}, id).Error
}

// LensBevel
func (s *Service) ListLensBevels() ([]map[string]interface{}, error) {
	var items []lenses.LensBevel
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = map[string]interface{}{"id_lens_bevel": it.IDLensBevel, "lens_bevel_name": it.LensBevelName, "description": it.Description}
	}
	return out, nil
}

func (s *Service) CreateLensBevel(name, desc string) (map[string]interface{}, error) {
	it := lenses.LensBevel{LensBevelName: name, Description: desc}
	if err := s.DB.Create(&it).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"id_lens_bevel": it.IDLensBevel, "lens_bevel_name": it.LensBevelName, "description": it.Description}, nil
}

func (s *Service) UpdateLensBevel(id int, data map[string]interface{}) error {
	return s.DB.Model(&lenses.LensBevel{}).Where("id_lens_bevel = ?", id).Updates(data).Error
}

func (s *Service) DeleteLensBevel(id int) error {
	return s.DB.Delete(&lenses.LensBevel{}, id).Error
}

// ══════════════════════════════════════════════════════════════════════════════
//  FRAME SHAPES
// ══════════════════════════════════════════════════════════════════════════════

func (s *Service) ListFrameShapes() ([]map[string]interface{}, error) {
	var items []frames.FrameShape
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = it.ToMap()
	}
	return out, nil
}

func (s *Service) CreateFrameShape(title string, desc *string) (map[string]interface{}, error) {
	it := frames.FrameShape{TitleFrameShape: title, Description: desc}
	if err := s.DB.Create(&it).Error; err != nil {
		return nil, err
	}
	return it.ToMap(), nil
}

func (s *Service) UpdateFrameShape(id int, data map[string]interface{}) error {
	res := s.DB.Model(&frames.FrameShape{}).Where("id_frame_shape = ?", id).Updates(data)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("not found")
	}
	return nil
}

func (s *Service) DeleteFrameShape(id int) error {
	return s.DB.Delete(&frames.FrameShape{}, id).Error
}

// FrameTypeMaterials
func (s *Service) ListFrameTypeMaterials() ([]map[string]interface{}, error) {
	var items []frames.FrameTypeMaterial
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = it.ToMap()
	}
	return out, nil
}

func (s *Service) CreateFrameTypeMaterial(material string) (map[string]interface{}, error) {
	it := frames.FrameTypeMaterial{Material: material}
	if err := s.DB.Create(&it).Error; err != nil {
		return nil, err
	}
	return it.ToMap(), nil
}

func (s *Service) UpdateFrameTypeMaterial(id int, material string) error {
	res := s.DB.Model(&frames.FrameTypeMaterial{}).Where("id_frame_type_material = ?", id).Update("material", material)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("not found")
	}
	return nil
}

func (s *Service) DeleteFrameTypeMaterial(id int) error {
	return s.DB.Delete(&frames.FrameTypeMaterial{}, id).Error
}

// ══════════════════════════════════════════════════════════════════════════════
//  PROFESSIONAL SERVICE SCOPES
// ══════════════════════════════════════════════════════════════════════════════

func (s *Service) ListProfServiceScopes() ([]map[string]interface{}, error) {
	var items []serviceModel.ProfessionalServiceScope
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = map[string]interface{}{"id_professional_service_scope": it.IDProfessionalServiceScope, "title": it.Title}
	}
	return out, nil
}

func (s *Service) CreateProfServiceScope(title string) (map[string]interface{}, error) {
	it := serviceModel.ProfessionalServiceScope{Title: title}
	if err := s.DB.Create(&it).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"id_professional_service_scope": it.IDProfessionalServiceScope, "title": it.Title}, nil
}

func (s *Service) UpdateProfServiceScope(id int, title string) error {
	return s.DB.Model(&serviceModel.ProfessionalServiceScope{}).Where("id_professional_service_scope = ?", id).Update("title", title).Error
}

func (s *Service) DeleteProfServiceScope(id int) error {
	return s.DB.Delete(&serviceModel.ProfessionalServiceScope{}, id).Error
}

// AdditionalServiceTypes
func (s *Service) ListAddServiceTypes() ([]map[string]interface{}, error) {
	var items []serviceModel.AdditionalServiceType
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = map[string]interface{}{"id_add_service_type": it.IDAddServiceType, "title": it.Title}
	}
	return out, nil
}

func (s *Service) CreateAddServiceType(title string) (map[string]interface{}, error) {
	it := serviceModel.AdditionalServiceType{Title: title}
	if err := s.DB.Create(&it).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"id_add_service_type": it.IDAddServiceType, "title": it.Title}, nil
}

func (s *Service) UpdateAddServiceType(id int, title string) error {
	res := s.DB.Model(&serviceModel.AdditionalServiceType{}).Where("id_add_service_type = ?", id).Update("title", title)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("not found")
	}
	return nil
}

func (s *Service) DeleteAddServiceType(id int) error {
	return s.DB.Delete(&serviceModel.AdditionalServiceType{}, id).Error
}

// ══════════════════════════════════════════════════════════════════════════════
//  INSURANCE
// ══════════════════════════════════════════════════════════════════════════════

func (s *Service) ListInsuranceCompanies() ([]map[string]interface{}, error) {
	rows, err := s.DB.Raw(`
		SELECT ic.id_insurance_company, ic.company_name, ic.id_insurance_coverage_type,
		       ict.coverage_name
		FROM insurance_company ic
		LEFT JOIN insurance_coverage_types ict ON ic.id_insurance_coverage_type = ict.id_insurance_coverage_type
		ORDER BY ic.company_name`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		var covID *int
		var covName *string
		rows.Scan(&id, &name, &covID, &covName)
		out = append(out, map[string]interface{}{
			"insurance_company_id":        id,
			"company_name":               name,
			"id_insurance_coverage_type":  covID,
			"coverage_name":              covName,
		})
	}
	return out, nil
}

func (s *Service) ListCoverageTypes() ([]map[string]interface{}, error) {
	var items []insurance.InsuranceCoverageType
	if err := s.DB.Order("coverage_name").Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = map[string]interface{}{
			"insurance_coverage_type_id": it.IDInsuranceCoverageType,
			"coverage_type_name":         it.CoverageName,
		}
	}
	return out, nil
}

func (s *Service) CreateInsuranceCompany(name string, coverageID *int) (map[string]interface{}, error) {
	if coverageID != nil {
		var exists bool
		s.DB.Raw(`SELECT EXISTS(SELECT 1 FROM insurance_coverage_types WHERE id_insurance_coverage_type = ?)`, *coverageID).Row().Scan(&exists)
		if !exists {
			return nil, fmt.Errorf("Invalid 'insurance_coverage_type_id'")
		}
	}

	// dedup
	q := s.DB.Model(&insurance.InsuranceCompany{}).Where("LOWER(company_name) = LOWER(?)", name)
	if coverageID == nil {
		q = q.Where("id_insurance_coverage_type IS NULL")
	} else {
		q = q.Where("id_insurance_coverage_type = ?", *coverageID)
	}
	var count int64
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("Insurance company already exists")
	}

	c := insurance.InsuranceCompany{CompanyName: name, IDInsuranceCoverageType: coverageID}
	if err := s.DB.Create(&c).Error; err != nil {
		return nil, err
	}
	resp := map[string]interface{}{
		"message":              "Insurance company added successfully",
		"insurance_company_id": c.IDInsuranceCompany,
	}
	if coverageID != nil {
		resp["insurance_coverage_type_id"] = *coverageID
	}
	return resp, nil
}

func (s *Service) UpdateInsuranceCompany(id int, data map[string]interface{}) (map[string]interface{}, error) {
	var c insurance.InsuranceCompany
	if err := s.DB.First(&c, id).Error; err != nil {
		return nil, err
	}

	newName := c.CompanyName
	if v, ok := data["company_name"]; ok {
		newName = strings.TrimSpace(fmt.Sprintf("%v", v))
	}

	newCovID := c.IDInsuranceCoverageType
	for _, key := range []string{"id_insurance_coverage_type", "insurance_coverage_type_id", "id_type_insurance_policy"} {
		if v, ok := data[key]; ok && v != nil {
			covID := toInt(v, 0)
			if covID > 0 {
				newCovID = &covID
			} else {
				newCovID = nil
			}
			break
		}
	}

	// dedup
	q := s.DB.Model(&insurance.InsuranceCompany{}).
		Where("id_insurance_company != ? AND LOWER(company_name) = LOWER(?)", id, newName)
	if newCovID == nil {
		q = q.Where("id_insurance_coverage_type IS NULL")
	} else {
		q = q.Where("id_insurance_coverage_type = ?", *newCovID)
	}
	var count int64
	q.Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("Insurance company already exists")
	}

	c.CompanyName = newName
	c.IDInsuranceCoverageType = newCovID

	for _, f := range []string{"contact_number", "contact_email", "address", "address_line_2", "city", "state", "zip_code"} {
		if v, ok := data[f]; ok {
			str := fmt.Sprintf("%v", v)
			switch f {
			case "contact_number":
				c.ContactNumber = &str
			case "contact_email":
				c.ContactEmail = &str
			case "address":
				c.Address = &str
			case "address_line_2":
				c.AddressLine2 = &str
			case "city":
				c.City = &str
			case "state":
				c.State = &str
			case "zip_code":
				c.ZipCode = &str
			}
		}
	}

	s.DB.Save(&c)

	var covName *string
	if c.IDInsuranceCoverageType != nil {
		var cn string
		s.DB.Raw(`SELECT coverage_name FROM insurance_coverage_types WHERE id_insurance_coverage_type = ?`, *c.IDInsuranceCoverageType).Row().Scan(&cn)
		covName = &cn
	}

	return map[string]interface{}{
		"insurance_company_id":       c.IDInsuranceCompany,
		"company_name":              c.CompanyName,
		"id_insurance_coverage_type": c.IDInsuranceCoverageType,
		"coverage_name":             covName,
	}, nil
}

func (s *Service) DeleteInsuranceCompany(id int) error {
	var count int64
	s.DB.Raw(`SELECT COUNT(*) FROM insurance_policy WHERE insurance_company_id = ?`, id).Row().Scan(&count)
	if count > 0 {
		return fmt.Errorf("Cannot delete. Insurance company is used in existing insurance policies.")
	}
	return s.DB.Delete(&insurance.InsuranceCompany{}, id).Error
}

func (s *Service) ListInsuranceTypes() ([]map[string]interface{}, error) {
	var items []insurance.InsuranceCoverageType
	if err := s.DB.Order("coverage_name").Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = map[string]interface{}{
			"id_type_insurance_policy": it.IDInsuranceCoverageType,
			"type_name":               it.CoverageName,
		}
	}
	return out, nil
}

// InsurancePaymentTypes
func (s *Service) ListInsurancePaymentTypes() ([]map[string]interface{}, error) {
	var items []insurance.InsurancePaymentType
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = it.ToMap()
	}
	return out, nil
}

func (s *Service) CreateInsurancePaymentType(name string, desc *string) (map[string]interface{}, error) {
	it := insurance.InsurancePaymentType{Name: name, Description: desc}
	if err := s.DB.Create(&it).Error; err != nil {
		return nil, err
	}
	return it.ToMap(), nil
}

func (s *Service) UpdateInsurancePaymentType(id int, data map[string]interface{}) error {
	return s.DB.Model(&insurance.InsurancePaymentType{}).Where("id_insurance_payment_type = ?", id).Updates(data).Error
}

func (s *Service) DeleteInsurancePaymentType(id int) error {
	return s.DB.Delete(&insurance.InsurancePaymentType{}, id).Error
}

// ══════════════════════════════════════════════════════════════════════════════
//  VENDOR BRANDS
// ══════════════════════════════════════════════════════════════════════════════

func (s *Service) ListVendorBrands() ([]map[string]interface{}, error) {
	query := `
		SELECT v.id_vendor, v.vendor_name, 'frames' AS brand_type, b.id_brand AS brand_id, b.brand_name
		FROM vendor_brand vb
		JOIN vendor v ON v.id_vendor = vb.id_vendor
		JOIN brand b ON b.id_brand = vb.id_brand
		UNION ALL
		SELECT v.id_vendor, v.vendor_name, 'lens', bl.id_brand_lens, bl.brand_name
		FROM vendor_brand_lens vbl
		JOIN vendor v ON v.id_vendor = vbl.id_vendor
		JOIN brand_lens bl ON bl.id_brand_lens = vbl.id_brand_lens
		UNION ALL
		SELECT v.id_vendor, v.vendor_name, 'contact_lens', bcl.id_brand_contact_lens, bcl.brand_name
		FROM vendor_brand_contact_lens vbcl
		JOIN vendor v ON v.id_vendor = vbcl.id_vendor
		JOIN brand_contact_lens bcl ON bcl.id_brand_contact_lens = vbcl.id_brand_contact_lens
		ORDER BY vendor_name, brand_type, brand_name`

	rows, err := s.DB.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]interface{}
	for rows.Next() {
		var vendorID, brandID int
		var vendorName, brandType, brandName string
		rows.Scan(&vendorID, &vendorName, &brandType, &brandID, &brandName)
		out = append(out, map[string]interface{}{
			"vendor_id":   vendorID,
			"vendor_name": vendorName,
			"brand_type":  brandType,
			"brand_id":    brandID,
			"brand_name":  brandName,
		})
	}
	return out, nil
}

// ══════════════════════════════════════════════════════════════════════════════
//  TICKET STATUSES
// ══════════════════════════════════════════════════════════════════════════════

func (s *Service) ListTicketStatuses() ([]map[string]interface{}, error) {
	var items []lab_ticket.LabTicketStatus
	if err := s.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = it.ToMap()
	}
	return out, nil
}

func (s *Service) CreateTicketStatus(status string) (map[string]interface{}, error) {
	var count int64
	s.DB.Model(&lab_ticket.LabTicketStatus{}).Where("ticket_status = ?", status).Count(&count)
	if count > 0 {
		return nil, fmt.Errorf("Status already exists")
	}
	it := lab_ticket.LabTicketStatus{TicketStatus: status}
	if err := s.DB.Create(&it).Error; err != nil {
		return nil, err
	}
	return it.ToMap(), nil
}

func (s *Service) UpdateTicketStatus(id int, status string) error {
	return s.DB.Model(&lab_ticket.LabTicketStatus{}).Where("id_lab_ticket_status = ?", id).Update("ticket_status", status).Error
}

func (s *Service) DeleteTicketStatus(id int) error {
	return s.DB.Delete(&lab_ticket.LabTicketStatus{}, id).Error
}

// ══════════════════════════════════════════════════════════════════════════════
//  LENS STATUSES — raw SQL since LensStatus model may not exist
// ══════════════════════════════════════════════════════════════════════════════

func (s *Service) ListLensStatuses() ([]map[string]interface{}, error) {
	rows, err := s.DB.Raw(`SELECT id_lens_status, status_name FROM lens_status ORDER BY id_lens_status`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		out = append(out, map[string]interface{}{"id_lens_status": id, "status_name": name})
	}
	return out, nil
}

func (s *Service) CreateLensStatus(name string) (map[string]interface{}, error) {
	var count int64
	s.DB.Raw(`SELECT COUNT(*) FROM lens_status WHERE status_name = ?`, name).Row().Scan(&count)
	if count > 0 {
		return nil, fmt.Errorf("Status already exists")
	}
	s.DB.Exec(`INSERT INTO lens_status (status_name) VALUES (?)`, name)
	var id int
	s.DB.Raw(`SELECT id_lens_status FROM lens_status WHERE status_name = ?`, name).Row().Scan(&id)
	return map[string]interface{}{"id_lens_status": id, "status_name": name}, nil
}

func (s *Service) UpdateLensStatus(id int, name string) error {
	return s.DB.Exec(`UPDATE lens_status SET status_name = ? WHERE id_lens_status = ?`, name, id).Error
}

func (s *Service) DeleteLensStatus(id int) error {
	return s.DB.Exec(`DELETE FROM lens_status WHERE id_lens_status = ?`, id).Error
}

// ══════════════════════════════════════════════════════════════════════════════
//  NOTIFY SETTINGS
// ══════════════════════════════════════════════════════════════════════════════

func (s *Service) GetNotifySetting(actionName string) (map[string]interface{}, error) {
	var it notifications.NotifySetting
	if err := s.DB.Where("action_name = ?", actionName).First(&it).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"id": it.ID, "action_name": it.ActionName, "email": it.Email, "sms": it.SMS}, nil
}

func (s *Service) UpsertNotifySetting(actionName string, data map[string]interface{}) (map[string]interface{}, error) {
	var it notifications.NotifySetting
	err := s.DB.Where("action_name = ?", actionName).First(&it).Error
	if err != nil {
		it = notifications.NotifySetting{ActionName: actionName, Email: true, SMS: true}
		s.DB.Create(&it)
	}
	if v, ok := data["email"]; ok {
		it.Email = toBool(v, it.Email)
	}
	if v, ok := data["sms"]; ok {
		it.SMS = toBool(v, it.SMS)
	}
	s.DB.Save(&it)
	return map[string]interface{}{"id": it.ID, "action_name": it.ActionName, "email": it.Email, "sms": it.SMS}, nil
}

// ══════════════════════════════════════════════════════════════════════════════
//  PAYMENT METHODS
// ══════════════════════════════════════════════════════════════════════════════

var ValidCategories = map[string]bool{
	"card": true, "installment": true, "fintech": true, "cash": true, "check": true,
	"gift": true, "insurance": true, "credit": true, "internal": true, "other": true,
}

func (s *Service) ListPaymentMethods() ([]map[string]interface{}, error) {
	var items []general.PaymentMethod
	if err := s.DB.Order("id_payment_method").Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = it.ToMap()
	}
	return out, nil
}

func (s *Service) CreatePaymentMethod(methodName, category string, shortName *string) (map[string]interface{}, error) {
	if !ValidCategories[category] {
		return nil, fmt.Errorf("Invalid category")
	}
	it := general.PaymentMethod{MethodName: methodName, Category: &category, ShortName: shortName}
	if err := s.DB.Create(&it).Error; err != nil {
		return nil, err
	}
	return it.ToMap(), nil
}

func (s *Service) UpdatePaymentMethod(id int, data map[string]interface{}) (map[string]interface{}, error) {
	var it general.PaymentMethod
	if err := s.DB.First(&it, id).Error; err != nil {
		return nil, err
	}
	if it.IsSystem {
		return nil, fmt.Errorf("This payment method is system-level and cannot be modified")
	}
	if v, ok := data["method_name"]; ok {
		n := strings.TrimSpace(fmt.Sprintf("%v", v))
		if n == "" {
			return nil, fmt.Errorf("method_name cannot be empty")
		}
		it.MethodName = n
	}
	if v, ok := data["category"]; ok {
		c := strings.TrimSpace(fmt.Sprintf("%v", v))
		if !ValidCategories[c] {
			return nil, fmt.Errorf("Invalid category")
		}
		it.Category = &c
	}
	if v, ok := data["short_name"]; ok {
		if v == nil {
			it.ShortName = nil
		} else {
			sn := strings.TrimSpace(fmt.Sprintf("%v", v))
			if sn == "" {
				it.ShortName = nil
			} else {
				it.ShortName = &sn
			}
		}
	}
	s.DB.Save(&it)
	return it.ToMap(), nil
}

func (s *Service) DeletePaymentMethod(id int) error {
	var it general.PaymentMethod
	if err := s.DB.First(&it, id).Error; err != nil {
		return err
	}
	if it.IsSystem {
		return fmt.Errorf("This payment method is system-level and cannot be deleted")
	}
	return s.DB.Delete(&general.PaymentMethod{}, id).Error
}

// ══════════════════════════════════════════════════════════════════════════════
//  JOB TITLES
// ══════════════════════════════════════════════════════════════════════════════

const DoctorJobTitleID = 1

func (s *Service) ListJobTitles() ([]map[string]interface{}, error) {
	var items []employees.JobTitle
	if err := s.DB.Order("title").Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(items))
	for i, it := range items {
		out[i] = it.ToMap()
	}
	return out, nil
}

func (s *Service) CreateJobTitle(title string, shortTitle *string) (map[string]interface{}, error) {
	it := employees.JobTitle{Title: title, Doctor: false, ShortTitle: shortTitle}
	if err := s.DB.Create(&it).Error; err != nil {
		return nil, err
	}
	return it.ToMap(), nil
}

func (s *Service) UpdateJobTitle(id int, title string, shortTitle *string) (map[string]interface{}, error) {
	var it employees.JobTitle
	if err := s.DB.First(&it, id).Error; err != nil {
		return nil, err
	}
	it.Title = title
	it.ShortTitle = shortTitle
	s.DB.Save(&it)
	return it.ToMap(), nil
}

func (s *Service) DeleteJobTitle(id int) error {
	if id == DoctorJobTitleID {
		return fmt.Errorf("Doctor job title cannot be deleted")
	}
	return s.DB.Delete(&employees.JobTitle{}, id).Error
}

// ── helpers ─────────────────────────────────────────────────────────────────

func toInt(v interface{}, def int) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case string:
		var i int
		fmt.Sscanf(n, "%d", &i)
		return i
	default:
		return def
	}
}

func toBool(v interface{}, def bool) bool {
	switch b := v.(type) {
	case bool:
		return b
	case float64:
		return b != 0
	case string:
		return b == "true" || b == "1"
	default:
		return def
	}
}
