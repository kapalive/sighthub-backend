package exameye

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	posteriorHandler "sighthub-backend/internal/handlers/posterior_handler"
	"sighthub-backend/internal/middleware"
	posteriorSvc "sighthub-backend/internal/services/posterior_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterPosteriorRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := posteriorSvc.New(db)
	h := posteriorHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	perm1 := middleware.ActivePermission(db, 1)
	perm3 := middleware.ActivePermission(db, 3)

	api := r.PathPrefix("/api/exam_eye/posterior").Subrouter()
	api.Use(jwtMW, perm1)

	api.Handle("/{exam_id:[0-9]+}", perm3(http.HandlerFunc(h.SavePosterior))).Methods("POST")
	api.HandleFunc("/{exam_id:[0-9]+}", h.GetPosterior).Methods("GET")
	api.HandleFunc("/{exam_id:[0-9]+}", h.UpdatePosterior).Methods("PUT")
}
