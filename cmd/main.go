package main

import (
	"fmt"
	"log"
	"net/http"
	"sighthub-backend/config"
	"sighthub-backend/internal/blacklist"

	"sighthub-backend/internal/routes"
	exameye "sighthub-backend/internal/routes/exam_eye"
	"os"
	"sighthub-backend/pkg/cache"
	"sighthub-backend/pkg/scheduler"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Recaptcha-Token")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// 📦 Конфигурация и база
	log.Println("Loading config...")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Expose config values as env vars for packages that read os.Getenv
	if cfg.TwilioAccountSID != "" {
		os.Setenv("TWILIO_ACCOUNT_SID", cfg.TwilioAccountSID)
		os.Setenv("TWILIO_AUTH_TOKEN", cfg.TwilioAuthToken)
		os.Setenv("TWILIO_PHONE_NUMBER", cfg.TwilioPhoneNumber)
	}

	log.Println("Connecting to database...")
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUsername, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // Пишем логи в консоль
		logger.Config{
			SlowThreshold:             200 * time.Millisecond, // Порог для медленных запросов
			LogLevel:                  logger.Error,           // <<<--- Уровень Info покажет ВСЕ SQL запросы GORM
			IgnoreRecordNotFoundError: true,                   // Не показывать ошибку "запись не найдена" в логах
			ParameterizedQueries:      false,                  // Поставь true, если хочешь видеть ? вместо реальных значений в SQL
			Colorful:                  true,                   // Цветной вывод
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger, // <-- ВОТ ТАК ПЕРЕДАЕМ ЛОГГЕР В GORM
		// Здесь могут быть другие твои настройки GORM, если они были
	})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	fmt.Println("Database connected successfully!")

	// Redis
	cache.InitRedisClient(cfg.RedisAddr)
	cache.StartDailyCachePurge("cache:")

	// 🧹 Blacklist очистка
	blacklist.StartCleaner(1 * time.Minute)

	// ⏱ Scheduler (replaces Celery)
	sched := scheduler.New()

	// 📡 Роутинг
	router := mux.NewRouter()

	routes.RegisterAuthRoutes(db, cache.RDB, cfg, router)
	routes.RegisterHomeRoutes(db, cache.RDB, cfg, router)
	routes.RegisterFrameLibraryRoutes(db, cache.RDB, cfg, router)
	routes.RegisterPriceBookRoutes(db, cache.RDB, cfg, router)
	routes.RegisterTasksRoutes(db, cache.RDB, cfg, router)
	routes.RegisterTimecardRoutes(db, cache.RDB, cfg, router)
	routes.RegisterInvoiceRoutes(db, cache.RDB, cfg, router)
	routes.RegisterOrderedInventoryRoutes(db, cache.RDB, cfg, router)
	routes.RegisterPosTerminalRoutes(db, cache.RDB, cfg, router)
	routes.RegisterReportAccountingRoutes(db, cache.RDB, cfg, router)
	routes.RegisterReportDailyRoutes(db, cache.RDB, cfg, router)
	routes.RegisterLicenseRoutes(db, router)
	routes.RegisterQuestionnaireRoutes(db, cache.RDB, cfg, router)
	routes.RegisterAppointmentBookRoutes(db, cache.RDB, cfg, router)
	routes.RegisterRequestAppointmentRoutes(db, router)
	routes.RegisterEmailTemplateRoutes(db, cache.RDB, cfg, router)
	routes.RegisterHelpdeskRoutes(db, cache.RDB, cfg, router)
	routes.RegisterEmployeeRoutes(db, cache.RDB, cfg, router)
	routes.RegisterStoreRoutes(db, cache.RDB, cfg, router)
	routes.RegisterPatientRoutes(db, cache.RDB, cfg, router)
	routes.RegisterProfileRoutes(db, cache.RDB, cfg, router)
	routes.RegisterARReportRoutes(db, cache.RDB, cfg, router)
	routes.RegisterDoctorDeskRoutes(db, cache.RDB, cfg, router)
	exameye.RegisterExamEyeRoutes(db, cache.RDB, cfg, router)
	exameye.RegisterHistoryRoutes(db, cache.RDB, cfg, router)
	exameye.RegisterCcHpiRoutes(db, cache.RDB, cfg, router)
	exameye.RegisterPreliminaryRoutes(db, cache.RDB, cfg, router)
	exameye.RegisterRefractionRoutes(db, cache.RDB, cfg, router)
	exameye.RegisterClFittingRoutes(db, cache.RDB, cfg, router)
	exameye.RegisterExternalSleRoutes(db, cache.RDB, cfg, router)
	exameye.RegisterPosteriorRoutes(db, cache.RDB, cfg, router)
	exameye.RegisterSpecialRoutes(db, cache.RDB, cfg, router)
	exameye.RegisterAssessmentRoutes(db, cache.RDB, cfg, router)
	exameye.RegisterReferralRoutes(db, cache.RDB, cfg, router)
	exameye.RegisterSuperRoutes(db, cache.RDB, cfg, router)
	exameye.RegisterSLRPRoutes(db, cache.RDB, cfg, router)
	routes.RegisterBillRoutes(db, cache.RDB, cfg, router)
	routes.RegisterTicketRoutes(db, cache.RDB, cfg, sched, router)
	routes.RegisterAccountingRoutes(db, cache.RDB, cfg, router)
	routes.RegisterClaimRoutes(db, cache.RDB, cfg, router)
	routes.RegisterDailyCloseRoutes(db, cache.RDB, cfg, router)
	routes.RegisterDashboardRoutes(db, cache.RDB, cfg, router)
	routes.RegisterInventoryRoutes(db, cache.RDB, cfg, router)
	routes.RegisterReportInventoryRoutes(db, cache.RDB, cfg, router)
	routes.RegisterReportLibraryRoutes(db, cache.RDB, cfg, router)
	routes.RegisterReportSalesRoutes(db, cache.RDB, cfg, router)
	routes.RegisterSaleRoutes(db, cache.RDB, cfg, router)
	routes.RegisterOrdersRoutes(db, cache.RDB, cfg, router)
	routes.RegisterSMSTemplateRoutes(db, cache.RDB, cfg, router)
	routes.RegisterSettingsRoutes(db, cache.RDB, cfg, router)
	routes.RegisterVendorRoutes(db, cache.RDB, cfg, router)

	// Integrations (domain routes)
	// Zeiss auth is now in price_book and ticket routes

	// favicon
	router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	})

	addr := ":" + cfg.Port
	log.Println("Server starting on", addr)
	if err := http.ListenAndServe(addr, corsMiddleware(router)); err != nil {
		log.Fatal(err)
	}

}
