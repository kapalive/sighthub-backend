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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
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

	// 📡 Роутинг
	router := mux.NewRouter()

	routes.RegisterAuthRoutes(db, cache.RDB, cfg, router)
	routes.RegisterHomeRoutes(db, cache.RDB, cfg, router)
	routes.RegisterFrameLibraryRoutes(db, cache.RDB, cfg, router)
	routes.RegisterPriceBookRoutes(db, cache.RDB, cfg, router)
	routes.RegisterTasksRoutes(db, cache.RDB, cfg, router)
	routes.RegisterTimecardRoutes(db, cache.RDB, cfg, router)
	routes.RegisterInvoiceRoutes(db, cache.RDB, cfg, router)
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

	addr := ":" + cfg.Port
	log.Println("Server starting on", addr)
	if err := http.ListenAndServe(addr, corsMiddleware(router)); err != nil {
		log.Fatal(err)
	}

}
