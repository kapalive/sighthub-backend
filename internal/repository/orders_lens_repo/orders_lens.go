// internal/repository/orders_lens_repo/orders_lens.go
package orders_lens_repo

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"sighthub-backend/internal/models/orders_lens"
)

type OrdersLensRepo struct{ DB *gorm.DB }

func NewOrdersLensRepo(db *gorm.DB) *OrdersLensRepo { return &OrdersLensRepo{DB: db} }

// GetByID возвращает заказ линзы по ID.
func (r *OrdersLensRepo) GetByID(id int64) (*orders_lens.OrdersLens, error) {
	var row orders_lens.OrdersLens
	err := r.DB.Preload("Status").First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// GetByPatientID возвращает все заказы линз для пациента.
func (r *OrdersLensRepo) GetByPatientID(patientID int64) ([]orders_lens.OrdersLens, error) {
	var rows []orders_lens.OrdersLens
	return rows, r.DB.Preload("Status").
		Where("patient_id = ?", patientID).
		Order("date_create DESC").
		Find(&rows).Error
}

// CreateInput — данные для нового заказа линз.
type CreateOrdersLensInput struct {
	NumberOrder        string
	DateCreate         time.Time
	PromisedDate       time.Time
	PromisedTimeBy     *string
	StatusOrdersLensID int
	LensID             *int
	PatientID          *int64
	Note               *string
}

// Create создаёт заказ линзы.
func (r *OrdersLensRepo) Create(inp CreateOrdersLensInput) (*orders_lens.OrdersLens, error) {
	ol := &orders_lens.OrdersLens{
		NumberOrder:        inp.NumberOrder,
		DateCreate:         inp.DateCreate,
		PromisedDate:       inp.PromisedDate,
		PromisedTimeBy:     inp.PromisedTimeBy,
		StatusOrdersLensID: inp.StatusOrdersLensID,
		LensID:             inp.LensID,
		PatientID:          inp.PatientID,
		Note:               inp.Note,
	}
	return ol, r.DB.Create(ol).Error
}

// UpdateInput — изменяемые поля.
type UpdateOrdersLensInput struct {
	PromisedDate       *time.Time
	PromisedTimeBy     *string
	StatusOrdersLensID *int
	Note               *string
}

// Update обновляет заказ линзы.
func (r *OrdersLensRepo) Update(id int64, inp UpdateOrdersLensInput) error {
	updates := map[string]interface{}{}
	if inp.PromisedDate != nil       { updates["promised_date"]        = *inp.PromisedDate }
	if inp.PromisedTimeBy != nil     { updates["promised_time_by"]     = *inp.PromisedTimeBy }
	if inp.StatusOrdersLensID != nil { updates["status_orders_lens_id"] = *inp.StatusOrdersLensID }
	if inp.Note != nil               { updates["note"]                  = *inp.Note }
	if len(updates) == 0 {
		return nil
	}
	return r.DB.Model(&orders_lens.OrdersLens{}).Where("id_orders_lens = ?", id).Updates(updates).Error
}

func (r *OrdersLensRepo) IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
