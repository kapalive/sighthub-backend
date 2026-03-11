package permission_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/permission"
)

type PermissionsBlockRepo struct{ DB *gorm.DB }

func NewPermissionsBlockRepo(db *gorm.DB) *PermissionsBlockRepo {
	return &PermissionsBlockRepo{DB: db}
}

func (r *PermissionsBlockRepo) GetAll() ([]permission.PermissionsBlock, error) {
	var items []permission.PermissionsBlock
	return items, r.DB.Find(&items).Error
}

func (r *PermissionsBlockRepo) GetByID(id int) (*permission.PermissionsBlock, error) {
	var item permission.PermissionsBlock
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *PermissionsBlockRepo) Create(p *permission.PermissionsBlock) error {
	return r.DB.Create(p).Error
}

func (r *PermissionsBlockRepo) Save(p *permission.PermissionsBlock) error {
	return r.DB.Save(p).Error
}
