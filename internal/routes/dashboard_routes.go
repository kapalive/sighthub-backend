package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	dashHandler "sighthub-backend/internal/handlers/dashboard_handler"
	dashSvc "sighthub-backend/internal/services/dashboard_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterDashboardRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := dashSvc.New(db)
	h := dashHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)

	api := r.PathPrefix("/api/dashboard").Subrouter()
	api.Use(jwtMW)

	api.HandleFunc("/weekly_income", h.GetWeeklyIncome).Methods("GET")
	api.HandleFunc("/appointments", h.GetAppointmentStatuses).Methods("GET")
	api.HandleFunc("/employee_sales", h.GetEmployeeSales).Methods("GET")
	api.HandleFunc("/employee_invoices", h.GetEmployeeInvoices).Methods("GET")
}
