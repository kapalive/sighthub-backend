package exameye

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	sleHandler "sighthub-backend/internal/handlers/external_sle_handler"
	"sighthub-backend/internal/middleware"
	sleSvc "sighthub-backend/internal/services/external_sle_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterExternalSleRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := sleSvc.New(db)
	h := sleHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	perm1 := middleware.ActivePermission(db, 1)
	perm3 := middleware.ActivePermission(db, 3)

	api := r.PathPrefix("/api/exam_eye/externalsle").Subrouter()
	api.Use(jwtMW, perm1)

	api.Handle("/{exam_id:[0-9]+}", perm3(http.HandlerFunc(h.SaveExternalSle))).Methods("POST")
	api.HandleFunc("/{exam_id:[0-9]+}", h.GetExternalSle).Methods("GET")
	api.HandleFunc("/{exam_id:[0-9]+}", h.UpdateExternalSle).Methods("PUT")
}
