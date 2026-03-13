package permission_service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	authModel "sighthub-backend/internal/models/auth"
	empModel "sighthub-backend/internal/models/employees"
	locModel "sighthub-backend/internal/models/location"
	permModel "sighthub-backend/internal/models/permission"
)

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service {
	return &Service{db: db}
}

// WAREHOUSE_FORBIDDEN_PERMISSION_IDS = STORE_PERMISSIONS - WAREHOUSE_PERMISSIONS - COMMON_PERMISSIONS
var warehouseForbiddenPermIDs = map[int]bool{
	1: true, 2: true, 3: true, 4: true, 5: true,
	6: true, 7: true, 8: true, 9: true, 10: true,
	31: true, 33: true, 34: true, 35: true,
}

type PermissionItem struct {
	CombinationID  int    `json:"combination_id"`
	PermissionName string `json:"permission_name"`
	Granted        bool   `json:"granted"`
}

type WarehouseAccessItem struct {
	WarehouseID   int    `json:"warehouse_id"`
	WarehouseName string `json:"warehouse_name"`
	Granted       bool   `json:"granted"`
}

// GetBlockPermissions returns employee permissions for a specific block at a location.
// Shows GENERAL + STORE combinations only (not warehouse).
func (s *Service) GetBlockPermissions(employeeID, blockID, locationID int) ([]PermissionItem, error) {
	var loc locModel.Location
	if err := s.db.First(&loc, locationID).Error; err != nil {
		return nil, errors.New("location not found")
	}

	employeeLoginID, err := s.getEmployeeLoginID(employeeID)
	if err != nil {
		return nil, err
	}

	// Get store subblock IDs for this location
	var storeSubBlockIDs []int
	if loc.StoreID != 0 {
		var subs []permModel.PermissionsSubBlockStore
		s.db.Where("store_id = ?", loc.StoreID).Find(&subs)
		for _, sb := range subs {
			storeSubBlockIDs = append(storeSubBlockIDs, sb.IDPermissionsSubBlock)
		}
	}

	// Build combination query — block + (general OR store-specific), no warehouse
	query := s.db.Table("permissions_combination pc").
		Select("pc.id_permissions_combination, pc.permissions_id, p.permission_name").
		Joins("JOIN permissions p ON p.permissions_id = pc.permissions_id").
		Where("pc.permissions_block_id = ?", blockID).
		Where("pc.permissions_sub_block_warehouse_id IS NULL")

	if len(storeSubBlockIDs) > 0 {
		query = query.Where(
			"(pc.permissions_sub_block_store_id IS NULL) OR (pc.permissions_sub_block_store_id IN ?)",
			storeSubBlockIDs,
		)
	} else {
		query = query.Where("pc.permissions_sub_block_store_id IS NULL")
	}

	type raw struct {
		IDPermissionsCombination int
		PermissionsID            int
		PermissionName           string
	}
	var rows []raw
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	// Deduplicate by permissions_id, keep first combo_id seen
	seen := map[int]raw{}
	order := []int{}
	for _, r := range rows {
		if _, exists := seen[r.PermissionsID]; !exists {
			seen[r.PermissionsID] = r
			order = append(order, r.PermissionsID)
		}
	}

	result := make([]PermissionItem, 0, len(order))
	for _, pid := range order {
		r := seen[pid]
		var ep empModel.EmployeePermission
		granted := s.db.Where("employee_login_id = ? AND permissions_combination_id = ? AND is_active = true",
			employeeLoginID, r.IDPermissionsCombination).First(&ep).Error == nil
		result = append(result, PermissionItem{
			CombinationID:  r.IDPermissionsCombination,
			PermissionName: r.PermissionName,
			Granted:        granted,
		})
	}
	return result, nil
}

// SetBlockPermission grants or revokes a specific combination, cascading to matching warehouse combos.
func (s *Service) SetBlockPermission(grantorUsername string, employeeID, blockID, locationID, combinationID int, grant bool) error {
	employeeLoginID, err := s.getEmployeeLoginID(employeeID)
	if err != nil {
		return err
	}

	var grantor authModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", grantorUsername).First(&grantor).Error; err != nil {
		return errors.New("granter not found")
	}

	var combo permModel.PermissionsCombination
	if err := s.db.First(&combo, combinationID).Error; err != nil {
		return errors.New("invalid permissions combination")
	}

	// 1. Grant/revoke the exact combination
	s.grantPermForCombo(employeeLoginID, combo.IDPermissionsCombination, grant, grantor.IDEmployeeLogin)

	// 2. Timecard special handling (perm 50)
	if combo.PermissionsID == 50 {
		var tc empModel.EmployeeTimecardLogin
		if s.db.Where("employee_id = ?", employeeID).First(&tc).Error == nil {
			tc.Active = grant
			s.db.Save(&tc)
		}
	}

	// 3. Cascade to warehouse combos if this is a store subblock combo and perm is not warehouse-forbidden
	if combo.PermissionsSubBlockStoreID != nil && !warehouseForbiddenPermIDs[combo.PermissionsID] {
		var storeSub permModel.PermissionsSubBlockStore
		if s.db.First(&storeSub, *combo.PermissionsSubBlockStoreID).Error == nil {
			// Find all warehouse locations for this store
			var whLocs []locModel.Location
			s.db.Where("store_id = ? AND warehouse_id IS NOT NULL", storeSub.StoreID).Find(&whLocs)

			for _, wl := range whLocs {
				var whSub permModel.PermissionsSubBlockWarehouse
				if s.db.Where("sub_block_name = ? AND warehouse_id = ?", storeSub.SubBlockName, wl.WarehouseID).First(&whSub).Error != nil {
					continue
				}
				var whCombo permModel.PermissionsCombination
				if s.db.Where("permissions_block_id = ? AND permissions_id = ? AND permissions_sub_block_warehouse_id = ? AND permissions_sub_block_store_id IS NULL",
					blockID, combo.PermissionsID, whSub.IDPermissionsSubBlock).First(&whCombo).Error != nil {
					continue
				}
				s.grantPermForCombo(employeeLoginID, whCombo.IDPermissionsCombination, grant, grantor.IDEmployeeLogin)
			}
		}
	}

	return nil
}

// GetWarehouseAccess returns warehouse access status for employee at a location's store.
func (s *Service) GetWarehouseAccess(employeeID, locationID int) ([]WarehouseAccessItem, error) {
	var loc locModel.Location
	if err := s.db.First(&loc, locationID).Error; err != nil {
		return nil, errors.New("location not found")
	}

	employeeLoginID, err := s.getEmployeeLoginID(employeeID)
	if err != nil {
		return nil, err
	}

	var whLocs []locModel.Location
	s.db.Where("store_id = ? AND warehouse_id IS NOT NULL", loc.StoreID).Preload("Warehouse").Find(&whLocs)

	result := make([]WarehouseAccessItem, 0, len(whLocs))
	for _, wl := range whLocs {
		if wl.Warehouse == nil {
			continue
		}
		var whSub permModel.PermissionsSubBlockWarehouse
		granted := false
		if s.db.Where("warehouse_id = ?", wl.WarehouseID).First(&whSub).Error == nil {
			var ep empModel.EmployeePermission
			granted = s.db.Joins("JOIN permissions_combination pc ON pc.id_permissions_combination = employee_permission.permissions_combination_id").
				Where("employee_permission.employee_login_id = ? AND employee_permission.is_active = true AND pc.permissions_sub_block_warehouse_id = ?",
					employeeLoginID, whSub.IDPermissionsSubBlock).
				First(&ep).Error == nil
		}
		whName := ""
		if wl.Warehouse.FullName != nil {
			whName = *wl.Warehouse.FullName
		}
		result = append(result, WarehouseAccessItem{
			WarehouseID:   *wl.WarehouseID,
			WarehouseName: whName,
			Granted:       granted,
		})
	}
	return result, nil
}

// SetWarehouseAccess grants or revokes all warehouse combos that intersect with employee store permissions.
func (s *Service) SetWarehouseAccess(grantorUsername string, employeeID, locationID, warehouseID int, grant bool) error {
	var loc locModel.Location
	if err := s.db.First(&loc, locationID).Error; err != nil {
		return errors.New("location not found")
	}

	employeeLoginID, err := s.getEmployeeLoginID(employeeID)
	if err != nil {
		return err
	}

	var grantor authModel.EmployeeLogin
	if err := s.db.Where("employee_login = ?", grantorUsername).First(&grantor).Error; err != nil {
		return errors.New("current user not found")
	}

	var whSub permModel.PermissionsSubBlockWarehouse
	if err := s.db.Where("warehouse_id = ?", warehouseID).First(&whSub).Error; err != nil {
		return errors.New("warehouse not found")
	}

	// All warehouse combos for this sub_block
	var whCombos []permModel.PermissionsCombination
	s.db.Where("permissions_sub_block_warehouse_id = ?", whSub.IDPermissionsSubBlock).Find(&whCombos)

	// Which permission_ids employee has active at the store
	type storePermRow struct {
		PermissionsID int
	}
	var storePerms []storePermRow
	s.db.Table("employee_permission ep").
		Select("pc.permissions_id").
		Joins("JOIN permissions_combination pc ON pc.id_permissions_combination = ep.permissions_combination_id").
		Where("ep.employee_login_id = ? AND ep.is_active = true AND pc.permissions_sub_block_store_id IS NOT NULL", employeeLoginID).
		Scan(&storePerms)
	storePermSet := map[int]bool{}
	for _, sp := range storePerms {
		storePermSet[sp.PermissionsID] = true
	}

	for _, combo := range whCombos {
		if storePermSet[combo.PermissionsID] && !warehouseForbiddenPermIDs[combo.PermissionsID] {
			s.grantPermForCombo(employeeLoginID, combo.IDPermissionsCombination, grant, grantor.IDEmployeeLogin)
		}
	}
	return nil
}

// AssignDefaultPermissionsForRoles assigns default permission combinations for roles at a store location.
// Used during add/update employee. Runs inside caller's transaction.
func (s *Service) AssignDefaultPermissionsForRoles(tx *gorm.DB, employeeLoginID, locationID int, roleIDs []int, grantByID int) {
	if tx == nil {
		tx = s.db
	}
	var loc locModel.Location
	if tx.First(&loc, locationID).Error != nil {
		return
	}

	var storeSubBlockIDs []int
	var storeSubs []permModel.PermissionsSubBlockStore
	tx.Where("store_id = ?", loc.StoreID).Find(&storeSubs)
	for _, sb := range storeSubs {
		storeSubBlockIDs = append(storeSubBlockIDs, sb.IDPermissionsSubBlock)
	}

	for _, roleID := range roleIDs {
		var rolePerms []permModel.RolePermission
		tx.Where("role_id = ?", roleID).Find(&rolePerms)
		for _, rp := range rolePerms {
			query := tx.Where("permissions_id = ?", rp.PermissionsID)
			if len(storeSubBlockIDs) > 0 {
				query = query.Where("permissions_sub_block_store_id IN ? OR permissions_sub_block_store_id IS NULL", storeSubBlockIDs)
			} else {
				query = query.Where("permissions_sub_block_store_id IS NULL")
			}
			var combos []permModel.PermissionsCombination
			query.Find(&combos)
			for _, combo := range combos {
				var existing empModel.EmployeePermission
				if tx.Where("employee_login_id = ? AND permissions_combination_id = ?",
					employeeLoginID, combo.IDPermissionsCombination).First(&existing).Error != nil {
					tx.Create(&empModel.EmployeePermission{
						EmployeeLoginID:          employeeLoginID,
						PermissionsCombinationID: combo.IDPermissionsCombination,
						GrantedBy:                grantByID,
						IsActive:                 true,
					})
				}
			}
		}
	}
}

// grantPermForCombo upserts a single EmployeePermission record.
func (s *Service) grantPermForCombo(employeeLoginID, combinationID int, grant bool, granterID int) {
	var ep empModel.EmployeePermission
	err := s.db.Where("employee_login_id = ? AND permissions_combination_id = ?",
		employeeLoginID, combinationID).First(&ep).Error
	if grant {
		now := time.Now()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.db.Create(&empModel.EmployeePermission{
				EmployeeLoginID:          employeeLoginID,
				PermissionsCombinationID: combinationID,
				GrantedBy:                granterID,
				GrantedAt:                &now,
				IsActive:                 true,
			})
		} else {
			ep.IsActive = true
			ep.GrantedBy = granterID
			ep.GrantedAt = &now
			s.db.Save(&ep)
		}
	} else {
		if err == nil {
			ep.IsActive = false
			s.db.Save(&ep)
		}
	}
}

func (s *Service) getEmployeeLoginID(employeeID int) (int, error) {
	var emp empModel.Employee
	if err := s.db.First(&emp, employeeID).Error; err != nil {
		return 0, errors.New("employee not found")
	}
	return int(emp.EmployeeLoginID), nil
}
