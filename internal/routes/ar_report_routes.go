package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	arHandler "sighthub-backend/internal/handlers/ar_report_handler"
	"sighthub-backend/internal/middleware"
	arSvc "sighthub-backend/internal/services/ar_report_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterARReportRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := arSvc.New(db)
	h := arHandler.New(svc)

	jwtMW  := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	perm33 := middleware.ActivePermission(db, 33)
	perm34 := middleware.ActivePermission(db, 34)
	perm35 := middleware.ActivePermission(db, 35)

	api := r.PathPrefix("/api/ar").Subrouter()
	api.Use(jwtMW, perm33)

	// ─── Balance due / Credits ────────────────────────────────────────────────
	api.Handle("/balance_due", perm34(http.HandlerFunc(h.GetBalanceDue))).Methods("GET")
	api.Handle("/credits",     perm35(http.HandlerFunc(h.GetCredits))).Methods("GET")

	// ─── Count sheets (list/create/delete share the same path) ───────────────
	api.Handle("/count_sheets",        perm34(http.HandlerFunc(h.GetCountSheets))).Methods("GET")
	api.Handle("/count_sheets",        perm34(http.HandlerFunc(h.CreateCountSheet))).Methods("POST")
	api.Handle("/count_sheets",        perm34(http.HandlerFunc(h.DeleteCountSheet))).Methods("DELETE")

	// ─── Count sheet items (query-param based) ────────────────────────────────
	api.Handle("/count_sheets/items",  perm34(http.HandlerFunc(h.GetCountSheetItems))).Methods("GET")
	api.Handle("/count_sheets/items",  perm34(http.HandlerFunc(h.AddInvoiceToCountSheet))).Methods("POST")

	// ─── Single count sheet CRUD ──────────────────────────────────────────────
	api.Handle("/count_sheets/{id:[0-9]+}",       perm34(http.HandlerFunc(h.GetCountSheetInfo))).Methods("GET")
	api.Handle("/count_sheets/{id:[0-9]+}",       perm34(http.HandlerFunc(h.UpdateCountSheetNotes))).Methods("PUT")
	api.Handle("/count_sheets/{id:[0-9]+}/close", perm34(http.HandlerFunc(h.CloseCountSheet))).Methods("POST")
}
