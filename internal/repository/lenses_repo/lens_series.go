package lenses_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/lenses"
)

type LensSeriesRepo struct{ DB *gorm.DB }

func NewLensSeriesRepo(db *gorm.DB) *LensSeriesRepo { return &LensSeriesRepo{DB: db} }

func (r *LensSeriesRepo) GetAll() ([]lenses.LensSeries, error) {
	var items []lenses.LensSeries
	return items, r.DB.Find(&items).Error
}

func (r *LensSeriesRepo) GetByID(id int) (*lenses.LensSeries, error) {
	var v lenses.LensSeries
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *LensSeriesRepo) Create(v *lenses.LensSeries) error { return r.DB.Create(v).Error }
func (r *LensSeriesRepo) Save(v *lenses.LensSeries) error   { return r.DB.Save(v).Error }
func (r *LensSeriesRepo) Delete(id int) error {
	return r.DB.Delete(&lenses.LensSeries{}, id).Error
}
