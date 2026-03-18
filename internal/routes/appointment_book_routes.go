package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	apptH "sighthub-backend/internal/handlers/appointment_book_handler/appointment"
	settApptH "sighthub-backend/internal/handlers/settings_handler/appointment"
	apptSvc "sighthub-backend/internal/services/appointment_service"
	settSvc "sighthub-backend/internal/services/settings_service"
	"sighthub-backend/internal/middleware"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterAppointmentBookRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	s := apptSvc.New(db)
	h := apptH.New(s, db)
	hSettings := settApptH.New(settSvc.New(db))

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)

	api := r.PathPrefix("/api/appointment-book").Subrouter()

	// ── StorePermission(12, 81): /location + /set_location ──────────────────
	storeR := api.PathPrefix("").Subrouter()
	storeR.Use(jwtMW, middleware.StorePermission(db, 12, 81))
	storeR.HandleFunc("/location", h.GetLocations).Methods("GET")
	storeR.HandleFunc("/set_location", h.SetLocation).Methods("PUT")

	// ── JWT + perm 6: read-only ──────────────────────────────────────────────
	readR := api.PathPrefix("").Subrouter()
	readR.Use(jwtMW, middleware.ActivePermission(db, 6))
	readR.HandleFunc("/appointment-duration", hSettings.GetAppointmentDuration).Methods("GET")
	readR.HandleFunc("/doctors", h.GetDoctors).Methods("GET")
	readR.HandleFunc("/status-appointments", h.GetStatusAppointments).Methods("GET")
	readR.HandleFunc("/appointments", h.GetAppointments).Methods("GET")
	readR.HandleFunc("/request-appointments", h.GetRequestAppointments).Methods("GET")
	readR.HandleFunc("/location/{id:[0-9]+}/work_hours", h.GetLocationWorkHours).Methods("GET")
	readR.HandleFunc("/doctor/{id:[0-9]+}/set_lunch", h.SetDoctorLunch).Methods("POST")

	// ── JWT + perm 8: create ─────────────────────────────────────────────────
	createR := api.PathPrefix("").Subrouter()
	createR.Use(jwtMW, middleware.ActivePermission(db, 8))
	createR.HandleFunc("/new-appointment", h.CreateAppointment).Methods("POST")
	createR.HandleFunc("/request-appointments/cancel", h.CancelRequestAppointment).Methods("POST")

	// ── JWT + perm 9: update ─────────────────────────────────────────────────
	updateR := api.PathPrefix("").Subrouter()
	updateR.Use(jwtMW, middleware.ActivePermission(db, 9))
	updateR.HandleFunc("/appointments/{id:[0-9]+}", h.UpdateAppointment).Methods("PUT")
	updateR.HandleFunc("/appointments/{id:[0-9]+}/status", h.UpdateAppointmentStatus).Methods("PUT")
	updateR.HandleFunc("/appointments/{id:[0-9]+}/insurance", h.UpdateAppointmentInsurance).Methods("PUT")
	updateR.HandleFunc("/send_sms/intake_form_link/{id:[0-9]+}", h.SendIntakeFormLink).Methods("POST")

	// ── JWT + perm 10: delete ────────────────────────────────────────────────
	deleteR := api.PathPrefix("").Subrouter()
	deleteR.Use(jwtMW, middleware.ActivePermission(db, 10))
	deleteR.HandleFunc("/appointments/{id:[0-9]+}", h.DeleteAppointment).Methods("DELETE")
}
