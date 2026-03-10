package lenses_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/lenses"
)

type LensBevelRepo struct{ DB *gorm.DB }

func NewLensBevelRepo(db *gorm.DB) *LensBevelRepo { return &LensBevelRepo{DB: db} }

func (r *LensBevelRepo) GetAll() ([]lenses.LensBevel, error) {
	var items []lenses.LensBevel
	return items, r.DB.Find(&items).Error
}

func (r *LensBevelRepo) GetByID(id int) (*lenses.LensBevel, error) {
	var v lenses.LensBevel
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *LensBevelRepo) Create(v *lenses.LensBevel) error { return r.DB.Create(v).Error }
func (r *LensBevelRepo) Save(v *lenses.LensBevel) error   { return r.DB.Save(v).Error }
func (r *LensBevelRepo) Delete(id int) error {
	return r.DB.Delete(&lenses.LensBevel{}, id).Error
}
