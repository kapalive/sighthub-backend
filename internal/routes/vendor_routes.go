package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	intHandler "sighthub-backend/internal/handlers/integration_handler"
	vendorHandler "sighthub-backend/internal/handlers/vendor_handler/vendor"
	"sighthub-backend/internal/middleware"
	intSvc "sighthub-backend/internal/services/integration_service"
	vendorSvc "sighthub-backend/internal/services/vendor_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterVendorRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := vendorSvc.New(db)
	h := vendorHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	basePerm := middleware.ActivePermission(db, 36)

	api := r.PathPrefix("/api/vendor").Subrouter()
	api.Use(jwtMW, basePerm)

	// Per-route permission middleware
	perm37 := middleware.RequireLocationPermission(db, 37) // view
	perm38 := middleware.RequireLocationPermission(db, 38) // add
	perm39 := middleware.RequireLocationPermission(db, 39) // edit
	perm40 := middleware.RequireLocationPermission(db, 40) // delete

	// --- Vendor CRUD ---
	api.Handle("/add", perm38(http.HandlerFunc(h.AddVendor))).Methods("POST")
	api.Handle("/{vendor_id:[0-9]+}", perm39(http.HandlerFunc(h.UpdateVendor))).Methods("PUT")
	api.Handle("/{vendor_id:[0-9]+}", perm40(http.HandlerFunc(h.DeleteVendor))).Methods("DELETE")
	api.Handle("/{vendor_id:[0-9]+}", perm37(http.HandlerFunc(h.GetVendor))).Methods("GET")
	api.HandleFunc("/list", h.ListVendors).Methods("GET")

	// --- Invoices ---
	api.HandleFunc("/{vendor_id:[0-9]+}/invoices", h.GetVendorInvoices).Methods("GET")

	// --- Agreement ---
	api.Handle("/{vendor_id:[0-9]+}/agreement", perm39(http.HandlerFunc(h.CreateAgreement))).Methods("POST")
	api.Handle("/{vendor_id:[0-9]+}/agreement/{agreement_id:[0-9]+}", perm39(http.HandlerFunc(h.UpdateAgreement))).Methods("PUT")
	api.Handle("/{vendor_id:[0-9]+}/agreement/{agreement_id:[0-9]+}", perm40(http.HandlerFunc(h.DeleteAgreement))).Methods("DELETE")

	// --- Brands ---
	api.Handle("/{vendor_id:[0-9]+}/add_brand", perm39(http.HandlerFunc(h.AddVendorBrand))).Methods("POST")
	api.Handle("/update_brand/{brand_id:[0-9]+}", perm39(http.HandlerFunc(h.UpdateVendorBrand))).Methods("PUT")
	api.Handle("/{vendor_id:[0-9]+}/brand/{brand_type}/{brand_id:[0-9]+}", perm39(http.HandlerFunc(h.DeleteVendorBrand))).Methods("DELETE")

	// --- Labs ---
	api.HandleFunc("/lab/list", h.ListLabs).Methods("GET")
	api.Handle("/lab", perm38(http.HandlerFunc(h.CreateLab))).Methods("POST")
	api.Handle("/lab/{id_lab:[0-9]+}", perm39(http.HandlerFunc(h.UpdateLab))).Methods("PUT")
	api.HandleFunc("/lab/{id_lab:[0-9]+}", h.GetLab).Methods("GET")
	api.Handle("/lab/{id_lab:[0-9]+}", perm40(http.HandlerFunc(h.DeleteLab))).Methods("DELETE")

	// --- Vendor-Lab links ---
	// add_lab removed — lab auto-links to vendor on POST /lab
	api.Handle("/{vendor_id:[0-9]+}/remove_lab/{lab_id:[0-9]+}", perm39(http.HandlerFunc(h.RemoveVendorLab))).Methods("DELETE")

	// --- Integration VW Labs ---
	hInt := intHandler.NewHandler(intSvc.New(db))
	api.HandleFunc("/integration/vw/labs", hInt.GetVisionWebLabs).Methods("GET")

	// --- Countries / States ---
	api.HandleFunc("/countries", h.GetCountries).Methods("GET")
	api.HandleFunc("/states/{country_id:[0-9]+}", h.GetStatesByCountry).Methods("GET")

	// --- Pricing Rules ---
	api.Handle("/{vendor_id:[0-9]+}/pricing-rules/{brand_type}/{brand_id:[0-9]+}", perm39(http.HandlerFunc(h.AddPricingRule))).Methods("POST")
	api.HandleFunc("/{vendor_id:[0-9]+}/pricing-rules/{brand_type}/{brand_id:[0-9]+}", h.GetPricingRules).Methods("GET")
	api.Handle("/{vendor_id:[0-9]+}/pricing-rules/{rule_id:[0-9]+}", perm39(http.HandlerFunc(h.UpdatePricingRule))).Methods("PUT")
	api.Handle("/{vendor_id:[0-9]+}/pricing-rules/{rule_id:[0-9]+}", perm40(http.HandlerFunc(h.DeletePricingRule))).Methods("DELETE")
}
