package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	billHandler "sighthub-backend/internal/handlers/bill_handler"
	"sighthub-backend/internal/middleware"
	billSvc "sighthub-backend/internal/services/bill_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterBillRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := billSvc.New(db)
	h := billHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	perm31 := middleware.ActivePermission(db, 31)

	api := r.PathPrefix("/api/bill").Subrouter()
	api.Use(jwtMW, perm31)

	api.HandleFunc("/forms", h.GetFormData).Methods("GET")
	api.HandleFunc("/forms/{id_form:[0-9]+}/fields", h.GetFormFields).Methods("GET")
	api.HandleFunc("/forms/{id_form:[0-9]+}/submit", h.SubmitForm).Methods("POST")
}
