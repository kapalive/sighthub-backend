package store_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	empModel  "sighthub-backend/internal/models/employees"
	genModel  "sighthub-backend/internal/models/general"
	locModel  "sighthub-backend/internal/models/location"
	permModel "sighthub-backend/internal/models/permission"
	pkgActivity "sighthub-backend/pkg/activitylog"
	pkgEmail    "sighthub-backend/pkg/email"
)

var storePermissions = []int{
	1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
	21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35,
	41, 42, 43, 44, 45, 46, 47,
	62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79,
}

var warehousePermissions = []int{
	11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 32,
	41, 42, 43, 44, 45, 46, 47,
	62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79,
}

// Service —————————————————————————————————————————————————————————

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service { return &Service{db: db} }

// --- DTOs ---

type StoreListItem struct {
	IDStore      int     `json:"id_store"`
	BusinessName *string `json:"business_name"`
	FullName     *string `json:"full_name"`
	ShortName    *string `json:"short_name"`
	Phone        *string `json:"phone"`
	Address      string  `json:"address"`
	Active       *bool   `json:"active"`
}

type StoreDetailResponse struct {
	Details  map[string]interface{} `json:"details"`
	Address  map[string]interface{} `json:"address"`
	Schedule interface{}            `json:"schedule"`
}

type StoreInput struct {
	Details  map[string]interface{} `json:"details"`
	Address  map[string]interface{} `json:"address"`
	Schedule map[string]interface{} `json:"schedule"`
}

type WarehouseListItem struct {
	ShortName     *string `json:"short_name"`
	WarehouseName *string `json:"warehouse_name"`
	ConnectStore  *string `json:"connect_store"`
	StoreID       *int    `json:"store_id"`
	WarehouseID   int     `json:"warehouse_id"`
	Address       string  `json:"address"`
	Phone         *string `json:"phone"`
	Active        *bool   `json:"active"`
}

type WarehouseDetailResponse struct {
	ShortName       *string `json:"short_name"`
	WarehouseName   *string `json:"warehouse_name"`
	ConnectStore    *string `json:"connect_store"`
	StoreID         *int    `json:"store_id"`
	StreetAddress   *string `json:"street_address"`
	AddressLine2    *string `json:"address_line_2"`
	City            *string `json:"city"`
	State           *string `json:"state"`
	PostalCode      *string `json:"postal_code"`
	Phone           *string `json:"phone"`
	Email           *string `json:"email"`
	CanReceiveItems *bool   `json:"can_receive_items"`
	Active          *bool   `json:"active"`
}

type SalesTaxListItem struct {
	IDSalesTax      int    `json:"id_sales_tax"`
	StateCode       string `json:"state_code"`
	SalesTaxPercent string `json:"sales_tax_percent"`
}

type SalesTaxItem struct {
	Code    string `json:"code"`
	State   string `json:"state"`
	SaleTax string `json:"sale_tax"`
}

// --- Stores ---

func (s *Service) GetAllStores() ([]StoreListItem, error) {
	var stores []locModel.Store
	if err := s.db.Find(&stores).Error; err != nil {
		return nil, err
	}
	result := make([]StoreListItem, 0, len(stores))
	for _, store := range stores {
		var loc locModel.Location
		s.db.Where("store_id = ? AND warehouse_id IS NULL", store.IDStore).First(&loc)

		parts := make([]string, 0, 5)
		if store.StreetAddress != nil { parts = append(parts, *store.StreetAddress) }
		if store.AddressLine2 != nil  { parts = append(parts, *store.AddressLine2) }
		if store.City != nil          { parts = append(parts, *store.City) }
		if store.State != nil         { parts = append(parts, *store.State) }
		if store.Country != nil       { parts = append(parts, *store.Country) }

		result = append(result, StoreListItem{
			IDStore:      store.IDStore,
			BusinessName: store.BusinessName,
			FullName:     store.FullName,
			ShortName:    store.ShortName,
			Phone:        store.Phone,
			Address:      strings.Join(parts, " "),
			Active:       loc.StoreActive,
		})
	}
	return result, nil
}

func (s *Service) GetStore(storeID int) (*StoreDetailResponse, error) {
	var store locModel.Store
	if err := s.db.First(&store, storeID).Error; err != nil {
		return nil, fmt.Errorf("store not found")
	}
	var loc locModel.Location
	if err := s.db.Where("store_id = ? AND warehouse_id IS NULL", storeID).First(&loc).Error; err != nil {
		return nil, fmt.Errorf("location not found")
	}

	var schedMap interface{}
	if loc.WorkShiftID != nil {
		var ws empModel.WorkShift
		if err := s.db.First(&ws, *loc.WorkShiftID).Error; err == nil {
			schedMap = ws.ToMap()
		}
	}

	return s.buildStoreDetailResponse(&store, &loc, schedMap), nil
}

func (s *Service) GetRequestAppointmentLink(storeID int) (string, error) {
	var store locModel.Store
	if err := s.db.First(&store, storeID).Error; err != nil {
		return "", fmt.Errorf("store not found")
	}
	var loc locModel.Location
	if err := s.db.Where("store_id = ? AND warehouse_id IS NULL", storeID).First(&loc).Error; err != nil {
		return "", fmt.Errorf("location not found")
	}
	var setting locModel.LocationAppointmentSettings
	if err := s.db.Where("location_id = ?", loc.IDLocation).First(&setting).Error; err != nil || !setting.RequestAppointmentEnabled {
		return "", fmt.Errorf("request appointment disabled for this location")
	}
	return fmt.Sprintf("https://sighthub.cloud/req/app/%s", store.Hash), nil
}

func (s *Service) CreateStore(input StoreInput) (*StoreDetailResponse, error) {
	details := input.Details
	address := input.Address
	schedData := input.Schedule

	if err := validateStoreData(details, address, false); err != nil {
		return nil, err
	}

	fullName := strFromMap(details, "full_name")
	shortName := strFromMap(details, "short_name")

	store := locModel.Store{
		BusinessName:  strPtr(strFromMap(details, "business_name")),
		FullName:      strPtr(fullName),
		ShortName:     strPtr(shortName),
		Phone:         strPtr(strFromMap(details, "phone")),
		Fax:           strPtr(strFromMap(details, "fax")),
		Email:         strPtr(strFromMap(details, "email")),
		NPI:           strPtr(strFromMap(details, "npi")),
		TaxN:          strPtr(strFromMap(details, "tax_n")),
		HPSA:          strPtr(strFromMap(details, "hpsa")),
		Logo:          strPtr(strFromMap(details, "logo")),
		StreetAddress: strPtr(strFromMap(address, "street_address")),
		AddressLine2:  strPtr(strFromMap(address, "address_line_2")),
		City:          strPtr(strFromMap(address, "city")),
		State:         strPtr(strFromMap(address, "state")),
		PostalCode:    strPtr(strFromMap(address, "postal_code")),
		Country:       strPtr(strFromMap(address, "country")),
	}

	var resp *StoreDetailResponse
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&store).Error; err != nil {
			return err
		}

		var salesTaxID *int
		if st := strFromMap(address, "state"); st != "" {
			var tax genModel.SalesTaxByState
			if err := tx.Where("state_code = ?", st).First(&tax).Error; err == nil {
				salesTaxID = &tax.IDSalesTax
			}
		}

		canReceive := boolFromMap(details, "can_receive_items", true)
		showcase := boolFromMap(details, "showcase", false)
		loc := locModel.Location{
			FullName:        fullName,
			ShortName:       strPtr(shortName),
			StreetAddress:   strPtr(strFromMap(address, "street_address")),
			AddressLine2:    strPtr(strFromMap(address, "address_line_2")),
			City:            strPtr(strFromMap(address, "city")),
			State:           strPtr(strFromMap(address, "state")),
			PostalCode:      strPtr(strFromMap(address, "postal_code")),
			Country:         strPtr(strFromMap(address, "country")),
			Phone:           strPtr(strFromMap(details, "phone")),
			Website:         strPtr(strFromMap(details, "website")),
			Email:           strPtr(strFromMap(details, "email")),
			Fax:             strPtr(strFromMap(details, "fax")),
			CanReceiveItems: &canReceive,
			Showcase:        &showcase,
			LogoPath:        strPtr(strFromMap(details, "logo")),
			StoreID:         store.IDStore,
			SalesTaxID:      salesTaxID,
		}

		var workShift *empModel.WorkShift
		if len(schedData) > 0 {
			ws, err := parseWorkShift(schedData)
			if err != nil {
				return fmt.Errorf("invalid work shift data: %w", err)
			}
			if err := tx.Create(ws).Error; err != nil {
				return err
			}
			wsID := int(ws.IDWorkShift)
			loc.WorkShiftID = &wsID
			workShift = ws
		}

		if err := tx.Create(&loc).Error; err != nil {
			return err
		}

		subBlock := permModel.PermissionsSubBlockStore{
			SubBlockName: fullName,
			StoreID:      &store.IDStore,
		}
		if err := tx.Create(&subBlock).Error; err != nil {
			return err
		}

		if err := createStorePermCombinations(tx, subBlock.IDPermissionsSubBlock); err != nil {
			return err
		}

		pkgActivity.Log(tx, "store", "create",
			pkgActivity.WithEntity(int64(store.IDStore)),
			pkgActivity.WithDetails(map[string]interface{}{"name": store.FullName}),
		)

		var schedMap interface{}
		if workShift != nil {
			schedMap = workShift.ToMap()
		}
		resp = s.buildStoreDetailResponse(&store, &loc, schedMap)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *Service) UpdateStore(storeID int, input StoreInput) error {
	details := input.Details
	address := input.Address
	schedData := input.Schedule

	if err := validateStoreData(details, address, true); err != nil {
		return err
	}

	var store locModel.Store
	if err := s.db.First(&store, storeID).Error; err != nil {
		return fmt.Errorf("store not found")
	}
	var loc locModel.Location
	if err := s.db.Where("store_id = ? AND warehouse_id IS NULL", storeID).First(&loc).Error; err != nil {
		return fmt.Errorf("location not found")
	}

	if v := strFromMap(details, "business_name"); v != "" { store.BusinessName = &v }
	if v := strFromMap(details, "full_name"); v != ""     { store.FullName = &v; loc.FullName = v }
	if v := strFromMap(details, "short_name"); v != ""    { store.ShortName = &v; loc.ShortName = &v }
	if v := strFromMap(details, "phone"); v != ""         { store.Phone = &v; loc.Phone = &v }
	if v := strFromMap(details, "email"); v != ""         { store.Email = &v; loc.Email = &v }
	if v := strFromMap(details, "fax"); v != ""           { store.Fax = &v; loc.Fax = &v }
	if v := strFromMap(details, "npi"); v != ""           { store.NPI = &v }
	if v := strFromMap(details, "tax_n"); v != ""         { store.TaxN = &v }
	if v := strFromMap(details, "hpsa"); v != ""          { store.HPSA = &v }
	if v := strFromMap(details, "logo"); v != ""          { store.Logo = &v; loc.LogoPath = &v }
	if v := strFromMap(details, "website"); v != ""       { loc.Website = &v }
	if _, ok := details["can_receive_items"]; ok {
		v := boolFromMap(details, "can_receive_items", true)
		loc.CanReceiveItems = &v
	}
	if _, ok := details["store_active"]; ok {
		v := boolFromMap(details, "store_active", false)
		loc.StoreActive = &v
	}

	if v := strFromMap(address, "street_address"); v != "" { store.StreetAddress = &v; loc.StreetAddress = &v }
	if v := strFromMap(address, "address_line_2"); v != "" { store.AddressLine2 = &v; loc.AddressLine2 = &v }
	if v := strFromMap(address, "city"); v != ""           { store.City = &v; loc.City = &v }
	if v := strFromMap(address, "state"); v != ""          { store.State = &v; loc.State = &v }
	if v := strFromMap(address, "postal_code"); v != ""    { store.PostalCode = &v; loc.PostalCode = &v }
	if v := strFromMap(address, "country"); v != ""        { store.Country = &v; loc.Country = &v }

	return s.db.Transaction(func(tx *gorm.DB) error {
		if len(schedData) > 0 {
			if loc.WorkShiftID != nil {
				var ws empModel.WorkShift
				if err := tx.First(&ws, *loc.WorkShiftID).Error; err == nil {
					updateWorkShift(&ws, schedData)
					tx.Save(&ws)
				}
			} else {
				ws, err := parseWorkShift(schedData)
				if err != nil {
					return fmt.Errorf("invalid work shift data: %w", err)
				}
				if err := tx.Create(ws).Error; err != nil {
					return err
				}
				wsID := int(ws.IDWorkShift)
				loc.WorkShiftID = &wsID
			}
		}
		if err := tx.Save(&store).Error; err != nil {
			return err
		}
		if err := tx.Save(&loc).Error; err != nil {
			return err
		}
		pkgActivity.Log(tx, "store", "update",
			pkgActivity.WithEntity(int64(storeID)),
			pkgActivity.WithDetails(map[string]interface{}{"name": store.FullName}),
		)
		return nil
	})
}

func (s *Service) ActivateStore(storeID int, active *bool) (map[string]interface{}, int, error) {
	var store locModel.Store
	if err := s.db.First(&store, storeID).Error; err != nil {
		return nil, 0, fmt.Errorf("store not found")
	}

	payload := map[string]interface{}{
		"hash":        store.Hash,
		"license_key": store.LicenseKey,
	}
	if active != nil {
		payload["active"] = *active
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "https://manage.sighthub.cloud/license/sighthub/store", bytes.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build request")
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	httpResp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to contact license service")
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(httpResp.Body)
	var responseData map[string]interface{}
	json.Unmarshal(respBody, &responseData) //nolint:errcheck

	if store.Email != nil && *store.Email != "" {
		var loc locModel.Location
		if err := s.db.Where("store_id = ? AND warehouse_id IS NULL", storeID).First(&loc).Error; err == nil {
			locID := int64(loc.IDLocation)
			msg := ""
			if m, ok := responseData["message"].(string); ok {
				msg = m
			}
			smtpCfg, err := pkgEmail.GetSMTPForLocation(s.db, &locID)
			if err == nil {
				pkgEmail.Send(smtpCfg, *store.Email, "License activation response", msg) //nolint:errcheck
			}
		}
	}

	return responseData, httpResp.StatusCode, nil
}

// --- Warehouses ---

func (s *Service) GetWarehouses() ([]WarehouseListItem, error) {
	var warehouses []locModel.Warehouse
	if err := s.db.Find(&warehouses).Error; err != nil {
		return nil, err
	}

	result := make([]WarehouseListItem, 0, len(warehouses))
	for _, wh := range warehouses {
		item := WarehouseListItem{
			ShortName:     wh.ShortName,
			WarehouseName: wh.FullName,
			WarehouseID:   wh.IDWarehouse,
		}

		var loc locModel.Location
		if err := s.db.Where("warehouse_id = ?", wh.IDWarehouse).First(&loc).Error; err == nil {
			storeID := loc.StoreID
			item.StoreID = &storeID
			item.Active = loc.StoreActive
			item.Phone = loc.Phone
			if item.Phone == nil {
				item.Phone = wh.Phone
			}
			parts := addrParts(loc.StreetAddress, loc.AddressLine2, loc.City, loc.State, loc.PostalCode, loc.Country)
			item.Address = strings.Join(parts, ", ")

			var st locModel.Store
			if err := s.db.First(&st, loc.StoreID).Error; err == nil {
				item.ConnectStore = st.FullName
			}
		} else {
			item.Phone = wh.Phone
			parts := addrParts(wh.StreetAddress, wh.AddressLine2, wh.City, wh.State, wh.PostalCode, wh.Country)
			item.Address = strings.Join(parts, ", ")
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *Service) GetWarehouse(warehouseID int) (*WarehouseDetailResponse, error) {
	var wh locModel.Warehouse
	if err := s.db.First(&wh, warehouseID).Error; err != nil {
		return nil, fmt.Errorf("warehouse not found")
	}

	resp := &WarehouseDetailResponse{
		ShortName:     wh.ShortName,
		WarehouseName: wh.FullName,
	}

	var loc locModel.Location
	if err := s.db.Where("warehouse_id = ?", warehouseID).First(&loc).Error; err == nil {
		storeID := loc.StoreID
		resp.StoreID = &storeID
		resp.StreetAddress = loc.StreetAddress
		resp.AddressLine2 = loc.AddressLine2
		resp.City = loc.City
		resp.State = loc.State
		resp.PostalCode = loc.PostalCode
		resp.Phone = loc.Phone
		resp.Email = loc.Email
		resp.CanReceiveItems = loc.CanReceiveItems
		resp.Active = loc.StoreActive

		var st locModel.Store
		if err := s.db.First(&st, loc.StoreID).Error; err == nil {
			resp.ConnectStore = st.FullName
		}
	} else {
		resp.StreetAddress = wh.StreetAddress
		resp.AddressLine2 = wh.AddressLine2
		resp.City = wh.City
		resp.State = wh.State
		resp.PostalCode = wh.PostalCode
		resp.Phone = wh.Phone
	}
	return resp, nil
}

func (s *Service) UpdateWarehouse(warehouseID int, data map[string]interface{}) error {
	var wh locModel.Warehouse
	if err := s.db.First(&wh, warehouseID).Error; err != nil {
		return fmt.Errorf("warehouse not found")
	}
	var loc locModel.Location
	if err := s.db.Where("warehouse_id = ?", warehouseID).First(&loc).Error; err != nil {
		return fmt.Errorf("location not found")
	}

	if v, ok := data["short_name"].(string); ok     { wh.ShortName = &v }
	if v, ok := data["warehouse_name"].(string); ok { wh.FullName = &v }
	if v, ok := data["street_address"].(string); ok { wh.StreetAddress = &v; loc.StreetAddress = &v }
	if v, ok := data["address_line_2"].(string); ok { wh.AddressLine2 = &v; loc.AddressLine2 = &v }
	if v, ok := data["city"].(string); ok            { wh.City = &v; loc.City = &v }
	if v, ok := data["state"].(string); ok           { wh.State = &v; loc.State = &v }
	if v, ok := data["postal_code"].(string); ok     { wh.PostalCode = &v; loc.PostalCode = &v }
	if v, ok := data["phone"].(string); ok           { wh.Phone = &v; loc.Phone = &v }
	if v, ok := data["email"].(string); ok           { loc.Email = &v }
	if v, ok := data["store_id"].(float64); ok       { loc.StoreID = int(v) }
	if _, ok := data["can_receive_items"]; ok {
		v := boolFromMap(data, "can_receive_items", true)
		loc.CanReceiveItems = &v
	}
	if _, ok := data["active"]; ok {
		v := boolFromMap(data, "active", false)
		loc.StoreActive = &v
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&wh).Error; err != nil {
			return err
		}
		if err := tx.Save(&loc).Error; err != nil {
			return err
		}
		pkgActivity.Log(tx, "store", "warehouse_update",
			pkgActivity.WithEntity(int64(warehouseID)),
			pkgActivity.WithDetails(map[string]interface{}{"name": wh.FullName}),
		)
		return nil
	})
}

func (s *Service) CreateWarehouse(data map[string]interface{}) (int, error) {
	shortName, _ := data["short_name"].(string)
	fullName, _ := data["warehouse_name"].(string)
	storeIDf, _ := data["store_id"].(float64)
	storeID := int(storeIDf)

	if shortName == "" || fullName == "" || storeID == 0 {
		return 0, fmt.Errorf("short_name, warehouse_name and store_id are required")
	}

	var whID int
	err := s.db.Transaction(func(tx *gorm.DB) error {
		wh := locModel.Warehouse{
			FullName:      &fullName,
			ShortName:     &shortName,
			StreetAddress: strPtr(strFromMap(data, "street_address")),
			AddressLine2:  strPtr(strFromMap(data, "address_line_2")),
			City:          strPtr(strFromMap(data, "city")),
			State:         strPtr(strFromMap(data, "state")),
			PostalCode:    strPtr(strFromMap(data, "postal_code")),
			Country:       strPtr(strFromMap(data, "country")),
			Phone:         strPtr(strFromMap(data, "phone")),
		}
		if err := tx.Create(&wh).Error; err != nil {
			return err
		}
		whID = wh.IDWarehouse

		canReceive := boolFromMap(data, "can_receive_items", true)
		active := boolFromMap(data, "active", false)
		loc := locModel.Location{
			FullName:        fullName,
			ShortName:       &shortName,
			StreetAddress:   strPtr(strFromMap(data, "street_address")),
			AddressLine2:    strPtr(strFromMap(data, "address_line_2")),
			City:            strPtr(strFromMap(data, "city")),
			State:           strPtr(strFromMap(data, "state")),
			PostalCode:      strPtr(strFromMap(data, "postal_code")),
			Country:         strPtr(strFromMap(data, "country")),
			Phone:           strPtr(strFromMap(data, "phone")),
			Email:           strPtr(strFromMap(data, "email")),
			CanReceiveItems: &canReceive,
			StoreActive:     &active,
			StoreID:         storeID,
			WarehouseID:     &wh.IDWarehouse,
		}
		if err := tx.Create(&loc).Error; err != nil {
			return err
		}

		subBlock := permModel.PermissionsSubBlockWarehouse{
			SubBlockName: fullName,
			WarehouseID:  &wh.IDWarehouse,
		}
		if err := tx.Create(&subBlock).Error; err != nil {
			return err
		}

		if err := createWarehousePermCombinations(tx, subBlock.IDPermissionsSubBlock); err != nil {
			return err
		}

		pkgActivity.Log(tx, "store", "warehouse_create",
			pkgActivity.WithEntity(int64(wh.IDWarehouse)),
			pkgActivity.WithDetails(map[string]interface{}{"name": wh.FullName}),
		)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return whID, nil
}

// --- Sales Tax ---

func (s *Service) GetSalesTaxList() ([]SalesTaxListItem, error) {
	var taxes []genModel.SalesTaxByState
	if err := s.db.Select("id_sales_tax, state_code, sales_tax_percent").Find(&taxes).Error; err != nil {
		return nil, err
	}
	result := make([]SalesTaxListItem, 0, len(taxes))
	for _, t := range taxes {
		result = append(result, SalesTaxListItem{
			IDSalesTax:      t.IDSalesTax,
			StateCode:       t.StateCode,
			SalesTaxPercent: fmt.Sprintf("%g", t.SalesTaxPercent),
		})
	}
	return result, nil
}

func (s *Service) GetSalesTaxes() ([]SalesTaxItem, error) {
	var taxes []genModel.SalesTaxByState
	if err := s.db.Where("tax_active = true").Find(&taxes).Error; err != nil {
		return nil, err
	}
	result := make([]SalesTaxItem, 0, len(taxes))
	for _, t := range taxes {
		result = append(result, SalesTaxItem{
			Code:    t.StateCode,
			State:   t.StateName,
			SaleTax: fmt.Sprintf("%.4f%%", t.SalesTaxPercent),
		})
	}
	return result, nil
}

// --- Helpers ---

func (s *Service) buildStoreDetailResponse(store *locModel.Store, loc *locModel.Location, sched interface{}) *StoreDetailResponse {
	return &StoreDetailResponse{
		Details: map[string]interface{}{
			"business_name":    store.BusinessName,
			"full_name":        store.FullName,
			"short_name":       store.ShortName,
			"hash":             store.Hash,
			"website":          loc.Website,
			"phone":            store.Phone,
			"email":            store.Email,
			"fax":              store.Fax,
			"can_receive_items": loc.CanReceiveItems,
			"logo":             store.Logo,
			"npi":              store.NPI,
			"tax_n":            store.TaxN,
			"hpsa":             store.HPSA,
			"store_active":     loc.StoreActive,
		},
		Address: map[string]interface{}{
			"street_address": store.StreetAddress,
			"address_line_2": store.AddressLine2,
			"city":           store.City,
			"state":          store.State,
			"postal_code":    store.PostalCode,
			"country":        store.Country,
		},
		Schedule: sched,
	}
}

func validateStoreData(details, address map[string]interface{}, partial bool) error {
	if !partial {
		if strings.TrimSpace(strFromMap(details, "full_name")) == "" {
			return fmt.Errorf("full_name is required")
		}
		if strings.TrimSpace(strFromMap(details, "short_name")) == "" {
			return fmt.Errorf("short_name is required")
		}
	}
	if sn, ok := details["short_name"].(string); ok && sn != "" {
		if len(strings.TrimSpace(sn)) != 2 {
			return fmt.Errorf("short_name must be 2 characters")
		}
	}
	if st, ok := address["state"].(string); ok && st != "" {
		if len(strings.TrimSpace(st)) != 2 {
			return fmt.Errorf("state must be 2 characters")
		}
	}
	if pc, ok := address["postal_code"].(string); ok && pc != "" {
		for _, c := range strings.TrimSpace(pc) {
			if c < '0' || c > '9' {
				return fmt.Errorf("postal_code must be numeric")
			}
		}
	}
	return nil
}

func parseWorkShift(data map[string]interface{}) (*empModel.WorkShift, error) {
	ld := strFromMap(data, "lunch_duration")
	if ld == "" {
		ld = "00:30:00"
	}
	ws := &empModel.WorkShift{
		Monday:             boolFromMap(data, "monday", true),
		Tuesday:            boolFromMap(data, "tuesday", true),
		Wednesday:          boolFromMap(data, "wednesday", true),
		Thursday:           boolFromMap(data, "thursday", true),
		Friday:             boolFromMap(data, "friday", true),
		Saturday:           boolFromMap(data, "saturday", false),
		Sunday:             boolFromMap(data, "sunday", false),
		MondayTimeStart:    parseTimeStr(strFromMap(data, "monday_time_start"), "10:00:00"),
		MondayTimeEnd:      parseTimeStr(strFromMap(data, "monday_time_end"), "19:00:00"),
		TuesdayTimeStart:   parseTimeStr(strFromMap(data, "tuesday_time_start"), "10:00:00"),
		TuesdayTimeEnd:     parseTimeStr(strFromMap(data, "tuesday_time_end"), "19:00:00"),
		WednesdayTimeStart: parseTimeStr(strFromMap(data, "wednesday_time_start"), "10:00:00"),
		WednesdayTimeEnd:   parseTimeStr(strFromMap(data, "wednesday_time_end"), "19:00:00"),
		ThursdayTimeStart:  parseTimeStr(strFromMap(data, "thursday_time_start"), "10:00:00"),
		ThursdayTimeEnd:    parseTimeStr(strFromMap(data, "thursday_time_end"), "19:00:00"),
		FridayTimeStart:    parseTimeStr(strFromMap(data, "friday_time_start"), "10:00:00"),
		FridayTimeEnd:      parseTimeStr(strFromMap(data, "friday_time_end"), "19:00:00"),
		SaturdayTimeStart:  parseTimeStrPtr(strFromMap(data, "saturday_time_start")),
		SaturdayTimeEnd:    parseTimeStrPtr(strFromMap(data, "saturday_time_end")),
		SundayTimeStart:    parseTimeStrPtr(strFromMap(data, "sunday_time_start")),
		SundayTimeEnd:      parseTimeStrPtr(strFromMap(data, "sunday_time_end")),
		LunchDuration:      &ld,
	}
	if title := strFromMap(data, "title"); title != "" {
		ws.Title = &title
	}
	return ws, nil
}

func updateWorkShift(ws *empModel.WorkShift, data map[string]interface{}) {
	if v, ok := data["monday"].(bool);    ok { ws.Monday = v }
	if v, ok := data["tuesday"].(bool);   ok { ws.Tuesday = v }
	if v, ok := data["wednesday"].(bool); ok { ws.Wednesday = v }
	if v, ok := data["thursday"].(bool);  ok { ws.Thursday = v }
	if v, ok := data["friday"].(bool);    ok { ws.Friday = v }
	if v, ok := data["saturday"].(bool);  ok { ws.Saturday = v }
	if v, ok := data["sunday"].(bool);    ok { ws.Sunday = v }

	setTime := func(key string, dst *time.Time) {
		if v, ok := data[key].(string); ok { *dst = parseTimeStr(v, "") }
	}
	setTime("monday_time_start",    &ws.MondayTimeStart)
	setTime("monday_time_end",      &ws.MondayTimeEnd)
	setTime("tuesday_time_start",   &ws.TuesdayTimeStart)
	setTime("tuesday_time_end",     &ws.TuesdayTimeEnd)
	setTime("wednesday_time_start", &ws.WednesdayTimeStart)
	setTime("wednesday_time_end",   &ws.WednesdayTimeEnd)
	setTime("thursday_time_start",  &ws.ThursdayTimeStart)
	setTime("thursday_time_end",    &ws.ThursdayTimeEnd)
	setTime("friday_time_start",    &ws.FridayTimeStart)
	setTime("friday_time_end",      &ws.FridayTimeEnd)

	setTimePtr := func(key string, dst **time.Time) {
		if v, ok := data[key].(string); ok { *dst = parseTimeStrPtr(v) }
	}
	setTimePtr("saturday_time_start", &ws.SaturdayTimeStart)
	setTimePtr("saturday_time_end",   &ws.SaturdayTimeEnd)
	setTimePtr("sunday_time_start",   &ws.SundayTimeStart)
	setTimePtr("sunday_time_end",     &ws.SundayTimeEnd)

	if v, ok := data["title"].(string);          ok { ws.Title = &v }
	if v, ok := data["lunch_duration"].(string);  ok { ws.LunchDuration = &v }
}

func createStorePermCombinations(tx *gorm.DB, subBlockID int) error {
	var perms []permModel.Permissions
	tx.Where("permissions_id IN ?", storePermissions).Find(&perms)
	for _, perm := range perms {
		var existing permModel.PermissionsCombination
		err := tx.Where("permissions_block_id = ? AND permissions_sub_block_store_id = ? AND permissions_id = ?",
			perm.PermissionsBlockID, subBlockID, perm.PermissionsID).First(&existing).Error
		if err != nil {
			combo := permModel.PermissionsCombination{
				PermissionsBlockID:         &perm.PermissionsBlockID,
				PermissionsSubBlockStoreID: &subBlockID,
				PermissionsID:              perm.PermissionsID,
			}
			if err := tx.Create(&combo).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func createWarehousePermCombinations(tx *gorm.DB, subBlockID int) error {
	var perms []permModel.Permissions
	tx.Where("permissions_id IN ?", warehousePermissions).Find(&perms)
	for _, perm := range perms {
		var existing permModel.PermissionsCombination
		err := tx.Where("permissions_block_id = ? AND permissions_sub_block_warehouse_id = ? AND permissions_id = ?",
			perm.PermissionsBlockID, subBlockID, perm.PermissionsID).First(&existing).Error
		if err != nil {
			combo := permModel.PermissionsCombination{
				PermissionsBlockID:             &perm.PermissionsBlockID,
				PermissionsSubBlockWarehouseID: &subBlockID,
				PermissionsID:                  perm.PermissionsID,
			}
			if err := tx.Create(&combo).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func addrParts(parts ...*string) []string {
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != nil && *p != "" {
			result = append(result, *p)
		}
	}
	return result
}

func parseTimeStr(s, def string) time.Time {
	if s == "" {
		s = def
	}
	t, _ := time.Parse("15:04:05", s)
	return t
}

func parseTimeStrPtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("15:04:05", s)
	if err != nil {
		return nil
	}
	return &t
}

func strFromMap(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}

func boolFromMap(m map[string]interface{}, key string, def bool) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
