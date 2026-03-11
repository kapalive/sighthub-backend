package permission_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/permission"
)

type TableAccessRepo struct{ DB *gorm.DB }

func NewTableAccessRepo(db *gorm.DB) *TableAccessRepo {
	return &TableAccessRepo{DB: db}
}

func (r *TableAccessRepo) GetByID(id int) (*permission.TableAccess, error) {
	var item permission.TableAccess
	if err := r.DB.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *TableAccessRepo) GetByRoleID(roleID int) ([]permission.TableAccess, error) {
	var items []permission.TableAccess
	return items, r.DB.Where("role_id = ?", roleID).Find(&items).Error
}

func (r *TableAccessRepo) GetByRoleAndTable(roleID int, tableName string) ([]permission.TableAccess, error) {
	var items []permission.TableAccess
	return items, r.DB.Where("role_id = ? AND table_name = ?", roleID, tableName).Find(&items).Error
}

// HasAccess checks if a role has a given permission on a table.
func (r *TableAccessRepo) HasAccess(roleID int, tableName string, permissionsID int) (bool, error) {
	var count int64
	err := r.DB.Model(&permission.TableAccess{}).
		Where("role_id = ? AND table_name = ? AND permissions_id = ?", roleID, tableName, permissionsID).
		Count(&count).Error
	return count > 0, err
}

func (r *TableAccessRepo) Create(ta *permission.TableAccess) error {
	return r.DB.Create(ta).Error
}

func (r *TableAccessRepo) Delete(id int) error {
	return r.DB.Delete(&permission.TableAccess{}, id).Error
}

func (r *TableAccessRepo) DeleteByRoleAndTable(roleID int, tableName string) error {
	return r.DB.Where("role_id = ? AND table_name = ?", roleID, tableName).
		Delete(&permission.TableAccess{}).Error
}
