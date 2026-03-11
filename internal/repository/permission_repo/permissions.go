package permission_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/permission"
)

type PermissionsRepo struct{ DB *gorm.DB }

func NewPermissionsRepo(db *gorm.DB) *PermissionsRepo { return &PermissionsRepo{DB: db} }

func (r *PermissionsRepo) GetAll() ([]permission.Permissions, error) {
	var items []permission.Permissions
	return items, r.DB.Find(&items).Error
}

func (r *PermissionsRepo) GetByID(id int) (*permission.Permissions, error) {
	var item permission.Permissions
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *PermissionsRepo) GetByBlockID(blockID int) ([]permission.Permissions, error) {
	var items []permission.Permissions
	return items, r.DB.Where("permissions_block_id = ?", blockID).Find(&items).Error
}

func (r *PermissionsRepo) Create(p *permission.Permissions) error {
	return r.DB.Create(p).Error
}

func (r *PermissionsRepo) Save(p *permission.Permissions) error {
	return r.DB.Save(p).Error
}

func (r *PermissionsRepo) Delete(id int) error {
	return r.DB.Delete(&permission.Permissions{}, id).Error
}
