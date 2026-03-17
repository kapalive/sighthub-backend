package super_service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	authModel      "sighthub-backend/internal/models/auth"
	empModel       "sighthub-backend/internal/models/employees"
	invoiceModel   "sighthub-backend/internal/models/invoices"
	assessModel    "sighthub-backend/internal/models/medical/vision_exam/assessment"
	superModel     "sighthub-backend/internal/models/medical/vision_exam/super_bill"
	locModel       "sighthub-backend/internal/models/location"
	svcModel       "sighthub-backend/internal/models/service"
	visionModel    "sighthub-backend/internal/models/vision_exam"
	pkgInvoice     "sighthub-backend/pkg/invoice"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

// ─── helpers ──────────────────────────────────────────────────────────────────

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

func (s *Service) validateExamOwnership(emp *empModel.Employee, examID int64) (*visionModel.EyeExam, error) {
	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	if exam.EmployeeID != int64(emp.IDEmployee) {
		return nil, errors.New("not authorized to update this exam")
	}
	return &exam, nil
}

// ─── input types ──────────────────────────────────────────────────────────────

type DiseaseInput struct {
	Code     *string `json:"code"`
	Title    *string `json:"title"`
	LevelID  *int64  `json:"level_id"`
	Type     *string `json:"type"`
	GroupSet *string `json:"group_set"`
}

type ServiceItem struct {
	Quantity int            `json:"quantity"`
	PbKey    string         `json:"pb_key"`
	Diseases []DiseaseInput `json:"diseases"`
}

type CreateSuperInput struct {
	CptHcpcsCode map[string]ServiceItem `json:"cpt_hcpcs_code"`
}

// same shape for invoice + update
type InvoiceInput = CreateSuperInput

// ─── create super eye exam ────────────────────────────────────────────────────

func (s *Service) CreateSuperEyeExam(username string, examID int64, input CreateSuperInput) (map[string]interface{}, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	exam, err := s.validateExamOwnership(emp, examID)
	if err != nil {
		return nil, err
	}
	if exam.Passed {
		return nil, errors.New("cannot create super bill for a completed exam")
	}
	if len(input.CptHcpcsCode) == 0 {
		return nil, errors.New("'cpt_hcpcs_code' object is required")
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	newSuper := superModel.SuperEyeExam{EyeExamID: examID}
	if err := tx.Create(&newSuper).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for serviceKey, item := range input.CptHcpcsCode {
		serviceID, err := strconv.ParseInt(serviceKey, 10, 64)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("invalid service key: %s", serviceKey)
		}
		var svc svcModel.ProfessionalService
		if err := tx.First(&svc, serviceID).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("professional service with id %s not found", serviceKey)
		}

		for _, d := range item.Diseases {
			code := "NO_CODE"
			if d.Code != nil && *d.Code != "" {
				code = *d.Code
			}
			title := "NO_TITLE"
			if d.Title != nil && *d.Title != "" {
				title = *d.Title
			}
			var levelID int64
			if d.LevelID != nil {
				levelID = *d.LevelID
			}
			dType := "custom"
			if d.Type != nil && *d.Type != "" {
				dType = *d.Type
			}

			sbd := superModel.SuperBillDiagnosis{
				SuperEyeExamID:        newSuper.IDSuperEyeExam,
				ProfessionalServiceID: serviceID,
			}
			if err := tx.Create(&sbd).Error; err != nil {
				tx.Rollback()
				return nil, err
			}

			sbdID := sbd.IDSuperBillDiagnosis
			defFalse := false
			disease := superModel.DiseaseSuperBill{
				LevelID:              levelID,
				Type:                 dType,
				Code:                 code,
				Title:                title,
				GroupSet:             d.GroupSet,
				Default:              &defFalse,
				SuperBillDiagnosisID: &sbdID,
				Include:              true,
			}
			if err := tx.Create(&disease).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message":           "SuperEyeExam created successfully",
		"super_eye_exam_id": newSuper.IDSuperEyeExam,
	}, nil
}

// ─── create super invoice ─────────────────────────────────────────────────────

func (s *Service) CreateSuperInvoice(username string, examID int64, input InvoiceInput) (map[string]interface{}, error) {
	emp, loc, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	var exam visionModel.EyeExam
	if err := s.db.First(&exam, examID).Error; err != nil {
		return nil, errors.New("exam not found")
	}
	if exam.Passed {
		return nil, errors.New("cannot create invoice for a completed exam")
	}

	var superExam superModel.SuperEyeExam
	if err := s.db.Where("eye_exam_id = ?", examID).First(&superExam).Error; err != nil {
		return nil, errors.New("super eye exam not found")
	}
	if superExam.InvoiceID != nil {
		return nil, errors.New("invoice already exists for this super bill")
	}

	// If no cpt_hcpcs_code provided, auto-populate from existing super bill diagnoses
	if len(input.CptHcpcsCode) == 0 {
		type row struct{ ProfessionalServiceID int64 }
		var rows []row
		s.db.Table("super_bill_diagnosis").
			Select("DISTINCT professional_service_id").
			Where("super_eye_exam_id = ?", superExam.IDSuperEyeExam).
			Scan(&rows)
		if len(rows) == 0 {
			return nil, errors.New("no services found in super bill")
		}
		input.CptHcpcsCode = make(map[string]ServiceItem, len(rows))
		for _, r := range rows {
			input.CptHcpcsCode[strconv.FormatInt(r.ProfessionalServiceID, 10)] = ServiceItem{Quantity: 1}
		}
	}

	shortName := ""
	if loc.ShortName != nil {
		shortName = *loc.ShortName
	}
	invoiceNumber, err := pkgInvoice.CreateInvoiceNumber(s.db, "S", shortName)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	discount := 0.0
	empID := int64(emp.IDEmployee)
	now := time.Now()
	newInvoice := invoiceModel.Invoice{
		NumberInvoice: invoiceNumber,
		DateCreate:    now,
		CreatedAt:     now,
		EmployeeID:    &empID,
		LocationID:    int64(exam.LocationID),
		PatientID:     exam.PatientID,
		TotalAmount:   0,
		PTBal:         0,
		InsBal:        0,
		Due:           0,
		TaxAmount:     0,
		Discount:      &discount,
		Finalized:     false,
	}
	if err := tx.Create(&newInvoice).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Link SuperEyeExam to invoice
	if err := tx.Model(&superExam).Update("invoice_id", newInvoice.IDInvoice).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var totalAmount, ptBal, insBal, due float64

	for serviceKey, item := range input.CptHcpcsCode {
		serviceID, err := strconv.ParseInt(serviceKey, 10, 64)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("invalid service key: %s", serviceKey)
		}
		var svc svcModel.ProfessionalService
		if err := tx.First(&svc, serviceID).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("professional service with id %s not found", serviceKey)
		}

		qty := item.Quantity
		if qty == 0 {
			qty = 1
		}
		price := svc.Price
		total := price * float64(qty)

		desc := ""
		if svc.InvoiceDesc != nil {
			desc = *svc.InvoiceDesc
		}
		totalTax := 0.0
		ptBal0 := 0.0
		taxable := false
		invItem := invoiceModel.InvoiceItemSale{
			InvoiceID:   newInvoice.IDInvoice,
			ItemType:    "Prof. service",
			ItemID:      &serviceID,
			Description: desc,
			Quantity:    qty,
			Price:       price,
			Total:       total,
			Taxable:     &taxable,
			TotalTax:    totalTax,
			PtBalance:   &ptBal0,
			InsBalance:  &total,
		}
		if err := tx.Create(&invItem).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		totalAmount += total
		insBal += total
		due += total
	}

	if err := tx.Model(&newInvoice).Updates(map[string]interface{}{
		"total_amount": totalAmount,
		"pt_bal":       ptBal,
		"ins_bal":      insBal,
		"due":          due,
	}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message":        "Invoice created successfully",
		"invoice_id":     newInvoice.IDInvoice,
		"invoice_number": newInvoice.NumberInvoice,
		"total_amount":   fmt.Sprintf("%.2f", totalAmount),
		"pt_bal":         fmt.Sprintf("%.2f", ptBal),
		"ins_bal":        fmt.Sprintf("%.2f", insBal),
		"due":            fmt.Sprintf("%.2f", due),
	}, nil
}

// ─── update super eye exam ────────────────────────────────────────────────────

func (s *Service) UpdateSuperEyeExam(username string, examID int64, input InvoiceInput) (map[string]interface{}, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	exam, err := s.validateExamOwnership(emp, examID)
	if err != nil {
		return nil, err
	}
	if exam.Passed {
		return nil, errors.New("exam has already been completed")
	}
	if len(input.CptHcpcsCode) == 0 {
		return nil, errors.New("'cpt_hcpcs_code' object is required")
	}

	var superExam superModel.SuperEyeExam
	if err := s.db.Where("eye_exam_id = ?", examID).First(&superExam).Error; err != nil {
		return nil, errors.New("super eye exam not found")
	}

	var invoice *invoiceModel.Invoice
	if superExam.InvoiceID != nil {
		var inv invoiceModel.Invoice
		if err := s.db.First(&inv, *superExam.InvoiceID).Error; err == nil {
			invoice = &inv
		}
	}
	if invoice != nil && invoice.Finalized {
		return nil, errors.New("cannot modify a finalized invoice")
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for serviceKey, item := range input.CptHcpcsCode {
		serviceID, err := strconv.ParseInt(serviceKey, 10, 64)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("invalid service key: %s", serviceKey)
		}
		var svc svcModel.ProfessionalService
		if err := tx.First(&svc, serviceID).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("professional service with item_id %s not found", serviceKey)
		}

		qty := item.Quantity
		if qty == 0 {
			qty = 1
		}
		price := svc.Price
		total := price * float64(qty)

		// Update/create invoice item if invoice exists
		if invoice != nil {
			pbKey := item.PbKey
			if pbKey == "" {
				pbKey = "Prof. service"
			}
			var invItem invoiceModel.InvoiceItemSale
			err := tx.Where("invoice_id = ? AND item_type = ? AND item_id = ?",
				invoice.IDInvoice, pbKey, serviceID).First(&invItem).Error

			if err == nil {
				// Update existing
				ptBal := total
				insBal := 0.0
				if invItem.InsBalance != nil {
					insBal = *invItem.InsBalance
					ptBal = total - insBal
				}
				if err := tx.Model(&invItem).Updates(map[string]interface{}{
					"quantity":   qty,
					"price":      price,
					"total":      total,
					"pt_balance": ptBal,
					"ins_balance": insBal,
				}).Error; err != nil {
					tx.Rollback()
					return nil, err
				}
			} else {
				// Create new
				desc := ""
				if svc.InvoiceDesc != nil {
					desc = *svc.InvoiceDesc
				}
				totalTax := 0.0
				insBal0 := 0.0
				taxable := false
				newItem := invoiceModel.InvoiceItemSale{
					InvoiceID:   invoice.IDInvoice,
					ItemType:    pbKey,
					ItemID:      &serviceID,
					Description: desc,
					Quantity:    qty,
					Price:       price,
					Total:       total,
					Taxable:     &taxable,
					TotalTax:    totalTax,
					PtBalance:   &total,
					InsBalance:  &insBal0,
				}
				if err := tx.Create(&newItem).Error; err != nil {
					tx.Rollback()
					return nil, err
				}
			}
		}

		// Always create new diagnoses
		for _, d := range item.Diseases {
			code := "NO_CODE"
			if d.Code != nil && *d.Code != "" {
				code = *d.Code
			}
			title := "NO_TITLE"
			if d.Title != nil && *d.Title != "" {
				title = *d.Title
			}
			var levelID int64
			if d.LevelID != nil {
				levelID = *d.LevelID
			}
			dType := "custom"
			if d.Type != nil && *d.Type != "" {
				dType = *d.Type
			}

			sbd := superModel.SuperBillDiagnosis{
				SuperEyeExamID:        superExam.IDSuperEyeExam,
				ProfessionalServiceID: serviceID,
			}
			if err := tx.Create(&sbd).Error; err != nil {
				tx.Rollback()
				return nil, err
			}

			sbdID := sbd.IDSuperBillDiagnosis
			defFalse := false
			disease := superModel.DiseaseSuperBill{
				LevelID:              levelID,
				Type:                 dType,
				Code:                 code,
				Title:                title,
				GroupSet:             d.GroupSet,
				Default:              &defFalse,
				SuperBillDiagnosisID: &sbdID,
				Include:              true,
			}
			if err := tx.Create(&disease).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	// Recalculate invoice totals
	var updatedInvoice *invoiceModel.Invoice
	if invoice != nil {
		var allItems []invoiceModel.InvoiceItemSale
		tx.Where("invoice_id = ?", invoice.IDInvoice).Find(&allItems)

		var total, ptBal, insBal float64
		for _, it := range allItems {
			total += it.Total
			if it.PtBalance != nil {
				ptBal += *it.PtBalance
			}
			if it.InsBalance != nil {
				insBal += *it.InsBalance
			}
		}
		due := ptBal + insBal
		if err := tx.Model(invoice).Updates(map[string]interface{}{
			"total_amount": total,
			"pt_bal":       ptBal,
			"ins_bal":      insBal,
			"due":          due,
		}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		invoice.TotalAmount = total
		invoice.PTBal = ptBal
		invoice.InsBal = insBal
		invoice.Due = due
		updatedInvoice = invoice
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	resp := map[string]interface{}{"message": "SuperEyeExam updated successfully"}
	if updatedInvoice != nil {
		resp["invoice_id"] = updatedInvoice.IDInvoice
		resp["invoice_number"] = updatedInvoice.NumberInvoice
		resp["total_amount"] = fmt.Sprintf("%.2f", updatedInvoice.TotalAmount)
		resp["pt_bal"] = fmt.Sprintf("%.2f", updatedInvoice.PTBal)
		resp["ins_bal"] = fmt.Sprintf("%.2f", updatedInvoice.InsBal)
		resp["due"] = fmt.Sprintf("%.2f", updatedInvoice.Due)
	}
	return resp, nil
}

// ─── get super eye exam ───────────────────────────────────────────────────────

func (s *Service) GetSuperEyeExam(examID int64) (map[string]interface{}, error) {
	var superExam superModel.SuperEyeExam
	if err := s.db.Where("eye_exam_id = ?", examID).First(&superExam).Error; err != nil {
		return map[string]interface{}{
			"exists":        false,
			"cpt_hcpcs_code": nil,
			"invoice_id":    nil,
		}, nil
	}

	var invoice *invoiceModel.Invoice
	if superExam.InvoiceID != nil {
		var inv invoiceModel.Invoice
		if err := s.db.First(&inv, *superExam.InvoiceID).Error; err == nil {
			invoice = &inv
		}
	}

	collectDiseases := func(serviceID int64) []map[string]interface{} {
		var sbds []superModel.SuperBillDiagnosis
		s.db.Where("professional_service_id = ? AND super_eye_exam_id = ?",
			serviceID, superExam.IDSuperEyeExam).Find(&sbds)

		result := []map[string]interface{}{}
		for _, sbd := range sbds {
			var diseases []superModel.DiseaseSuperBill
			s.db.Where("super_bill_diagnosis_id = ?", sbd.IDSuperBillDiagnosis).Find(&diseases)
			for _, dr := range diseases {
				result = append(result, map[string]interface{}{
					"id_disease_super_bill": dr.IDDiseaseSuperBill,
					"level_id":              dr.LevelID,
					"type":                  dr.Type,
					"code":                  dr.Code,
					"title":                 dr.Title,
					"group_set":             dr.GroupSet,
					"include":               dr.Include,
				})
			}
		}
		return result
	}

	cptCode := map[string]interface{}{}

	if invoice != nil {
		var items []invoiceModel.InvoiceItemSale
		s.db.Where("invoice_id = ?", invoice.IDInvoice).Find(&items)
		for _, item := range items {
			if item.ItemID == nil {
				continue
			}
			sid := *item.ItemID
			var svc svcModel.ProfessionalService
			s.db.First(&svc, sid)
			ptBal := item.Total
			if item.PtBalance != nil {
				ptBal = *item.PtBalance
			}
			insBal := 0.0
			if item.InsBalance != nil {
				insBal = *item.InsBalance
			}
			cptCode[strconv.FormatInt(sid, 10)] = map[string]interface{}{
				"item_id":      sid,
				"cpt_hcpcs_code": svc.CptHcpcsCode,
				"description":  svc.InvoiceDesc,
				"quantity":     item.Quantity,
				"price":        fmt.Sprintf("%.2f", item.Price),
				"total":        fmt.Sprintf("%.2f", item.Total),
				"pt_balance":   fmt.Sprintf("%.2f", ptBal),
				"ins_balance":  fmt.Sprintf("%.2f", insBal),
				"include":      true,
				"diseases":     collectDiseases(sid),
			}
		}
	} else {
		// No invoice — get services from SuperBillDiagnosis
		type Row struct {
			ProfessionalServiceID int64
		}
		var rows []Row
		s.db.Table("super_bill_diagnosis").
			Select("DISTINCT professional_service_id").
			Where("super_eye_exam_id = ?", superExam.IDSuperEyeExam).
			Scan(&rows)

		for _, row := range rows {
			sid := row.ProfessionalServiceID
			var svc svcModel.ProfessionalService
			s.db.First(&svc, sid)
			cptCode[strconv.FormatInt(sid, 10)] = map[string]interface{}{
				"item_id":      sid,
				"cpt_hcpcs_code": svc.CptHcpcsCode,
				"description":  svc.InvoiceDesc,
				"quantity":     nil,
				"price":        fmt.Sprintf("%.2f", svc.Price),
				"total":        nil,
				"pt_balance":   nil,
				"ins_balance":  nil,
				"include":      nil,
				"diseases":     collectDiseases(sid),
			}
		}
	}

	resp := map[string]interface{}{
		"exists":        true,
		"invoice_id":    nil,
		"invoice_number": nil,
		"total_amount":  nil,
		"pt_bal":        "0.00",
		"ins_bal":       "0.00",
		"due":           nil,
		"cpt_hcpcs_code": cptCode,
	}
	if invoice != nil {
		resp["invoice_id"] = invoice.IDInvoice
		resp["invoice_number"] = invoice.NumberInvoice
		resp["total_amount"] = fmt.Sprintf("%.2f", invoice.TotalAmount)
		resp["pt_bal"] = fmt.Sprintf("%.2f", invoice.PTBal)
		resp["ins_bal"] = fmt.Sprintf("%.2f", invoice.InsBal)
		resp["due"] = fmt.Sprintf("%.2f", invoice.Due)
	}
	return resp, nil
}

// ─── delete diagnosis ─────────────────────────────────────────────────────────

func (s *Service) DeleteDiagnosisByID(username string, examID, itemID, diseaseSuperBillID int64) (map[string]interface{}, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	exam, err := s.validateExamOwnership(emp, examID)
	if err != nil {
		return nil, err
	}
	if exam.Passed {
		return nil, errors.New("exam has already been completed")
	}

	var superExam superModel.SuperEyeExam
	if err := s.db.Where("eye_exam_id = ?", examID).First(&superExam).Error; err != nil {
		return nil, errors.New("super eye exam not found")
	}

	var disease superModel.DiseaseSuperBill
	if err := s.db.First(&disease, diseaseSuperBillID).Error; err != nil {
		return nil, fmt.Errorf("no DiseaseSuperBill found with id '%d'", diseaseSuperBillID)
	}
	if disease.Default != nil && *disease.Default {
		return nil, errors.New("cannot delete a default disease")
	}
	if disease.SuperBillDiagnosisID == nil {
		return nil, errors.New("no SuperBillDiagnosis linked to this DiseaseSuperBill")
	}

	var sbd superModel.SuperBillDiagnosis
	if err := s.db.First(&sbd, *disease.SuperBillDiagnosisID).Error; err != nil {
		return nil, errors.New("no SuperBillDiagnosis linked to this DiseaseSuperBill")
	}
	if sbd.SuperEyeExamID != superExam.IDSuperEyeExam || sbd.ProfessionalServiceID != itemID {
		return nil, errors.New("this disease does not match the given exam or item")
	}

	tx := s.db.Begin()
	if err := tx.Delete(&disease).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Delete(&sbd).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":               fmt.Sprintf("DiseaseSuperBill with id '%d' and linked SuperBillDiagnosis deleted successfully", diseaseSuperBillID),
		"exam_id":               examID,
		"item_id":               itemID,
		"disease_super_bill_id": diseaseSuperBillID,
	}, nil
}

// ─── delete item from invoice ─────────────────────────────────────────────────

func (s *Service) DeleteItemFromInvoice(username string, examID, itemID int64) (map[string]interface{}, error) {
	emp, _, err := s.getEmployeeAndLocation(username)
	if err != nil {
		return nil, err
	}
	exam, err := s.validateExamOwnership(emp, examID)
	if err != nil {
		return nil, err
	}
	if exam.Passed {
		return nil, errors.New("exam has already been completed")
	}

	var superExam superModel.SuperEyeExam
	if err := s.db.Where("eye_exam_id = ?", examID).First(&superExam).Error; err != nil {
		return nil, errors.New("super eye exam not found")
	}

	var invItem invoiceModel.InvoiceItemSale
	if err := s.db.Where("invoice_id = ? AND item_id = ? AND item_type = ?",
		superExam.InvoiceID, itemID, "Prof. service").First(&invItem).Error; err != nil {
		return nil, errors.New("item not found in invoice")
	}

	tx := s.db.Begin()

	// Delete related SuperBillDiagnosis + DiseaseSuperBill
	var sbds []superModel.SuperBillDiagnosis
	tx.Where("super_eye_exam_id = ? AND professional_service_id = ?",
		superExam.IDSuperEyeExam, itemID).Find(&sbds)

	for _, sbd := range sbds {
		defFalse := false
		tx.Where("super_bill_diagnosis_id = ? AND \"default\" = ?",
			sbd.IDSuperBillDiagnosis, defFalse).Delete(&superModel.DiseaseSuperBill{})
		tx.Delete(&sbd)
	}

	deletedTotal := invItem.Total
	tx.Delete(&invItem)

	// Recalculate invoice totals
	var invoice invoiceModel.Invoice
	if superExam.InvoiceID != nil {
		if tx.First(&invoice, *superExam.InvoiceID).Error == nil {
			newTotal := invoice.TotalAmount - deletedTotal
			newDue := invoice.Due - deletedTotal
			if newTotal < 0 {
				newTotal = 0
			}
			if newDue < 0 {
				newDue = 0
			}
			tx.Model(&invoice).Updates(map[string]interface{}{
				"total_amount": newTotal,
				"due":          newDue,
			})
			invoice.TotalAmount = newTotal
			invoice.Due = newDue
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":         "Item and associated diagnoses removed successfully",
		"exam_id":         examID,
		"item_id":         itemID,
		"remaining_total": fmt.Sprintf("%.2f", invoice.TotalAmount),
		"remaining_due":   fmt.Sprintf("%.2f", invoice.Due),
	}, nil
}

// ─── professional services ────────────────────────────────────────────────────

func (s *Service) GetProfessionalServices(typeID *int) ([]map[string]interface{}, error) {
	var services []svcModel.ProfessionalService
	q := s.db.Model(&svcModel.ProfessionalService{})
	if typeID != nil {
		q = q.Where("professional_service_type_id = ?", *typeID)
	}
	if err := q.Find(&services).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(services))
	for _, svc := range services {
		result = append(result, map[string]interface{}{
			"item_id":        svc.IDProfessionalService,
			"cpt_hcpcs_code": svc.CptHcpcsCode,
			"description":    svc.InvoiceDesc,
			"pb_key":         "Prof. service",
		})
	}
	return result, nil
}

// ─── super bill diseases ──────────────────────────────────────────────────────

type SuperBillDiseaseResult struct {
	LevelID  int64    `json:"level_id"`
	Code     string   `json:"code"`
	Title    string   `json:"title"`
	Type     string   `json:"type"`
	GroupSet *string  `json:"group_set"`
	Children []string `json:"children"`
}

func (s *Service) GetSuperBillDiseases(examID int64) ([]SuperBillDiseaseResult, error) {
	var ads []assessModel.AssessmentDiagnosis
	if err := s.db.
		Joins("JOIN assessment_eye ON assessment_diagnosis_eye.assessment_eye_id = assessment_eye.id_assessment_eye").
		Where("assessment_eye.eye_exam_id = ?", examID).
		Find(&ads).Error; err != nil {
		return nil, err
	}

	type key struct {
		code    *string
		levelID *int64
		dtype   *string
	}
	seen := map[key]bool{}
	result := []SuperBillDiseaseResult{}

	for _, ad := range ads {
		k := key{ad.Code, ad.LevelID, ad.Type}
		if seen[k] {
			continue
		}
		seen[k] = true

		levelID := int64(0)
		if ad.LevelID != nil {
			levelID = *ad.LevelID
		}
		code := "NO_CODE"
		if ad.Code != nil && *ad.Code != "" {
			code = *ad.Code
		}
		title := "NO_TITLE"
		if ad.Title != nil && *ad.Title != "" {
			title = *ad.Title
		}
		dType := "custom"
		if ad.Type != nil && *ad.Type != "" {
			dType = *ad.Type
		}

		result = append(result, SuperBillDiseaseResult{
			LevelID:  levelID,
			Code:     code,
			Title:    title,
			Type:     dType,
			GroupSet: nil,
			Children: []string{},
		})
	}
	return result, nil
}
