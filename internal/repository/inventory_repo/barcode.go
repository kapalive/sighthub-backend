// internal/repository/inventory_repo/barcode.go
package inventory_repo

import (
	"errors"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/inventory"
)

type BarcodeRepo struct{ DB *gorm.DB }

func NewBarcodeRepo(db *gorm.DB) *BarcodeRepo { return &BarcodeRepo{DB: db} }

// GetByInventoryID возвращает штрих-код для единицы инвентаря.
func (r *BarcodeRepo) GetByInventoryID(inventoryID int64) (*inventory.Barcode, error) {
	var row inventory.Barcode
	err := r.DB.Where("inventory_id = ?", inventoryID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// GetByID возвращает штрих-код по ID.
func (r *BarcodeRepo) GetByID(id int64) (*inventory.Barcode, error) {
	var row inventory.Barcode
	err := r.DB.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// GetByLocation возвращает все штрих-коды для локации через JOIN.
func (r *BarcodeRepo) GetByLocation(locationID int64) ([]inventory.Barcode, error) {
	var rows []inventory.Barcode
	err := r.DB.
		Joins("JOIN inventory ON inventory.id_inventory = barcode.inventory_id").
		Where("inventory.location_id = ?", locationID).
		Find(&rows).Error
	return rows, err
}

// Upsert создаёт или обновляет штрих-код для inventoryID.
func (r *BarcodeRepo) Upsert(b *inventory.Barcode) error {
	existing, err := r.GetByInventoryID(b.InventoryID)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.DB.Create(b).Error
	}
	b.IDBarcode = existing.IDBarcode
	return r.DB.Save(b).Error
}

// Delete удаляет штрих-код.
func (r *BarcodeRepo) Delete(id int64) error {
	return r.DB.Delete(&inventory.Barcode{}, id).Error
}

func (r *BarcodeRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
