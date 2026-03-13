package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	"sighthub-backend/internal/handlers/report_inventory_handler"
	"sighthub-backend/internal/middleware"
	"sighthub-backend/internal/services/report_inventory_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterReportInventoryRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	s := report_inventory_service.New(db)
	h := report_inventory_handler.New(s, db)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	permMW := middleware.ActivePermission(db, 11)

	api := r.PathPrefix("/api/report-inventory").Subrouter()
	api.Use(jwtMW, permMW)

	api.HandleFunc("/frame-interaction", h.FrameInteraction).Methods("GET")
	api.HandleFunc("/missing_inventory", h.MissingInventory).Methods("GET")
	api.HandleFunc("/receipt_by_brand", h.ReceiptByBrand).Methods("GET")
	api.HandleFunc("/list_of_receipts", h.ListOfReceipts).Methods("GET")
	api.HandleFunc("/internal_transfers", h.InternalTransfers).Methods("GET")
	api.HandleFunc("/locations/can-receive", h.CanReceiveLocations).Methods("GET")
	api.HandleFunc("/locations/all", h.AllLocations).Methods("GET")
}
