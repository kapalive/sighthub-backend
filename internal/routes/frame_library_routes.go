package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	"sighthub-backend/internal/handlers/frame_library_handler"
	"sighthub-backend/internal/middleware"
	"sighthub-backend/internal/services/frame_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterFrameLibraryRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	svc := frame_service.New(db)
	h := frame_library_handler.New(svc)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	// permission 28 = frame library write access
	writeMW := middleware.ActivePermission(db, 28)

	api := r.PathPrefix("/api/frame_library").Subrouter()
	api.Use(jwtMW)

	// Read-only routes (JWT only)
	api.HandleFunc("/search", h.SearchModels).Methods("GET")
	api.HandleFunc("/vendor_brand", h.GetVendorBrands).Methods("GET")
	api.HandleFunc("/vendor-brand", h.GetVendorBrandCombinations).Methods("GET")
	api.HandleFunc("/products", h.GetProducts).Methods("GET")
	api.HandleFunc("/materials_frame", h.GetMaterialsFrame).Methods("GET")
	api.HandleFunc("/frame-type-materials", h.GetFrameTypeMaterials).Methods("GET")
	api.HandleFunc("/shapes", h.GetFrameShapes).Methods("GET")
	api.HandleFunc("/lens/materials", h.GetLensMaterials).Methods("GET")
	api.HandleFunc("/models_by_product/{product_id}", h.GetModelsByProduct).Methods("GET")
	api.HandleFunc("/update_model/{id_model}", h.UpdateModel).Methods("PUT")

	// Write routes (require permission 28)
	write := api.NewRoute().Subrouter()
	write.Use(writeMW)
	write.HandleFunc("/add_model", h.AddProduct).Methods("POST")
	write.HandleFunc("/add_variant", h.AddVariant).Methods("POST")
	write.HandleFunc("/custom_glasses/{id_model}", h.CreateCustomGlasses).Methods("POST")
}
