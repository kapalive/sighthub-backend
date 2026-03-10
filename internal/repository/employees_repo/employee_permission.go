package employees_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"sighthub-backend/internal/models/employees"
	"sighthub-backend/internal/models/location"
	"sighthub-backend/internal/models/permission"
)

// warehouseForbiddenPermissionIDs is the set of permissions_ids that exist in the
// store permission set but must NOT be mirrored to warehouses.
// STORE has IDs 1-35, 41-47, 62-81.
// WAREHOUSE has IDs 11-30, 32, 41-47, 62-81.
// Forbidden = store-only = 1-10, 31, 33-35.
var warehouseForbiddenPermissionIDs = func() map[int]struct{} {
	forbidden := make(map[int]struct{})
	for i := 1; i <= 10; i++ {
		forbidden[i] = struct{}{}
	}
	forbidden[31] = struct{}{}
	for i := 33; i <= 35; i++ {
		forbidden[i] = struct{}{}
	}
	return forbidden
}()

// timecardPermissionID is the permissions_id for the timecard permission.
// Granting/revoking it also activates/deactivates EmployeeTimecardLogin.
const timecardPermissionID = 50

// ─────────────────────────────────────────────
// DTO types
// ─────────────────────────────────────────────

type PermissionStatus struct {
	CombinationID  int    `json:"combination_id"`
	PermissionName string `json:"permission_name"`
	Granted        bool   `json:"granted"`
}

type WarehousePermissionStatus struct {
	WarehouseID   int    `json:"warehouse_id"`
	WarehouseName string `json:"warehouse_name"`
	Granted       bool   `json:"granted"`
}

type LocationItem struct {
	LocationID   int    `json:"location_id"`
	LocationName string `json:"location_name"`
}

type WarehouseItem struct {
	WarehouseID int    `json:"warehouse_id"`
	FullName    string `json:"full_name"`
}

// ─────────────────────────────────────────────
// PermissionRepo
// ─────────────────────────────────────────────

type PermissionRepo struct {
	DB *gorm.DB
}

func NewPermissionRepo(db *gorm.DB) *PermissionRepo {
	return &PermissionRepo{DB: db}
}

// GetBlockPermissions returns the grant status of each permission in a block
// for the given employee, scoped to the store's sub-block (or general).
func (r *PermissionRepo) GetBlockPermissions(employeeLoginID, blockID int) ([]PermissionStatus, error) {
	// Raw query: join permissions_combination → permissions to get names,
	// then left-join employee_permission to determine grant status.
	type row struct {
		CombinationID       int
		PermissionName      string
		PermissionsID       int
		SubBlockStoreID     *int
		SubBlockWarehouseID *int
		IsActive            *bool
	}

	var rows []row
	err := r.DB.Raw(`
		SELECT
			pc.id_permissions_combination  AS combination_id,
			p.permission_name,
			p.permissions_id,
			pc.permissions_sub_block_store_id      AS sub_block_store_id,
			pc.permissions_sub_block_warehouse_id  AS sub_block_warehouse_id,
			ep.is_active
		FROM permissions_combination pc
		JOIN permissions p ON p.permissions_id = pc.permissions_id
		LEFT JOIN employee_permission ep
			ON ep.permissions_combination_id = pc.id_permissions_combination
			AND ep.employee_login_id = ?
		WHERE pc.permissions_block_id = ?
		  AND pc.permissions_sub_block_warehouse_id IS NULL
		ORDER BY p.permissions_id, pc.id_permissions_combination
	`, employeeLoginID, blockID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	// Deduplicate: one entry per permissions_id, preferring granted combos.
	// Also filter: only general (no sub_block_store_id) combinations are shown
	// at this level; store-specific filtering happens at the service layer.
	seen := make(map[int]bool)       // permissions_id → already added
	seenGranted := make(map[int]bool) // permissions_id → has a granted row

	// First pass: mark which permissions_ids have a granted combination.
	for _, rw := range rows {
		if rw.IsActive != nil && *rw.IsActive {
			seenGranted[rw.PermissionsID] = true
		}
	}

	result := make([]PermissionStatus, 0, len(rows))
	for _, rw := range rows {
		// Skip warehouse-scoped rows.
		if rw.SubBlockWarehouseID != nil {
			continue
		}
		if seen[rw.PermissionsID] {
			continue
		}
		granted := rw.IsActive != nil && *rw.IsActive
		// If this combo is not granted but another combo for the same permission is
		// granted, skip this row in favour of the granted one (handled below).
		if !granted && seenGranted[rw.PermissionsID] {
			continue
		}
		seen[rw.PermissionsID] = true
		result = append(result, PermissionStatus{
			CombinationID:  rw.CombinationID,
			PermissionName: rw.PermissionName,
			Granted:        granted,
		})
	}
	return result, nil
}

// SetPermission grants or revokes a single permission combination for an employee.
// When grant=true it creates or reactivates the EmployeePermission record.
// When grant=false it deactivates it.
// If the combination's permissions_id equals timecardPermissionID the method
// also activates / deactivates the linked EmployeeTimecardLogin.
func (r *PermissionRepo) SetPermission(employeeLoginID, combinationID int, grant bool, grantedBy int) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		var existing employees.EmployeePermission
		err := tx.Where(
			"employee_login_id = ? AND permissions_combination_id = ?",
			employeeLoginID, combinationID,
		).First(&existing).Error

		if grant {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ep := employees.EmployeePermission{
					EmployeeLoginID:          employeeLoginID,
					PermissionsCombinationID: combinationID,
					GrantedBy:                grantedBy,
					GrantedAt:                &now,
					IsActive:                 true,
				}
				if err := tx.Create(&ep).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else if !existing.IsActive {
				if err := tx.Model(&existing).Updates(map[string]interface{}{
					"is_active":  true,
					"granted_by": grantedBy,
					"granted_at": now,
				}).Error; err != nil {
					return err
				}
			}

			// Mirror to warehouse if not forbidden
			if err := r.mirrorToWarehouse(tx, employeeLoginID, combinationID, true, grantedBy, now); err != nil {
				return err
			}
		} else {
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if err == nil {
				if err := tx.Model(&existing).Update("is_active", false).Error; err != nil {
					return err
				}
			}
			if err := r.mirrorToWarehouse(tx, employeeLoginID, combinationID, false, grantedBy, now); err != nil {
				return err
			}
		}

		// Handle timecard activation
		var combo permission.PermissionsCombination
		if err := tx.First(&combo, "id_permissions_combination = ?", combinationID).Error; err == nil {
			if combo.PermissionsID == timecardPermissionID {
				if err := r.setTimecardLoginActive(tx, employeeLoginID, grant); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// mirrorToWarehouse mirrors a store permission to an equivalent warehouse permission
// when the permission is not in the forbidden list.
func (r *PermissionRepo) mirrorToWarehouse(tx *gorm.DB, employeeLoginID, storeCombinationID int, grant bool, grantedBy int, now time.Time) error {
	var storeCombo permission.PermissionsCombination
	if err := tx.First(&storeCombo, "id_permissions_combination = ?", storeCombinationID).Error; err != nil {
		return nil // not found → nothing to mirror
	}

	if _, forbidden := warehouseForbiddenPermissionIDs[storeCombo.PermissionsID]; forbidden {
		return nil
	}
	if storeCombo.PermissionsSubBlockStoreID == nil {
		return nil // general combination; no warehouse mirror needed here
	}

	// Find warehouse sub-block with the same name as the store sub-block
	var storeSub permission.PermissionsSubBlockStore
	if err := tx.First(&storeSub, "id_permissions_sub_block = ?", *storeCombo.PermissionsSubBlockStoreID).Error; err != nil {
		return nil
	}

	var warehouseSub permission.PermissionsSubBlockWarehouse
	if err := tx.Where("sub_block_name = ?", storeSub.SubBlockName).First(&warehouseSub).Error; err != nil {
		return nil // no matching warehouse sub-block
	}

	// Find the warehouse combination
	var whCombo permission.PermissionsCombination
	if err := tx.Where(
		"permissions_id = ? AND permissions_sub_block_warehouse_id = ?",
		storeCombo.PermissionsID, warehouseSub.IDPermissionsSubBlock,
	).First(&whCombo).Error; err != nil {
		return nil
	}

	var existing employees.EmployeePermission
	err := tx.Where(
		"employee_login_id = ? AND permissions_combination_id = ?",
		employeeLoginID, whCombo.IDPermissionsCombination,
	).First(&existing).Error

	if grant {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ep := employees.EmployeePermission{
				EmployeeLoginID:          employeeLoginID,
				PermissionsCombinationID: whCombo.IDPermissionsCombination,
				GrantedBy:                grantedBy,
				GrantedAt:                &now,
				IsActive:                 true,
			}
			return tx.Create(&ep).Error
		} else if err != nil {
			return err
		} else if !existing.IsActive {
			return tx.Model(&existing).Updates(map[string]interface{}{
				"is_active":  true,
				"granted_by": grantedBy,
				"granted_at": now,
			}).Error
		}
	} else {
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			return tx.Model(&existing).Update("is_active", false).Error
		}
	}
	return nil
}

// setTimecardLoginActive activates or deactivates the EmployeeTimecardLogin
// linked to this employee (matched via employee_id on EmployeeTimecardLogin).
func (r *PermissionRepo) setTimecardLoginActive(tx *gorm.DB, employeeLoginID int, active bool) error {
	// Resolve employee_id from employee_login_id
	var emp employees.Employee
	if err := tx.Where("employee_login_id = ?", employeeLoginID).First(&emp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	empID := emp.IDEmployee
	return tx.Model(&employees.EmployeeTimecardLogin{}).
		Where("employee_id = ?", empID).
		Update("active", active).Error
}

// GetWarehousePermissions returns grant status for each warehouse accessible
// from the given store, for the given employee.
func (r *PermissionRepo) GetWarehousePermissions(employeeLoginID, storeID int) ([]WarehousePermissionStatus, error) {
	type row struct {
		WarehouseID   int
		WarehouseName *string
		Granted       bool
	}

	// A warehouse is "granted" when the employee has at least one active
	// warehouse-scoped permission for it.
	var rows []row
	err := r.DB.Raw(`
		SELECT
			w.id_warehouse  AS warehouse_id,
			w.full_name     AS warehouse_name,
			COALESCE(bool_or(ep.is_active), false) AS granted
		FROM location l
		JOIN warehouse w ON w.id_warehouse = l.warehouse_id
		LEFT JOIN permissions_sub_block_warehouse psbw ON psbw.warehouse_id = w.id_warehouse
		LEFT JOIN permissions_combination pc ON pc.permissions_sub_block_warehouse_id = psbw.id_permissions_sub_block
		LEFT JOIN employee_permission ep
			ON ep.permissions_combination_id = pc.id_permissions_combination
			AND ep.employee_login_id = ?
		WHERE l.store_id = ?
		  AND l.warehouse_id IS NOT NULL
		  AND COALESCE(l.store_active, false) = true
		GROUP BY w.id_warehouse, w.full_name
		ORDER BY w.id_warehouse
	`, employeeLoginID, storeID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]WarehousePermissionStatus, 0, len(rows))
	for _, rw := range rows {
		name := ""
		if rw.WarehouseName != nil {
			name = *rw.WarehouseName
		}
		result = append(result, WarehousePermissionStatus{
			WarehouseID:   rw.WarehouseID,
			WarehouseName: name,
			Granted:       rw.Granted,
		})
	}
	return result, nil
}

// SetWarehousePermission grants or revokes all warehouse-scoped permissions for
// one warehouse for the given employee, subject to:
//  1. The permission must not be in warehouseForbiddenPermissionIDs.
//  2. The employee must already hold the equivalent store permission (grant only).
func (r *PermissionRepo) SetWarehousePermission(employeeLoginID, warehouseID int, grant bool, grantedBy int) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		// Find all warehouse combinations for this warehouse
		var whSubs []permission.PermissionsSubBlockWarehouse
		if err := tx.Where("warehouse_id = ?", warehouseID).Find(&whSubs).Error; err != nil {
			return err
		}

		for _, wSub := range whSubs {
			var combos []permission.PermissionsCombination
			if err := tx.Where(
				"permissions_sub_block_warehouse_id = ?", wSub.IDPermissionsSubBlock,
			).Find(&combos).Error; err != nil {
				return err
			}

			for _, combo := range combos {
				if _, forbidden := warehouseForbiddenPermissionIDs[combo.PermissionsID]; forbidden {
					continue
				}

				if grant {
					// Check employee holds the store permission
					if !r.employeeHasStorePermission(tx, employeeLoginID, combo.PermissionsID) {
						continue
					}
				}

				var existing employees.EmployeePermission
				err := tx.Where(
					"employee_login_id = ? AND permissions_combination_id = ?",
					employeeLoginID, combo.IDPermissionsCombination,
				).First(&existing).Error

				if grant {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						ep := employees.EmployeePermission{
							EmployeeLoginID:          employeeLoginID,
							PermissionsCombinationID: combo.IDPermissionsCombination,
							GrantedBy:                grantedBy,
							GrantedAt:                &now,
							IsActive:                 true,
						}
						if err := tx.Create(&ep).Error; err != nil {
							return err
						}
					} else if err != nil {
						return err
					} else if !existing.IsActive {
						if err := tx.Model(&existing).Updates(map[string]interface{}{
							"is_active":  true,
							"granted_by": grantedBy,
							"granted_at": now,
						}).Error; err != nil {
							return err
						}
					}
				} else {
					if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
						return err
					}
					if err == nil {
						if err := tx.Model(&existing).Update("is_active", false).Error; err != nil {
							return err
						}
					}
				}
			}
		}
		return nil
	})
}

// employeeHasStorePermission checks whether the employee has an active store-level
// (non-warehouse) permission for the given permissions_id.
func (r *PermissionRepo) employeeHasStorePermission(tx *gorm.DB, employeeLoginID, permissionsID int) bool {
	var count int64
	tx.Table("employee_permission ep").
		Joins("JOIN permissions_combination pc ON pc.id_permissions_combination = ep.permissions_combination_id").
		Where("ep.employee_login_id = ?", employeeLoginID).
		Where("pc.permissions_id = ?", permissionsID).
		Where("pc.permissions_sub_block_warehouse_id IS NULL").
		Where("ep.is_active = true").
		Count(&count)
	return count > 0
}

// GetShowcaseLocations returns active showcase locations for the given store.
func (r *PermissionRepo) GetShowcaseLocations(storeID int) ([]LocationItem, error) {
	var locs []location.Location
	err := r.DB.
		Where("store_id = ?", storeID).
		Where("COALESCE(showcase, false) = true").
		Where("COALESCE(store_active, false) = true").
		Find(&locs).Error
	if err != nil {
		return nil, err
	}

	items := make([]LocationItem, 0, len(locs))
	for _, l := range locs {
		items = append(items, LocationItem{
			LocationID:   l.IDLocation,
			LocationName: l.FullName,
		})
	}
	return items, nil
}

// GetWarehousesByLocation returns all warehouses linked to the given location.
func (r *PermissionRepo) GetWarehousesByLocation(locationID int) ([]WarehouseItem, error) {
	type row struct {
		WarehouseID int
		FullName    *string
	}
	var rows []row
	err := r.DB.Raw(`
		SELECT w.id_warehouse, w.full_name
		FROM location l
		JOIN warehouse w ON w.id_warehouse = l.warehouse_id
		WHERE l.id_location = ?
	`, locationID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]WarehouseItem, 0, len(rows))
	for _, rw := range rows {
		name := ""
		if rw.FullName != nil {
			name = *rw.FullName
		}
		items = append(items, WarehouseItem{
			WarehouseID: rw.WarehouseID,
			FullName:    name,
		})
	}
	return items, nil
}

// GetRoles returns all roles.
func (r *PermissionRepo) GetRoles() ([]permission.Role, error) {
	var roles []permission.Role
	if err := r.DB.Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}
