// internal/repository/inventory_repo/batch.go
package inventory_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/inventory"
)

type BatchRepo struct{ DB *gorm.DB }

func NewBatchRepo(db *gorm.DB) *BatchRepo { return &BatchRepo{DB: db} }

// GetByID возвращает партию по ID.
func (r *BatchRepo) GetByID(id int64) (*inventory.Batch, error) {
	var row inventory.Batch
	err := r.DB.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// GetByLocation возвращает все партии для локации.
func (r *BatchRepo) GetByLocation(locationID int64) ([]inventory.Batch, error) {
	var rows []inventory.Batch
	return rows, r.DB.Where("location_id = ?", locationID).Order("created_at DESC").Find(&rows).Error
}

// GetByBrand возвращает партии по бренду в локации.
func (r *BatchRepo) GetByBrand(locationID, brandID int64) ([]inventory.Batch, error) {
	var rows []inventory.Batch
	return rows, r.DB.
		Where("location_id = ? AND brand_id = ?", locationID, brandID).
		Order("created_at DESC").
		Find(&rows).Error
}

// CreateBatchInput — данные новой партии.
type CreateBatchInput struct {
	LocationID        int64
	BrandID           int64
	Qty               int
	Cost              float64
	EmployeeIDPrepBy  int64
	EmployeeIDCreated int64
	Notes             *string
}

// Create создаёт партию.
func (r *BatchRepo) Create(inp CreateBatchInput) (*inventory.Batch, error) {
	now := time.Now()
	b := &inventory.Batch{
		LocationID:        inp.LocationID,
		BrandID:           inp.BrandID,
		Qty:               inp.Qty,
		Cost:              inp.Cost,
		EmployeeIDPrepBy:  inp.EmployeeIDPrepBy,
		EmployeeIDCreated: inp.EmployeeIDCreated,
		EmployeeIDUpdated: inp.EmployeeIDCreated,
		Notes:             inp.Notes,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	return b, r.DB.Create(b).Error
}

// UpdateQty обновляет количество в партии.
func (r *BatchRepo) UpdateQty(id int64, qty int, employeeID int64) error {
	return r.DB.Model(&inventory.Batch{}).Where("id_batch = ?", id).
		Updates(map[string]interface{}{
			"qty":                 qty,
			"employee_id_updated": employeeID,
			"updated_at":          time.Now(),
		}).Error
}

func (r *BatchRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
