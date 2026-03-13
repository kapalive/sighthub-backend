package routes

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sighthub-backend/config"
	commHandler "sighthub-backend/internal/handlers/commission_handler"
	empHandler "sighthub-backend/internal/handlers/employee_handler"
	permHandler "sighthub-backend/internal/handlers/permission_handler"
	"sighthub-backend/internal/middleware"
	commSvc "sighthub-backend/internal/services/commission_service"
	empSvc "sighthub-backend/internal/services/employee_service"
	permSvc "sighthub-backend/internal/services/permission_service"
	pkgAuth "sighthub-backend/pkg/auth"
)

func RegisterEmployeeRoutes(db *gorm.DB, rdb *redis.Client, cfg *config.Config, r *mux.Router) {
	api := r.PathPrefix("/api/employee").Subrouter()

	jwtMW := pkgAuth.JWTMiddleware(cfg.JWTSecretKey, rdb)
	permMW := middleware.ActivePermission(db, 48)

	api.Use(jwtMW, permMW)

	emp := empHandler.New(empSvc.New(db))
	perm := permHandler.New(permSvc.New(db))
	comm := commHandler.New(commSvc.New(db))

	// Employee general
	api.HandleFunc("/login/availability", emp.CheckLoginAvailability).Methods("GET")
	api.HandleFunc("/location", emp.GetLocations).Methods("GET")
	api.HandleFunc("/warehouses/{location_id:[0-9]+}", emp.GetWarehousesByLocation).Methods("GET")
	api.HandleFunc("/roles", emp.GetRoles).Methods("GET")
	api.HandleFunc("/", emp.ListEmployees).Methods("GET")
	api.HandleFunc("/employee/general", emp.GetEmployeeGeneral).Methods("GET")
	api.HandleFunc("/add", emp.AddEmployee).Methods("POST")
	api.HandleFunc("/{employee_id:[0-9]+}", emp.GetEmployee).Methods("GET")
	api.HandleFunc("/{employee_id:[0-9]+}", emp.UpdateEmployee).Methods("PUT")
	api.HandleFunc("/deactivate/{login_id:[0-9]+}", emp.DeactivateTimecard).Methods("POST")

	// Timecard
	api.HandleFunc("/time_card", emp.ListTimecards).Methods("GET")
	api.HandleFunc("/time_card", emp.CreateTimecard).Methods("POST")
	api.HandleFunc("/time_card/{timecard_login_id:[0-9]+}", emp.GetTimecardHistory).Methods("GET")
	api.HandleFunc("/{employee_id:[0-9]+}/time_card", emp.GetEmployeeTimecardHistory).Methods("GET")
	api.HandleFunc("/{employee_id:[0-9]+}/time_card", emp.CreateTimecard).Methods("POST")
	api.HandleFunc("/{employee_id:[0-9]+}/time_card", emp.UpdateTimecard).Methods("PUT")

	// Schedule
	api.HandleFunc("/{employee_id:[0-9]+}/schedule", emp.GetSchedule).Methods("GET")
	api.HandleFunc("/{employee_id:[0-9]+}/schedule", emp.CreateSchedule).Methods("POST")
	api.HandleFunc("/{employee_id:[0-9]+}/schedule", emp.UpdateSchedule).Methods("PUT")

	// Off days
	api.HandleFunc("/{employee_id:[0-9]+}/off_day", emp.AddOffDay).Methods("POST")
	api.HandleFunc("/{employee_id:[0-9]+}/off_day", emp.ListOffDays).Methods("GET")
	api.HandleFunc("/{employee_id:[0-9]+}/off_day/{date_str}", emp.RemoveOffDay).Methods("DELETE")

	// Commissions
	api.HandleFunc("/{employee_id:[0-9]+}/commissions", comm.GetCommissions).Methods("GET")
	api.HandleFunc("/{employee_id:[0-9]+}/commission/current", comm.GetCurrentCommission).Methods("GET")
	api.HandleFunc("/{employee_id:[0-9]+}/commissions/history", comm.GetCommissionHistory).Methods("GET")
	api.HandleFunc("/{employee_id:[0-9]+}/commissions/{commission_id:[0-9]+}", comm.UpdateCommission).Methods("PUT")
	api.HandleFunc("/{employee_id:[0-9]+}/commissions", comm.CreateCommission).Methods("POST")

	// Job titles
	api.HandleFunc("/job-titles", emp.GetJobTitles).Methods("GET")

	// Permissions — block-based (15 blocks)
	permBlocks := map[string]int{
		"docktor_desk":       1,
		"appointment_manage": 2,
		"inventory":          3,
		"pricebook":          4,
		"claim_billing":      5,
		"accountant":         6,
		"ar_report":          7,
		"vendors":            8,
		"sale_report":        9,
		"employees":          10,
		"stores":             11,
		"time_card":          12,
		"patient":            13,
		"invoice":            14,
		"setting":            15,
	}
	for name, blockID := range permBlocks {
		bid := blockID // capture for closure
		// Use a sub-route with block_id embedded via the handler
		permRoute := "/permissions/" + name + "/{employee_id:[0-9]+}"
		api.HandleFunc(permRoute, makeBlockPermissionsHandler(perm, bid)).Methods("GET")
		api.HandleFunc(permRoute, makeBlockPermissionsHandler(perm, bid)).Methods("POST")
	}

	// Permissions — warehouse access
	api.HandleFunc("/permissions/warehouses/{employee_id:[0-9]+}", perm.GetWarehouseAccess).Methods("GET")
	api.HandleFunc("/permissions/warehouses/{employee_id:[0-9]+}", perm.SetWarehouseAccess).Methods("POST")
}

// makeBlockPermissionsHandler wraps block_id into mux vars so handlers can read it.
func makeBlockPermissionsHandler(h *permHandler.Handler, blockID int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Inject block_id into the mux vars map
		vars := mux.Vars(r)
		vars["block_id"] = strconv.Itoa(blockID)
		switch r.Method {
		case "GET":
			h.GetBlockPermissions(w, r)
		case "POST":
			h.SetBlockPermission(w, r)
		}
	}
}
