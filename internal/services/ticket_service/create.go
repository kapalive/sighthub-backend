package ticket_service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	"sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/frames"
	"sighthub-backend/internal/models/inventory"
	invoiceModel "sighthub-backend/internal/models/invoices"
	labTicketModel "sighthub-backend/internal/models/lab_ticket"
	"sighthub-backend/internal/models/lenses"
	locationModel "sighthub-backend/internal/models/location"
	"sighthub-backend/internal/models/prescriptions"
	serviceModel "sighthub-backend/internal/models/service"
	"sighthub-backend/internal/models/types"
	"sighthub-backend/internal/models/vendors"
	"sighthub-backend/pkg/activitylog"
	"sighthub-backend/pkg/sku"
)

// CreateTicketRequest holds the JSON body for POST /api/ticket/invoice/{invoice_id}.
type CreateTicketRequest struct {
	GOrC              string `json:"g_or_c"`
	NumberTicket      string `json:"number_ticket"`
	LabID             *int   `json:"lab_id"`
	LabTicketStatusID *int64 `json:"lab_ticket_status_id"`
	OrdersLensID      *int64 `json:"orders_lens_id"`
	IDRX              *int64 `json:"id_rx"`

	// Frame
	POF                 bool    `json:"pof"`
	ItemID              *int64  `json:"item_id"`
	BValue              *int    `json:"b_value"`
	EDValue             *int    `json:"ed_value"`
	CircValue           *int    `json:"circ_value"`
	FrameTypeMaterialID *int    `json:"frame_type_material_id"`
	FrameName           *string `json:"frame_name"`
	BrandName           *string `json:"brand_name"`
	MaterialsFrame      *string `json:"materials_frame"`
	MaterialsTemple     *string `json:"materials_temple"`
	Color               *string `json:"color"`
	SizeLensWidth       *string `json:"size_lens_width"`
	SizeBridgeWidth     *string `json:"size_bridge_width"`
	SizeTempleLength    *string `json:"size_temple_length"`

	// Lens
	LensStatus            *string `json:"lens_status"`
	LensOrder             *string `json:"lens_order"`
	EdgeThickness         *string `json:"edge_thickness"`
	CenterThickness       *string `json:"center_thickness"`
	LensSafetyThicknessID *int    `json:"lens_safety_thickness_id"`
	LensEdgeID            *int    `json:"lens_edge_id"`
	LensTypeColor         *string `json:"lens_type_color"`
	TintPercent           *int    `json:"tint_percent"`
	NotesColor            *string `json:"notes_color"`
	LensTintColorID       *int    `json:"lens_tint_color_id"`

	// Powers override (glasses)
	ODSph            *string  `json:"od_sph"`
	OSSph            *string  `json:"os_sph"`
	ODCyl            *string  `json:"od_cyl"`
	OSCyl            *string  `json:"os_cyl"`
	ODAxis           *string  `json:"od_axis"`
	OSAxis           *string  `json:"os_axis"`
	ODAdd            *float64 `json:"od_add"`
	OSAdd            *float64 `json:"os_add"`
	ODHPrism         *float64 `json:"od_h_prism"`
	ODHPrismDir      *string  `json:"od_h_prism_direction"`
	OSHPrism         *float64 `json:"os_h_prism"`
	OSHPrismDir      *string  `json:"os_h_prism_direction"`
	ODVPrism         *float64 `json:"od_v_prism"`
	ODVPrismDir      *string  `json:"od_v_prism_direction"`
	OSVPrism         *float64 `json:"os_v_prism"`
	OSVPrismDir      *string  `json:"os_v_prism_direction"`
	ODSegHD          *string  `json:"od_seg_hd"`
	OSSegHD          *string  `json:"os_seg_hd"`
	ODOC             *string  `json:"od_oc"`
	OSOC             *string  `json:"os_oc"`
	ODBC             *string  `json:"od_bc"`
	OSBC             *string  `json:"os_bc"`
	ODBVD            *string  `json:"od_bvd"`
	OSBVD            *string  `json:"os_bvd"`
	ODDT             *string  `json:"od_dt"`
	OSDT             *string  `json:"os_dt"`
	ODNR             *string  `json:"od_nr"`
	OSNR             *string  `json:"os_nr"`
	OUDT             *string  `json:"ou_dt"`
	OUNR             *string  `json:"ou_nr"`

	// Contacts (g_or_c == "c")
	ODContLens                    *string  `json:"od_cont_lens"`
	OSContLens                    *string  `json:"os_cont_lens"`
	ContactODBc                   *string  `json:"contact_od_bc"`
	ContactOSBc                   *string  `json:"contact_os_bc"`
	ODDia                         *float64 `json:"od_dia"`
	OSDia                         *float64 `json:"os_dia"`
	ODPwr                         *string  `json:"od_pwr"`
	OSPwr                         *string  `json:"os_pwr"`
	ContactODCyl                  *string  `json:"contact_od_cyl"`
	ContactOSCyl                  *string  `json:"contact_os_cyl"`
	ContactODAxis                 *string  `json:"contact_od_axis"`
	ContactOSAxis                 *string  `json:"contact_os_axis"`
	ContactODAdd                  *string  `json:"contact_od_add"`
	ContactOSAdd                  *string  `json:"contact_os_add"`
	ODColor                       *string  `json:"od_color"`
	OSColor                       *string  `json:"os_color"`
	ODType                        *string  `json:"od_type"`
	OSType                        *string  `json:"os_type"`
	ExpirationDate                *string  `json:"expiration_date"`
	ContactODHPrismDir            *string  `json:"contact_od_h_prism_direction"`
	ContactOSHPrismDir            *string  `json:"contact_os_h_prism_direction"`
	ContactODVPrismDir            *string  `json:"contact_od_v_prism_direction"`
	ContactOSVPrismDir            *string  `json:"contact_os_v_prism_direction"`
	LabTicketContactLensServicesID *int    `json:"lab_ticket_contact_lens_services_id"`
	ODAnnualSupply                *bool    `json:"od_annual_supply"`
	OSAnnualSupply                *bool    `json:"os_annual_supply"`
	ODTotalQty                    *int     `json:"od_total_qty"`
	OSTotalQty                    *int     `json:"os_total_qty"`
	Reasons                       *string  `json:"reasons"`
	Modality                      *string  `json:"modality"`
	BrandContactLensID            *int64   `json:"brand_contact_lens_id"`

	// Common
	Tray            *string `json:"tray"`
	Notified        *string `json:"notified"`
	Amt             *string `json:"amt"`
	OurNote         *string `json:"our_note"`
	LabInstructions *string `json:"lab_instructions"`
	DatePromise     *string `json:"date_promise"`
}

// CreateTicketResult is returned on success.
type CreateTicketResult struct {
	Message     string `json:"message"`
	IDLabTicket int64  `json:"id_lab_ticket"`
	GOrC        string `json:"g_or_c"`
}

// empLocation resolves the current user's employee + location from JWT username.
func (s *Service) empLocation(username string) (*employees.Employee, *locationModel.Location, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, nil, fmt.Errorf("login not found")
	}

	var emp employees.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, nil, fmt.Errorf("employee not found")
	}

	if emp.LocationID == nil {
		return nil, nil, fmt.Errorf("employee not assigned to a location")
	}

	var loc locationModel.Location
	if err := s.db.First(&loc, *emp.LocationID).Error; err != nil {
		return nil, nil, fmt.Errorf("location not found")
	}

	return &emp, &loc, nil
}

// CreateTicket creates a new LabTicket with associated powers/lens/frame (glasses)
// or powers-contact/contact (contacts), depending on g_or_c.
func (s *Service) CreateTicket(username string, invoiceID int64, req *CreateTicketRequest) (*CreateTicketResult, error) {
	// Resolve employee + location
	emp, loc, err := s.empLocation(username)
	if err != nil {
		return nil, err
	}
	employeeID := int64(emp.IDEmployee)

	// Load invoice
	var invoice invoiceModel.Invoice
	if err := s.db.First(&invoice, invoiceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("invoice with id %d not found", invoiceID)
		}
		return nil, err
	}
	if invoice.PatientID == nil {
		return nil, fmt.Errorf("invoice %d does not have a patient associated", invoiceID)
	}
	patientID := *invoice.PatientID

	// Defaults
	gOrC := req.GOrC
	if gOrC == "" {
		gOrC = "g"
	}
	numberTicket := req.NumberTicket
	if numberTicket == "" {
		numberTicket = fmt.Sprintf("T-%d", invoiceID)
	}
	labTicketStatusID := int64(2)
	if req.LabTicketStatusID != nil {
		labTicketStatusID = *req.LabTicketStatusID
	}
	now := time.Now()
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	var (
		labTicketPowersID        *int64
		labTicketLensID          *int64
		labTicketFrameID         *int64
		labTicketPowersContactID *int64
		labTicketContactID       *int64
	)

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if gOrC == "g" {
		// ======================== GLASSES ========================

		// --- POWERS ---
		powers := &labTicketModel.LabTicketPowers{}
		if req.IDRX != nil {
			var pp prescriptions.PatientPrescription
			if err := tx.Where("id_patient_prescription = ?", *req.IDRX).First(&pp).Error; err == nil {
				var g prescriptions.GlassesPrescription
				if err := tx.Where("prescription_id = ?", pp.IDPatientPrescription).First(&g).Error; err == nil {
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
					if g.OdHPrismDirection != nil {
						d := labTicketModel.HPrismDirection(*g.OdHPrismDirection)
						powers.ODHPrismDirection = &d
					}
					if g.OsHPrismDirection != nil {
						d := labTicketModel.HPrismDirection(*g.OsHPrismDirection)
						powers.OSHPrismDirection = &d
					}
					powers.ODVPrism = g.OdVPrism
					powers.OSVPrism = g.OsVPrism
					if g.OdVPrismDirection != nil {
						d := labTicketModel.VPrismDirection(*g.OdVPrismDirection)
						powers.ODVPrismDirection = &d
					}
					if g.OsVPrismDirection != nil {
						d := labTicketModel.VPrismDirection(*g.OsVPrismDirection)
						powers.OSVPrismDirection = &d
					}
					// od_dpd -> od_dt, os_dpd -> os_dt
					if g.OdDpd != nil {
						s := fmt.Sprintf("%.2f", *g.OdDpd)
						powers.ODDT = &s
					}
					if g.OsDpd != nil {
						s := fmt.Sprintf("%.2f", *g.OsDpd)
						powers.OSDT = &s
					}
				}
			}
		}
		// Override from request body
		overrideStrPtr(&powers.ODSph, req.ODSph)
		overrideStrPtr(&powers.OSSph, req.OSSph)
		overrideStrPtr(&powers.ODCyl, req.ODCyl)
		overrideStrPtr(&powers.OSCyl, req.OSCyl)
		overrideStrPtr(&powers.ODAxis, req.ODAxis)
		overrideStrPtr(&powers.OSAxis, req.OSAxis)
		overrideF64Ptr(&powers.ODAdd, req.ODAdd)
		overrideF64Ptr(&powers.OSAdd, req.OSAdd)
		overrideF64Ptr(&powers.ODHPrism, req.ODHPrism)
		overrideF64Ptr(&powers.OSHPrism, req.OSHPrism)
		overrideHPrismDir(&powers.ODHPrismDirection, req.ODHPrismDir)
		overrideHPrismDir(&powers.OSHPrismDirection, req.OSHPrismDir)
		overrideF64Ptr(&powers.ODVPrism, req.ODVPrism)
		overrideF64Ptr(&powers.OSVPrism, req.OSVPrism)
		overrideVPrismDir(&powers.ODVPrismDirection, req.ODVPrismDir)
		overrideVPrismDir(&powers.OSVPrismDirection, req.OSVPrismDir)
		overrideStrPtr(&powers.ODSegHD, req.ODSegHD)
		overrideStrPtr(&powers.OSSegHD, req.OSSegHD)
		overrideStrPtr(&powers.ODOC, req.ODOC)
		overrideStrPtr(&powers.OSOC, req.OSOC)
		overrideStrPtr(&powers.ODBC, req.ODBC)
		overrideStrPtr(&powers.OSBC, req.OSBC)
		overrideStrPtr(&powers.ODBVD, req.ODBVD)
		overrideStrPtr(&powers.OSBVD, req.OSBVD)
		overrideStrPtr(&powers.ODDT, req.ODDT)
		overrideStrPtr(&powers.OSDT, req.OSDT)
		overrideStrPtr(&powers.ODNR, req.ODNR)
		overrideStrPtr(&powers.OSNR, req.OSNR)
		overrideStrPtr(&powers.OUDT, req.OUDT)
		overrideStrPtr(&powers.OUNR, req.OUNR)

		if err := tx.Create(powers).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("create powers: %w", err)
		}
		labTicketPowersID = &powers.IDLabTicketPowers

		// --- LENS ---
		lensStatus := "Uncut"
		if req.LensStatus != nil {
			lensStatus = *req.LensStatus
		}
		lensOrderStr := "Pair"
		if req.LensOrder != nil {
			lensOrderStr = *req.LensOrder
		}
		lensOrder := labTicketModel.LabTicketLensOrder(lensOrderStr)

		ticketLens := &labTicketModel.LabTicketLens{
			LensStatus: &lensStatus,
			LensOrder:  &lensOrder,
		}

		// Find lens item in invoice (must be exactly one)
		var lensItems []invoiceModel.InvoiceItemSale
		tx.Where("invoice_id = ? AND item_type IN ?", invoiceID, []string{"Lens", "Lenses"}).
			Find(&lensItems)
		if len(lensItems) > 1 {
			tx.Rollback()
			return nil, fmt.Errorf("invoice has %d lens items — cannot create lab ticket (expected 1)", len(lensItems))
		}
		if len(lensItems) == 1 && lensItems[0].ItemID != nil {
			lensItem := lensItems[0]
			var ln lenses.Lenses
			if err := tx.First(&ln, *lensItem.ItemID).Error; err == nil {
				ticketLens.LensesMaterialsID = ln.LensesMaterialsID
				ticketLens.LensTypesID = ln.LensTypeID
				ticketLens.LensesID = &ln.IDLenses
				ticketLens.VwDesignCode = ln.VwDesignCode
				ticketLens.VwMaterialCode = ln.VwMaterialCode
			}
		}

		// Check if tint additional service exists
		tintAllowed := false
		var svcItems []invoiceModel.InvoiceServicesItem
		if err := tx.Where("invoice_id = ?", invoiceID).Find(&svcItems).Error; err == nil {
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

		// Accept fields from request
		ticketLens.EdgeThickness = req.EdgeThickness
		ticketLens.CenterThickness = req.CenterThickness
		ticketLens.LensSafetyThicknessID = req.LensSafetyThicknessID
		ticketLens.LensEdgeID = req.LensEdgeID
		ticketLens.NotesColor = req.NotesColor

		if tintAllowed {
			ticketLens.LensTypeColor = req.LensTypeColor
			ticketLens.TintPercent = req.TintPercent
			ticketLens.LensTintColorID = req.LensTintColorID
		}

		if err := tx.Create(ticketLens).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("create lens: %w", err)
		}
		labTicketLensID = &ticketLens.IDLabTicketLens

		// --- FRAME ---
		frameStatus := 1
		frame := &labTicketModel.LabTicketFrame{
			LabTicketStatus: frameStatus,
			BValue:          req.BValue,
			EDValue:         req.EDValue,
			CircValue:       req.CircValue,
		}

		var inv *inventory.Inventory

		if req.POF {
			// Patient Own Frame
			pofStr := "true"
			statusStr := "Patient Own Frame"
			frame.POF = &pofStr
			frame.Status = &statusStr
			frame.FrameName = req.FrameName
			frame.BrandName = req.BrandName
			frame.MaterialsFrame = req.MaterialsFrame
			frame.MaterialsTemple = req.MaterialsTemple
			frame.Color = req.Color
			frame.SizeLensWidth = req.SizeLensWidth
			frame.SizeBridgeWidth = req.SizeBridgeWidth
			frame.SizeTempleLength = req.SizeTempleLength
			frame.FrameTypeMaterialID = req.FrameTypeMaterialID
		} else {
			pofStr := "false"
			frame.POF = &pofStr

			if req.ItemID != nil {
				// Find inventory by id or SKU
				itemIDVal := *req.ItemID
				normalizedSKU := sku.Normalize(strconv.FormatInt(itemIDVal, 10))
				var foundInv inventory.Inventory
				if err := tx.Where("id_inventory = ? OR sku = ?", itemIDVal, normalizedSKU).First(&foundInv).Error; err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("frame not found")
				}
				inv = &foundInv

				// Verify item belongs to invoice
				var existingItem invoiceModel.InvoiceItemSale
				err := tx.Where("invoice_id = ? AND item_type = ? AND item_id = ?",
					invoiceID, "Frames", inv.IDInventory).First(&existingItem).Error

				if errors.Is(err, gorm.ErrRecordNotFound) {
					// Not in invoice yet -- validate status
					status := string(inv.StatusItemsInventory)
					if status != "Ready for Sale" && status != "SOLD" && status != "Ordered" {
						tx.Rollback()
						return nil, fmt.Errorf("item %d has invalid status '%s'", inv.IDInventory, status)
					}
					if status == "SOLD" && inv.InvoiceID != invoice.IDInvoice {
						tx.Rollback()
						return nil, fmt.Errorf("frame already sold on another invoice")
					}
				} else if err != nil {
					tx.Rollback()
					return nil, err
				}

				// Set status
				status := string(inv.StatusItemsInventory)
				if status == "SOLD" {
					s := "Frame in Store"
					frame.Status = &s
				} else if status == "Ordered" {
					s := "Ordered"
					frame.Status = &s
				} else {
					s := "Frame in Store"
					frame.Status = &s
				}

				// Fill from Model
				if inv.ModelID != nil {
					var mdl frames.Model
					if err := tx.Preload("Product").First(&mdl, *inv.ModelID).Error; err == nil {
						frame.ModelTitleVariant = &mdl.TitleVariant
						frame.MaterialsFrame = mdl.MaterialsFrame
						frame.MaterialsTemple = mdl.MaterialsTemple
						frame.Color = mdl.Color
						frame.SizeLensWidth = mdl.SizeLensWidth
						frame.SizeBridgeWidth = mdl.SizeBridgeWidth
						frame.SizeTempleLength = mdl.SizeTempleLength

						if mdl.Product != nil {
							fn := mdl.Product.TitleProduct + " " + mdl.TitleVariant
							frame.FrameName = &fn
							// Load brand
							if mdl.Product.BrandID != nil {
								var brand vendors.Brand
								if err := tx.First(&brand, *mdl.Product.BrandID).Error; err == nil {
									frame.BrandName = brand.BrandName
								}
							}
							// Load vendor
							if mdl.Product.VendorID != nil {
								var vendor vendors.Vendor
								if err := tx.First(&vendor, *mdl.Product.VendorID).Error; err == nil {
									frame.VendorName = &vendor.VendorName
								}
							}
							// Load manufacturer
							if mdl.Product.ManufacturerID != nil {
								var mfr vendors.Manufacturer
								if err := tx.First(&mfr, *mdl.Product.ManufacturerID).Error; err == nil {
									frame.ManufacturerName = &mfr.ManufacturerName
								}
							}
						}
					}
				}
			}
		}

		if err := tx.Create(frame).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("create frame: %w", err)
		}
		labTicketFrameID = &frame.IDLabTicketFrame

		// --- Invoice + transactions (stock frame only, not POF) ---
		if inv != nil && !req.POF {
			var existingFrameItem invoiceModel.InvoiceItemSale
			err := tx.Where("invoice_id = ? AND item_type = ? AND item_id = ?",
				invoice.IDInvoice, "Frames", inv.IDInventory).First(&existingFrameItem).Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Find price from PriceBook
				var price float64
				var pb inventory.PriceBook
				if err := tx.Where("inventory_id = ?", inv.IDInventory).First(&pb).Error; err == nil {
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
					Discount:    0,
					Total:       price,
					Taxable:     boolPtr(false),
					TotalTax:    0,
				}
				if err := tx.Create(&newItem).Error; err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("create invoice item: %w", err)
				}

				oldInvoiceID := inv.InvoiceID
				locID := int64(loc.IDLocation)

				// Transfer from warehouse if needed
				if loc.WarehouseID != nil && inv.LocationID == int64(*loc.WarehouseID) {
					whID := int64(*loc.WarehouseID)
					txn := inventory.InventoryTransaction{
						InventoryID:    &inv.IDInventory,
						FromLocationID: &whID,
						ToLocationID:   &locID,
						TransferredBy:  employeeID,
						StatusItems:    types.StatusItemsInventory("TRANSFERRED TO SHOWCASE"),
						TransactionType: "Transfer",
						DateTransaction: time.Now(),
					}
					if err := tx.Create(&txn).Error; err != nil {
						tx.Rollback()
						return nil, fmt.Errorf("create transfer txn: %w", err)
					}
					inv.LocationID = locID
				}

				// Update inventory status
				if string(inv.StatusItemsInventory) != "Ordered" {
					inv.StatusItemsInventory = types.StatusInventorySOLD
				}
				inv.InvoiceID = invoice.IDInvoice
				if err := tx.Save(inv).Error; err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("update inventory: %w", err)
				}

				// Sale transaction
				saleTxn := inventory.InventoryTransaction{
					InventoryID:     &inv.IDInventory,
					FromLocationID:  &locID,
					TransferredBy:   employeeID,
					InvoiceID:       &invoice.IDInvoice,
					OldInvoiceID:    &oldInvoiceID,
					StatusItems:     inv.StatusItemsInventory,
					TransactionType: "Sale",
					DateTransaction: time.Now(),
				}
				if err := tx.Create(&saleTxn).Error; err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("create sale txn: %w", err)
				}

				// Recalculate invoice totals
				invoice.TotalAmount = invoice.TotalAmount + price
				if invoice.Discount == nil {
					z := 0.0
					invoice.Discount = &z
				}
				discount := *invoice.Discount
				invoice.FinalAmount = roundTo2(invoice.TotalAmount - discount)
				gcBal := 0.0
				if invoice.GiftCardBal != nil {
					gcBal = *invoice.GiftCardBal
				}
				invoice.Due = roundTo2(invoice.FinalAmount - invoice.PTBal - invoice.InsBal - gcBal)
				if err := tx.Save(&invoice).Error; err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("update invoice: %w", err)
				}
			}
		}

	} else if gOrC == "c" {
		// ======================== CONTACTS ========================

		// --- Powers Contact ---
		pc := &labTicketModel.LabTicketPowersContact{}
		if req.IDRX != nil {
			var pp prescriptions.PatientPrescription
			if err := tx.Where("id_patient_prescription = ?", *req.IDRX).First(&pp).Error; err == nil {
				var cl prescriptions.ContactLensPrescription
				if err := tx.Where("prescription_id = ?", pp.IDPatientPrescription).First(&cl).Error; err == nil {
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
					if cl.OdType != nil {
						t := labTicketModel.ContactLensType(*cl.OdType)
						pc.ODType = &t
					}
					if cl.OsType != nil {
						t := labTicketModel.ContactLensType(*cl.OsType)
						pc.OSType = &t
					}
					pc.ExpirationDate = cl.ExpirationDate
					if cl.OdHPrismDirection != nil {
						d := labTicketModel.HPrismDirection(*cl.OdHPrismDirection)
						pc.ODHPrismDirection = &d
					}
					if cl.OsHPrismDirection != nil {
						d := labTicketModel.HPrismDirection(*cl.OsHPrismDirection)
						pc.OSHPrismDirection = &d
					}
					if cl.OdVPrismDirection != nil {
						d := labTicketModel.VPrismDirection(*cl.OdVPrismDirection)
						pc.ODVPrismDirection = &d
					}
					if cl.OsVPrismDirection != nil {
						d := labTicketModel.VPrismDirection(*cl.OsVPrismDirection)
						pc.OSVPrismDirection = &d
					}
				}
			}
		}

		// Override from request
		overrideStrPtr(&pc.ODContLens, req.ODContLens)
		overrideStrPtr(&pc.OSContLens, req.OSContLens)
		overrideStrPtr(&pc.ODBC, req.ContactODBc)
		overrideStrPtr(&pc.OSBC, req.ContactOSBc)
		overrideF64Ptr(&pc.ODDia, req.ODDia)
		overrideF64Ptr(&pc.OSDia, req.OSDia)
		overrideStrPtr(&pc.ODPwr, req.ODPwr)
		overrideStrPtr(&pc.OSPwr, req.OSPwr)
		overrideStrPtr(&pc.ODCyl, req.ContactODCyl)
		overrideStrPtr(&pc.OSCyl, req.ContactOSCyl)
		overrideStrPtr(&pc.ODAxis, req.ContactODAxis)
		overrideStrPtr(&pc.OSAxis, req.ContactOSAxis)
		overrideStrPtr(&pc.ODAdd, req.ContactODAdd)
		overrideStrPtr(&pc.OSAdd, req.ContactOSAdd)
		overrideStrPtr(&pc.ODColor, req.ODColor)
		overrideStrPtr(&pc.OSColor, req.OSColor)
		if req.ODType != nil {
			t := labTicketModel.ContactLensType(*req.ODType)
			pc.ODType = &t
		}
		if req.OSType != nil {
			t := labTicketModel.ContactLensType(*req.OSType)
			pc.OSType = &t
		}
		if req.ExpirationDate != nil {
			if t, err := time.Parse("2006-01-02", *req.ExpirationDate); err == nil {
				pc.ExpirationDate = &t
			}
		}
		if req.ContactODHPrismDir != nil {
			d := labTicketModel.HPrismDirection(*req.ContactODHPrismDir)
			pc.ODHPrismDirection = &d
		}
		if req.ContactOSHPrismDir != nil {
			d := labTicketModel.HPrismDirection(*req.ContactOSHPrismDir)
			pc.OSHPrismDirection = &d
		}
		if req.ContactODVPrismDir != nil {
			d := labTicketModel.VPrismDirection(*req.ContactODVPrismDir)
			pc.ODVPrismDirection = &d
		}
		if req.ContactOSVPrismDir != nil {
			d := labTicketModel.VPrismDirection(*req.ContactOSVPrismDir)
			pc.OSVPrismDirection = &d
		}

		if err := tx.Create(pc).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("create powers contact: %w", err)
		}
		labTicketPowersContactID = &pc.IDLabTicketPowersContact

		// --- LabTicketContact ---
		contactSvcID := 1
		if req.LabTicketContactLensServicesID != nil {
			contactSvcID = *req.LabTicketContactLensServicesID
		}
		odAnnual := true
		if req.ODAnnualSupply != nil {
			odAnnual = *req.ODAnnualSupply
		}
		osAnnual := true
		if req.OSAnnualSupply != nil {
			osAnnual = *req.OSAnnualSupply
		}

		ltContact := &labTicketModel.LabTicketContact{
			LabTicketContactLensServicesID: contactSvcID,
			ODAnnualSupply:                odAnnual,
			OSAnnualSupply:                osAnnual,
			ODTotalQty:                    req.ODTotalQty,
			OSTotalQty:                    req.OSTotalQty,
			Reasons:                       req.Reasons,
			Modality:                      req.Modality,
			BrandContactLensID:            req.BrandContactLensID,
		}

		if err := tx.Create(ltContact).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("create contact: %w", err)
		}
		labTicketContactID = &ltContact.IDLabTicketContact

	} else {
		tx.Rollback()
		return nil, fmt.Errorf("g_or_c must be 'g' or 'c'")
	}

	// ======================== Create the ticket ========================
	var datePromise *time.Time
	if req.DatePromise != nil {
		if t, err := time.Parse("2006-01-02", *req.DatePromise); err == nil {
			datePromise = &t
		}
	}

	ticket := &labTicketModel.LabTicket{
		GOrC:                     &gOrC,
		NumberTicket:             numberTicket,
		LabID:                    req.LabID,
		LabTicketStatusID:        labTicketStatusID,
		DateCreate:               &currentDate,
		DatePromise:              datePromise,
		PatientID:                patientID,
		OrdersLensID:             req.OrdersLensID,
		InvoiceID:                invoiceID,
		Tray:                     req.Tray,
		Notified:                 req.Notified,
		Amt:                      req.Amt,
		OurNote:                  req.OurNote,
		LabInstructions:          req.LabInstructions,
		EmployeeID:               employeeID,
		LabTicketPowersID:        labTicketPowersID,
		LabTicketLensID:          labTicketLensID,
		LabTicketFrameID:         labTicketFrameID,
		LabTicketPowersContactID: labTicketPowersContactID,
		LabTicketContactID:       labTicketContactID,
	}

	if err := tx.Create(ticket).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create ticket: %w", err)
	}

	// Activity log
	details := map[string]interface{}{
		"number_ticket": numberTicket,
		"invoice_id":    invoiceID,
		"patient_id":    patientID,
		"g_or_c":        gOrC,
		"lab_id":        req.LabID,
	}
	detailsJSON, _ := json.Marshal(details)
	_ = activitylog.Log(tx, "ticket", "create",
		activitylog.WithEmployee(employeeID),
		activitylog.WithLocation(loc.IDLocation),
		activitylog.WithEntity(ticket.IDLabTicket),
		activitylog.WithDetailsRaw(detailsJSON),
	)

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &CreateTicketResult{
		Message:     "Ticket created successfully",
		IDLabTicket: ticket.IDLabTicket,
		GOrC:        gOrC,
	}, nil
}

// ──── helpers ────

func overrideStrPtr(dst **string, src *string) {
	if src != nil {
		*dst = src
	}
}

func overrideF64Ptr(dst **float64, src *float64) {
	if src != nil {
		*dst = src
	}
}

func overrideHPrismDir(dst **labTicketModel.HPrismDirection, src *string) {
	if src != nil {
		d := labTicketModel.HPrismDirection(*src)
		*dst = &d
	}
}

func overrideVPrismDir(dst **labTicketModel.VPrismDirection, src *string) {
	if src != nil {
		d := labTicketModel.VPrismDirection(*src)
		*dst = &d
	}
}

func boolPtr(b bool) *bool { return &b }

func roundTo2(f float64) float64 {
	return math.Round(f*100) / 100
}
