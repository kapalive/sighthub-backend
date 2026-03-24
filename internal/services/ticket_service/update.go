package ticket_service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/frames"
	"sighthub-backend/internal/models/inventory"
	invoiceModel "sighthub-backend/internal/models/invoices"
	labTicketModel "sighthub-backend/internal/models/lab_ticket"
	locationModel "sighthub-backend/internal/models/location"
	"sighthub-backend/internal/models/prescriptions"
	serviceModel "sighthub-backend/internal/models/service"
	"sighthub-backend/internal/models/types"
	"sighthub-backend/internal/models/vendors"
	"sighthub-backend/pkg/activitylog"
	pkgSKU "sighthub-backend/pkg/sku"
)

// ── Request types ────────────────────────────────────────────────────────────

type UpdateTicketRequest struct {
	NumberTicket      *string `json:"number_ticket"`
	DateCreate        *string `json:"date_create"`
	Date              *string `json:"date"` // alias
	DatePromise       *string `json:"date_promise"`
	LabID             *int    `json:"lab_id"`
	LabTicketStatusID *int64  `json:"lab_ticket_status_id"`
	OrdersLensID      *int64  `json:"orders_lens_id"`
	Tray              *string `json:"tray"`
	Amt               *string `json:"amt"`
	OurNote           *string `json:"our_note"`
	LabInstructions   *string `json:"lab_instructions"`
	Notified          *string `json:"notified"`
	ShipTo            *string `json:"ship_to"`
	IDRX              *int64  `json:"id_rx"`

	// Nested objects for glasses
	Powers *PowersPayload `json:"powers"`
	Lens   *LensPayload   `json:"lens"`
	Frame  *FramePayload  `json:"frame"`

	// Nested for contacts
	ContactPowers *json.RawMessage `json:"contact_powers"`
	TicketContact *json.RawMessage `json:"ticket_contact"`
}

type PowersPayload struct {
	ODSph       *string  `json:"od_sph"`
	OSSph       *string  `json:"os_sph"`
	ODCyl       *string  `json:"od_cyl"`
	OSCyl       *string  `json:"os_cyl"`
	ODAxis      *string  `json:"od_axis"`
	OSAxis      *string  `json:"os_axis"`
	ODAdd       *float64 `json:"od_add"`
	OSAdd       *float64 `json:"os_add"`
	ODHPrism    *float64 `json:"od_h_prism"`
	ODHPrismDir *string  `json:"od_h_prism_direction"`
	OSHPrism    *float64 `json:"os_h_prism"`
	OSHPrismDir *string  `json:"os_h_prism_direction"`
	ODVPrism    *float64 `json:"od_v_prism"`
	ODVPrismDir *string  `json:"od_v_prism_direction"`
	OSVPrism    *float64 `json:"os_v_prism"`
	OSVPrismDir *string  `json:"os_v_prism_direction"`
	ODSegHD     *string  `json:"od_seg_hd"`
	OSSegHD     *string  `json:"os_seg_hd"`
	ODOC        *string  `json:"od_oc"`
	OSOC        *string  `json:"os_oc"`
	ODBC        *string  `json:"od_bc"`
	OSBC        *string  `json:"os_bc"`
	ODDT        *string  `json:"od_dt"`
	OSDT        *string  `json:"os_dt"`
	ODNR        *string  `json:"od_nr"`
	OSNR        *string  `json:"os_nr"`
	OUDT        *string  `json:"ou_dt"`
	OUNR        *string  `json:"ou_nr"`
}

type LensPayload struct {
	EdgeThickness         *string `json:"edge_thickness"`
	CenterThickness       *string `json:"center_thickness"`
	LensSafetyThicknessID *int    `json:"lens_safety_thickness_id"`
	LensEdgeID            *int    `json:"lens_edge_id"`
	LensStatus            *string `json:"lens_status"`
	LensOrder             *string `json:"lens_order"`
	LensType              *string `json:"lens_type"`
	LensesMaterial        *string `json:"lenses_material"`
	NotesColor            *string `json:"notes_color"`
	LensesID              *int    `json:"lenses_id"`
}

type FramePayload struct {
	// POF or stock
	POF    *bool  `json:"pof"`
	ItemID *int64 `json:"item_id"` // inventory_id
	SKU    *string `json:"sku"`

	// Manual frame fields
	FrameName           *string  `json:"frame_name"`
	BrandName           *string  `json:"brand_name"`
	VendorName          *string  `json:"vendor_name"`
	ManufacturerName    *string  `json:"manufacturer_name"`
	MaterialsFrame      *string  `json:"materials_frame"`
	MaterialsTemple     *string  `json:"materials_temple"`
	Color               *string  `json:"color"`
	SizeLensWidth       *string  `json:"size_lens_width"`
	SizeBridgeWidth     *string  `json:"size_bridge_width"`
	SizeTempleLength    *string  `json:"size_temple_length"`
	ModelTitleVariant   *string  `json:"model_title_variant"`
	FrameTypeMaterialID *int     `json:"frame_type_material_id"`
	FrameShapeID        *int     `json:"frame_shape_id"`
	Status              *string  `json:"status"`
	DropShip            *bool    `json:"drop_ship"`
	ShipTo              *string  `json:"ship_to"`

	// Measurements (accept both string and number from frontend)
	BValue    *json.Number `json:"b_value"`
	EDValue   *json.Number `json:"ed_value"`
	CircValue *json.Number `json:"circ_value"`
	AValue    *json.Number `json:"a_value"`
	BDim      *json.Number `json:"b_dim"`
	Circum    *json.Number `json:"circum"`
	ED        *json.Number `json:"ed"`

	// Optical
	Panto          *float64 `json:"panto"`
	WrapAngle      *float64 `json:"wrap_angle"`
	HeadEyeRatio   *float64 `json:"head_eye_ratio"`
	StabilityCoeff *float64 `json:"stability_coeff"`
	BC             *float64 `json:"bc"`
	HeadCape       *string  `json:"head_cape"`
	CorridorR      *string  `json:"corridor_r"`
	CorridorL      *string  `json:"corridor_l"`

	// Aliases
	DBL    *string `json:"dbl"`
	Temple *string `json:"temple"`
}

// jsonNumToInt converts json.Number to *int, returns nil if nil or invalid
func jsonNumToInt(n *json.Number) *int {
	if n == nil {
		return nil
	}
	s := n.String()
	v, err := strconv.Atoi(s)
	if err != nil {
		// try parsing as float then truncate
		if f, err2 := strconv.ParseFloat(s, 64); err2 == nil {
			vi := int(f)
			return &vi
		}
		return nil
	}
	return &v
}

// ── Update ───────────────────────────────────────────────────────────────────

func (s *Service) UpdateTicket(username string, ticketID int64, req *UpdateTicketRequest) (map[string]interface{}, error) {
	emp, loc, err := s.empLocation(username)
	if err != nil {
		return nil, err
	}

	var ticket labTicketModel.LabTicket
	if err := s.db.Preload("Powers").Preload("Lens").Preload("Frame").
		Preload("PowersContact").Preload("Contact").
		First(&ticket, ticketID).Error; err != nil {
		return nil, fmt.Errorf("ticket %d not found", ticketID)
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// ── Common fields ────────────────────────────────────────────────────
	if req.NumberTicket != nil {
		ticket.NumberTicket = *req.NumberTicket
	}
	if req.DateCreate != nil {
		if t, err := time.Parse("2006-01-02", *req.DateCreate); err == nil {
			ticket.DateCreate = &t
		}
	} else if req.Date != nil {
		if t, err := time.Parse("2006-01-02", *req.Date); err == nil {
			ticket.DateCreate = &t
		}
	}
	if req.DatePromise != nil {
		if t, err := time.Parse("2006-01-02", *req.DatePromise); err == nil {
			ticket.DatePromise = &t
		}
	}
	if req.LabID != nil {
		ticket.LabID = req.LabID
	}
	if req.LabTicketStatusID != nil {
		ticket.LabTicketStatusID = *req.LabTicketStatusID
	}
	if req.OrdersLensID != nil {
		ticket.OrdersLensID = req.OrdersLensID
	}
	if req.Tray != nil {
		ticket.Tray = req.Tray
	}
	if req.Amt != nil {
		ticket.Amt = req.Amt
	}
	if req.OurNote != nil {
		ticket.OurNote = req.OurNote
	}
	if req.LabInstructions != nil {
		ticket.LabInstructions = req.LabInstructions
	}
	if req.Notified != nil {
		ticket.Notified = req.Notified
	}
	if req.ShipTo != nil {
		ticket.ShipTo = req.ShipTo
	}

	ticketType := ""
	if ticket.GOrC != nil {
		ticketType = *ticket.GOrC
	}

	// ── Tint check ───────────────────────────────────────────────────────
	tintAllowed := false
	if ticket.InvoiceID != 0 {
		var svcItems []invoiceModel.InvoiceServicesItem
		if err := tx.Where("invoice_id = ?", ticket.InvoiceID).Find(&svcItems).Error; err == nil {
			for _, si := range svcItems {
				if si.AdditionalServiceID != nil {
					var addSvc serviceModel.AdditionalService
					if err := tx.First(&addSvc, *si.AdditionalServiceID).Error; err == nil {
						if addSvc.Tint != nil && *addSvc.Tint {
							tintAllowed = true
							break
						}
					}
				}
			}
		}
	}

	// ── GLASSES branch ───────────────────────────────────────────────────
	if ticketType == "g" {
		// Powers from rx
		if req.IDRX != nil && ticket.Powers != nil {
			s.refillPowersFromRx(tx, ticket.Powers, *req.IDRX)
		}
		// Powers from payload
		if req.Powers != nil && ticket.Powers != nil {
			s.patchPowers(ticket.Powers, req.Powers)
			tx.Save(ticket.Powers)
		}

		// Lens
		if req.Lens != nil && ticket.Lens != nil {
			s.patchLens(ticket.Lens, req.Lens, tintAllowed)
			tx.Save(ticket.Lens)
		}

		// Frame
		if req.Frame != nil && ticket.Frame != nil {
			if err := s.applyFramePayload(tx, &ticket, req.Frame, emp, loc); err != nil {
				tx.Rollback()
				return nil, err
			}
			tx.Save(ticket.Frame)
		}
	}

	// ── CONTACTS branch ──────────────────────────────────────────────────
	if ticketType == "c" {
		if req.IDRX != nil && ticket.PowersContact != nil {
			s.refillContactFromRx(tx, ticket.PowersContact, *req.IDRX)
		}
		if req.ContactPowers != nil && ticket.PowersContact != nil {
			// Decode and patch
			var cp PowersPayload
			if json.Unmarshal(*req.ContactPowers, &cp) == nil {
				s.patchContactPowers(ticket.PowersContact, &cp)
			}
			tx.Save(ticket.PowersContact)
		}
	}

	tx.Save(&ticket)

	// Activity log
	employeeID := int64(emp.IDEmployee)
	details := map[string]interface{}{
		"number_ticket": ticket.NumberTicket,
		"invoice_id":    ticket.InvoiceID,
	}
	detailsJSON, _ := json.Marshal(details)
	_ = activitylog.Log(tx, "ticket", "update",
		activitylog.WithEmployee(employeeID),
		activitylog.WithLocation(loc.IDLocation),
		activitylog.WithEntity(ticket.IDLabTicket),
		activitylog.WithDetailsRaw(detailsJSON),
	)

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return map[string]interface{}{
		"message":    "LabTicket updated successfully",
		"lab_ticket": ticket.ToMap(),
	}, nil
}

// ── Powers helpers ───────────────────────────────────────────────────────────

func (s *Service) refillPowersFromRx(tx *gorm.DB, powers *labTicketModel.LabTicketPowers, rxID int64) {
	var pp prescriptions.PatientPrescription
	if err := tx.Where("id_patient_prescription = ?", rxID).First(&pp).Error; err != nil {
		return
	}
	var g prescriptions.GlassesPrescription
	if err := tx.Where("prescription_id = ?", pp.IDPatientPrescription).First(&g).Error; err != nil {
		return
	}
	powers.ODSph = g.OdSph
	powers.OSSph = g.OsSph
	powers.ODCyl = g.OdCyl
	powers.OSCyl = g.OsCyl
	powers.ODAxis = g.OdAxis
	powers.OSAxis = g.OsAxis
	powers.ODAdd = g.OdAdd
	powers.OSAdd = g.OsAdd
	powers.ODHPrism = g.OdHPrism
	powers.OSHPrism = g.OsHPrism
	powers.ODVPrism = g.OdVPrism
	powers.OSVPrism = g.OsVPrism
	if g.OdHPrismDirection != nil {
		d := labTicketModel.HPrismDirection(*g.OdHPrismDirection)
		powers.ODHPrismDirection = &d
	}
	if g.OsHPrismDirection != nil {
		d := labTicketModel.HPrismDirection(*g.OsHPrismDirection)
		powers.OSHPrismDirection = &d
	}
	if g.OdVPrismDirection != nil {
		d := labTicketModel.VPrismDirection(*g.OdVPrismDirection)
		powers.ODVPrismDirection = &d
	}
	if g.OsVPrismDirection != nil {
		d := labTicketModel.VPrismDirection(*g.OsVPrismDirection)
		powers.OSVPrismDirection = &d
	}
	if g.OdDpd != nil {
		str := fmt.Sprintf("%.2f", *g.OdDpd)
		powers.ODDT = &str
	}
	if g.OsDpd != nil {
		str := fmt.Sprintf("%.2f", *g.OsDpd)
		powers.OSDT = &str
	}
	tx.Save(powers)
}

func (s *Service) patchPowers(p *labTicketModel.LabTicketPowers, pl *PowersPayload) {
	overrideStrPtr(&p.ODSph, pl.ODSph)
	overrideStrPtr(&p.OSSph, pl.OSSph)
	overrideStrPtr(&p.ODCyl, pl.ODCyl)
	overrideStrPtr(&p.OSCyl, pl.OSCyl)
	overrideStrPtr(&p.ODAxis, pl.ODAxis)
	overrideStrPtr(&p.OSAxis, pl.OSAxis)
	overrideF64Ptr(&p.ODAdd, pl.ODAdd)
	overrideF64Ptr(&p.OSAdd, pl.OSAdd)
	overrideF64Ptr(&p.ODHPrism, pl.ODHPrism)
	overrideF64Ptr(&p.OSHPrism, pl.OSHPrism)
	overrideHPrismDir(&p.ODHPrismDirection, pl.ODHPrismDir)
	overrideHPrismDir(&p.OSHPrismDirection, pl.OSHPrismDir)
	overrideF64Ptr(&p.ODVPrism, pl.ODVPrism)
	overrideF64Ptr(&p.OSVPrism, pl.OSVPrism)
	overrideVPrismDir(&p.ODVPrismDirection, pl.ODVPrismDir)
	overrideVPrismDir(&p.OSVPrismDirection, pl.OSVPrismDir)
	overrideStrPtr(&p.ODSegHD, pl.ODSegHD)
	overrideStrPtr(&p.OSSegHD, pl.OSSegHD)
	overrideStrPtr(&p.ODOC, pl.ODOC)
	overrideStrPtr(&p.OSOC, pl.OSOC)
	overrideStrPtr(&p.ODBC, pl.ODBC)
	overrideStrPtr(&p.OSBC, pl.OSBC)
	overrideStrPtr(&p.ODDT, pl.ODDT)
	overrideStrPtr(&p.OSDT, pl.OSDT)
	overrideStrPtr(&p.ODNR, pl.ODNR)
	overrideStrPtr(&p.OSNR, pl.OSNR)
	overrideStrPtr(&p.OUDT, pl.OUDT)
	overrideStrPtr(&p.OUNR, pl.OUNR)
}

// ── Lens helpers ─────────────────────────────────────────────────────────────

func (s *Service) patchLens(lens *labTicketModel.LabTicketLens, pl *LensPayload, tintAllowed bool) {
	if pl.LensStatus != nil {
		lens.LensStatus = pl.LensStatus
	}
	if pl.LensOrder != nil {
		lo := labTicketModel.LabTicketLensOrder(*pl.LensOrder)
		lens.LensOrder = &lo
	}
	if pl.EdgeThickness != nil {
		lens.EdgeThickness = pl.EdgeThickness
	}
	if pl.CenterThickness != nil {
		lens.CenterThickness = pl.CenterThickness
	}
	if pl.LensSafetyThicknessID != nil {
		lens.LensSafetyThicknessID = pl.LensSafetyThicknessID
	}
	if pl.LensEdgeID != nil {
		lens.LensEdgeID = pl.LensEdgeID
	}
	if pl.NotesColor != nil {
		lens.NotesColor = pl.NotesColor
	}
	if tintAllowed && pl.NotesColor != nil {
		lens.NotesColor = pl.NotesColor
	}

	// Lens catalog reference — auto-fill VW codes
	if pl.LensesID != nil {
		lens.LensesID = pl.LensesID
		// Look up VW codes from lenses table
		type vwRow struct {
			VwDesignCode   *string
			VwMaterialCode *string
		}
		var vw vwRow
		s.db.Table("lenses").Select("vw_design_code, vw_material_code").
			Where("id_lenses = ?", *pl.LensesID).Scan(&vw)
		lens.VwDesignCode = vw.VwDesignCode
		lens.VwMaterialCode = vw.VwMaterialCode
	}
}

// ── Frame helpers ────────────────────────────────────────────────────────────

func (s *Service) applyFramePayload(tx *gorm.DB, ticket *labTicketModel.LabTicket, fp *FramePayload, emp *employees.Employee, loc *locationModel.Location) error {
	frame := ticket.Frame

	// Sizes with aliases (string fields on frame)
	if fp.SizeLensWidth != nil {
		frame.SizeLensWidth = fp.SizeLensWidth
	} else if fp.AValue != nil {
		s := fp.AValue.String()
		frame.SizeLensWidth = &s
	}
	if fp.SizeBridgeWidth != nil {
		frame.SizeBridgeWidth = fp.SizeBridgeWidth
	} else if fp.DBL != nil {
		frame.SizeBridgeWidth = fp.DBL
	}
	if fp.SizeTempleLength != nil {
		frame.SizeTempleLength = fp.SizeTempleLength
	} else if fp.Temple != nil {
		frame.SizeTempleLength = fp.Temple
	}

	// B, ED, Circ with aliases (int fields on frame)
	if fp.BValue != nil {
		frame.BValue = jsonNumToInt(fp.BValue)
	} else if fp.BDim != nil {
		frame.BValue = jsonNumToInt(fp.BDim)
	}
	if fp.EDValue != nil {
		frame.EDValue = jsonNumToInt(fp.EDValue)
	} else if fp.ED != nil {
		frame.EDValue = jsonNumToInt(fp.ED)
	}
	if fp.CircValue != nil {
		frame.CircValue = jsonNumToInt(fp.CircValue)
	} else if fp.Circum != nil {
		frame.CircValue = jsonNumToInt(fp.Circum)
	}

	// Simple fields
	if fp.FrameName != nil {
		frame.FrameName = fp.FrameName
	}
	if fp.BrandName != nil {
		frame.BrandName = fp.BrandName
	}
	if fp.MaterialsFrame != nil {
		frame.MaterialsFrame = fp.MaterialsFrame
	}
	if fp.MaterialsTemple != nil {
		frame.MaterialsTemple = fp.MaterialsTemple
	}
	if fp.Color != nil {
		frame.Color = fp.Color
	}
	if fp.ModelTitleVariant != nil {
		frame.ModelTitleVariant = fp.ModelTitleVariant
	}
	if fp.FrameShapeID != nil {
		frame.FrameShapeID = fp.FrameShapeID
	}
	if fp.FrameTypeMaterialID != nil {
		frame.FrameTypeMaterialID = fp.FrameTypeMaterialID
	}
	if fp.Status != nil {
		frame.Status = fp.Status
	}
	if fp.VendorName != nil {
		frame.VendorName = fp.VendorName
	}
	if fp.ManufacturerName != nil {
		frame.ManufacturerName = fp.ManufacturerName
	}
	if fp.DropShip != nil {
		frame.DropShip = *fp.DropShip
	}
	if fp.ShipTo != nil {
		frame.ShipTo = fp.ShipTo
	}
	if fp.Panto != nil {
		frame.Panto = fp.Panto
	}
	if fp.WrapAngle != nil {
		frame.WrapAngle = fp.WrapAngle
	}
	if fp.HeadEyeRatio != nil {
		frame.HeadEyeRatio = fp.HeadEyeRatio
	}
	if fp.StabilityCoeff != nil {
		frame.StabilityCoeff = fp.StabilityCoeff
	}
	if fp.BC != nil {
		frame.BC = fp.BC
	}
	if fp.HeadCape != nil {
		frame.HeadCape = fp.HeadCape
	}
	if fp.CorridorR != nil {
		frame.CorridorR = fp.CorridorR
	}
	if fp.CorridorL != nil {
		frame.CorridorL = fp.CorridorL
	}

	// Inventory assignment (item_id or sku)
	if fp.ItemID != nil || fp.SKU != nil {
		return s.assignInventoryToFrame(tx, ticket, fp, emp, loc)
	}

	return nil
}

func (s *Service) assignInventoryToFrame(tx *gorm.DB, ticket *labTicketModel.LabTicket, fp *FramePayload, emp *employees.Employee, loc *locationModel.Location) error {
	frame := ticket.Frame
	locID := int64(loc.IDLocation)
	employeeID := int64(emp.IDEmployee)

	var inv inventory.Inventory
	if fp.ItemID != nil {
		if err := tx.Where("id_inventory = ?", *fp.ItemID).First(&inv).Error; err != nil {
			// Try as SKU
			skuStr := fmt.Sprintf("%d", *fp.ItemID)
			norm := pkgSKU.Normalize(skuStr)
			if err := tx.Where("sku = ? OR sku = ?", skuStr, norm).First(&inv).Error; err != nil {
				return fmt.Errorf("inventory item not found")
			}
		}
	} else if fp.SKU != nil {
		norm := pkgSKU.Normalize(*fp.SKU)
		if err := tx.Where("sku = ? OR sku = ?", *fp.SKU, norm).First(&inv).Error; err != nil {
			return fmt.Errorf("inventory item not found")
		}
	}

	st := string(inv.StatusItemsInventory)
	if st != "Ready for Sale" && st != "SOLD" {
		return fmt.Errorf("item %d has invalid status '%s'", inv.IDInventory, st)
	}
	if st == "SOLD" && inv.InvoiceID != ticket.InvoiceID {
		return fmt.Errorf("item %d already sold on another invoice", inv.IDInventory)
	}

	// Fill frame from Model
	var mdl frames.Model
	if err := tx.Preload("Product").First(&mdl, inv.ModelID).Error; err == nil {
		if mdl.Product != nil {
			fn := mdl.Product.TitleProduct + " " + mdl.TitleVariant
			if fp.FrameName == nil {
				frame.FrameName = &fn
			}
			if fp.ModelTitleVariant == nil {
				frame.ModelTitleVariant = &mdl.TitleVariant
			}
			if fp.BrandName == nil && mdl.Product.BrandID != nil {
				var brand vendors.Brand
				if tx.First(&brand, *mdl.Product.BrandID).Error == nil {
					frame.BrandName = brand.BrandName
				}
			}
			if fp.MaterialsFrame == nil {
				frame.MaterialsFrame = mdl.MaterialsFrame
			}
			if fp.MaterialsTemple == nil {
				frame.MaterialsTemple = mdl.MaterialsTemple
			}
			if fp.Color == nil {
				frame.Color = mdl.Color
			}
			if frame.SizeLensWidth == nil {
				frame.SizeLensWidth = mdl.SizeLensWidth
			}
			if frame.SizeBridgeWidth == nil {
				frame.SizeBridgeWidth = mdl.SizeBridgeWidth
			}
			if frame.SizeTempleLength == nil {
				frame.SizeTempleLength = mdl.SizeTempleLength
			}
			if mdl.Product.VendorID != nil {
				var vendor vendors.Vendor
				if tx.First(&vendor, *mdl.Product.VendorID).Error == nil {
					frame.VendorName = &vendor.VendorName
				}
			}
		}
	}

	// Set status
	if st == "SOLD" {
		s := "Frame in Store"
		frame.Status = &s
	} else {
		s := "Frame in Store"
		frame.Status = &s
	}

	// Add to invoice if not already there
	var invoice invoiceModel.Invoice
	if err := tx.First(&invoice, ticket.InvoiceID).Error; err == nil {
		var existCnt int64
		tx.Model(&invoiceModel.InvoiceItemSale{}).
			Where("invoice_id = ? AND item_type = ? AND item_id = ?", invoice.IDInvoice, "Frames", inv.IDInventory).
			Count(&existCnt)

		if existCnt == 0 {
			var pb inventory.PriceBook
			price := 0.0
			if tx.Where("inventory_id = ?", inv.IDInventory).First(&pb).Error == nil {
				if pb.PbSellingPrice != nil {
					price = *pb.PbSellingPrice
				}
			}

			desc := "Frame"
			if frame.FrameName != nil {
				desc = *frame.FrameName
			}
			newItem := invoiceModel.InvoiceItemSale{
				InvoiceID:   invoice.IDInvoice,
				ItemType:    "Frames",
				ItemID:      &inv.IDInventory,
				Description: desc,
				Quantity:    1,
				Price:       price,
				Total:       price,
				Taxable:     boolPtr(false),
			}
			tx.Create(&newItem)

			// Warehouse transfer if needed
			oldInvoiceID := inv.InvoiceID
			if loc.WarehouseID != nil && inv.LocationID == int64(*loc.WarehouseID) {
				whID := int64(*loc.WarehouseID)
				tx.Create(&inventory.InventoryTransaction{
					InventoryID:     &inv.IDInventory,
					FromLocationID:  &whID,
					ToLocationID:    &locID,
					TransferredBy:   employeeID,
					StatusItems:     types.StatusItemsInventory("TRANSFERRED TO SHOWCASE"),
					TransactionType: "Transfer",
					DateTransaction: time.Now(),
				})
				inv.LocationID = locID
			}

			// Mark as SOLD
			inv.StatusItemsInventory = types.StatusInventorySOLD
			inv.InvoiceID = invoice.IDInvoice
			tx.Save(&inv)

			// Sale transaction
			tx.Create(&inventory.InventoryTransaction{
				InventoryID:     &inv.IDInventory,
				FromLocationID:  &locID,
				TransferredBy:   employeeID,
				InvoiceID:       &invoice.IDInvoice,
				OldInvoiceID:    &oldInvoiceID,
				StatusItems:     types.StatusInventorySOLD,
				TransactionType: "Sale",
				DateTransaction: time.Now(),
			})

			// Recalculate invoice
			invoice.TotalAmount += price
			if invoice.Discount == nil {
				z := 0.0
				invoice.Discount = &z
			}
			invoice.FinalAmount = roundTo2(invoice.TotalAmount - *invoice.Discount)
			gcBal := 0.0
			if invoice.GiftCardBal != nil {
				gcBal = *invoice.GiftCardBal
			}
			invoice.Due = roundTo2(invoice.FinalAmount - invoice.PTBal - invoice.InsBal - gcBal)
			tx.Save(&invoice)
		}
	}

	return nil
}

// ── Contact helpers ──────────────────────────────────────────────────────────

func (s *Service) refillContactFromRx(tx *gorm.DB, pc *labTicketModel.LabTicketPowersContact, rxID int64) {
	var pp prescriptions.PatientPrescription
	if tx.Where("id_patient_prescription = ?", rxID).First(&pp).Error != nil {
		return
	}
	var cl prescriptions.ContactLensPrescription
	if tx.Where("prescription_id = ?", pp.IDPatientPrescription).First(&cl).Error != nil {
		return
	}
	pc.ODContLens = cl.OdContLens
	pc.OSContLens = cl.OsContLens
	pc.ODBC = cl.OdBc
	pc.OSBC = cl.OsBc
	pc.ODDia = cl.OdDia
	pc.OSDia = cl.OsDia
	pc.ODPwr = cl.OdPwr
	pc.OSPwr = cl.OsPwr
	pc.ODCyl = cl.OdCyl
	pc.OSCyl = cl.OsCyl
	pc.ODAxis = cl.OdAxis
	pc.OSAxis = cl.OsAxis
	pc.ODAdd = cl.OdAdd
	pc.OSAdd = cl.OsAdd
	pc.ODColor = cl.OdColor
	pc.OSColor = cl.OsColor
	if cl.ExpirationDate != nil {
		pc.ExpirationDate = cl.ExpirationDate
	}
	tx.Save(pc)
}

func (s *Service) patchContactPowers(pc *labTicketModel.LabTicketPowersContact, pl *PowersPayload) {
	overrideStrPtr(&pc.ODContLens, pl.ODSph) // reuse struct, map appropriately
	overrideStrPtr(&pc.OSContLens, pl.OSSph)
	overrideStrPtr(&pc.ODBC, pl.ODBC)
	overrideStrPtr(&pc.OSBC, pl.OSBC)
	overrideStrPtr(&pc.ODPwr, pl.ODAxis)
	overrideStrPtr(&pc.OSPwr, pl.OSAxis)
	overrideStrPtr(&pc.ODCyl, pl.ODCyl)
	overrideStrPtr(&pc.OSCyl, pl.OSCyl)
	overrideStrPtr(&pc.ODAxis, pl.ODAxis)
	overrideStrPtr(&pc.OSAxis, pl.OSAxis)
	overrideStrPtr(&pc.ODAdd, pl.ODDT)
	overrideStrPtr(&pc.OSAdd, pl.OSDT)
}

func strPtr(s string) *string { return &s }
