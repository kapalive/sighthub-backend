package general_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/general"
)

type PQRSRepo struct{ DB *gorm.DB }

func NewPQRSRepo(db *gorm.DB) *PQRSRepo { return &PQRSRepo{DB: db} }

func (r *PQRSRepo) GetAll() ([]general.PQRS, error) {
	var items []general.PQRS
	return items, r.DB.Preload("GroupPQRSRef").Order("code").Find(&items).Error
}

func (r *PQRSRepo) GetByID(id int64) (*general.PQRS, error) {
	var v general.PQRS
	if err := r.DB.Preload("GroupPQRSRef").First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *PQRSRepo) GetByGroupID(groupID int) ([]general.PQRS, error) {
	var items []general.PQRS
	return items, r.DB.Where("pqrs_group_id = ?", groupID).Order("code").Find(&items).Error
}

func (r *PQRSRepo) Create(v *general.PQRS) error { return r.DB.Create(v).Error }
func (r *PQRSRepo) Save(v *general.PQRS) error   { return r.DB.Save(v).Error }
