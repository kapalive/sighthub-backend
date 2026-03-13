// internal/repository/inventory_repo/missing.go
package inventory_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/inventory"
)

type MissingRepo struct{ DB *gorm.DB }

func NewMissingRepo(db *gorm.DB) *MissingRepo { return &MissingRepo{DB: db} }

// GetByCountSheet возвращает все записи "missing" для count-листа.
func (r *MissingRepo) GetByCountSheet(countSheetID int64) ([]inventory.Missing, error) {
	var rows []inventory.Missing
	return rows, r.DB.Where("inventory_count_id = ?", countSheetID).Find(&rows).Error
}

// GetByLocation возвращает все missing-записи для локации.
func (r *MissingRepo) GetByLocation(locationID int64) ([]inventory.Missing, error) {
	var rows []inventory.Missing
	return rows, r.DB.
		Where("location_id = ?", locationID).
		Order("reported_date DESC").
		Find(&rows).Error
}

// CreateMissingInput — данные для записи о потере.
type CreateMissingInput struct {
	InventoryCountID int64
	InventoryID      int64
	LocationID       int64
	BrandID          int64
	ModelID          int64
	Quantity         int
	Cost             float64
	Notes            *string
}

// Create создаёт запись об отсутствующей позиции.
func (r *MissingRepo) Create(inp CreateMissingInput) (*inventory.Missing, error) {
	m := &inventory.Missing{
		InventoryCountID: inp.InventoryCountID,
		InventoryID:      inp.InventoryID,
		LocationID:       inp.LocationID,
		BrandID:          &inp.BrandID,
		ModelID:          inp.ModelID,
		Quantity:         inp.Quantity,
		Cost:             inp.Cost,
		ReportedDate:     time.Now(),
		Notes:            inp.Notes,
	}
	return m, r.DB.Create(m).Error
}

func (r *MissingRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
