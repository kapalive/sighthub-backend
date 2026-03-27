package routes

import (
	"github.com/gorilla/mux"
	"gorm.io/gorm"

	"sighthub-backend/config"
	"sighthub-backend/internal/handlers/orders_handler"
	"sighthub-backend/internal/middleware"
	pkgAuth "sighthub-backend/pkg/auth"

	"github.com/redis/go-redis/v9"
)

func RegisterOrdersRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	h := orders_handler.New(db)
	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	permMW := middleware.StorePermission(db, 14, 62) // block=14 (Invoice), perm=62 (Add Item)

	api := r.PathPrefix("/api/orders").Subrouter()
	api.Use(jwtMW, permMW)

	api.HandleFunc("/order-list", h.OrderList).Methods("GET")
	api.HandleFunc("/locations", h.Locations).Methods("GET")
	api.HandleFunc("/ticket-statuses", h.TicketStatuses).Methods("GET")
	api.HandleFunc("/invoice-statuses", h.InvoiceStatuses).Methods("GET")
	api.HandleFunc("/lab-status", h.LabStatus).Methods("GET")
	api.HandleFunc("/contact-status", h.ContactStatus).Methods("GET")
	api.HandleFunc("/invoice-status", h.InvoiceStatus).Methods("GET")
	api.HandleFunc("/sms-templates", h.SMSTemplates).Methods("GET")
	api.HandleFunc("/status-notification-map", h.StatusNotificationMap).Methods("GET")
	api.HandleFunc("/send-sms", h.SendSMS).Methods("POST")
	api.HandleFunc("/notify", h.NotifyPatient).Methods("POST")
	api.HandleFunc("/ticket/{id:[0-9]+}/dashboard-note", h.UpdateTicketNote).Methods("POST")
	api.HandleFunc("/invoice/{id:[0-9]+}/dashboard-note", h.UpdateInvoiceNote).Methods("POST")
}
