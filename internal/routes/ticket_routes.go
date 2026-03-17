package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	ticketHandler "sighthub-backend/internal/handlers/ticket_handler"
	ticketSvc "sighthub-backend/internal/services/ticket_service"
	pkgAuth "sighthub-backend/pkg/auth"
	"sighthub-backend/pkg/scheduler"
)

func RegisterTicketRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, sched *scheduler.Scheduler, r *mux.Router) {
	svc := ticketSvc.New(db, sched)
	h := ticketHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)

	api := r.PathPrefix("/api/ticket").Subrouter()
	api.Use(jwtMW)

	api.HandleFunc("/{ticket_id:[0-9]+}/notify-patient", h.NotifyPatient).Methods("POST")
}
