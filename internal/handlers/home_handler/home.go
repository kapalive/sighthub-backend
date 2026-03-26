package home_handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	"sighthub-backend/internal/models/general"
	"sighthub-backend/internal/models/insurance"
	"sighthub-backend/internal/models/invoices"
	"sighthub-backend/internal/models/location"
	"sighthub-backend/internal/models/marketing"
	"sighthub-backend/internal/models/patients"
	"sighthub-backend/internal/middleware"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// GET /home
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	emp := middleware.EmployeeFromCtx(r.Context())
	loc := middleware.LocationFromCtx(r.Context())
	allowed := middleware.AllowedNavigationFromCtx(r.Context())

	var items []general.NavigationItem
	h.db.Order("position ASC, id_navigation_item ASC").Find(&items)

	allowedPerms := middleware.AllowedPermIDsFromCtx(r.Context())
	navItems := []map[string]interface{}{}
	for _, item := range items {
		itemPath := normalizePath(item.Path)

		// If item has permissions_id, check if user has that permission
		if item.PermissionsID != nil {
			if _, ok := allowedPerms[*item.PermissionsID]; !ok {
				continue
			}
			navItems = append(navItems, map[string]interface{}{
				"name": item.Label,
				"url":  itemPath,
				"icon": item.Icon,
			})
			continue
		}

		if _, ok := allowed[itemPath]; ok {
			navItems = append(navItems, map[string]interface{}{
				"name": item.Label,
				"url":  itemPath,
				"icon": item.Icon,
			})
			continue
		}
		// check prefix match
		for p := range allowed {
			p = normalizePath(p)
			if strings.HasPrefix(itemPath, p+"/") || strings.HasPrefix(p, itemPath+"/") {
				navItems = append(navItems, map[string]interface{}{
					"name": item.Label,
					"url":  itemPath,
					"icon": item.Icon,
				})
				break
			}
		}
	}

	shortName := ""
	if loc.ShortName != nil {
		shortName = *loc.ShortName
	}

	jsonResponse(w, map[string]interface{}{
		"location_id":     loc.IDLocation,
		"employee_id":     emp.IDEmployee,
		"first_name":      emp.FirstName,
		"store_name":      loc.FullName,
		"short_name":      shortName,
		"home_navigation": navItems,
	}, http.StatusOK)
}

// GET /notify/counts
func (h *Handler) NotifyCounts(w http.ResponseWriter, r *http.Request) {
	emp := middleware.EmployeeFromCtx(r.Context())
	loc := middleware.LocationFromCtx(r.Context())
	allowed := middleware.AllowedNavigationFromCtx(r.Context())

	counts := []map[string]interface{}{}

	if middleware.PathAllowed(allowed, "/accounting") {
		c := h.countVendorTermsDue(loc)
		counts = append(counts, map[string]interface{}{
			"url":   "/accounting",
			"count": c,
		})
	}

	if middleware.PathAllowed(allowed, "/tasks") {
		c := h.countTasksTodoAssigned(emp.IDEmployee)
		counts = append(counts, map[string]interface{}{
			"url":   "/tasks",
			"count": c,
		})
	}

	if middleware.PathAllowed(allowed, "/appointment") {
		c := h.countRequestAppointmentsUnprocessed(loc.IDLocation)
		counts = append(counts, map[string]interface{}{
			"url":   "/appointment/requests",
			"count": c,
		})
	}

	jsonResponse(w, map[string]interface{}{"counts": counts}, http.StatusOK)
}

// GET /set-stores-list
func (h *Handler) GetStoresList(w http.ResponseWriter, r *http.Request) {
	allowedLocs := middleware.PermittedLocationsFromCtx(r.Context())
	allowedSet := make(map[int]struct{}, len(allowedLocs))
	for _, l := range allowedLocs {
		allowedSet[l.IDLocation] = struct{}{}
	}

	var stores []location.Store
	h.db.Find(&stores)

	storeList := []map[string]interface{}{}

	for _, store := range stores {
		// Main store location (no warehouse)
		var storeLoc location.Location
		err := h.db.Where("store_id = ? AND warehouse_id IS NULL", store.IDStore).First(&storeLoc).Error
		if err != nil {
			continue
		}
		if _, ok := allowedSet[storeLoc.IDLocation]; !ok {
			continue
		}

		// Warehouse locations
		var whLocs []location.Location
		h.db.Where("store_id = ? AND warehouse_id IS NOT NULL", store.IDStore).Find(&whLocs)

		warehouses := []map[string]interface{}{}
		for _, wl := range whLocs {
			if _, ok := allowedSet[wl.IDLocation]; !ok {
				continue
			}
			var wh location.Warehouse
			whName := ""
			whFull := ""
			if wl.WarehouseID != nil {
				h.db.First(&wh, *wl.WarehouseID)
				if wh.ShortName != nil {
					whName = *wh.ShortName
				}
				if wh.FullName != nil {
					whFull = *wh.FullName
				}
			}
			warehouses = append(warehouses, map[string]interface{}{
				"location_id":  wl.IDLocation,
				"warehouse_id": wl.WarehouseID,
				"short_name":   whName,
				"full_name":    whFull,
			})
		}

		storeInfo := map[string]interface{}{
			"location_id": storeLoc.IDLocation,
			"full_name":   store.FullName,
			"short_name":  store.ShortName,
		}
		if len(warehouses) > 0 {
			storeInfo["warehouses"] = warehouses
		}
		storeList = append(storeList, storeInfo)
	}

	jsonResponse(w, storeList, http.StatusOK)
}

// POST /set_store
func (h *Handler) SetStore(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	permittedIDs := middleware.PermittedLocationIDsFromCtx(r.Context())

	var login authModel.EmployeeLogin
	if err := h.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}

	var body struct {
		LocationID int `json:"location_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.LocationID == 0 {
		jsonError(w, "Missing or invalid location_id", http.StatusBadRequest)
		return
	}

	permitted := false
	for _, id := range permittedIDs {
		if id == body.LocationID {
			permitted = true
			break
		}
	}
	if !permitted {
		jsonError(w, "Permission denied for this location", http.StatusForbidden)
		return
	}

	var loc location.Location
	if err := h.db.Where("id_location = ? AND store_active = true", body.LocationID).First(&loc).Error; err != nil {
		jsonError(w, "Location not found", http.StatusNotFound)
		return
	}

	locID := int64(body.LocationID)
	h.db.Model(&struct {
		EmployeeLoginID int64
		LocationID      *int64
	}{}).Table("employee").Where("employee_login_id = ?", login.IDEmployeeLogin).
		Update("location_id", &locID)

	jsonResponse(w, map[string]interface{}{
		"location_id": loc.IDLocation,
		"location":    loc.FullName,
	}, http.StatusOK)
}

// GET /invoice/search
func (h *Handler) SearchInvoice(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var login authModel.EmployeeLogin
	if err := h.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}
	var emp struct {
		LocationID *int64
	}
	h.db.Table("employee").Select("location_id").
		Where("employee_login_id = ?", login.IDEmployeeLogin).Scan(&emp)
	if emp.LocationID == nil {
		jsonError(w, "Employee has no location", http.StatusNotFound)
		return
	}
	locID := *emp.LocationID

	q := strings.TrimSpace(r.URL.Query().Get("number_invoice"))
	if q == "" {
		jsonError(w, "Invoice number not provided", http.StatusBadRequest)
		return
	}
	like := "%" + q + "%"

	var invList []invoices.Invoice
	if isDigits(q) {
		h.db.Where("CAST(id_invoice AS TEXT) ILIKE ? OR number_invoice ILIKE ?", like, like).Find(&invList)
	} else {
		h.db.Where("number_invoice ILIKE ?", like).Find(&invList)
	}

	if len(invList) == 0 {
		jsonError(w, "No invoices found", http.StatusNotFound)
		return
	}

	accessible := []map[string]interface{}{}
	for _, inv := range invList {
		num := inv.NumberInvoice
		if strings.HasPrefix(num, "V") {
			if inv.LocationID == locID {
				accessible = append(accessible, map[string]interface{}{
					"id_invoice":     inv.IDInvoice,
					"number_invoice": num,
					"redirect_url":   fmt.Sprintf("/inventory/receipts/vendors/invoice/%d/%d/*", inv.IDInvoice, inv.VendorID),
				})
			}
		} else if strings.HasPrefix(num, "I") {
			if inv.LocationID == locID {
				accessible = append(accessible, map[string]interface{}{
					"id_invoice":     inv.IDInvoice,
					"number_invoice": num,
					"redirect_url":   fmt.Sprintf("/inventory/transfers/invoice/%d", inv.IDInvoice),
				})
			} else if inv.ToLocationID != nil && *inv.ToLocationID == locID {
				accessible = append(accessible, map[string]interface{}{
					"id_invoice":     inv.IDInvoice,
					"number_invoice": num,
					"redirect_url":   fmt.Sprintf("/receipts/store/invoice/%d", inv.IDInvoice),
				})
			}
		} else if strings.HasPrefix(num, "S") {
			if inv.LocationID == locID {
				accessible = append(accessible, map[string]interface{}{
					"id_invoice":     inv.IDInvoice,
					"number_invoice": num,
					"redirect_url":   fmt.Sprintf("/patient/%d/invoice/%d", derefInt64(inv.PatientID), inv.IDInvoice),
				})
			}
		}
	}

	if len(accessible) == 0 {
		jsonError(w, "No accessible invoices found", http.StatusForbidden)
		return
	}
	jsonResponse(w, map[string]interface{}{"invoices": accessible}, http.StatusOK)
}

// GET /locations
func (h *Handler) GetLocations(w http.ResponseWriter, r *http.Request) {
	var locs []location.Location
	h.db.Where("showcase = true").Find(&locs)

	result := make([]map[string]interface{}, 0, len(locs))
	for _, l := range locs {
		result = append(result, map[string]interface{}{
			"location_id": l.IDLocation,
			"location":    l.FullName,
		})
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /insurance/companies
func (h *Handler) GetInsuranceCompanies(w http.ResponseWriter, r *http.Request) {
	var companies []insurance.InsuranceCompany
	h.db.Find(&companies)

	result := make([]map[string]interface{}, 0, len(companies))
	for _, c := range companies {
		result = append(result, map[string]interface{}{
			"insurance_company_id": c.IDInsuranceCompany,
			"company_name":         c.CompanyName,
		})
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /insurance/types
func (h *Handler) GetInsuranceTypes(w http.ResponseWriter, r *http.Request) {
	var types []insurance.InsuranceCoverageType
	h.db.Order("coverage_name asc").Find(&types)

	result := make([]map[string]interface{}, 0, len(types))
	for _, t := range types {
		result = append(result, map[string]interface{}{
			"id_type_insurance_policy": t.IDInsuranceCoverageType,
			"type_name":                t.CoverageName,
		})
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /languages
func (h *Handler) GetLanguages(w http.ResponseWriter, r *http.Request) {
	var langs []patients.PreferredLanguage
	h.db.Find(&langs)

	result := make([]map[string]interface{}, 0, len(langs))
	for _, l := range langs {
		result = append(result, map[string]interface{}{
			"id_preferred_language": l.IDPreferredLanguage,
			"language":              l.Language,
		})
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /patient/search
func (h *Handler) SearchPatients(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var login authModel.EmployeeLogin
	if err := h.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}
	var emp struct{ LocationID *int64 }
	h.db.Table("employee").Select("location_id").
		Where("employee_login_id = ?", login.IDEmployeeLogin).Scan(&emp)
	if emp.LocationID == nil {
		jsonError(w, "Employee has no location", http.StatusNotFound)
		return
	}
	locID := *emp.LocationID

	q := r.URL.Query()
	db := h.db.Model(&patients.Patient{}).Where("location_id = ?", locID)

	if v := strings.TrimSpace(q.Get("first_name")); v != "" {
		db = db.Where("first_name ILIKE ?", v+"%")
	}
	if v := strings.TrimSpace(q.Get("last_name")); v != "" {
		db = db.Where("last_name ILIKE ?", v+"%")
	}
	if v := strings.TrimSpace(q.Get("dob")); v != "" {
		db = db.Where("dob = ?", v)
	}
	if v := strings.TrimSpace(q.Get("phone")); v != "" {
		db = db.Where("phone ILIKE ?", v+"%")
	}
	if v := strings.TrimSpace(q.Get("email")); v != "" {
		db = db.Where("email ILIKE ?", v+"%")
	}
	if v := strings.TrimSpace(q.Get("city")); v != "" {
		db = db.Where("city ILIKE ?", v+"%")
	}
	if v := strings.TrimSpace(q.Get("state")); v != "" {
		db = db.Where("state ILIKE ?", v+"%")
	}
	if v := strings.TrimSpace(q.Get("zip_code")); v != "" {
		db = db.Where("zip_code ILIKE ?", "%"+v+"%")
	}

	var patList []patients.Patient
	db.Order("last_name asc, first_name asc").Limit(25).Find(&patList)

	if len(patList) == 0 {
		jsonError(w, "No patients found with the specified parameters", http.StatusNotFound)
		return
	}

	result := make([]map[string]interface{}, 0, len(patList))
	for _, p := range patList {
		dob := ""
		if p.DOB != nil {
			dob = p.DOB.Format("2006-01-02")
		}
		addr := strings.TrimSpace(fmt.Sprintf("%s %s %s %s",
			strVal(p.StreetAddress), strVal(p.City), strVal(p.State), strVal(p.ZipCode)))
		result = append(result, map[string]interface{}{
			"id":         p.IDPatient,
			"last_name":  p.LastName,
			"first_name": p.FirstName,
			"dob":        dob,
			"address":    addr,
			"phone":      strVal(p.Phone),
			"email":      strVal(p.Email),
			"gender":     func() string { if p.Gender != nil { return string(*p.Gender) }; return "" }(),
		})
	}

	jsonResponse(w, map[string]interface{}{
		"total_items": len(result),
		"patients":    result,
	}, http.StatusOK)
}

// GET /patient/recently-viewed
func (h *Handler) RecentlyViewedPatients(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var login authModel.EmployeeLogin
	if err := h.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		jsonResponse(w, []interface{}{}, http.StatusOK)
		return
	}
	var emp struct{ LocationID *int64 }
	h.db.Table("employee").Select("location_id").
		Where("employee_login_id = ?", login.IDEmployeeLogin).Scan(&emp)
	if emp.LocationID == nil {
		jsonResponse(w, []interface{}{}, http.StatusOK)
		return
	}
	locID := *emp.LocationID

	var entries []patients.RecentlyViewedPatient
	h.db.Preload("Patient").
		Where("location_id = ?", locID).
		Order("datetime_viewed desc").
		Limit(30).Find(&entries)

	result := make([]map[string]interface{}, 0, len(entries))
	for _, e := range entries {
		if e.Patient == nil {
			continue
		}
		p := e.Patient
		dob := ""
		if p.DOB != nil {
			dob = p.DOB.Format("2006-01-02")
		}
		viewed := ""
		if e.DatetimeViewed != nil {
			viewed = e.DatetimeViewed.Format(time.RFC3339)
		}
		result = append(result, map[string]interface{}{
			"id_patient":      p.IDPatient,
			"first_name":      p.FirstName,
			"last_name":       p.LastName,
			"dob":             dob,
			"email":           strVal(p.Email),
			"phone":           strVal(p.Phone),
			"datetime_viewed": viewed,
		})
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /express_pass
func (h *Handler) ExpressPass(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var login authModel.EmployeeLogin
	if err := h.db.Where("employee_login = ?", strings.ToUpper(username)).First(&login).Error; err != nil {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}
	if !login.Active {
		jsonError(w, "Account is inactive", http.StatusForbidden)
		return
	}

	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Password) == "" {
		jsonError(w, "Password is required", http.StatusBadRequest)
		return
	}
	if !login.CheckPassword(body.Password) {
		jsonError(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	pin, err := generateUniqueExpressLogin(h.db, &login)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, map[string]string{"express_login": pin}, http.StatusOK)
}

// POST /gift_card/new
func (h *Handler) CreateGiftCard(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())

	var login authModel.EmployeeLogin
	if err := h.db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}
	var emp struct{ LocationID *int64 }
	h.db.Table("employee").Select("location_id").
		Where("employee_login_id = ?", login.IDEmployeeLogin).Scan(&emp)
	if emp.LocationID == nil {
		jsonError(w, "Employee has no location", http.StatusNotFound)
		return
	}

	var body struct {
		Quantity       int     `json:"quantity"`
		Nominal        float64 `json:"nominal"`
		LocationID     *int    `json:"location_id"`
		ExpirationDate *string `json:"expiration_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if body.Nominal == 0 {
		jsonError(w, "Nominal value is required", http.StatusBadRequest)
		return
	}
	qty := body.Quantity
	if qty <= 0 {
		qty = 1
	}
	locID := int(*emp.LocationID)
	if body.LocationID != nil {
		locID = *body.LocationID
	}
	expDate := time.Now().UTC().AddDate(1, 0, 0)
	if body.ExpirationDate != nil {
		t, err := time.Parse(time.RFC3339, *body.ExpirationDate)
		if err == nil {
			expDate = t
		}
	}

	nominal := fmt.Sprintf("%.2f", body.Nominal)
	cards := make([]map[string]interface{}, 0, qty)
	for i := 0; i < qty; i++ {
		code := generateUniqueGiftCardCode(h.db)
		gc := marketing.GiftCard{
			Code:           code,
			Nominal:        nominal,
			Balance:        nominal,
			ExpirationDate: &expDate,
			LocationID:     locID,
			Status:         "active",
		}
		h.db.Create(&gc)
		cards = append(cards, map[string]interface{}{
			"id_gift_card":    gc.IDGiftCard,
			"code":            gc.Code,
			"nominal":         body.Nominal,
			"balance":         body.Nominal,
			"expiration_date": expDate.Format("2006-01-02"),
			"status":          gc.Status,
		})
	}
	jsonResponse(w, map[string]interface{}{
		"message":    fmt.Sprintf("%d gift card(s) created", qty),
		"gift_cards": cards,
	}, http.StatusCreated)
}

// GET /gift_card/details
func (h *Handler) GetGiftCardDetails(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	locID := getEmployeeLocationID(h.db, username)
	if locID == 0 {
		jsonError(w, "Employee or location not found", http.StatusNotFound)
		return
	}

	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		jsonError(w, "Gift card code is required", http.StatusBadRequest)
		return
	}

	var gc marketing.GiftCard
	if err := h.db.Where("location_id = ? AND code = ?", locID, code).First(&gc).Error; err != nil {
		jsonError(w, "Gift card not found", http.StatusNotFound)
		return
	}

	expDate := ""
	if gc.ExpirationDate != nil {
		expDate = gc.ExpirationDate.Format("2006-01-02")
	}
	createdAt := ""
	if gc.CreatedAt != nil {
		createdAt = gc.CreatedAt.Format(time.RFC3339)
	}

	var loc location.Location
	h.db.First(&loc, locID)

	jsonResponse(w, map[string]interface{}{
		"gift_card_id":     gc.IDGiftCard,
		"code":             gc.Code,
		"balance":          gc.Balance,
		"expiration_date":  expDate,
		"nominal":          gc.Nominal,
		"location_name":    loc.FullName,
		"created_at":       createdAt,
		"status":           gc.Status,
	}, http.StatusOK)
}

// GET /gift_card/list
func (h *Handler) ListGiftCards(w http.ResponseWriter, r *http.Request) {
	username := pkgAuth.UsernameFromContext(r.Context())
	locID := getEmployeeLocationID(h.db, username)
	if locID == 0 {
		jsonError(w, "Employee or location not found", http.StatusNotFound)
		return
	}

	includeExpired := strings.ToLower(r.URL.Query().Get("exp")) == "true"

	db := h.db.Where("location_id = ?", locID)
	if !includeExpired {
		today := time.Now().Format("2006-01-02")
		db = db.Where("expiration_date IS NULL OR expiration_date >= ?", today)
	}

	var cards []marketing.GiftCard
	db.Order("created_at desc").Find(&cards)

	var loc location.Location
	h.db.First(&loc, locID)

	result := make([]map[string]interface{}, 0, len(cards))
	for _, gc := range cards {
		expDate := ""
		if gc.ExpirationDate != nil {
			expDate = gc.ExpirationDate.Format("2006-01-02")
		}
		createdAt := ""
		if gc.CreatedAt != nil {
			createdAt = gc.CreatedAt.Format(time.RFC3339)
		}
		result = append(result, map[string]interface{}{
			"gift_card_id":    gc.IDGiftCard,
			"code":            gc.Code,
			"balance":         gc.Balance,
			"expiration_date": expDate,
			"nominal":         gc.Nominal,
			"location_name":   loc.FullName,
			"created_at":      createdAt,
			"status":          gc.Status,
		})
	}
	jsonResponse(w, result, http.StatusOK)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func (h *Handler) countVendorTermsDue(loc *location.Location) int {
	const windowDays = 5
	today := time.Now()
	cutoff := today.AddDate(0, 0, windowDays)

	var count int64
	h.db.Table("vendor_ap_invoice vai").
		Joins("JOIN vendor_location_account vla ON vla.id_vendor_location_account = vai.vendor_location_account_id").
		Where("vla.location_id = ? AND vai.due_date <= ? AND vai.status = 'open'",
			loc.IDLocation, cutoff.Format("2006-01-02")).
		Count(&count)
	return int(count)
}

func (h *Handler) countTasksTodoAssigned(employeeID int) int {
	// table "task" does not exist yet — return 0
	return 0
}

func (h *Handler) countRequestAppointmentsUnprocessed(locationID int) int {
	// table "appointment_request" does not exist yet — return 0
	return 0
}

func generateUniqueExpressLogin(db *gorm.DB, login *authModel.EmployeeLogin) (string, error) {
	for i := 0; i < 200; i++ {
		pin := fmt.Sprintf("%05d", rand.Intn(100000))
		if err := login.SetExpressLogin(pin); err != nil {
			continue
		}
		res := db.Model(login).Updates(map[string]interface{}{
			"express_login":        login.ExpressLogin,
			"express_login_digest": login.ExpressLoginDigest,
		})
		if res.Error == nil {
			return pin, nil
		}
	}
	return "", fmt.Errorf("could not generate unique PIN (try again)")
}

func generateUniqueGiftCardCode(db *gorm.DB) string {
	for {
		code := strconv.Itoa(100_000_000 + rand.Intn(900_000_000))
		var count int64
		db.Model(&marketing.GiftCard{}).Where("code = ?", code).Count(&count)
		if count == 0 {
			return code
		}
	}
}

func getEmployeeLocationID(db *gorm.DB, username string) int {
	var login authModel.EmployeeLogin
	if err := db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return 0
	}
	var emp struct{ LocationID *int64 }
	db.Table("employee").Select("location_id").
		Where("employee_login_id = ?", login.IDEmployeeLogin).Scan(&emp)
	if emp.LocationID == nil {
		return 0
	}
	return int(*emp.LocationID)
}

func normalizePath(path string) string {
	if len(path) == 0 {
		return "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}
	if path != "/" {
		for len(path) > 1 && path[len(path)-1] == '/' {
			path = path[:len(path)-1]
		}
	}
	return path
}

func isDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func derefInt64(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func jsonResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
