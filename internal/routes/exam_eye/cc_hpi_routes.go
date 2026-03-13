package exameye

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	ccHpiHandler "sighthub-backend/internal/handlers/cc_hpi_handler"
	"sighthub-backend/internal/middleware"
	ccHpiSvc "sighthub-backend/internal/services/cc_hpi_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterCcHpiRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := ccHpiSvc.New(db)
	h := ccHpiHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	perm1 := middleware.ActivePermission(db, 1)
	perm3 := middleware.ActivePermission(db, 3)

	api := r.PathPrefix("/api/exam_eye/cc_hpi").Subrouter()
	api.Use(jwtMW, perm1)

	api.Handle("/{exam_id:[0-9]+}", perm3(http.HandlerFunc(h.SaveCcHpi))).Methods("POST")
	api.HandleFunc("/{exam_id:[0-9]+}", h.GetCcHpi).Methods("GET")
	api.HandleFunc("/{exam_id:[0-9]+}", h.UpdateCcHpi).Methods("PUT")
}
