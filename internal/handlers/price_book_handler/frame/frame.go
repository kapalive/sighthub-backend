package frame

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"sighthub-backend/internal/services/price_book_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

type Handler struct {
	svc *price_book_service.Service
	db  *gorm.DB
}

func New(svc *price_book_service.Service, db *gorm.DB) *Handler {
	return &Handler{svc: svc, db: db}
}

// GET /api/price-book/vendor-brand
func (h *Handler) GetVendorBrandCombinations(w http.ResponseWriter, r *http.Request) {
	results, err := h.svc.GetPBVendorBrandCombinations()
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/price-book/products?vendor_id=X&brand_id=Y
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
	results, err := h.svc.GetPBProducts(vendorID, brandID)
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// GET /api/price-book/models?product_id=X
func (h *Handler) GetModels(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(r.URL.Query().Get("product_id"), 10, 64)
	if err != nil || productID == 0 {
		jsonError(w, "product_id is required", http.StatusBadRequest)
		return
	}
	results, err := h.svc.GetPBModels(productID)
	if err != nil {
		jsonError(w, "Query failed", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, results, http.StatusOK)
}

// POST /api/price-book/custom_glasses/{inventory_id}
func (h *Handler) CreateCustomGlasses(w http.ResponseWriter, r *http.Request) {
	inventoryID, err := strconv.Atoi(mux.Vars(r)["inventory_id"])
	if err != nil || inventoryID == 0 {
		jsonError(w, "Invalid inventory_id", http.StatusBadRequest)
		return
	}
	employeeID := h.getEmployeeID(r)

	var body map[string]interface{}
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&body)
	}
	if body == nil {
		body = map[string]interface{}{}
	}

	in := price_book_service.CustomGlassesInput{}
	if v, ok := body["title_variant"].(string); ok { in.TitleVariant = &v }
	if v, ok := body["lens_color"].(string); ok { in.LensColor = &v }
	if v, ok := body["lens_material"].(string); ok { in.LensMaterial = &v }
	if v, ok := body["size_lens_width"].(string); ok { in.SizeLensWidth = &v }
	if v, ok := body["size_bridge_width"].(string); ok { in.SizeBridgeWidth = &v }
	if v, ok := body["size_temple_length"].(string); ok { in.SizeTempleLength = &v }
	if v, ok := body["sunglass"].(bool); ok { in.Sunglass = &v }
	if v, ok := body["photo"].(bool); ok { in.Photo = &v }
	if v, ok := body["polor"].(bool); ok { in.Polor = &v }
	if v, ok := body["mirror"].(bool); ok { in.Mirror = &v }
	if v, ok := body["backside_ar"].(bool); ok { in.BacksideAR = &v }
	if v, ok := body["upc"].(string); ok { in.UPC = &v }
	if v, ok := body["accessories"].(string); ok { in.Accessories = &v }
	if v, ok := body["type_product"].(string); ok {
		in.TypeProduct = &v
	} else if v, ok := body["type_frame"].(string); ok {
		in.TypeProduct = &v
	}
	if v, ok := body["brand_id"].(float64); ok { id := int64(v); in.BrandID = &id }
	if v, ok := body["item_list_cost"].(float64); ok { in.ItemListCost = &v }
	if v, ok := body["item_discount"].(float64); ok { in.ItemDiscount = &v }
	if v, ok := body["item_net"].(float64); ok { in.ItemNet = &v }
	if v, ok := body["pb_selling_price"].(float64); ok { in.PbSellingPrice = &v }
	if v, ok := body["lens_cost"].(float64); ok { in.LensCost = &v }
	if v, ok := body["accessories_cost"].(float64); ok { in.AccessoriesCost = &v }

	newModelID, linkedType, err := h.svc.CreateCustomGlasses(inventoryID, employeeID, in)
	if err != nil {
		code := http.StatusInternalServerError
		if err.Error() == "inventory not found" || err.Error() == "model not found" {
			code = http.StatusNotFound
		}
		jsonError(w, err.Error(), code)
		return
	}
	jsonResponse(w, map[string]interface{}{
		"message":             "Custom model created; PriceBook (if found) updated",
		"new_model_id":        newModelID,
		"linked_product_type": linkedType,
	}, http.StatusCreated)
}

// POST /api/price-book/revert_custom_glasses/{inventory_id}
func (h *Handler) RevertCustomGlasses(w http.ResponseWriter, r *http.Request) {
	inventoryID, err := strconv.Atoi(mux.Vars(r)["inventory_id"])
	if err != nil || inventoryID == 0 {
		jsonError(w, "Invalid inventory_id", http.StatusBadRequest)
		return
	}
	employeeID := h.getEmployeeID(r)

	var body map[string]interface{}
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&body)
	}

	pbPrices := map[string]*float64{}
	if body != nil {
		for _, k := range []string{"pb_selling_price", "lens_cost", "accessories_cost"} {
			if v, ok := body[k].(float64); ok {
				vv := v
				pbPrices[k] = &vv
			}
		}
	}

	oldModelID, currentModelID, err := h.svc.RevertCustomGlasses(inventoryID, employeeID, pbPrices)
	if err != nil {
		code := http.StatusInternalServerError
		if err.Error() == "inventory not found" { code = http.StatusNotFound }
		jsonError(w, err.Error(), code)
		return
	}
	if oldModelID == 0 && currentModelID == 0 {
		jsonResponse(w, map[string]interface{}{
			"message":         "This item is already a regular model.",
			"already_regular": true,
		}, http.StatusOK)
		return
	}
	jsonResponse(w, map[string]interface{}{
		"message":                  "Inventory reverted to regular model.",
		"restored_model_id":        oldModelID,
		"previous_custom_model_id": currentModelID,
	}, http.StatusOK)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func (h *Handler) getEmployeeID(r *http.Request) int64 {
	username := pkgAuth.UsernameFromContext(r.Context())
	if username == "" { return 0 }
	type loginRow struct{ IDEmployeeLogin int64 }
	var login loginRow
	h.db.Table("employee_login").Select("id_employee_login").
		Where("employee_login = ?", username).Scan(&login)
	if login.IDEmployeeLogin == 0 { return 0 }
	type empRow struct{ IDEmployee int64 }
	var emp empRow
	h.db.Table("employee").Select("id_employee").
		Where("employee_login_id = ?", login.IDEmployeeLogin).Scan(&emp)
	return emp.IDEmployee
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
