package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	questionnaireH "sighthub-backend/internal/handlers/questionnaire_handler"
	questionnaireSvc "sighthub-backend/internal/services/questionnaire_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterQuestionnaireRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	s := questionnaireSvc.New(db)
	h := questionnaireH.New(s, db)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)

	api := r.PathPrefix("/api/questionnaire").Subrouter()
	api.Use(jwtMW)

	api.HandleFunc("/referral", h.CreateReferral).Methods("POST")
	api.HandleFunc("/referral_sources", h.GetReferralSources).Methods("GET")
	api.HandleFunc("/visit_reasons", h.GetVisitReasons).Methods("GET")
}
