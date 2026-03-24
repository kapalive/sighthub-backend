package visionweb_service

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	labTicketModel "sighthub-backend/internal/models/lab_ticket"
	"sighthub-backend/internal/models/vendors"
)

// ─── SOAP / Order constants ─────────────────────────────────────────────────

const (
	// QA environment (switch to production URLs when ready)
	orderUploadURL = "https://services.visionwebqa.com/FileUpload.asmx"
	soapUsername   = "SighthubQA"
	soapPassword   = "vision"
	vwCustomerUser = "SighthubCustomer"
	vwCustomerPass = "vision2020"
	vwRefID        = "ROSIGHTHUB"
)

// ─── Result types ───────────────────────────────────────────────────────────

type OrderResult struct {
	OrderID    string `json:"order_id"`
	VWOrderID  string `json:"vw_order_id,omitempty"`
	Status     string `json:"status"`
	ErrorList  string `json:"error_list,omitempty"`
	RawResponse string `json:"raw_response,omitempty"`
}

// VW response XML
type singleOrderResp struct {
	XMLName       xml.Name `xml:"SingleOrder"`
	OrderID       string   `xml:"OrderId"`
	PatientName   string   `xml:"PatientName"`
	VWebOrderID   string   `xml:"VWebOrderId"`
	VWebExchangeID string  `xml:"VWebExchangeId"`
	Status        string   `xml:"Status"`
	ErrorList     string   `xml:"ErrorList"`
	SentDate      string   `xml:"SentDate"`
}

// SOAP response envelope
type soapUploadResp struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		UploadFileResponse struct {
			Result string `xml:"UploadFileResult"`
		} `xml:"UploadFileResponse"`
	} `xml:"Body"`
}

// ─── Input for placing an order ─────────────────────────────────────────────

type PlaceOrderInput struct {
	TicketID   int64
	EmployeeID int64
}

// ─── validation errors ──────────────────────────────────────────────────────

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (e *ValidationErrors) Error() string {
	msgs := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "; ")
}

// ─── Order requirements check ───────────────────────────────────────────────

type FieldStatus struct {
	Field    string      `json:"field"`
	Label    string      `json:"label"`
	Required bool        `json:"required"`
	Filled   bool        `json:"filled"`
	Value    interface{} `json:"value,omitempty"`
	Source   string      `json:"source"`
}

type OrderRequirements struct {
	Ready  bool          `json:"ready"`
	Fields []FieldStatus `json:"fields"`
}

func (s *Service) CheckOrderRequirements(ticketID int64) *OrderRequirements {
	var ticket labTicketModel.LabTicket
	err := s.db.
		Preload("Powers").
		Preload("Lens").
		Preload("Frame").
		Preload("Frame.FrameTypeMaterial").
		Preload("Lab").
		First(&ticket, ticketID).Error
	if err != nil {
		return &OrderRequirements{Ready: false, Fields: []FieldStatus{
			{Field: "ticket", Label: "Lab Ticket", Required: true, Filled: false, Source: "lab_ticket"},
		}}
	}

	var fields []FieldStatus
	add := func(field, label, source string, required bool, filled bool, val interface{}) {
		fields = append(fields, FieldStatus{Field: field, Label: label, Required: required, Filled: filled, Value: val, Source: source})
	}

	// Lab
	labFilled := ticket.LabID != nil && ticket.Lab != nil
	labName := interface{}(nil)
	if labFilled {
		labName = ticket.Lab.VendorName
	}
	add("lab_id", "Laboratory", "lab_ticket", true, labFilled, labName)

	// VW Account
	vwOK := false
	if labFilled {
		var vla vendors.VendorLocationAccount
		type invRow struct{ LocationID int64 }
		var inv invRow
		s.db.Table("invoice").Select("location_id").Where("id_invoice = ?", ticket.InvoiceID).Scan(&inv)
		if inv.LocationID > 0 {
			err := s.db.Where("vendor_id = ? AND location_id = ? AND vw_slo_id IS NOT NULL AND source = 'vision_web'",
				ticket.Lab.IDVendor, inv.LocationID).First(&vla).Error
			vwOK = err == nil && vla.VwSloID != nil && vla.VwBill != nil && vla.VwShip != nil
			val := interface{}(nil)
			if vwOK {
				val = fmt.Sprintf("slo=%d bill=%s ship=%s", *vla.VwSloID, *vla.VwBill, *vla.VwShip)
			}
			add("vw_account", "VisionWeb Account (ship#/bill#)", "vendor_location_account", true, vwOK, val)
		} else {
			add("vw_account", "VisionWeb Account (ship#/bill#)", "vendor_location_account", true, false, nil)
		}
	} else {
		add("vw_account", "VisionWeb Account (ship#/bill#)", "vendor_location_account", true, false, nil)
	}

	// Lens VW codes — from lab_ticket_lens directly
	lens := ticket.Lens
	hasLens := lens != nil
	designOK := hasLens && lens.VwDesignCode != nil && *lens.VwDesignCode != ""
	materialOK := hasLens && lens.VwMaterialCode != nil && *lens.VwMaterialCode != ""
	add("lenses_id", "Lens (from VW catalog)", "lab_ticket_lens.lenses_id", true, hasLens && lens.LensesID != nil, valStr(hasLens, func() interface{} { return lens.LensesID }))
	add("vw_design_code", "Lens Design (VW)", "lab_ticket_lens.vw_design_code", true, designOK, valStr(hasLens, func() interface{} { return lens.VwDesignCode }))
	add("vw_material_code", "Lens Material (VW)", "lab_ticket_lens.vw_material_code", true, materialOK, valStr(hasLens, func() interface{} { return lens.VwMaterialCode }))

	// Powers
	p := ticket.Powers
	hasPowers := p != nil
	add("od_sph", "OD Sphere", "lab_ticket_powers", true, hasPowers && p.ODSph != nil && *p.ODSph != "", valStr(hasPowers, func() interface{} { return p.ODSph }))
	add("os_sph", "OS Sphere", "lab_ticket_powers", true, hasPowers && p.OSSph != nil && *p.OSSph != "", valStr(hasPowers, func() interface{} { return p.OSSph }))
	add("od_cyl", "OD Cylinder", "lab_ticket_powers", false, hasPowers && p.ODCyl != nil && *p.ODCyl != "", valStr(hasPowers, func() interface{} { return p.ODCyl }))
	add("os_cyl", "OS Cylinder", "lab_ticket_powers", false, hasPowers && p.OSCyl != nil && *p.OSCyl != "", valStr(hasPowers, func() interface{} { return p.OSCyl }))
	add("od_axis", "OD Axis", "lab_ticket_powers", false, hasPowers && p.ODAxis != nil && *p.ODAxis != "", valStr(hasPowers, func() interface{} { return p.ODAxis }))
	add("os_axis", "OS Axis", "lab_ticket_powers", false, hasPowers && p.OSAxis != nil && *p.OSAxis != "", valStr(hasPowers, func() interface{} { return p.OSAxis }))
	add("od_dt", "OD Distance PD", "lab_ticket_powers", true, hasPowers && p.ODDT != nil && *p.ODDT != "", valStr(hasPowers, func() interface{} { return p.ODDT }))
	add("os_dt", "OS Distance PD", "lab_ticket_powers", true, hasPowers && p.OSDT != nil && *p.OSDT != "", valStr(hasPowers, func() interface{} { return p.OSDT }))
	add("od_nr", "OD Near PD", "lab_ticket_powers", false, hasPowers && p.ODNR != nil && *p.ODNR != "", valStr(hasPowers, func() interface{} { return p.ODNR }))
	add("os_nr", "OS Near PD", "lab_ticket_powers", false, hasPowers && p.OSNR != nil && *p.OSNR != "", valStr(hasPowers, func() interface{} { return p.OSNR }))
	add("od_add", "OD Add", "lab_ticket_powers", false, hasPowers && p.ODAdd != nil, valStr(hasPowers, func() interface{} { return p.ODAdd }))
	add("os_add", "OS Add", "lab_ticket_powers", false, hasPowers && p.OSAdd != nil, valStr(hasPowers, func() interface{} { return p.OSAdd }))
	add("od_seg_hd", "OD Seg Height", "lab_ticket_powers", false, hasPowers && p.ODSegHD != nil && *p.ODSegHD != "", valStr(hasPowers, func() interface{} { return p.ODSegHD }))
	add("os_seg_hd", "OS Seg Height", "lab_ticket_powers", false, hasPowers && p.OSSegHD != nil && *p.OSSegHD != "", valStr(hasPowers, func() interface{} { return p.OSSegHD }))
	add("od_oc", "OD Optical Center", "lab_ticket_powers", false, hasPowers && p.ODOC != nil && *p.ODOC != "", valStr(hasPowers, func() interface{} { return p.ODOC }))
	add("os_oc", "OS Optical Center", "lab_ticket_powers", false, hasPowers && p.OSOC != nil && *p.OSOC != "", valStr(hasPowers, func() interface{} { return p.OSOC }))
	add("od_h_prism", "OD Horiz Prism", "lab_ticket_powers", false, hasPowers && p.ODHPrism != nil, valStr(hasPowers, func() interface{} { return p.ODHPrism }))
	add("os_h_prism", "OS Horiz Prism", "lab_ticket_powers", false, hasPowers && p.OSHPrism != nil, valStr(hasPowers, func() interface{} { return p.OSHPrism }))
	add("od_v_prism", "OD Vert Prism", "lab_ticket_powers", false, hasPowers && p.ODVPrism != nil, valStr(hasPowers, func() interface{} { return p.ODVPrism }))
	add("os_v_prism", "OS Vert Prism", "lab_ticket_powers", false, hasPowers && p.OSVPrism != nil, valStr(hasPowers, func() interface{} { return p.OSVPrism }))
	add("od_bc", "OD Base Curve", "lab_ticket_powers", false, hasPowers && p.ODBC != nil && *p.ODBC != "", valStr(hasPowers, func() interface{} { return p.ODBC }))
	add("os_bc", "OS Base Curve", "lab_ticket_powers", false, hasPowers && p.OSBC != nil && *p.OSBC != "", valStr(hasPowers, func() interface{} { return p.OSBC }))

	// Frame
	f := ticket.Frame
	hasFrame := f != nil
	add("size_lens_width", "Frame A / Eye Size", "lab_ticket_frame", true, hasFrame && f.SizeLensWidth != nil && *f.SizeLensWidth != "", valStr(hasFrame, func() interface{} { return f.SizeLensWidth }))
	add("b_value", "Frame B (box)", "lab_ticket_frame", true, hasFrame && f.BValue != nil, valStr(hasFrame, func() interface{} { return f.BValue }))
	add("ed_value", "Frame ED", "lab_ticket_frame", true, hasFrame && f.EDValue != nil, valStr(hasFrame, func() interface{} { return f.EDValue }))
	add("size_bridge_width", "Frame DBL (bridge)", "lab_ticket_frame", true, hasFrame && f.SizeBridgeWidth != nil && *f.SizeBridgeWidth != "", valStr(hasFrame, func() interface{} { return f.SizeBridgeWidth }))
	add("size_lens_width", "Frame Eye Size", "lab_ticket_frame", false, hasFrame && f.SizeLensWidth != nil, valStr(hasFrame, func() interface{} { return f.SizeLensWidth }))
	add("size_temple_length", "Frame Temple Length", "lab_ticket_frame", false, hasFrame && f.SizeTempleLength != nil, valStr(hasFrame, func() interface{} { return f.SizeTempleLength }))
	add("frame_type_material", "Frame Type/Material", "lab_ticket_frame", false, hasFrame && f.FrameTypeMaterialID != nil, valStr(hasFrame && f.FrameTypeMaterial != nil, func() interface{} { return f.FrameTypeMaterial.Material }))
	add("brand_name", "Frame Brand", "lab_ticket_frame", false, hasFrame && f.BrandName != nil, valStr(hasFrame, func() interface{} { return f.BrandName }))
	add("model_title_variant", "Frame Model", "lab_ticket_frame", false, hasFrame && f.ModelTitleVariant != nil, valStr(hasFrame, func() interface{} { return f.ModelTitleVariant }))
	add("color", "Frame Color", "lab_ticket_frame", false, hasFrame && f.Color != nil, valStr(hasFrame, func() interface{} { return f.Color }))
	add("panto", "Panto Angle", "lab_ticket_frame", false, hasFrame && f.Panto != nil, valStr(hasFrame, func() interface{} { return f.Panto }))
	add("wrap_angle", "Wrap Angle", "lab_ticket_frame", false, hasFrame && f.WrapAngle != nil, valStr(hasFrame, func() interface{} { return f.WrapAngle }))

	// Special instructions
	add("lab_instructions", "Special Instructions", "lab_ticket", false, ticket.LabInstructions != nil && *ticket.LabInstructions != "", ticket.LabInstructions)

	// Check if all required are filled
	ready := true
	for _, f := range fields {
		if f.Required && !f.Filled {
			ready = false
			break
		}
	}

	return &OrderRequirements{Ready: ready, Fields: fields}
}

func valStr(has bool, fn func() interface{}) interface{} {
	if !has {
		return nil
	}
	return fn()
}

// ─── Main method ────────────────────────────────────────────────────────────

func (s *Service) PlaceOrder(ticketID int64) (*OrderResult, error) {
	// 1. Load ticket with all relations
	var ticket labTicketModel.LabTicket
	err := s.db.
		Preload("Powers").
		Preload("Lens").
		Preload("Frame").
		Preload("Frame.FrameTypeMaterial").
		Preload("Lab").
		First(&ticket, ticketID).Error
	if err != nil {
		return nil, fmt.Errorf("ticket not found")
	}

	// 2-9. Collect ALL validation issues at once
	var errs []ValidationError

	// Ticket type
	if ticket.GOrC != nil && *ticket.GOrC != "g" {
		errs = append(errs, ValidationError{"g_or_c", "VisionWeb orders are only supported for glasses tickets (g_or_c = 'g')"})
	}

	// 3. Load patient
	type patientRow struct {
		FirstName string
		LastName  string
	}
	var patient patientRow
	if err := s.db.Table("patient").Select("first_name, last_name").
		Where("id_patient = ?", ticket.PatientID).Scan(&patient).Error; err != nil {
		errs = append(errs, ValidationError{"patient", "patient not found"})
	} else if patient.LastName == "" {
		errs = append(errs, ValidationError{"patient", "patient last name is required"})
	}

	// 4. Load invoice to get location_id
	type invoiceRow struct {
		LocationID int64
	}
	var invoice invoiceRow
	if err := s.db.Table("invoice").Select("location_id").
		Where("id_invoice = ?", ticket.InvoiceID).Scan(&invoice).Error; err != nil {
		errs = append(errs, ValidationError{"invoice", "invoice not found"})
	}

	// 5. Lab
	var vla vendors.VendorLocationAccount
	var vlaOK bool
	if ticket.LabID == nil || ticket.Lab == nil {
		errs = append(errs, ValidationError{"lab", "lab is not selected on this ticket"})
	} else {
		lab := ticket.Lab
		if invoice.LocationID == 0 {
			// skip VW account check — invoice already errored
		} else {
			// 6. VW account (lab IS the vendor now)
			err = s.db.Where("vendor_id = ? AND location_id = ? AND vw_slo_id IS NOT NULL AND source = 'vision_web'",
				lab.IDVendor, invoice.LocationID).First(&vla).Error
			if err != nil {
				errs = append(errs, ValidationError{"vw_account", fmt.Sprintf("no VisionWeb account configured for lab '%s' (vendor_id=%d) at this location (location_id=%d) — need ship#/bill# in vendor_location_account", lab.VendorName, lab.IDVendor, invoice.LocationID)})
			} else if vla.VwSloID == nil || vla.VwBill == nil || vla.VwShip == nil {
				errs = append(errs, ValidationError{"vw_account", fmt.Sprintf("VisionWeb account for lab '%s' is incomplete — missing: %s", lab.VendorName, missingVWFields(&vla))})
			} else {
				vlaOK = true
			}
		}
	}

	// 7. Lens VW codes — read directly from lab_ticket_lens
	var vwDesign, vwMaterial string
	if ticket.Lens == nil {
		errs = append(errs, ValidationError{"lens", "lens data is required"})
	} else {
		if ticket.Lens.VwDesignCode != nil && *ticket.Lens.VwDesignCode != "" {
			vwDesign = *ticket.Lens.VwDesignCode
		}
		if ticket.Lens.VwMaterialCode != nil && *ticket.Lens.VwMaterialCode != "" {
			vwMaterial = *ticket.Lens.VwMaterialCode
		}
		if vwDesign == "" {
			errs = append(errs, ValidationError{"vw_design_code", "lens has no VisionWeb design code — select a lens from VW catalog (lenses_id)"})
		}
		if vwMaterial == "" {
			errs = append(errs, ValidationError{"vw_material_code", "lens has no VisionWeb material code — select a lens from VW catalog (lenses_id)"})
		}
	}

	// 8. Treatments
	var treatments []struct {
		VwCode string
	}
	s.db.Table("lab_ticket_invoice_item ltii").
		Select("lt.vw_code").
		Joins("JOIN invoice_item_sale iis ON iis.id_invoice_sale = ltii.invoice_item_id").
		Joins("JOIN invoice_services_item isi ON isi.invoice_id = iis.invoice_id").
		Joins("JOIN lens_treatments lt ON lt.id_lens_treatments = isi.additional_service_id").
		Where("ltii.lab_ticket_id = ? AND lt.vw_code IS NOT NULL", ticketID).
		Scan(&treatments)

	// 9. Powers
	if ticket.Powers == nil {
		errs = append(errs, ValidationError{"powers", "prescription powers are required"})
	} else {
		p := ticket.Powers
		if p.ODSph == nil || *p.ODSph == "" {
			errs = append(errs, ValidationError{"od_sph", "OD Sphere is required"})
		}
		if p.OSSph == nil || *p.OSSph == "" {
			errs = append(errs, ValidationError{"os_sph", "OS Sphere is required"})
		}
		if p.ODDT == nil || *p.ODDT == "" {
			errs = append(errs, ValidationError{"od_dt", "OD Distance PD is required"})
		}
		if p.OSDT == nil || *p.OSDT == "" {
			errs = append(errs, ValidationError{"os_dt", "OS Distance PD is required"})
		}
	}

	// Frame
	if ticket.Frame == nil {
		errs = append(errs, ValidationError{"frame", "frame data is required"})
	} else {
		f := ticket.Frame
		if f.SizeLensWidth == nil || *f.SizeLensWidth == "" {
			errs = append(errs, ValidationError{"size_lens_width", "frame A / Eye Size is required"})
		}
		if f.BValue == nil {
			errs = append(errs, ValidationError{"b_value", "frame B measurement is required"})
		}
		if f.EDValue == nil {
			errs = append(errs, ValidationError{"ed_value", "frame ED measurement is required"})
		}
		if f.SizeBridgeWidth == nil || *f.SizeBridgeWidth == "" {
			errs = append(errs, ValidationError{"size_bridge_width", "frame DBL (bridge width) is required"})
		}
	}

	if len(errs) > 0 {
		return nil, &ValidationErrors{Errors: errs}
	}

	// If VW account wasn't resolved, bail (shouldn't happen after validation above)
	_ = vlaOK

	// 10. Build VWOrder XML
	orderID := fmt.Sprintf("SH-%d-%d", ticketID, time.Now().Unix())

	items := []vwItem{
		{Name: "Username", Value: vwCustomerUser},
		{Name: "Password", Value: vwCustomerPass},
		{Name: "OrderId", Value: orderID},
		{Name: "PONumber", Value: ticket.NumberTicket},
		{Name: "SupplierName", Value: fmt.Sprintf("%d", *vla.VwSloID)},
		{Name: "BillAccount", Value: *vla.VwBill},
		{Name: "ShipAccount", Value: *vla.VwShip},
		{Name: "JobType", Value: "Frame To Come"},
		{Name: "Eyes", Value: resolveEyes(ticket.Powers)},
	}

	// Patient
	items = append(items, vwItem{Name: "PatLastName", Value: patient.LastName})
	items = append(items, vwItem{Name: "PatFirstName", Value: patient.FirstName})

	// Powers (RE = OD, LE = OS in VW terminology)
	p := ticket.Powers
	items = appendIfSet(items, "RESph", p.ODSph)
	items = appendIfSet(items, "RECyl", p.ODCyl)
	items = appendIfSet(items, "REAxis", p.ODAxis)
	items = appendIfSet(items, "REDistPD", p.ODDT)
	items = appendIfSet(items, "RENearPD", p.ODNR)
	if p.ODHPrism != nil {
		items = append(items, vwItem{Name: "REHorizPrismValue", Value: fmt.Sprintf("%.2f", *p.ODHPrism)})
	}
	if p.ODHPrismDirection != nil {
		items = append(items, vwItem{Name: "REHorizPrismDirection", Value: string(*p.ODHPrismDirection)})
	}
	if p.ODVPrism != nil {
		items = append(items, vwItem{Name: "REVerticalPrismValue", Value: fmt.Sprintf("%.2f", *p.ODVPrism)})
	}
	if p.ODVPrismDirection != nil {
		items = append(items, vwItem{Name: "REVerticalPrismDirection", Value: string(*p.ODVPrismDirection)})
	}
	if p.ODAdd != nil {
		items = append(items, vwItem{Name: "REAdd", Value: fmt.Sprintf("%.2f", *p.ODAdd)})
	}
	items = appendIfSet(items, "RESegHeight", p.ODSegHD)
	items = appendIfSet(items, "REOpticalCenter", p.ODOC)

	items = appendIfSet(items, "LESph", p.OSSph)
	items = appendIfSet(items, "LECyl", p.OSCyl)
	items = appendIfSet(items, "LEAxis", p.OSAxis)
	items = appendIfSet(items, "LEDistPD", p.OSDT)
	items = appendIfSet(items, "LENearPD", p.OSNR)
	if p.OSHPrism != nil {
		items = append(items, vwItem{Name: "LEHorizPrismValue", Value: fmt.Sprintf("%.2f", *p.OSHPrism)})
	}
	if p.OSHPrismDirection != nil {
		items = append(items, vwItem{Name: "LEHorizPrismDirection", Value: string(*p.OSHPrismDirection)})
	}
	if p.OSVPrism != nil {
		items = append(items, vwItem{Name: "LEVerticalPrismValue", Value: fmt.Sprintf("%.2f", *p.OSVPrism)})
	}
	if p.OSVPrismDirection != nil {
		items = append(items, vwItem{Name: "LEVerticalPrismDirection", Value: string(*p.OSVPrismDirection)})
	}
	if p.OSAdd != nil {
		items = append(items, vwItem{Name: "LEAdd", Value: fmt.Sprintf("%.2f", *p.OSAdd)})
	}
	items = appendIfSet(items, "LESegHeight", p.OSSegHD)
	items = appendIfSet(items, "LEOpticalCenter", p.OSOC)

	// Lens specification — same for both eyes
	items = append(items, vwItem{Name: "RELensDesign", Value: vwDesign})
	items = append(items, vwItem{Name: "RELensMaterial", Value: vwMaterial})
	items = append(items, vwItem{Name: "LELensDesign", Value: vwDesign})
	items = append(items, vwItem{Name: "LELensMaterial", Value: vwMaterial})

	// Treatments (up to 3 per eye, same both eyes)
	for i, t := range treatments {
		if i >= 3 {
			break
		}
		num := fmt.Sprintf("%d", i+1)
		items = append(items, vwItem{Name: "RETreatment" + num, Value: t.VwCode})
		items = append(items, vwItem{Name: "LETreatment" + num, Value: t.VwCode})
	}

	// Frame specification
	f := ticket.Frame
	frameType := "ZYL" // default
	if f.FrameTypeMaterial != nil {
		frameType = mapFrameType(f.FrameTypeMaterial.Material)
	}
	items = append(items, vwItem{Name: "FrameType", Value: frameType})
	if f.SizeLensWidth != nil && *f.SizeLensWidth != "" {
		items = append(items, vwItem{Name: "ABox", Value: *f.SizeLensWidth})
	}
	if f.BValue != nil {
		items = append(items, vwItem{Name: "BBox", Value: fmt.Sprintf("%d", *f.BValue)})
	}
	if f.SizeBridgeWidth != nil {
		items = append(items, vwItem{Name: "Dbl", Value: *f.SizeBridgeWidth})
	}
	if f.EDValue != nil {
		items = append(items, vwItem{Name: "ED", Value: fmt.Sprintf("%d", *f.EDValue)})
	}
	if f.SizeLensWidth != nil {
		items = append(items, vwItem{Name: "Eye", Value: *f.SizeLensWidth})
	}
	if f.SizeTempleLength != nil {
		items = append(items, vwItem{Name: "FrameTempleLength", Value: *f.SizeTempleLength})
	}
	if f.BrandName != nil {
		items = append(items, vwItem{Name: "FrameManufacturer", Value: *f.BrandName})
	}
	if f.ModelTitleVariant != nil {
		items = append(items, vwItem{Name: "FrameModel", Value: *f.ModelTitleVariant})
	}
	if f.Color != nil {
		items = append(items, vwItem{Name: "FrameColor", Value: *f.Color})
	}

	// Special instructions
	if ticket.LabInstructions != nil && *ticket.LabInstructions != "" {
		instr := *ticket.LabInstructions
		if len(instr) > 60 {
			instr = instr[:60]
		}
		items = append(items, vwItem{Name: "SpecialInstructions1", Value: instr})
	}

	// Additional parameters
	if f.Panto != nil {
		items = append(items, vwItem{Name: "PantoAngle", Value: fmt.Sprintf("%.1f", *f.Panto)})
	}
	if f.WrapAngle != nil {
		items = append(items, vwItem{Name: "WrapAngle", Value: fmt.Sprintf("%.1f", *f.WrapAngle)})
	}
	if p.ODBC != nil && *p.ODBC != "" {
		items = append(items, vwItem{Name: "REBaseCurve", Value: *p.ODBC})
	}
	if p.OSBC != nil && *p.OSBC != "" {
		items = append(items, vwItem{Name: "LEBaseCurve", Value: *p.OSBC})
	}

	xmlPayload := buildVWOrderXML(items)

	// 11. Send SOAP request
	soapBody := buildSOAPEnvelope(xmlPayload, orderID, *vla.VwSloID)

	req, _ := http.NewRequest("POST", orderUploadURL, strings.NewReader(soapBody))
	req.Header.Set("Content-Type", "text/xml")

	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("VisionWeb request failed: %w", err)
	}
	defer resp.Body.Close()

	respData, _ := io.ReadAll(resp.Body)
	respStr := string(respData)

	// 12. Parse response
	result := &OrderResult{
		OrderID:     orderID,
		RawResponse: respStr,
	}

	// Try to extract SingleOrder or ORDER_MSG from SOAP response
	var soapResp soapUploadResp
	if xml.Unmarshal(respData, &soapResp) == nil && soapResp.Body.UploadFileResponse.Result != "" {
		inner := soapResp.Body.UploadFileResponse.Result

		// Try SingleOrder (tracking response)
		var single singleOrderResp
		if xml.Unmarshal([]byte(inner), &single) == nil && single.Status != "" {
			result.VWOrderID = single.VWebOrderID
			result.Status = single.Status
			result.ErrorList = single.ErrorList
			return result, nil
		}

		// Try ORDER_MSG (file upload response)
		type orderMsg struct {
			XMLName xml.Name `xml:"ORDER_MSG"`
			Text    string   `xml:",chardata"`
		}
		var msg orderMsg
		if xml.Unmarshal([]byte(inner), &msg) == nil && msg.Text != "" {
			if strings.Contains(strings.ToLower(msg.Text), "successfully") {
				result.Status = "Sent"
			} else {
				result.Status = "Error"
				result.ErrorList = msg.Text
			}
			result.RawResponse = msg.Text
			return result, nil
		}
	}

	// Fallback: try direct parse of response body
	var single singleOrderResp
	if xml.Unmarshal(respData, &single) == nil && single.Status != "" {
		result.VWOrderID = single.VWebOrderID
		result.Status = single.Status
		result.ErrorList = single.ErrorList
		return result, nil
	}

	// If response doesn't parse, check HTTP status
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("VisionWeb returned HTTP %d: %s", resp.StatusCode, respStr)
	}

	result.Status = "submitted"
	return result, nil
}

// ─── XML helpers ────────────────────────────────────────────────────────────

type vwItem struct {
	Name  string
	Value string
}

func buildVWOrderXML(items []vwItem) string {
	var b strings.Builder
	b.WriteString("<VWOrder>\n")
	for _, it := range items {
		b.WriteString("  <Item>\n")
		b.WriteString("    <FieldName>")
		b.WriteString(xmlEscape(it.Name))
		b.WriteString("</FieldName>\n")
		b.WriteString("    <FieldValue>")
		b.WriteString(xmlEscape(it.Value))
		b.WriteString("</FieldValue>\n")
		b.WriteString("  </Item>\n")
	}
	b.WriteString("</VWOrder>")
	return b.String()
}

func buildSOAPEnvelope(vwOrderXML string, orderID string, sloID int) string {
	msgGUID := fmt.Sprintf("%s_%d", orderID, time.Now().UnixNano())
	return fmt.Sprintf(`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ser="http://services.visionweb.com">
   <soapenv:Header/>
   <soapenv:Body>
      <ser:UploadFile>
         <ser:username>%s</ser:username>
         <ser:pswd>%s</ser:pswd>
         <ser:filestring><![CDATA[%s]]></ser:filestring>
         <ser:subordid>%s</ser:subordid>
         <ser:refid>%s</ser:refid>
         <ser:guid></ser:guid>
         <ser:msgguid>%s</ser:msgguid>
         <ser:sloid>%d</ser:sloid>
         <ser:cbsid></ser:cbsid>
         <ser:ordtype></ser:ordtype>
         <ser:filename></ser:filename>
      </ser:UploadFile>
   </soapenv:Body>
</soapenv:Envelope>`, soapUsername, soapPassword, vwOrderXML, orderID, vwRefID, msgGUID, sloID)
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func appendIfSet(items []vwItem, name string, val *string) []vwItem {
	if val != nil && *val != "" {
		items = append(items, vwItem{Name: name, Value: *val})
	}
	return items
}

func resolveEyes(p *labTicketModel.LabTicketPowers) string {
	if p == nil {
		return "B"
	}
	hasOD := p.ODSph != nil && *p.ODSph != ""
	hasOS := p.OSSph != nil && *p.OSSph != ""
	if hasOD && hasOS {
		return "B"
	}
	if hasOD {
		return "R"
	}
	if hasOS {
		return "L"
	}
	return "B"
}

func missingVWFields(vla *vendors.VendorLocationAccount) string {
	var missing []string
	if vla.VwSloID == nil {
		missing = append(missing, "vw_slo_id")
	}
	if vla.VwBill == nil {
		missing = append(missing, "vw_bill")
	}
	if vla.VwShip == nil {
		missing = append(missing, "vw_ship")
	}
	return strings.Join(missing, ", ")
}

// mapFrameType maps our FrameTypeMaterial.Material string to VW's FrameType codes:
// METAL, ZYL, DRILL, RIMLESS, INDUST
func mapFrameType(material string) string {
	m := strings.ToUpper(material)
	switch {
	case strings.Contains(m, "METAL"), strings.Contains(m, "TITANIUM"), strings.Contains(m, "STAINLESS"):
		return "METAL"
	case strings.Contains(m, "DRILL"):
		return "DRILL"
	case strings.Contains(m, "RIMLESS"):
		return "RIMLESS"
	case strings.Contains(m, "INDUST"), strings.Contains(m, "SAFETY"):
		return "INDUST"
	default:
		// ZYL covers: Zylonite, Acetate, Plastic, Nylon, Carbon Fiber, etc.
		return "ZYL"
	}
}
