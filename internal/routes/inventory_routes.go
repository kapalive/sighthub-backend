package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	"sighthub-backend/internal/handlers/inventory_handler"
	"sighthub-backend/internal/handlers/invoice_handler"
	"sighthub-backend/internal/middleware"
	"sighthub-backend/internal/services/inventory_service"
	"sighthub-backend/internal/services/invoice_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterInventoryRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := inventory_service.New(db)
	h := inventory_handler.New(svc, db)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	// permission 11 = inventory access
	invMW := middleware.ActivePermission(db, 11)

	api := r.PathPrefix("/api/inventory").Subrouter()
	api.Use(jwtMW, invMW)

	// ── Helper GETs ─────────────────────────────────────────────────────
	api.HandleFunc("/vendors", h.GetVendorsWithBrands).Methods("GET")
	api.HandleFunc("/brands", h.GetBrands).Methods("GET")
	api.HandleFunc("/stores", h.GetStores).Methods("GET")
	api.HandleFunc("/brands/{brand_id:[0-9]+}/products", h.GetProductsByBrand).Methods("GET")
	api.HandleFunc("/products/{product_id:[0-9]+}/variants", h.GetVariantsByProduct).Methods("GET")
	api.HandleFunc("/item-statuses", h.GetItemStatuses).Methods("GET")
	api.HandleFunc("/products/search", h.SearchProducts).Methods("GET")
	api.HandleFunc("/search-model", h.SearchModel).Methods("GET")
	api.HandleFunc("/stock-by-model", h.GetStockByModel).Methods("GET")
	api.HandleFunc("/inventory-receipt", h.GetInventoryReceipt).Methods("GET")

	// ── Lookup & status ─────────────────────────────────────────────────
	api.HandleFunc("/lookup", h.LookupBySKU).Methods("GET")
	api.HandleFunc("/history", h.GetInventoryHistory).Methods("GET")
	api.HandleFunc("/update_price/{inventory_id:[0-9]+}", h.UpdatePrice).Methods("PUT")
	api.HandleFunc("/update_state", h.UpdateInventoryState).Methods("POST")
	api.HandleFunc("/calc/sell-price", h.CalcSellPrice).Methods("POST")

	// ── Add inventory ───────────────────────────────────────────────────
	api.HandleFunc("/add", h.AddInventoryItem).Methods("POST")

	// ── All-filters (paginated + CSV) ───────────────────────────────────
	api.HandleFunc("/all-filters", h.GetInventoryByFilter).Methods("GET")

	// ── Count sheets ────────────────────────────────────────────────────
	api.HandleFunc("/count_sheets", h.GetCountSheets).Methods("GET")
	api.HandleFunc("/count_sheets", h.CreateCountSheet).Methods("POST")
	api.HandleFunc("/count_sheets", h.DeleteCountSheet).Methods("DELETE")
	api.HandleFunc("/count_sheets/{id_count_sheet:[0-9]+}", h.GetCountSheetInfo).Methods("GET")
	api.HandleFunc("/count_sheets/{id_count_sheet:[0-9]+}", h.UpdateCountSheetNotes).Methods("PUT")
	api.HandleFunc("/count_sheets/items", h.GetCountSheetItems).Methods("GET")
	api.HandleFunc("/count_sheets/items", h.AddItemToCountSheet).Methods("POST")
	api.HandleFunc("/count_sheets/items", h.DeleteItemFromCountSheet).Methods("DELETE")
	api.HandleFunc("/count_sheets/close/{id_count_sheet:[0-9]+}", h.CloseCountSheet).Methods("POST")

	// ── Labels (TODO: depends on utils_printer) ─────────────────────────
	api.HandleFunc("/preview-label", h.PreviewLabel).Methods("GET")
	api.HandleFunc("/print-label", h.PrintLabel).Methods("GET")
	api.HandleFunc("/print-labels-by-vendor", h.PrintLabelsByVendor).Methods("GET")

	// ── Local Transfers ──────────────────────────────────────────────────
	invSvc := invoice_service.New(db)
	invH := invoice_handler.New(invSvc)
	api.HandleFunc("/local-transfer", invH.CreateLocalTransfer).Methods("POST")
	api.HandleFunc("/local-transfers", invH.GetLocalTransfers).Methods("GET")
	api.HandleFunc("/local-transfer", invH.ReverseLocalTransfer).Methods("DELETE")
}
