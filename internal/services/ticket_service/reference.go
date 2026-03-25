package ticket_service

import (
	"fmt"

	frameModel "sighthub-backend/internal/models/frames"
	labTicketModel "sighthub-backend/internal/models/lab_ticket"
	lensModel "sighthub-backend/internal/models/lenses"
	vendorModel "sighthub-backend/internal/models/vendors"
)

// GET /statuses
func (s *Service) GetStatuses() ([]map[string]interface{}, error) {
	var rows []labTicketModel.LabTicketStatus
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"lab_ticket_status_id": r.IDLabTicketStatus,
			"ticket_status":        r.TicketStatus,
		})
	}
	return result, nil
}

// GET /lens-statuses — raw query since Go model may not exist
func (s *Service) GetLensStatuses() ([]map[string]interface{}, error) {
	type row struct {
		IDLensStatus int    `gorm:"column:id_lens_status"`
		StatusName   string `gorm:"column:status_name"`
	}
	var rows []row
	if err := s.db.Table("lens_status").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"id_lens_status": r.IDLensStatus,
			"status_name":    r.StatusName,
		})
	}
	return result, nil
}

// GET /labs
func (s *Service) GetLabs() ([]map[string]interface{}, error) {
	var rows []vendorModel.Vendor
	if err := s.db.Where("lab = true AND visible = true").Order("vendor_name").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"lab_id":    r.IDVendor,
			"title_lab": r.VendorName,
		})
	}
	return result, nil
}

func (s *Service) GetLabsForEmployee(username string) ([]map[string]interface{}, error) {
	_, loc, err := s.empLocation(username)
	if err != nil || loc == nil {
		// fallback: return labs without account info
		return s.GetLabs()
	}
	locationID := loc.IDLocation

	var rows []vendorModel.Vendor
	if err := s.db.Where("lab = true AND visible = true").Order("vendor_name").Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		entry := map[string]interface{}{
			"lab_id":    r.IDVendor,
			"title_lab": r.VendorName,
		}

		type vlaRow struct {
			AccountNumber *string `gorm:"column:account_number"`
			VwSloID       *int    `gorm:"column:vw_slo_id"`
			VwBill        *string `gorm:"column:vw_bill"`
			VwShip        *string `gorm:"column:vw_ship"`
			Source        *string `gorm:"column:source"`
		}
		var vla vlaRow
		s.db.Table("vendor_location_account").
			Where("vendor_id = ? AND location_id = ?", r.IDVendor, locationID).
			Order("created_at DESC").
			First(&vla)
		entry["account_number"] = vla.AccountNumber
		entry["vw_slo_id"] = vla.VwSloID
		entry["vw_bill"] = vla.VwBill
		entry["vw_ship"] = vla.VwShip
		entry["source"] = vla.Source

		result = append(result, entry)
	}
	return result, nil
}

// GET /frame-type-materials
func (s *Service) GetFrameTypeMaterials() ([]map[string]interface{}, error) {
	var rows []frameModel.FrameTypeMaterial
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"frame_type_material_id": r.IDFrameTypeMaterial,
			"material":               r.Material,
		})
	}
	return result, nil
}

// GET /frame_shapes
func (s *Service) GetFrameShapes() ([]map[string]interface{}, error) {
	var rows []frameModel.FrameShape
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"frame_shape_id":   r.IDFrameShape,
			"frame_shape_name": r.TitleFrameShape,
		})
	}
	return result, nil
}

// GET /lens/type
func (s *Service) GetLensTypes() ([]map[string]interface{}, error) {
	var rows []lensModel.LensType
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"id_lens_type": r.IDLensType,
			"type_name":    r.TypeName,
		})
	}
	return result, nil
}

// GET /lens/materials
func (s *Service) GetLensMaterials() ([]map[string]interface{}, error) {
	var rows []lensModel.LensesMaterial
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"id_lenses_materials": r.IDLensesMaterials,
			"material_name":      r.MaterialName,
		})
	}
	return result, nil
}

// POST /lens/materials
func (s *Service) AddLensMaterial(materialName string) (map[string]interface{}, error) {
	if materialName == "" {
		return nil, ErrMaterialNameRequired
	}
	mat := lensModel.LensesMaterial{MaterialName: materialName}
	if err := s.db.Create(&mat).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message":             "Lens material added successfully",
		"id_lenses_materials": mat.IDLensesMaterials,
	}, nil
}

// GET /lens/style
func (s *Service) GetLensStyles() ([]map[string]interface{}, error) {
	var rows []lensModel.LensStyle
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"id_lens_style": r.IDLensStyle,
			"style_name":    r.StyleName,
		})
	}
	return result, nil
}

// GET /lens/lens_tint_color
func (s *Service) GetLensTintColors() ([]map[string]interface{}, error) {
	var rows []lensModel.LensTintColor
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"id_lens_tint_color":   r.IDLensTintColor,
			"lens_tint_color_name": r.LensTintColorName,
		})
	}
	return result, nil
}

// GET /lens/lens_sample_color
func (s *Service) GetLensSampleColors() ([]map[string]interface{}, error) {
	var rows []lensModel.LensSampleColor
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"id_lens_sample_color":   r.IDLensSampleColor,
			"lens_sample_color_name": r.LensSampleColorName,
		})
	}
	return result, nil
}

// GET /lens/safety_thickness
func (s *Service) GetLensSafetyThicknesses() ([]map[string]interface{}, error) {
	var rows []lensModel.LensSafetyThickness
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"id_lens_safety_thickness": r.IDLensSafetyThickness,
			"safety_thickness_name":    r.SafetyThicknessName,
		})
	}
	return result, nil
}

// GET /lens/lens_bevel
func (s *Service) GetLensBevels() ([]map[string]interface{}, error) {
	var rows []lensModel.LensBevel
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"id_lens_bevel":   r.IDLensBevel,
			"lens_bevel_name": r.LensBevelName,
		})
	}
	return result, nil
}

// GET /lens/lens_edge
func (s *Service) GetLensEdges() ([]map[string]interface{}, error) {
	var rows []lensModel.LensEdge
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"id_lens_edge":   r.IDLensEdge,
			"lens_edge_name": r.LensEdgeName,
		})
	}
	return result, nil
}

// GET /lens/series
func (s *Service) GetLensSeries() ([]map[string]interface{}, error) {
	var rows []lensModel.LensSeries
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"id_lens_series": r.IDLensSeries,
			"series_name":    r.SeriesName,
		})
	}
	return result, nil
}

// GET /contact/services
func (s *Service) GetContactServices() ([]map[string]interface{}, error) {
	var rows []labTicketModel.LabTicketContactLensService
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"lab_ticket_contact_lens_services_id": r.IDLabTicketContactLensServices,
			"services_name":                       r.ServicesName,
		})
	}
	return result, nil
}

// GET /contact_lens/brands
func (s *Service) GetContactLensBrands() ([]map[string]interface{}, error) {
	var rows []vendorModel.BrandContactLens
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"id_brand_contact_lens": r.IDBrandContactLens,
			"brand_name":            r.BrandName,
		})
	}
	return result, nil
}

// sentinel error
var ErrMaterialNameRequired = fmt.Errorf("field 'material_name' is required")

// keep import happy
var _ = fmt.Sprintf
