// internal/repository/inventory_repo/inventory_count.go
package inventory_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/inventory"
)

type InventoryCountRepo struct{ DB *gorm.DB }

func NewInventoryCountRepo(db *gorm.DB) *InventoryCountRepo {
	return &InventoryCountRepo{DB: db}
}

// GetAll возвращает все count-листы для локации.
func (r *InventoryCountRepo) GetAll(locationID int64) ([]inventory.InventoryCount, error) {
	var rows []inventory.InventoryCount
	return rows, r.DB.
		Where("location_id = ?", locationID).
		Order("created_date DESC").
		Find(&rows).Error
}

// GetByID возвращает count-лист по ID.
func (r *InventoryCountRepo) GetByID(id int64) (*inventory.InventoryCount, error) {
	var row inventory.InventoryCount
	err := r.DB.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// GetOpenByBrandAndLocation возвращает открытый count-лист для бренда в локации.
func (r *InventoryCountRepo) GetOpenByBrandAndLocation(brandID, locationID int64) (*inventory.InventoryCount, error) {
	var row inventory.InventoryCount
	err := r.DB.
		Where("brand_id = ? AND location_id = ? AND status = true", brandID, locationID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// CreateCountSheetInput — данные для нового count-листа.
type CreateCountSheetInput struct {
	BrandID           int64
	LocationID        int64
	PrepByDate        time.Time
	PrepByEmployeeID  int64
	CreatedEmployeeID int64
	Quantity          int
	Cost              float64
	Notes             *string
}

// Create создаёт count-лист (статус OPEN=true).
func (r *InventoryCountRepo) Create(inp CreateCountSheetInput) (*inventory.InventoryCount, error) {
	now := time.Now()
	ic := &inventory.InventoryCount{
		BrandID:           inp.BrandID,
		LocationID:        inp.LocationID,
		Status:            true,
		PrepByDate:        inp.PrepByDate,
		PrepByEmployeeID:  inp.PrepByEmployeeID,
		CreatedDate:       now,
		CreatedEmployeeID: inp.CreatedEmployeeID,
		UpdatedDate:       now,
		UpdatedEmployeeID: inp.CreatedEmployeeID,
		Quantity:          inp.Quantity,
		Cost:              inp.Cost,
		Notes:             inp.Notes,
	}
	return ic, r.DB.Create(ic).Error
}

// UpdateNotes обновляет заметки count-листа.
func (r *InventoryCountRepo) UpdateNotes(id int64, notes *string, employeeID int64) error {
	return r.DB.Model(&inventory.InventoryCount{}).Where("id_inventory_count = ?", id).
		Updates(map[string]interface{}{
			"notes":               notes,
			"updated_employee_id": employeeID,
			"updated_date":        time.Now(),
		}).Error
}

// Close закрывает count-лист (статус CLOSED=false) и очищает temp-таблицу.
func (r *InventoryCountRepo) Close(id int64, employeeID int64) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&inventory.InventoryCount{}).Where("id_inventory_count = ?", id).
			Updates(map[string]interface{}{
				"status":              false,
				"updated_employee_id": employeeID,
				"updated_date":        time.Now(),
			}).Error; err != nil {
			return err
		}
		return tx.Where("inventory_count_id = ?", id).Delete(&inventory.TempCountInventory{}).Error
	})
}

// Delete удаляет count-лист (только закрытые).
func (r *InventoryCountRepo) Delete(id int64) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("inventory_count_id = ?", id).Delete(&inventory.TempCountInventory{})
		return tx.Delete(&inventory.InventoryCount{}, id).Error
	})
}

func (r *InventoryCountRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
