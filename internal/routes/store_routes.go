package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	storeHandler "sighthub-backend/internal/handlers/store_handler"
	"sighthub-backend/internal/middleware"
	storeSvc "sighthub-backend/internal/services/store_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterStoreRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := storeSvc.New(db)
	h := storeHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	permMW := middleware.ActivePermission(db, 49)

	api := r.PathPrefix("/api/store").Subrouter()
	api.Use(jwtMW, permMW)

	// Stores
	api.HandleFunc("", h.GetAllStores).Methods("GET")
	api.HandleFunc("/", h.GetAllStores).Methods("GET")
	api.HandleFunc("", h.CreateStore).Methods("POST")
	api.HandleFunc("/", h.CreateStore).Methods("POST")
	api.HandleFunc("/{store_id:[0-9]+}", h.GetStore).Methods("GET")
	api.HandleFunc("/{store_id:[0-9]+}/request-appointment-link", h.GetRequestAppointmentLink).Methods("GET")
	api.HandleFunc("/{store_id:[0-9]+}", h.UpdateStore).Methods("PUT")
	api.HandleFunc("/{store_id:[0-9]+}/activate", h.ActivateStore).Methods("PUT")

	// Warehouses
	api.HandleFunc("/warehouses", h.GetWarehouses).Methods("GET")
	api.HandleFunc("/warehouses", h.CreateWarehouse).Methods("POST")
	api.HandleFunc("/warehouses/{warehouse_id:[0-9]+}", h.GetWarehouse).Methods("GET")
	api.HandleFunc("/warehouses/{warehouse_id:[0-9]+}", h.UpdateWarehouse).Methods("PUT")

	// Sales tax
	api.HandleFunc("/sales_tax_list", h.GetSalesTaxList).Methods("GET")
	api.HandleFunc("/sales_taxes", h.GetSalesTaxes).Methods("GET")
}
