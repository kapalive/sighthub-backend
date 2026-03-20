package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	"sighthub-backend/internal/handlers/settings_handler/appointment"
	"sighthub-backend/internal/handlers/settings_handler/frame"
	"sighthub-backend/internal/handlers/settings_handler/insurance"
	"sighthub-backend/internal/handlers/settings_handler/jobtitle"
	"sighthub-backend/internal/handlers/settings_handler/lens"
	"sighthub-backend/internal/handlers/settings_handler/notify"
	"sighthub-backend/internal/handlers/settings_handler/payment"
	"sighthub-backend/internal/handlers/settings_handler/professional"
	"sighthub-backend/internal/handlers/settings_handler/smtp"
	"sighthub-backend/internal/handlers/settings_handler/ticket"
	"sighthub-backend/internal/handlers/settings_handler/vendor"
	"net/http"

	homeHandler "sighthub-backend/internal/handlers/home_handler"
	intHandler "sighthub-backend/internal/handlers/integration_handler"
	"sighthub-backend/internal/middleware"
	intSvc "sighthub-backend/internal/services/integration_service"
	"sighthub-backend/internal/services/settings_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterSettingsRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := settings_service.New(db)

	hSMTP := smtp.New(svc)
	hAppt := appointment.New(svc)
	hLens := lens.New(svc)
	hFrame := frame.New(svc)
	hIns := insurance.New(svc)
	hVendor := vendor.New(svc)
	hProf := professional.New(svc)
	hTicket := ticket.New(svc)
	hNotify := notify.New(svc)
	hPay := payment.New(svc)
	hJob := jobtitle.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	baseMW := middleware.ActivePermission(db, 80)

	api := r.PathPrefix("/api/settings").Subrouter()
	api.Use(jwtMW, baseMW)

	// ── SMTP ───────────────────────────────────────────────────────────────
	api.HandleFunc("/smtp-config", hSMTP.List).Methods("GET")
	api.HandleFunc("/smtp-config/{id:[0-9]+}", hSMTP.Get).Methods("GET")
	api.HandleFunc("/smtp-config", hSMTP.Create).Methods("POST")
	api.HandleFunc("/smtp-config/{id:[0-9]+}", hSMTP.Update).Methods("PUT")
	api.HandleFunc("/smtp-test/{id:[0-9]+}", hSMTP.Test).Methods("POST")

	// ── Appointment Reasons ────────────────────────────────────────────────
	api.HandleFunc("/appointment/reasons", hAppt.ListReasons).Methods("GET")
	api.HandleFunc("/appointment/reasons", hAppt.CreateReason).Methods("POST")
	api.HandleFunc("/appointment/reasons/{id:[0-9]+}", hAppt.UpdateReason).Methods("PUT")
	api.HandleFunc("/appointment/reasons/{id:[0-9]+}", hAppt.DeleteReason).Methods("DELETE")

	// ── Locations (showcase) ───────────────────────────────────────────────
	api.HandleFunc("/locations", hAppt.GetShowcaseLocations).Methods("GET")

	// ── Location Appointment Settings ──────────────────────────────────────
	api.HandleFunc("/request-appointment", hAppt.SetRequestAppointment).Methods("POST")
	api.HandleFunc("/request-appointment", hAppt.GetRequestAppointmentSettings).Methods("GET")
	api.HandleFunc("/intake-form", hAppt.SetIntakeForm).Methods("POST")
	api.HandleFunc("/intake-form", hAppt.GetIntakeFormSettings).Methods("GET")
	api.HandleFunc("/appointment-duration", hAppt.SetAppointmentDuration).Methods("POST")
	api.HandleFunc("/appointment-duration", hAppt.GetAppointmentDuration).Methods("GET")

	// ── Lens Types ─────────────────────────────────────────────────────────
	api.HandleFunc("/lens/type", hLens.ListTypes).Methods("GET")
	api.HandleFunc("/lens/type", hLens.CreateType).Methods("POST")
	api.HandleFunc("/lens/type/{id:[0-9]+}", hLens.UpdateType).Methods("PUT")
	api.HandleFunc("/lens/type/{id:[0-9]+}", hLens.DeleteType).Methods("DELETE")

	// ── Lens Materials ─────────────────────────────────────────────────────
	api.HandleFunc("/lens/materials", hLens.ListMaterials).Methods("GET")
	api.HandleFunc("/lens/materials", hLens.CreateMaterial).Methods("POST")
	api.HandleFunc("/lens/materials/{id:[0-9]+}", hLens.UpdateMaterial).Methods("PUT")
	api.HandleFunc("/lens/materials/{id:[0-9]+}", hLens.DeleteMaterial).Methods("DELETE")

	// ── Lens Special Features ──────────────────────────────────────────────
	api.HandleFunc("/lens/special", hLens.ListSpecialFeatures).Methods("GET")
	api.HandleFunc("/lens/special", hLens.CreateSpecialFeature).Methods("POST")
	api.HandleFunc("/lens/special/{id:[0-9]+}", hLens.UpdateSpecialFeature).Methods("PUT")
	api.HandleFunc("/lens/special/{id:[0-9]+}", hLens.DeleteSpecialFeature).Methods("DELETE")

	// ── Lens Series ────────────────────────────────────────────────────────
	api.HandleFunc("/lens/series", hLens.ListSeries).Methods("GET")
	api.HandleFunc("/lens/series", hLens.CreateSeries).Methods("POST")
	api.HandleFunc("/lens/series/{id:[0-9]+}", hLens.UpdateSeries).Methods("PUT")
	api.HandleFunc("/lens/series/{id:[0-9]+}", hLens.DeleteSeries).Methods("DELETE")

	// ── VCodes ─────────────────────────────────────────────────────────────
	api.HandleFunc("/lens/v_codes", hLens.ListVCodes).Methods("GET")
	api.HandleFunc("/lens/v_codes", hLens.CreateVCode).Methods("POST")
	api.HandleFunc("/lens/v_codes/{id:[0-9]+}", hLens.UpdateVCode).Methods("PUT")
	api.HandleFunc("/lens/v_codes/{id:[0-9]+}", hLens.DeleteVCode).Methods("DELETE")

	// ── Lens Styles ────────────────────────────────────────────────────────
	api.HandleFunc("/lens/style", hLens.ListStyles).Methods("GET")
	api.HandleFunc("/lens/style", hLens.CreateStyle).Methods("POST")
	api.HandleFunc("/lens/style/{id:[0-9]+}", hLens.UpdateStyle).Methods("PUT")
	api.HandleFunc("/lens/style/{id:[0-9]+}", hLens.DeleteStyle).Methods("DELETE")

	// ── Tint Colors ────────────────────────────────────────────────────────
	api.HandleFunc("/lens/lens_tint_color", hLens.ListTintColors).Methods("GET")
	api.HandleFunc("/lens/lens_tint_color", hLens.CreateTintColor).Methods("POST")
	api.HandleFunc("/lens/lens_tint_color/{id:[0-9]+}", hLens.UpdateTintColor).Methods("PUT")
	api.HandleFunc("/lens/lens_tint_color/{id:[0-9]+}", hLens.DeleteTintColor).Methods("DELETE")

	// ── Sample Colors ──────────────────────────────────────────────────────
	api.HandleFunc("/lens/lens_sample_color", hLens.ListSampleColors).Methods("GET")
	api.HandleFunc("/lens/lens_sample_color", hLens.CreateSampleColor).Methods("POST")
	api.HandleFunc("/lens/lens_sample_color/{id:[0-9]+}", hLens.UpdateSampleColor).Methods("PUT")
	api.HandleFunc("/lens/lens_sample_color/{id:[0-9]+}", hLens.DeleteSampleColor).Methods("DELETE")

	// ── Safety Thickness ───────────────────────────────────────────────────
	api.HandleFunc("/lens/safety_thickness", hLens.ListSafetyThickness).Methods("GET")
	api.HandleFunc("/lens/safety_thickness", hLens.CreateSafetyThickness).Methods("POST")
	api.HandleFunc("/lens/safety_thickness/{id:[0-9]+}", hLens.UpdateSafetyThickness).Methods("PUT")
	api.HandleFunc("/lens/safety_thickness/{id:[0-9]+}", hLens.DeleteSafetyThickness).Methods("DELETE")

	// ── Bevels ─────────────────────────────────────────────────────────────
	api.HandleFunc("/lens/lens_bevel", hLens.ListBevels).Methods("GET")
	api.HandleFunc("/lens/lens_bevel", hLens.CreateBevel).Methods("POST")
	api.HandleFunc("/lens/lens_bevel/{id:[0-9]+}", hLens.UpdateBevel).Methods("PUT")
	api.HandleFunc("/lens/lens_bevel/{id:[0-9]+}", hLens.DeleteBevel).Methods("DELETE")

	// ── Lens Statuses ──────────────────────────────────────────────────────
	api.HandleFunc("/lens-statuses", hLens.ListLensStatuses).Methods("GET")
	api.HandleFunc("/lens-statuses", hLens.CreateLensStatus).Methods("POST")
	api.HandleFunc("/lens-statuses/{id:[0-9]+}", hLens.UpdateLensStatus).Methods("PUT")
	api.HandleFunc("/lens-statuses/{id:[0-9]+}", hLens.DeleteLensStatus).Methods("DELETE")

	// ── Frame Shapes ───────────────────────────────────────────────────────
	api.HandleFunc("/frame_shapes", hFrame.ListShapes).Methods("GET")
	api.HandleFunc("/frame_shapes", hFrame.CreateShape).Methods("POST")
	api.HandleFunc("/frame_shapes/{id:[0-9]+}", hFrame.UpdateShape).Methods("PUT")
	api.HandleFunc("/frame_shapes/{id:[0-9]+}", hFrame.DeleteShape).Methods("DELETE")

	// ── Frame Type Materials ───────────────────────────────────────────────
	api.HandleFunc("/frame-type-materials", hFrame.ListTypeMaterials).Methods("GET")
	api.HandleFunc("/frame-type-materials", hFrame.CreateTypeMaterial).Methods("POST")
	api.HandleFunc("/frame-type-materials/{id:[0-9]+}", hFrame.UpdateTypeMaterial).Methods("PUT")
	api.HandleFunc("/frame-type-materials/{id:[0-9]+}", hFrame.DeleteTypeMaterial).Methods("DELETE")

	// ── Professional Service Scopes ────────────────────────────────────────
	api.HandleFunc("/professional_service_scopes", hProf.ListScopes).Methods("GET")
	api.HandleFunc("/professional_service_scopes", hProf.CreateScope).Methods("POST")
	api.HandleFunc("/professional_service_scopes/{id:[0-9]+}", hProf.UpdateScope).Methods("PUT")
	api.HandleFunc("/professional_service_scopes/{id:[0-9]+}", hProf.DeleteScope).Methods("DELETE")

	// ── Additional Service Types ───────────────────────────────────────────
	api.HandleFunc("/additional_types", hProf.ListAddTypes).Methods("GET")
	api.HandleFunc("/additional_types", hProf.CreateAddType).Methods("POST")
	api.HandleFunc("/additional_types/{id:[0-9]+}", hProf.UpdateAddType).Methods("PUT")
	api.HandleFunc("/additional_types/{id:[0-9]+}", hProf.DeleteAddType).Methods("DELETE")

	// ── Insurance Companies ────────────────────────────────────────────────
	api.HandleFunc("/insurance/companies", hIns.ListCompanies).Methods("GET")
	api.HandleFunc("/insurance/companies", hIns.CreateCompany).Methods("POST")
	api.HandleFunc("/insurance/companies/{id:[0-9]+}", hIns.UpdateCompany).Methods("PUT")
	api.HandleFunc("/insurance/companies/{id:[0-9]+}", hIns.DeleteCompany).Methods("DELETE")

	// ── Insurance Coverage Types ───────────────────────────────────────────
	api.HandleFunc("/insurance/coverage_types", hIns.ListCoverageTypes).Methods("GET")

	// ── Insurance Types ────────────────────────────────────────────────────
	api.HandleFunc("/insurance/types", hIns.ListInsuranceTypes).Methods("GET")

	// ── Insurance Payment Types ────────────────────────────────────────────
	api.HandleFunc("/insurance-payment-types", hIns.ListPaymentTypes).Methods("GET")
	api.HandleFunc("/insurance-payment-types", hIns.CreatePaymentType).Methods("POST")
	api.HandleFunc("/insurance-payment-types/{id:[0-9]+}", hIns.UpdatePaymentType).Methods("PUT")
	api.HandleFunc("/insurance-payment-types/{id:[0-9]+}", hIns.DeletePaymentType).Methods("DELETE")

	// ── Vendor Brands ──────────────────────────────────────────────────────
	api.HandleFunc("/vendor-brands", hVendor.ListBrands).Methods("GET")

	// ── Ticket Statuses ────────────────────────────────────────────────────
	api.HandleFunc("/ticket-statuses", hTicket.List).Methods("GET")
	api.HandleFunc("/ticket-statuses", hTicket.Create).Methods("POST")
	api.HandleFunc("/ticket-statuses/{id:[0-9]+}", hTicket.Update).Methods("PUT")
	api.HandleFunc("/ticket-statuses/{id:[0-9]+}", hTicket.Delete).Methods("DELETE")

	// ── Notify Settings ────────────────────────────────────────────────────
	api.HandleFunc("/notify-settings/{action}", hNotify.Get).Methods("GET")
	api.HandleFunc("/notify-settings/{action}", hNotify.Upsert).Methods("PUT")

	// ── Payment Methods ────────────────────────────────────────────────────
	api.HandleFunc("/payment-methods", hPay.List).Methods("GET")
	api.HandleFunc("/payment-methods", hPay.Create).Methods("POST")
	api.HandleFunc("/payment-methods/{id:[0-9]+}", hPay.Update).Methods("PUT")
	api.HandleFunc("/payment-methods/{id:[0-9]+}", hPay.Delete).Methods("DELETE")

	// ── Job Titles ─────────────────────────────────────────────────────────
	api.HandleFunc("/job-titles", hJob.List).Methods("GET")
	api.HandleFunc("/job-titles", hJob.Create).Methods("POST")
	api.HandleFunc("/job-titles/{id:[0-9]+}", hJob.Update).Methods("PUT")
	api.HandleFunc("/job-titles/{id:[0-9]+}", hJob.Delete).Methods("DELETE")

	// ── Integrations (VisionWeb, Zeiss) ─────────────────────────────────
	hInt := intHandler.NewHandler(intSvc.New(db))
	api.HandleFunc("/integration", hInt.ListIntegrations).Methods("GET")
	api.HandleFunc("/integration/{code}", hInt.GetIntegration).Methods("GET")
	api.HandleFunc("/integration/{code}", hInt.SetIntegration).Methods("POST")

	// ── Set Stores List (same as /home/set-stores-list) ─────────────────
	hHome := homeHandler.New(db)
	storeMW := middleware.StorePermission(db, 12, 81)
	api.Handle("/set-stores-list", storeMW(http.HandlerFunc(hHome.GetStoresList))).Methods("GET")
}
