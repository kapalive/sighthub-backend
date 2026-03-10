// internal/repository/inventory_repo/temp_count_inventory.go
package inventory_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/inventory"
)

type TempCountInventoryRepo struct{ DB *gorm.DB }

func NewTempCountInventoryRepo(db *gorm.DB) *TempCountInventoryRepo {
	return &TempCountInventoryRepo{DB: db}
}

// TempCountItem — расширенная строка для отображения позиции в count-листе.
type TempCountItem struct {
	IDTempCount      int       `json:"id_temp_count"`
	InventoryCountID int64     `json:"inventory_count_id"`
	InventoryID      int64     `json:"inventory_id"`
	SKU              string    `json:"sku"`
	LocationID       int       `json:"location_id"`
	BrandID          int       `json:"brand_id"`
	InStock          bool      `json:"in_stock"`
	CountDate        time.Time `json:"count_date"`
	// price_book
	PbSellingPrice *float64 `json:"pb_selling_price,omitempty"`
	PbCost         *float64 `json:"pb_cost,omitempty"`
}

// GetByCountSheet возвращает позиции count-листа с данными инвентаря.
func (r *TempCountInventoryRepo) GetByCountSheet(countSheetID int64) ([]TempCountItem, error) {
	var rows []TempCountItem
	err := r.DB.
		Table("temp_count_inventory t").
		Select("t.id_temp_count, t.inventory_count_id, t.inventory_id, i.sku, t.location_id, t.brand_id, t.in_stock, t.count_date, pb.pb_selling_price, pb.pb_cost").
		Joins("JOIN inventory i ON i.id_inventory = t.inventory_id").
		Joins("LEFT JOIN price_book pb ON pb.inventory_id = t.inventory_id").
		Where("t.inventory_count_id = ?", countSheetID).
		Scan(&rows).Error
	return rows, err
}

// Add добавляет единицу инвентаря в count-лист.
func (r *TempCountInventoryRepo) Add(inv *inventory.TempCountInventory) error {
	inv.CountDate = time.Now()
	return r.DB.Create(inv).Error
}

// SetInStock обновляет флаг in_stock для позиции.
func (r *TempCountInventoryRepo) SetInStock(id int, inStock bool) error {
	return r.DB.Model(&inventory.TempCountInventory{}).
		Where("id_temp_count = ?", id).
		Update("in_stock", inStock).Error
}

// Remove удаляет позицию из count-листа и перемещает в Missing.
func (r *TempCountInventoryRepo) Remove(id int, missingData *inventory.Missing) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&inventory.TempCountInventory{}, id).Error; err != nil {
			return err
		}
		if missingData != nil {
			missingData.ReportedDate = time.Now()
			return tx.Create(missingData).Error
		}
		return nil
	})
}

// ExistsInCount проверяет, есть ли inventoryID уже в данном count-листе.
func (r *TempCountInventoryRepo) ExistsInCount(inventoryID int64, countSheetID int64) (bool, error) {
	var count int64
	err := r.DB.Model(&inventory.TempCountInventory{}).
		Where("inventory_id = ? AND inventory_count_id = ?", inventoryID, countSheetID).
		Count(&count).Error
	return count > 0, err
}

func (r *TempCountInventoryRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
