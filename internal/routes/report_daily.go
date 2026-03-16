package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	"sighthub-backend/internal/handlers/report_daily_handler"
	"sighthub-backend/internal/middleware"
	"sighthub-backend/internal/services/report_daily_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterReportDailyRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	s := report_daily_service.New(db)
	h := report_daily_handler.New(s)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	storeMW := middleware.StorePermission(db, 12, 81)

	api := r.PathPrefix("/api/report_daily").Subrouter()
	api.Use(jwtMW, storeMW)

	api.HandleFunc("/sales", h.DailySalesSummary).Methods("GET")
	api.HandleFunc("/monthly_sales_summary", h.MonthlySalesSummary).Methods("GET")
	api.HandleFunc("/ytd_sales_summary", h.YTDSalesSummary).Methods("GET")
	api.HandleFunc("/sales_cash", h.DailySalesCash).Methods("GET")
	api.HandleFunc("/journal_report", h.JournalReport).Methods("GET")
	api.HandleFunc("/journal_transfer", h.JournalTransfer).Methods("GET")
	api.HandleFunc("/journal_receipts", h.JournalReceipts).Methods("GET")
	api.HandleFunc("/all_reports", h.AllReports).Methods("GET")
	api.HandleFunc("/locations", h.Locations).Methods("GET")
	api.HandleFunc("/employees", h.Employees).Methods("GET")
}
