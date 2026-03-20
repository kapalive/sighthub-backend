package invoice

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	clModel "sighthub-backend/internal/models/contact_lens"
	empModel "sighthub-backend/internal/models/employees"
	frameModel "sighthub-backend/internal/models/frames"
	generalModel "sighthub-backend/internal/models/general"
	insModel "sighthub-backend/internal/models/insurance"
	invModel "sighthub-backend/internal/models/inventory"
	"sighthub-backend/internal/models/invoices"
	labModel "sighthub-backend/internal/models/lab_ticket"
	lensModel "sighthub-backend/internal/models/lenses"
	locModel "sighthub-backend/internal/models/location"
	miscModel "sighthub-backend/internal/models/misc"
	patModel "sighthub-backend/internal/models/patients"
	svcModel "sighthub-backend/internal/models/service"
	vendorModel "sighthub-backend/internal/models/vendors"
	pkgSKU "sighthub-backend/pkg/sku"
)

// ─── Service ─────────────────────────────────────────────────────────────────

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ─── Input DTOs ───────────────────────────────────────────────────────────────

type AddItemInput struct {
	ItemID       interface{} `json:"item_id"`
	IDInvoiceSale interface{} `json:"id_invoice_sale"`
	Quantity     float64     `json:"quantity"`
	Price        *float64    `json:"price"`
	Discount     float64     `json:"discount"`
	Taxable      interface{} `json:"taxable"`
	Description  string      `json:"description"`
	Cost         float64     `json:"cost"`
	SaleKey      string      `json:"sale_key"`
}

type AddItemsInput struct {
	PbKey   string         `json:"pb_key"`
	SaleKey string         `json:"sale_key"`
	Notes   *string        `json:"notes"`
	Items   []AddItemInput `json:"items"`
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (s *Service) getEmployeeAndLocation(username string) (*empModel.Employee, *locModel.Location, error) {
	var login authModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, nil, errors.New("employee or location not found")
	}
	var emp empModel.Employee
	if err := s.db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return nil, nil, errors.New("employee or location not found")
	}
	if emp.LocationID == nil {
		return nil, nil, errors.New("employee or location not found")
	}
	var loc locModel.Location
	if err := s.db.First(&loc, *emp.LocationID).Error; err != nil {
		return nil, nil, errors.New("employee or location not found")
	}
	return &emp, &loc, nil
}

func (s *Service) sumPatientPayments(tx *gorm.DB, invoiceID int64) float64 {
	var total float64
	tx.Model(&patModel.PaymentHistory{}).
		Where("invoice_id = ? AND (payment_method_id IS NULL OR payment_method_id != 14)", invoiceID).
		Select("COALESCE(SUM(amount), 0)").Scan(&total)
	return total
}

func (s *Service) sumInsurancePayments(tx *gorm.DB, invoiceID int64) float64 {
	var payments []insModel.InsurancePayment
	tx.Where("invoice_id = ?", invoiceID).Find(&payments)
	var total float64
	for _, p := range payments {
		if v, err := strconv.ParseFloat(p.Amount, 64); err == nil {
			total += v
		}
	}
	return total
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }
func round4(v float64) float64 { return math.Round(v*10000) / 10000 }

func (s *Service) recalcInvoice(tx *gorm.DB, invoice *invoices.Invoice) error {
	var items []invoices.InvoiceItemSale
	if err := tx.Where("invoice_id = ?", invoice.IDInvoice).Find(&items).Error; err != nil {
		return err
	}

	var totalAmount, taxAmount, ptResp, insResp float64
	for _, item := range items {
		totalAmount += item.Total
		taxAmount += item.TotalTax
		if item.PtBalance != nil {
			ptResp += *item.PtBalance
		}
		if item.InsBalance != nil {
			insResp += *item.InsBalance
		}
	}

	invoice.TotalAmount = round2(totalAmount)
	invoice.TaxAmount = round4(taxAmount)
	invoice.FinalAmount = invoice.TotalAmount

	discount := 0.0
	if invoice.Discount != nil {
		discount = *invoice.Discount
	}
	giftCard := 0.0
	if invoice.GiftCardBal != nil {
		giftCard = *invoice.GiftCardBal
	}

	ptPaid := s.sumPatientPayments(tx, invoice.IDInvoice)
	insPaid := s.sumInsurancePayments(tx, invoice.IDInvoice)

	ptBal := ptResp - discount - ptPaid - giftCard
	if ptBal < 0 {
		ptBal = 0
	}
	invoice.PTBal = round2(ptBal)

	insBal := insResp - insPaid
	if insBal < 0 {
		insBal = 0
	}
	invoice.InsBal = round2(insBal)
	invoice.Due = round2(invoice.PTBal + invoice.InsBal)

	return tx.Save(invoice).Error
}

func (s *Service) addInventoryTx(tx *gorm.DB, inventoryID int64, fromLocID *int64, toLocID *int64,
	transferredBy, invoiceID int64, oldInvoiceID *int64, statusItems, txType string) error {
	return tx.Exec(`
		INSERT INTO inventory_transaction
		(inventory_id, from_location_id, to_location_id, transferred_by, invoice_id, old_invoice_id, status_items, transaction_type, date_transaction)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())`,
		inventoryID, fromLocID, toLocID, transferredBy, invoiceID, oldInvoiceID, statusItems, txType,
	).Error
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrFloat64(v float64) *float64 { return &v }
func ptrStr(v string) *string       { return &v }

func normKey(v string) *string {
	s := strings.TrimSpace(v)
	if s == "" {
		return nil
	}
	return &s
}

func parseBool(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t == "true" || t == "1" || t == "yes"
	case float64:
		return t != 0
	}
	return false
}

func parseInt64(v interface{}) (int64, bool) {
	switch t := v.(type) {
	case float64:
		return int64(t), true
	case int64:
		return t, true
	case int:
		return int64(t), true
	case string:
		if t == "" || t == "0" {
			return 0, false
		}
		n, err := strconv.ParseInt(t, 10, 64)
		return n, err == nil
	}
	return 0, false
}

func buildInsuranceInfo(db *gorm.DB, invoiceID, patientID int64) map[string]interface{} {
	var link invoices.InvoiceInsurancePolicy
	if err := db.Where("invoice_id = ?", invoiceID).First(&link).Error; err != nil {
		return nil
	}
	var policy insModel.InsurancePolicy
	if err := db.Preload("InsuranceCompany").First(&policy, link.InsurancePolicyID).Error; err != nil {
		return nil
	}
	var holder patModel.InsuranceHolderPatients
	memberNum := (*string)(nil)
	if err := db.Where("insurance_policy_id = ? AND patient_id = ?", policy.IDInsurancePolicy, patientID).First(&holder).Error; err == nil {
		memberNum = holder.MemberNumber
	}
	companyName := ""
	if policy.InsuranceCompany != nil {
		companyName = policy.InsuranceCompany.CompanyName
	}
	return map[string]interface{}{
		"insurance_policy_id": strconv.FormatInt(policy.IDInsurancePolicy, 10),
		"member_number":       memberNum,
		"insurance_name":      companyName,
	}
}

func buildInsuranceInfoForList(db *gorm.DB, invoiceID int64, invoice *invoices.Invoice) map[string]interface{} {
	if invoice.InsurancePolicyID == nil {
		return nil
	}
	var link invoices.InvoiceInsurancePolicy
	if err := db.Where("invoice_id = ?", invoiceID).First(&link).Error; err != nil {
		return nil
	}
	var policy insModel.InsurancePolicy
	if err := db.Preload("InsuranceCompany").First(&policy, link.InsurancePolicyID).Error; err != nil {
		return nil
	}
	companyName := ""
	if policy.InsuranceCompany != nil {
		companyName = policy.InsuranceCompany.CompanyName
	}
	return map[string]interface{}{
		"id_insurance_policy": policy.IDInsurancePolicy,
		"group_number":        policy.GroupNumber,
		"coverage_details":    policy.CoverageDetails,
		"insurance_name":      companyName,
	}
}

func buildPaymentHistory(db *gorm.DB, invoiceID int64) []map[string]interface{} {
	var ptPayments []patModel.PaymentHistory
	db.Where("invoice_id = ?", invoiceID).Find(&ptPayments)

	var insPayments []insModel.InsurancePayment
	db.Where("invoice_id = ?", invoiceID).Find(&insPayments)

	result := make([]map[string]interface{}, 0, len(ptPayments)+len(insPayments))

	for _, p := range ptPayments {
		var methodName string
		var pm generalModel.PaymentMethod
		if p.PaymentMethodID != nil {
			if err := db.First(&pm, *p.PaymentMethodID).Error; err == nil {
				methodName = pm.MethodName
			}
		}
		if methodName == "" {
			methodName = "N/A"
		}
		result = append(result, map[string]interface{}{
			"date":   p.PaymentTimestamp.Format("01/02/2006 15:04"),
			"type":   "Patient",
			"method": methodName,
			"amount": fmt.Sprintf("$%.2f", p.Amount),
		})
	}
	for _, p := range insPayments {
		var methodName string
		var pt insModel.InsurancePaymentType
		if err := db.First(&pt, p.PaymentTypeID).Error; err == nil {
			methodName = pt.Name
		}
		if methodName == "" {
			methodName = "N/A"
		}
		dateStr := "N/A"
		if p.CreatedAt != nil {
			dateStr = p.CreatedAt.Format("01/02/2006 15:04")
		}
		result = append(result, map[string]interface{}{
			"date":   dateStr,
			"type":   "Insurance",
			"method": methodName,
			"amount": fmt.Sprintf("$%s", p.Amount),
		})
	}
	return result
}

func buildLocationData(loc *locModel.Location) map[string]interface{} {
	defaultLogo := "../logo-eyesync-vector.svg"
	logoPath := defaultLogo
	if loc.LogoPath != nil {
		logoPath = *loc.LogoPath
	}
	return map[string]interface{}{
		"full_name":      loc.FullName,
		"street_address": derefStr(loc.StreetAddress),
		"address_line_2": derefStr(loc.AddressLine2),
		"city":           derefStr(loc.City),
		"state":          derefStr(loc.State),
		"postal_code":    derefStr(loc.PostalCode),
		"phone":          derefStr(loc.Phone),
		"website":        derefStr(loc.Website),
		"logo_path":      logoPath,
	}
}

// ─── GET /invoice ────────────────────────────────────────────────────────────

func (s *Service) GetInvoiceList(username string, patientID int64) ([]map[string]interface{}, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	_ = emp

	var patient patModel.Patient
	if err := s.db.First(&patient, patientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}

	var invList []invoices.Invoice
	s.db.Where("patient_id = ? AND location_id = ?", patientID, loc.IDLocation).
		Order("date_create DESC").Find(&invList)

	if len(invList) == 0 {
		return []map[string]interface{}{}, nil
	}

	// Which invoices have lab tickets?
	invoiceIDs := make([]int64, len(invList))
	for i, inv := range invList {
		invoiceIDs[i] = inv.IDInvoice
	}
	var labInvoiceIDs []int64
	s.db.Model(&labModel.LabTicket{}).
		Where("invoice_id IN ?", invoiceIDs).
		Pluck("invoice_id", &labInvoiceIDs)
	labSet := make(map[int64]bool)
	for _, id := range labInvoiceIDs {
		labSet[id] = true
	}

	result := make([]map[string]interface{}, 0, len(invList))
	for _, inv := range invList {
		ptPaid := s.sumPatientPayments(s.db, inv.IDInvoice)
		insPaid := s.sumInsurancePayments(s.db, inv.IDInvoice)

		// Load employee for invoice
		repName := ""
		if inv.EmployeeID != nil {
			var rep empModel.Employee
			if s.db.First(&rep, *inv.EmployeeID).Error == nil {
				repName = rep.FirstName + " " + rep.LastName
			}
		}

		var labVal interface{} = nil
		if labSet[inv.IDInvoice] {
			labVal = "G"
		}

		row := map[string]interface{}{
			"id_invoice":     inv.IDInvoice,
			"date":           inv.DateCreate.Format(time.RFC3339),
			"invoice_number": inv.NumberInvoice,
			"lab":            labVal,
			"tax_amount":     fmt.Sprintf("%.2f", inv.TaxAmount),
			"final_amount":   fmt.Sprintf("%.2f", inv.FinalAmount),
			"total_amount":   fmt.Sprintf("%.2f", inv.TotalAmount),
			"discount":       fmt.Sprintf("%.2f", func() float64 { if inv.Discount != nil { return *inv.Discount }; return 0 }()),
			"due":            fmt.Sprintf("%.2f", inv.Due),
			"pt_bal":         fmt.Sprintf("%.2f", inv.PTBal),
			"ins_bal":        fmt.Sprintf("%.2f", inv.InsBal),
			"pt_paid":        fmt.Sprintf("%.2f", ptPaid),
			"ins_paid":       fmt.Sprintf("%.2f", insPaid),
			"representative": repName,
		}
		if insInfo := buildInsuranceInfoForList(s.db, inv.IDInvoice, &inv); insInfo != nil {
			row["insurance_policy"] = insInfo
		}
		result = append(result, row)
	}
	return result, nil
}

// ─── POST /invoice ────────────────────────────────────────────────────────────

func (s *Service) CreateInvoice(username string, patientID int64) (map[string]interface{}, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var patient patModel.Patient
	if err := s.db.First(&patient, patientID).Error; err != nil {
		return nil, errors.New("patient not found")
	}

	shortName := ""
	if loc.ShortName != nil {
		shortName = *loc.ShortName
	}

	// Generate invoice number: "S" + shortName + max_id+1 formatted
	var maxID int64
	s.db.Raw("SELECT COALESCE(MAX(id_invoice), 0) FROM invoice").Scan(&maxID)
	invoiceNumber := fmt.Sprintf("S%s%07d", shortName, maxID+1)

	discount := 0.0
	inv := invoices.Invoice{
		NumberInvoice: invoiceNumber,
		DateCreate:    time.Now(),
		Discount:      &discount,
		TotalAmount:   0,
		FinalAmount:   0,
		PTBal:         0,
		InsBal:        0,
		Due:           0,
		EmployeeID:    func() *int64 { v := int64(emp.IDEmployee); return &v }(),
		LocationID:    int64(loc.IDLocation),
		PatientID:     &patientID,
	}

	if err := s.db.Create(&inv).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":        "Invoice created successfully",
		"invoice_id":     inv.IDInvoice,
		"invoice_number": invoiceNumber,
	}, nil
}

// ─── POST /invoice/{id}/remake ────────────────────────────────────────────────

func (s *Service) CreateRemakeInvoice(username string, invoiceID int64) (map[string]interface{}, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var original invoices.Invoice
	if err := s.db.First(&original, invoiceID).Error; err != nil {
		return nil, errors.New("invoice not found")
	}
	if original.LocationID != int64(loc.IDLocation) {
		return nil, errors.New("remake can only be created at the location where the original invoice was created")
	}

	// Build remake number
	var origLoc locModel.Location
	s.db.First(&origLoc, original.LocationID)
	shortName := ""
	if origLoc.ShortName != nil {
		shortName = *origLoc.ShortName
	}
	prefix := fmt.Sprintf("S%s%d", shortName, original.IDInvoice)
	var countExisting int64
	s.db.Model(&invoices.Invoice{}).Where("number_invoice LIKE ?", prefix+"R%").Count(&countExisting)
	remakeNumber := fmt.Sprintf("%sR%d", prefix, countExisting+1)

	discount := 0.0
	remake := invoices.Invoice{
		NumberInvoice: remakeNumber,
		DateCreate:    time.Now(),
		Discount:      &discount,
		TotalAmount:   0,
		FinalAmount:   0,
		PTBal:         0,
		InsBal:        0,
		Due:           0,
		EmployeeID:    func() *int64 { v := int64(emp.IDEmployee); return &v }(),
		LocationID:    original.LocationID,
		PatientID:     original.PatientID,
		Remake:        true,
	}
	if err := s.db.Create(&remake).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":        "Remake invoice created successfully",
		"invoice_id":     remake.IDInvoice,
		"invoice_number": remake.NumberInvoice,
	}, nil
}

// ─── GET /payment-methods ─────────────────────────────────────────────────────

func (s *Service) GetPaymentMethods() ([]map[string]interface{}, error) {
	var methods []generalModel.PaymentMethod
	if err := s.db.Find(&methods).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, len(methods))
	for i, m := range methods {
		result[i] = map[string]interface{}{
			"id_payment_method": m.IDPaymentMethod,
			"method_name":       m.MethodName,
			"short_name":        m.ShortName,
		}
	}
	return result, nil
}

// ─── GET /invoice/{id} ────────────────────────────────────────────────────────

func (s *Service) GetInvoice(username string, invoiceID int64, groupByFrame bool) (map[string]interface{}, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	_ = emp

	var inv invoices.Invoice
	if err := s.db.Preload("Employee").Preload("PaymentMethod").Preload("Location").First(&inv, invoiceID).Error; err != nil {
		return nil, errors.New("invoice not found")
	}

	creatorName := ""
	if inv.Employee != nil {
		creatorName = inv.Employee.FirstName + " " + inv.Employee.LastName
	}
	paymentMethodName := ""
	if inv.PaymentMethod != nil {
		paymentMethodName = inv.PaymentMethod.MethodName
	}

	var patID int64
	if inv.PatientID != nil {
		patID = *inv.PatientID
	}
	insuranceInfo := buildInsuranceInfo(s.db, invoiceID, patID)

	var itemList []invoices.InvoiceItemSale
	s.db.Where("invoice_id = ?", invoiceID).Find(&itemList)

	itemsData := []map[string]interface{}{}
	groupedFrames := map[string]map[string]interface{}{}

	for _, item := range itemList {
		desc := item.Description

		if groupByFrame && item.ItemType == "Frames" {
			key := desc
			if _, ok := groupedFrames[key]; !ok {
				groupedFrames[key] = map[string]interface{}{
					"description": key,
					"_qty": 0.0, "_discount": 0.0, "_total": 0.0,
					"_total_tax": 0.0, "_pt_balance": 0.0, "_ins_balance": 0.0,
					"_price": item.Price, "_taxable": item.Taxable != nil && *item.Taxable,
				}
			}
			g := groupedFrames[key]
			g["_qty"] = g["_qty"].(float64) + float64(item.Quantity)
			g["_discount"] = g["_discount"].(float64) + item.Discount
			g["_total"] = g["_total"].(float64) + item.Total
			g["_total_tax"] = g["_total_tax"].(float64) + item.TotalTax
			if item.PtBalance != nil {
				g["_pt_balance"] = g["_pt_balance"].(float64) + *item.PtBalance
			}
			if item.InsBalance != nil {
				g["_ins_balance"] = g["_ins_balance"].(float64) + *item.InsBalance
			}
			if item.Taxable != nil && *item.Taxable {
				g["_taxable"] = true
			}
			groupedFrames[key] = g
			continue
		}

		var itemName *string
		var addInfo *string
		itype := item.ItemType
		if item.ItemID != nil {
			switch itype {
			case "Lens":
				var obj lensModel.Lenses
				if s.db.First(&obj, *item.ItemID).Error == nil {
					itemName = ptrStr(obj.LensName)
				}
			case "Contact Lens":
				var obj clModel.ContactLensItem
				if s.db.First(&obj, *item.ItemID).Error == nil {
					itemName = ptrStr(obj.NameContact)
				}
			case "Treatment":
				var obj lensModel.LensTreatments
				if s.db.First(&obj, *item.ItemID).Error == nil {
					itemName = ptrStr(obj.ItemNbr)
				}
			case "Prof. service":
				var obj svcModel.ProfessionalService
				if s.db.First(&obj, *item.ItemID).Error == nil {
					itemName = ptrStr(obj.ItemNumber)
				}
			case "Add service":
				var obj svcModel.AdditionalService
				if s.db.First(&obj, *item.ItemID).Error == nil && obj.ItemNumber != nil {
					itemName = obj.ItemNumber
				}
			case "misc":
				var obj miscModel.MiscInvoiceItem
				if s.db.First(&obj, *item.ItemID).Error == nil {
					itemName = ptrStr(obj.ItemNumber)
				}
			case "Frames":
				var frame invModel.Inventory
				if s.db.First(&frame, *item.ItemID).Error == nil {
					if frame.StatusItemsInventory == "Ordered" {
						addInfo = ptrStr("Pending")
					}
				}
			}
		}

		ptBal := "0.00"
		if item.PtBalance != nil {
			ptBal = fmt.Sprintf("%.2f", *item.PtBalance)
		}
		insBal := "0.00"
		if item.InsBalance != nil {
			insBal = fmt.Sprintf("%.2f", *item.InsBalance)
		}
		itemIDStr := "0"
		if item.ItemID != nil {
			itemIDStr = strconv.FormatInt(*item.ItemID, 10)
		}

		itemsData = append(itemsData, map[string]interface{}{
			"item_sale_id": strconv.FormatInt(item.IDInvoiceSale, 10),
			"pb_key":       item.ItemType,
			"sale_key":     item.SaleKey,
			"item_id":      itemIDStr,
			"item_name":    itemName,
			"description":  desc,
			"quantity":     strconv.Itoa(item.Quantity),
			"price":        fmt.Sprintf("%.2f", item.Price),
			"discount":     fmt.Sprintf("%.2f", item.Discount),
			"total":        fmt.Sprintf("%.2f", item.Total),
			"taxable":      item.Taxable != nil && *item.Taxable,
			"total_tax":    fmt.Sprintf("%.2f", item.TotalTax),
			"pt_balance":   ptBal,
			"ins_balance":  insBal,
			"add_info":     addInfo,
		})
	}

	if groupByFrame {
		for _, g := range groupedFrames {
			itemsData = append(itemsData, map[string]interface{}{
				"description": g["description"],
				"quantity":    fmt.Sprintf("%.0f", g["_qty"].(float64)),
				"price":       fmt.Sprintf("%.2f", g["_price"].(float64)),
				"discount":    fmt.Sprintf("%.2f", g["_discount"].(float64)),
				"total":       fmt.Sprintf("%.2f", g["_total"].(float64)),
				"taxable":     g["_taxable"].(bool),
				"total_tax":   fmt.Sprintf("%.4f", g["_total_tax"].(float64)),
				"pt_balance":  fmt.Sprintf("%.2f", g["_pt_balance"].(float64)),
				"ins_balance": fmt.Sprintf("%.2f", g["_ins_balance"].(float64)),
			})
		}
	}

	ptPaid := s.sumPatientPayments(s.db, invoiceID)
	insPaid := s.sumInsurancePayments(s.db, invoiceID)

	discount := "0.00"
	if inv.Discount != nil {
		discount = fmt.Sprintf("%.2f", *inv.Discount)
	}
	giftCardBal := "0.00"
	if inv.GiftCardBal != nil {
		giftCardBal = fmt.Sprintf("%.2f", *inv.GiftCardBal)
	}
	locationName := ""
	if inv.Location != nil {
		locationName = inv.Location.FullName
	}

	return map[string]interface{}{
		"invoice_id":       strconv.FormatInt(inv.IDInvoice, 10),
		"invoice_number":   inv.NumberInvoice,
		"location":         locationName,
		"sold_by":          creatorName,
		"invoice_date":     inv.CreatedAt.Format("2006-01-02 15:04:05"),
		"payment_method":   paymentMethodName,
		"items":            itemsData,
		"total_amount":     fmt.Sprintf("%.2f", inv.TotalAmount),
		"final_amount":     fmt.Sprintf("%.2f", inv.FinalAmount),
		"tax_amount":       fmt.Sprintf("%.2f", inv.TaxAmount),
		"discount":         discount,
		"pt_bal":           fmt.Sprintf("%.2f", inv.PTBal),
		"ins_bal":          fmt.Sprintf("%.2f", inv.InsBal),
		"pt_paid":          fmt.Sprintf("%.2f", ptPaid),
		"ins_paid":         fmt.Sprintf("%.2f", insPaid),
		"gift_card_balance": giftCardBal,
		"due":              fmt.Sprintf("%.2f", inv.Due),
		"notes":            inv.Notified,
		"status_reason":    inv.Reason,
		"insurance_policy": insuranceInfo,
		"finalized":        inv.Finalized,
	}, nil
}

// ─── BuildInvoiceHTMLContext ───────────────────────────────────────────────────

func (s *Service) BuildInvoiceHTMLContext(invoiceID int64) (map[string]interface{}, error) {
	var inv invoices.Invoice
	if err := s.db.Preload("Employee").Preload("PaymentMethod").Preload("Location").Preload("Patient").First(&inv, invoiceID).Error; err != nil {
		return nil, errors.New("invoice not found")
	}
	if inv.Location == nil {
		return nil, errors.New("location not found for this invoice")
	}
	if inv.Patient == nil {
		return nil, errors.New("patient not found")
	}

	creatorName := "N/A"
	if inv.Employee != nil {
		creatorName = inv.Employee.FirstName + " " + inv.Employee.LastName
	}
	paymentMethodName := "N/A"
	if inv.PaymentMethod != nil {
		paymentMethodName = inv.PaymentMethod.MethodName
	}

	patient := inv.Patient
	patientName := patient.FirstName + " " + patient.LastName
	phones := strings.TrimSpace(strings.Join([]string{
		derefStr(patient.Phone),
		derefStr(patient.PhoneHome),
		derefStr(patient.CellWork),
	}, " "))
	patientAddress := strings.TrimSpace(strings.Join([]string{
		derefStr(patient.StreetAddress),
		derefStr(patient.City),
	}, " "))
	patientState := strings.TrimSpace(strings.Join([]string{
		derefStr(patient.State),
		derefStr(patient.ZipCode),
	}, " "))

	var patID2 int64
	if inv.PatientID != nil {
		patID2 = *inv.PatientID
	}
	insuranceInfo := buildInsuranceInfo(s.db, invoiceID, patID2)

	var itemList []invoices.InvoiceItemSale
	s.db.Where("invoice_id = ?", invoiceID).Find(&itemList)

	itemsData := make([]map[string]interface{}, 0, len(itemList))
	for _, item := range itemList {
		itemsData = append(itemsData, map[string]interface{}{
			"id":          strconv.FormatInt(item.IDInvoiceSale, 10),
			"description": item.Description,
			"quantity":    strconv.Itoa(item.Quantity),
			"price":       fmt.Sprintf("%.2f", item.Price),
			"discount":    fmt.Sprintf("%.2f", item.Discount),
			"total":       fmt.Sprintf("%.2f", item.Total),
			"total_tax":   fmt.Sprintf("%.2f", item.TotalTax),
		})
	}

	paymentHistory := buildPaymentHistory(s.db, invoiceID)

	discount := "0.00"
	if inv.Discount != nil {
		discount = fmt.Sprintf("%.2f", *inv.Discount)
	}
	giftCardBal := "0.00"
	if inv.GiftCardBal != nil {
		giftCardBal = fmt.Sprintf("%.2f", *inv.GiftCardBal)
	}

	return map[string]interface{}{
		"invoice_id":       strconv.FormatInt(inv.IDInvoice, 10),
		"invoice_number":   inv.NumberInvoice,
		"sold_by":          creatorName,
		"invoice_date":     inv.CreatedAt.Format("01/02/2006"),
		"patient_name":     patientName,
		"patient_address":  patientAddress,
		"patient_state":    patientState,
		"patient_phone":    phones,
		"payment_method":   paymentMethodName,
		"items":            itemsData,
		"total_amount":     fmt.Sprintf("%.2f", inv.TotalAmount),
		"final_amount":     fmt.Sprintf("%.2f", inv.FinalAmount),
		"pt_balance":       fmt.Sprintf("%.2f", inv.PTBal),
		"insurance_balance": fmt.Sprintf("%.2f", inv.InsBal),
		"gift_card_balance": giftCardBal,
		"tax_amount":       fmt.Sprintf("%.2f", inv.TaxAmount),
		"due_amount":       fmt.Sprintf("%.2f", inv.Due),
		"discount":         discount,
		"insurance_policy": insuranceInfo,
		"payment_history":  paymentHistory,
		"location":         buildLocationData(inv.Location),
		"barcode":          "",
		"qrcode":           "",
		// patient address fields
		"patient_address_street_1": derefStr(patient.StreetAddress),
		"patient_address_street_2": derefStr(patient.AddressLine2),
		"patient_address_city":     derefStr(patient.City),
		"patient_address_state":    derefStr(patient.State),
		"patient_address_zip":      derefStr(patient.ZipCode),
	}, nil
}

// RenderInvoiceHTML renders the invoice HTML template and returns the HTML string.
func (s *Service) RenderInvoiceHTML(invoiceID int64) (string, error) {
	ctx, err := s.BuildInvoiceHTMLContext(invoiceID)
	if err != nil {
		return "", err
	}

	templatesDir := os.Getenv("PDF_TEMPLATES_DIR")
	if templatesDir == "" {
		templatesDir = "internal/templates/pdf"
	}
	tmplPath := filepath.Join(templatesDir, "invoice.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("template execute error: %w", err)
	}
	return buf.String(), nil
}

// GenerateInvoicePDF renders the HTML template and converts it to PDF via wkhtmltopdf.
func (s *Service) GenerateInvoicePDF(invoiceID int64) ([]byte, string, error) {
	htmlContent, err := s.RenderInvoiceHTML(invoiceID)
	if err != nil {
		return nil, "", err
	}

	// Get invoice number for filename
	var inv invoices.Invoice
	s.db.Select("number_invoice").First(&inv, invoiceID)

	cmd := exec.Command("wkhtmltopdf", "--quiet", "-", "-")
	cmd.Stdin = strings.NewReader(htmlContent)
	pdfBytes, err := cmd.Output()
	if err != nil {
		return nil, "", fmt.Errorf("pdf generation failed: %w", err)
	}
	return pdfBytes, inv.NumberInvoice, nil
}

// ─── GET /lookup ──────────────────────────────────────────────────────────────

func (s *Service) LookupBySKU(username, rawSKU string) (map[string]interface{}, error) {
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	normalized := pkgSKU.Normalize(rawSKU)

	// Find inventory in current location or its warehouse
	var frame invModel.Inventory
	query := s.db.Where("sku = ? AND (location_id = ?", normalized, loc.IDLocation)
	if loc.WarehouseID != nil {
		query = s.db.Where("sku = ? AND (location_id = ? OR location_id = ?)", normalized, loc.IDLocation, *loc.WarehouseID)
	}
	if err := query.First(&frame).Error; err != nil {
		return nil, errors.New("item not found")
	}

	if frame.ModelID == nil {
		return nil, errors.New("model not found for this item")
	}
	var model frameModel.Model
	if err := s.db.First(&model, *frame.ModelID).Error; err != nil {
		return nil, errors.New("model not found for this item")
	}
	var product frameModel.Product
	if err := s.db.First(&product, model.ProductID).Error; err != nil {
		return nil, errors.New("product not found for this model")
	}

	brandName := ""
	if product.BrandID != nil {
		var brand vendorModel.Brand
		if s.db.First(&brand, *product.BrandID).Error == nil && brand.BrandName != nil {
			brandName = *brand.BrandName
		}
	}

	var pbe invModel.PriceBook
	var sellingPrice *float64
	if s.db.Where("inventory_id = ?", frame.IDInventory).First(&pbe).Error == nil {
		sellingPrice = pbe.PbSellingPrice
	}

	description := strings.TrimSpace(fmt.Sprintf("%s %s %s", brandName, product.TitleProduct, model.TitleVariant))

	result := map[string]interface{}{
		"pb_key":      "Frames",
		"item_id":     frame.IDInventory,
		"price":       sellingPrice,
		"item_name":   frame.SKU,
		"description": description,
	}

	// Auto-transfer from warehouse to showcase check
	if loc.WarehouseID != nil && frame.LocationID == int64(*loc.WarehouseID) {
		if loc.Showcase == nil || !*loc.Showcase {
			result["auto_transfer"] = "showcase"
		}
	}

	return result, nil
}

// ─── GET /invoice/statuses ────────────────────────────────────────────────────

func (s *Service) GetInvoiceStatuses() ([]map[string]interface{}, error) {
	var statuses []invoices.StatusInvoice
	if err := s.db.Where("id_status_invoice NOT IN (24, 25, 26, 27)").
		Order("id_status_invoice").Find(&statuses).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, len(statuses))
	for i, st := range statuses {
		result[i] = map[string]interface{}{
			"status_invoice_id": st.IDStatusInvoice,
			"status_invoice":    st.StatusInvoiceValue,
		}
	}
	return result, nil
}

// ─── PUT /invoice/{id} (add items) ───────────────────────────────────────────

func (s *Service) AddItemsToInvoice(username string, invoiceID int64, input AddItemsInput) (map[string]interface{}, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return nil, errors.New("invoice not found")
	}
	if inv.Finalized {
		return nil, errors.New("invoice is finalized (locked) and cannot be updated")
	}

	pbKey := strings.TrimSpace(input.PbKey)
	if pbKey == "" {
		return nil, errors.New("pb_key is required")
	}

	var newItemIDs []int64

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Update notes if provided
		if input.Notes != nil {
			if err := tx.Model(&inv).Update("notified", *input.Notes).Error; err != nil {
				return err
			}
			inv.Notified = input.Notes
		}

		// Get sales tax rate for this location
		taxRate := 0.0
		if loc.SalesTaxID != nil {
			var taxRecord generalModel.SalesTaxByState
			if tx.First(&taxRecord, *loc.SalesTaxID).Error == nil {
				taxRate = taxRecord.SalesTaxPercent / 100
			}
		}

		globalSaleKey := normKey(input.SaleKey)

		for _, itemData := range input.Items {
			var (
				itemID       *int64
				description  string
				price        = 0.0
				lineCost     = 0.0
				existingItem *invoices.InvoiceItemSale
			)
			quantity := itemData.Quantity
			if quantity == 0 {
				quantity = 1
			}
			discount := itemData.Discount
			taxable := parseBool(itemData.Taxable)
			lineSaleKey := normKey(itemData.SaleKey)

			if lineSaleKey == nil {
				lineSaleKey = globalSaleKey
			}
			if itemData.Price != nil {
				price = *itemData.Price
			}

			switch pbKey {
			case "Frames":
				rawID, hasID := parseInt64(itemData.ItemID)
				if !hasID || rawID == 0 {
					// Dummy frame (item_id=0)
					i0 := int64(0)
					itemID = &i0
					description = strings.TrimSpace(itemData.Description)
					if description == "" {
						return errors.New("frames dummy item_id=0 requires non-empty 'description'")
					}
					lineCost = round2(itemData.Cost * quantity)
				} else {
					var frame invModel.Inventory
					if tx.First(&frame, rawID).Error != nil {
						continue
					}
					if frame.StatusItemsInventory != "Ready for Sale" && frame.StatusItemsInventory != "Ordered" {
						return fmt.Errorf("item %d not ready for sale: %s", rawID, frame.StatusItemsInventory)
					}
					// Check missing item
					var missingItem invModel.Missing
					if tx.Where("inventory_id = ? AND location_id = ?", rawID, loc.IDLocation).First(&missingItem).Error == nil {
						var ic invModel.InventoryCount
						if tx.First(&ic, missingItem.InventoryCountID).Error == nil && !ic.Status {
							return fmt.Errorf("item %d is marked as missing", rawID)
						}
					}

					// Get description from brand/product/model
					if frame.ModelID != nil {
						var model frameModel.Model
						var product frameModel.Product
						var brand vendorModel.Brand
						tx.First(&model, *frame.ModelID)
						tx.First(&product, model.ProductID)
						brandName := ""
						if product.BrandID != nil {
							if tx.First(&brand, *product.BrandID).Error == nil && brand.BrandName != nil {
								brandName = *brand.BrandName
							}
						}
						description = strings.TrimSpace(fmt.Sprintf("%s %s %s", brandName, product.TitleProduct, model.TitleVariant))
					}

					fid := rawID
					itemID = &fid
					var pbe invModel.PriceBook
					if tx.Where("inventory_id = ?", rawID).First(&pbe).Error == nil {
						if pbe.PbSellingPrice != nil {
							price = *pbe.PbSellingPrice
						}
						if pbe.ItemListCost != nil {
							lineCost = round2(*pbe.ItemListCost * quantity)
						}
					}
				}

			case "Lens":
				lineSaleKey = ptrStr("Lens Options")
				rawID, ok := parseInt64(itemData.ItemID)
				if !ok {
					continue
				}
				var lens lensModel.Lenses
				if tx.First(&lens, rawID).Error != nil {
					continue
				}
				lid := int64(lens.IDLenses)
				itemID = &lid
				description = derefStr(lens.Description)
				if description == "" {
					description = lens.LensName
				}
				quantity *= 2
				if itemData.Price != nil {
					price = *itemData.Price / 2
				} else if lens.Price != nil {
					price = *lens.Price / 2
				}
				if lens.Cost != nil {
					lineCost = round2(*lens.Cost / 2 * quantity)
				}

			case "Contact Lens":
				rawID, ok := parseInt64(itemData.ItemID)
				if !ok {
					continue
				}
				var cl clModel.ContactLensItem
				if tx.First(&cl, rawID).Error != nil {
					continue
				}
				clid := int64(cl.IDContactLensItem)
				itemID = &clid
				description = derefStr(cl.InvoiceDesc)
				if description == "" {
					description = "Contact Lens"
				}
				if itemData.Price != nil {
					price = *itemData.Price
				} else if cl.SellingPrice != nil {
					price = *cl.SellingPrice
				}
				if cl.Cost != nil {
					lineCost = round2(*cl.Cost * quantity)
				}
				// Check for existing item to update
				if rawLineID, ok := parseInt64(itemData.IDInvoiceSale); ok && rawLineID != 0 {
					var existing invoices.InvoiceItemSale
					if tx.Where("id_invoice_sale = ? AND invoice_id = ? AND item_type = ?", rawLineID, invoiceID, pbKey).First(&existing).Error == nil {
						if existing.ItemID != nil && *existing.ItemID != *itemID {
							return fmt.Errorf("contact Lens item_id mismatch")
						}
						existingItem = &existing
					} else {
						return fmt.Errorf("invoice item not found: id_invoice_sale=%d", rawLineID)
					}
				}

			case "Treatment":
				rawID, ok := parseInt64(itemData.ItemID)
				if !ok {
					continue
				}
				var tr lensModel.LensTreatments
				if tx.First(&tr, rawID).Error != nil {
					continue
				}
				trid := tr.IDLensTreatments
				itemID = &trid
				description = derefStr(tr.Description)
				if description == "" {
					description = "Treatment"
				}
				quantity *= 2
				if itemData.Price != nil {
					price = *itemData.Price / 2
				} else if tr.Price != nil {
					price = *tr.Price / 2
				}
				if tr.Cost != nil {
					lineCost = round2(*tr.Cost / 2 * quantity)
				}

			case "Prof. service":
				rawID, ok := parseInt64(itemData.ItemID)
				if !ok {
					continue
				}
				var srv svcModel.ProfessionalService
				if tx.First(&srv, rawID).Error != nil {
					continue
				}
				srvid := srv.IDProfessionalService
				itemID = &srvid
				description = derefStr(srv.InvoiceDesc)
				price = srv.Price
				lineCost = round2(srv.Cost * quantity)

			case "Add service":
				rawID, ok := parseInt64(itemData.ItemID)
				if !ok {
					continue
				}
				var asrv svcModel.AdditionalService
				if tx.First(&asrv, rawID).Error != nil {
					continue
				}
				asrvid := asrv.IDAdditionalService
				itemID = &asrvid
				description = asrv.InvoiceDesc
				price = asrv.Price
				lineCost = round2(asrv.CostPrice * quantity)

			case "misc":
				rawID, ok := parseInt64(itemData.ItemID)
				if !ok {
					continue
				}
				var misc miscModel.MiscInvoiceItem
				if tx.First(&misc, rawID).Error != nil {
					continue
				}
				miscid := misc.IDMiscItem
				itemID = &miscid
				if lineSaleKey == nil {
					if misc.SaleKey != nil {
						lineSaleKey = misc.SaleKey
					} else {
						lineSaleKey = ptrStr(pbKey)
					}
				}
				description = misc.Description
				if itemData.Price != nil {
					price = *itemData.Price
				} else if misc.Price != nil {
					if v, err := strconv.ParseFloat(*misc.Price, 64); err == nil {
						price = v
					}
				}
				if misc.Cost != nil {
					if v, err := strconv.ParseFloat(*misc.Cost, 64); err == nil {
						lineCost = round2(v * quantity)
					}
				}
				taxable = false

			default:
				if itemID == nil {
					continue
				}
			}

			// Compute total and tax
			total := round2(quantity*price - discount)
			taxAmount := 0.0
			if taxable {
				taxAmount = round4(total * taxRate)
			}

			if lineSaleKey == nil && pbKey != "Lens" {
				lineSaleKey = ptrStr(pbKey)
			}

			ptBal := total
			insBal := 0.0

			if existingItem != nil {
				// Update Contact Lens existing item
				existingItem.Quantity = int(quantity)
				existingItem.Price = price
				existingItem.Discount = discount
				existingItem.Total = total
				existingItem.TotalTax = taxAmount
				existingItem.Cost = lineCost
				existingItem.PtBalance = &ptBal
				existingItem.InsBalance = &insBal
				if err := tx.Save(existingItem).Error; err != nil {
					return err
				}
				newItemIDs = append(newItemIDs, existingItem.IDInvoiceSale)
			} else {
				trueBool := true
				falseBool := false
				var taxPtr *bool
				if taxable {
					taxPtr = &trueBool
				} else {
					taxPtr = &falseBool
				}

				newItem := invoices.InvoiceItemSale{
					InvoiceID:   invoiceID,
					ItemType:    pbKey,
					SaleKey:     lineSaleKey,
					ItemID:      itemID,
					Description: description,
					Quantity:    int(quantity),
					Price:       price,
					Discount:    discount,
					Total:       total,
					Taxable:     taxPtr,
					TotalTax:    taxAmount,
					Cost:        lineCost,
					PtBalance:   &ptBal,
					InsBalance:  &insBal,
				}
				if err := tx.Create(&newItem).Error; err != nil {
					return err
				}
				newItemIDs = append(newItemIDs, newItem.IDInvoiceSale)

				// Remake: add negative mirror line
				if inv.Remake {
					negPt := -ptBal
					negIns := -insBal
					negDiscount := -discount
					negItem := invoices.InvoiceItemSale{
						InvoiceID:   invoiceID,
						ItemType:    pbKey,
						SaleKey:     lineSaleKey,
						ItemID:      itemID,
						Description: description,
						Quantity:    -int(quantity),
						Price:       price,
						Discount:    negDiscount,
						Total:       -total,
						Taxable:     taxPtr,
						TotalTax:    -taxAmount,
						Cost:        -lineCost,
						PtBalance:   &negPt,
						InsBalance:  &negIns,
					}
					if err := tx.Create(&negItem).Error; err != nil {
						return err
					}
				}

				// Inventory movements for Frames with real item_id
				if pbKey == "Frames" && itemID != nil && *itemID != 0 {
					var frame invModel.Inventory
					if tx.First(&frame, *itemID).Error == nil {
						oldInvoiceID := frame.InvoiceID

						// Transfer from warehouse to showcase if needed
						if loc.WarehouseID != nil && frame.LocationID == int64(*loc.WarehouseID) {
							whID := int64(*loc.WarehouseID)
							locID := int64(loc.IDLocation)
							s.addInventoryTx(tx, frame.IDInventory, &whID, &locID,
								int64(emp.IDEmployee), oldInvoiceID, &oldInvoiceID,
								"TRANSFERRED TO SHOWCASE", "Transfer")
							frame.LocationID = int64(loc.IDLocation)
						}

						if frame.StatusItemsInventory != "Ordered" {
							frame.StatusItemsInventory = "SOLD"
						}
						frame.InvoiceID = invoiceID
						tx.Save(&frame)

						// Auto-count: if frame is in a count sheet, mark counted
						var missingInCount invModel.Missing
						if tx.Where("inventory_id = ? AND location_id = ?", frame.IDInventory, loc.IDLocation).First(&missingInCount).Error == nil {
							var activeIC invModel.InventoryCount
							if tx.Where("id_inventory_count = ? AND status = true", missingInCount.InventoryCountID).First(&activeIC).Error == nil {
								locIntID := loc.IDLocation
								brandIntID := 0
								if missingInCount.BrandID != nil {
									brandIntID = int(*missingInCount.BrandID)
								}
								var vendorIntID *int
							if missingInCount.VendorID != nil {
								v := int(*missingInCount.VendorID)
								vendorIntID = &v
							}
							tx.Create(&invModel.TempCountInventory{
									InventoryID:      frame.IDInventory,
									LocationID:       locIntID,
									BrandID:          &brandIntID,
									VendorID:         vendorIntID,
									InStock:          true,
									InventoryCountID: activeIC.IDInventoryCount,
									CountDate:        time.Now(),
								})
								icID := activeIC.IDInventoryCount
								s.addInventoryTx(tx, frame.IDInventory,
									int64Ptr(int64(loc.IDLocation)), nil,
									int64(emp.IDEmployee), invoiceID, &icID,
									"SOLD", "Count Sheet: Counted")
								tx.Delete(&missingInCount)
							}
						}

						// Sale transaction
						newStatus := string(frame.StatusItemsInventory)
						fromLocID := int64(loc.IDLocation)
						s.addInventoryTx(tx, frame.IDInventory,
							&fromLocID, nil,
							int64(emp.IDEmployee), invoiceID, &oldInvoiceID,
							newStatus, "Sale")
					}
				}
			}
		}

		return s.recalcInvoice(tx, &inv)
	})

	if err != nil {
		return nil, err
	}

	resp := map[string]interface{}{"message": "Invoice updated successfully"}
	if len(newItemIDs) > 0 {
		resp["item_sale_ids"] = newItemIDs
	}
	return resp, nil
}

func int64Ptr(v int64) *int64 { return &v }

// ─── GET /invoices/search ─────────────────────────────────────────────────────

func (s *Service) SearchInvoices(username, q string) ([]map[string]interface{}, error) {
	if q == "" {
		return []map[string]interface{}{}, nil
	}
	_, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}

	var invList []invoices.Invoice
	s.db.Preload("Patient").Preload("StatusInvoice").
		Where("location_id = ? AND number_invoice ILIKE ?", loc.IDLocation, "%"+q+"%").
		Order("date_create DESC").Limit(20).Find(&invList)

	result := make([]map[string]interface{}, 0, len(invList))
	for _, inv := range invList {
		patientName := (*string)(nil)
		if inv.Patient != nil {
			name := inv.Patient.FirstName + " " + inv.Patient.LastName
			patientName = &name
		}
		statusVal := (*string)(nil)
		if inv.StatusInvoice != nil {
			statusVal = &inv.StatusInvoice.StatusInvoiceValue
		}
		result = append(result, map[string]interface{}{
			"id_invoice":     inv.IDInvoice,
			"number_invoice": inv.NumberInvoice,
			"date_create":    inv.DateCreate.Format(time.RFC3339),
			"patient_id":     inv.PatientID,
			"patient_name":   patientName,
			"due":            fmt.Sprintf("%.2f", inv.Due),
			"total_amount":   fmt.Sprintf("%.2f", inv.TotalAmount),
			"status":         statusVal,
		})
	}
	return result, nil
}

// ─── DELETE /invoice/{id} ─────────────────────────────────────────────────────

func (s *Service) DeleteInvoice(username string, invoiceID int64) error {
	_, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return err
	}

	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return errors.New("invoice not found")
	}
	if inv.Finalized {
		return errors.New("invoice is finalized (locked) and cannot be updated")
	}

	var itemsCount int64
	s.db.Model(&invoices.InvoiceItemSale{}).Where("invoice_id = ?", invoiceID).Count(&itemsCount)
	if itemsCount > 0 {
		return errors.New("cannot delete invoice with items")
	}

	if inv.TotalAmount != 0 || inv.FinalAmount != 0 {
		return errors.New("cannot delete invoice with non-zero amounts")
	}

	var ticketsCount int64
	s.db.Model(&labModel.LabTicket{}).Where("invoice_id = ?", invoiceID).Count(&ticketsCount)
	if ticketsCount > 0 {
		return errors.New("cannot delete invoice with associated lab tickets")
	}

	return s.db.Delete(&inv).Error
}

// ─── PUT /invoice/finalize/{id} ───────────────────────────────────────────────

func (s *Service) FinalizeInvoice(invoiceID int64) error {
	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return errors.New("invoice not found")
	}
	if inv.Finalized {
		return errors.New("invoice is already finalized")
	}
	return s.db.Model(&inv).Update("finalized", true).Error
}

// ─── PUT /invoice/unfinalize/{id} ─────────────────────────────────────────────

func (s *Service) UnfinalizeInvoice(invoiceID int64) error {
	var inv invoices.Invoice
	if err := s.db.First(&inv, invoiceID).Error; err != nil {
		return errors.New("invoice not found")
	}
	if !inv.Finalized {
		return errors.New("invoice is already unlocked")
	}
	return s.db.Model(&inv).Update("finalized", false).Error
}
