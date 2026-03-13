package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	profileHandler "sighthub-backend/internal/handlers/profile_handler"
	profileSvc "sighthub-backend/internal/services/profile_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterProfileRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := profileSvc.New(db, rdb, cfg)
	h := profileHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)

	api := r.PathPrefix("/api/profile").Subrouter()
	api.Use(jwtMW)

	api.HandleFunc("/express_pass", h.ExpressPass).Methods("POST")
	api.HandleFunc("/info", h.GetInfo).Methods("GET")
	api.HandleFunc("/change_password", h.ChangePassword).Methods("POST")
	api.Handle("/invoice/{invoice_id:[0-9]+}", http.HandlerFunc(h.GetInvoiceByID)).Methods("GET")
}
