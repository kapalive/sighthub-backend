package permission_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/permission"
)

type PermissionsSubBlockStoreRepo struct{ DB *gorm.DB }

func NewPermissionsSubBlockStoreRepo(db *gorm.DB) *PermissionsSubBlockStoreRepo {
	return &PermissionsSubBlockStoreRepo{DB: db}
}

func (r *PermissionsSubBlockStoreRepo) GetAll() ([]permission.PermissionsSubBlockStore, error) {
	var items []permission.PermissionsSubBlockStore
	return items, r.DB.Find(&items).Error
}

func (r *PermissionsSubBlockStoreRepo) GetByID(id int) (*permission.PermissionsSubBlockStore, error) {
	var item permission.PermissionsSubBlockStore
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *PermissionsSubBlockStoreRepo) GetByStoreID(storeID int) ([]permission.PermissionsSubBlockStore, error) {
	var items []permission.PermissionsSubBlockStore
	return items, r.DB.Where("store_id = ?", storeID).Find(&items).Error
}

func (r *PermissionsSubBlockStoreRepo) Create(p *permission.PermissionsSubBlockStore) error {
	return r.DB.Create(p).Error
}

func (r *PermissionsSubBlockStoreRepo) Save(p *permission.PermissionsSubBlockStore) error {
	return r.DB.Save(p).Error
}

func (r *PermissionsSubBlockStoreRepo) Delete(id int) error {
	return r.DB.Delete(&permission.PermissionsSubBlockStore{}, id).Error
}
