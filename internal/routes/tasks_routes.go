package routes

import (
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	tasksH "sighthub-backend/internal/handlers/tasks_handler"
	taskSvc "sighthub-backend/internal/services/task_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterTasksRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	s := taskSvc.New(db)
	h := tasksH.New(s, db)

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)

	api := r.PathPrefix("/api/tasks").Subrouter()
	api.Use(jwtMW)

	api.HandleFunc("", h.GetTasks).Methods("GET")
	api.HandleFunc("", h.CreateTask).Methods("POST")
	api.HandleFunc("/employees", h.GetEmployees).Methods("GET")
	api.HandleFunc("/invoice/search", h.SearchInvoice).Methods("GET")
	api.HandleFunc("/resource/{id_tasks_resource:[0-9]+}", h.DeleteTaskResource).Methods("DELETE")
	api.HandleFunc("/comments/{id_task_comment:[0-9]+}", h.UpdateComment).Methods("PUT")
	api.HandleFunc("/comments/{id_task_comment:[0-9]+}", h.DeleteComment).Methods("DELETE")
	api.HandleFunc("/{id_task:[0-9]+}", h.GetTask).Methods("GET")
	api.HandleFunc("/{id_task:[0-9]+}", h.UpdateTask).Methods("PUT")
	api.HandleFunc("/{id_task:[0-9]+}", h.DeleteTask).Methods("DELETE")
	api.HandleFunc("/{id_task:[0-9]+}/comments", h.CreateComment).Methods("POST")
}
