package contact_lens

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

// GET /api/price-book/contact_lens/vendors
func (h *Handler) GetVendors(w http.ResponseWriter, r *http.Request) {
	results, err := h.svc.GetCLVendors()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/price-book/contact_lens/brands
func (h *Handler) GetBrands(w http.ResponseWriter, r *http.Request) {
	results, err := h.svc.GetCLBrands()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/price-book/contact_lens/list
func (h *Handler) GetList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := price_book_service.CLListFilters{}
	if v := q.Get("vendor_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil { f.VendorID = &id }
	}
	if v := q.Get("brand_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil { f.BrandID = &id }
	}
	if v := q.Get("can_lookup"); v != "" {
		b := v == "1" || v == "true" || v == "t" || v == "yes" || v == "y"
		if v == "0" || v == "false" || v == "f" || v == "no" || v == "n" {
			b = false
			f.CanLookup = &b
		} else if b {
			f.CanLookup = &b
		}
	}

	results, err := h.svc.GetCLList(f)
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/price-book/contact_lens/{lens_id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["lens_id"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid lens_id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GetCL(id)
	if err != nil {
		jsonError(w, "Contact lens not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /api/price-book/contact_lens
func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NameContact  string  `json:"name_contact"`
		BrandID      int     `json:"brand_id"`
		InvoiceDesc  string  `json:"invoice_desc"`
		SellingPrice float64 `json:"selling_price"`
		Cost         float64 `json:"cost"`
		VendorID     int     `json:"vendor_id"`
		InsVCode     *string `json:"ins_v_code"`
		CanLookup    bool    `json:"can_lookup"`
	}
	body.CanLookup = true // default
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if body.NameContact == "" || body.BrandID == 0 || body.InvoiceDesc == "" || body.VendorID == 0 {
		jsonError(w, "Required: name_contact, brand_id, invoice_desc, selling_price, cost, vendor_id", http.StatusBadRequest)
		return
	}

	id, err := h.svc.AddCL(price_book_service.AddCLInput{
		NameContact:  body.NameContact,
		BrandID:      body.BrandID,
		InvoiceDesc:  body.InvoiceDesc,
		SellingPrice: body.SellingPrice,
		Cost:         body.Cost,
		VendorID:     body.VendorID,
		InsVCode:     body.InsVCode,
		CanLookup:    body.CanLookup,
	})
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]interface{}{"message": "Contact lens added successfully", "id_contact_lens_item": id}, http.StatusCreated)
}

// PUT /api/price-book/contact_lens/{lens_id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["lens_id"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid lens_id", http.StatusBadRequest)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	in := price_book_service.UpdateCLInput{}
	if v, ok := body["name_contact"].(string); ok && v != "" { in.NameContact = &v }
	if v, ok := body["brand_id"].(float64); ok { vv := int(v); in.BrandID = &vv }
	if v, ok := body["invoice_desc"].(string); ok { in.InvoiceDesc = &v }
	if v, ok := body["selling_price"].(float64); ok { in.SellingPrice = &v }
	if v, ok := body["cost"].(float64); ok { in.Cost = &v }
	if v, ok := body["vendor_id"].(float64); ok { vv := int(v); in.VendorID = &vv }
	if _, ok := body["ins_v_code"]; ok {
		in.SetVCode = true
		switch val := body["ins_v_code"].(type) {
		case string:
			in.InsVCode = &val
		case nil:
			in.InsVCode = nil
		}
	}
	if v, ok := body["can_lookup"].(bool); ok { in.CanLookup = &v }

	if err := h.svc.UpdateCL(id, in); err != nil {
		code := http.StatusBadRequest
		if err.Error() == "contact lens not found" { code = http.StatusNotFound }
		jsonError(w, err.Error(), code)
		return
	}
	jsonResponse(w, map[string]string{"message": "Contact lens updated successfully"}, http.StatusOK)
}

// DELETE /api/price-book/contact_lens/{lens_id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["lens_id"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid lens_id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteCL(id); err != nil {
		code := http.StatusBadRequest
		if err.Error() == "contact lens not found" { code = http.StatusNotFound }
		jsonError(w, err.Error(), code)
		return
	}
	jsonResponse(w, map[string]string{"message": "Contact lens deleted successfully"}, http.StatusOK)
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
