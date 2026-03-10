package lenses_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/lenses"
)

type PriceBookLensRepo struct{ DB *gorm.DB }

func NewPriceBookLensRepo(db *gorm.DB) *PriceBookLensRepo { return &PriceBookLensRepo{DB: db} }

func (r *PriceBookLensRepo) GetByID(id int64) (*lenses.PriceBookLens, error) {
	var v lenses.PriceBookLens
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *PriceBookLensRepo) GetByLensID(lensID int) ([]lenses.PriceBookLens, error) {
	var items []lenses.PriceBookLens
	return items, r.DB.Where("lenses_id = ?", lensID).Find(&items).Error
}

func (r *PriceBookLensRepo) GetByBrandID(brandID int) ([]lenses.PriceBookLens, error) {
	var items []lenses.PriceBookLens
	return items, r.DB.Where("brand_lens_id = ?", brandID).Find(&items).Error
}

func (r *PriceBookLensRepo) Create(v *lenses.PriceBookLens) error { return r.DB.Create(v).Error }
func (r *PriceBookLensRepo) Save(v *lenses.PriceBookLens) error   { return r.DB.Save(v).Error }
func (r *PriceBookLensRepo) Delete(id int64) error {
	return r.DB.Delete(&lenses.PriceBookLens{}, id).Error
}
