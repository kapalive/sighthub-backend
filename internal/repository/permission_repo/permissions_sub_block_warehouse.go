package permission_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/permission"
)

type PermissionsSubBlockWarehouseRepo struct{ DB *gorm.DB }

func NewPermissionsSubBlockWarehouseRepo(db *gorm.DB) *PermissionsSubBlockWarehouseRepo {
	return &PermissionsSubBlockWarehouseRepo{DB: db}
}

func (r *PermissionsSubBlockWarehouseRepo) GetAll() ([]permission.PermissionsSubBlockWarehouse, error) {
	var items []permission.PermissionsSubBlockWarehouse
	return items, r.DB.Find(&items).Error
}

func (r *PermissionsSubBlockWarehouseRepo) GetByID(id int) (*permission.PermissionsSubBlockWarehouse, error) {
	var item permission.PermissionsSubBlockWarehouse
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *PermissionsSubBlockWarehouseRepo) GetByWarehouseID(warehouseID int) ([]permission.PermissionsSubBlockWarehouse, error) {
	var items []permission.PermissionsSubBlockWarehouse
	return items, r.DB.Where("warehouse_id = ?", warehouseID).Find(&items).Error
}

func (r *PermissionsSubBlockWarehouseRepo) Create(p *permission.PermissionsSubBlockWarehouse) error {
	return r.DB.Create(p).Error
}

func (r *PermissionsSubBlockWarehouseRepo) Save(p *permission.PermissionsSubBlockWarehouse) error {
	return r.DB.Save(p).Error
}

func (r *PermissionsSubBlockWarehouseRepo) Delete(id int) error {
	return r.DB.Delete(&permission.PermissionsSubBlockWarehouse{}, id).Error
}
