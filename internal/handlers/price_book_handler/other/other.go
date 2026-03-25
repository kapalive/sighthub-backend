package other

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"sighthub-backend/internal/services/price_book_service"
)

type Handler struct {
	svc *price_book_service.Service
}

func New(svc *price_book_service.Service) *Handler {
	return &Handler{svc: svc}
}

// ═══════════════════════ TREATMENTS ═══════════════════════════════════════════

// GET /api/price-book/treatments/vendors
func (h *Handler) GetTreatmentVendors(w http.ResponseWriter, r *http.Request) {
	results, err := h.svc.GetTreatmentVendors()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/price-book/treatments?vendor_id=X
func (h *Handler) GetTreatments(w http.ResponseWriter, r *http.Request) {
	var vendorID *int
	if v := r.URL.Query().Get("vendor_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil { vendorID = &id }
	}
	var source *string
	if v := r.URL.Query().Get("source"); v != "" {
		source = &v
	}
	results, err := h.svc.GetTreatments(vendorID, source)
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/price-book/treatments/{id}
func (h *Handler) GetTreatment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GetTreatment(id)
	if err != nil {
		jsonError(w, "Lens treatment not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /api/price-book/treatments
func (h *Handler) CreateTreatment(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	itemNbr, _ := body["item_nbr"].(string)
	vendorIDf, _ := body["vendor_id"].(float64)
	pricef, _ := body["price"].(float64)
	if itemNbr == "" || vendorIDf == 0 || pricef == 0 {
		jsonError(w, "Required: item_nbr, vendor_id, price", http.StatusBadRequest)
		return
	}

	in := price_book_service.CreateTreatmentInput{
		ItemNbr:   itemNbr,
		VendorID:  int(vendorIDf),
		Price:     pricef,
		CanLookup: true,
	}
	if v, ok := body["description"].(string); ok { in.Description = &v }
	if v, ok := body["cost"].(float64); ok { in.Cost = &v }
	if v, ok := body["v_codes_lens_id"].(float64); ok { vv := int(v); in.VCodesLensID = &vv }
	if v, ok := body["can_lookup"].(bool); ok { in.CanLookup = v }
	if arr, ok := body["special_features"].([]interface{}); ok {
		for _, v := range arr {
			if f, ok := v.(float64); ok {
				in.SpecialFeatures = append(in.SpecialFeatures, int(f))
			}
		}
	}

	id, err := h.svc.CreateTreatment(in)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]interface{}{"message": "Lens treatment created successfully", "id_lens_treatments": id}, http.StatusCreated)
}

// PUT /api/price-book/treatments/{id}
func (h *Handler) UpdateTreatment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid id", http.StatusBadRequest)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	in := price_book_service.UpdateTreatmentInput{}
	if v, ok := body["item_nbr"].(string); ok && v != "" { in.ItemNbr = &v }
	if v, ok := body["description"].(string); ok && v != "" { in.Description = &v }
	if v, ok := body["vendor_id"].(float64); ok { vv := int(v); in.VendorID = &vv }
	if _, ok := body["v_codes_lens_id"]; ok {
		switch val := body["v_codes_lens_id"].(type) {
		case float64:
			vv := int(val); in.VCodesLensID = &vv
		case nil:
			vv := -1; in.VCodesLensID = &vv
		}
	}
	if v, ok := body["price"].(float64); ok { in.Price = &v }
	if _, ok := body["cost"]; ok {
		switch val := body["cost"].(type) {
		case float64:
			in.Cost = &val
		case string:
			if val == "" { in.ClearCost = true } else {
				var f float64
				if _, err := strconv.ParseFloat(val, 64); err == nil {
					f, _ = strconv.ParseFloat(val, 64)
					in.Cost = &f
				}
			}
		case nil:
			in.ClearCost = true
		}
	}
	if v, ok := body["can_lookup"].(bool); ok { in.CanLookup = &v }
	if arr, ok := body["special_features"].([]interface{}); ok {
		sfs := make([]int, 0, len(arr))
		for _, v := range arr {
			if f, ok := v.(float64); ok {
				sfs = append(sfs, int(f))
			}
		}
		in.SpecialFeatures = &sfs
	}
	if arr, ok := body["v_codes"].([]interface{}); ok {
		vcs := make([]int, 0, len(arr))
		for _, v := range arr {
			if f, ok := v.(float64); ok {
				vcs = append(vcs, int(f))
			}
		}
		in.VCodes = &vcs
	}

	if err := h.svc.UpdateTreatment(id, in); err != nil {
		code := http.StatusBadRequest
		if err.Error() == "lens treatment not found" { code = http.StatusNotFound }
		jsonError(w, err.Error(), code)
		return
	}
	jsonResponse(w, map[string]string{"message": "Lens treatment updated successfully"}, http.StatusOK)
}

// DELETE /api/price-book/treatments/{id}
func (h *Handler) DeleteTreatment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteTreatment(id); err != nil {
		code := http.StatusBadRequest
		if err.Error() == "lens treatment not found" { code = http.StatusNotFound }
		jsonError(w, err.Error(), code)
		return
	}
	jsonResponse(w, map[string]string{"message": "Lens treatment deleted successfully"}, http.StatusOK)
}

// ═══════════════════════ PROFESSIONAL SERVICES ════════════════════════════════

// GET /api/price-book/professional_service_types
func (h *Handler) GetProfServiceTypes(w http.ResponseWriter, r *http.Request) {
	results, err := h.svc.GetProfServiceTypes()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/price-book/professional_service_scopes
func (h *Handler) GetProfServiceScopes(w http.ResponseWriter, r *http.Request) {
	results, err := h.svc.GetProfServiceScopes()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/price-book/professional_services?type_id=X
func (h *Handler) GetProfServices(w http.ResponseWriter, r *http.Request) {
	var typeID *int
	if v := r.URL.Query().Get("type_id"); v != "" && v != "null" {
		if id, err := strconv.Atoi(v); err == nil { typeID = &id }
	}
	results, err := h.svc.GetProfServices(typeID)
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/price-book/professional_service/{id}
func (h *Handler) GetProfService(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GetProfService(id)
	if err != nil {
		jsonError(w, "Professional service not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /api/price-book/professional_service
func (h *Handler) AddProfService(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	itemNumber, _ := body["item_number"].(string)
	typeIDf, _ := body["type_id"].(float64)
	if itemNumber == "" || typeIDf == 0 {
		jsonError(w, "Item number and professional service type ID are required", http.StatusBadRequest)
		return
	}

	in := price_book_service.AddProfServiceInput{
		ItemNumber:         itemNumber,
		TypeID:             int(typeIDf),
		ReferringPhysician: false,
		Visible:            true,
	}
	if v, ok := body["cpt_hcpcs_code"].(string); ok { in.CptHcpcsCode = &v }
	if v, ok := body["scope_id"].(float64); ok { vv := int(v); in.ScopeID = &vv }
	if v, ok := body["description"].(string); ok { in.Description = &v }
	if v, ok := body["price"].(float64); ok { in.Price = v }
	if v, ok := body["cost"].(float64); ok { in.Cost = v }
	if v, ok := body["sort1"].(float64); ok { in.Sort1 = &v }
	if v, ok := body["sort2"].(float64); ok { in.Sort2 = &v }
	if v, ok := body["referring_physician"].(bool); ok { in.ReferringPhysician = v }
	if v, ok := body["visible"].(bool); ok { in.Visible = v }
	if v, ok := body["mfr_number"].(string); ok { in.MfrNumber = &v }

	id, err := h.svc.AddProfService(in)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]interface{}{"message": "Professional service added successfully", "service_id": id}, http.StatusCreated)
}

// PUT /api/price-book/professional_service/{id}
func (h *Handler) UpdateProfService(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid id", http.StatusBadRequest)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	in := price_book_service.UpdateProfServiceInput{}
	if v, ok := body["item_number"].(string); ok { in.ItemNumber = &v }
	if v, ok := body["cpt_hcpcs_code"].(string); ok { in.CptHcpcsCode = &v }
	if v, ok := body["type_id"].(float64); ok { vv := int(v); in.TypeID = &vv }
	if _, ok := body["scope_id"]; ok {
		switch val := body["scope_id"].(type) {
		case float64:
			vv := int(val); in.ScopeID = &vv
		case string:
			if val == "" { vv := -1; in.ScopeID = &vv }
		case nil:
			vv := -1; in.ScopeID = &vv
		}
	}
	if v, ok := body["description"].(string); ok { in.Description = &v }
	if v, ok := body["price"].(float64); ok { in.Price = &v }
	if v, ok := body["cost"].(float64); ok { in.Cost = &v }
	if v, ok := body["sort1"].(float64); ok { in.Sort1 = &v }
	if v, ok := body["sort2"].(float64); ok { in.Sort2 = &v }
	if v, ok := body["referring_physician"].(bool); ok { in.ReferringPhysician = &v }
	if v, ok := body["visible"].(bool); ok { in.Visible = &v }
	if v, ok := body["mfr_number"].(string); ok { in.MfrNumber = &v }

	if err := h.svc.UpdateProfService(id, in); err != nil {
		code := http.StatusBadRequest
		if err.Error() == "professional service not found" { code = http.StatusNotFound }
		jsonError(w, err.Error(), code)
		return
	}
	jsonResponse(w, map[string]string{"message": "Professional service updated successfully"}, http.StatusOK)
}

// DELETE /api/price-book/professional_service/{id}
func (h *Handler) DeleteProfService(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteProfService(id); err != nil {
		code := http.StatusBadRequest
		if err.Error() == "professional service not found" { code = http.StatusNotFound }
		jsonError(w, err.Error(), code)
		return
	}
	jsonResponse(w, map[string]string{"message": "Professional service deleted successfully"}, http.StatusOK)
}

// ═══════════════════════ ADDITIONAL SERVICES ══════════════════════════════════

// GET /api/price-book/add_types
func (h *Handler) GetAddTypes(w http.ResponseWriter, r *http.Request) {
	results, err := h.svc.GetAddTypes()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/price-book/additional_services?type_id=X
func (h *Handler) GetAdditionalServices(w http.ResponseWriter, r *http.Request) {
	var typeID *int
	if v := r.URL.Query().Get("type_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil { typeID = &id }
	}
	results, err := h.svc.GetAdditionalServices(typeID)
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/price-book/additional_service/{id}
func (h *Handler) GetAdditionalService(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GetAdditionalService(id)
	if err != nil {
		jsonError(w, "Additional service not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /api/price-book/additional_service
func (h *Handler) AddAdditionalService(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	itemNumber, _ := body["item_number"].(string)
	typeIDf, _ := body["type_id"].(float64)
	invoiceDesc, _ := body["invoice_desc"].(string)
	if itemNumber == "" || typeIDf == 0 || invoiceDesc == "" {
		jsonError(w, "Item number, service type, and invoice description are required", http.StatusBadRequest)
		return
	}

	in := price_book_service.AddServiceInput{
		ItemNumber:  itemNumber,
		TypeID:      int(typeIDf),
		InvoiceDesc: invoiceDesc,
		Visible:     true,
	}
	if v, ok := body["price"].(float64); ok { in.Price = v }
	if v, ok := body["cost_price"].(float64); ok { in.CostPrice = v }
	toBoolP := func(key string) *bool {
		switch val := body[key].(type) {
		case bool: return &val
		case string:
			b := val == "true" || val == "1" || val == "yes"
			return &b
		}
		return nil
	}
	in.SrCost = toBoolP("sr_cost")
	in.UV = toBoolP("uv")
	in.AR = toBoolP("ar")
	in.Tint = toBoolP("tint")
	in.Drill = toBoolP("drill")
	in.Send = toBoolP("send")
	in.ReportOmit = toBoolP("report_omit")
	if v, ok := body["ins_v_code"].(string); ok { in.InsVCode = &v }
	if v, ok := body["class_level"].(string); ok { in.ClassLevel = &v }
	if v, ok := body["ins_v_code_add"].(string); ok { in.InsVCodeAdd = &v }
	if v, ok := body["sort1"].(float64); ok { in.Sort1 = &v }
	if v, ok := body["sort2"].(float64); ok { in.Sort2 = &v }
	if v, ok := body["visible"].(bool); ok { in.Visible = v }
	if v, ok := body["mfr_number"].(string); ok { in.MfrNumber = &v }
	in.Photochromatic = toBoolP("photochromatic")
	in.Polarized = toBoolP("polarized")
	in.CanDrill = toBoolP("can_drill")
	in.HighIndex = toBoolP("high_index")
	in.Digital = toBoolP("digital")

	id, err := h.svc.AddAdditionalService(in)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]interface{}{"message": "Additional service added successfully", "id_additional_service": id}, http.StatusCreated)
}

// PUT /api/price-book/additional_service/{id}
func (h *Handler) UpdateAdditionalService(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid id", http.StatusBadRequest)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	in := price_book_service.UpdateAddServiceInput{}
	if v, ok := body["item_number"].(string); ok { in.ItemNumber = &v }
	if v, ok := body["type_id"].(float64); ok { vv := int(v); in.TypeID = &vv }
	if v, ok := body["invoice_desc"].(string); ok { in.InvoiceDesc = &v }
	if v, ok := body["price"].(float64); ok { in.Price = &v }
	if v, ok := body["cost_price"].(float64); ok { in.CostPrice = &v }
	if v, ok := body["sr_cost"].(bool); ok { in.SrCost = &v }
	if v, ok := body["uv"].(bool); ok { in.UV = &v }
	if v, ok := body["ar"].(bool); ok { in.AR = &v }
	if v, ok := body["tint"].(bool); ok { in.Tint = &v }
	if v, ok := body["drill"].(bool); ok { in.Drill = &v }
	if v, ok := body["send"].(bool); ok { in.Send = &v }
	if v, ok := body["report_omit"].(bool); ok { in.ReportOmit = &v }
	if v, ok := body["ins_v_code"].(string); ok { in.InsVCode = &v }
	if v, ok := body["class_level"].(string); ok { in.ClassLevel = &v }
	if v, ok := body["ins_v_code_add"].(string); ok { in.InsVCodeAdd = &v }
	if v, ok := body["sort1"].(float64); ok { in.Sort1 = &v }
	if v, ok := body["sort2"].(float64); ok { in.Sort2 = &v }
	if v, ok := body["visible"].(bool); ok { in.Visible = &v }
	if v, ok := body["mfr_number"].(string); ok { in.MfrNumber = &v }
	if v, ok := body["photochromatic"].(bool); ok { in.Photochromatic = &v }
	if v, ok := body["polarized"].(bool); ok { in.Polarized = &v }
	if v, ok := body["can_drill"].(bool); ok { in.CanDrill = &v }
	if v, ok := body["high_index"].(bool); ok { in.HighIndex = &v }
	if v, ok := body["digital"].(bool); ok { in.Digital = &v }

	if err := h.svc.UpdateAdditionalService(id, in); err != nil {
		code := http.StatusBadRequest
		if err.Error() == "additional service not found" { code = http.StatusNotFound }
		jsonError(w, err.Error(), code)
		return
	}
	jsonResponse(w, map[string]string{"message": "Additional service updated successfully"}, http.StatusOK)
}

// DELETE /api/price-book/additional_service/{id}
func (h *Handler) DeleteAdditionalService(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteAdditionalService(id); err != nil {
		code := http.StatusBadRequest
		if err.Error() == "additional service not found" { code = http.StatusNotFound }
		jsonError(w, err.Error(), code)
		return
	}
	jsonResponse(w, map[string]string{"message": "Additional service deleted successfully"}, http.StatusOK)
}

// ═══════════════════════ MISC ITEMS ═══════════════════════════════════════════

// GET /api/price-book/misc_items
func (h *Handler) GetMiscItems(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := price_book_service.MiscListFilters{}
	if v := q.Get("pb_key"); v != "" { f.PbKey = &v }
	if v := q.Get("q"); v != "" { f.Q = &v }
	if v := q.Get("active"); v != "" {
		b := v == "1" || v == "true"
		f.Active = &b
	}
	if v := q.Get("visible"); v != "" && (v == "1" || v == "true") {
		t := true
		f.Active = &t
		f.LookupOnly = true
	} else if v := q.Get("lookup_only"); v == "1" || v == "true" {
		f.LookupOnly = true
	}

	results, err := h.svc.GetMiscItems(f)
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/price-book/misc_items/{id}
func (h *Handler) GetMiscItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GetMiscItem(id)
	if err != nil {
		jsonError(w, "Misc item not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /api/price-book/misc_items
func (h *Handler) AddMiscItem(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		body = map[string]interface{}{}
	}

	itemNumber, _ := body["item_number"].(string)
	description, _ := body["description"].(string)
	if itemNumber == "" {
		jsonError(w, "item_number is required", http.StatusBadRequest)
		return
	}
	if description == "" {
		jsonError(w, "description is required", http.StatusBadRequest)
		return
	}

	in := price_book_service.AddMiscItemInput{
		ItemNumber:  itemNumber,
		Description: description,
	}
	if v, ok := body["sale_key"].(string); ok && v != "" { in.SaleKey = &v }
	if v, ok := body["cost"].(string); ok { in.Cost = &v }
	if v, ok := body["price"].(string); ok { in.Price = &v }
	if v, ok := body["can_lookup"].(bool); ok { in.CanLookup = &v }
	if v, ok := body["active"].(bool); ok { in.Active = &v }

	result, err := h.svc.AddMiscItem(in)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, map[string]interface{}{"message": "Misc item created", "item": result}, http.StatusCreated)
}

// PUT /api/price-book/misc_items/{id}
func (h *Handler) UpdateMiscItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid id", http.StatusBadRequest)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if _, ok := body["pb_key"]; ok {
		jsonError(w, "pb_key_cannot_be_updated", http.StatusBadRequest)
		return
	}

	in := price_book_service.UpdateMiscItemInput{}
	if v, ok := body["item_number"].(string); ok { in.ItemNumber = &v }
	if v, ok := body["description"].(string); ok { in.Description = &v }
	if v, ok := body["price"].(string); ok { in.Price = &v }
	if v, ok := body["cost"].(string); ok { in.Cost = &v }
	if v, ok := body["can_lookup"].(bool); ok { in.CanLookup = &v }
	if v, ok := body["active"].(bool); ok { in.Active = &v }
	if _, ok := body["sale_key"]; ok {
		in.SetSaleKey = true
		if v, ok := body["sale_key"].(string); ok { in.SaleKey = &v }
	}

	result, err := h.svc.UpdateMiscItem(id, in)
	if err != nil {
		code := http.StatusBadRequest
		if err.Error() == "misc item not found" { code = http.StatusNotFound }
		jsonError(w, err.Error(), code)
		return
	}
	jsonResponse(w, map[string]interface{}{"message": "Misc item updated", "item": result}, http.StatusOK)
}

// DELETE /api/price-book/misc_items/{id}
func (h *Handler) DeleteMiscItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteMiscItem(id); err != nil {
		code := http.StatusBadRequest
		if err.Error() == "misc item not found" { code = http.StatusNotFound }
		if err.Error() == "cannot delete: referenced by other records" { code = http.StatusConflict }
		jsonError(w, err.Error(), code)
		return
	}
	jsonResponse(w, map[string]interface{}{"message": "Misc item deleted", "id_misc_item": id}, http.StatusOK)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

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
