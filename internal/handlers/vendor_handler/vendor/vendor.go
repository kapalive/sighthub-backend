package vendor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	svc "sighthub-backend/internal/services/vendor_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct {
	svc *svc.Service
}

func New(s *svc.Service) *Handler { return &Handler{svc: s} }

// POST /add
func (h *Handler) AddVendor(w http.ResponseWriter, r *http.Request) {
	var req svc.AddVendorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	id, err := h.svc.AddVendor(req)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.VwAccount != nil {
		var locID int64
		username := pkgAuth.UsernameFromContext(r.Context())
		if username != "" {
			h.svc.DB().Table("employee e").
				Select("e.location_id").
				Joins("JOIN employee_login el ON el.id_employee_login = e.employee_login_id").
				Where("el.employee_login = ?", username).Scan(&locID)
		}
		if locID > 0 {
			h.svc.UpsertVWAccount(id, locID, req.VwAccount)
		}
	}

	jsonResponse(w, map[string]interface{}{"message": "Vendor added successfully", "vendor_id": id}, http.StatusCreated)
}

// PUT /{vendor_id}
func (h *Handler) UpdateVendor(w http.ResponseWriter, r *http.Request) {
	vendorID, err := parseID(r, "vendor_id")
	if err != nil {
		jsonError(w, "invalid vendor_id", http.StatusBadRequest)
		return
	}
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var rep *svc.RepInput
	if repRaw, ok := body["rep"]; ok && repRaw != nil {
		repBytes, _ := json.Marshal(repRaw)
		rep = &svc.RepInput{}
		json.Unmarshal(repBytes, rep)
	}
	delete(body, "rep")

	var vwAccount map[string]interface{}
	if vwRaw, ok := body["vw_account"]; ok && vwRaw != nil {
		if m, ok := vwRaw.(map[string]interface{}); ok {
			vwAccount = m
		}
	}
	delete(body, "vw_account")

	if err := h.svc.UpdateVendor(vendorID, body, rep); err != nil {
		status := errorStatus(err)
		jsonError(w, err.Error(), status)
		return
	}

	if vwAccount != nil {
		var locID int64
		username := pkgAuth.UsernameFromContext(r.Context())
		if username != "" {
			h.svc.DB().Table("employee e").
				Select("e.location_id").
				Joins("JOIN employee_login el ON el.id_employee_login = e.employee_login_id").
				Where("el.employee_login = ?", username).Scan(&locID)
		}
		if locID > 0 {
			h.svc.UpsertVWAccount(vendorID, locID, vwAccount)
		}
	}

	jsonResponse(w, map[string]string{"message": "Vendor updated successfully"}, http.StatusOK)
}

// DELETE /{vendor_id}
func (h *Handler) DeleteVendor(w http.ResponseWriter, r *http.Request) {
	vendorID, err := parseID(r, "vendor_id")
	if err != nil {
		jsonError(w, "invalid vendor_id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteVendor(vendorID); err != nil {
		status := errorStatus(err)
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]string{"message": "Vendor deleted successfully"}, http.StatusOK)
}

// GET /{vendor_id}
func (h *Handler) GetVendor(w http.ResponseWriter, r *http.Request) {
	vendorID, err := parseID(r, "vendor_id")
	if err != nil {
		jsonError(w, "invalid vendor_id", http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	var locID int64
	if username != "" {
		h.svc.DB().Table("employee e").
			Select("e.location_id").
			Joins("JOIN employee_login el ON el.id_employee_login = e.employee_login_id").
			Where("el.employee_login = ?", username).
			Scan(&locID)
	}
	result, err := h.svc.GetVendor(vendorID, locID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /list
func (h *Handler) ListVendors(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	includeDetails := r.URL.Query().Get("include_details") == "true"

	result, err := h.svc.ListVendors(page, includeDetails)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /{vendor_id}/invoices
func (h *Handler) GetVendorInvoices(w http.ResponseWriter, r *http.Request) {
	vendorID, err := parseID(r, "vendor_id")
	if err != nil {
		jsonError(w, "invalid vendor_id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GetVendorInvoices(vendorID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /{vendor_id}/agreement
func (h *Handler) CreateAgreement(w http.ResponseWriter, r *http.Request) {
	vendorID, err := parseID(r, "vendor_id")
	if err != nil {
		jsonError(w, "invalid vendor_id", http.StatusBadRequest)
		return
	}
	var input svc.AgreementInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	result, err := h.svc.CreateAgreement(vendorID, input)
	if err != nil {
		status := errorStatus(err)
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]interface{}{
		"message":   "Agreement was created successfully",
		"agreement": result,
	}, http.StatusCreated)
}

// PUT /{vendor_id}/agreement/{agreement_id}
func (h *Handler) UpdateAgreement(w http.ResponseWriter, r *http.Request) {
	vendorID, err := parseID(r, "vendor_id")
	if err != nil {
		jsonError(w, "invalid vendor_id", http.StatusBadRequest)
		return
	}
	agrID, err := parseID(r, "agreement_id")
	if err != nil {
		jsonError(w, "invalid agreement_id", http.StatusBadRequest)
		return
	}
	var input svc.AgreementInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	result, err := h.svc.UpdateAgreement(vendorID, agrID, input)
	if err != nil {
		status := errorStatus(err)
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]interface{}{
		"message":   "Agreement updated successfully",
		"agreement": result,
	}, http.StatusOK)
}

// DELETE /{vendor_id}/agreement/{agreement_id}
func (h *Handler) DeleteAgreement(w http.ResponseWriter, r *http.Request) {
	vendorID, err := parseID(r, "vendor_id")
	if err != nil {
		jsonError(w, "invalid vendor_id", http.StatusBadRequest)
		return
	}
	agrID, err := parseID(r, "agreement_id")
	if err != nil {
		jsonError(w, "invalid agreement_id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteAgreement(vendorID, agrID); err != nil {
		status := errorStatus(err)
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]interface{}{
		"message":      "Agreement deleted successfully",
		"agreement_id": agrID,
	}, http.StatusOK)
}

// POST /{vendor_id}/add_brand
func (h *Handler) AddVendorBrand(w http.ResponseWriter, r *http.Request) {
	vendorID, err := parseID(r, "vendor_id")
	if err != nil {
		jsonError(w, "invalid vendor_id", http.StatusBadRequest)
		return
	}
	_ = vendorID // vendorID passed via URL
	var input svc.AddBrandInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	result, err := h.svc.AddVendorBrand(vendorID, input)
	if err != nil {
		status := errorStatus(err)
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, result, http.StatusCreated)
}

// PUT /update_brand/{brand_id}
func (h *Handler) UpdateVendorBrand(w http.ResponseWriter, r *http.Request) {
	brandID, err := parseID(r, "brand_id")
	if err != nil {
		jsonError(w, "invalid brand_id", http.StatusBadRequest)
		return
	}
	var input svc.UpdateBrandInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateVendorBrand(brandID, input); err != nil {
		status := errorStatus(err)
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]interface{}{
		"message":  fmt.Sprintf("%s brand updated successfully", input.KeyBrand),
		"brand_id": brandID,
	}, http.StatusOK)
}

// DELETE /{vendor_id}/brand/{brand_type}/{brand_id}
func (h *Handler) DeleteVendorBrand(w http.ResponseWriter, r *http.Request) {
	vendorID, err := parseID(r, "vendor_id")
	if err != nil {
		jsonError(w, "invalid vendor_id", http.StatusBadRequest)
		return
	}
	brandType := mux.Vars(r)["brand_type"]
	brandID, err := parseID(r, "brand_id")
	if err != nil {
		jsonError(w, "invalid brand_id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteVendorBrand(vendorID, brandType, brandID); err != nil {
		status := errorStatus(err)
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]string{
		"message": fmt.Sprintf("%s brand link deleted for vendor %d", capitalize(brandType), vendorID),
	}, http.StatusOK)
}

// GET /lab/list
func (h *Handler) ListLabs(w http.ResponseWriter, r *http.Request) {
	var locationID *int64
	username := pkgAuth.UsernameFromContext(r.Context())
	if username != "" {
		var locID int64
		h.svc.DB().Table("employee e").
			Select("e.location_id").
			Joins("JOIN employee_login el ON el.id_employee_login = e.employee_login_id").
			Where("el.employee_login = ?", username).
			Scan(&locID)
		if locID > 0 {
			locationID = &locID
		}
	}
	result, err := h.svc.ListLabs(locationID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /lab
func (h *Handler) CreateLab(w http.ResponseWriter, r *http.Request) {
	var input svc.LabInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	username := pkgAuth.UsernameFromContext(r.Context())
	if username != "" {
		var locID int64
		h.svc.DB().Table("employee e").
			Select("e.location_id").
			Joins("JOIN employee_login el ON el.id_employee_login = e.employee_login_id").
			Where("el.employee_login = ?", username).
			Scan(&locID)
		input.LocationID = locID
	}
	result, err := h.svc.CreateLab(input)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, result, http.StatusCreated)
}

// PUT /lab/{id_lab}
func (h *Handler) UpdateLab(w http.ResponseWriter, r *http.Request) {
	labID, err := parseID(r, "id_lab")
	if err != nil {
		jsonError(w, "invalid id_lab", http.StatusBadRequest)
		return
	}
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	result, err := h.svc.UpdateLab(labID, data)
	if err != nil {
		status := errorStatus(err)
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /lab/{id_lab}
func (h *Handler) GetLab(w http.ResponseWriter, r *http.Request) {
	labID, err := parseID(r, "id_lab")
	if err != nil {
		jsonError(w, "invalid id_lab", http.StatusBadRequest)
		return
	}
	var locationID *int64
	username := pkgAuth.UsernameFromContext(r.Context())
	if username != "" {
		var locID int64
		h.svc.DB().Table("employee e").
			Select("e.location_id").
			Joins("JOIN employee_login el ON el.id_employee_login = e.employee_login_id").
			Where("el.employee_login = ?", username).
			Scan(&locID)
		if locID > 0 {
			locationID = &locID
		}
	}
	result, err := h.svc.GetLab(labID, locationID)
	if err != nil {
		status := errorStatus(err)
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// DELETE /lab/{id_lab}
func (h *Handler) DeleteLab(w http.ResponseWriter, r *http.Request) {
	labID, err := parseID(r, "id_lab")
	if err != nil {
		jsonError(w, "invalid id_lab", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteLab(labID); err != nil {
		status := errorStatus(err)
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]string{"message": "Lab deleted successfully"}, http.StatusOK)
}

// PUT /{vendor_id}/add_lab
func (h *Handler) AddVendorLab(w http.ResponseWriter, r *http.Request) {
	vendorID, err := parseID(r, "vendor_id")
	if err != nil {
		jsonError(w, "invalid vendor_id", http.StatusBadRequest)
		return
	}
	var body struct {
		LabID *int `json:"lab_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.LabID == nil {
		jsonError(w, "lab_id is required", http.StatusBadRequest)
		return
	}
	err = h.svc.AddVendorLab(vendorID, *body.LabID)
	if err != nil {
		switch err.(type) {
		case *svc.AlreadyExistsError:
			jsonResponse(w, map[string]string{"message": err.Error()}, http.StatusOK)
			return
		default:
			status := errorStatus(err)
			jsonError(w, err.Error(), status)
			return
		}
	}
	jsonResponse(w, map[string]string{"message": "Lab added to vendor successfully"}, http.StatusCreated)
}

// DELETE /{vendor_id}/remove_lab/{lab_id}
func (h *Handler) RemoveVendorLab(w http.ResponseWriter, r *http.Request) {
	vendorID, err := parseID(r, "vendor_id")
	if err != nil {
		jsonError(w, "invalid vendor_id", http.StatusBadRequest)
		return
	}
	labID, err := parseID(r, "lab_id")
	if err != nil {
		jsonError(w, "invalid lab_id", http.StatusBadRequest)
		return
	}
	if err := h.svc.RemoveVendorLab(vendorID, labID); err != nil {
		status := errorStatus(err)
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]string{"message": "Lab removed from vendor successfully"}, http.StatusOK)
}

// GET /countries
func (h *Handler) GetCountries(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetCountries()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// GET /states/{country_id}
func (h *Handler) GetStatesByCountry(w http.ResponseWriter, r *http.Request) {
	countryID, err := parseID(r, "country_id")
	if err != nil {
		jsonError(w, "invalid country_id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GetStatesByCountry(countryID)
	if err != nil {
		status := errorStatus(err)
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /{vendor_id}/pricing-rules/{brand_type}/{brand_id}
func (h *Handler) AddPricingRule(w http.ResponseWriter, r *http.Request) {
	vendorID, err := parseID(r, "vendor_id")
	if err != nil {
		jsonError(w, "invalid vendor_id", http.StatusBadRequest)
		return
	}
	brandType := mux.Vars(r)["brand_type"]
	brandID, err := parseID(r, "brand_id")
	if err != nil {
		jsonError(w, "invalid brand_id", http.StatusBadRequest)
		return
	}
	var input svc.PricingRuleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	result, err := h.svc.AddPricingRule(vendorID, brandType, brandID, input)
	if err != nil {
		status := errorStatus(err)
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]interface{}{
		"message": "Pricing rule created",
		"rule":    result,
	}, http.StatusCreated)
}

// GET /{vendor_id}/pricing-rules/{brand_type}/{brand_id}
func (h *Handler) GetPricingRules(w http.ResponseWriter, r *http.Request) {
	brandType := mux.Vars(r)["brand_type"]
	brandID, err := parseID(r, "brand_id")
	if err != nil {
		jsonError(w, "invalid brand_id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GetPricingRules(brandType, brandID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// PUT /{vendor_id}/pricing-rules/{rule_id}
func (h *Handler) UpdatePricingRule(w http.ResponseWriter, r *http.Request) {
	vendorID, err := parseID(r, "vendor_id")
	if err != nil {
		jsonError(w, "invalid vendor_id", http.StatusBadRequest)
		return
	}
	ruleID, err := parseID(r, "rule_id")
	if err != nil {
		jsonError(w, "invalid rule_id", http.StatusBadRequest)
		return
	}
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	result, err := h.svc.UpdatePricingRule(vendorID, ruleID, data)
	if err != nil {
		status := errorStatus(err)
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]interface{}{
		"message": "Pricing rule updated",
		"rule":    result,
	}, http.StatusOK)
}

// DELETE /{vendor_id}/pricing-rules/{rule_id}
func (h *Handler) DeletePricingRule(w http.ResponseWriter, r *http.Request) {
	vendorID, err := parseID(r, "vendor_id")
	if err != nil {
		jsonError(w, "invalid vendor_id", http.StatusBadRequest)
		return
	}
	ruleID, err := parseID(r, "rule_id")
	if err != nil {
		jsonError(w, "invalid rule_id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeletePricingRule(vendorID, ruleID); err != nil {
		status := errorStatus(err)
		jsonError(w, err.Error(), status)
		return
	}
	jsonResponse(w, map[string]string{"message": "Pricing rule deleted"}, http.StatusOK)
}

// --- helpers ---

func parseID(r *http.Request, key string) (int, error) {
	return strconv.Atoi(mux.Vars(r)[key])
}

func jsonResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data) //nolint:errcheck
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}

func errorStatus(err error) int {
	switch err.(type) {
	case *svc.NotFoundError:
		return http.StatusNotFound
	case *svc.ConflictError:
		return http.StatusConflict
	case *svc.AlreadyExistsError:
		return http.StatusOK
	default:
		return http.StatusBadRequest
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

