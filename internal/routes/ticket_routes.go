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

	// ── GET endpoints ───────────────────────────────────────────────────
	api.HandleFunc("/", h.ListTickets).Methods("GET")
	api.HandleFunc("/search", h.SearchTickets).Methods("GET")
	api.HandleFunc("/statuses", h.GetStatuses).Methods("GET")
	api.HandleFunc("/lens-statuses", h.GetLensStatuses).Methods("GET")
	api.HandleFunc("/labs", h.GetLabs).Methods("GET")
	api.HandleFunc("/frame-type-materials", h.GetFrameTypeMaterials).Methods("GET")
	api.HandleFunc("/frame_shapes", h.GetFrameShapes).Methods("GET")
	api.HandleFunc("/vendor_brand", h.GetVendorBrand).Methods("GET")
	api.HandleFunc("/products", h.GetProducts).Methods("GET")
	api.HandleFunc("/lens/type", h.GetLensTypes).Methods("GET")
	api.HandleFunc("/lens/materials", h.GetLensMaterials).Methods("GET")
	api.HandleFunc("/lens/materials", h.AddLensMaterial).Methods("POST")
	api.HandleFunc("/lens/style", h.GetLensStyles).Methods("GET")
	api.HandleFunc("/lens/lens_tint_color", h.GetLensTintColors).Methods("GET")
	api.HandleFunc("/lens/lens_sample_color", h.GetLensSampleColors).Methods("GET")
	api.HandleFunc("/lens/safety_thickness", h.GetLensSafetyThicknesses).Methods("GET")
	api.HandleFunc("/lens/lens_bevel", h.GetLensBevels).Methods("GET")
	api.HandleFunc("/lens/lens_edge", h.GetLensEdges).Methods("GET")
	api.HandleFunc("/lens/series", h.GetLensSeries).Methods("GET")
	api.HandleFunc("/contact/services", h.GetContactServices).Methods("GET")
	api.HandleFunc("/contact_lens/brands", h.GetContactLensBrands).Methods("GET")
	api.HandleFunc("/invoice/{invoice_id:[0-9]+}", h.GetTicketsByInvoice).Methods("GET")
	api.HandleFunc("/{id_lab_ticket:[0-9]+}", h.GetTicketByID).Methods("GET")

	// ── POST / PUT endpoints ────────────────────────────────────────────
	api.HandleFunc("/invoice/{invoice_id:[0-9]+}", h.CreateTicket).Methods("POST")
	api.HandleFunc("/{id_lab_ticket:[0-9]+}", h.UpdateTicket).Methods("PUT")
	api.HandleFunc("/{ticket_id:[0-9]+}/notify-patient", h.NotifyPatient).Methods("POST")
}
