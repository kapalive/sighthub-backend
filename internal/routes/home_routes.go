package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	"sighthub-backend/internal/handlers/home_handler"
	"sighthub-backend/internal/middleware"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterHomeRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	h := home_handler.New(db)
	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	navMW := middleware.NavigationPermission(db)
	storeMW := middleware.StorePermission(db, 12, 81)

	// All home routes require JWT
	api := r.PathPrefix("/api/home").Subrouter()
	api.Use(jwtMW)

	// Navigation-permission routes
	nav := api.NewRoute().Subrouter()
	nav.Use(navMW)
	nav.HandleFunc("/home", h.Home).Methods("GET")
	nav.HandleFunc("/notify/counts", h.NotifyCounts).Methods("GET")

	// Store-permission routes (block=12, perm=81)
	store := api.NewRoute().Subrouter()
	store.Use(storeMW)
	store.HandleFunc("/set-stores-list", h.GetStoresList).Methods("GET")
	store.HandleFunc("/set_store", h.SetStore).Methods("POST")

	// JWT-only routes
	api.HandleFunc("/invoice/search", h.SearchInvoice).Methods("GET")
	api.HandleFunc("/locations", h.GetLocations).Methods("GET")
	api.HandleFunc("/insurance/companies", h.GetInsuranceCompanies).Methods("GET")
	api.HandleFunc("/insurance/types", h.GetInsuranceTypes).Methods("GET")
	api.HandleFunc("/languages", h.GetLanguages).Methods("GET")
	api.HandleFunc("/patient/search", h.SearchPatients).Methods("GET")
	api.HandleFunc("/patient/recently-viewed", h.RecentlyViewedPatients).Methods("GET")
	api.HandleFunc("/express_pass", h.ExpressPass).Methods("POST")
	api.HandleFunc("/gift_card/new", h.CreateGiftCard).Methods("POST")
	api.HandleFunc("/gift_card/details", h.GetGiftCardDetails).Methods("GET")
	api.HandleFunc("/gift_card/list", h.ListGiftCards).Methods("GET")
}
