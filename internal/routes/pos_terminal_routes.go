package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	posH "sighthub-backend/internal/handlers/pos_terminal_handler"
	"sighthub-backend/internal/middleware"
	posSvc "sighthub-backend/internal/services/pos_terminal_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterPosTerminalRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := posSvc.New(db)
	h := posH.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)

	// Permission 64 = patient payments (read/use terminals)
	// Permission 80 = terminal management (create/update/delete/provision)
	perm64 := middleware.ActivePermission(db, 64)
	perm80 := middleware.ActivePermission(db, 80)

	api := r.PathPrefix("/api").Subrouter()
	api.Use(jwtMW)

	// ─── Terminal CRUD (perm 80 for management) ──────────────────────────────
	api.Handle("/pos/terminals",
		perm64(http.HandlerFunc(h.ListTerminals)),
	).Methods("GET")
	api.Handle("/pos/terminals",
		perm80(http.HandlerFunc(h.CreateTerminal)),
	).Methods("POST")
	api.Handle("/pos/terminals/{terminal_id:[0-9]+}",
		perm80(http.HandlerFunc(h.UpdateTerminal)),
	).Methods("PUT")
	api.Handle("/pos/terminals/{terminal_id:[0-9]+}/default",
		perm80(http.HandlerFunc(h.SetDefaultTerminal)),
	).Methods("POST")
	api.Handle("/pos/terminals/{terminal_id:[0-9]+}",
		perm80(http.HandlerFunc(h.DeleteTerminal)),
	).Methods("DELETE")

	// ─── POS Start / Commit (perm 64) ────────────────────────────────────────
	api.Handle("/invoice/{invoice_id:[0-9]+}/pos/start",
		perm64(http.HandlerFunc(h.PosStart)),
	).Methods("POST")
	api.Handle("/invoice/{invoice_id:[0-9]+}/pos/commit",
		perm64(http.HandlerFunc(h.PosCommit)),
	).Methods("POST")

	// ─── SPIn Config (perm 80) ───────────────────────────────────────────────
	api.Handle("/pos/spin-config/provision",
		perm80(http.HandlerFunc(h.ProvisionSpinConfig)),
	).Methods("POST")
	api.Handle("/pos/spin-config",
		perm80(http.HandlerFunc(h.GetSpinConfig)),
	).Methods("GET")

	// ─── Transaction detail (perm 64) ────────────────────────────────────────
	api.Handle("/pos/tx/{tx_id:[0-9]+}",
		perm64(http.HandlerFunc(h.GetTransaction)),
	).Methods("GET")
}
