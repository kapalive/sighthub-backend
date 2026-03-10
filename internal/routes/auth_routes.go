package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	"sighthub-backend/internal/handlers/auth_handler"
	"sighthub-backend/internal/services/auth_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterAuthRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := auth_service.New(db, rdb, cfg)
	h := auth_handler.New(svc, cfg)
	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)

	auth := r.PathPrefix("/api/auth").Subrouter()

	// Публичные маршруты
	auth.HandleFunc("/login", h.Login).Methods("POST")
	auth.HandleFunc("/login_with_pin", h.LoginWithPin).Methods("POST")
	auth.HandleFunc("/token/refresh", h.Refresh).Methods("POST")
	auth.HandleFunc("/token-check", h.TokenCheck).Methods("POST")

	// Защищённые маршруты (требуют валидный JWT)
	protected := auth.NewRoute().Subrouter()
	protected.Use(jwtMW)
	protected.HandleFunc("/logout", h.Logout).Methods("POST")
}
