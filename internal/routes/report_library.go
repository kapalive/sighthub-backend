package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	"sighthub-backend/internal/handlers/report_library_handler"
	"sighthub-backend/internal/services/report_library_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterReportLibraryRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	s := report_library_service.New(db)
	h := report_library_handler.New(s, db)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)

	api := r.PathPrefix("/api/report-library").Subrouter()
	api.Use(jwtMW)

	api.HandleFunc("/all_reports", h.AllReports).Methods("GET")

	// Sales
	api.HandleFunc("/sales/invoice_summary", h.InvoiceSummary).Methods("GET")
	api.HandleFunc("/sales/invoice_classification", h.InvoiceClassification).Methods("GET")
	api.HandleFunc("/sales/vendor_brand_margin_report", h.VendorBrandMarginReport).Methods("GET")
	api.HandleFunc("/sales/sales_by_location", h.SalesByLocation).Methods("GET")
	api.HandleFunc("/sales/sales_by_frame", h.SalesByFrame).Methods("GET")
	api.HandleFunc("/sales/sales_average", h.SalesAverage).Methods("GET")
	api.HandleFunc("/sales/sales_breakdown_by_product_type", h.SalesBreakdownByProductType).Methods("GET")
	api.HandleFunc("/sales/sales_by_employee", h.SalesByEmployee).Methods("GET")

	// Gift cards
	api.HandleFunc("/sales/gift_card_balance", h.GiftCardBalance).Methods("GET")
	api.HandleFunc("/sales/gift_card_details", h.GiftCardDetails).Methods("GET")
	api.HandleFunc("/sales/gift_card_activity", h.GiftCardActivity).Methods("GET")

	// First questionnaire
	api.HandleFunc("/first_questionnaire/referral", h.QuestionnaireReferral).Methods("GET")
	api.HandleFunc("/first_questionnaire/reasons", h.QuestionnaireReasons).Methods("GET")
}
