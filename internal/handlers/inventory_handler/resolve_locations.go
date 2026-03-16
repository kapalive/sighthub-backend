package inventory_handler

import (
	"net/http"

	"gorm.io/gorm"

	pkgAuth "sighthub-backend/pkg/auth"
)

// resolveLocationIDs returns the effective location IDs for the current user.
//
// Resolution order:
//  1. If the caller already supplied explicit IDs, use them as-is.
//  2. Permission-based: single SQL matching Python's
//     @store_permission_required(block_id=12, permission_id=81).
//  3. Fallback: employee → location → store → all active store locations.
func resolveLocationIDs(db *gorm.DB, r *http.Request, explicit []int64) ([]int64, error) {
	if len(explicit) > 0 {
		return explicit, nil
	}

	username := pkgAuth.UsernameFromContext(r.Context())

	// ── 1. Permission-based — one query, matches Python exactly ──────────
	var locIDs []int64
	db.Raw(`
		SELECT DISTINCT l.id_location
		FROM employee_login el
		JOIN employee_permission ep ON ep.employee_login_id = el.id_employee_login
		JOIN permissions_combination pc ON pc.id_permissions_combination = ep.permissions_combination_id
		LEFT JOIN permissions_sub_block_store pbs ON pbs.id_permissions_sub_block = pc.permissions_sub_block_store_id
		LEFT JOIN permissions_sub_block_warehouse pbw ON pbw.id_permissions_sub_block = pc.permissions_sub_block_warehouse_id
		JOIN location l ON (
			(pbs.store_id IS NOT NULL AND l.store_id = pbs.store_id)
			OR
			(pbw.warehouse_id IS NOT NULL AND l.warehouse_id = pbw.warehouse_id)
		)
		WHERE el.employee_login = ?
		  AND el.active = true
		  AND ep.is_active = true
		  AND pc.permissions_id = 81
		  AND l.store_active = true
		ORDER BY l.id_location
	`, username).Scan(&locIDs)

	if len(locIDs) > 0 {
		return locIDs, nil
	}

	// ── 2. Fallback: employee → store → all store locations ──────────────
	db.Raw(`
		SELECT l2.id_location
		FROM employee_login el
		JOIN employee e ON e.employee_login_id = el.id_employee_login
		JOIN location l ON l.id_location = e.location_id
		JOIN location l2 ON l2.store_id = l.store_id AND l2.store_active = true
		WHERE el.employee_login = ?
		ORDER BY l2.id_location
	`, username).Scan(&locIDs)

	return locIDs, nil
}
