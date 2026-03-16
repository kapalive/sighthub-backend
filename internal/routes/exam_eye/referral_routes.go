package exameye

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	referralHandler "sighthub-backend/internal/handlers/referral_handler"
	"sighthub-backend/internal/middleware"
	referralSvc "sighthub-backend/internal/services/referral_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterReferralRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := referralSvc.New(db)
	h := referralHandler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	perm1 := middleware.ActivePermission(db, 1)
	perm3 := middleware.ActivePermission(db, 3)

	api := r.PathPrefix("/api/exam_eye/referral").Subrouter()
	api.Use(jwtMW, perm1)

	api.Handle("/{exam_id:[0-9]+}/letter", perm3(http.HandlerFunc(h.SaveReferral))).Methods("POST")
	api.HandleFunc("/{exam_id:[0-9]+}", h.GetReferral).Methods("GET")
	api.HandleFunc("/{exam_id:[0-9]+}/letter/{letter_id:[0-9]+}", h.GetReferralLetterByID).Methods("GET")
	api.HandleFunc("/{exam_id:[0-9]+}/letter/{letter_id:[0-9]+}", h.UpdateReferralLetter).Methods("PUT")
	api.HandleFunc("/{exam_id:[0-9]+}/letter/{letter_id:[0-9]+}", h.DeleteReferralLetterOrDoctor).Methods("DELETE")
	api.HandleFunc("/doctors", h.GetAllReferralDoctors).Methods("GET")
	api.HandleFunc("/doctors", h.CreateReferralDoctor).Methods("POST")
	api.HandleFunc("/doctors/{doctor_id:[0-9]+}", h.UpdateReferralDoctor).Methods("PUT")
	api.HandleFunc("/doctors/{doctor_id:[0-9]+}", h.DeleteReferralDoctor).Methods("DELETE")
	api.HandleFunc("/referral-letters/{letter_id:[0-9]+}/html", h.PrintReferralLetter).Methods("GET")
	api.HandleFunc("/referral-letters/{letter_id:[0-9]+}/fax", h.FaxReferralLetter).Methods("POST")
	api.HandleFunc("/referral-letters/{letter_id:[0-9]+}/email", h.EmailReferralLetter).Methods("POST")
	api.HandleFunc("/tests-build/{exam_id:[0-9]+}", h.BuildTests).Methods("GET")
}
