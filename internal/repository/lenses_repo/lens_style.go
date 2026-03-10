package lenses_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/lenses"
)

type LensStyleRepo struct{ DB *gorm.DB }

func NewLensStyleRepo(db *gorm.DB) *LensStyleRepo { return &LensStyleRepo{DB: db} }

func (r *LensStyleRepo) GetAll() ([]lenses.LensStyle, error) {
	var items []lenses.LensStyle
	return items, r.DB.Find(&items).Error
}

func (r *LensStyleRepo) GetByID(id int) (*lenses.LensStyle, error) {
	var v lenses.LensStyle
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *LensStyleRepo) Create(v *lenses.LensStyle) error { return r.DB.Create(v).Error }
func (r *LensStyleRepo) Save(v *lenses.LensStyle) error   { return r.DB.Save(v).Error }
func (r *LensStyleRepo) Delete(id int) error {
	return r.DB.Delete(&lenses.LensStyle{}, id).Error
}
