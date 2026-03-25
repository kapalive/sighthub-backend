package zeiss_service

import (
	"fmt"

	labTicketModel "sighthub-backend/internal/models/lab_ticket"
)

type ZeissFieldStatus struct {
	Field    string      `json:"field"`
	Label    string      `json:"label"`
	Required bool        `json:"required"`
	Filled   bool        `json:"filled"`
	Value    interface{} `json:"value,omitempty"`
	Source   string      `json:"source"`
}

type ZeissOrderRequirements struct {
	Ready  bool               `json:"ready"`
	Fields []ZeissFieldStatus `json:"fields"`
}

func (s *CatalogService) CheckZeissOrderRequirements(ticketID int64, employeeID int64) *ZeissOrderRequirements {
	var ticket labTicketModel.LabTicket
	err := s.db.
		Preload("Powers").
		Preload("Lens").
		Preload("Frame").
		Preload("Lab").
		First(&ticket, ticketID).Error
	if err != nil {
		return &ZeissOrderRequirements{Ready: false, Fields: []ZeissFieldStatus{
			{Field: "ticket", Label: "Lab Ticket", Required: true, Filled: false, Source: "lab_ticket"},
		}}
	}

	var fields []ZeissFieldStatus
	add := func(field, label, source string, required, filled bool, val interface{}) {
		fields = append(fields, ZeissFieldStatus{Field: field, Label: label, Required: required, Filled: filled, Value: val, Source: source})
	}

	// ── Zeiss Auth ──
	authOK := s.auth.IsAuthenticated(employeeID)
	add("zeiss_auth", "Zeiss Authentication", "zeiss_token", true, authOK, nil)

	// ── Customer Number ──
	status := s.auth.GetAuthStatus(employeeID)
	custOK := status.CustomerNumber != nil && *status.CustomerNumber != ""
	var custVal interface{}
	if custOK {
		custVal = *status.CustomerNumber
	}
	add("customer_number", "Zeiss Customer Number", "zeiss_token", true, custOK, custVal)

	// ── Lab = CARL ZEISS (id=69) ──
	labOK := ticket.LabID != nil && *ticket.LabID == ZeissVendorID
	labName := interface{}(nil)
	if ticket.Lab != nil {
		labName = ticket.Lab.VendorName
	}
	add("lab_id", "Laboratory (CARL ZEISS)", "lab_ticket", true, labOK, labName)

	// ── Lens commercial code ──
	lens := ticket.Lens
	hasLens := lens != nil
	codeOK := hasLens && lens.VwDesignCode != nil && *lens.VwDesignCode != ""
	var codeVal interface{}
	if codeOK {
		codeVal = *lens.VwDesignCode
	}
	// Auto-sync from invoice if missing
	if hasLens && !codeOK {
		codeVal, codeOK = s.trySyncLensFromInvoice(ticket.InvoiceID, lens)
	}
	add("commercial_code", "Lens Commercial Code (Zeiss)", "lab_ticket_lens", true, codeOK, codeVal)

	// ── Coating (optional but common) ──
	coatingOK := false
	var coatingVal interface{}
	if ticket.InvoiceID > 0 {
		type treatRow struct {
			VwCode      *string
			Description *string
		}
		var treats []treatRow
		s.db.Raw(`
			SELECT lt.vw_code, lt.description
			FROM lab_ticket_invoice_item ltii
			JOIN invoice_item_sale iis ON iis.id_invoice_sale = ltii.invoice_item_id
			JOIN lens_treatments lt ON lt.id_lens_treatments = iis.item_id AND iis.item_type = 'Treatment'
			WHERE ltii.lab_ticket_id = ? AND lt.source = 'zeiss_only'
		`, ticketID).Scan(&treats)
		if len(treats) > 0 && treats[0].VwCode != nil {
			coatingOK = true
			coatingVal = *treats[0].VwCode
		}
	}
	add("coating_code", "Coating (Zeiss)", "invoice_treatments", false, coatingOK, coatingVal)

	// ── RX Data ──
	p := ticket.Powers
	hp := p != nil
	add("od_sph", "OD Sphere", "lab_ticket_powers", true, hp && p.ODSph != nil && *p.ODSph != "", valP(hp, p, func() interface{} { return p.ODSph }))
	add("os_sph", "OS Sphere", "lab_ticket_powers", true, hp && p.OSSph != nil && *p.OSSph != "", valP(hp, p, func() interface{} { return p.OSSph }))
	add("od_cyl", "OD Cylinder", "lab_ticket_powers", false, hp && p.ODCyl != nil && *p.ODCyl != "", valP(hp, p, func() interface{} { return p.ODCyl }))
	add("os_cyl", "OS Cylinder", "lab_ticket_powers", false, hp && p.OSCyl != nil && *p.OSCyl != "", valP(hp, p, func() interface{} { return p.OSCyl }))
	add("od_axis", "OD Axis", "lab_ticket_powers", false, hp && p.ODAxis != nil && *p.ODAxis != "", valP(hp, p, func() interface{} { return p.ODAxis }))
	add("os_axis", "OS Axis", "lab_ticket_powers", false, hp && p.OSAxis != nil && *p.OSAxis != "", valP(hp, p, func() interface{} { return p.OSAxis }))
	add("od_add", "OD Addition", "lab_ticket_powers", false, hp && p.ODAdd != nil, valP(hp, p, func() interface{} { return p.ODAdd }))
	add("os_add", "OS Addition", "lab_ticket_powers", false, hp && p.OSAdd != nil, valP(hp, p, func() interface{} { return p.OSAdd }))

	// ── Centration ──
	add("od_dt", "OD Distance PD", "lab_ticket_powers", true, hp && p.ODDT != nil && *p.ODDT != "", valP(hp, p, func() interface{} { return p.ODDT }))
	add("os_dt", "OS Distance PD", "lab_ticket_powers", true, hp && p.OSDT != nil && *p.OSDT != "", valP(hp, p, func() interface{} { return p.OSDT }))
	add("od_seg_hd", "OD Fitting Height", "lab_ticket_powers", false, hp && p.ODSegHD != nil && *p.ODSegHD != "", valP(hp, p, func() interface{} { return p.ODSegHD }))
	add("os_seg_hd", "OS Fitting Height", "lab_ticket_powers", false, hp && p.OSSegHD != nil && *p.OSSegHD != "", valP(hp, p, func() interface{} { return p.OSSegHD }))
	add("od_bvd", "OD Back Vertex Distance", "lab_ticket_powers", false, hp && p.ODBVD != nil && *p.ODBVD != "", valP(hp, p, func() interface{} { return p.ODBVD }))
	add("os_bvd", "OS Back Vertex Distance", "lab_ticket_powers", false, hp && p.OSBVD != nil && *p.OSBVD != "", valP(hp, p, func() interface{} { return p.OSBVD }))

	// ── Prism ──
	add("od_h_prism", "OD Horiz Prism", "lab_ticket_powers", false, hp && p.ODHPrism != nil, valP(hp, p, func() interface{} { return p.ODHPrism }))
	add("os_h_prism", "OS Horiz Prism", "lab_ticket_powers", false, hp && p.OSHPrism != nil, valP(hp, p, func() interface{} { return p.OSHPrism }))

	// ── Frame ──
	f := ticket.Frame
	hf := f != nil
	add("size_lens_width", "Frame Eye Size (A)", "lab_ticket_frame", true, hf && f.SizeLensWidth != nil && *f.SizeLensWidth != "", valF(hf, func() interface{} { return f.SizeLensWidth }))
	add("b_value", "Frame B", "lab_ticket_frame", true, hf && f.BValue != nil, valF(hf, func() interface{} { return f.BValue }))
	add("size_bridge_width", "Frame DBL", "lab_ticket_frame", true, hf && f.SizeBridgeWidth != nil && *f.SizeBridgeWidth != "", valF(hf, func() interface{} { return f.SizeBridgeWidth }))
	add("panto", "Pantoscopic Angle", "lab_ticket_frame", false, hf && f.Panto != nil, valF(hf, func() interface{} { return f.Panto }))
	add("wrap_angle", "Frame Bow Angle", "lab_ticket_frame", false, hf && f.WrapAngle != nil, valF(hf, func() interface{} { return f.WrapAngle }))

	// ── Special Instructions ──
	add("lab_instructions", "Special Instructions", "lab_ticket", false, ticket.LabInstructions != nil && *ticket.LabInstructions != "", ticket.LabInstructions)

	// Check readiness
	ready := true
	for _, f := range fields {
		if f.Required && !f.Filled {
			ready = false
			break
		}
	}

	return &ZeissOrderRequirements{Ready: ready, Fields: fields}
}

func (s *CatalogService) trySyncLensFromInvoice(invoiceID int64, lens *labTicketModel.LabTicketLens) (interface{}, bool) {
	type lensRow struct {
		ItemID int64
	}
	var lr lensRow
	s.db.Table("invoice_item_sale").Select("item_id").
		Where("invoice_id = ? AND item_type IN ? AND item_id IS NOT NULL", invoiceID, []string{"Lens", "Lenses"}).
		Limit(1).Scan(&lr)
	if lr.ItemID == 0 {
		return nil, false
	}

	var ln struct {
		IDLenses       int64
		VwDesignCode   *string `gorm:"column:vw_design_code"`
		VwMaterialCode *string `gorm:"column:vw_material_code"`
		Source         *string
	}
	if s.db.Table("lenses").Where("id_lenses = ?", lr.ItemID).Scan(&ln).Error != nil || ln.IDLenses == 0 {
		return nil, false
	}
	if ln.Source == nil || *ln.Source != "zeiss_only" {
		return nil, false
	}
	if ln.VwDesignCode == nil || *ln.VwDesignCode == "" {
		return nil, false
	}

	// Persist
	lid := int(ln.IDLenses)
	lens.LensesID = &lid
	lens.VwDesignCode = ln.VwDesignCode
	lens.VwMaterialCode = ln.VwMaterialCode
	s.db.Model(lens).Updates(map[string]interface{}{
		"lenses_id":        ln.IDLenses,
		"vw_design_code":   ln.VwDesignCode,
		"vw_material_code": ln.VwMaterialCode,
	})

	return *ln.VwDesignCode, true
}

func valP(has bool, p *labTicketModel.LabTicketPowers, fn func() interface{}) interface{} {
	if !has || p == nil {
		return nil
	}
	return fn()
}

func valF(has bool, fn func() interface{}) interface{} {
	if !has {
		return nil
	}
	return fn()
}

func init() {
	_ = fmt.Sprintf // ensure fmt is used
}
