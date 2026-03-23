package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	pbframe "sighthub-backend/internal/handlers/price_book_handler/frame"
	pblens "sighthub-backend/internal/handlers/price_book_handler/lens"
	pbcl "sighthub-backend/internal/handlers/price_book_handler/contact_lens"
	pbother "sighthub-backend/internal/handlers/price_book_handler/other"
	vwhandler "sighthub-backend/internal/handlers/visionweb_handler"
	"sighthub-backend/internal/middleware"
	"sighthub-backend/internal/services/price_book_service"
	"sighthub-backend/internal/services/visionweb_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterPriceBookRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := price_book_service.New(db)

	hFrame := pbframe.New(svc, db)
	hLens := pblens.New(svc)
	hCL := pbcl.New(svc)
	hOther := pbother.New(svc)
	hVW := vwhandler.NewHandler(visionweb_service.New(db))

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	// permission 26 = price-book read, 28 = create, 29 = update, 30 = delete
	readMW := middleware.ActivePermission(db, 26)
	writeMW := middleware.ActivePermission(db, 28)
	updateMW := middleware.ActivePermission(db, 29)
	deleteMW := middleware.ActivePermission(db, 30)

	api := r.PathPrefix("/api/price_book").Subrouter()
	api.Use(jwtMW)
	api.Use(readMW)

	// ─── Frames ───────────────────────────────────────────────────────────────
	api.HandleFunc("/vendor_brand", hFrame.GetVendorBrandCombinations).Methods("GET")
	api.HandleFunc("/products", hFrame.GetProducts).Methods("GET")
	api.HandleFunc("/models", hFrame.GetModels).Methods("GET")

	frameWrite := api.NewRoute().Subrouter()
	frameWrite.Use(writeMW)
	frameWrite.HandleFunc("/custom_glasses/{inventory_id}", hFrame.CreateCustomGlasses).Methods("POST")
	frameWrite.HandleFunc("/revert_custom_glasses/{inventory_id}", hFrame.RevertCustomGlasses).Methods("POST")

	// ─── VisionWeb Import (requires write permission) ─────────────────────
	vwWrite := api.NewRoute().Subrouter()
	vwWrite.Use(writeMW)
	vwWrite.HandleFunc("/import-vw", hVW.ImportFromVisionWeb).Methods("POST")

	// ─── Lenses ───────────────────────────────────────────────────────────────
	api.HandleFunc("/lens/brand_vendor", hLens.GetLensBrandsVendors).Methods("GET")
	api.HandleFunc("/lens/type", hLens.GetLensTypes).Methods("GET")
	api.HandleFunc("/lens/materials", hLens.GetLensMaterials).Methods("GET")
	api.HandleFunc("/lens/special", hLens.GetLensSpecialFeatures).Methods("GET")
	api.HandleFunc("/lens/series", hLens.GetLensSeries).Methods("GET")
	api.HandleFunc("/lens/v_codes", hLens.GetVCodes).Methods("GET")
	api.HandleFunc("/lens/list", hLens.GetLensList).Methods("GET")
	api.HandleFunc("/lens/{lens_id}", hLens.GetLens).Methods("GET")

	lensWrite := api.NewRoute().Subrouter()
	lensWrite.Use(writeMW)
	lensWrite.HandleFunc("/lens/materials", hLens.AddLensMaterial).Methods("POST")
	lensWrite.HandleFunc("/lens/special", hLens.AddLensSpecialFeature).Methods("POST")
	lensWrite.HandleFunc("/lens/series", hLens.AddLensSeries).Methods("POST")
	lensWrite.HandleFunc("/lens/v_codes", hLens.AddVCode).Methods("POST")
	lensWrite.HandleFunc("/lens", hLens.AddLens).Methods("POST")

	lensUpdate := api.NewRoute().Subrouter()
	lensUpdate.Use(updateMW)
	lensUpdate.HandleFunc("/lens/{lens_id}", hLens.UpdateLens).Methods("PUT")

	lensDelete := api.NewRoute().Subrouter()
	lensDelete.Use(deleteMW)
	lensDelete.HandleFunc("/lens/{lens_id}", hLens.DeleteLens).Methods("DELETE")

	// ─── Contact Lenses ───────────────────────────────────────────────────────
	api.HandleFunc("/contact_lens/vendors", hCL.GetVendors).Methods("GET")
	api.HandleFunc("/contact_lens/brands", hCL.GetBrands).Methods("GET")
	api.HandleFunc("/contact_lens/list", hCL.GetList).Methods("GET")
	api.HandleFunc("/contact_lens/{lens_id}", hCL.Get).Methods("GET")

	clWrite := api.NewRoute().Subrouter()
	clWrite.Use(writeMW)
	clWrite.HandleFunc("/contact_lens", hCL.Add).Methods("POST")

	clUpdate := api.NewRoute().Subrouter()
	clUpdate.Use(updateMW)
	clUpdate.HandleFunc("/contact_lens/{lens_id}", hCL.Update).Methods("PUT")

	clDelete := api.NewRoute().Subrouter()
	clDelete.Use(deleteMW)
	clDelete.HandleFunc("/contact_lens/{lens_id}", hCL.Delete).Methods("DELETE")

	// ─── Treatments ───────────────────────────────────────────────────────────
	api.HandleFunc("/treatments/vendors", hOther.GetTreatmentVendors).Methods("GET")
	api.HandleFunc("/treatments", hOther.GetTreatments).Methods("GET")
	api.HandleFunc("/treatments/{id}", hOther.GetTreatment).Methods("GET")

	// Treatments: no manual POST — imported from VisionWeb only
	// Update and delete still available for price/cost adjustments
	trUpdate := api.NewRoute().Subrouter()
	trUpdate.Use(updateMW)
	trUpdate.HandleFunc("/treatments/{id}", hOther.UpdateTreatment).Methods("PUT")

	trDelete := api.NewRoute().Subrouter()
	trDelete.Use(deleteMW)
	trDelete.HandleFunc("/treatments/{id}", hOther.DeleteTreatment).Methods("DELETE")

	// ─── Professional Services ────────────────────────────────────────────────
	api.HandleFunc("/professional_service_types", hOther.GetProfServiceTypes).Methods("GET")
	api.HandleFunc("/professional_service_scopes", hOther.GetProfServiceScopes).Methods("GET")
	api.HandleFunc("/professional_services", hOther.GetProfServices).Methods("GET")
	api.HandleFunc("/professional_service/{id}", hOther.GetProfService).Methods("GET")

	psWrite := api.NewRoute().Subrouter()
	psWrite.Use(writeMW)
	psWrite.HandleFunc("/professional_service", hOther.AddProfService).Methods("POST")

	psUpdate := api.NewRoute().Subrouter()
	psUpdate.Use(updateMW)
	psUpdate.HandleFunc("/professional_service/{id}", hOther.UpdateProfService).Methods("PUT")

	psDelete := api.NewRoute().Subrouter()
	psDelete.Use(deleteMW)
	psDelete.HandleFunc("/professional_service/{id}", hOther.DeleteProfService).Methods("DELETE")

	// ─── Additional Services ──────────────────────────────────────────────────
	api.HandleFunc("/add_types", hOther.GetAddTypes).Methods("GET")
	api.HandleFunc("/additional_services", hOther.GetAdditionalServices).Methods("GET")
	api.HandleFunc("/additional_service/{id}", hOther.GetAdditionalService).Methods("GET")

	asWrite := api.NewRoute().Subrouter()
	asWrite.Use(writeMW)
	asWrite.HandleFunc("/additional_service", hOther.AddAdditionalService).Methods("POST")

	asUpdate := api.NewRoute().Subrouter()
	asUpdate.Use(updateMW)
	asUpdate.HandleFunc("/additional_service/{id}", hOther.UpdateAdditionalService).Methods("PUT")

	asDelete := api.NewRoute().Subrouter()
	asDelete.Use(deleteMW)
	asDelete.HandleFunc("/additional_service/{id}", hOther.DeleteAdditionalService).Methods("DELETE")

	// ─── Misc Items ───────────────────────────────────────────────────────────
	api.HandleFunc("/misc_items", hOther.GetMiscItems).Methods("GET")
	api.HandleFunc("/misc_items/{id}", hOther.GetMiscItem).Methods("GET")

	miscWrite := api.NewRoute().Subrouter()
	miscWrite.Use(writeMW)
	miscWrite.HandleFunc("/misc_items", hOther.AddMiscItem).Methods("POST")

	miscUpdate := api.NewRoute().Subrouter()
	miscUpdate.Use(updateMW)
	miscUpdate.HandleFunc("/misc_items/{id}", hOther.UpdateMiscItem).Methods("PUT")

	miscDelete := api.NewRoute().Subrouter()
	miscDelete.Use(deleteMW)
	miscDelete.HandleFunc("/misc_items/{id}", hOther.DeleteMiscItem).Methods("DELETE")
}
