package middleware

import (
	"context"
	"net/http"

	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	"sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/location"
	"sighthub-backend/internal/models/permission"
	pkgAuth "sighthub-backend/pkg/auth"
)

// ─── context keys ─────────────────────────────────────────────────────────────

type ctxKey int

const (
	keyLogin               ctxKey = iota
	keyEmployee            ctxKey = iota
	keyLocation            ctxKey = iota
	keyAllowedNavigation   ctxKey = iota
	keyAllowedPermIDs      ctxKey = iota
	keyPermittedLocationIDs ctxKey = iota
	keyPermittedLocations  ctxKey = iota
)

// ─── Context helpers (used by handlers) ───────────────────────────────────────

func LoginFromCtx(ctx context.Context) *authModel.EmployeeLogin {
	v, _ := ctx.Value(keyLogin).(*authModel.EmployeeLogin)
	return v
}

func EmployeeFromCtx(ctx context.Context) *employees.Employee {
	v, _ := ctx.Value(keyEmployee).(*employees.Employee)
	return v
}

func LocationFromCtx(ctx context.Context) *location.Location {
	v, _ := ctx.Value(keyLocation).(*location.Location)
	return v
}

func AllowedNavigationFromCtx(ctx context.Context) map[string]struct{} {
	v, _ := ctx.Value(keyAllowedNavigation).(map[string]struct{})
	return v
}

func AllowedPermIDsFromCtx(ctx context.Context) map[int]struct{} {
	v, _ := ctx.Value(keyAllowedPermIDs).(map[int]struct{})
	return v
}

func PermittedLocationIDsFromCtx(ctx context.Context) []int {
	v, _ := ctx.Value(keyPermittedLocationIDs).([]int)
	return v
}

func PermittedLocationsFromCtx(ctx context.Context) []location.Location {
	v, _ := ctx.Value(keyPermittedLocations).([]location.Location)
	return v
}

// ─── Navigation permission map ─────────────────────────────────────────────────

var navigationPermissionMap = map[string]int{
	"/doctor-desk":       1,
	"/appointment":       6,
	"/inventory":         11,
	"/price-book":        26,
	"/claim-and-billing": 31,
	"/accounting":        32,
	"/vendors":           36,
	"/stores":            49,
	"/patient":           51,
	"/settings":          80,
	"/employees":         48,
}

var defaultNavigation = map[string]struct{}{
	"/":        {},
	"/tasks/":  {},
	"/daily":   {},
	"/reports": {},
}

// permissions_ids that are forbidden on warehouse locations
var warehouseForbiddenPermIDs = map[int]struct{}{
	1:  {}, // doctor-desk
	6:  {}, // appointment
	31: {}, // claim-and-billing
	32: {}, // accountant
	51: {}, // patient
	80: {}, // settings
	48: {}, // employees
}

// ─── Shared DB helpers ─────────────────────────────────────────────────────────

func loadLoginAndEmployee(db *gorm.DB, username string) (*authModel.EmployeeLogin, *employees.Employee, *location.Location, error) {
	var login authModel.EmployeeLogin
	if err := db.Where("employee_login = ?", username).First(&login).Error; err != nil {
		return nil, nil, nil, err
	}

	var emp employees.Employee
	if err := db.Where("employee_login_id = ?", login.IDEmployeeLogin).First(&emp).Error; err != nil {
		return &login, nil, nil, err
	}

	if emp.LocationID == nil {
		return &login, &emp, nil, nil
	}

	var loc location.Location
	if err := db.First(&loc, *emp.LocationID).Error; err != nil {
		return &login, &emp, nil, err
	}

	return &login, &emp, &loc, nil
}

func normalizePath(path string) string {
	if len(path) == 0 {
		return "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}
	if path != "/" {
		for len(path) > 1 && path[len(path)-1] == '/' {
			path = path[:len(path)-1]
		}
	}
	return path
}

// ─── NavigationPermission middleware ──────────────────────────────────────────
//
// Sets in context: Login, Employee, Location, AllowedNavigation, AllowedPermIDs.
// Equivalent to Python's @navigation_permission_required decorator.

func NavigationPermission(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username := pkgAuth.UsernameFromContext(r.Context())
			if username == "" {
				jsonErr(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			login, emp, loc, err := loadLoginAndEmployee(db, username)
			if err != nil || loc == nil {
				jsonErr(w, "Employee or location not found", http.StatusNotFound)
				return
			}

			isWarehouse := loc.WarehouseID != nil

			// Load all active employee permissions + combinations
			type row struct {
				PermissionsID              int
				SubBlockStoreID            *int
				SubBlockWarehouseID        *int
				SubBlockStoreStoreID       *int
				SubBlockWarehouseWarehouseID *int
			}

			var rows []row
			db.Table("employee_permission ep").
				Select(`pc.permissions_id,
					pc.permissions_sub_block_store_id,
					pc.permissions_sub_block_warehouse_id,
					pbs.store_id AS sub_block_store_store_id,
					pbw.warehouse_id AS sub_block_warehouse_warehouse_id`).
				Joins("JOIN permissions_combination pc ON pc.id_permissions_combination = ep.permissions_combination_id").
				Joins("LEFT JOIN permissions_sub_block_store pbs ON pbs.id_permissions_sub_block = pc.permissions_sub_block_store_id").
				Joins("LEFT JOIN permissions_sub_block_warehouse pbw ON pbw.id_permissions_sub_block = pc.permissions_sub_block_warehouse_id").
				Where("ep.employee_login_id = ? AND ep.is_active = true", login.IDEmployeeLogin).
				Scan(&rows)

			allowedPermIDs := make(map[int]struct{})

			for _, rw := range rows {
				// global permission (no store or warehouse subblock)
				if rw.SubBlockStoreID == nil && rw.SubBlockWarehouseID == nil {
					allowedPermIDs[rw.PermissionsID] = struct{}{}
					continue
				}

				if !isWarehouse {
					// store location: only store combos matching this store
					if rw.SubBlockStoreID != nil && rw.SubBlockStoreStoreID != nil &&
						*rw.SubBlockStoreStoreID == loc.StoreID {
						allowedPermIDs[rw.PermissionsID] = struct{}{}
					}
				} else {
					// warehouse location: only warehouse combos matching this warehouse
					if rw.SubBlockWarehouseID != nil && rw.SubBlockWarehouseWarehouseID != nil &&
						loc.WarehouseID != nil && *rw.SubBlockWarehouseWarehouseID == *loc.WarehouseID {
						if _, forbidden := warehouseForbiddenPermIDs[rw.PermissionsID]; !forbidden {
							allowedPermIDs[rw.PermissionsID] = struct{}{}
						}
					}
				}
			}

			// Remove warehouse-forbidden perms if on warehouse
			if isWarehouse {
				for pid := range warehouseForbiddenPermIDs {
					delete(allowedPermIDs, pid)
				}
			}

			// Build allowed navigation URLs
			allowedNav := make(map[string]struct{})
			for url := range defaultNavigation {
				allowedNav[normalizePath(url)] = struct{}{}
			}
			for url, permID := range navigationPermissionMap {
				if _, ok := allowedPermIDs[permID]; ok {
					allowedNav[normalizePath(url)] = struct{}{}
				}
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, keyLogin, login)
			ctx = context.WithValue(ctx, keyEmployee, emp)
			ctx = context.WithValue(ctx, keyLocation, loc)
			ctx = context.WithValue(ctx, keyAllowedNavigation, allowedNav)
			ctx = context.WithValue(ctx, keyAllowedPermIDs, allowedPermIDs)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ─── StorePermission middleware ────────────────────────────────────────────────
//
// Resolves permitted location IDs for a given block+permission combo.
// Sets in context: Login, Employee, PermittedLocationIDs, PermittedLocations.
// Equivalent to Python's @store_permission_required(block_id, permission_id).

func StorePermission(db *gorm.DB, blockID, permissionID int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username := pkgAuth.UsernameFromContext(r.Context())
			if username == "" {
				jsonErr(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			login, emp, _, err := loadLoginAndEmployee(db, username)
			if err != nil {
				jsonErr(w, "Employee not found", http.StatusNotFound)
				return
			}

			// Get combination IDs matching block+permission
			var comboIDs []int
			db.Model(&permission.PermissionsCombination{}).
				Where("permissions_id = ? AND permissions_block_id = ?", permissionID, blockID).
				Pluck("id_permissions_combination", &comboIDs)

			if len(comboIDs) == 0 {
				jsonErr(w, "No permission combinations found", http.StatusForbidden)
				return
			}

			// Get active combos for this user
			type subBlockRow struct {
				SubBlockStoreID     *int
				SubBlockWarehouseID *int
			}
			var sbRows []subBlockRow
			db.Table("employee_permission ep").
				Select("pc.permissions_sub_block_store_id AS sub_block_store_id, pc.permissions_sub_block_warehouse_id AS sub_block_warehouse_id").
				Joins("JOIN permissions_combination pc ON pc.id_permissions_combination = ep.permissions_combination_id").
				Where("ep.employee_login_id = ? AND ep.is_active = true AND ep.permissions_combination_id IN ?",
					login.IDEmployeeLogin, comboIDs).
				Scan(&sbRows)

			var storeSubBlockIDs, warehouseSubBlockIDs []int
			for _, sb := range sbRows {
				if sb.SubBlockStoreID != nil {
					storeSubBlockIDs = append(storeSubBlockIDs, *sb.SubBlockStoreID)
				}
				if sb.SubBlockWarehouseID != nil {
					warehouseSubBlockIDs = append(warehouseSubBlockIDs, *sb.SubBlockWarehouseID)
				}
			}

			locationIDSet := make(map[int]struct{})

			if len(storeSubBlockIDs) > 0 {
				var storeIDs []int
				db.Model(&permission.PermissionsSubBlockStore{}).
					Where("id_permissions_sub_block IN ?", storeSubBlockIDs).
					Pluck("store_id", &storeIDs)

				if len(storeIDs) > 0 {
					var locIDs []int
					db.Model(&location.Location{}).
						Where("store_id IN ? AND store_active = true", storeIDs).
						Pluck("id_location", &locIDs)
					for _, id := range locIDs {
						locationIDSet[id] = struct{}{}
					}
				}
			}

			if len(warehouseSubBlockIDs) > 0 {
				var warehouseIDs []int
				db.Model(&permission.PermissionsSubBlockWarehouse{}).
					Where("id_permissions_sub_block IN ?", warehouseSubBlockIDs).
					Pluck("warehouse_id", &warehouseIDs)

				if len(warehouseIDs) > 0 {
					var locIDs []int
					db.Model(&location.Location{}).
						Where("warehouse_id IN ? AND store_active = true", warehouseIDs).
						Pluck("id_location", &locIDs)
					for _, id := range locIDs {
						locationIDSet[id] = struct{}{}
					}
				}
			}

			permittedIDs := make([]int, 0, len(locationIDSet))
			for id := range locationIDSet {
				permittedIDs = append(permittedIDs, id)
			}
			sortInts(permittedIDs)

			var locations []location.Location
			if len(permittedIDs) > 0 {
				db.Where("id_location IN ?", permittedIDs).Find(&locations)
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, keyLogin, login)
			ctx = context.WithValue(ctx, keyEmployee, emp)
			ctx = context.WithValue(ctx, keyPermittedLocationIDs, permittedIDs)
			ctx = context.WithValue(ctx, keyPermittedLocations, locations)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ─── FPermissions middleware ───────────────────────────────────────────────────
//
// Checks that the user has at least one active permission from the given
// permissions_combination_id list.
// Equivalent to Python's @fpermissions_required(combination_ids).

func FPermissions(db *gorm.DB, combinationIDs ...int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username := pkgAuth.UsernameFromContext(r.Context())
			if username == "" {
				jsonErr(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			var login authModel.EmployeeLogin
			if err := db.Where("employee_login = ?", username).First(&login).Error; err != nil {
				jsonErr(w, "User not found", http.StatusNotFound)
				return
			}

			var count int64
			db.Model(&employees.EmployeePermission{}).
				Where("employee_login_id = ? AND permissions_combination_id IN ? AND is_active = true",
					login.IDEmployeeLogin, combinationIDs).
				Count(&count)

			if count == 0 {
				jsonErr(w, "Permission denied: insufficient rights", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ─── ActivePermission middleware ───────────────────────────────────────────────
//
// Checks that the user has an active permission for the given permissions_id
// (regardless of location). Equivalent to Python's @active_permission_required.

func ActivePermission(db *gorm.DB, permissionID int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username := pkgAuth.UsernameFromContext(r.Context())
			if username == "" {
				jsonErr(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			var login authModel.EmployeeLogin
			if err := db.Where("employee_login = ?", username).First(&login).Error; err != nil {
				jsonErr(w, "User not found", http.StatusNotFound)
				return
			}

			var count int64
			db.Table("employee_permission ep").
				Joins("JOIN permissions_combination pc ON pc.id_permissions_combination = ep.permissions_combination_id").
				Where("ep.employee_login_id = ? AND ep.is_active = true AND pc.permissions_id = ?",
					login.IDEmployeeLogin, permissionID).
				Count(&count)

			if count == 0 {
				jsonErr(w, "Permission denied", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ─── GetPermittedLocationIDs helper ───────────────────────────────────────────
//
// Standalone helper (not middleware) that returns permitted location IDs
// for a given username and permissionID. Useful inside handlers as a fallback.

func GetPermittedLocationIDs(db *gorm.DB, username string, permissionID int) []int {
	var login authModel.EmployeeLogin
	if err := db.Where("employee_login = ? AND active = true", username).First(&login).Error; err != nil {
		return nil
	}

	type subBlockRow struct {
		SubBlockStoreID     *int
		SubBlockWarehouseID *int
	}
	var rows []subBlockRow
	db.Table("employee_permission ep").
		Select("pc.permissions_sub_block_store_id AS sub_block_store_id, pc.permissions_sub_block_warehouse_id AS sub_block_warehouse_id").
		Joins("JOIN permissions_combination pc ON pc.id_permissions_combination = ep.permissions_combination_id").
		Where("ep.employee_login_id = ? AND ep.is_active = true AND pc.permissions_id = ?",
			login.IDEmployeeLogin, permissionID).
		Scan(&rows)

	locationIDSet := make(map[int]struct{})

	for _, rw := range rows {
		if rw.SubBlockStoreID != nil {
			var storeID *int
			db.Model(&permission.PermissionsSubBlockStore{}).
				Where("id_permissions_sub_block = ?", *rw.SubBlockStoreID).
				Pluck("store_id", &storeID)
			if storeID != nil {
				var locIDs []int
				db.Model(&location.Location{}).
					Where("store_id = ? AND store_active = true", *storeID).
					Pluck("id_location", &locIDs)
				for _, id := range locIDs {
					locationIDSet[id] = struct{}{}
				}
			}
		} else if rw.SubBlockWarehouseID != nil {
			var warehouseID *int
			db.Model(&permission.PermissionsSubBlockWarehouse{}).
				Where("id_permissions_sub_block = ?", *rw.SubBlockWarehouseID).
				Pluck("warehouse_id", &warehouseID)
			if warehouseID != nil {
				var locIDs []int
				db.Model(&location.Location{}).
					Where("warehouse_id = ? AND store_active = true", *warehouseID).
					Pluck("id_location", &locIDs)
				for _, id := range locIDs {
					locationIDSet[id] = struct{}{}
				}
			}
		}
	}

	result := make([]int, 0, len(locationIDSet))
	for id := range locationIDSet {
		result = append(result, id)
	}
	sortInts(result)
	return result
}

// ─── PathAllowed helper ────────────────────────────────────────────────────────

func PathAllowed(allowed map[string]struct{}, requiredPath string) bool {
	rp := normalizePath(requiredPath)
	if _, ok := allowed[rp]; ok {
		return true
	}
	for p := range allowed {
		p = normalizePath(p)
		if len(rp) > len(p) && rp[:len(p)+1] == p+"/" {
			return true
		}
		if len(p) > len(rp) && p[:len(rp)+1] == rp+"/" {
			return true
		}
	}
	return false
}

// ─── RequireLocationPermission middleware ────────────────────────────────────
//
// Checks that the user has the given permission_id at their current location.
// Equivalent to Python's @require_location_permission(permission_id).

func RequireLocationPermission(db *gorm.DB, permissionIDs ...int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username := pkgAuth.UsernameFromContext(r.Context())
			if username == "" {
				jsonErr(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			login, emp, _, err := loadLoginAndEmployee(db, username)
			if err != nil || emp == nil || emp.LocationID == nil {
				jsonErr(w, "Employee not assigned to a location", http.StatusForbidden)
				return
			}

			var loc location.Location
			if err := db.First(&loc, *emp.LocationID).Error; err != nil {
				jsonErr(w, "Location not found", http.StatusNotFound)
				return
			}

			var count int64
			q := db.Table("employee_permission ep").
				Joins("JOIN permissions_combination pc ON pc.id_permissions_combination = ep.permissions_combination_id").
				Joins("LEFT JOIN permissions_sub_block_store pbs ON pbs.id_permissions_sub_block = pc.permissions_sub_block_store_id").
				Joins("LEFT JOIN permissions_sub_block_warehouse pbw ON pbw.id_permissions_sub_block = pc.permissions_sub_block_warehouse_id").
				Where("ep.employee_login_id = ? AND ep.is_active = true AND pc.permissions_id IN ?",
					login.IDEmployeeLogin, permissionIDs).
				Where(`(
					(pc.permissions_sub_block_store_id IS NULL AND pc.permissions_sub_block_warehouse_id IS NULL)
					OR pbs.store_id = ?
					OR pbw.warehouse_id = ?
				)`, loc.StoreID, loc.WarehouseID)

			q.Count(&count)
			if count == 0 {
				jsonErr(w, "Permission denied", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ─── internal helpers ──────────────────────────────────────────────────────────

func jsonErr(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(`{"error":"` + msg + `"}`))
}

func sortInts(s []int) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
