// internal/repository/inventory_repo/inventory.go
package inventory_repo

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/inventory"
	"sighthub-backend/internal/models/types"
)

type InventoryRepo struct{ DB *gorm.DB }

func NewInventoryRepo(db *gorm.DB) *InventoryRepo { return &InventoryRepo{DB: db} }

// ─────────────────────────────────────────────
// LIST / FILTER
// ─────────────────────────────────────────────

// InventoryListItem — строка в списке инвентаря с данными из price_book.
type InventoryListItem struct {
	IDInventory        int64     `json:"id_inventory"`
	SKU                string    `json:"sku"`
	StatusItems        string    `json:"status_items_inventory"`
	LocationID         int64     `json:"location_id"`
	ModelID            *int64    `json:"model_id,omitempty"`
	InvoiceID          int64     `json:"invoice_id"`
	CreatedDate        time.Time `json:"created_date"`
	// price_book fields
	PbSellingPrice    *float64 `json:"pb_selling_price,omitempty"`
	PbCost            *float64 `json:"pb_cost,omitempty"`
	PbStoreTierPrice  *float64 `json:"pb_store_tier_price,omitempty"`
}

// FilterInput — параметры фильтрации.
type InventoryFilterInput struct {
	LocationID *int64
	VendorID   *int64
	BrandID    *int64
	ModelID    *int64
	Status     *string
	SKU        *string
	Limit      int
	Offset     int
}

// GetByFilter возвращает инвентарь с применением фильтров + данные price_book.
func (r *InventoryRepo) GetByFilter(f InventoryFilterInput) ([]InventoryListItem, error) {
	q := r.DB.
		Table("inventory i").
		Select("i.id_inventory, i.sku, i.status_items_inventory, i.location_id, i.model_id, i.invoice_id, i.created_date, pb.pb_selling_price, pb.pb_cost, pb.pb_store_tier_price").
		Joins("LEFT JOIN price_book pb ON pb.inventory_id = i.id_inventory")

	if f.LocationID != nil { q = q.Where("i.location_id = ?", *f.LocationID) }
	if f.Status != nil     { q = q.Where("i.status_items_inventory = ?", *f.Status) }
	if f.SKU != nil        { q = q.Where("i.sku ILIKE ?", "%"+*f.SKU+"%") }
	if f.ModelID != nil    { q = q.Where("i.model_id = ?", *f.ModelID) }
	if f.BrandID != nil {
		q = q.Joins("JOIN model m ON m.id_model = i.model_id").
			Where("m.brand_id = ?", *f.BrandID)
	}
	if f.VendorID != nil {
		if f.BrandID == nil {
			q = q.Joins("JOIN model m ON m.id_model = i.model_id")
		}
		q = q.Joins("JOIN brand b ON b.id_brand = m.brand_id").
			Where("b.vendor_id = ?", *f.VendorID)
	}

	lim := f.Limit
	if lim <= 0 { lim = 100 }
	q = q.Order("i.created_date DESC").Limit(lim).Offset(f.Offset)

	var rows []InventoryListItem
	return rows, q.Scan(&rows).Error
}

// ─────────────────────────────────────────────
// SINGLE ITEM
// ─────────────────────────────────────────────

// GetByID возвращает единицу инвентаря.
func (r *InventoryRepo) GetByID(id int64) (*inventory.Inventory, error) {
	var row inventory.Inventory
	err := r.DB.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// GetBySKU ищет инвентарь по SKU.
func (r *InventoryRepo) GetBySKU(sku string) (*inventory.Inventory, error) {
	var row inventory.Inventory
	err := r.DB.Where("sku = ?", sku).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

// CreateInput — входные данные для нового инвентаря.
type CreateInventoryInput struct {
	LocationID           int64
	ModelID              *int64
	InvoiceID            int64
	EmployeeID           *int64
	OrdersLensID         *int64
	VariantCRSLProductID *int
	Status               types.StatusItemsInventory
}

// Create создаёт единицу инвентаря, генерирует SKU и создаёт пустую запись в price_book.
func (r *InventoryRepo) Create(inp CreateInventoryInput) (*inventory.Inventory, error) {
	inv := &inventory.Inventory{
		LocationID:           inp.LocationID,
		ModelID:              inp.ModelID,
		InvoiceID:            inp.InvoiceID,
		EmployeeID:           inp.EmployeeID,
		OrdersLensID:         inp.OrdersLensID,
		VariantCRSLProductID: inp.VariantCRSLProductID,
		StatusItemsInventory: inp.Status,
		CreatedDate:          ptrStr(time.Now().Format("15:04:05")),
	}
	inv.GenerateSKU()

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(inv).Error; err != nil {
			return err
		}
		pb := inventory.PriceBook{InventoryID: inv.IDInventory}
		return tx.Create(&pb).Error
	})
	return inv, err
}

// ─────────────────────────────────────────────
// UPDATE STATUS
// ─────────────────────────────────────────────

// UpdateStatus обновляет статус единицы инвентаря.
func (r *InventoryRepo) UpdateStatus(id int64, status types.StatusItemsInventory) error {
	return r.DB.Model(&inventory.Inventory{}).
		Where("id_inventory = ?", id).
		Update("status_items_inventory", status).Error
}

// UpdateState обновляет статус + опциональные поля (invoice_id, location_id).
type UpdateStateInput struct {
	Status     types.StatusItemsInventory
	InvoiceID  *int64
	LocationID *int64
}

func (r *InventoryRepo) UpdateState(id int64, inp UpdateStateInput) error {
	updates := map[string]interface{}{"status_items_inventory": inp.Status}
	if inp.InvoiceID != nil  { updates["invoice_id"]  = *inp.InvoiceID }
	if inp.LocationID != nil { updates["location_id"] = *inp.LocationID }
	return r.DB.Model(&inventory.Inventory{}).Where("id_inventory = ?", id).Updates(updates).Error
}

// ─────────────────────────────────────────────
// SEARCH
// ─────────────────────────────────────────────

// SearchBySKU ищет инвентарь по точному/частичному SKU в локации.
func (r *InventoryRepo) SearchBySKU(sku string, locationID int64) ([]InventoryListItem, error) {
	var rows []InventoryListItem
	err := r.DB.
		Table("inventory i").
		Select("i.id_inventory, i.sku, i.status_items_inventory, i.location_id, i.model_id, i.invoice_id, i.created_date, pb.pb_selling_price, pb.pb_cost, pb.pb_store_tier_price").
		Joins("LEFT JOIN price_book pb ON pb.inventory_id = i.id_inventory").
		Where("i.location_id = ? AND i.sku ILIKE ?", locationID, "%"+sku+"%").
		Limit(20).
		Scan(&rows).Error
	return rows, err
}

// GetItemStatuses возвращает допустимые значения enum status_items_inventory.
func (r *InventoryRepo) GetItemStatuses() []string {
	return []string{
		string(types.StatusInventoryReadyForSale),
		string(types.StatusInventoryDefective),
		string(types.StatusInventoryOnReturn),
		string(types.StatusInventoryICTToReceiveInMN),
		string(types.StatusInventoryICTSentAndNotReceived),
		string(types.StatusInventorySOLD),
		string(types.StatusInventoryMissing),
		string(types.StatusInventoryRemoved),
	}
}

// GenerateSKU возвращает уникальный SKU на основе modelID.
func (r *InventoryRepo) GenerateSKU(modelID int64) string {
	return fmt.Sprintf("%03d/%03d", modelID%1000, time.Now().UnixMilli()%1000)
}

func (r *InventoryRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func ptrStr(s string) *string { return &s }
