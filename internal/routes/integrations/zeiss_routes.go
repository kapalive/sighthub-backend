package integrations

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	zHandler "sighthub-backend/internal/handlers/zeiss_handler"
	"sighthub-backend/internal/middleware"
	zSvc "sighthub-backend/internal/services/zeiss_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterZeissRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := zSvc.New(db)
	h := zHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	perm1 := middleware.ActivePermission(db, 1)

	api := r.PathPrefix("/api/integrations/zeiss").Subrouter()

	// oauth2redirect is called by ZEISS — no JWT required
	api.HandleFunc("/oauth2redirect", h.OAuth2Redirect).Methods("GET")

	// all other endpoints require auth
	secured := api.PathPrefix("").Subrouter()
	secured.Use(jwtMW, perm1)

	secured.HandleFunc("/auth-url", h.AuthURL).Methods("GET")
	secured.HandleFunc("/exchange", h.Exchange).Methods("POST")
	secured.HandleFunc("/token", h.Token).Methods("GET")
	secured.HandleFunc("/refresh", h.Refresh).Methods("POST")
	secured.HandleFunc("/logout", h.Logout).Methods("POST")
	secured.HandleFunc("/", h.Index).Methods("GET")
	secured.HandleFunc("", h.Index).Methods("GET")
}
