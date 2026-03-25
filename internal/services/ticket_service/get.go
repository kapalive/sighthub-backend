package ticket_service

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	employeeModel "sighthub-backend/internal/models/employees"
	frameModel "sighthub-backend/internal/models/frames"
	inventoryModel "sighthub-backend/internal/models/inventory"
	invoiceModel "sighthub-backend/internal/models/invoices"
	labTicketModel "sighthub-backend/internal/models/lab_ticket"
	lensModel "sighthub-backend/internal/models/lenses"
	patientModel "sighthub-backend/internal/models/patients"
	serviceModel "sighthub-backend/internal/models/service"
	vendorModel "sighthub-backend/internal/models/vendors"
)

// ── List tickets (GET /) ────────────────────────────────────────────────────

type ListTicketsParams struct {
	LocationID *string
	DateFrom   *string
	DateTo     *string
	StatusID   *string
}

func (s *Service) ListTickets(p ListTicketsParams) ([]map[string]interface{}, error) {
	q := s.db.Model(&labTicketModel.LabTicket{})

	if p.LocationID != nil && *p.LocationID != "" {
		q = q.Where("location_id = ?", *p.LocationID)
	}
	if p.DateFrom != nil && *p.DateFrom != "" && p.DateTo != nil && *p.DateTo != "" {
		q = q.Where("date_create >= ? AND date_create <= ?", *p.DateFrom, *p.DateTo)
	}
	if p.StatusID != nil && *p.StatusID != "" {
		q = q.Where("lab_ticket_status_id = ?", *p.StatusID)
	}

	var tickets []labTicketModel.LabTicket
	if err := q.Find(&tickets).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(tickets))
	for i := range tickets {
		result = append(result, tickets[i].ToMap())
	}
	return result, nil
}

// ── Tickets by invoice (GET /invoice/{invoice_id}) ──────────────────────────

func (s *Service) GetTicketsByInvoice(invoiceID int64) ([]map[string]interface{}, error) {
	var tickets []labTicketModel.LabTicket
	if err := s.db.Where("invoice_id = ?", invoiceID).Find(&tickets).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(tickets))
	for i := range tickets {
		t := &tickets[i]

		var empName *string
		var emp employeeModel.Employee
		if err := s.db.First(&emp, t.EmployeeID).Error; err == nil {
			n := fmt.Sprintf("%s %s", emp.FirstName, emp.LastName)
			empName = &n
		}

		var patName *string
		var pat patientModel.Patient
		if err := s.db.First(&pat, t.PatientID).Error; err == nil {
			n := fmt.Sprintf("%s %s", pat.FirstName, pat.LastName)
			patName = &n
		}

		var invoiceNumber *string
		var inv invoiceModel.Invoice
		if err := s.db.First(&inv, t.InvoiceID).Error; err == nil {
			invoiceNumber = &inv.NumberInvoice
		}

		m := map[string]interface{}{
			"id":                 t.IDLabTicket,
			"number":            t.NumberTicket,
			"g_or_c":            t.GOrC,
			"date_create":       nil,
			"date_promise":      nil,
			"status_id":         t.LabTicketStatusID,
			"patient_id":        t.PatientID,
			"lens_order_id":     t.OrdersLensID,
			"invoice_id":        t.InvoiceID,
			"employee_id":       t.EmployeeID,
			"employee_full_name": empName,
			"patient_full_name":  patName,
			"invoice_number":    invoiceNumber,
		}
		if t.DateCreate != nil {
			m["date_create"] = t.DateCreate.Format("2006-01-02")
		}
		if t.DatePromise != nil {
			m["date_promise"] = t.DatePromise.Format("2006-01-02")
		}
		result = append(result, m)
	}
	return result, nil
}

// ── Detailed ticket by ID (GET /{id_lab_ticket}) ────────────────────────────

func (s *Service) GetTicketByID(ticketID int64) (map[string]interface{}, error) {
	var ticket labTicketModel.LabTicket
	err := s.db.
		Preload("Powers").
		Preload("Lens").
		Preload("Frame").
		Preload("PowersContact").
		Preload("Contact").
		Preload("Contact.ContactLensService").
		First(&ticket, ticketID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ticket not found")
		}
		return nil, err
	}

	// Employee full name
	var empFullName *string
	var emp employeeModel.Employee
	if err := s.db.First(&emp, ticket.EmployeeID).Error; err == nil {
		n := fmt.Sprintf("%s %s", emp.FirstName, emp.LastName)
		empFullName = &n
	}

	// Patient full name
	var patFullName *string
	var pat patientModel.Patient
	if err := s.db.First(&pat, ticket.PatientID).Error; err == nil {
		n := fmt.Sprintf("%s %s", pat.FirstName, pat.LastName)
		patFullName = &n
	}

	// Invoice info
	var invID *int64
	var invNumber *string
	var inv invoiceModel.Invoice
	if err := s.db.First(&inv, ticket.InvoiceID).Error; err == nil {
		invID = &inv.IDInvoice
		invNumber = &inv.NumberInvoice
	}

	resp := map[string]interface{}{
		"id_lab_ticket":        ticket.IDLabTicket,
		"number_ticket":        ticket.NumberTicket,
		"g_or_c":               ticket.GOrC,
		"date":                 nil,
		"date_promise":         nil,
		"lab_id":               ticket.LabID,
		"lab_ticket_status_id": ticket.LabTicketStatusID,
		"patient_id":           ticket.PatientID,
		"orders_lens_id":       ticket.OrdersLensID,
		"invoice_id":           ticket.InvoiceID,
		"tray":                 ticket.Tray,
		"amt":                  ticket.Amt,
		"our_note":             ticket.OurNote,
		"lab_instructions":     ticket.LabInstructions,
		"employee_id":          ticket.EmployeeID,
		"employee_full_name":   empFullName,
		"patient_full_name":    patFullName,
		"id_invoice":           invID,
		"number_invoice":       invNumber,
		"ship_to":              ticket.ShipTo,
		"order_id":             ticket.VwOrderID,
	}
	if ticket.DateCreate != nil {
		resp["date"] = ticket.DateCreate.Format("2006-01-02")
	}
	if ticket.DatePromise != nil {
		resp["date_promise"] = ticket.DatePromise.Format("2006-01-02")
	}

	// tint_allowed flag + allowed lens options
	tintAllowed := false
	var allowedTypes []map[string]interface{}
	var allowedMaterials []map[string]interface{}
	if ticket.InvoiceID != 0 {
		tintAllowed = s.tintAllowedForInvoice(ticket.InvoiceID)
		allowedTypes, allowedMaterials = s.allowedLensOptionsForInvoice(ticket.InvoiceID)
	}
	resp["tint_allowed"] = tintAllowed
	resp["allowed_lens_types"] = allowedTypes
	resp["allowed_lenses_materials"] = allowedMaterials
	resp["lock_lens_options_from_invoice"] = true

	gOrC := ""
	if ticket.GOrC != nil {
		gOrC = *ticket.GOrC
	}

	if gOrC == "g" {
		// Powers
		var powers interface{}
		if ticket.Powers != nil {
			powers = ticket.Powers.ToMap()
		}
		resp["powers"] = powers

		// Lens
		var lensInfo interface{}
		if ticket.Lens != nil {
			lens := ticket.Lens
			allowedTypeIDs := make(map[int]bool)
			for _, t := range allowedTypes {
				if id, ok := t["id"].(int); ok {
					allowedTypeIDs[id] = true
				}
			}
			allowedMatIDs := make(map[int64]bool)
			for _, m := range allowedMaterials {
				if id, ok := m["id"].(int64); ok {
					allowedMatIDs[id] = true
				}
			}

			typeNameByID := func(typeID *int) interface{} {
				if typeID == nil {
					return nil
				}
				for _, t := range allowedTypes {
					if id, ok := t["id"].(int); ok && id == *typeID {
						return t["name"]
					}
				}
				return nil
			}
			matNameByID := func(matID *int) interface{} {
				if matID == nil {
					return nil
				}
				for _, m := range allowedMaterials {
					if id, ok := m["id"].(int64); ok && id == int64(*matID) {
						return m["name"]
					}
				}
				return nil
			}

			li := map[string]interface{}{
				"lens_status":              lens.LensStatus,
				"lens_order":               lens.LensOrder,
				"lens_type":                typeNameByID(lens.LensTypesID),
				"lenses_material":          matNameByID(lens.LensesMaterialsID),
				"edge_thickness":           lens.EdgeThickness,
				"lens_safety_thickness_id": lens.LensSafetyThicknessID,
				"lens_edge_id":             lens.LensEdgeID,
				"center_thickness":         lens.CenterThickness,
				"notes_color":              lens.NotesColor,
				"lenses_id":                lens.LensesID,
				"vw_design_code":           lens.VwDesignCode,
				"vw_material_code":         lens.VwMaterialCode,
			}
			if !tintAllowed {
				li["notes_color"] = nil
			}
			lensInfo = li
		}
		resp["lens"] = lensInfo

		// Frame
		var frameInfo interface{}
		if ticket.Frame != nil {
			f := ticket.Frame
			isPOF := f.POF != nil && *f.POF == "true"
			frameInfo = map[string]interface{}{
				"pof":                 isPOF,
				"model_title_variant": f.ModelTitleVariant,
				"materials_frame":     f.MaterialsFrame,
				"materials_temple":    f.MaterialsTemple,
				"color":               f.Color,
				"size_lens_width":     f.SizeLensWidth,
				"size_bridge_width":   f.SizeBridgeWidth,
				"size_temple_length":  f.SizeTempleLength,
				"frame_shape_id":      f.FrameShapeID,
				"drop_ship":           f.DropShip,
				"b_dim":               f.BValue,
				"circum":              f.CircValue,
				"ed":                  f.EDValue,
				"panto":               f.Panto,
				"wrap_angle":          f.WrapAngle,
				"head_eye_ratio":      f.HeadEyeRatio,
				"stability_coeff":     f.StabilityCoeff,
				"bc":                  f.BC,
				"status":              f.Status,
				"frame_name":          f.FrameName,
				"brand_name":          f.BrandName,
				"vendor_name":         f.VendorName,
				"manufacturer_name":   f.ManufacturerName,
				"head_cape":           f.HeadCape,
				"corridor_r":          f.CorridorR,
				"corridor_l":          f.CorridorL,
			}
		}
		resp["frame"] = frameInfo
		resp["contact_powers"] = nil

	} else if gOrC == "c" {
		resp["powers"] = nil
		resp["lens"] = nil
		resp["frame"] = nil

		var contactPowers interface{}
		if ticket.PowersContact != nil {
			cp := ticket.PowersContact
			var expDate *string
			if cp.ExpirationDate != nil {
				d := cp.ExpirationDate.Format("2006-01-02")
				expDate = &d
			}
			contactPowers = map[string]interface{}{
				"od_cont_lens":         cp.ODContLens,
				"os_cont_lens":         cp.OSContLens,
				"od_bc":                cp.ODBC,
				"os_bc":                cp.OSBC,
				"od_dia":               cp.ODDia,
				"os_dia":               cp.OSDia,
				"od_pwr":               cp.ODPwr,
				"os_pwr":               cp.OSPwr,
				"od_cyl":               cp.ODCyl,
				"os_cyl":               cp.OSCyl,
				"od_axis":              cp.ODAxis,
				"os_axis":              cp.OSAxis,
				"od_add":               cp.ODAdd,
				"os_add":               cp.OSAdd,
				"od_color":             cp.ODColor,
				"os_color":             cp.OSColor,
				"od_type":              cp.ODType,
				"os_type":              cp.OSType,
				"expiration_date":      expDate,
				"od_h_prism_direction": cp.ODHPrismDirection,
				"os_h_prism_direction": cp.OSHPrismDirection,
				"od_v_prism_direction": cp.ODVPrismDirection,
				"os_v_prism_direction": cp.OSVPrismDirection,
			}
		}
		resp["contact_powers"] = contactPowers

		var ticketContactInfo interface{}
		if ticket.Contact != nil {
			ttc := ticket.Contact
			ticketContactInfo = map[string]interface{}{
				"id_lab_ticket_contact":               ttc.IDLabTicketContact,
				"lab_ticket_contact_lens_services_id": ttc.LabTicketContactLensServicesID,
				"od_annual_supply":                    ttc.ODAnnualSupply,
				"os_annual_supply":                    ttc.OSAnnualSupply,
				"od_total_qty":                        ttc.ODTotalQty,
				"os_total_qty":                        ttc.OSTotalQty,
				"reasons":                             ttc.Reasons,
				"modality":                            ttc.Modality,
				"brand_contact_lens_id":               ttc.BrandContactLensID,
			}
		}
		resp["ticket_contact"] = ticketContactInfo

	} else {
		resp["powers"] = nil
		resp["lens"] = nil
		resp["frame"] = nil
		resp["contact_powers"] = nil
	}

	return resp, nil
}

// ── Search tickets (GET /search) ────────────────────────────────────────────

func (s *Service) SearchTickets(ticketNumber, invoiceNumber string) ([]map[string]interface{}, error) {
	q := s.db.Model(&labTicketModel.LabTicket{})

	if ticketNumber != "" {
		q = q.Where("number_ticket = ?", ticketNumber)
	}
	if invoiceNumber != "" {
		// In Python this filters by invoice_id = invoice_number (string->int match)
		q = q.Where("invoice_id = ?", invoiceNumber)
	}

	var tickets []labTicketModel.LabTicket
	if err := q.Find(&tickets).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(tickets))
	for i := range tickets {
		result = append(result, tickets[i].ToMap())
	}
	return result, nil
}

// ── Vendor-Brand combinations (GET /vendor_brand) ───────────────────────────

func (s *Service) GetVendorBrandCombinations() ([]map[string]interface{}, error) {
	type vbRow struct {
		VendorID   int    `gorm:"column:id_vendor"`
		VendorName string `gorm:"column:vendor_name"`
		BrandID    int    `gorm:"column:id_brand"`
		BrandName  string `gorm:"column:brand_name"`
	}
	var rows []vbRow
	err := s.db.
		Table("vendor").
		Select("DISTINCT vendor.id_vendor, vendor.vendor_name, brand.id_brand, brand.brand_name").
		Joins("JOIN product ON product.vendor_id = vendor.id_vendor").
		Joins("JOIN brand ON product.brand_id = brand.id_brand").
		Joins("JOIN model ON model.product_id = product.id_product").
		Joins("JOIN inventory ON inventory.model_id = model.id_model").
		Order("brand.brand_name").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"vendor_id":   r.VendorID,
			"vendor_name": r.VendorName,
			"brand_id":    r.BrandID,
			"brand_name":  r.BrandName,
		})
	}
	return result, nil
}

// ── Products (GET /products) ────────────────────────────────────────────────

func (s *Service) GetProducts(vendorID, brandID string) ([]map[string]interface{}, error) {
	q := s.db.
		Table("product").
		Select("DISTINCT product.id_product, product.title_product").
		Joins("JOIN model ON product.id_product = model.product_id").
		Joins("JOIN inventory ON model.id_model = inventory.model_id")

	if vendorID != "" {
		q = q.Where("product.vendor_id = ?", vendorID)
	}
	if brandID != "" {
		q = q.Where("product.brand_id = ?", brandID)
	}

	type prodRow struct {
		IDProduct    int64  `gorm:"column:id_product"`
		TitleProduct string `gorm:"column:title_product"`
	}
	var rows []prodRow
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"product_id":    r.IDProduct,
			"title_product": r.TitleProduct,
		})
	}
	return result, nil
}

// ── Helper: tint_allowed_for_invoice ────────────────────────────────────────

func (s *Service) tintAllowedForInvoice(invoiceID int64) bool {
	var qtySum int64
	s.db.
		Table("invoice_item_sale").
		Select("COALESCE(SUM(invoice_item_sale.quantity), 0)").
		Joins("JOIN additional_service ON invoice_item_sale.item_id = additional_service.id_additional_service").
		Where("invoice_item_sale.invoice_id = ? AND invoice_item_sale.item_type = ? AND additional_service.tint = true",
			invoiceID, "Add service").
		Scan(&qtySum)
	return qtySum > 0
}

// ── Helper: allowed_lens_options_for_invoice ────────────────────────────────

func (s *Service) allowedLensOptionsForInvoice(invoiceID int64) ([]map[string]interface{}, []map[string]interface{}) {
	type row struct {
		TypeID *int   `gorm:"column:type_id"`
		MatID  *int64 `gorm:"column:mat_id"`
		QtySum int    `gorm:"column:qty_sum"`
	}
	var rows []row
	s.db.
		Table("invoice_item_sale").
		Select("lenses.lens_type_id AS type_id, lenses.lenses_materials_id AS mat_id, SUM(invoice_item_sale.quantity) AS qty_sum").
		Joins("JOIN lenses ON lenses.id_lenses = invoice_item_sale.item_id").
		Where("invoice_item_sale.invoice_id = ? AND invoice_item_sale.item_type = ?", invoiceID, "Lens").
		Group("lenses.lens_type_id, lenses.lenses_materials_id").
		Scan(&rows)

	typeIDs := make(map[int]bool)
	matIDs := make(map[int64]bool)
	for _, r := range rows {
		if r.QtySum > 0 {
			if r.TypeID != nil {
				typeIDs[*r.TypeID] = true
			}
			if r.MatID != nil {
				matIDs[*r.MatID] = true
			}
		}
	}

	var types []map[string]interface{}
	if len(typeIDs) > 0 {
		ids := make([]int, 0, len(typeIDs))
		for id := range typeIDs {
			ids = append(ids, id)
		}
		var lensTypes []lensModel.LensType
		s.db.Where("id_lens_type IN ?", ids).Find(&lensTypes)
		for _, lt := range lensTypes {
			types = append(types, map[string]interface{}{
				"id":   lt.IDLensType,
				"name": lt.TypeName,
			})
		}
	}

	var materials []map[string]interface{}
	if len(matIDs) > 0 {
		ids := make([]int64, 0, len(matIDs))
		for id := range matIDs {
			ids = append(ids, id)
		}
		var lensMats []lensModel.LensesMaterial
		s.db.Where("id_lenses_materials IN ?", ids).Find(&lensMats)
		for _, lm := range lensMats {
			materials = append(materials, map[string]interface{}{
				"id":   lm.IDLensesMaterials,
				"name": lm.MaterialName,
			})
		}
	}

	if types == nil {
		types = []map[string]interface{}{}
	}
	if materials == nil {
		materials = []map[string]interface{}{}
	}
	return types, materials
}

// Ensure models are referenced to avoid unused import errors.
var (
	_ = frameModel.Product{}
	_ = inventoryModel.Inventory{}
	_ = serviceModel.AdditionalService{}
	_ = vendorModel.Vendor{}
)
