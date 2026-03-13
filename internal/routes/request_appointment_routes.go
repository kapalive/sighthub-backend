package routes

import (
	"github.com/gorilla/mux"
	"gorm.io/gorm"

	reqApptH "sighthub-backend/internal/handlers/request_appointment_handler"
	reqApptSvc "sighthub-backend/internal/services/request_appointment_service"
)

func RegisterRequestAppointmentRoutes(db *gorm.DB, r *mux.Router) {
	s := reqApptSvc.New(db)
	h := reqApptH.New(s)

	api := r.PathPrefix("/api/request-appointment").Subrouter()

	// All endpoints are public (no JWT) — called by patients from the browser
	api.HandleFunc("/locations", h.GetLocations).Methods("GET")
	api.HandleFunc("/doctors", h.GetDoctors).Methods("GET")
	api.HandleFunc("/doctor-slot-availability", h.GetDoctorSlotAvailability).Methods("GET")
	api.HandleFunc("/request-appointment", h.CreateRequestAppointment).Methods("POST")
	api.HandleFunc("/intake-form", h.CreateIntakeForm).Methods("POST")
	api.HandleFunc("/intake-form/{id:[0-9]+}", h.GetIntakeForm).Methods("GET")
	api.HandleFunc("/intake-form/{id:[0-9]+}", h.UpdateIntakeForm).Methods("PUT")
	api.HandleFunc("/check-appointment/{id:[0-9]+}", h.CheckAppointment).Methods("GET")
}
