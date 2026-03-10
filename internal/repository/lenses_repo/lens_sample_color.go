package lenses_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/lenses"
)

type LensSampleColorRepo struct{ DB *gorm.DB }

func NewLensSampleColorRepo(db *gorm.DB) *LensSampleColorRepo { return &LensSampleColorRepo{DB: db} }

func (r *LensSampleColorRepo) GetAll() ([]lenses.LensSampleColor, error) {
	var items []lenses.LensSampleColor
	return items, r.DB.Find(&items).Error
}

func (r *LensSampleColorRepo) GetByID(id int) (*lenses.LensSampleColor, error) {
	var v lenses.LensSampleColor
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *LensSampleColorRepo) Create(v *lenses.LensSampleColor) error { return r.DB.Create(v).Error }
func (r *LensSampleColorRepo) Save(v *lenses.LensSampleColor) error   { return r.DB.Save(v).Error }
func (r *LensSampleColorRepo) Delete(id int) error {
	return r.DB.Delete(&lenses.LensSampleColor{}, id).Error
}
