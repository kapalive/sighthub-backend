package lenses_repo

import (
	"errors"
	"gorm.io/gorm"
	"sighthub-backend/internal/models/lenses"
)

type VCodesLensRepo struct{ DB *gorm.DB }

func NewVCodesLensRepo(db *gorm.DB) *VCodesLensRepo { return &VCodesLensRepo{DB: db} }

func (r *VCodesLensRepo) GetAll() ([]lenses.VCodesLens, error) {
	var items []lenses.VCodesLens
	return items, r.DB.Find(&items).Error
}

func (r *VCodesLensRepo) GetByID(id int) (*lenses.VCodesLens, error) {
	var v lenses.VCodesLens
	if err := r.DB.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, nil }
		return nil, err
	}
	return &v, nil
}

func (r *VCodesLensRepo) Create(v *lenses.VCodesLens) error { return r.DB.Create(v).Error }
func (r *VCodesLensRepo) Save(v *lenses.VCodesLens) error   { return r.DB.Save(v).Error }
func (r *VCodesLensRepo) Delete(id int) error {
	return r.DB.Delete(&lenses.VCodesLens{}, id).Error
}
