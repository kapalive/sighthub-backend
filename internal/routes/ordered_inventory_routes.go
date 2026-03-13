package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	orderedH "sighthub-backend/internal/handlers/ordered_inventory_handler"
	"sighthub-backend/internal/middleware"
	orderedSvc "sighthub-backend/internal/services/ordered_inventory_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterOrderedInventoryRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := orderedSvc.New(db)
	h := orderedH.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)

	// permission 11=base, 18=create (Python uses @require_location_permission(18) for all 3 endpoints)
	baseMW   := middleware.ActivePermission(db, 11)
	createMW := middleware.ActivePermission(db, 18)

	api := r.PathPrefix("/api/ordered-inventory").Subrouter()
	api.Use(jwtMW)
	api.Use(baseMW)
	api.Use(createMW)

	api.HandleFunc("/add", h.AddOrderedItem).Methods("POST")
	api.HandleFunc("/pending", h.GetPendingItems).Methods("GET")
	api.HandleFunc("/receive", h.ReceiveOrderedItem).Methods("POST")
}
