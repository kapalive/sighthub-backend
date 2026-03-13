package exameye

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	prelimHandler "sighthub-backend/internal/handlers/preliminary_handler"
	"sighthub-backend/internal/middleware"
	prelimSvc "sighthub-backend/internal/services/preliminary_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterPreliminaryRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := prelimSvc.New(db)
	h := prelimHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	perm1 := middleware.ActivePermission(db, 1)
	perm3 := middleware.ActivePermission(db, 3)

	api := r.PathPrefix("/api/exam_eye/preliminary").Subrouter()
	api.Use(jwtMW, perm1)

	// /prescription_list must be registered BEFORE /{exam_id} to avoid routing conflict
	api.HandleFunc("/prescription_list", h.GetPrescriptionList).Methods("GET")

	// Main preliminary endpoints
	api.Handle("/{exam_id:[0-9]+}", perm3(http.HandlerFunc(h.SavePreliminary))).Methods("POST")
	api.HandleFunc("/{exam_id:[0-9]+}", h.GetPreliminary).Methods("GET")
	api.HandleFunc("/{exam_id:[0-9]+}", h.UpdatePreliminary).Methods("PUT")

	// Entrance rx endpoints
	api.HandleFunc("/{exam_id:[0-9]+}/entrance_rx", h.FillEntranceRx).Methods("POST")
	api.HandleFunc("/{exam_id:[0-9]+}/entrance_rx", h.UpdateEntranceRx).Methods("PUT")
	api.HandleFunc("/{exam_id:[0-9]+}/entrance_rx", h.GetEntranceRx).Methods("GET")
	api.HandleFunc("/{exam_id:[0-9]+}/entrance_rx", h.DeleteEntranceRx).Methods("DELETE")

	// Near point testing endpoints
	api.HandleFunc("/{exam_id:[0-9]+}/near_point_testing", h.CreateNearPointTesting).Methods("POST")
	api.HandleFunc("/{exam_id:[0-9]+}/near_point_testing", h.UpdateNearPointTesting).Methods("PUT")
	api.HandleFunc("/{exam_id:[0-9]+}/near_point_testing", h.GetNearPointTesting).Methods("GET")
}
