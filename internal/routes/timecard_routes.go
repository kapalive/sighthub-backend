package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	timecardH "sighthub-backend/internal/handlers/timecard_handler"
	timecardSvc "sighthub-backend/internal/services/timecard_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterTimecardRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	s := timecardSvc.New(db, rdb, cfg)
	h := timecardH.New(s)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)

	api := r.PathPrefix("/api/timecard").Subrouter()

	// Публичные маршруты
	api.HandleFunc("/session_status", h.SessionStatus).Methods("GET")
	api.HandleFunc("/check", h.Check).Methods("POST")
	api.HandleFunc("/login", h.Login).Methods("POST")

	// Защищённые маршруты (JWT)
	protected := api.NewRoute().Subrouter()
	protected.Use(jwtMW)
	protected.HandleFunc("/info", h.Info).Methods("GET")
	protected.HandleFunc("/change_password", h.ChangePassword).Methods("POST")
	protected.HandleFunc("/history/{history_id:[0-9]+}/note", h.UpdateHistoryNote).Methods("PUT")
	protected.HandleFunc("/logout", h.Logout).Methods("POST")
}
