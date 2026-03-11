package permission_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/permission"
)

type PermissionsCombinationRepo struct{ DB *gorm.DB }

func NewPermissionsCombinationRepo(db *gorm.DB) *PermissionsCombinationRepo {
	return &PermissionsCombinationRepo{DB: db}
}

func (r *PermissionsCombinationRepo) GetAll() ([]permission.PermissionsCombination, error) {
	var items []permission.PermissionsCombination
	return items, r.DB.Find(&items).Error
}

func (r *PermissionsCombinationRepo) GetByID(id int) (*permission.PermissionsCombination, error) {
	var item permission.PermissionsCombination
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *PermissionsCombinationRepo) GetByPermissionsID(permissionsID int) ([]permission.PermissionsCombination, error) {
	var items []permission.PermissionsCombination
	return items, r.DB.Where("permissions_id = ?", permissionsID).Find(&items).Error
}

func (r *PermissionsCombinationRepo) GetByBlockAndPermission(blockID, permissionsID int) ([]permission.PermissionsCombination, error) {
	var items []permission.PermissionsCombination
	return items, r.DB.Where("permissions_block_id = ? AND permissions_id = ?", blockID, permissionsID).Find(&items).Error
}

func (r *PermissionsCombinationRepo) Create(p *permission.PermissionsCombination) error {
	return r.DB.Create(p).Error
}

func (r *PermissionsCombinationRepo) Delete(id int) error {
	return r.DB.Delete(&permission.PermissionsCombination{}, id).Error
}
