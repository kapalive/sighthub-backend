package inventory_handler

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// GET /vendors
func (h *Handler) GetVendorsWithBrands(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetVendorsWithBrands()
	if err != nil {
		jsonResponse(w, 500, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// GET /brands
func (h *Handler) GetBrands(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.GetBrands()
	if err != nil {
		jsonResponse(w, 500, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// GET /stores
func (h *Handler) GetStores(w http.ResponseWriter, r *http.Request) {
	locIDs, err := resolveLocationIDs(h.db, r, nil)
	if err != nil || len(locIDs) == 0 {
		jsonResponse(w, 200, []map[string]interface{}{})
		return
	}
	result, err := h.svc.GetStores(locIDs)
	if err != nil {
		jsonResponse(w, 500, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// GET /brands/{brand_id}/products
func (h *Handler) GetProductsByBrand(w http.ResponseWriter, r *http.Request) {
	brandID, err := strconv.ParseInt(mux.Vars(r)["brand_id"], 10, 64)
	if err != nil {
		jsonResponse(w, 400, map[string]string{"error": "invalid brand_id"})
		return
	}
	result, err := h.svc.GetProductsByBrand(brandID)
	if err != nil {
		jsonResponse(w, 500, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// GET /products/{product_id}/variants
func (h *Handler) GetVariantsByProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(mux.Vars(r)["product_id"], 10, 64)
	if err != nil {
		jsonResponse(w, 400, map[string]string{"error": "invalid product_id"})
		return
	}
	result, err := h.svc.GetVariantsByProduct(productID)
	if err != nil {
		jsonResponse(w, 500, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// GET /item-statuses
func (h *Handler) GetItemStatuses(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, 200, h.svc.GetItemStatuses())
}

// GET /products/search
func (h *Handler) SearchProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		jsonResponse(w, 400, map[string]string{"error": "Search query is required"})
		return
	}
	result, err := h.svc.SearchProducts(q)
	if err != nil {
		jsonResponse(w, 500, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// GET /search-model
func (h *Handler) SearchModel(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("model")
	if q == "" {
		jsonResponse(w, 400, map[string]string{"error": "Parameter 'model' is required"})
		return
	}
	result, err := h.svc.SearchModel(q)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// GET /stock-by-model
func (h *Handler) GetStockByModel(w http.ResponseWriter, r *http.Request) {
	modelID, err := strconv.ParseInt(r.URL.Query().Get("model_id"), 10, 64)
	if err != nil || modelID == 0 {
		jsonResponse(w, 400, map[string]string{"error": "Parameter 'model_id' is required"})
		return
	}
	result, err := h.svc.GetStockByModel(modelID)
	if err != nil {
		jsonResponse(w, 500, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// GET /inventory-receipt
func (h *Handler) GetInventoryReceipt(w http.ResponseWriter, r *http.Request) {
	productID, _ := strconv.ParseInt(r.URL.Query().Get("product_id"), 10, 64)
	locationID, _ := strconv.ParseInt(r.URL.Query().Get("location_id"), 10, 64)
	if productID == 0 || locationID == 0 {
		jsonResponse(w, 400, map[string]string{"error": "Parameters 'product_id' and 'location_id' are required"})
		return
	}
	result, err := h.svc.GetInventoryReceipt(productID, locationID)
	if err != nil {
		jsonResponse(w, 500, map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}

// POST /calc/sell-price
func (h *Handler) CalcSellPrice(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BrandType string  `json:"brand_type"`
		BrandID   int64   `json:"brand_id"`
		ListPrice float64 `json:"list_price"`
	}
	if err := decodeJSON(r, &body); err != nil {
		jsonResponse(w, 400, map[string]string{"error": "invalid request body"})
		return
	}
	if body.BrandType == "" || body.ListPrice < 0 {
		jsonResponse(w, 400, map[string]string{"error": "brand_type, brand_id, and list_price are required"})
		return
	}
	result, err := h.svc.CalcSellPrice(body.BrandType, body.BrandID, body.ListPrice)
	if err != nil {
		jsonResponse(w, httpStatus(err), map[string]string{"error": err.Error()})
		return
	}
	jsonResponse(w, 200, result)
}
