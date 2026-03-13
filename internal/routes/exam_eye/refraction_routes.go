package exameye

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	refrHandler "sighthub-backend/internal/handlers/refraction_handler"
	"sighthub-backend/internal/middleware"
	refrSvc "sighthub-backend/internal/services/refraction_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterRefractionRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := refrSvc.New(db)
	h := refrHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	perm1 := middleware.ActivePermission(db, 1)
	perm3 := middleware.ActivePermission(db, 3)

	api := r.PathPrefix("/api/exam_eye/refraction").Subrouter()
	api.Use(jwtMW, perm1)

	api.Handle("/{exam_id:[0-9]+}", perm3(http.HandlerFunc(h.SaveRefraction))).Methods("POST")
	api.HandleFunc("/{exam_id:[0-9]+}", h.GetRefraction).Methods("GET")
	api.HandleFunc("/{exam_id:[0-9]+}", h.UpdateRefraction).Methods("PUT")
}
