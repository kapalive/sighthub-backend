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
	"sighthub-backend/internal/middleware"
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
	api.HandleFunc("/smtp", hSMTP.List).Methods("GET")
	api.HandleFunc("/smtp/{id:[0-9]+}", hSMTP.Get).Methods("GET")
	api.HandleFunc("/smtp", hSMTP.Create).Methods("POST")
	api.HandleFunc("/smtp/{id:[0-9]+}", hSMTP.Update).Methods("PUT")
	api.HandleFunc("/smtp/{id:[0-9]+}/test", hSMTP.Test).Methods("POST")

	// ── Appointment Reasons ────────────────────────────────────────────────
	api.HandleFunc("/appointment_reasons", hAppt.ListReasons).Methods("GET")
	api.HandleFunc("/appointment_reasons", hAppt.CreateReason).Methods("POST")
	api.HandleFunc("/appointment_reasons/{id:[0-9]+}", hAppt.UpdateReason).Methods("PUT")
	api.HandleFunc("/appointment_reasons/{id:[0-9]+}", hAppt.DeleteReason).Methods("DELETE")

	// ── Locations (showcase) ───────────────────────────────────────────────
	api.HandleFunc("/locations", hAppt.GetShowcaseLocations).Methods("GET")

	// ── Location Appointment Settings ──────────────────────────────────────
	api.HandleFunc("/request_appointment", hAppt.SetRequestAppointment).Methods("POST")
	api.HandleFunc("/request_appointment", hAppt.GetRequestAppointmentSettings).Methods("GET")
	api.HandleFunc("/intake_form", hAppt.SetIntakeForm).Methods("POST")
	api.HandleFunc("/intake_form", hAppt.GetIntakeFormSettings).Methods("GET")
	api.HandleFunc("/appointment_duration", hAppt.SetAppointmentDuration).Methods("POST")
	api.HandleFunc("/appointment_duration", hAppt.GetAppointmentDuration).Methods("GET")

	// ── Lens Types ─────────────────────────────────────────────────────────
	api.HandleFunc("/lens_types", hLens.ListTypes).Methods("GET")
	api.HandleFunc("/lens_types", hLens.CreateType).Methods("POST")
	api.HandleFunc("/lens_types/{id:[0-9]+}", hLens.UpdateType).Methods("PUT")
	api.HandleFunc("/lens_types/{id:[0-9]+}", hLens.DeleteType).Methods("DELETE")

	// ── Lens Materials ─────────────────────────────────────────────────────
	api.HandleFunc("/lens_materials", hLens.ListMaterials).Methods("GET")
	api.HandleFunc("/lens_materials", hLens.CreateMaterial).Methods("POST")
	api.HandleFunc("/lens_materials/{id:[0-9]+}", hLens.UpdateMaterial).Methods("PUT")
	api.HandleFunc("/lens_materials/{id:[0-9]+}", hLens.DeleteMaterial).Methods("DELETE")

	// ── Lens Special Features ──────────────────────────────────────────────
	api.HandleFunc("/lens_special_features", hLens.ListSpecialFeatures).Methods("GET")
	api.HandleFunc("/lens_special_features", hLens.CreateSpecialFeature).Methods("POST")
	api.HandleFunc("/lens_special_features/{id:[0-9]+}", hLens.UpdateSpecialFeature).Methods("PUT")
	api.HandleFunc("/lens_special_features/{id:[0-9]+}", hLens.DeleteSpecialFeature).Methods("DELETE")

	// ── Lens Series ────────────────────────────────────────────────────────
	api.HandleFunc("/lens_series", hLens.ListSeries).Methods("GET")
	api.HandleFunc("/lens_series", hLens.CreateSeries).Methods("POST")
	api.HandleFunc("/lens_series/{id:[0-9]+}", hLens.UpdateSeries).Methods("PUT")
	api.HandleFunc("/lens_series/{id:[0-9]+}", hLens.DeleteSeries).Methods("DELETE")

	// ── VCodes ─────────────────────────────────────────────────────────────
	api.HandleFunc("/v_codes", hLens.ListVCodes).Methods("GET")
	api.HandleFunc("/v_codes", hLens.CreateVCode).Methods("POST")
	api.HandleFunc("/v_codes/{id:[0-9]+}", hLens.UpdateVCode).Methods("PUT")
	api.HandleFunc("/v_codes/{id:[0-9]+}", hLens.DeleteVCode).Methods("DELETE")

	// ── Lens Styles ────────────────────────────────────────────────────────
	api.HandleFunc("/lens_styles", hLens.ListStyles).Methods("GET")
	api.HandleFunc("/lens_styles", hLens.CreateStyle).Methods("POST")
	api.HandleFunc("/lens_styles/{id:[0-9]+}", hLens.UpdateStyle).Methods("PUT")
	api.HandleFunc("/lens_styles/{id:[0-9]+}", hLens.DeleteStyle).Methods("DELETE")

	// ── Tint Colors ────────────────────────────────────────────────────────
	api.HandleFunc("/lens_tint_colors", hLens.ListTintColors).Methods("GET")
	api.HandleFunc("/lens_tint_colors", hLens.CreateTintColor).Methods("POST")
	api.HandleFunc("/lens_tint_colors/{id:[0-9]+}", hLens.UpdateTintColor).Methods("PUT")
	api.HandleFunc("/lens_tint_colors/{id:[0-9]+}", hLens.DeleteTintColor).Methods("DELETE")

	// ── Sample Colors ──────────────────────────────────────────────────────
	api.HandleFunc("/lens_sample_colors", hLens.ListSampleColors).Methods("GET")
	api.HandleFunc("/lens_sample_colors", hLens.CreateSampleColor).Methods("POST")
	api.HandleFunc("/lens_sample_colors/{id:[0-9]+}", hLens.UpdateSampleColor).Methods("PUT")
	api.HandleFunc("/lens_sample_colors/{id:[0-9]+}", hLens.DeleteSampleColor).Methods("DELETE")

	// ── Safety Thickness ───────────────────────────────────────────────────
	api.HandleFunc("/lens_safety_thickness", hLens.ListSafetyThickness).Methods("GET")
	api.HandleFunc("/lens_safety_thickness", hLens.CreateSafetyThickness).Methods("POST")
	api.HandleFunc("/lens_safety_thickness/{id:[0-9]+}", hLens.UpdateSafetyThickness).Methods("PUT")
	api.HandleFunc("/lens_safety_thickness/{id:[0-9]+}", hLens.DeleteSafetyThickness).Methods("DELETE")

	// ── Bevels ─────────────────────────────────────────────────────────────
	api.HandleFunc("/lens_bevels", hLens.ListBevels).Methods("GET")
	api.HandleFunc("/lens_bevels", hLens.CreateBevel).Methods("POST")
	api.HandleFunc("/lens_bevels/{id:[0-9]+}", hLens.UpdateBevel).Methods("PUT")
	api.HandleFunc("/lens_bevels/{id:[0-9]+}", hLens.DeleteBevel).Methods("DELETE")

	// ── Lens Statuses ──────────────────────────────────────────────────────
	api.HandleFunc("/lens_statuses", hLens.ListLensStatuses).Methods("GET")
	api.HandleFunc("/lens_statuses", hLens.CreateLensStatus).Methods("POST")
	api.HandleFunc("/lens_statuses/{id:[0-9]+}", hLens.UpdateLensStatus).Methods("PUT")
	api.HandleFunc("/lens_statuses/{id:[0-9]+}", hLens.DeleteLensStatus).Methods("DELETE")

	// ── Frame Shapes ───────────────────────────────────────────────────────
	api.HandleFunc("/frame_shapes", hFrame.ListShapes).Methods("GET")
	api.HandleFunc("/frame_shapes", hFrame.CreateShape).Methods("POST")
	api.HandleFunc("/frame_shapes/{id:[0-9]+}", hFrame.UpdateShape).Methods("PUT")
	api.HandleFunc("/frame_shapes/{id:[0-9]+}", hFrame.DeleteShape).Methods("DELETE")

	// ── Frame Type Materials ───────────────────────────────────────────────
	api.HandleFunc("/frame_type_materials", hFrame.ListTypeMaterials).Methods("GET")
	api.HandleFunc("/frame_type_materials", hFrame.CreateTypeMaterial).Methods("POST")
	api.HandleFunc("/frame_type_materials/{id:[0-9]+}", hFrame.UpdateTypeMaterial).Methods("PUT")
	api.HandleFunc("/frame_type_materials/{id:[0-9]+}", hFrame.DeleteTypeMaterial).Methods("DELETE")

	// ── Professional Service Scopes ────────────────────────────────────────
	api.HandleFunc("/professional_service_scopes", hProf.ListScopes).Methods("GET")
	api.HandleFunc("/professional_service_scopes", hProf.CreateScope).Methods("POST")
	api.HandleFunc("/professional_service_scopes/{id:[0-9]+}", hProf.UpdateScope).Methods("PUT")
	api.HandleFunc("/professional_service_scopes/{id:[0-9]+}", hProf.DeleteScope).Methods("DELETE")

	// ── Additional Service Types ───────────────────────────────────────────
	api.HandleFunc("/additional_service_types", hProf.ListAddTypes).Methods("GET")
	api.HandleFunc("/additional_service_types", hProf.CreateAddType).Methods("POST")
	api.HandleFunc("/additional_service_types/{id:[0-9]+}", hProf.UpdateAddType).Methods("PUT")
	api.HandleFunc("/additional_service_types/{id:[0-9]+}", hProf.DeleteAddType).Methods("DELETE")

	// ── Insurance Companies ────────────────────────────────────────────────
	api.HandleFunc("/insurance_companies", hIns.ListCompanies).Methods("GET")
	api.HandleFunc("/insurance_companies", hIns.CreateCompany).Methods("POST")
	api.HandleFunc("/insurance_companies/{id:[0-9]+}", hIns.UpdateCompany).Methods("PUT")
	api.HandleFunc("/insurance_companies/{id:[0-9]+}", hIns.DeleteCompany).Methods("DELETE")

	// ── Insurance Coverage Types ───────────────────────────────────────────
	api.HandleFunc("/insurance_coverage_types", hIns.ListCoverageTypes).Methods("GET")

	// ── Insurance Types ────────────────────────────────────────────────────
	api.HandleFunc("/insurance_types", hIns.ListInsuranceTypes).Methods("GET")

	// ── Insurance Payment Types ────────────────────────────────────────────
	api.HandleFunc("/insurance_payment_types", hIns.ListPaymentTypes).Methods("GET")
	api.HandleFunc("/insurance_payment_types", hIns.CreatePaymentType).Methods("POST")
	api.HandleFunc("/insurance_payment_types/{id:[0-9]+}", hIns.UpdatePaymentType).Methods("PUT")
	api.HandleFunc("/insurance_payment_types/{id:[0-9]+}", hIns.DeletePaymentType).Methods("DELETE")

	// ── Vendor Brands ──────────────────────────────────────────────────────
	api.HandleFunc("/vendor_brands", hVendor.ListBrands).Methods("GET")

	// ── Ticket Statuses ────────────────────────────────────────────────────
	api.HandleFunc("/ticket_statuses", hTicket.List).Methods("GET")
	api.HandleFunc("/ticket_statuses", hTicket.Create).Methods("POST")
	api.HandleFunc("/ticket_statuses/{id:[0-9]+}", hTicket.Update).Methods("PUT")
	api.HandleFunc("/ticket_statuses/{id:[0-9]+}", hTicket.Delete).Methods("DELETE")

	// ── Notify Settings ────────────────────────────────────────────────────
	api.HandleFunc("/notify/{action}", hNotify.Get).Methods("GET")
	api.HandleFunc("/notify/{action}", hNotify.Upsert).Methods("PUT")

	// ── Payment Methods ────────────────────────────────────────────────────
	api.HandleFunc("/payment_methods", hPay.List).Methods("GET")
	api.HandleFunc("/payment_methods", hPay.Create).Methods("POST")
	api.HandleFunc("/payment_methods/{id:[0-9]+}", hPay.Update).Methods("PUT")
	api.HandleFunc("/payment_methods/{id:[0-9]+}", hPay.Delete).Methods("DELETE")

	// ── Job Titles ─────────────────────────────────────────────────────────
	api.HandleFunc("/job_titles", hJob.List).Methods("GET")
	api.HandleFunc("/job_titles", hJob.Create).Methods("POST")
	api.HandleFunc("/job_titles/{id:[0-9]+}", hJob.Update).Methods("PUT")
	api.HandleFunc("/job_titles/{id:[0-9]+}", hJob.Delete).Methods("DELETE")
}
