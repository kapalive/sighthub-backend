// internal/repository/location_repo/warehouse.go
// CRUD для Warehouse + связанных Location/Permissions.
// Аналог warehouse-части store_routes.py.
package location_repo

import (
	"strings"

	"gorm.io/gorm"
	locmodel "sighthub-backend/internal/models/location"
	permmodel "sighthub-backend/internal/models/permission"
)

// defaultWarehousePermissionIDs — ID прав по умолчанию для нового склада.
var defaultWarehousePermissionIDs = []int{
	11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28,
	29, 30, 32, 41, 42, 43, 44, 45, 46, 47, 62, 63, 64, 65, 66, 67, 68, 69,
	70, 71, 72, 73, 74, 75, 76, 77, 78, 79,
}

// ─────────────────────── input/response types ──────────────

// CreateWarehouseInput — входные данные для POST /warehouses.
type CreateWarehouseInput struct {
	ShortName     string  `json:"short_name"`
	WarehouseName string  `json:"warehouse_name"`
	StoreID       int     `json:"store_id"`
	StreetAddress *string `json:"street_address"`
	AddressLine2  *string `json:"address_line_2"`
	City          *string `json:"city"`
	State         *string `json:"state"`
	PostalCode    *string `json:"postal_code"`
	Country       *string `json:"country"`
	Phone         *string `json:"phone"`
	Email         *string `json:"email"`
	CanReceiveItems bool  `json:"can_receive_items"`
	Active        bool    `json:"active"`
}

// UpdateWarehouseInput — входные данные для PUT /warehouses/:id.
type UpdateWarehouseInput struct {
	ShortName     *string `json:"short_name"`
	WarehouseName *string `json:"warehouse_name"`
	StoreID       *int    `json:"store_id"`
	StreetAddress *string `json:"street_address"`
	AddressLine2  *string `json:"address_line_2"`
	City          *string `json:"city"`
	State         *string `json:"state"`
	PostalCode    *string `json:"postal_code"`
	Phone         *string `json:"phone"`
	Email         *string `json:"email"`
	CanReceiveItems *bool  `json:"can_receive_items"`
	Active        *bool   `json:"active"`
}

// WarehouseListItem — элемент списка для GET /warehouses.
type WarehouseListItem struct {
	IDWarehouse   int     `json:"warehouse_id"`
	ShortName     *string `json:"short_name"`
	WarehouseName *string `json:"warehouse_name"`
	ConnectStore  *string `json:"connect_store"`
	StoreID       *int    `json:"store_id"`
	Address       string  `json:"address"`
	Phone         *string `json:"phone"`
	Active        *bool   `json:"active"`
}

// WarehouseDetail — детальный ответ для GET /warehouses/:id.
type WarehouseDetail struct {
	IDWarehouse   int     `json:"warehouse_id"`
	ShortName     *string `json:"short_name"`
	WarehouseName *string `json:"warehouse_name"`
	ConnectStore  *string `json:"connect_store"`
	StoreID       *int    `json:"store_id"`
	StreetAddress *string `json:"street_address"`
	AddressLine2  *string `json:"address_line_2"`
	City          *string `json:"city"`
	State         *string `json:"state"`
	PostalCode    *string `json:"postal_code"`
	Phone         *string `json:"phone"`
	Email         *string `json:"email"`
	CanReceiveItems *bool  `json:"can_receive_items"`
	Active        *bool   `json:"active"`
}

// ─────────────────────── repo ──────────────────────────────

// WarehouseRepo — репозиторий для Warehouse.
type WarehouseRepo struct {
	DB *gorm.DB
}

func NewWarehouseRepo(db *gorm.DB) *WarehouseRepo { return &WarehouseRepo{DB: db} }

// GetAll возвращает список всех складов.
func (r *WarehouseRepo) GetAll() ([]WarehouseListItem, error) {
	var warehouses []locmodel.Warehouse
	if err := r.DB.Find(&warehouses).Error; err != nil {
		return nil, err
	}

	result := make([]WarehouseListItem, 0, len(warehouses))
	for _, wh := range warehouses {
		item := WarehouseListItem{
			IDWarehouse:   wh.IDWarehouse,
			ShortName:     wh.ShortName,
			WarehouseName: wh.FullName,
			Phone:         wh.Phone,
		}

		var loc locmodel.Location
		if err := r.DB.Where("warehouse_id = ?", wh.IDWarehouse).First(&loc).Error; err == nil {
			// Адрес из location (приоритет)
			parts := []string{}
			for _, p := range []*string{loc.StreetAddress, loc.AddressLine2, loc.City, loc.State, loc.PostalCode, loc.Country} {
				if p != nil && *p != "" {
					parts = append(parts, *p)
				}
			}
			item.Address = strings.Join(parts, ", ")

			if loc.Phone != nil {
				item.Phone = loc.Phone
			}
			item.Active = loc.StoreActive

			storeID := loc.StoreID
			item.StoreID = &storeID

			var store locmodel.Store
			if err2 := r.DB.First(&store, loc.StoreID).Error; err2 == nil {
				item.ConnectStore = store.FullName
			}
		} else {
			// Адрес из самого warehouse
			parts := []string{}
			for _, p := range []*string{wh.StreetAddress, wh.AddressLine2, wh.City, wh.State, wh.PostalCode, wh.Country} {
				if p != nil && *p != "" {
					parts = append(parts, *p)
				}
			}
			item.Address = strings.Join(parts, ", ")
		}

		result = append(result, item)
	}
	return result, nil
}

// GetByID возвращает детальную информацию о складе.
func (r *WarehouseRepo) GetByID(warehouseID int) (*WarehouseDetail, error) {
	var wh locmodel.Warehouse
	if err := r.DB.First(&wh, warehouseID).Error; err != nil {
		return nil, err
	}

	detail := &WarehouseDetail{
		IDWarehouse:   wh.IDWarehouse,
		ShortName:     wh.ShortName,
		WarehouseName: wh.FullName,
	}

	var loc locmodel.Location
	if err := r.DB.Where("warehouse_id = ?", warehouseID).First(&loc).Error; err == nil {
		storeID := loc.StoreID
		detail.StoreID = &storeID
		detail.StreetAddress = loc.StreetAddress
		detail.AddressLine2 = loc.AddressLine2
		detail.City = loc.City
		detail.State = loc.State
		detail.PostalCode = loc.PostalCode
		detail.Phone = loc.Phone
		detail.Email = loc.Email
		detail.CanReceiveItems = loc.CanReceiveItems
		detail.Active = loc.StoreActive

		var store locmodel.Store
		if err2 := r.DB.First(&store, loc.StoreID).Error; err2 == nil {
			detail.ConnectStore = store.FullName
		}
	} else {
		// Нет location — берём из самого warehouse
		detail.StreetAddress = wh.StreetAddress
		detail.AddressLine2 = wh.AddressLine2
		detail.City = wh.City
		detail.State = wh.State
		detail.PostalCode = wh.PostalCode
		detail.Phone = wh.Phone
	}

	return detail, nil
}

// Create создаёт Warehouse + Location + PermissionsSubBlockWarehouse в одной транзакции.
func (r *WarehouseRepo) Create(input CreateWarehouseInput) (int, error) {
	var newID int

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		sn := input.ShortName
		fn := input.WarehouseName

		// 1. Warehouse
		wh := locmodel.Warehouse{
			FullName:      &fn,
			ShortName:     &sn,
			StreetAddress: input.StreetAddress,
			AddressLine2:  input.AddressLine2,
			City:          input.City,
			State:         input.State,
			PostalCode:    input.PostalCode,
			Country:       input.Country,
			Phone:         input.Phone,
		}
		if err := tx.Create(&wh).Error; err != nil {
			return err
		}

		// 2. Location
		canReceive := input.CanReceiveItems
		active := input.Active
		loc := locmodel.Location{
			FullName:        fn,
			ShortName:       &sn,
			StreetAddress:   input.StreetAddress,
			AddressLine2:    input.AddressLine2,
			City:            input.City,
			State:           input.State,
			PostalCode:      input.PostalCode,
			Country:         input.Country,
			Phone:           input.Phone,
			Email:           input.Email,
			CanReceiveItems: &canReceive,
			StoreActive:     &active,
			StoreID:         input.StoreID,
			WarehouseID:     &wh.IDWarehouse,
		}
		if err := tx.Create(&loc).Error; err != nil {
			return err
		}

		// 3. PermissionsSubBlockWarehouse
		whID := wh.IDWarehouse
		sub := permmodel.PermissionsSubBlockWarehouse{
			SubBlockName: fn,
			WarehouseID:  &whID,
		}
		if err := tx.Create(&sub).Error; err != nil {
			return err
		}

		// 4. Bulk-insert PermissionsCombination
		if err := insertWarehousePermissions(tx, sub.IDPermissionsSubBlock, defaultWarehousePermissionIDs); err != nil {
			return err
		}

		newID = wh.IDWarehouse
		return nil
	})

	return newID, err
}

// Update обновляет Warehouse и его Location.
func (r *WarehouseRepo) Update(warehouseID int, input UpdateWarehouseInput) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var wh locmodel.Warehouse
		if err := tx.First(&wh, warehouseID).Error; err != nil {
			return err
		}
		var loc locmodel.Location
		if err := tx.Where("warehouse_id = ?", warehouseID).First(&loc).Error; err != nil {
			return err
		}

		// Обновляем Warehouse
		if input.ShortName != nil {
			wh.ShortName = input.ShortName
		}
		if input.WarehouseName != nil {
			wh.FullName = input.WarehouseName
		}
		if input.StreetAddress != nil {
			wh.StreetAddress = input.StreetAddress
			loc.StreetAddress = input.StreetAddress
		}
		if input.AddressLine2 != nil {
			wh.AddressLine2 = input.AddressLine2
			loc.AddressLine2 = input.AddressLine2
		}
		if input.City != nil {
			wh.City = input.City
			loc.City = input.City
		}
		if input.State != nil {
			wh.State = input.State
			loc.State = input.State
		}
		if input.PostalCode != nil {
			wh.PostalCode = input.PostalCode
			loc.PostalCode = input.PostalCode
		}
		if input.Phone != nil {
			wh.Phone = input.Phone
			loc.Phone = input.Phone
		}

		// Обновляем Location
		if input.Email != nil {
			loc.Email = input.Email
		}
		if input.StoreID != nil {
			loc.StoreID = *input.StoreID
		}
		if input.CanReceiveItems != nil {
			loc.CanReceiveItems = input.CanReceiveItems
		}
		if input.Active != nil {
			loc.StoreActive = input.Active
		}

		if err := tx.Save(&wh).Error; err != nil {
			return err
		}
		return tx.Save(&loc).Error
	})
}

// ─────────────────────── helpers ───────────────────────────

// insertWarehousePermissions создаёт PermissionsCombination для склада.
func insertWarehousePermissions(tx *gorm.DB, subBlockID int, permIDs []int) error {
	type permRow struct {
		PermissionsID      int
		PermissionsBlockID int
	}
	var rows []permRow
	tx.Table("permissions").
		Select("permissions_id, permissions_block_id").
		Where("permissions_id IN ?", permIDs).
		Scan(&rows)

	for _, row := range rows {
		var count int64
		tx.Model(&permmodel.PermissionsCombination{}).
			Where("permissions_block_id = ? AND permissions_sub_block_warehouse_id = ? AND permissions_id = ?",
				row.PermissionsBlockID, subBlockID, row.PermissionsID).
			Count(&count)
		if count == 0 {
			blockID := row.PermissionsBlockID
			if err := tx.Create(&permmodel.PermissionsCombination{
				PermissionsBlockID:             &blockID,
				PermissionsSubBlockWarehouseID: &subBlockID,
				PermissionsID:                  row.PermissionsID,
			}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
