package exameye

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	slrpHandler "sighthub-backend/internal/handlers/slrp_handler"
	"sighthub-backend/internal/middleware"
	slrpSvc "sighthub-backend/internal/services/slrp_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterSLRPRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := slrpSvc.New(db)
	h := slrpHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	perm1 := middleware.ActivePermission(db, 1)
	perm3 := middleware.ActivePermission(db, 3)

	api := r.PathPrefix("/api/exam_eye/slrp").Subrouter()
	api.Use(jwtMW, perm1)

	api.Handle("/{exam_id:[0-9]+}", perm3(http.HandlerFunc(h.CreateSLRP))).Methods("POST")
	api.HandleFunc("/{exam_id:[0-9]+}", h.UpdateSLRP).Methods("PUT")
	api.HandleFunc("/{exam_id:[0-9]+}", h.GetSLRP).Methods("GET")
}
