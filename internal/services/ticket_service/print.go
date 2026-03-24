package ticket_service

import (
	"fmt"

	empModel "sighthub-backend/internal/models/employees"
	generalModel "sighthub-backend/internal/models/general"
	invoiceModel "sighthub-backend/internal/models/invoices"
	labTicketModel "sighthub-backend/internal/models/lab_ticket"
	lensModel "sighthub-backend/internal/models/lenses"
	locationModel "sighthub-backend/internal/models/location"
	patModel "sighthub-backend/internal/models/patients"
	svcModel "sighthub-backend/internal/models/service"
	vendorModel "sighthub-backend/internal/models/vendors"
)

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func (s *Service) PrintTicket(ticketID int64, includeInvoice bool) (map[string]interface{}, error) {
	var ticket labTicketModel.LabTicket
	if err := s.db.
		Preload("Powers").
		Preload("Lens").
		Preload("Frame").
		First(&ticket, ticketID).Error; err != nil {
		return nil, fmt.Errorf("ticket not found")
	}

	// ── Lab name ──
	var labName *string
	if ticket.LabID != nil {
		var lab vendorModel.Vendor
		if s.db.First(&lab, *ticket.LabID).Error == nil {
			labName = &lab.VendorName
		}
	}

	// ── Location ──
	var inv invoiceModel.Invoice
	s.db.First(&inv, ticket.InvoiceID)
	var locInfo map[string]interface{}
	var loc locationModel.Location
	if s.db.First(&loc, inv.LocationID).Error == nil {
		locInfo = map[string]interface{}{
			"full_name": loc.FullName,
		}
	}

	// ── Patient ──
	var pat patModel.Patient
	s.db.First(&pat, ticket.PatientID)
	patInfo := map[string]interface{}{
		"full_name":                  fmt.Sprintf("%s %s", pat.FirstName, pat.LastName),
		"phone":                      pat.Phone,
		"patient_address_street_1":   pat.StreetAddress,
		"patient_address_city":       pat.City,
		"patient_address_state":      pat.State,
		"patient_address_zip":        pat.ZipCode,
	}

	// ── Powers ──
	var powers map[string]interface{}
	if ticket.Powers != nil {
		p := ticket.Powers
		powers = map[string]interface{}{
			"od_sph": deref(p.ODSph), "od_cyl": deref(p.ODCyl), "od_axis": deref(p.ODAxis),
			"od_add": p.ODAdd, "od_dt": deref(p.ODDT), "od_nr": deref(p.ODNR),
			"od_seg_hd": deref(p.ODSegHD), "od_oc": deref(p.ODOC), "od_bc": deref(p.ODBC),
			"od_h_prism": p.ODHPrism, "od_h_prism_direction": p.ODHPrismDirection,
			"od_v_prism": p.ODVPrism, "od_v_prism_direction": p.ODVPrismDirection,
			"os_sph": deref(p.OSSph), "os_cyl": deref(p.OSCyl), "os_axis": deref(p.OSAxis),
			"os_add": p.OSAdd, "os_dt": deref(p.OSDT), "os_nr": deref(p.OSNR),
			"os_seg_hd": deref(p.OSSegHD), "os_oc": deref(p.OSOC), "os_bc": deref(p.OSBC),
			"os_h_prism": p.OSHPrism, "os_h_prism_direction": p.OSHPrismDirection,
			"os_v_prism": p.OSVPrism, "os_v_prism_direction": p.OSVPrismDirection,
		}
	}

	// ── Frame ──
	var frameInfo map[string]interface{}
	if ticket.Frame != nil {
		f := ticket.Frame
		frameInfo = map[string]interface{}{
			"brand_name":         f.BrandName,
			"item_type":          f.ItemType,
			"model_title_variant": f.ModelTitleVariant,
			"color":               f.Color,
			"frame_source":       f.FrameSource,
			"size_lens_width":     f.SizeLensWidth,
			"size_bridge_width":   f.SizeBridgeWidth,
			"size_temple_length":  f.SizeTempleLength,
			"b_dim":               f.BValue,
			"ed":                  f.EDValue,
			"circum":              f.CircValue,
			"panto":               f.Panto,
			"wrap_angle":          f.WrapAngle,
			"head_eye_ratio":      f.HeadEyeRatio,
			"stability_coeff":     f.StabilityCoeff,
			"head_cape":           f.HeadCape,
			"corridor_r":          f.CorridorR,
			"corridor_l":          f.CorridorL,
			"pof":                 f.POF,
		}
	}

	// ── Lens items + treatments + additional from invoice ──
	var lensItems []map[string]interface{}
	var addServices []map[string]interface{}
	lensFeatures := map[string]interface{}{
		"sr_cost": false, "uv": false, "ar": false,
		"tint": false, "drill": false, "send": false,
	}

	var invoiceItems []invoiceModel.InvoiceItemSale
	s.db.Where("invoice_id = ? AND item_type IN ?", ticket.InvoiceID,
		[]string{"Lens", "Treatment", "Add service"}).
		Find(&invoiceItems)

	for _, item := range invoiceItems {
		if item.ItemID == nil {
			continue
		}
		switch item.ItemType {
		case "Lens":
			var l lensModel.Lenses
			if s.db.Preload("LensesMaterial").First(&l, *item.ItemID).Error == nil {
				matName := ""
				if l.LensesMaterial != nil {
					matName = l.LensesMaterial.MaterialName
				}
				lensItems = append(lensItems, map[string]interface{}{
					"name":        l.LensName,
					"description": deref(l.Description),
					"material":    matName,
				})
			}
		case "Treatment":
			var t lensModel.LensTreatments
			if s.db.First(&t, *item.ItemID).Error == nil {
				addServices = append(addServices, map[string]interface{}{
					"name":        t.ItemNbr,
					"description": deref(t.Description),
				})
			}
		case "Add service":
			var a svcModel.AdditionalService
			if s.db.First(&a, *item.ItemID).Error == nil {
				name := ""
				if a.ItemNumber != nil {
					name = *a.ItemNumber
				}
				addServices = append(addServices, map[string]interface{}{
					"name":        name,
					"description": a.InvoiceDesc,
				})
				// Aggregate lens features
				if a.SrCost != nil && *a.SrCost {
					lensFeatures["sr_cost"] = true
				}
				if a.UV != nil && *a.UV {
					lensFeatures["uv"] = true
				}
				if a.AR != nil && *a.AR {
					lensFeatures["ar"] = true
				}
				if a.Tint != nil && *a.Tint {
					lensFeatures["tint"] = true
				}
				if a.Drill != nil && *a.Drill {
					lensFeatures["drill"] = true
				}
				if a.Send != nil && *a.Send {
					lensFeatures["send"] = true
				}
			}
		}
	}

	// ── Lens tint info ──
	var lensInfo map[string]interface{}
	if ticket.Lens != nil {
		lensInfo = map[string]interface{}{
			"lens_type_color": ticket.Lens.LensTypeColor,
			"tint_percent":    ticket.Lens.TintPercent,
		}
	}

	// ── Date formatting ──
	var dateStr, datePromiseStr *string
	if ticket.DateCreate != nil {
		d := ticket.DateCreate.Format("2006-01-02")
		dateStr = &d
	}
	if ticket.DatePromise != nil {
		d := ticket.DatePromise.Format("2006-01-02")
		datePromiseStr = &d
	}

	labTicketData := map[string]interface{}{
		"id_lab_ticket":  ticket.IDLabTicket,
		"number_ticket":  ticket.NumberTicket,
		"lab":            labName,
		"lab_id":         ticket.LabID,
		"our_note":       ticket.OurNote,
		"date_promise":   datePromiseStr,
		"date":           dateStr,
		"powers":         powers,
		"frame":          frameInfo,
		"lens_items":     lensItems,
		"add_services":   addServices,
		"lens_features":  lensFeatures,
		"lens":           lensInfo,
		"location":       locInfo,
		"patient_id":     ticket.PatientID,
	}

	result := map[string]interface{}{
		"lab_ticket": labTicketData,
		"patient":    patInfo,
	}

	// ── Invoice block (optional) ──
	if includeInvoice {
		// Sold by
		soldBy := ""
		if inv.EmployeeID != nil {
			var emp empModel.Employee
			if s.db.First(&emp, *inv.EmployeeID).Error == nil {
				soldBy = emp.FirstName + " " + emp.LastName
			}
		}

		// All invoice items
		var allItems []invoiceModel.InvoiceItemSale
		s.db.Where("invoice_id = ?", ticket.InvoiceID).Find(&allItems)
		invItems := make([]map[string]interface{}, 0, len(allItems))
		for _, it := range allItems {
			invItems = append(invItems, map[string]interface{}{
				"id":          it.IDInvoiceSale,
				"description": it.Description,
				"quantity":    it.Quantity,
				"price":       fmt.Sprintf("%.2f", it.Price),
				"total":       fmt.Sprintf("%.2f", it.Total),
				"discount":    fmt.Sprintf("%.2f", it.Discount),
			})
		}

		// Payment history
		var payments []patModel.PaymentHistory
		s.db.Where("invoice_id = ?", ticket.InvoiceID).Find(&payments)
		payHistory := make([]map[string]interface{}, 0, len(payments))
		for _, ph := range payments {
			methodName := ""
			if ph.PaymentMethodID != nil {
				var pm generalModel.PaymentMethod
				if s.db.First(&pm, *ph.PaymentMethodID).Error == nil {
					methodName = pm.MethodName
				}
			}
			payHistory = append(payHistory, map[string]interface{}{
				"date":   ph.PaymentTimestamp.Format("2006-01-02"),
				"type":   methodName,
				"method": methodName,
				"amount": fmt.Sprintf("%.2f", ph.Amount),
			})
		}

		// Paid = sum of payments
		paid := 0.0
		for _, ph := range payments {
			paid += ph.Amount
		}

		result["invoice"] = map[string]interface{}{
			"invoice_id":      inv.IDInvoice,
			"invoice_number":  inv.NumberInvoice,
			"invoice_date":    inv.DateCreate.Format("2006-01-02"),
			"sold_by":         soldBy,
			"total_amount":    fmt.Sprintf("%.2f", inv.TotalAmount),
			"final_amount":    fmt.Sprintf("%.2f", inv.FinalAmount),
			"due_amount":      fmt.Sprintf("%.2f", inv.Due),
			"paid":            fmt.Sprintf("%.2f", paid),
			"items":           invItems,
			"payment_history": payHistory,
			"patient_info":    patInfo,
		}
	}

	return result, nil
}
