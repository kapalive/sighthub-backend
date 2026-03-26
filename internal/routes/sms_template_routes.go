package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	"sighthub-backend/internal/handlers/sms_template_handler"
	"sighthub-backend/internal/middleware"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterSMSTemplateRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	h := sms_template_handler.New(db)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)

	api := r.PathPrefix("/api/sms-templates").Subrouter()
	api.Use(jwtMW, middleware.ActivePermission(db, 80))

	api.HandleFunc("", h.GetAll).Methods("GET")
	api.HandleFunc("/", h.GetAll).Methods("GET")
	api.HandleFunc("", h.Create).Methods("POST")
	api.HandleFunc("/", h.Create).Methods("POST")
	api.HandleFunc("/{id:[0-9]+}", h.Update).Methods("PUT")
	api.HandleFunc("/{id:[0-9]+}", h.Delete).Methods("DELETE")
	api.HandleFunc("/variables", h.GetVariables).Methods("GET")
}
