package frame_library_handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"sighthub-backend/internal/services/frame_service"
)

type Handler struct {
	svc *frame_service.Service
}

func New(svc *frame_service.Service) *Handler {
	return &Handler{svc: svc}
}

// GET /api/frame-library/search
func (h *Handler) SearchModels(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filters := frame_service.SearchFilters{}
	if v := q.Get("brand_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filters.BrandID = &id
		}
	}
	if v := q.Get("brand_name"); v != "" {
		filters.BrandName = &v
	}
	if v := q.Get("vendor_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filters.VendorID = &id
		}
	}
	if v := q.Get("vendor_name"); v != "" {
		filters.VendorName = &v
	}
	if v := q.Get("product_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filters.ProductID = &id
		}
	}
	if v := q.Get("title_product"); v != "" {
		filters.TitleProduct = &v
	}
	if v := q.Get("id_model"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filters.IDModel = &id
		}
	}
	if v := q.Get("upc"); v != "" {
		filters.UPC = &v
	}
	if v := q.Get("gtin"); v != "" {
		filters.GTIN = &v
	}

	results, err := h.svc.SearchModels(filters)
	if err != nil {
		jsonError(w, "Search failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/frame-library/vendor_brand?vendor_id=X
func (h *Handler) GetVendorBrands(w http.ResponseWriter, r *http.Request) {
	vendorID, err := strconv.Atoi(r.URL.Query().Get("vendor_id"))
	if err != nil || vendorID == 0 {
		jsonResponse(w, []interface{}{}, http.StatusBadRequest)
		return
	}

	results, err := h.svc.GetVendorBrands(vendorID)
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/frame-library/vendor-brand
func (h *Handler) GetVendorBrandCombinations(w http.ResponseWriter, r *http.Request) {
	results, err := h.svc.GetVendorBrandCombinations()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/frame-library/products?vendor_id=X&brand_id=Y
func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	var vendorID, brandID *int64
	if v := r.URL.Query().Get("vendor_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			vendorID = &id
		}
	}
	if v := r.URL.Query().Get("brand_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			brandID = &id
		}
	}

	results, err := h.svc.GetProducts(vendorID, brandID)
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/frame-library/materials_frame
func (h *Handler) GetMaterialsFrame(w http.ResponseWriter, r *http.Request) {
	materials, err := h.svc.GetFrameMaterials()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, materials, http.StatusOK)
}

// GET /api/frame-library/models_by_product/{product_id}?materials_frame=X
func (h *Handler) GetModelsByProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.Atoi(mux.Vars(r)["product_id"])
	if err != nil || productID == 0 {
		jsonError(w, "Invalid product_id", http.StatusBadRequest)
		return
	}

	var materialFilter *string
	if v := r.URL.Query().Get("materials_frame"); v != "" {
		materialFilter = &v
	}

	results, err := h.svc.GetModelsByProduct(productID, materialFilter)
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// PUT /api/frame-library/update_model/{id_model}
func (h *Handler) UpdateModel(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id_model"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid id_model", http.StatusBadRequest)
		return
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	in := frame_service.UpdateModelInput{}
	if v, ok := body["title_variant"].(string); ok {
		in.TitleVariant = &v
	}
	if v, ok := body["lens_color"].(string); ok {
		in.LensColor = &v
	}
	if v, ok := body["size_lens_width"].(string); ok {
		in.SizeLensWidth = &v
	}
	if v, ok := body["size_bridge_width"].(string); ok {
		in.SizeBridgeWidth = &v
	}
	if v, ok := body["size_temple_length"].(string); ok {
		in.SizeTempleLength = &v
	}
	if v, ok := body["sunglass"].(bool); ok {
		in.Sunglass = &v
	}
	if v, ok := body["photo"].(bool); ok {
		in.Photo = &v
	}
	if v, ok := body["polor"].(bool); ok {
		in.Polor = &v
	}
	if v, ok := body["mirror"].(bool); ok {
		in.Mirror = &v
	}
	if v, ok := body["backside_ar"].(bool); ok {
		in.BacksideAR = &v
	}
	if v, ok := body["lens_material"].(string); ok {
		in.LensMaterial = &v
	}
	if v, ok := body["upc"].(string); ok {
		in.UPC = &v
	}
	if v, ok := body["ean"].(string); ok {
		in.EAN = &v
	}
	if v, ok := body["mfg_number"].(string); ok {
		in.MfgNumber = &v
	}
	if v, ok := body["mfr_serial_number"].(string); ok {
		in.MfrSerialNumber = &v
	}
	if v, ok := body["accessories"].(string); ok {
		in.Accessories = &v
	}
	if v, ok := body["materials_frame"].(string); ok {
		in.MaterialsFrame = &v
	}
	if v, ok := body["materials_temple"].(string); ok {
		in.MaterialsTemple = &v
	}
	if v, ok := body["color"].(string); ok {
		in.Color = &v
	}
	if v, ok := body["color_template"].(string); ok {
		in.ColorTemplate = &v
	}
	if v, ok := body["shape"].(string); ok {
		in.Shape = &v
	}

	model, err := h.svc.UpdateModel(id, in)
	if err != nil {
		jsonError(w, "Model not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, model.ToMap(), http.StatusOK)
}

// POST /api/frame-library/add_model
func (h *Handler) AddProduct(w http.ResponseWriter, r *http.Request) {
	var body struct {
		VendorID     int64   `json:"vendor_id"`
		BrandID      int64   `json:"brand_id"`
		TitleProduct *string `json:"title_product"`
		Title        *string `json:"title"`
		TypeProduct  string  `json:"type_product"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	title := ""
	if body.TitleProduct != nil {
		title = *body.TitleProduct
	} else if body.Title != nil {
		title = *body.Title
	}

	if body.VendorID == 0 || body.BrandID == 0 || title == "" || body.TypeProduct == "" {
		jsonError(w, "vendor_id, brand_id, title_product, and type_product are required", http.StatusBadRequest)
		return
	}

	product, err := h.svc.AddProduct(frame_service.AddProductInput{
		VendorID:     body.VendorID,
		BrandID:      body.BrandID,
		TitleProduct: title,
		TypeProduct:  body.TypeProduct,
	})
	if err != nil {
		msg := err.Error()
		code := http.StatusBadRequest
		if strings.Contains(msg, "not found") {
			code = http.StatusNotFound
		}
		jsonError(w, msg, code)
		return
	}
	jsonResponse(w, product.ToMap(), http.StatusCreated)
}

// POST /api/frame-library/add_variant
func (h *Handler) AddVariant(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ProductID        int64   `json:"product_id"`
		Title            string  `json:"title"`
		LensColor        *string `json:"lens_color"`
		LensMaterial     *string `json:"lens_material"`
		SizeLensWidth    *string `json:"size_lens_width"`
		SizeBridgeWidth  *string `json:"size_bridge_width"`
		SizeTempleLength *string `json:"size_temple_length"`
		Sunglass         *bool   `json:"sunglass"`
		Photo            *bool   `json:"photo"`
		Polor            *bool   `json:"polor"`
		Mirror           *bool   `json:"mirror"`
		BacksideAR       *bool   `json:"backside_ar"`
		UPC              *string `json:"upc"`
		EAN              *string `json:"ean"`
		MfgNumber        *string `json:"mfg_number"`
		MfrSerialNumber  *string `json:"mfr_serial_number"`
		Accessories      *string `json:"accessories"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if body.ProductID == 0 || body.Title == "" {
		jsonError(w, "product_id and title of variant are required", http.StatusBadRequest)
		return
	}

	model, err := h.svc.AddVariant(frame_service.AddVariantInput{
		ProductID:        body.ProductID,
		TitleVariant:     body.Title,
		LensColor:        body.LensColor,
		LensMaterial:     body.LensMaterial,
		SizeLensWidth:    body.SizeLensWidth,
		SizeBridgeWidth:  body.SizeBridgeWidth,
		SizeTempleLength: body.SizeTempleLength,
		Sunglass:         body.Sunglass,
		Photo:            body.Photo,
		Polor:            body.Polor,
		Mirror:           body.Mirror,
		BacksideAR:       body.BacksideAR,
		UPC:              body.UPC,
		EAN:              body.EAN,
		MfgNumber:        body.MfgNumber,
		MfrSerialNumber:  body.MfrSerialNumber,
		Accessories:      body.Accessories,
	})
	if err != nil {
		msg := err.Error()
		code := http.StatusBadRequest
		if strings.Contains(msg, "not found") {
			code = http.StatusNotFound
		}
		jsonError(w, msg, code)
		return
	}
	jsonResponse(w, model.ToMap(), http.StatusCreated)
}

// POST /api/frame-library/custom_glasses/{id_model}
func (h *Handler) CreateCustomGlasses(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id_model"])
	if err != nil || id == 0 {
		jsonError(w, "Invalid id_model", http.StatusBadRequest)
		return
	}

	var body struct {
		TitleVariant     *string `json:"title_variant"`
		LensColor        *string `json:"lens_color"`
		LensMaterial     *string `json:"lens_material"`
		SizeLensWidth    *string `json:"size_lens_width"`
		SizeBridgeWidth  *string `json:"size_bridge_width"`
		SizeTempleLength *string `json:"size_temple_length"`
		Sunglass         *bool   `json:"sunglass"`
		Photo            *bool   `json:"photo"`
		Polor            *bool   `json:"polor"`
		Mirror           *bool   `json:"mirror"`
		BacksideAR       *bool   `json:"backside_ar"`
		UPC              *string `json:"upc"`
		Accessories      *string `json:"accessories"`
		BrandID          *int64  `json:"brand_id"`
	}
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&body)
	}

	model, err := h.svc.CreateCustomGlasses(id, frame_service.CustomGlassesInput{
		TitleVariant:     body.TitleVariant,
		LensColor:        body.LensColor,
		LensMaterial:     body.LensMaterial,
		SizeLensWidth:    body.SizeLensWidth,
		SizeBridgeWidth:  body.SizeBridgeWidth,
		SizeTempleLength: body.SizeTempleLength,
		Sunglass:         body.Sunglass,
		Photo:            body.Photo,
		Polor:            body.Polor,
		Mirror:           body.Mirror,
		BacksideAR:       body.BacksideAR,
		UPC:              body.UPC,
		Accessories:      body.Accessories,
		BrandID:          body.BrandID,
	})
	if err != nil {
		msg := err.Error()
		code := http.StatusInternalServerError
		if strings.Contains(msg, "not found") {
			code = http.StatusNotFound
		}
		jsonError(w, msg, code)
		return
	}
	jsonResponse(w, model.ToMap(), http.StatusCreated)
}

// GET /api/frame_library/frame-type-materials
func (h *Handler) GetFrameTypeMaterials(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetFrameTypeMaterials()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, data, http.StatusOK)
}

// GET /api/frame_library/shapes
func (h *Handler) GetFrameShapes(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetFrameShapes()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, data, http.StatusOK)
}

// GET /api/frame_library/lens/materials
func (h *Handler) GetLensMaterials(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.GetLensMaterials()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, data, http.StatusOK)
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
