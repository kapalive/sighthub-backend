package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	ddHandler "sighthub-backend/internal/handlers/doctor_desk_handler"
	"sighthub-backend/internal/middleware"
	ddSvc "sighthub-backend/internal/services/doctor_desk_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterDoctorDeskRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := ddSvc.New(db)
	h := ddHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	perm1  := middleware.ActivePermission(db, 1)
	perm3  := middleware.ActivePermission(db, 3)
	perm4  := middleware.ActivePermission(db, 4)
	perm5  := middleware.ActivePermission(db, 5)
	perm9  := middleware.ActivePermission(db, 9)
	perm58 := middleware.ActivePermission(db, 58)
	perm59 := middleware.ActivePermission(db, 59)
	perm60 := middleware.ActivePermission(db, 60)
	perm61 := middleware.ActivePermission(db, 61)

	api := r.PathPrefix("/api/doctor_desk").Subrouter()
	api.Use(jwtMW, perm1)

	// ─── Appointments ──────────────────────────────────────────────────────────
	api.HandleFunc("/appointments", h.GetAppointments).Methods("GET")
	api.Handle("/appointment/status/{id:[0-9]+}", perm9(http.HandlerFunc(h.UpdateAppointmentStatus))).Methods("PUT")

	// ─── Doctors / Locations ──────────────────────────────────────────────────
	api.HandleFunc("/doctors", h.GetDoctors).Methods("GET")
	api.HandleFunc("/location", h.GetShowcaseLocations).Methods("GET")

	// ─── Patient search ────────────────────────────────────────────────────────
	api.HandleFunc("/patient-search", h.SearchPatients).Methods("GET")

	// ─── Exams ─────────────────────────────────────────────────────────────────
	api.HandleFunc("/exams/unsigned", h.GetUnsignedExams).Methods("GET")
	api.HandleFunc("/patient/{id:[0-9]+}/exams", h.GetPatientExams).Methods("GET")

	// ─── Files — patient list BEFORE generic file routes ──────────────────────
	api.Handle("/files/patient/{id:[0-9]+}", perm58(http.HandlerFunc(h.GetPatientFiles))).Methods("GET")
	api.Handle("/files/{id:[0-9]+}", perm59(http.HandlerFunc(h.UploadExamFile))).Methods("POST")
	api.Handle("/files/{id:[0-9]+}", perm60(http.HandlerFunc(h.UpdateExamFile))).Methods("PUT")
	api.Handle("/files/{id:[0-9]+}", perm58(http.HandlerFunc(h.GetExamFile))).Methods("GET")
	api.Handle("/files/{id:[0-9]+}", perm61(http.HandlerFunc(h.DeleteExamFile))).Methods("DELETE")

	// ─── Patient notes ─────────────────────────────────────────────────────────
	api.HandleFunc("/patient/{id:[0-9]+}/notes", h.GetPatientNotes).Methods("GET")
	api.Handle("/patient/{id:[0-9]+}/notes", perm3(http.HandlerFunc(h.CreatePatientNote))).Methods("POST")
	api.Handle("/patient/{id:[0-9]+}/notes/{note_id:[0-9]+}", perm4(http.HandlerFunc(h.UpdatePatientNote))).Methods("PUT")
	api.HandleFunc("/patient/{id:[0-9]+}/notes/{note_id:[0-9]+}", h.GetPatientNote).Methods("GET")
	api.Handle("/patient/{id:[0-9]+}/notes/{note_id:[0-9]+}", perm5(http.HandlerFunc(h.DeletePatientNote))).Methods("DELETE")

	// ─── Patient info ──────────────────────────────────────────────────────────
	api.HandleFunc("/patient/{id:[0-9]+}/info", h.GetPatientInfo).Methods("GET")
	api.HandleFunc("/patient/{id:[0-9]+}/info", h.UpdatePatientInfo).Methods("PUT")

	// ─── Log call ──────────────────────────────────────────────────────────────
	api.HandleFunc("/{id:[0-9]+}/log-call", h.LogCall).Methods("POST")

	// ─── Medical history ───────────────────────────────────────────────────────
	api.HandleFunc("/patient/{id:[0-9]+}/medications", h.GetMedications).Methods("GET")
	api.HandleFunc("/patient/{id:[0-9]+}/allergy", h.GetAllergies).Methods("GET")
	api.HandleFunc("/patient/{id:[0-9]+}/diagnoses", h.GetDiagnoses).Methods("GET")
}
