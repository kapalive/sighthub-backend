package lens

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

// GET /api/price-book/lens/brand_vendor (legacy — returns all)
func (h *Handler) GetLensBrandsVendors(w http.ResponseWriter, r *http.Request) {
	results, err := h.svc.GetLensBrandsVendors()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/price_book/lens/vendors
func (h *Handler) GetLensVendors(w http.ResponseWriter, r *http.Request) {
	results, err := h.svc.GetLensVendors()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/price_book/lens/brands/{vendor_id}
func (h *Handler) GetLensBrandsByVendor(w http.ResponseWriter, r *http.Request) {
	vendorID, err := strconv.Atoi(mux.Vars(r)["vendor_id"])
	if err != nil || vendorID == 0 {
		jsonError(w, "invalid vendor_id", http.StatusBadRequest)
		return
	}
	results, err := h.svc.GetLensBrandsByVendor(vendorID)
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/price-book/lens/type
func (h *Handler) GetLensTypes(w http.ResponseWriter, r *http.Request) {
	results, err := h.svc.GetLensTypes()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/price-book/lens/materials
func (h *Handler) GetLensMaterials(w http.ResponseWriter, r *http.Request) {
	results, err := h.svc.GetLensMaterials()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// POST /api/price-book/lens/materials
func (h *Handler) AddLensMaterial(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MaterialName string `json:"material_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.MaterialName == "" {
		jsonError(w, "Field 'material_name' is required", http.StatusBadRequest)
		return
	}
	id, err := h.svc.AddLensMaterial(body.MaterialName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]interface{}{"message": "Lens material added successfully", "id_lenses_materials": id}, http.StatusCreated)
}

// GET /api/price-book/lens/special
func (h *Handler) GetLensSpecialFeatures(w http.ResponseWriter, r *http.Request) {
	results, err := h.svc.GetLensSpecialFeatures()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// POST /api/price-book/lens/special
func (h *Handler) AddLensSpecialFeature(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FeatureName string `json:"feature_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.FeatureName == "" {
		jsonError(w, "Field 'feature_name' is required", http.StatusBadRequest)
		return
	}
	id, err := h.svc.AddLensSpecialFeature(body.FeatureName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]interface{}{"message": "Lens special feature added successfully", "id_lens_special_features": id}, http.StatusCreated)
}

// GET /api/price-book/lens/series
func (h *Handler) GetLensSeries(w http.ResponseWriter, r *http.Request) {
	results, err := h.svc.GetLensSeries()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// POST /api/price-book/lens/series
func (h *Handler) AddLensSeries(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SeriesName string `json:"series_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.SeriesName == "" {
		jsonError(w, "Field 'series_name' is required", http.StatusBadRequest)
		return
	}
	id, err := h.svc.AddLensSeries(body.SeriesName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]interface{}{"message": "Lens series added successfully", "id_lens_series": id}, http.StatusCreated)
}

// GET /api/price-book/lens/v_codes
func (h *Handler) GetVCodes(w http.ResponseWriter, r *http.Request) {
	results, err := h.svc.GetVCodes()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// POST /api/price-book/lens/v_codes
func (h *Handler) AddVCode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Code == "" {
		jsonError(w, "Field 'code' is required", http.StatusBadRequest)
		return
	}
	id, err := h.svc.AddVCode(body.Code)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]interface{}{"message": "V-code added successfully", "id_v_codes_lens": id}, http.StatusCreated)
}

// GET /api/price-book/lens/list
func (h *Handler) GetLensList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := price_book_service.LensFilters{}
	if v := qint(q.Get("brand_id"), q.Get("brand")); v != nil { f.BrandID = v }
	if v := qint(q.Get("vendor_id"), q.Get("vendor")); v != nil { f.VendorID = v }
	if v := qint(q.Get("type_id")); v != nil { f.TypeID = v }
	if v := qint(q.Get("material_id")); v != nil { f.MaterialID = v }
	if v := qint(q.Get("special_feature_id")); v != nil { f.SpecialFeatureID = v }
	if v := qint(q.Get("series_id")); v != nil { f.SeriesID = v }
	if v := q.Get("source"); v != "" { f.Source = &v }
	if v := q.Get("search"); v != "" { f.Search = &v }
	if v := qint(q.Get("page")); v != nil { f.Page = *v } else { f.Page = 1 }
	if v := qint(q.Get("per_page")); v != nil { f.PerPage = *v } else { f.PerPage = 25 }

	results, err := h.svc.GetLensList(f)
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/price-book/lens/{lens_id}
func (h *Handler) GetLens(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["lens_id"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid lens_id", http.StatusBadRequest)
		return
	}
	result, err := h.svc.GetLens(id)
	if err != nil {
		jsonError(w, "Lens not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, result, http.StatusOK)
}

// POST /api/price-book/lens
func (h *Handler) AddLens(w http.ResponseWriter, r *http.Request) {
	var body struct {
		LensName          string  `json:"lens_name"`
		BrandLensID       int     `json:"brand_lens_id"`
		LensTypeID        *int    `json:"lens_type_id"`
		LensesMaterialsID *int    `json:"lenses_materials_id"`
		LensSeriesID      *int    `json:"lens_series_id"`
		Description       *string `json:"description"`
		VendorID          *int    `json:"vendor_id"`
		Price             float64 `json:"price"`
		Cost              float64 `json:"cost"`
		VCodes            []int   `json:"v_codes"`
		SpecialFeatures   []int   `json:"special_features"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if body.LensName == "" || body.BrandLensID == 0 {
		jsonError(w, "Missing required fields: lens_name, brand_lens_id, price, cost", http.StatusBadRequest)
		return
	}

	id, err := h.svc.AddLens(price_book_service.AddLensInput{
		LensName:          body.LensName,
		BrandLensID:       body.BrandLensID,
		LensTypeID:        body.LensTypeID,
		LensesMaterialsID: body.LensesMaterialsID,
		LensSeriesID:      body.LensSeriesID,
		Description:       body.Description,
		VendorID:          body.VendorID,
		Price:             body.Price,
		Cost:              body.Cost,
		VCodes:            body.VCodes,
		SpecialFeatures:   body.SpecialFeatures,
	})
	if err != nil {
		code := http.StatusBadRequest
		jsonError(w, err.Error(), code)
		return
	}
	jsonResponse(w, map[string]interface{}{"message": "Lens added successfully", "id_lenses": id}, http.StatusCreated)
}

// PUT /api/price-book/lens/{lens_id}
func (h *Handler) UpdateLens(w http.ResponseWriter, r *http.Request) {
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

	in := price_book_service.UpdateLensInput{}
	if v, ok := body["lens_name"].(string); ok && v != "" { in.LensName = &v }
	if v, ok := body["brand_lens_id"].(float64); ok { vv := int(v); in.BrandLensID = &vv }
	if v, ok := body["lens_type_id"].(float64); ok { vv := int(v); in.LensTypeID = &vv }
	if v, ok := body["lenses_materials_id"].(float64); ok { vv := int(v); in.LensesMaterialsID = &vv }
	if _, ok := body["lens_series_id"]; ok {
		switch val := body["lens_series_id"].(type) {
		case float64:
			vv := int(val)
			in.LensSeriesID = &vv
		case string:
			if val == "" { vv := -1; in.LensSeriesID = &vv }
		case nil:
			vv := -1
			in.LensSeriesID = &vv
		}
	}
	if v, ok := body["description"].(string); ok { in.Description = &v }
	if _, ok := body["vendor_id"]; ok {
		switch val := body["vendor_id"].(type) {
		case float64:
			vv := int(val)
			in.VendorID = &vv
		case nil:
			vv := -1
			in.VendorID = &vv
		}
	}
	if v, ok := body["price"].(float64); ok { in.Price = &v }
	if v, ok := body["cost"].(float64); ok { in.Cost = &v }
	if v, ok := body["v_codes"].([]interface{}); ok {
		ids := make([]int, 0, len(v))
		for _, x := range v {
			if f, ok := x.(float64); ok { ids = append(ids, int(f)) }
		}
		in.VCodes = &ids
	}
	if v, ok := body["special_features"].([]interface{}); ok {
		ids := make([]int, 0, len(v))
		for _, x := range v {
			if f, ok := x.(float64); ok { ids = append(ids, int(f)) }
		}
		in.SpecialFeatures = &ids
	}

	if err := h.svc.UpdateLens(id, in); err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, map[string]string{"message": "Lens updated successfully"}, http.StatusOK)
}

// DELETE /api/price-book/lens/{lens_id}
func (h *Handler) DeleteLens(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["lens_id"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid lens_id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteLens(id); err != nil {
		code := http.StatusBadRequest
		if err.Error() == "lens not found" { code = http.StatusNotFound }
		jsonError(w, err.Error(), code)
		return
	}
	jsonResponse(w, map[string]string{"message": "Lens deleted successfully"}, http.StatusOK)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func qint(vals ...string) *int {
	for _, v := range vals {
		if v == "" || v == "null" { continue }
		if n, err := strconv.Atoi(v); err == nil {
			return &n
		}
	}
	return nil
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
