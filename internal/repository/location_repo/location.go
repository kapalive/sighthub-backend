// internal/repository/location_repo/location.go
// Операции с Location: поиск по ID, store_id, warehouse_id.
package location_repo

import (
	"errors"

	"gorm.io/gorm"
	locmodel "sighthub-backend/internal/models/location"
)

// LocationRepo — репозиторий для таблицы location.
type LocationRepo struct {
	DB *gorm.DB
}

func NewLocationRepo(db *gorm.DB) *LocationRepo { return &LocationRepo{DB: db} }

// GetByID возвращает Location по первичному ключу.
func (r *LocationRepo) GetByID(id int) (*locmodel.Location, error) {
	var loc locmodel.Location
	err := r.DB.First(&loc, id).Error
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

// GetByStoreID возвращает основной Location для магазина
// (warehouse_id IS NULL — признак основной локации стора).
func (r *LocationRepo) GetByStoreID(storeID int) (*locmodel.Location, error) {
	var loc locmodel.Location
	err := r.DB.Where("store_id = ? AND warehouse_id IS NULL", storeID).First(&loc).Error
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

// GetByWarehouseID возвращает Location, привязанную к warehouse.
func (r *LocationRepo) GetByWarehouseID(warehouseID int) (*locmodel.Location, error) {
	var loc locmodel.Location
	err := r.DB.Where("warehouse_id = ?", warehouseID).First(&loc).Error
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

// GetAllActive возвращает все активные локации.
func (r *LocationRepo) GetAllActive() ([]locmodel.Location, error) {
	var locs []locmodel.Location
	err := r.DB.Where("store_active = true").Find(&locs).Error
	return locs, err
}

// GetAllByStore возвращает все Location для данного store_id (включая warehouse).
func (r *LocationRepo) GetAllByStore(storeID int) ([]locmodel.Location, error) {
	var locs []locmodel.Location
	err := r.DB.Where("store_id = ?", storeID).Find(&locs).Error
	return locs, err
}

// Exists проверяет наличие Location по ID.
func (r *LocationRepo) Exists(id int) (bool, error) {
	var count int64
	err := r.DB.Model(&locmodel.Location{}).Where("id_location = ?", id).Count(&count).Error
	return count > 0, err
}

// IsNotFound возвращает true если ошибка означает "запись не найдена".
func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// findSalesTaxByState ищет id_sales_tax по коду штата.
// Используется в store/warehouse repo при создании Location.
func findSalesTaxByState(db *gorm.DB, stateCode string) *int {
	var id int
	err := db.Table("sales_tax_by_state").
		Select("id_sales_tax").
		Where("state_code = ?", stateCode).
		Scan(&id).Error
	if err != nil || id == 0 {
		return nil
	}
	return &id
}
