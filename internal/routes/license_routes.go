package routes

import (
	"github.com/gorilla/mux"
	"gorm.io/gorm"

	licenseH "sighthub-backend/internal/handlers/license_handler"
	licenseSvc "sighthub-backend/internal/services/license_service"
)

func RegisterLicenseRoutes(db *gorm.DB, r *mux.Router) {
	s := licenseSvc.New(db)
	h := licenseH.New(s)

	api := r.PathPrefix("/api/license").Subrouter()
	api.HandleFunc("/kms/store", h.KMSStore).Methods("POST")
}
