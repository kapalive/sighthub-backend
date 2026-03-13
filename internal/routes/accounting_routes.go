package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	accHandler "sighthub-backend/internal/handlers/accounting_handler"
	accSvc "sighthub-backend/internal/services/accounting_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterAccountingRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := accSvc.New(db)
	h := accHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)

	api := r.PathPrefix("/api/accounting").Subrouter()

	// No auth
	api.HandleFunc("/vendors", h.GetVendors).Methods("GET")
	api.HandleFunc("/payments_methods", h.GetPaymentMethods).Methods("GET")
	api.HandleFunc("/stores", h.GetStores).Methods("GET")
	api.HandleFunc("/locations", h.GetLocations).Methods("GET")

	// JWT required
	auth := api.NewRoute().Subrouter()
	auth.Use(jwtMW)

	auth.HandleFunc("/{vendor_id:[0-9]+}/quickbooks-header", h.GetVendorQuickbooksHeader).Methods("GET")
	auth.HandleFunc("/vendor-invoices/{vendor_id:[0-9]+}", h.GetVendorInvoicesList).Methods("GET")
	auth.HandleFunc("/invoices/{vendor_id:[0-9]+}", h.GetInvoicesByVendor).Methods("GET")
	auth.HandleFunc("/vendor-bills", h.CreateVendorBill).Methods("POST")
	auth.HandleFunc("/transactions/{vendor_id:[0-9]+}", h.GetTransactionsByVendor).Methods("GET")
	auth.HandleFunc("/add_payment", h.AddPaymentToVendor).Methods("POST")
	auth.HandleFunc("/vendors-balances", h.GetVendorsBalances).Methods("GET")
	auth.HandleFunc("/vendors/{vendor_id:[0-9]+}/account-number", h.ListVendorLocationAccounts).Methods("GET")
	auth.HandleFunc("/vendors/{vendor_id:[0-9]+}/account-number", h.CreateVendorAccountNumber).Methods("POST")
	auth.HandleFunc("/vendors/{vendor_id:[0-9]+}/account-number/{acc_id:[0-9]+}", h.UpdateVendorAccountNumber).Methods("PUT")
	auth.HandleFunc("/vendors/{vendor_id:[0-9]+}/account-number/{acc_id:[0-9]+}", h.DeleteVendorAccountNumber).Methods("DELETE")
	auth.HandleFunc("/notify/terms", h.GetTermsNotifyList).Methods("GET")
	auth.HandleFunc("/return-to-vendor-invoices/{vendor_id:[0-9]+}", h.GetReturnToVendorInvoices).Methods("GET")
	auth.HandleFunc("/return-to-vendor-invoices/{rtv_id:[0-9]+}/credit", h.UpdateReturnToVendorCredit).Methods("PUT")
	auth.HandleFunc("/ledger/{vendor_id:[0-9]+}", h.GetVendorLedger).Methods("GET")
	auth.HandleFunc("/ledger/entry", h.GetLedgerEntry).Methods("GET")
}
