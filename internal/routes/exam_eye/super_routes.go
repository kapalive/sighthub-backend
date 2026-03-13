package exameye

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	superHandler "sighthub-backend/internal/handlers/super_handler"
	"sighthub-backend/internal/middleware"
	superSvc "sighthub-backend/internal/services/super_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterSuperRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := superSvc.New(db)
	h := superHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	perm1 := middleware.ActivePermission(db, 1)
	perm3 := middleware.ActivePermission(db, 3)

	api := r.PathPrefix("/api/exam_eye/super").Subrouter()
	api.Use(jwtMW, perm1)

	api.Handle("/{exam_id:[0-9]+}", perm3(http.HandlerFunc(h.CreateSuperEyeExam))).Methods("POST")
	api.Handle("/{exam_id:[0-9]+}/invoice", perm3(http.HandlerFunc(h.CreateSuperInvoice))).Methods("POST")
	api.HandleFunc("/{exam_id:[0-9]+}", h.UpdateSuperEyeExam).Methods("PUT")
	api.HandleFunc("/{exam_id:[0-9]+}", h.GetSuperEyeExam).Methods("GET")
	api.HandleFunc("/{exam_id:[0-9]+}/item/{item_id:[0-9]+}/diagnosis", h.DeleteDiagnosisByID).Methods("DELETE")
	api.HandleFunc("/{exam_id:[0-9]+}/item/{item_id:[0-9]+}", h.DeleteItemFromInvoice).Methods("DELETE")
	api.HandleFunc("/prof-serv", h.GetProfessionalServices).Methods("GET")
	api.HandleFunc("/super-bill-diseases", h.GetSuperBillDiseases).Methods("GET")
}
