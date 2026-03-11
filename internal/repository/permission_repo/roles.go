package permission_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/permission"
)

type RoleRepo struct{ DB *gorm.DB }

func NewRoleRepo(db *gorm.DB) *RoleRepo { return &RoleRepo{DB: db} }

func (r *RoleRepo) GetAll() ([]permission.Role, error) {
	var items []permission.Role
	return items, r.DB.Find(&items).Error
}

func (r *RoleRepo) GetByID(id int) (*permission.Role, error) {
	var item permission.Role
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *RoleRepo) GetByKey(key string) (*permission.Role, error) {
	var item permission.Role
	if err := r.DB.Where("key = ?", key).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *RoleRepo) Create(role *permission.Role) error {
	return r.DB.Create(role).Error
}

func (r *RoleRepo) Save(role *permission.Role) error {
	return r.DB.Save(role).Error
}

func (r *RoleRepo) Delete(id int) error {
	return r.DB.Delete(&permission.Role{}, id).Error
}
