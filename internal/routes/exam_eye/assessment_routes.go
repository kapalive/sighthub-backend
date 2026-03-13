package exameye

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	assessmentHandler "sighthub-backend/internal/handlers/assessment_handler"
	"sighthub-backend/internal/middleware"
	assessmentSvc "sighthub-backend/internal/services/assessment_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterAssessmentRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := assessmentSvc.New(db)
	h := assessmentHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	perm1 := middleware.ActivePermission(db, 1)
	perm3 := middleware.ActivePermission(db, 3)

	api := r.PathPrefix("/api/exam_eye/assessment").Subrouter()
	api.Use(jwtMW, perm1)

	api.Handle("/{exam_id:[0-9]+}", perm3(http.HandlerFunc(h.SaveAssessment))).Methods("POST")
	api.HandleFunc("/{exam_id:[0-9]+}", h.GetAssessments).Methods("GET")
	api.HandleFunc("/{exam_id:[0-9]+}", h.UpdateAssessment).Methods("PUT")
	api.HandleFunc("/{exam_id:[0-9]+}/{assessment_id:[0-9]+}/diagnosis/{diagnosis_id:[0-9]+}", h.DeleteAssessmentDiagnosis).Methods("DELETE")
	api.HandleFunc("/{exam_id:[0-9]+}/{assessment_id:[0-9]+}/pqrs/{pqrs_id:[0-9]+}", h.DeleteAssessmentPQRS).Methods("DELETE")
	api.HandleFunc("/{exam_id:[0-9]+}/{assessment_id:[0-9]+}", h.DeleteAssessment).Methods("DELETE")
	api.HandleFunc("/search", h.SearchDiagnosis).Methods("GET")
	api.HandleFunc("/my-top-diseases", h.GetMyTopDiseases).Methods("GET")
	api.HandleFunc("/my-top-diseases", h.AddMyTopDisease).Methods("POST")
	api.HandleFunc("/my-top-diseases/{disease_id:[0-9]+}", h.DeleteMyTopDisease).Methods("DELETE")
	api.HandleFunc("/pqrs", h.GetAllPQRS).Methods("GET")
}
