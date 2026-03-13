package exameye

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	eeHandler "sighthub-backend/internal/handlers/exam_eye_handler"
	"sighthub-backend/internal/middleware"
	eeSvc "sighthub-backend/internal/services/exam_eye_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterExamEyeRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := eeSvc.New(db)
	h := eeHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	perm1 := middleware.ActivePermission(db, 1)
	perm3 := middleware.ActivePermission(db, 3)
	perm4 := middleware.ActivePermission(db, 4)
	perm5 := middleware.ActivePermission(db, 5)

	api := r.PathPrefix("/api/exam_eye").Subrouter()
	api.Use(jwtMW, perm1)

	// Exam types
	api.HandleFunc("/exam-types", h.GetExamTypes).Methods("GET")

	// Exam CRUD
	api.Handle("/new", perm3(http.HandlerFunc(h.StartNewExam))).Methods("POST")
	api.Handle("/submit/{exam_id:[0-9]+}", perm3(http.HandlerFunc(h.SubmitExam))).Methods("PUT")
	api.HandleFunc("/{id_exam:[0-9]+}", h.GetExamDetails).Methods("GET")
	api.Handle("/cancel/{id_exam:[0-9]+}", perm5(http.HandlerFunc(h.CancelExam))).Methods("DELETE")
	api.Handle("/unlock/{exam_id:[0-9]+}", perm4(http.HandlerFunc(h.UnlockExam))).Methods("PUT")
	api.Handle("/change-type/{exam_id:[0-9]+}", perm3(http.HandlerFunc(h.ChangeExamType))).Methods("PUT")

	// Notes
	api.HandleFunc("/notes", h.GetNotes).Methods("GET")
	api.Handle("/notes", perm3(http.HandlerFunc(h.AddNote))).Methods("POST")
	api.Handle("/notes", perm4(http.HandlerFunc(h.UpdateNote))).Methods("PUT")
	api.Handle("/notes", perm5(http.HandlerFunc(h.DeleteNote))).Methods("DELETE")
}
