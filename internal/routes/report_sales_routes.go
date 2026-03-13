package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	"sighthub-backend/internal/handlers/report_sales_handler"
	"sighthub-backend/internal/middleware"
	"sighthub-backend/internal/services/report_sales_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterReportSalesRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	s := report_sales_service.New(db)
	h := report_sales_handler.New(s, db)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	permMW := middleware.ActivePermission(db, 11)

	api := r.PathPrefix("/api/report-sales").Subrouter()
	api.Use(jwtMW, permMW)

	api.HandleFunc("/breakdown", h.Breakdown).Methods("GET")
}
