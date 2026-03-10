// internal/repository/location_repo/store.go
// CRUD для Store + связанных Location/WorkShift/Permissions.
// Аналог store_bp из store_routes.py.
package location_repo

import (
	"strings"

	"gorm.io/gorm"
	empmodel "sighthub-backend/internal/models/employees"
	locmodel "sighthub-backend/internal/models/location"
	permmodel "sighthub-backend/internal/models/permission"
)

// defaultStorePermissionIDs — ID прав, создаваемых по умолчанию для каждого стора.
var defaultStorePermissionIDs = []int{
	1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
	21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 41, 42, 43,
	44, 45, 46, 47, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75,
	76, 77, 78, 79,
}

// ─────────────────────── input types ───────────────────────

// StoreDetailsInput — поля деталей при создании/обновлении стора.
type StoreDetailsInput struct {
	FullName        string  `json:"full_name"`
	ShortName       string  `json:"short_name"`
	BusinessName    *string `json:"business_name"`
	Phone           *string `json:"phone"`
	Fax             *string `json:"fax"`
	Email           *string `json:"email"`
	Website         *string `json:"website"`
	NPI             *string `json:"npi"`
	TaxN            *string `json:"tax_n"`
	HPSA            *string `json:"hpsa"`
	Logo            *string `json:"logo"`
	CanReceiveItems bool    `json:"can_receive_items"`
	Showcase        bool    `json:"showcase"`
	StoreActive     *bool   `json:"store_active"`
}

// StoreAddressInput — адресные поля.
type StoreAddressInput struct {
	StreetAddress *string `json:"street_address"`
	AddressLine2  *string `json:"address_line_2"`
	City          *string `json:"city"`
	State         *string `json:"state"`
	PostalCode    *string `json:"postal_code"`
	Country       *string `json:"country"`
}

// WorkShiftInput — расписание.
type WorkShiftInput struct {
	Title              *string `json:"title"`
	Monday             *bool   `json:"monday"`
	Tuesday            *bool   `json:"tuesday"`
	Wednesday          *bool   `json:"wednesday"`
	Thursday           *bool   `json:"thursday"`
	Friday             *bool   `json:"friday"`
	Saturday           *bool   `json:"saturday"`
	Sunday             *bool   `json:"sunday"`
	MondayStart        *string `json:"monday_time_start"`
	MondayEnd          *string `json:"monday_time_end"`
	TuesdayStart       *string `json:"tuesday_time_start"`
	TuesdayEnd         *string `json:"tuesday_time_end"`
	WednesdayStart     *string `json:"wednesday_time_start"`
	WednesdayEnd       *string `json:"wednesday_time_end"`
	ThursdayStart      *string `json:"thursday_time_start"`
	ThursdayEnd        *string `json:"thursday_time_end"`
	FridayStart        *string `json:"friday_time_start"`
	FridayEnd          *string `json:"friday_time_end"`
	SaturdayStart      *string `json:"saturday_time_start"`
	SaturdayEnd        *string `json:"saturday_time_end"`
	SundayStart        *string `json:"sunday_time_start"`
	SundayEnd          *string `json:"sunday_time_end"`
	LunchDuration      *string `json:"lunch_duration"`
}

// CreateStoreInput — полный входной JSON для создания стора.
type CreateStoreInput struct {
	Details  StoreDetailsInput  `json:"details"`
	Address  StoreAddressInput  `json:"address"`
	Schedule *WorkShiftInput    `json:"schedule"`
}

// UpdateStoreInput — входные данные для частичного обновления.
type UpdateStoreInput struct {
	Details  map[string]interface{} `json:"details"`
	Address  map[string]interface{} `json:"address"`
	Schedule *WorkShiftInput        `json:"schedule"`
}

// ─────────────────────── response types ────────────────────

// StoreListItem — элемент списка для GET /stores.
type StoreListItem struct {
	IDStore      int     `json:"id_store"`
	BusinessName *string `json:"business_name"`
	FullName     *string `json:"full_name"`
	ShortName    *string `json:"short_name"`
	Phone        *string `json:"phone"`
	Address      string  `json:"address"`
	Active       *bool   `json:"active"`
}

// StoreDetail — детальный ответ для GET /stores/:id.
type StoreDetail struct {
	Details  map[string]interface{} `json:"details"`
	Address  map[string]interface{} `json:"address"`
	Schedule map[string]interface{} `json:"schedule"`
}

// ─────────────────────── repo ──────────────────────────────

// StoreRepo — репозиторий для Store.
type StoreRepo struct {
	DB *gorm.DB
}

func NewStoreRepo(db *gorm.DB) *StoreRepo { return &StoreRepo{DB: db} }

// GetAll возвращает список всех сторов с их активным статусом из location.
func (r *StoreRepo) GetAll() ([]StoreListItem, error) {
	var stores []locmodel.Store
	if err := r.DB.Find(&stores).Error; err != nil {
		return nil, err
	}

	result := make([]StoreListItem, 0, len(stores))
	for _, s := range stores {
		var loc locmodel.Location
		r.DB.Where("store_id = ? AND warehouse_id IS NULL", s.IDStore).First(&loc)

		parts := []string{}
		for _, p := range []*string{s.StreetAddress, s.AddressLine2, s.City, s.State, s.Country} {
			if p != nil && *p != "" {
				parts = append(parts, *p)
			}
		}
		result = append(result, StoreListItem{
			IDStore:      s.IDStore,
			BusinessName: s.BusinessName,
			FullName:     s.FullName,
			ShortName:    s.ShortName,
			Phone:        s.Phone,
			Address:      strings.Join(parts, " "),
			Active:       loc.StoreActive,
		})
	}
	return result, nil
}

// GetByID возвращает Store по ID.
func (r *StoreRepo) GetByID(id int) (*locmodel.Store, error) {
	var s locmodel.Store
	err := r.DB.First(&s, id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetDetail возвращает StoreDetail с details/address/schedule для GET /stores/:id.
func (r *StoreRepo) GetDetail(storeID int) (*StoreDetail, error) {
	var s locmodel.Store
	if err := r.DB.First(&s, storeID).Error; err != nil {
		return nil, err
	}

	var loc locmodel.Location
	if err := r.DB.Where("store_id = ? AND warehouse_id IS NULL", storeID).First(&loc).Error; err != nil {
		return nil, err
	}

	var schedule map[string]interface{}
	if loc.WorkShiftID != nil {
		var ws empmodel.WorkShift
		if err := r.DB.First(&ws, *loc.WorkShiftID).Error; err == nil {
			schedule = ws.ToMap()
		}
	}

	return &StoreDetail{
		Details: map[string]interface{}{
			"business_name":    s.BusinessName,
			"full_name":        s.FullName,
			"short_name":       s.ShortName,
			"hash":             s.Hash,
			"website":          loc.Website,
			"phone":            s.Phone,
			"email":            s.Email,
			"fax":              s.Fax,
			"can_receive_items": loc.CanReceiveItems,
			"logo":             s.Logo,
			"npi":              s.NPI,
			"tax_n":            s.TaxN,
			"hpsa":             s.HPSA,
			"store_active":     loc.StoreActive,
		},
		Address: map[string]interface{}{
			"street_address": s.StreetAddress,
			"address_line_2": s.AddressLine2,
			"city":           s.City,
			"state":          s.State,
			"postal_code":    s.PostalCode,
			"country":        s.Country,
		},
		Schedule: schedule,
	}, nil
}

// Create создаёт Store + Location + (опционально) WorkShift + PermissionsSubBlock.
// Всё выполняется в одной транзакции.
func (r *StoreRepo) Create(input CreateStoreInput) (*locmodel.Store, *locmodel.Location, error) {
	var createdStore locmodel.Store
	var createdLoc locmodel.Location

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Store
		fn := input.Details.FullName
		sn := input.Details.ShortName
		store := locmodel.Store{
			FullName:      &fn,
			ShortName:     &sn,
			BusinessName:  input.Details.BusinessName,
			Phone:         input.Details.Phone,
			Fax:           input.Details.Fax,
			Email:         input.Details.Email,
			NPI:           input.Details.NPI,
			TaxN:          input.Details.TaxN,
			HPSA:          input.Details.HPSA,
			Logo:          input.Details.Logo,
			StreetAddress: input.Address.StreetAddress,
			AddressLine2:  input.Address.AddressLine2,
			City:          input.Address.City,
			State:         input.Address.State,
			PostalCode:    input.Address.PostalCode,
			Country:       input.Address.Country,
		}
		if err := tx.Create(&store).Error; err != nil {
			return err
		}

		// 2. SalesTax lookup
		var taxID *int
		if input.Address.State != nil && *input.Address.State != "" {
			taxID = findSalesTaxByState(tx, *input.Address.State)
		}

		// 3. WorkShift (опционально)
		var workShiftID *int
		if input.Schedule != nil {
			ws := buildWorkShift(input.Schedule)
			if err := tx.Create(&ws).Error; err != nil {
				return err
			}
			id := int(ws.IDWorkShift)
			workShiftID = &id
		}

		// 4. Location
		canReceive := input.Details.CanReceiveItems
		showcase := input.Details.Showcase
		loc := locmodel.Location{
			FullName:        fn,
			ShortName:       &sn,
			StreetAddress:   input.Address.StreetAddress,
			AddressLine2:    input.Address.AddressLine2,
			City:            input.Address.City,
			State:           input.Address.State,
			PostalCode:      input.Address.PostalCode,
			Country:         input.Address.Country,
			Phone:           input.Details.Phone,
			Website:         input.Details.Website,
			Email:           input.Details.Email,
			Fax:             input.Details.Fax,
			CanReceiveItems: &canReceive,
			Showcase:        &showcase,
			LogoPath:        input.Details.Logo,
			StoreID:         store.IDStore,
			SalesTaxID:      taxID,
			WorkShiftID:     workShiftID,
		}
		if err := tx.Create(&loc).Error; err != nil {
			return err
		}

		// 5. PermissionsSubBlockStore
		storeID := store.IDStore
		sub := permmodel.PermissionsSubBlockStore{
			SubBlockName: fn,
			StoreID:      &storeID,
		}
		if err := tx.Create(&sub).Error; err != nil {
			return err
		}

		// 6. Bulk-insert PermissionsCombination для дефолтных прав
		if err := insertStorePermissions(tx, sub.IDPermissionsSubBlock, defaultStorePermissionIDs); err != nil {
			return err
		}

		createdStore = store
		createdLoc = loc
		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	return &createdStore, &createdLoc, nil
}

// Update обновляет Store и его Location, опционально WorkShift.
func (r *StoreRepo) Update(storeID int, input UpdateStoreInput) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var store locmodel.Store
		if err := tx.First(&store, storeID).Error; err != nil {
			return err
		}
		var loc locmodel.Location
		if err := tx.Where("store_id = ? AND warehouse_id IS NULL", storeID).First(&loc).Error; err != nil {
			return err
		}

		// Обновляем Store-поля из details и address
		applyStoreFields(&store, input.Details, input.Address)
		if err := tx.Save(&store).Error; err != nil {
			return err
		}

		// Обновляем Location-поля
		applyLocationFields(&loc, input.Details, input.Address)
		if err := tx.Save(&loc).Error; err != nil {
			return err
		}

		// WorkShift
		if input.Schedule != nil {
			if loc.WorkShiftID != nil {
				var ws empmodel.WorkShift
				if err := tx.First(&ws, *loc.WorkShiftID).Error; err == nil {
					applyWorkShift(&ws, input.Schedule)
					if err := tx.Save(&ws).Error; err != nil {
						return err
					}
				}
			} else {
				ws := buildWorkShift(input.Schedule)
				if err := tx.Create(&ws).Error; err != nil {
					return err
				}
				id := int(ws.IDWorkShift)
				loc.WorkShiftID = &id
				if err := tx.Save(&loc).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// ─────────────────────── helpers ───────────────────────────

// buildWorkShift создаёт WorkShift из input (используются дефолты из Python).
func buildWorkShift(in *WorkShiftInput) empmodel.WorkShift {
	ws := empmodel.WorkShift{
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		Saturday:  false,
		Sunday:    false,
	}
	ws.Title = in.Title
	applyWorkShift(&ws, in)
	return ws
}

// applyWorkShift применяет изменения из input к существующему WorkShift.
func applyWorkShift(ws *empmodel.WorkShift, in *WorkShiftInput) {
	if in.Monday != nil {
		ws.Monday = *in.Monday
	}
	if in.Tuesday != nil {
		ws.Tuesday = *in.Tuesday
	}
	if in.Wednesday != nil {
		ws.Wednesday = *in.Wednesday
	}
	if in.Thursday != nil {
		ws.Thursday = *in.Thursday
	}
	if in.Friday != nil {
		ws.Friday = *in.Friday
	}
	if in.Saturday != nil {
		ws.Saturday = *in.Saturday
	}
	if in.Sunday != nil {
		ws.Sunday = *in.Sunday
	}
	if in.LunchDuration != nil {
		ws.LunchDuration = in.LunchDuration
	}
}

// applyStoreFields применяет details/address map к Store struct.
func applyStoreFields(s *locmodel.Store, details, address map[string]interface{}) {
	if v, ok := strVal(details, "business_name"); ok {
		s.BusinessName = v
	}
	if v, ok := strVal(details, "full_name"); ok {
		s.FullName = v
	}
	if v, ok := strVal(details, "short_name"); ok {
		s.ShortName = v
	}
	if v, ok := strVal(details, "phone"); ok {
		s.Phone = v
	}
	if v, ok := strVal(details, "email"); ok {
		s.Email = v
	}
	if v, ok := strVal(details, "fax"); ok {
		s.Fax = v
	}
	if v, ok := strVal(details, "npi"); ok {
		s.NPI = v
	}
	if v, ok := strVal(details, "tax_n"); ok {
		s.TaxN = v
	}
	if v, ok := strVal(details, "hpsa"); ok {
		s.HPSA = v
	}
	if v, ok := strVal(details, "logo"); ok {
		s.Logo = v
	}
	if v, ok := strVal(address, "street_address"); ok {
		s.StreetAddress = v
	}
	if v, ok := strVal(address, "address_line_2"); ok {
		s.AddressLine2 = v
	}
	if v, ok := strVal(address, "city"); ok {
		s.City = v
	}
	if v, ok := strVal(address, "state"); ok {
		s.State = v
	}
	if v, ok := strVal(address, "postal_code"); ok {
		s.PostalCode = v
	}
	if v, ok := strVal(address, "country"); ok {
		s.Country = v
	}
}

// applyLocationFields применяет details/address map к Location struct.
func applyLocationFields(l *locmodel.Location, details, address map[string]interface{}) {
	if v, ok := strVal(details, "full_name"); ok {
		if v != nil {
			l.FullName = *v
		}
	}
	if v, ok := strVal(details, "short_name"); ok {
		l.ShortName = v
	}
	if v, ok := strVal(details, "phone"); ok {
		l.Phone = v
	}
	if v, ok := strVal(details, "email"); ok {
		l.Email = v
	}
	if v, ok := strVal(details, "website"); ok {
		l.Website = v
	}
	if v, ok := strVal(details, "fax"); ok {
		l.Fax = v
	}
	if v, ok := strVal(details, "logo"); ok {
		l.LogoPath = v
	}
	if raw, ok := details["can_receive_items"]; ok {
		if b, ok := raw.(bool); ok {
			l.CanReceiveItems = &b
		}
	}
	if raw, ok := details["store_active"]; ok {
		if b, ok := raw.(bool); ok {
			l.StoreActive = &b
		}
	}
	if v, ok := strVal(address, "street_address"); ok {
		l.StreetAddress = v
	}
	if v, ok := strVal(address, "address_line_2"); ok {
		l.AddressLine2 = v
	}
	if v, ok := strVal(address, "city"); ok {
		l.City = v
	}
	if v, ok := strVal(address, "state"); ok {
		l.State = v
	}
	if v, ok := strVal(address, "postal_code"); ok {
		l.PostalCode = v
	}
	if v, ok := strVal(address, "country"); ok {
		l.Country = v
	}
}

// insertStorePermissions создаёт PermissionsCombination для стора.
// Пропускает комбинации, которые уже существуют.
func insertStorePermissions(tx *gorm.DB, subBlockID int, permIDs []int) error {
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
			Where("permissions_block_id = ? AND permissions_sub_block_store_id = ? AND permissions_id = ?",
				row.PermissionsBlockID, subBlockID, row.PermissionsID).
			Count(&count)
		if count == 0 {
			blockID := row.PermissionsBlockID
			if err := tx.Create(&permmodel.PermissionsCombination{
				PermissionsBlockID:         &blockID,
				PermissionsSubBlockStoreID: &subBlockID,
				PermissionsID:              row.PermissionsID,
			}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// strVal извлекает *string из map[string]interface{}.
// ok=true означает ключ присутствует в map (значение может быть nil).
func strVal(m map[string]interface{}, key string) (*string, bool) {
	if m == nil {
		return nil, false
	}
	raw, ok := m[key]
	if !ok {
		return nil, false
	}
	if raw == nil {
		return nil, true
	}
	if s, ok := raw.(string); ok {
		return &s, true
	}
	return nil, true
}
