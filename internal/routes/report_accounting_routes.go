package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	"sighthub-backend/internal/handlers/report_accounting_handler"
	"sighthub-backend/internal/middleware"
	svc "sighthub-backend/internal/services/report_accounting_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterReportAccountingRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	s := svc.New(db)
	h := report_accounting_handler.New(s)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	storeMW := middleware.StorePermission(db, 12, 81)

	api := r.PathPrefix("/api/report_accounting").Subrouter()
	api.Use(jwtMW)
	api.Use(storeMW)

	api.HandleFunc("/monthly_summary", h.MonthlySummary).Methods("GET")
	api.HandleFunc("/daily_detail", h.DailyDetail).Methods("GET")
	api.HandleFunc("/payment_summary", h.PaymentSummary).Methods("GET")
	api.HandleFunc("/payment_details", h.PaymentDetails).Methods("GET")
	api.HandleFunc("/payment_categories", h.PaymentCategories).Methods("GET")
	api.HandleFunc("/insurance_companies", h.InsuranceCompanies).Methods("GET")
	api.HandleFunc("/payment_types", h.PaymentTypes).Methods("GET")
	api.HandleFunc("/ar_insurance_aging", h.ARInsuranceAging).Methods("GET")
}
