package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	helpdeskH "sighthub-backend/internal/handlers/helpdesk_handler"
	helpdeskSvc "sighthub-backend/internal/services/helpdesk_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterHelpdeskRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	s := helpdeskSvc.New(db, cfg)
	h := helpdeskH.New(s)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)

	api := r.PathPrefix("/api/helpdesk").Subrouter()

	// JWT-protected routes
	jwtR := api.PathPrefix("").Subrouter()
	jwtR.Use(jwtMW)
	jwtR.HandleFunc("", h.ListTickets).Methods("GET")
	jwtR.HandleFunc("/", h.ListTickets).Methods("GET")
	jwtR.HandleFunc("", h.CreateTicket).Methods("POST")
	jwtR.HandleFunc("/", h.CreateTicket).Methods("POST")
	jwtR.HandleFunc("/{ticket_id:[0-9]+}", h.GetTicket).Methods("GET")

	// Webhook from external helpdesk — HMAC only, no JWT
	api.HandleFunc("/{ticket_id:[0-9]+}", h.ReceiveReply).Methods("POST")
}
