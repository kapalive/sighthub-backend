package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	claimHandler "sighthub-backend/internal/handlers/claim_handler"
	claimSvc "sighthub-backend/internal/services/claim_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterClaimRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := claimSvc.New(db)
	h := claimHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)

	api := r.PathPrefix("/api/claim").Subrouter()
	api.Use(jwtMW)

	// Invoices
	api.HandleFunc("/invoices", h.GetInsuranceInvoices).Methods("GET")
	api.HandleFunc("/insurance-companies", h.GetInsuranceCompanies).Methods("GET")
	api.HandleFunc("/insurance-coverage-types", h.GetInsuranceCoverageTypes).Methods("GET")
	api.HandleFunc("/insurance-payment-types", h.GetInsurancePaymentTypes).Methods("GET")
	api.HandleFunc("/invoices/{invoice_id:[0-9]+}/insurance-payment", h.GetInvoicePaymentSummary).Methods("GET")
	api.HandleFunc("/invoices/{invoice_id:[0-9]+}/insurance-payment", h.AddInsurancePayment).Methods("POST")
	api.HandleFunc("/invoices/{invoice_id:[0-9]+}/patient", h.GetInvoicePatient).Methods("GET")
	api.HandleFunc("/invoices/{invoice_id:[0-9]+}/responsible-party", h.GetResponsibleParty).Methods("GET")
	api.HandleFunc("/invoices/{invoice_id:[0-9]+}/insurance-status", h.UpdateInsuranceStatus).Methods("PUT")
	api.HandleFunc("/invoices/{invoice_id:[0-9]+}/secondary-insurance", h.GetSecondaryInsurance).Methods("GET")
	api.HandleFunc("/invoices/{invoice_id:[0-9]+}/insurance-info", h.GetInvoiceInsuranceInfo).Methods("GET")
	api.HandleFunc("/invoices/{invoice_id:[0-9]+}/claim-info", h.GetClaimInfo).Methods("GET")

	// Super bill
	api.HandleFunc("/super-bill/{invoice_id:[0-9]+}", h.GetSuperBill).Methods("GET")
	api.HandleFunc("/super-bill/{invoice_id:[0-9]+}", h.UpdateSuperBill).Methods("PUT")

	// Doctors
	api.HandleFunc("/doctors", h.GetDoctors).Methods("GET")

	// Claim templates
	api.HandleFunc("/templates", h.ListClaimTemplates).Methods("GET")
	api.HandleFunc("/templates", h.CreateClaimTemplate).Methods("POST")
	api.HandleFunc("/templates/{template_id:[0-9]+}", h.GetClaimTemplate).Methods("GET")
	api.HandleFunc("/templates/{template_id:[0-9]+}", h.UpdateClaimTemplate).Methods("PUT")
	api.HandleFunc("/templates/{template_id:[0-9]+}", h.DeleteClaimTemplate).Methods("DELETE")
	api.HandleFunc("/templates/{template_id:[0-9]+}/render", h.RenderClaimPDF).Methods("POST")
	api.HandleFunc("/templates/{template_id:[0-9]+}/pdf", h.GetTemplatePDF).Methods("GET")
	api.HandleFunc("/templates/{template_id:[0-9]+}/preview", h.PreviewTemplatePDF).Methods("GET")
}
