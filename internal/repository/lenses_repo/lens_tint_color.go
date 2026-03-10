package lenses_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/lenses"
)

type LensTintColorRepo struct{ DB *gorm.DB }

func NewLensTintColorRepo(db *gorm.DB) *LensTintColorRepo { return &LensTintColorRepo{DB: db} }

func (r *LensTintColorRepo) GetAll() ([]lenses.LensTintColor, error) {
	var items []lenses.LensTintColor
	return items, r.DB.Find(&items).Error
}

func (r *LensTintColorRepo) GetByID(id int) (*lenses.LensTintColor, error) {
	var v lenses.LensTintColor
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *LensTintColorRepo) Create(v *lenses.LensTintColor) error { return r.DB.Create(v).Error }
func (r *LensTintColorRepo) Save(v *lenses.LensTintColor) error   { return r.DB.Save(v).Error }
func (r *LensTintColorRepo) Delete(id int) error {
	return r.DB.Delete(&lenses.LensTintColor{}, id).Error
}
