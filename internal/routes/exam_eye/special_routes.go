package exameye

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	specialHandler "sighthub-backend/internal/handlers/special_handler"
	"sighthub-backend/internal/middleware"
	specialSvc "sighthub-backend/internal/services/special_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterSpecialRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := specialSvc.New(db)
	h := specialHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	perm1 := middleware.ActivePermission(db, 1)
	perm3 := middleware.ActivePermission(db, 3)

	api := r.PathPrefix("/api/exam_eye/special").Subrouter()
	api.Use(jwtMW, perm1)

	api.Handle("/{exam_id:[0-9]+}", perm3(http.HandlerFunc(h.SaveSpecial))).Methods("POST")
	api.HandleFunc("/{exam_id:[0-9]+}", h.GetSpecial).Methods("GET")
	api.HandleFunc("/{exam_id:[0-9]+}", h.UpdateSpecial).Methods("PUT")
	api.HandleFunc("/file/{file_id:[0-9]+}", h.DeleteSpecialEyeFile).Methods("DELETE")
}
