package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	dailyCloseHandler "sighthub-backend/internal/handlers/daily_close_handler"
	dailyCloseSvc "sighthub-backend/internal/services/daily_close_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterDailyCloseRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := dailyCloseSvc.New(db)
	h := dailyCloseHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)

	api := r.PathPrefix("/api/daily_close").Subrouter()
	api.Use(jwtMW)

	// Count sheet CRUD
	api.HandleFunc("/count_sheet", h.CreateDailyClose).Methods("POST")
	api.HandleFunc("/count_sheet", h.GetDailyCloseSummary).Methods("GET")
	api.HandleFunc("/count_sheet", h.UpdateDailyClose).Methods("PUT")

	// Payment methods
	api.HandleFunc("/payment_methods", h.GetPaymentMethods).Methods("GET")

	// Reports
	api.HandleFunc("/daily_close_detail", h.GetDailyCloseDetail).Methods("GET")
	api.HandleFunc("/invoices_payments", h.GetInvoicesPayments).Methods("GET")
	api.HandleFunc("/payments_summary", h.GetPaymentsSummary).Methods("GET")
	api.HandleFunc("/transfer_credit_summary", h.GetTransferCreditSummary).Methods("GET")

	// HTML reports
	api.HandleFunc("/daily_close_report_html", h.RenderDailyCloseReportHTML).Methods("GET")
	api.HandleFunc("/count_sheet_html", h.RenderCountSheetHTML).Methods("GET")
}
