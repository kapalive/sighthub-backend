package lenses_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/lenses"
)

type LensSpecialFeatureRepo struct{ DB *gorm.DB }

func NewLensSpecialFeatureRepo(db *gorm.DB) *LensSpecialFeatureRepo { return &LensSpecialFeatureRepo{DB: db} }

func (r *LensSpecialFeatureRepo) GetAll() ([]lenses.LensSpecialFeature, error) {
	var items []lenses.LensSpecialFeature
	return items, r.DB.Find(&items).Error
}

func (r *LensSpecialFeatureRepo) GetByID(id int) (*lenses.LensSpecialFeature, error) {
	var v lenses.LensSpecialFeature
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *LensSpecialFeatureRepo) Create(v *lenses.LensSpecialFeature) error { return r.DB.Create(v).Error }
func (r *LensSpecialFeatureRepo) Save(v *lenses.LensSpecialFeature) error   { return r.DB.Save(v).Error }
func (r *LensSpecialFeatureRepo) Delete(id int) error {
	return r.DB.Delete(&lenses.LensSpecialFeature{}, id).Error
}
