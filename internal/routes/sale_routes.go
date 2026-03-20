package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	"sighthub-backend/internal/handlers/sale_handler"
	"sighthub-backend/internal/middleware"
	"sighthub-backend/internal/services/sale_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterSaleRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	s := sale_service.New(db)
	h := sale_handler.New(s, db)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	baseMW := middleware.AnyActivePermission(db, 41, 42, 43, 44, 45, 46)

	api := r.PathPrefix("/api/sale").Subrouter()
	api.Use(jwtMW, baseMW)

	// Simple lookups
	api.HandleFunc("/stores", h.GetLocations).Methods("GET")
	api.HandleFunc("/vendors", h.GetVendors).Methods("GET")
	api.HandleFunc("/vendors/{vendor_id:[0-9]+}/brands", h.GetVendorBrands).Methods("GET")
	api.HandleFunc("/employees", h.GetEmployees).Methods("GET")

	// Sale items (complex filter)
	api.HandleFunc("/item", h.GetItems).Methods("GET")

	// Yearly comparisons (extra permissions via ActivePermission inline)
	perm42 := middleware.ActivePermission(db, 42)
	api.Handle("/yearly_comparison_by_rep", perm42(http.HandlerFunc(h.YearlyComparisonByRep))).Methods("GET")

	perm43 := middleware.ActivePermission(db, 43)
	api.Handle("/yearly_comparison_by_brand", perm43(http.HandlerFunc(h.YearlyComparisonByBrand))).Methods("GET")

	perm44 := middleware.ActivePermission(db, 44)
	api.Handle("/professional_codes", perm44(http.HandlerFunc(h.ProfessionalCodes))).Methods("GET")

	// Insurance
	api.HandleFunc("/insurance", h.GetInsuranceCompanies).Methods("GET")

	perm45 := middleware.ActivePermission(db, 45)
	api.Handle("/insurance_report", perm45(http.HandlerFunc(h.InsuranceReport))).Methods("GET")

	// Commission
	perm46 := middleware.ActivePermission(db, 46)
	api.Handle("/commission", perm46(http.HandlerFunc(h.Commission))).Methods("GET")

	// Sales report
	api.HandleFunc("/sales_report", h.SalesReport).Methods("GET")

	// Referral report (block_id=12, permission_id=81)
	referralMW := middleware.StorePermission(db, 12, 81)
	api.Handle("/referral_report", referralMW(http.HandlerFunc(h.ReferralReport))).Methods("GET")
}
