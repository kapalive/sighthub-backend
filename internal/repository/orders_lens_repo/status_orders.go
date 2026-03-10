// internal/repository/orders_lens_repo/status_orders.go
package orders_lens_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/orders_lens"
)

type StatusOrdersRepo struct{ DB *gorm.DB }

func NewStatusOrdersRepo(db *gorm.DB) *StatusOrdersRepo { return &StatusOrdersRepo{DB: db} }

func (r *StatusOrdersRepo) GetAll() ([]orders_lens.StatusOrdersLens, error) {
	var rows []orders_lens.StatusOrdersLens
	return rows, r.DB.Find(&rows).Error
}

func (r *StatusOrdersRepo) GetByID(id int) (*orders_lens.StatusOrdersLens, error) {
	var row orders_lens.StatusOrdersLens
	err := r.DB.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}
