package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	invoiceH "sighthub-backend/internal/handlers/invoice_handler"
	"sighthub-backend/internal/middleware"
	invoiceSvc "sighthub-backend/internal/services/invoice_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterInvoiceRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := invoiceSvc.New(db)
	h := invoiceH.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)

	// permission IDs: 11=base, 17=read, 18=create, 19=update/confirm, 20=delete
	//                 21=read return, 22=create return, 23=update return/payment, 24=delete return
	baseMW         := middleware.ActivePermission(db, 11)
	readMW         := middleware.ActivePermission(db, 17)
	createMW       := middleware.ActivePermission(db, 18)
	updateMW       := middleware.ActivePermission(db, 19)
	deleteMW       := middleware.ActivePermission(db, 20)
	readReturnMW   := middleware.ActivePermission(db, 21)
	createReturnMW := middleware.ActivePermission(db, 22)
	updateReturnMW := middleware.ActivePermission(db, 23)
	deleteReturnMW := middleware.ActivePermission(db, 24)

	api := r.PathPrefix("/api/invoice").Subrouter()
	api.Use(jwtMW)
	api.Use(baseMW)

	// ─── Read ─────────────────────────────────────────────────────────────────
	read := api.NewRoute().Subrouter()
	read.Use(readMW)
	read.HandleFunc("/search", h.SearchInvoice).Methods("GET")
	read.HandleFunc("/vendors", h.GetVendors).Methods("GET")
	read.HandleFunc("/locations", h.GetLocations).Methods("GET")
	read.HandleFunc("/view/{invoice_id}", h.ViewInvoice).Methods("GET")
	read.HandleFunc("/view-item/{invoice_id}", h.ViewInvoiceItem).Methods("GET")
	read.HandleFunc("/receipt", h.GetReceipts).Methods("GET")
	read.HandleFunc("/receipt/{invoice_id}", h.GetReceipt).Methods("GET")
	read.HandleFunc("/vendor-contacts", h.GetVendorContacts).Methods("GET")
	read.HandleFunc("/location-contacts", h.GetLocationContacts).Methods("GET")
	read.HandleFunc("/shipments", h.GetShipments).Methods("GET")
	read.HandleFunc("/transfers", h.GetTransfers).Methods("GET")
	read.HandleFunc("/shipment/{shipment_id}", h.GetShipment).Methods("GET")
	read.HandleFunc("/vendor_invoice/{id}", h.GetVendorInvoice).Methods("GET")

	// ─── Create ───────────────────────────────────────────────────────────────
	create := api.NewRoute().Subrouter()
	create.Use(createMW)
	create.HandleFunc("/create", h.CreateInvoice).Methods("POST")
	create.HandleFunc("/shipment/create", h.CreateShipment).Methods("POST")
	create.HandleFunc("/vendor_invoices/create", h.CreateVendorInvoice).Methods("POST")

	// ─── Update / Confirm ─────────────────────────────────────────────────────
	update := api.NewRoute().Subrouter()
	update.Use(updateMW)
	update.HandleFunc("/update/{invoice_id}", h.UpdateInvoice).Methods("PUT")
	update.HandleFunc("/receipt/confirm", h.ConfirmReceipt).Methods("PUT")
	update.HandleFunc("/receipt/{invoice_id}/pay", h.PayTransfer).Methods("POST")
	update.HandleFunc("/shipment/update/{shipment_id}", h.UpdateShipment).Methods("PUT")
	update.HandleFunc("/vendor_invoices/{id}", h.UpdateVendorInvoice).Methods("PUT")

	// ─── Delete ───────────────────────────────────────────────────────────────
	del := api.NewRoute().Subrouter()
	del.Use(deleteMW)
	del.HandleFunc("/delete-item", h.DeleteItem).Methods("DELETE")
	del.HandleFunc("/delete/{invoice_id}", h.DeleteInvoice).Methods("DELETE")

	// ─── Return Invoices — Read ───────────────────────────────────────────────
	readReturn := api.NewRoute().Subrouter()
	readReturn.Use(readReturnMW)
	readReturn.HandleFunc("/return_invoices", h.GetReturnInvoices).Methods("GET")
	readReturn.HandleFunc("/return_invoice/shipping-services", h.GetShippingServices).Methods("GET")
	readReturn.HandleFunc("/return_invoice/payment-methods", h.GetReturnPaymentMethods).Methods("GET")
	readReturn.HandleFunc("/return_invoice/{id}/payments", h.GetReturnPayments).Methods("GET")
	readReturn.HandleFunc("/return_invoice/{id}", h.GetReturnInvoice).Methods("GET")

	// ─── Return Invoices — Create ─────────────────────────────────────────────
	createReturn := api.NewRoute().Subrouter()
	createReturn.Use(createReturnMW)
	createReturn.HandleFunc("/return_invoice", h.CreateReturnInvoice).Methods("POST")

	// ─── Return Invoices — Update / Payment ──────────────────────────────────
	updateReturn := api.NewRoute().Subrouter()
	updateReturn.Use(updateReturnMW)
	updateReturn.HandleFunc("/return_invoice/{id}", h.UpdateReturnInvoice).Methods("PUT")
	updateReturn.HandleFunc("/return_invoice/{id}/payment", h.AddReturnPayment).Methods("POST")
	updateReturn.HandleFunc("/return_invoice/{id}/payment/{payment_id}", h.DeleteReturnPayment).Methods("DELETE")

	// ─── Return Invoices — Delete ─────────────────────────────────────────────
	deleteReturn := api.NewRoute().Subrouter()
	deleteReturn.Use(deleteReturnMW)
	deleteReturn.HandleFunc("/return_invoice/{id}", h.DeleteReturnInvoice).Methods("DELETE")
}
