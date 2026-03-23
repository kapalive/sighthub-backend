package vendor_service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	clModel "sighthub-backend/internal/models/contact_lens"
	frameModel "sighthub-backend/internal/models/frames"
	genModel "sighthub-backend/internal/models/general"
	lensModel "sighthub-backend/internal/models/lenses"
	otherModel "sighthub-backend/internal/models/other_products"
	vModel "sighthub-backend/internal/models/vendors"
	pkgActivity "sighthub-backend/pkg/activitylog"
)

// Service ————————————————————————————————————————————————————————————

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service { return &Service{db: db} }
func (s *Service) DB() *gorm.DB { return s.db }

// =====================================================================
// Vendor CRUD
// =====================================================================

type VendorInput struct {
	VendorName    string  `json:"vendor_name"`
	ShortName     *string `json:"short_name"`
	Phone         *string `json:"phone"`
	StreetAddress *string `json:"street_address"`
	AddressLine2  *string `json:"address_line_2"`
	City          *string `json:"city"`
	CountryID     *int    `json:"country_id"`
	StateID       *int    `json:"state_id"`
	Zip           *string `json:"zip"`
	Website       *string `json:"website"`
	Fax           *string `json:"fax"`
	Email         *string `json:"email"`
}

type RepInput struct {
	Name          *string `json:"name"`
	Title         *string `json:"title"`
	Phone         *string `json:"phone"`
	AddressLine2  *string `json:"address_line_2"`
	City          *string `json:"city"`
	CountryID     *int    `json:"country_id"`
	Email         *string `json:"email"`
	Fax           *string `json:"fax"`
	StateID       *int    `json:"state_id"`
	StreetAddress *string `json:"street_address"`
	Zip           *string `json:"zip"`
}

type AddVendorRequest struct {
	Vendor VendorInput `json:"vendor"`
	Rep    *RepInput   `json:"rep"`
}

func (s *Service) AddVendor(req AddVendorRequest) (int, error) {
	if req.Vendor.VendorName == "" {
		return 0, errors.New("Vendor name is required")
	}

	v := vModel.Vendor{
		VendorName:    req.Vendor.VendorName,
		ShortName:     req.Vendor.ShortName,
		Phone:         req.Vendor.Phone,
		StreetAddress: req.Vendor.StreetAddress,
		AddressLine2:  req.Vendor.AddressLine2,
		City:          req.Vendor.City,
		CountryID:     req.Vendor.CountryID,
		StateID:       req.Vendor.StateID,
		ZipCode:       req.Vendor.Zip,
		Website:       req.Vendor.Website,
		Fax:           req.Vendor.Fax,
		Email:         req.Vendor.Email,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&v).Error; err != nil {
			return err
		}
		if req.Rep != nil {
			rep := vModel.RepContactVendor{
				VendorID:      v.IDVendor,
				Name:          req.Rep.Name,
				Title:         req.Rep.Title,
				Phone:         req.Rep.Phone,
				AddressLine2:  req.Rep.AddressLine2,
				City:          req.Rep.City,
				CountryID:     req.Rep.CountryID,
				Email:         req.Rep.Email,
				Fax:           req.Rep.Fax,
				StateID:       req.Rep.StateID,
				StreetAddress: req.Rep.StreetAddress,
				Zip:           req.Rep.Zip,
			}
			if err := tx.Create(&rep).Error; err != nil {
				return err
			}
		}
		_ = pkgActivity.Log(tx, "vendor", "create", pkgActivity.WithEntity(int64(v.IDVendor)),
			pkgActivity.WithDetails(map[string]interface{}{"name": v.VendorName}))
		return nil
	})
	if err != nil {
		return 0, err
	}
	return v.IDVendor, nil
}

type UpdateVendorRequest struct {
	Data map[string]interface{}
	Rep  *RepInput
}

func (s *Service) UpdateVendor(vendorID int, data map[string]interface{}, rep *RepInput) error {
	var v vModel.Vendor
	if err := s.db.First(&v, "id_vendor = ?", vendorID).Error; err != nil {
		return errors.New("Vendor not found")
	}

	fields := map[string]string{
		"vendor_name":    "vendor_name",
		"short_name":     "short_name",
		"phone":          "phone",
		"street_address": "street_address",
		"address_line_2": "address_line_2",
		"city":           "city",
		"state_id":       "state_id",
		"country_id":     "country_id",
		"website":        "website",
		"fax":            "fax",
		"email":          "email",
	}
	updates := map[string]interface{}{}
	for jsonKey, col := range fields {
		if val, ok := data[jsonKey]; ok {
			updates[col] = val
		}
	}
	// "zip" maps to "zip_code"
	if val, ok := data["zip"]; ok {
		updates["zip_code"] = val
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			if err := tx.Model(&v).Updates(updates).Error; err != nil {
				return err
			}
		}
		if rep != nil {
			var existing vModel.RepContactVendor
			err := tx.First(&existing, "vendor_id = ?", vendorID).Error
			if err == nil {
				repUpdates := buildRepUpdates(rep)
				if len(repUpdates) > 0 {
					if err := tx.Model(&existing).Updates(repUpdates).Error; err != nil {
						return err
					}
				}
			} else {
				newRep := vModel.RepContactVendor{
					VendorID:      vendorID,
					Name:          rep.Name,
					Title:         rep.Title,
					Phone:         rep.Phone,
					AddressLine2:  rep.AddressLine2,
					City:          rep.City,
					CountryID:     rep.CountryID,
					Email:         rep.Email,
					Fax:           rep.Fax,
					StateID:       rep.StateID,
					StreetAddress: rep.StreetAddress,
					Zip:           rep.Zip,
				}
				if err := tx.Create(&newRep).Error; err != nil {
					return err
				}
			}
		}
		_ = pkgActivity.Log(tx, "vendor", "update", pkgActivity.WithEntity(int64(vendorID)))
		return nil
	})
}

func buildRepUpdates(rep *RepInput) map[string]interface{} {
	u := map[string]interface{}{}
	if rep.Name != nil {
		u["name"] = *rep.Name
	}
	if rep.Title != nil {
		u["title"] = *rep.Title
	}
	if rep.Phone != nil {
		u["phone"] = *rep.Phone
	}
	if rep.AddressLine2 != nil {
		u["address_line_2"] = *rep.AddressLine2
	}
	if rep.City != nil {
		u["city"] = *rep.City
	}
	if rep.CountryID != nil {
		u["country_id"] = *rep.CountryID
	}
	if rep.Email != nil {
		u["email"] = *rep.Email
	}
	if rep.Fax != nil {
		u["fax"] = *rep.Fax
	}
	if rep.StateID != nil {
		u["state_id"] = *rep.StateID
	}
	if rep.StreetAddress != nil {
		u["street_address"] = *rep.StreetAddress
	}
	if rep.Zip != nil {
		u["zip"] = *rep.Zip
	}
	return u
}

func (s *Service) DeleteVendor(vendorID int) error {
	var v vModel.Vendor
	if err := s.db.First(&v, "id_vendor = ?", vendorID).Error; err != nil {
		return errors.New("Vendor not found")
	}

	var cnt int64
	s.db.Model(&vModel.Vendor{}).Raw(
		"SELECT COUNT(*) FROM invoice WHERE vendor_id = ?", vendorID,
	).Count(&cnt)
	if cnt > 0 {
		return &ConflictError{Msg: "Vendor cannot be deleted because there are associated invoices."}
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		_ = pkgActivity.Log(tx, "vendor", "delete", pkgActivity.WithEntity(int64(vendorID)))
		return tx.Delete(&v).Error
	})
}

func (s *Service) GetVendor(vendorID int, locationIDs ...int64) (map[string]interface{}, error) {
	var v vModel.Vendor
	if err := s.db.First(&v, "id_vendor = ?", vendorID).Error; err != nil {
		return nil, errors.New("Vendor not found")
	}

	// Frames brands
	var framesBrands []vModel.VendorBrand
	s.db.Preload("Brand").Where("id_vendor = ?", vendorID).Find(&framesBrands)
	framesList := make([]map[string]interface{}, 0, len(framesBrands))
	for _, fb := range framesBrands {
		if fb.Brand == nil {
			continue
		}
		b := fb.Brand
		item := map[string]interface{}{
			"brand_id":           b.IDBrand,
			"brand_name":         b.BrandName,
			"short_name":         b.ShortName,
			"description":        b.Description,
			"return_policy":      b.ReturnPolicy,
			"note":               b.Note,
			"print_model_on_tag": b.PrintModelOnTag,
			"print_price_on_tag": b.PrintPriceOnTag,
			"discount":           b.Discount,
			"can_lookup":         b.CanLookup,
		}
		// type_items_of_brand
		if b.TypeItemsOfBrandID != nil {
			var tib vModel.TypeItemsOfBrand
			if s.db.First(&tib, "id_type_items_of_brand = ?", *b.TypeItemsOfBrandID).Error == nil {
				item["type_items_of_brand"] = tib.ToMap()
			}
		}
		framesList = append(framesList, item)
	}

	// Lens brands
	var lensBrands []vModel.VendorBrandLens
	s.db.Preload("BrandLens").Where("id_vendor = ?", vendorID).Find(&lensBrands)
	lensList := make([]map[string]interface{}, 0, len(lensBrands))
	for _, lb := range lensBrands {
		if lb.BrandLens == nil {
			continue
		}
		b := lb.BrandLens
		lensList = append(lensList, map[string]interface{}{
			"brand_id":           b.IDBrandLens,
			"brand_name":         b.BrandName,
			"short_name":         b.ShortName,
			"description":        b.Description,
			"return_policy":      b.ReturnPolicy,
			"note":               b.Note,
			"print_model_on_tag": b.PrintModelOnTag,
			"print_price_on_tag": b.PrintPriceOnTag,
			"discount":           b.Discount,
			"can_lookup":         b.CanLookup,
		})
	}

	// Contact lens brands
	var clBrands []vModel.VendorBrandContactLens
	s.db.Preload("BrandContactLens").Where("id_vendor = ?", vendorID).Find(&clBrands)
	clList := make([]map[string]interface{}, 0, len(clBrands))
	for _, cl := range clBrands {
		if cl.BrandContactLens == nil {
			continue
		}
		b := cl.BrandContactLens
		clList = append(clList, map[string]interface{}{
			"brand_id":           b.IDBrandContactLens,
			"brand_name":         b.BrandName,
			"short_name":         b.ShortName,
			"description":        b.Description,
			"return_policy":      b.ReturnPolicy,
			"note":               b.Note,
			"print_model_on_tag": b.PrintModelOnTag,
			"print_price_on_tag": b.PrintPriceOnTag,
			"discount":           b.Discount,
			"can_lookup":         b.CanLookup,
		})
	}

	// Labs
	var vendorLabs []vModel.VendorLabs
	s.db.Preload("Lab").Where("vendor_id = ?", vendorID).Find(&vendorLabs)
	var labsList interface{}
	if len(vendorLabs) > 0 {
		labs := make([]map[string]interface{}, 0, len(vendorLabs))
		for _, vl := range vendorLabs {
			if vl.Lab != nil {
				labs = append(labs, vl.Lab.ToMap())
			}
		}
		labsList = labs
	}

	// Agreements
	var agreements []vModel.Agreement
	s.db.Where("vendor_id = ?", vendorID).Find(&agreements)
	agrList := make([]map[string]interface{}, 0, len(agreements))
	for _, a := range agreements {
		item := map[string]interface{}{
			"id_agreement": a.IDAgreement,
			"link_to_file": a.LinkToFile,
			"title":        a.Title,
		}
		if a.DateAgreement != nil {
			item["date_agreement"] = a.DateAgreement.Format("2006-01-02")
		} else {
			item["date_agreement"] = nil
		}
		if a.DateEnd != nil {
			item["date_end"] = a.DateEnd.Format("2006-01-02")
		} else {
			item["date_end"] = nil
		}
		agrList = append(agrList, item)
	}

	// Rep contact
	var repContact vModel.RepContactVendor
	var repData interface{}
	if s.db.First(&repContact, "vendor_id = ?", vendorID).Error == nil {
		rd := repContact.ToMap()
		delete(rd, "id_rep_contact_vendor")
		repData = rd
	}

	result := map[string]interface{}{
		"vendor_name":    v.VendorName,
		"vendor_id":      v.IDVendor,
		"short_name":     v.ShortName,
		"phone":          v.Phone,
		"street_address": v.StreetAddress,
		"address_line_2": v.AddressLine2,
		"city":           v.City,
		"state_id":       v.StateID,
		"zip":            v.ZipCode,
		"website":        v.Website,
		"fax":            v.Fax,
		"email":          v.Email,
		"country_id":     v.CountryID,
		"rep":            repData,
		"brands": map[string]interface{}{
			"frames":      framesList,
			"lens":        lensList,
			"contact_lens": clList,
		},
		"labs":       labsList,
		"agreements": agrList,
	}
	return result, nil
}

type VendorListResult struct {
	CurrentPage  int                      `json:"current_page"`
	TotalPages   int                      `json:"total_pages"`
	TotalVendors int64                    `json:"total_vendors"`
	Vendors      []map[string]interface{} `json:"vendors"`
}

func (s *Service) ListVendors(page int, includeDetails bool) (*VendorListResult, error) {
	const perPage = 20
	if page < 1 {
		page = 1
	}

	var total int64
	s.db.Model(&vModel.Vendor{}).Count(&total)

	var vendors []vModel.Vendor
	s.db.Order("vendor_name ASC").Offset((page - 1) * perPage).Limit(perPage).Find(&vendors)

	list := make([]map[string]interface{}, 0, len(vendors))
	for _, v := range vendors {
		if includeDetails {
			list = append(list, map[string]interface{}{
				"vendor_id":      v.IDVendor,
				"vendor_name":    v.VendorName,
				"short_name":     v.ShortName,
				"phone":          v.Phone,
				"street_address": v.StreetAddress,
				"address_line_2": v.AddressLine2,
				"city":           v.City,
				"state_id":       v.StateID,
				"zip":            v.ZipCode,
				"country_id":     v.CountryID,
				"website":        v.Website,
				"fax":            v.Fax,
				"email":          v.Email,
			})
		} else {
			list = append(list, map[string]interface{}{
				"vendor_id":   v.IDVendor,
				"vendor_name": v.VendorName,
			})
		}
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return &VendorListResult{
		CurrentPage:  page,
		TotalPages:   totalPages,
		TotalVendors: total,
		Vendors:      list,
	}, nil
}

func (s *Service) GetVendorInvoices(vendorID int) ([]map[string]interface{}, error) {
	rows, err := s.db.Raw(`
		SELECT DISTINCT
			i.id_invoice,
			i.number_invoice,
			i.date_create,
			i.total_amount,
			i.final_amount,
			i.employee_id,
			i.doctor_id,
			l.full_name AS location_name,
			p.first_name AS patient_first_name,
			p.last_name AS patient_last_name
		FROM invoice i
		JOIN inventory inv ON inv.invoice_id = i.id_invoice
		JOIN location l ON l.id_location = i.location_id
		JOIN patient p ON p.id_patient = i.patient_id
		WHERE inv.vendor_id = ?
	`, vendorID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var (
			idInvoice        int64
			numberInvoice    string
			dateCreate       time.Time
			totalAmount      float64
			finalAmount      float64
			employeeID       *int64
			doctorID         *int64
			locationName     *string
			patientFirstName *string
			patientLastName  *string
		)
		if err := rows.Scan(&idInvoice, &numberInvoice, &dateCreate, &totalAmount, &finalAmount,
			&employeeID, &doctorID, &locationName, &patientFirstName, &patientLastName); err != nil {
			return nil, err
		}
		result = append(result, map[string]interface{}{
			"id_invoice":         idInvoice,
			"number_invoice":     numberInvoice,
			"date_create":        dateCreate,
			"total_amount":       totalAmount,
			"final_amount":       finalAmount,
			"created_by":         employeeID,
			"last_modified_by":   doctorID,
			"location_name":      locationName,
			"patient_first_name": patientFirstName,
			"patient_last_name":  patientLastName,
		})
	}
	return result, nil
}

// =====================================================================
// Agreement CRUD
// =====================================================================

type AgreementInput struct {
	LinkToFile    *string `json:"link_to_file"`
	Title         *string `json:"title"`
	DateAgreement *string `json:"date_agreement"`
	DateEnd       *string `json:"date_end"`
}

func (s *Service) CreateAgreement(vendorID int, input AgreementInput) (map[string]interface{}, error) {
	var v vModel.Vendor
	if err := s.db.First(&v, "id_vendor = ?", vendorID).Error; err != nil {
		return nil, errors.New("Vendor not found")
	}
	if input.LinkToFile == nil || input.Title == nil {
		return nil, errors.New("link_to_file and title are required")
	}

	dateAgreement, err := parseOptionalDate(input.DateAgreement)
	if err != nil {
		return nil, fmt.Errorf("invalid date_agreement format, expected YYYY-MM-DD")
	}
	dateEnd, err := parseOptionalDate(input.DateEnd)
	if err != nil {
		return nil, fmt.Errorf("invalid date_end format, expected YYYY-MM-DD")
	}

	var cleanedPath *string
	if input.LinkToFile != nil {
		cp := cleanFilePath(*input.LinkToFile)
		cleanedPath = &cp
	}

	agr := vModel.Agreement{
		LinkToFile:    cleanedPath,
		Title:         input.Title,
		DateAgreement: dateAgreement,
		DateEnd:       dateEnd,
		VendorID:      vendorID,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&agr).Error; err != nil {
			return err
		}
		_ = pkgActivity.Log(tx, "vendor", "agreement_create", pkgActivity.WithEntity(int64(vendorID)))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return agr.ToMap(), nil
}

func (s *Service) UpdateAgreement(vendorID, agreementID int, input AgreementInput) (map[string]interface{}, error) {
	var v vModel.Vendor
	if err := s.db.First(&v, "id_vendor = ?", vendorID).Error; err != nil {
		return nil, errors.New("Vendor not found")
	}

	var agr vModel.Agreement
	if err := s.db.First(&agr, "id_agreement = ? AND vendor_id = ?", agreementID, vendorID).Error; err != nil {
		return nil, &NotFoundError{Msg: "Agreement not found for this vendor"}
	}

	if input.DateAgreement != nil {
		d, err := parseDate(*input.DateAgreement)
		if err != nil {
			return nil, fmt.Errorf("invalid date_agreement format (YYYY-MM-DD)")
		}
		agr.DateAgreement = &d
	}
	if input.DateEnd != nil {
		d, err := parseDate(*input.DateEnd)
		if err != nil {
			return nil, fmt.Errorf("invalid date_end format (YYYY-MM-DD)")
		}
		agr.DateEnd = &d
	}
	if input.LinkToFile != nil {
		cp := cleanFilePath(*input.LinkToFile)
		agr.LinkToFile = &cp
	}
	if input.Title != nil {
		agr.Title = input.Title
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&agr).Error; err != nil {
			return err
		}
		_ = pkgActivity.Log(tx, "vendor", "agreement_update", pkgActivity.WithEntity(int64(agreementID)))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return agr.ToMap(), nil
}

func (s *Service) DeleteAgreement(vendorID, agreementID int) error {
	var v vModel.Vendor
	if err := s.db.First(&v, "id_vendor = ?", vendorID).Error; err != nil {
		return errors.New("Vendor not found")
	}

	var agr vModel.Agreement
	if err := s.db.First(&agr, "id_agreement = ? AND vendor_id = ?", agreementID, vendorID).Error; err != nil {
		return &NotFoundError{Msg: "Agreement not found for this vendor"}
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		_ = pkgActivity.Log(tx, "vendor", "agreement_delete", pkgActivity.WithEntity(int64(agreementID)))
		return tx.Delete(&agr).Error
	})
}

// =====================================================================
// Brand management
// =====================================================================

type AddBrandInput struct {
	KeyBrand        string  `json:"key_brand"`
	BrandName       string  `json:"brand_name"`
	ShortName       *string `json:"short_name"`
	Description     *string `json:"description"`
	ReturnPolicy    *string `json:"return_policy"`
	Note            *string `json:"note"`
	PrintModelOnTag *bool   `json:"print_model_on_tag"`
	PrintPriceOnTag *bool   `json:"print_price_on_tag"`
	Discount        *int    `json:"discount"`
	CanLookup       *bool   `json:"can_lookup"`
}

func (s *Service) AddVendorBrand(vendorID int, input AddBrandInput) (map[string]interface{}, error) {
	if input.BrandName == "" || input.KeyBrand == "" {
		return nil, errors.New("Brand name and key_brand are required")
	}

	var vendorExists int64
	s.db.Model(&vModel.Vendor{}).Where("id_vendor = ?", vendorID).Count(&vendorExists)
	if vendorExists == 0 {
		return nil, &NotFoundError{Msg: "Vendor not found"}
	}

	key := strings.ToLower(input.KeyBrand)

	var typeItem vModel.TypeItemsOfBrand
	if err := s.db.First(&typeItem, "type_name = ?", key).Error; err != nil {
		return nil, fmt.Errorf("No matching type_items_of_brand for key_brand: %s", input.KeyBrand)
	}
	tibID := typeItem.IDTypeItemsOfBrand

	printModel := true
	if input.PrintModelOnTag != nil {
		printModel = *input.PrintModelOnTag
	}
	printPrice := true
	if input.PrintPriceOnTag != nil {
		printPrice = *input.PrintPriceOnTag
	}
	discount := 0
	if input.Discount != nil {
		discount = *input.Discount
	}
	canLookup := true
	if input.CanLookup != nil {
		canLookup = *input.CanLookup
	}

	var resultMsg string

	err := s.db.Transaction(func(tx *gorm.DB) error {
		switch key {
		case "frames":
			brand := vModel.Brand{
				BrandName:          &input.BrandName,
				ShortName:          input.ShortName,
				ReturnPolicy:       input.ReturnPolicy,
				Note:               input.Note,
				PrintModelOnTag:    printModel,
				PrintPriceOnTag:    printPrice,
				Discount:           &discount,
				CanLookup:          canLookup,
				TypeItemsOfBrandID: &tibID,
			}
			if err := tx.Create(&brand).Error; err != nil {
				return err
			}
			link := vModel.VendorBrand{IDVendor: vendorID, IDBrand: brand.IDBrand}
			if err := tx.Create(&link).Error; err != nil {
				return err
			}

		case "lens":
			brand := vModel.BrandLens{
				BrandName:          input.BrandName,
				ShortName:          input.ShortName,
				Description:        input.Description,
				ReturnPolicy:       input.ReturnPolicy,
				Note:               input.Note,
				PrintModelOnTag:    printModel,
				PrintPriceOnTag:    printPrice,
				Discount:           &discount,
				CanLookup:          canLookup,
				TypeItemsOfBrandID: &tibID,
			}
			if err := tx.Create(&brand).Error; err != nil {
				return err
			}
			link := vModel.VendorBrandLens{IDVendor: vendorID, IDBrandLens: brand.IDBrandLens}
			if err := tx.Create(&link).Error; err != nil {
				return err
			}

		case "contact_lens":
			brand := vModel.BrandContactLens{
				BrandName:          input.BrandName,
				ShortName:          input.ShortName,
				Description:        input.Description,
				ReturnPolicy:       input.ReturnPolicy,
				Note:               input.Note,
				PrintModelOnTag:    printModel,
				PrintPriceOnTag:    printPrice,
				Discount:           discount,
				CanLookup:          canLookup,
				TypeItemsOfBrandID: &tibID,
			}
			if err := tx.Create(&brand).Error; err != nil {
				return err
			}
			link := vModel.VendorBrandContactLens{IDVendor: vendorID, IDBrandContactLens: brand.IDBrandContactLens}
			if err := tx.Create(&link).Error; err != nil {
				return err
			}

		default:
			return errors.New("Invalid key_brand. Must be one of: frames, lens, contact_lens")
		}

		_ = pkgActivity.Log(tx, "vendor", "brand_add", pkgActivity.WithEntity(int64(vendorID)))
		return nil
	})
	if err != nil {
		return nil, err
	}

	resultMsg = fmt.Sprintf("%s brand added successfully", input.KeyBrand)
	return map[string]interface{}{
		"message":                resultMsg,
		"brand_name":             input.BrandName,
		"type_items_of_brand_id": tibID,
	}, nil
}

type UpdateBrandInput struct {
	KeyBrand        string  `json:"key_brand"`
	BrandName       *string `json:"brand_name"`
	ShortName       *string `json:"short_name"`
	Description     *string `json:"description"`
	ReturnPolicy    *string `json:"return_policy"`
	Note            *string `json:"note"`
	PrintModelOnTag *bool   `json:"print_model_on_tag"`
	PrintPriceOnTag *bool   `json:"print_price_on_tag"`
	Discount        *int    `json:"discount"`
	CanLookup       *bool   `json:"can_lookup"`
}

func (s *Service) UpdateVendorBrand(brandID int, input UpdateBrandInput) error {
	if input.KeyBrand == "" {
		return errors.New("key_brand is required")
	}

	key := strings.ToLower(input.KeyBrand)
	updates := map[string]interface{}{}
	if input.BrandName != nil {
		updates["brand_name"] = *input.BrandName
	}
	if input.ShortName != nil {
		updates["short_name"] = *input.ShortName
	}
	if input.ReturnPolicy != nil {
		updates["return_policy"] = *input.ReturnPolicy
	}
	if input.Note != nil {
		updates["note"] = *input.Note
	}
	if input.PrintModelOnTag != nil {
		updates["print_model_on_tag"] = *input.PrintModelOnTag
	}
	if input.PrintPriceOnTag != nil {
		updates["print_price_on_tag"] = *input.PrintPriceOnTag
	}
	if input.Discount != nil {
		updates["discount"] = *input.Discount
	}
	if input.CanLookup != nil {
		updates["can_lookup"] = *input.CanLookup
	}

	switch key {
	case "frames":
		if input.Description != nil {
			updates["description"] = *input.Description
		}
		var brand vModel.Brand
		if err := s.db.First(&brand, "id_brand = ?", brandID).Error; err != nil {
			return &NotFoundError{Msg: fmt.Sprintf("Brand not found for key_brand: %s and id: %d", key, brandID)}
		}
		return s.db.Model(&brand).Updates(updates).Error

	case "lens":
		if input.Description != nil {
			updates["description"] = *input.Description
		}
		var brand vModel.BrandLens
		if err := s.db.First(&brand, "id_brand_lens = ?", brandID).Error; err != nil {
			return &NotFoundError{Msg: fmt.Sprintf("Brand not found for key_brand: %s and id: %d", key, brandID)}
		}
		return s.db.Model(&brand).Updates(updates).Error

	case "contact_lens":
		if input.Description != nil {
			updates["description"] = *input.Description
		}
		var brand vModel.BrandContactLens
		if err := s.db.First(&brand, "id_brand_contact_lens = ?", brandID).Error; err != nil {
			return &NotFoundError{Msg: fmt.Sprintf("Brand not found for key_brand: %s and id: %d", key, brandID)}
		}
		return s.db.Model(&brand).Updates(updates).Error

	default:
		return errors.New("Invalid key_brand. Must be one of: frames, lens, contact_lens")
	}
}

func (s *Service) DeleteVendorBrand(vendorID int, brandType string, brandID int) error {
	key := strings.ToLower(brandType)
	var isUsed bool

	switch key {
	case "frames":
		var cnt int64
		s.db.Model(&frameModel.Product{}).Where("brand_id = ?", brandID).Count(&cnt)
		if cnt == 0 {
			s.db.Model(&otherModel.CrossSellProduct{}).Where("brand_id = ?", brandID).Count(&cnt)
		}
		isUsed = cnt > 0

		if isUsed {
			return &ConflictError{Msg: "Cannot delete frames brand: it is used in the system"}
		}
		var link vModel.VendorBrand
		if err := s.db.First(&link, "id_vendor = ? AND id_brand = ?", vendorID, brandID).Error; err != nil {
			return &NotFoundError{Msg: "No brand link found for this vendor"}
		}
		return s.db.Transaction(func(tx *gorm.DB) error {
			_ = pkgActivity.Log(tx, "vendor", "brand_delete", pkgActivity.WithEntity(int64(vendorID)))
			return tx.Delete(&link).Error
		})

	case "lens":
		var cnt int64
		s.db.Model(&lensModel.Lenses{}).Where("brand_lens_id = ?", brandID).Count(&cnt)
		isUsed = cnt > 0
		if isUsed {
			return &ConflictError{Msg: "Cannot delete lens brand: it is used in the system"}
		}
		var link vModel.VendorBrandLens
		if err := s.db.First(&link, "id_vendor = ? AND id_brand_lens = ?", vendorID, brandID).Error; err != nil {
			return &NotFoundError{Msg: "No brand link found for this vendor"}
		}
		return s.db.Transaction(func(tx *gorm.DB) error {
			_ = pkgActivity.Log(tx, "vendor", "brand_delete", pkgActivity.WithEntity(int64(vendorID)))
			return tx.Delete(&link).Error
		})

	case "contact_lens":
		var cnt int64
		s.db.Model(&clModel.ContactLensItem{}).Where("brand_contact_lens_id = ?", brandID).Count(&cnt)
		isUsed = cnt > 0
		if isUsed {
			return &ConflictError{Msg: "Cannot delete contact_lens brand: it is used in the system"}
		}
		var link vModel.VendorBrandContactLens
		if err := s.db.First(&link, "id_vendor = ? AND id_brand_contact_lens = ?", vendorID, brandID).Error; err != nil {
			return &NotFoundError{Msg: "No brand link found for this vendor"}
		}
		return s.db.Transaction(func(tx *gorm.DB) error {
			_ = pkgActivity.Log(tx, "vendor", "brand_delete", pkgActivity.WithEntity(int64(vendorID)))
			return tx.Delete(&link).Error
		})

	default:
		return errors.New("Invalid brand type. Must be 'frames', 'lens', or 'contact_lens'")
	}
}

// =====================================================================
// Lab CRUD
// =====================================================================

type LabInput struct {
	TitleLab      string  `json:"title_lab"`
	ShortName     *string `json:"short_name"`
	IsInternal    *bool   `json:"is_internal"`
	Phone         *string `json:"phone"`
	Email         *string `json:"email"`
	StreetAddress *string `json:"street_address"`
	AddressLine2  *string `json:"address_line_2"`
	City          *string `json:"city"`
	ZipCode       *string `json:"zip_code"`
	StateID       *int    `json:"state_id"`
	CountryID     *int    `json:"country_id"`
	VendorID      *int    `json:"vendor_id"`
	BrandLensID   *int    `json:"brand_lens_id"`
	Source        *string `json:"source"`
	// VW account fields (saved to vendor_location_account)
	VwSloID       *int    `json:"vw_slo_id"`
	VwBill        *string `json:"vw_bill"`
	VwShip        *string `json:"vw_ship"`
	AccountNumber *string `json:"account_number"`
	LocationID    int64   `json:"-"` // set by handler from JWT
}

func (s *Service) ListLabs(locationID *int64) ([]map[string]interface{}, error) {
	var labs []vModel.Lab
	s.db.Order("title_lab ASC").Find(&labs)
	result := make([]map[string]interface{}, 0, len(labs))
	for _, l := range labs {
		entry := map[string]interface{}{
			"lab_id":    l.IDLab,
			"title_lab": l.TitleLab,
		}
		// If location provided, attach VW account info from vendor_location_account
		if locationID != nil && l.VendorID != nil {
			type vlaRow struct {
				AccountNumber *string `gorm:"column:account_number"`
				VwSloID       *int    `gorm:"column:vw_slo_id"`
				VwBill        *string `gorm:"column:vw_bill"`
				VwShip        *string `gorm:"column:vw_ship"`
				Source        *string `gorm:"column:source"`
			}
			var vla vlaRow
			s.db.Table("vendor_location_account").
				Where("vendor_id = ? AND location_id = ? AND is_active = true", *l.VendorID, *locationID).
				First(&vla)
			entry["account_number"] = vla.AccountNumber
			entry["vw_slo_id"] = vla.VwSloID
			entry["vw_bill"] = vla.VwBill
			entry["vw_ship"] = vla.VwShip
			entry["source"] = vla.Source
		}
		result = append(result, entry)
	}
	return result, nil
}

func (s *Service) CreateLab(input LabInput) (map[string]interface{}, error) {
	if input.TitleLab == "" {
		return nil, errors.New("title_lab is required")
	}

	isInternal := false
	if input.IsInternal != nil {
		isInternal = *input.IsInternal
	}

	lab := vModel.Lab{
		TitleLab:      input.TitleLab,
		ShortName:     input.ShortName,
		IsInternal:    isInternal,
		Phone:         input.Phone,
		Email:         input.Email,
		StreetAddress: input.StreetAddress,
		AddressLine2:  input.AddressLine2,
		City:          input.City,
		ZipCode:       input.ZipCode,
		StateID:       input.StateID,
		CountryID:     input.CountryID,
		VendorID:      input.VendorID,
		BrandLensID:   input.BrandLensID,
		Source:        input.Source,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&lab).Error; err != nil {
			return err
		}
		// Auto-link lab to vendor
		if input.VendorID != nil {
			tx.Exec("INSERT INTO vendor_labs (vendor_id, lab_id) VALUES (?, ?) ON CONFLICT DO NOTHING", *input.VendorID, lab.IDLab)
		}
		// Create vendor_location_account if VW fields provided
		if input.VendorID != nil && input.LocationID > 0 && (input.VwSloID != nil || input.VwBill != nil || input.AccountNumber != nil) {
			src := "custom"
			if input.Source != nil {
				src = *input.Source
			}
			tx.Exec(`INSERT INTO vendor_location_account (vendor_id, location_id, account_number, vw_slo_id, vw_bill, vw_ship, source, is_active)
				VALUES (?, ?, ?, ?, ?, ?, ?, true)
				ON CONFLICT DO NOTHING`,
				*input.VendorID, input.LocationID, input.AccountNumber, input.VwSloID, input.VwBill, input.VwShip, src)
		}
		_ = pkgActivity.Log(tx, "vendor", "lab_create", pkgActivity.WithEntity(int64(lab.IDLab)))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return lab.ToMap(), nil
}

func (s *Service) GetLab(labID int, locationID *int64) (map[string]interface{}, error) {
	var lab vModel.Lab
	if err := s.db.First(&lab, "id_lab = ?", labID).Error; err != nil {
		return nil, &NotFoundError{Msg: "Lab not found"}
	}
	result := lab.ToMap()
	if locationID != nil && lab.VendorID != nil {
		type vlaRow struct {
			AccountNumber *string `gorm:"column:account_number"`
			VwSloID       *int    `gorm:"column:vw_slo_id"`
			VwBill        *string `gorm:"column:vw_bill"`
			VwShip        *string `gorm:"column:vw_ship"`
			Source        *string `gorm:"column:source"`
		}
		var vla vlaRow
		s.db.Table("vendor_location_account").
			Where("vendor_id = ? AND location_id = ? AND is_active = true", *lab.VendorID, *locationID).
			First(&vla)
		result["account_number"] = vla.AccountNumber
		result["vw_slo_id"] = vla.VwSloID
		result["vw_bill"] = vla.VwBill
		result["vw_ship"] = vla.VwShip
		result["source"] = vla.Source
	}
	return result, nil
}

func (s *Service) UpdateLab(labID int, data map[string]interface{}) (map[string]interface{}, error) {
	var lab vModel.Lab
	if err := s.db.First(&lab, "id_lab = ?", labID).Error; err != nil {
		return nil, &NotFoundError{Msg: "Lab not found"}
	}

	updates := map[string]interface{}{}
	fieldMap := map[string]string{
		"title_lab":      "title_lab",
		"short_name":     "short_name",
		"is_internal":    "is_internal",
		"phone":          "phone",
		"email":          "email",
		"street_address": "street_address",
		"address_line_2": "address_line_2",
		"city":           "city",
		"zip_code":       "zip_code",
		"state_id":       "state_id",
		"country_id":     "country_id",
	}
	for jsonKey, col := range fieldMap {
		if val, ok := data[jsonKey]; ok {
			updates[col] = val
		}
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			if err := tx.Model(&lab).Updates(updates).Error; err != nil {
				return err
			}
		}
		_ = pkgActivity.Log(tx, "vendor", "lab_update", pkgActivity.WithEntity(int64(labID)))
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Reload
	s.db.First(&lab, "id_lab = ?", labID)
	return lab.ToMap(), nil
}

func (s *Service) DeleteLab(labID int) error {
	var lab vModel.Lab
	if err := s.db.First(&lab, "id_lab = ?", labID).Error; err != nil {
		return &NotFoundError{Msg: "Lab not found"}
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		_ = pkgActivity.Log(tx, "vendor", "lab_delete", pkgActivity.WithEntity(int64(labID)))
		return tx.Delete(&lab).Error
	})
}

func (s *Service) AddVendorLab(vendorID, labID int) error {
	var v vModel.Vendor
	if err := s.db.First(&v, "id_vendor = ?", vendorID).Error; err != nil {
		return &NotFoundError{Msg: "Vendor not found"}
	}
	var lab vModel.Lab
	if err := s.db.First(&lab, "id_lab = ?", labID).Error; err != nil {
		return &NotFoundError{Msg: "Lab not found"}
	}

	var existing vModel.VendorLabs
	if s.db.First(&existing, "vendor_id = ? AND lab_id = ?", vendorID, labID).Error == nil {
		return &AlreadyExistsError{Msg: "Lab already associated with this vendor"}
	}

	vl := vModel.VendorLabs{VendorID: vendorID, LabID: labID}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&vl).Error; err != nil {
			return err
		}
		_ = pkgActivity.Log(tx, "vendor", "lab_link", pkgActivity.WithEntity(int64(vendorID)),
			pkgActivity.WithDetails(map[string]interface{}{"lab_id": labID}))
		return nil
	})
}

func (s *Service) RemoveVendorLab(vendorID, labID int) error {
	var v vModel.Vendor
	if err := s.db.First(&v, "id_vendor = ?", vendorID).Error; err != nil {
		return &NotFoundError{Msg: "Vendor not found"}
	}
	var lab vModel.Lab
	if err := s.db.First(&lab, "id_lab = ?", labID).Error; err != nil {
		return &NotFoundError{Msg: "Lab not found"}
	}

	var vl vModel.VendorLabs
	if err := s.db.First(&vl, "vendor_id = ? AND lab_id = ?", vendorID, labID).Error; err != nil {
		return &NotFoundError{Msg: "Lab not associated with this vendor"}
	}

	// Check if lab can be fully deleted
	var ticketCount int64
	s.db.Table("lab_ticket").Where("lab_id = ?", labID).Count(&ticketCount)
	var otherVendorCount int64
	s.db.Table("vendor_labs").Where("lab_id = ? AND vendor_id != ?", labID, vendorID).Count(&otherVendorCount)

	canDelete := ticketCount == 0 && otherVendorCount == 0

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Always unlink from this vendor
		tx.Delete(&vl)
		if canDelete {
			// Full delete: lab + vendor_location_account
			tx.Exec("DELETE FROM vendor_location_account WHERE vendor_id = ?", vendorID)
			tx.Delete(&lab)
		}
		_ = pkgActivity.Log(tx, "vendor", "lab_delete", pkgActivity.WithEntity(int64(vendorID)),
			pkgActivity.WithDetails(map[string]interface{}{"lab_id": labID, "full_delete": canDelete}))
		return nil
	})
}

// =====================================================================
// Countries / States
// =====================================================================

func (s *Service) GetCountries() ([]map[string]interface{}, error) {
	var countries []genModel.Country
	s.db.Order("country ASC").Find(&countries)
	result := make([]map[string]interface{}, 0, len(countries))
	for _, c := range countries {
		result = append(result, map[string]interface{}{
			"id_country": c.IDCountry,
			"country":    c.Country,
		})
	}
	return result, nil
}

func (s *Service) GetStatesByCountry(countryID int) ([]map[string]interface{}, error) {
	var states []genModel.SalesTaxByState
	s.db.Where("country_id = ?", countryID).Order("state_name ASC").Find(&states)
	if len(states) == 0 {
		return nil, &NotFoundError{Msg: "No states found for the given country ID"}
	}
	result := make([]map[string]interface{}, 0, len(states))
	for _, st := range states {
		result = append(result, map[string]interface{}{
			"state_id":   st.IDSalesTax,
			"state_name": st.StateName,
		})
	}
	return result, nil
}

// =====================================================================
// Pricing Rules
// =====================================================================

type PricingRuleInput struct {
	SellingPrice    *float64 `json:"selling_price"`
	MinSellingPrice *float64 `json:"min_selling_price"`
	MinPrice        *float64 `json:"min_price"`
	MaxPrice        *float64 `json:"max_price"`
	Multiplier      *float64 `json:"multiplier"`
	LowerMultiplier *float64 `json:"lower_multiplier"`
	RoundingTargets *[]int   `json:"rounding_targets"`
}

func (s *Service) AddPricingRule(vendorID int, brandType string, brandID int, input PricingRuleInput) (map[string]interface{}, error) {
	if brandType != "frames" && brandType != "lens" && brandType != "contact_lens" {
		return nil, errors.New("Invalid brand_type")
	}

	isBase := input.SellingPrice != nil

	if isBase {
		if input.MinSellingPrice == nil {
			return nil, errors.New("min_selling_price is required for base rule")
		}
		if *input.SellingPrice <= 0 {
			return nil, errors.New("selling_price must be positive")
		}
		if *input.MinSellingPrice <= 0 {
			return nil, errors.New("min_selling_price must be positive")
		}
		input.MinPrice = nil
		input.MaxPrice = nil
		input.Multiplier = nil
		input.LowerMultiplier = nil
	} else {
		if input.Multiplier == nil {
			return nil, errors.New("multiplier is required for range rule")
		}
		if *input.Multiplier <= 0 {
			return nil, errors.New("multiplier must be positive")
		}
		if input.LowerMultiplier != nil && *input.LowerMultiplier <= 0 {
			return nil, errors.New("lower_multiplier must be positive")
		}
		if input.MinPrice != nil && input.MaxPrice != nil && *input.MinPrice >= *input.MaxPrice {
			return nil, errors.New("min_price must be less than max_price")
		}
		input.SellingPrice = nil
		input.MinSellingPrice = nil
	}

	if input.RoundingTargets != nil {
		for _, t := range *input.RoundingTargets {
			if t < 0 || t > 9 {
				return nil, errors.New("rounding_targets must be a list of integers 0-9")
			}
		}
	}

	// Check brand exists
	var brandExists int64
	switch brandType {
	case "frames":
		s.db.Model(&vModel.Brand{}).Where("id_brand = ?", brandID).Count(&brandExists)
	case "lens":
		s.db.Model(&vModel.BrandLens{}).Where("id_brand_lens = ?", brandID).Count(&brandExists)
	case "contact_lens":
		s.db.Model(&vModel.BrandContactLens{}).Where("id_brand_contact_lens = ?", brandID).Count(&brandExists)
	}
	if brandExists == 0 {
		return nil, &NotFoundError{Msg: "Brand not found"}
	}

	rule := vModel.PricingRule{
		BrandType: brandType,
		BrandID:   brandID,
	}
	if input.SellingPrice != nil {
		v := fmt.Sprintf("%.2f", *input.SellingPrice)
		rule.SellingPrice = &v
	}
	if input.MinSellingPrice != nil {
		v := fmt.Sprintf("%.2f", *input.MinSellingPrice)
		rule.MinSellingPrice = &v
	}
	if input.MinPrice != nil {
		v := fmt.Sprintf("%.2f", *input.MinPrice)
		rule.MinPrice = &v
	}
	if input.MaxPrice != nil {
		v := fmt.Sprintf("%.2f", *input.MaxPrice)
		rule.MaxPrice = &v
	}
	if input.Multiplier != nil {
		v := fmt.Sprintf("%.2f", *input.Multiplier)
		rule.Multiplier = &v
	}
	if input.LowerMultiplier != nil {
		v := fmt.Sprintf("%.2f", *input.LowerMultiplier)
		rule.LowerMultiplier = &v
	}
	if input.RoundingTargets != nil {
		rt, _ := json.Marshal(*input.RoundingTargets)
		rule.RoundingTargets = rt
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&rule).Error; err != nil {
			return err
		}
		_ = pkgActivity.Log(tx, "vendor", "pricing_rule_create", pkgActivity.WithEntity(int64(vendorID)))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rule.ToMap(), nil
}

func (s *Service) GetPricingRules(brandType string, brandID int) (map[string]interface{}, error) {
	if brandType != "frames" && brandType != "lens" && brandType != "contact_lens" {
		return nil, errors.New("Invalid brand_type")
	}

	var rules []vModel.PricingRule
	s.db.Where("brand_type = ? AND brand_id = ?", brandType, brandID).
		Order("min_price ASC NULLS FIRST").Find(&rules)

	rulesList := make([]map[string]interface{}, 0, len(rules))
	for _, r := range rules {
		rulesList = append(rulesList, r.ToMap())
	}

	return map[string]interface{}{
		"brand_type": brandType,
		"brand_id":   brandID,
		"rules":      rulesList,
	}, nil
}

func (s *Service) UpdatePricingRule(vendorID, ruleID int, data map[string]interface{}) (map[string]interface{}, error) {
	var rule vModel.PricingRule
	if err := s.db.First(&rule, "id_pricing_rule = ?", ruleID).Error; err != nil {
		return nil, &NotFoundError{Msg: "Pricing rule not found"}
	}

	if rule.IsBase() {
		if v, ok := data["selling_price"]; ok {
			if v != nil {
				f := toFloat(v)
				if f <= 0 {
					return nil, errors.New("selling_price must be positive")
				}
				s := fmt.Sprintf("%.2f", f)
				rule.SellingPrice = &s
			} else {
				rule.SellingPrice = nil
			}
		}
		if v, ok := data["min_selling_price"]; ok {
			if v != nil {
				f := toFloat(v)
				if f <= 0 {
					return nil, errors.New("min_selling_price must be positive")
				}
				s := fmt.Sprintf("%.2f", f)
				rule.MinSellingPrice = &s
			} else {
				rule.MinSellingPrice = nil
			}
		}
	} else {
		if v, ok := data["min_price"]; ok {
			if v != nil {
				s := fmt.Sprintf("%.2f", toFloat(v))
				rule.MinPrice = &s
			} else {
				rule.MinPrice = nil
			}
		}
		if v, ok := data["max_price"]; ok {
			if v != nil {
				s := fmt.Sprintf("%.2f", toFloat(v))
				rule.MaxPrice = &s
			} else {
				rule.MaxPrice = nil
			}
		}
		if v, ok := data["multiplier"]; ok {
			f := toFloat(v)
			if f <= 0 {
				return nil, errors.New("multiplier must be positive")
			}
			s := fmt.Sprintf("%.2f", f)
			rule.Multiplier = &s
		}
		if v, ok := data["lower_multiplier"]; ok {
			if v != nil {
				f := toFloat(v)
				if f <= 0 {
					return nil, errors.New("lower_multiplier must be positive")
				}
				s := fmt.Sprintf("%.2f", f)
				rule.LowerMultiplier = &s
			} else {
				rule.LowerMultiplier = nil
			}
		}
		// Cross-check min < max
		if rule.MinPrice != nil && rule.MaxPrice != nil {
			if toFloat(*rule.MinPrice) >= toFloat(*rule.MaxPrice) {
				return nil, errors.New("min_price must be less than max_price")
			}
		}
	}

	if v, ok := data["rounding_targets"]; ok {
		if v != nil {
			arr, ok := v.([]interface{})
			if !ok {
				return nil, errors.New("rounding_targets must be a list of integers 0-9")
			}
			targets := make([]int, 0, len(arr))
			for _, item := range arr {
				t := int(toFloat(item))
				if t < 0 || t > 9 {
					return nil, errors.New("rounding_targets must be a list of integers 0-9")
				}
				targets = append(targets, t)
			}
			rt, _ := json.Marshal(targets)
			rule.RoundingTargets = rt
		} else {
			rule.RoundingTargets = nil
		}
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&rule).Error; err != nil {
			return err
		}
		_ = pkgActivity.Log(tx, "vendor", "pricing_rule_update", pkgActivity.WithEntity(int64(ruleID)))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rule.ToMap(), nil
}

func (s *Service) DeletePricingRule(vendorID, ruleID int) error {
	var rule vModel.PricingRule
	if err := s.db.First(&rule, "id_pricing_rule = ?", ruleID).Error; err != nil {
		return &NotFoundError{Msg: "Pricing rule not found"}
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		_ = pkgActivity.Log(tx, "vendor", "pricing_rule_delete", pkgActivity.WithEntity(int64(ruleID)))
		return tx.Delete(&rule).Error
	})
}

// =====================================================================
// Helpers
// =====================================================================

func parseOptionalDate(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func cleanFilePath(raw string) string {
	parts := strings.SplitAfter(raw, "/mnt/tank/data/")
	if len(parts) > 1 {
		return strings.TrimSpace(parts[len(parts)-1])
	}
	return strings.TrimSpace(raw)
}

func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		f := 0.0
		fmt.Sscanf(val, "%f", &f)
		return f
	default:
		return 0
	}
}

// =====================================================================
// Error types
// =====================================================================

type NotFoundError struct{ Msg string }

func (e *NotFoundError) Error() string { return e.Msg }

type ConflictError struct{ Msg string }

func (e *ConflictError) Error() string { return e.Msg }

type AlreadyExistsError struct{ Msg string }

func (e *AlreadyExistsError) Error() string { return e.Msg }
