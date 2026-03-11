package permission_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/permission"
)

type RolePermissionRepo struct{ DB *gorm.DB }

func NewRolePermissionRepo(db *gorm.DB) *RolePermissionRepo {
	return &RolePermissionRepo{DB: db}
}

func (r *RolePermissionRepo) GetByID(id int) (*permission.RolePermission, error) {
	var item permission.RolePermission
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *RolePermissionRepo) GetByRoleID(roleID int) ([]permission.RolePermission, error) {
	var items []permission.RolePermission
	return items, r.DB.Where("role_id = ?", roleID).Find(&items).Error
}

// GetPermissionIDsByRole returns just the permissions_id slice — useful for building sets.
func (r *RolePermissionRepo) GetPermissionIDsByRole(roleID int) ([]int, error) {
	var ids []int
	return ids, r.DB.Model(&permission.RolePermission{}).
		Where("role_id = ?", roleID).
		Pluck("permissions_id", &ids).Error
}

func (r *RolePermissionRepo) Add(roleID, permissionsID int) error {
	return r.DB.Create(&permission.RolePermission{
		RoleID:        roleID,
		PermissionsID: permissionsID,
	}).Error
}

func (r *RolePermissionRepo) Remove(id int) error {
	return r.DB.Delete(&permission.RolePermission{}, id).Error
}

func (r *RolePermissionRepo) RemoveByRoleAndPermission(roleID, permissionsID int) error {
	return r.DB.Where("role_id = ? AND permissions_id = ?", roleID, permissionsID).
		Delete(&permission.RolePermission{}).Error
}

func (r *RolePermissionRepo) RemoveAllForRole(roleID int) error {
	return r.DB.Where("role_id = ?", roleID).Delete(&permission.RolePermission{}).Error
}
