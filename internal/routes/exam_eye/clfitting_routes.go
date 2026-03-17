package exameye

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	clHandler "sighthub-backend/internal/handlers/cl_fitting_handler"
	"sighthub-backend/internal/middleware"
	clSvc "sighthub-backend/internal/services/cl_fitting_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterClFittingRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := clSvc.New(db)
	h := clHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	perm1 := middleware.ActivePermission(db, 1)
	perm3 := middleware.ActivePermission(db, 3)

	api := r.PathPrefix("/api/exam_eye/clfitting").Subrouter()
	api.Use(jwtMW, perm1)

	api.Handle("/{exam_id:[0-9]+}", perm3(http.HandlerFunc(h.SaveClFitting))).Methods("POST")
	api.HandleFunc("/{exam_id:[0-9]+}", h.GetClFitting).Methods("GET")
	api.HandleFunc("/{exam_id:[0-9]+}", h.UpdateClFitting).Methods("PUT")
	api.HandleFunc("/contact_lens/brands", h.GetContactLensBrands).Methods("GET")
}
