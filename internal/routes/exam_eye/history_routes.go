package exameye

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	histHandler "sighthub-backend/internal/handlers/history_handler"
	"sighthub-backend/internal/middleware"
	histSvc "sighthub-backend/internal/services/history_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterHistoryRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := histSvc.New(db)
	h := histHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	perm1 := middleware.ActivePermission(db, 1)
	perm3 := middleware.ActivePermission(db, 3)

	api := r.PathPrefix("/api/exam_eye/history").Subrouter()
	api.Use(jwtMW, perm1)

	// Races / Ethnicities — registered before /{exam_id} to avoid matching conflict
	api.HandleFunc("/races", h.GetRaces).Methods("GET")
	api.HandleFunc("/ethnicities", h.GetEthnicities).Methods("GET")

	// Medications
	api.HandleFunc("/medications/{exam_id:[0-9]+}", h.GetMedications).Methods("GET")
	api.HandleFunc("/medications/{exam_id:[0-9]+}", h.SaveMedication).Methods("POST")
	api.HandleFunc("/medications/{exam_id:[0-9]+}/{medication_id:[0-9]+}", h.DeleteMedication).Methods("DELETE")

	// Allergies
	api.HandleFunc("/allergies/{exam_id:[0-9]+}", h.GetAllergies).Methods("GET")
	api.HandleFunc("/allergies/{exam_id:[0-9]+}", h.SaveAllergy).Methods("POST")
	api.HandleFunc("/allergies/{exam_id:[0-9]+}/{allergy_id:[0-9]+}", h.DeleteAllergy).Methods("DELETE")

	// History by exam — patient-info before /{exam_id} GET to avoid conflict
	api.HandleFunc("/{exam_id:[0-9]+}/patient-info", h.GetPatientInfo).Methods("GET")
	api.Handle("/{exam_id:[0-9]+}", perm3(http.HandlerFunc(h.SaveHistory))).Methods("POST")
	api.HandleFunc("/{exam_id:[0-9]+}", h.GetHistory).Methods("GET")
	api.HandleFunc("/{exam_id:[0-9]+}", h.UpdateHistory).Methods("PUT")
}
