package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	emailTplH "sighthub-backend/internal/handlers/email_template_handler"
	emailTplSvc "sighthub-backend/internal/services/email_template_service"
	"sighthub-backend/internal/middleware"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterEmailTemplateRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	s := emailTplSvc.New(db)
	h := emailTplH.New(s)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)

	api := r.PathPrefix("/api/email-templates").Subrouter()
	api.Use(jwtMW, middleware.ActivePermission(db, 80))

	api.HandleFunc("", h.GetAllTemplates).Methods("GET")
	api.HandleFunc("/", h.GetAllTemplates).Methods("GET")
	api.HandleFunc("/settings", h.SetOrgTemplate).Methods("POST")
}
