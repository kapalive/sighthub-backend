// internal/repository/inventory_repo/shipment.go
package inventory_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/inventory"
)

type ShipmentRepo struct{ DB *gorm.DB }

func NewShipmentRepo(db *gorm.DB) *ShipmentRepo { return &ShipmentRepo{DB: db} }

// GetAll возвращает все отгрузки для локации.
func (r *ShipmentRepo) GetAll(locationID int64) ([]inventory.Shipment, error) {
	var rows []inventory.Shipment
	return rows, r.DB.
		Where("location_id = ?", locationID).
		Order("date_received DESC").
		Find(&rows).Error
}

// GetByID возвращает отгрузку по ID.
func (r *ShipmentRepo) GetByID(id int64) (*inventory.Shipment, error) {
	var row inventory.Shipment
	err := r.DB.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// GetByVendorInvoice возвращает отгрузки по vendor invoice ID.
func (r *ShipmentRepo) GetByVendorInvoice(vendorInvoiceID int64) ([]inventory.Shipment, error) {
	var rows []inventory.Shipment
	return rows, r.DB.Where("vendor_invoice_id = ?", vendorInvoiceID).Find(&rows).Error
}

// CreateShipmentInput — данные для новой отгрузки.
type CreateShipmentInput struct {
	VendorID          int64
	LocationID        int64
	BrandID           int64
	QtyOk             int
	QtyHold           int
	QtyShort          int
	QtyOver           int
	Cost              float64
	EmployeeIDPrepBy  int64
	EmployeeIDCreated int64
	Status            *string
	Notes             *string
	VendorInvoiceID   *int64
}

// Create создаёт запись об отгрузке.
func (r *ShipmentRepo) Create(inp CreateShipmentInput) (*inventory.Shipment, error) {
	s := &inventory.Shipment{
		VendorID:          inp.VendorID,
		LocationID:        inp.LocationID,
		BrandID:           inp.BrandID,
		QtyOk:             inp.QtyOk,
		QtyHold:           inp.QtyHold,
		QtyShort:          inp.QtyShort,
		QtyOver:           inp.QtyOver,
		Cost:              inp.Cost,
		EmployeeIDPrepBy:  inp.EmployeeIDPrepBy,
		EmployeeIDCreated: inp.EmployeeIDCreated,
		DateReceived:      time.Now(),
		Status:            inp.Status,
		Notes:             inp.Notes,
		VendorInvoiceID:   inp.VendorInvoiceID,
	}
	return s, r.DB.Create(s).Error
}

// UpdateShipmentInput — изменяемые поля.
type UpdateShipmentInput struct {
	QtyOk           *int
	QtyHold         *int
	QtyShort        *int
	QtyOver         *int
	Cost            *float64
	Status          *string
	Notes           *string
	VendorInvoiceID *int64
}

// Update обновляет запись об отгрузке.
func (r *ShipmentRepo) Update(id int64, inp UpdateShipmentInput) error {
	updates := map[string]interface{}{}
	if inp.QtyOk != nil           { updates["qty_ok"]            = *inp.QtyOk }
	if inp.QtyHold != nil         { updates["qty_hold"]          = *inp.QtyHold }
	if inp.QtyShort != nil        { updates["qty_short"]         = *inp.QtyShort }
	if inp.QtyOver != nil         { updates["qty_over"]          = *inp.QtyOver }
	if inp.Cost != nil            { updates["cost"]              = *inp.Cost }
	if inp.Status != nil          { updates["status"]            = *inp.Status }
	if inp.Notes != nil           { updates["notes"]             = *inp.Notes }
	if inp.VendorInvoiceID != nil { updates["vendor_invoice_id"] = *inp.VendorInvoiceID }
	if len(updates) == 0 {
		return nil
	}
	return r.DB.Model(&inventory.Shipment{}).Where("id_shipment = ?", id).Updates(updates).Error
}

func (r *ShipmentRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
